package remote

import (
	"errors"
	"sync"
	"time"
)

// ReconnectHandler handles reconnection logic for remote sessions
type ReconnectHandler struct {
	sessionManager *SessionManager
	config         *ReconnectConfig
	attempts       map[string]*reconnectAttempt
	mu             sync.RWMutex
}

// ReconnectConfig holds configuration for reconnection handling
type ReconnectConfig struct {
	// Maximum number of reconnection attempts
	MaxAttempts int
	// Base delay between reconnection attempts (exponential backoff)
	BaseDelay time.Duration
	// Maximum delay between reconnection attempts
	MaxDelay time.Duration
	// Token validity period after disconnect
	TokenValidity time.Duration
	// Whether to merge pending messages on reconnect
	MergePendingMessages bool
}

// DefaultReconnectConfig returns the default reconnect configuration
func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxAttempts:          5,
		BaseDelay:            1 * time.Second,
		MaxDelay:             30 * time.Second,
		TokenValidity:        5 * time.Minute,
		MergePendingMessages: true,
	}
}

// reconnectAttempt tracks reconnection attempts for a session
type reconnectAttempt struct {
	sessionID    string
	attemptCount int
	lastAttempt  time.Time
	nextDelay    time.Duration
}

// ReconnectRequest represents a reconnection request
type ReconnectRequest struct {
	ReconnectToken string            `json:"reconnect_token"`
	ClientInfo     map[string]string `json:"client_info,omitempty"`
	LastMessageID  string            `json:"last_message_id,omitempty"`
}

// ReconnectResponse represents a reconnection response
type ReconnectResponse struct {
	Success            bool              `json:"success"`
	SessionID          string            `json:"session_id,omitempty"`
	NewReconnectToken  string            `json:"new_reconnect_token,omitempty"`
	PendingMessages    []PendingMessage  `json:"pending_messages,omitempty"`
	Error              string            `json:"error,omitempty"`
	RetryAfterMs       int64             `json:"retry_after_ms,omitempty"`
	RemainingAttempts  int               `json:"remaining_attempts,omitempty"`
}

// NewReconnectHandler creates a new reconnect handler
func NewReconnectHandler(sessionManager *SessionManager, config *ReconnectConfig) *ReconnectHandler {
	if config == nil {
		config = DefaultReconnectConfig()
	}

	return &ReconnectHandler{
		sessionManager: sessionManager,
		config:         config,
		attempts:       make(map[string]*reconnectAttempt),
	}
}

// HandleReconnect processes a reconnection request
func (rh *ReconnectHandler) HandleReconnect(req *ReconnectRequest) (*ReconnectResponse, error) {
	if req.ReconnectToken == "" {
		return &ReconnectResponse{
			Success: false,
			Error:   "reconnect token is required",
		}, ErrMissingReconnectToken
	}

	// Get session by reconnect token
	session, err := rh.sessionManager.GetSessionByReconnectToken(req.ReconnectToken)
	if err != nil {
		switch err {
		case ErrInvalidReconnectToken:
			return &ReconnectResponse{
				Success: false,
				Error:   "invalid reconnect token",
			}, err
		case ErrReconnectTokenExpired:
			return &ReconnectResponse{
				Success: false,
				Error:   "reconnect token expired, please re-authenticate",
			}, err
		default:
			return &ReconnectResponse{
				Success: false,
				Error:   err.Error(),
			}, err
		}
	}

	// Check attempt limits
	rh.mu.Lock()
	attempt, exists := rh.attempts[session.ID]
	if !exists {
		attempt = &reconnectAttempt{
			sessionID:    session.ID,
			attemptCount: 0,
			nextDelay:    rh.config.BaseDelay,
		}
		rh.attempts[session.ID] = attempt
	}

	attempt.attemptCount++
	attempt.lastAttempt = time.Now()

	if attempt.attemptCount > rh.config.MaxAttempts {
		rh.mu.Unlock()
		// Close the session after too many failed attempts
		rh.sessionManager.CloseSession(session.ID)
		return &ReconnectResponse{
			Success:           false,
			Error:             "maximum reconnection attempts exceeded",
			RemainingAttempts: 0,
		}, ErrMaxReconnectAttemptsExceeded
	}
	rh.mu.Unlock()

	// Perform reconnection
	reconnectedSession, err := rh.sessionManager.Reconnect(req.ReconnectToken)
	if err != nil {
		// Calculate retry delay with exponential backoff
		rh.mu.Lock()
		retryDelay := attempt.nextDelay
		attempt.nextDelay = minDuration(attempt.nextDelay*2, rh.config.MaxDelay)
		remainingAttempts := rh.config.MaxAttempts - attempt.attemptCount
		rh.mu.Unlock()

		return &ReconnectResponse{
			Success:           false,
			Error:             err.Error(),
			RetryAfterMs:      retryDelay.Milliseconds(),
			RemainingAttempts: remainingAttempts,
		}, err
	}

	// Get pending messages if configured
	var pendingMessages []PendingMessage
	if rh.config.MergePendingMessages {
		pendingMessages, _ = rh.sessionManager.GetAndClearPendingMessages(reconnectedSession.ID)

		// Filter messages if last message ID provided
		if req.LastMessageID != "" {
			pendingMessages = filterMessagesAfter(pendingMessages, req.LastMessageID)
		}
	}

	// Update client info if provided
	if req.ClientInfo != nil {
		reconnectedSession.mu.Lock()
		if reconnectedSession.ClientInfo == nil {
			reconnectedSession.ClientInfo = make(map[string]string)
		}
		for k, v := range req.ClientInfo {
			reconnectedSession.ClientInfo[k] = v
		}
		reconnectedSession.mu.Unlock()
	}

	// Clear attempt counter on successful reconnect
	rh.mu.Lock()
	delete(rh.attempts, session.ID)
	rh.mu.Unlock()

	return &ReconnectResponse{
		Success:           true,
		SessionID:         reconnectedSession.ID,
		NewReconnectToken: reconnectedSession.ReconnectToken,
		PendingMessages:   pendingMessages,
	}, nil
}

// GenerateReconnectInfo generates reconnection info for a session
func (rh *ReconnectHandler) GenerateReconnectInfo(sessionID string) (*ReconnectInfo, error) {
	session, err := rh.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return &ReconnectInfo{
		ReconnectToken: session.ReconnectToken,
		ExpiresAt:      session.ReconnectExpiry,
		ValidForMs:     time.Until(session.ReconnectExpiry).Milliseconds(),
	}, nil
}

// ReconnectInfo contains information needed for reconnection
type ReconnectInfo struct {
	ReconnectToken string    `json:"reconnect_token"`
	ExpiresAt      time.Time `json:"expires_at"`
	ValidForMs     int64     `json:"valid_for_ms"`
}

// ClearAttempts clears reconnection attempt tracking for a session
func (rh *ReconnectHandler) ClearAttempts(sessionID string) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	delete(rh.attempts, sessionID)
}

// GetAttemptInfo returns information about reconnection attempts
func (rh *ReconnectHandler) GetAttemptInfo(sessionID string) *AttemptInfo {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	attempt, exists := rh.attempts[sessionID]
	if !exists {
		return &AttemptInfo{
			SessionID:         sessionID,
			AttemptCount:      0,
			RemainingAttempts: rh.config.MaxAttempts,
		}
	}

	return &AttemptInfo{
		SessionID:         sessionID,
		AttemptCount:      attempt.attemptCount,
		LastAttempt:       attempt.lastAttempt,
		NextDelay:         attempt.nextDelay,
		RemainingAttempts: rh.config.MaxAttempts - attempt.attemptCount,
	}
}

// AttemptInfo contains information about reconnection attempts
type AttemptInfo struct {
	SessionID         string        `json:"session_id"`
	AttemptCount      int           `json:"attempt_count"`
	LastAttempt       time.Time     `json:"last_attempt,omitempty"`
	NextDelay         time.Duration `json:"next_delay,omitempty"`
	RemainingAttempts int           `json:"remaining_attempts"`
}

// filterMessagesAfter filters messages to only include those after a given message ID
func filterMessagesAfter(messages []PendingMessage, lastMessageID string) []PendingMessage {
	// In a real implementation, messages would have IDs
	// For now, return all messages
	return messages
}

// minDuration returns the smaller of two durations
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// Errors
var (
	ErrMissingReconnectToken         = errors.New("reconnect token is required")
	ErrMaxReconnectAttemptsExceeded  = errors.New("maximum reconnection attempts exceeded")
)
