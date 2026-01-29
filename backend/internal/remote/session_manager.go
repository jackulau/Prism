package remote

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/security"
)

// SessionState represents the current state of a remote session
type SessionState int

const (
	StateActive SessionState = iota
	StateDisconnected
	StateReconnecting
	StateClosed
)

func (s SessionState) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateDisconnected:
		return "disconnected"
	case StateReconnecting:
		return "reconnecting"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// SessionConfig holds configuration for the session manager
type SessionConfig struct {
	HeartbeatInterval    time.Duration
	HeartbeatTimeout     time.Duration
	IdleTimeout          time.Duration
	IdleWarningBefore    time.Duration
	ReconnectTokenExpiry time.Duration
	CleanupInterval      time.Duration
}

// DefaultSessionConfig returns the default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		HeartbeatInterval:    30 * time.Second,
		HeartbeatTimeout:     90 * time.Second,
		IdleTimeout:          30 * time.Minute,
		IdleWarningBefore:    5 * time.Minute,
		ReconnectTokenExpiry: 5 * time.Minute,
		CleanupInterval:      1 * time.Minute,
	}
}

// RemoteSession represents an active remote connection session
type RemoteSession struct {
	ID              string
	UserID          string
	AuthSession     *security.Claims
	LastActivity    time.Time
	LastHeartbeat   time.Time
	ReconnectToken  string
	ReconnectExpiry time.Time
	State           SessionState
	CreatedAt       time.Time
	DisconnectedAt  *time.Time

	// Pending messages while disconnected
	PendingMessages []PendingMessage

	// Metadata
	ClientInfo map[string]string

	mu sync.RWMutex
}

// GetInfo returns a snapshot of session information (thread-safe)
func (s *RemoteSession) GetInfo() SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := SessionInfo{
		ID:              s.ID,
		UserID:          s.UserID,
		State:           s.State.String(),
		CreatedAt:       s.CreatedAt.UnixMilli(),
		LastActivity:    s.LastActivity.UnixMilli(),
		LastHeartbeat:   s.LastHeartbeat.UnixMilli(),
		PendingMessages: len(s.PendingMessages),
		ClientInfo:      make(map[string]string),
	}

	if s.DisconnectedAt != nil {
		ts := s.DisconnectedAt.UnixMilli()
		info.DisconnectedAt = &ts
	}

	for k, v := range s.ClientInfo {
		info.ClientInfo[k] = v
	}

	return info
}

// SessionInfo represents a thread-safe snapshot of session information
type SessionInfo struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	State           string            `json:"state"`
	CreatedAt       int64             `json:"created_at"`
	LastActivity    int64             `json:"last_activity"`
	LastHeartbeat   int64             `json:"last_heartbeat"`
	DisconnectedAt  *int64            `json:"disconnected_at,omitempty"`
	ClientInfo      map[string]string `json:"client_info,omitempty"`
	PendingMessages int               `json:"pending_messages"`
}

// PendingMessage represents a message queued during disconnection
type PendingMessage struct {
	Type      string
	Data      interface{}
	Timestamp time.Time
}

// SessionManager manages remote connection sessions
type SessionManager struct {
	sessions        map[string]*RemoteSession
	reconnectTokens map[string]string // reconnect token -> session ID
	userSessions    map[string][]string // user ID -> session IDs
	config          *SessionConfig
	mu              sync.RWMutex

	// Event callbacks
	onSessionCreated     func(*RemoteSession)
	onSessionClosed      func(*RemoteSession)
	onSessionDisconnect  func(*RemoteSession)
	onSessionReconnect   func(*RemoteSession)
	onIdleWarning        func(*RemoteSession)
	onHeartbeatTimeout   func(*RemoteSession)

	// Cleanup ticker
	stopCleanup chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionConfig) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	sm := &SessionManager{
		sessions:        make(map[string]*RemoteSession),
		reconnectTokens: make(map[string]string),
		userSessions:    make(map[string][]string),
		config:          config,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

// CreateSession creates a new remote session
func (sm *SessionManager) CreateSession(userID string, authSession *security.Claims, clientInfo map[string]string) (*RemoteSession, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	reconnectToken, err := generateReconnectToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &RemoteSession{
		ID:              sessionID,
		UserID:          userID,
		AuthSession:     authSession,
		LastActivity:    now,
		LastHeartbeat:   now,
		ReconnectToken:  reconnectToken,
		ReconnectExpiry: now.Add(sm.config.ReconnectTokenExpiry),
		State:           StateActive,
		CreatedAt:       now,
		ClientInfo:      clientInfo,
		PendingMessages: make([]PendingMessage, 0),
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.reconnectTokens[reconnectToken] = sessionID
	sm.userSessions[userID] = append(sm.userSessions[userID], sessionID)
	sm.mu.Unlock()

	if sm.onSessionCreated != nil {
		sm.onSessionCreated(session)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*RemoteSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// GetSessionByReconnectToken retrieves a session using a reconnect token
func (sm *SessionManager) GetSessionByReconnectToken(token string) (*RemoteSession, error) {
	sm.mu.RLock()
	sessionID, exists := sm.reconnectTokens[token]
	sm.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidReconnectToken
	}

	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.RLock()
	expired := time.Now().After(session.ReconnectExpiry)
	session.mu.RUnlock()

	if expired {
		return nil, ErrReconnectTokenExpired
	}

	return session, nil
}

// GetUserSessions returns all sessions for a user
func (sm *SessionManager) GetUserSessions(userID string) []*RemoteSession {
	sm.mu.RLock()
	sessionIDs := sm.userSessions[userID]
	sm.mu.RUnlock()

	sessions := make([]*RemoteSession, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		if session, err := sm.GetSession(id); err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// UpdateActivity updates the last activity timestamp
func (sm *SessionManager) UpdateActivity(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	session.LastActivity = time.Now()
	session.mu.Unlock()

	return nil
}

// UpdateHeartbeat updates the last heartbeat timestamp
func (sm *SessionManager) UpdateHeartbeat(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	session.LastHeartbeat = time.Now()
	if session.State == StateReconnecting {
		session.State = StateActive
		session.DisconnectedAt = nil
	}
	session.mu.Unlock()

	return nil
}

// MarkDisconnected marks a session as disconnected
func (sm *SessionManager) MarkDisconnected(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	if session.State == StateClosed {
		session.mu.Unlock()
		return ErrSessionClosed
	}

	now := time.Now()
	session.State = StateDisconnected
	session.DisconnectedAt = &now
	session.ReconnectExpiry = now.Add(sm.config.ReconnectTokenExpiry)
	session.mu.Unlock()

	if sm.onSessionDisconnect != nil {
		sm.onSessionDisconnect(session)
	}

	return nil
}

// Reconnect attempts to reconnect a session using a reconnect token
func (sm *SessionManager) Reconnect(token string) (*RemoteSession, error) {
	session, err := sm.GetSessionByReconnectToken(token)
	if err != nil {
		return nil, err
	}

	// Generate new reconnect token first (before locking)
	newToken, err := generateReconnectToken()
	if err != nil {
		return nil, err
	}

	// Lock in consistent order: sm.mu first, then session.mu
	// This prevents deadlock with cleanup loop
	sm.mu.Lock()
	session.mu.Lock()

	if session.State == StateClosed {
		session.mu.Unlock()
		sm.mu.Unlock()
		return nil, ErrSessionClosed
	}

	// Update session state
	now := time.Now()
	oldToken := session.ReconnectToken
	session.ReconnectToken = newToken
	session.ReconnectExpiry = now.Add(sm.config.ReconnectTokenExpiry)
	session.State = StateActive
	session.LastActivity = now
	session.LastHeartbeat = now
	session.DisconnectedAt = nil

	// Update token map (already holding sm.mu)
	delete(sm.reconnectTokens, oldToken)
	sm.reconnectTokens[newToken] = session.ID

	session.mu.Unlock()
	sm.mu.Unlock()

	if sm.onSessionReconnect != nil {
		sm.onSessionReconnect(session)
	}

	return session, nil
}

// AddPendingMessage adds a message to the pending queue for a disconnected session
func (sm *SessionManager) AddPendingMessage(sessionID string, msgType string, data interface{}) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State != StateDisconnected && session.State != StateReconnecting {
		return ErrSessionNotDisconnected
	}

	session.PendingMessages = append(session.PendingMessages, PendingMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	})

	return nil
}

// GetAndClearPendingMessages retrieves and clears pending messages
func (sm *SessionManager) GetAndClearPendingMessages(sessionID string) ([]PendingMessage, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	messages := session.PendingMessages
	session.PendingMessages = make([]PendingMessage, 0)

	return messages, nil
}

// CloseSession closes a session permanently
func (sm *SessionManager) CloseSession(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	session.State = StateClosed
	reconnectToken := session.ReconnectToken
	userID := session.UserID
	session.mu.Unlock()

	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	delete(sm.reconnectTokens, reconnectToken)

	// Remove from user sessions
	userSessionIDs := sm.userSessions[userID]
	for i, id := range userSessionIDs {
		if id == sessionID {
			sm.userSessions[userID] = append(userSessionIDs[:i], userSessionIDs[i+1:]...)
			break
		}
	}
	if len(sm.userSessions[userID]) == 0 {
		delete(sm.userSessions, userID)
	}
	sm.mu.Unlock()

	if sm.onSessionClosed != nil {
		sm.onSessionClosed(session)
	}

	return nil
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []*RemoteSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*RemoteSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetStats returns session statistics
func (sm *SessionManager) GetStats() SessionStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := SessionStats{
		TotalSessions: len(sm.sessions),
	}

	for _, session := range sm.sessions {
		session.mu.RLock()
		switch session.State {
		case StateActive:
			stats.ActiveSessions++
		case StateDisconnected:
			stats.DisconnectedSessions++
		case StateReconnecting:
			stats.ReconnectingSessions++
		}
		session.mu.RUnlock()
	}

	stats.TotalUsers = len(sm.userSessions)

	return stats
}

// SessionStats holds session statistics
type SessionStats struct {
	TotalSessions         int `json:"total_sessions"`
	ActiveSessions        int `json:"active_sessions"`
	DisconnectedSessions  int `json:"disconnected_sessions"`
	ReconnectingSessions  int `json:"reconnecting_sessions"`
	TotalUsers            int `json:"total_users"`
}

// SetOnSessionCreated sets the callback for session creation
func (sm *SessionManager) SetOnSessionCreated(fn func(*RemoteSession)) {
	sm.onSessionCreated = fn
}

// SetOnSessionClosed sets the callback for session closure
func (sm *SessionManager) SetOnSessionClosed(fn func(*RemoteSession)) {
	sm.onSessionClosed = fn
}

// SetOnSessionDisconnect sets the callback for session disconnection
func (sm *SessionManager) SetOnSessionDisconnect(fn func(*RemoteSession)) {
	sm.onSessionDisconnect = fn
}

// SetOnSessionReconnect sets the callback for session reconnection
func (sm *SessionManager) SetOnSessionReconnect(fn func(*RemoteSession)) {
	sm.onSessionReconnect = fn
}

// SetOnIdleWarning sets the callback for idle warning
func (sm *SessionManager) SetOnIdleWarning(fn func(*RemoteSession)) {
	sm.onIdleWarning = fn
}

// SetOnHeartbeatTimeout sets the callback for heartbeat timeout
func (sm *SessionManager) SetOnHeartbeatTimeout(fn func(*RemoteSession)) {
	sm.onHeartbeatTimeout = fn
}

// cleanupLoop periodically cleans up stale sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanup()
		case <-sm.stopCleanup:
			return
		}
	}
}

// cleanup removes stale sessions and handles timeouts
func (sm *SessionManager) cleanup() {
	now := time.Now()
	sessionsToClose := make([]string, 0)
	sessionsToWarn := make([]*RemoteSession, 0)
	sessionsTimedOut := make([]*RemoteSession, 0)

	sm.mu.RLock()
	for _, session := range sm.sessions {
		session.mu.RLock()
		state := session.State
		lastHeartbeat := session.LastHeartbeat
		lastActivity := session.LastActivity
		reconnectExpiry := session.ReconnectExpiry
		session.mu.RUnlock()

		switch state {
		case StateActive:
			// Check heartbeat timeout
			if now.Sub(lastHeartbeat) > sm.config.HeartbeatTimeout {
				sessionsTimedOut = append(sessionsTimedOut, session)
			}

			// Check idle timeout warning
			idleTime := now.Sub(lastActivity)
			warningTime := sm.config.IdleTimeout - sm.config.IdleWarningBefore
			if idleTime >= warningTime && idleTime < sm.config.IdleTimeout {
				sessionsToWarn = append(sessionsToWarn, session)
			} else if idleTime >= sm.config.IdleTimeout {
				sessionsToClose = append(sessionsToClose, session.ID)
			}

		case StateDisconnected:
			// Check if reconnect token expired
			if now.After(reconnectExpiry) {
				sessionsToClose = append(sessionsToClose, session.ID)
			}
		}
	}
	sm.mu.RUnlock()

	// Handle heartbeat timeouts
	for _, session := range sessionsTimedOut {
		sm.MarkDisconnected(session.ID)
		if sm.onHeartbeatTimeout != nil {
			sm.onHeartbeatTimeout(session)
		}
	}

	// Send idle warnings
	for _, session := range sessionsToWarn {
		if sm.onIdleWarning != nil {
			sm.onIdleWarning(session)
		}
	}

	// Close expired sessions
	for _, sessionID := range sessionsToClose {
		sm.CloseSession(sessionID)
	}
}

// Stop stops the session manager
func (sm *SessionManager) Stop() {
	close(sm.stopCleanup)

	// Collect session IDs first to avoid lock contention
	sm.mu.RLock()
	sessionIDs := make([]string, 0, len(sm.sessions))
	for sessionID := range sm.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}
	sm.mu.RUnlock()

	// Close all sessions without holding the main lock
	for _, sessionID := range sessionIDs {
		sm.CloseSession(sessionID)
	}
}

// Errors
var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionClosed         = errors.New("session is closed")
	ErrInvalidReconnectToken = errors.New("invalid reconnect token")
	ErrReconnectTokenExpired = errors.New("reconnect token expired")
	ErrSessionNotDisconnected = errors.New("session is not disconnected")
)

// Helper functions

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "rs_" + hex.EncodeToString(bytes), nil
}

func generateReconnectToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "rt_" + hex.EncodeToString(bytes), nil
}
