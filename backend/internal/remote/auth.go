package remote

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/security"
)

// Errors for authentication
var (
	ErrInvalidCredentials           = errors.New("invalid credentials")
	ErrSessionExpired               = errors.New("session expired")
	ErrSessionNotFound              = errors.New("session not found")
	ErrConnectionLimitExceeded      = errors.New("connection limit exceeded")
	ErrConnectionLimitPerIPExceeded = errors.New("connection limit per IP exceeded")
	ErrRemoteAccessDisabled         = errors.New("remote access is disabled")
	ErrInvalidToken                 = errors.New("invalid token")
)

// RemoteSession represents an authenticated remote session
type RemoteSession struct {
	ID           string    `json:"id"`
	Token        string    `json:"-"` // Don't serialize the token
	TokenHash    string    `json:"-"` // Store hash for validation
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
}

// IsExpired checks if the session has expired
func (s *RemoteSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// RemoteAuthConfig holds configuration for remote authentication
type RemoteAuthConfig struct {
	// Enabled indicates if remote access is enabled
	Enabled bool

	// Password is the hashed password for remote access
	PasswordHash string

	// SessionDuration is how long a session is valid
	SessionDuration time.Duration

	// MaxSessions is the maximum number of concurrent sessions
	MaxSessions int

	// EncryptionService for secure operations
	EncryptionService *security.EncryptionService
}

// DefaultRemoteAuthConfig returns default authentication configuration
func DefaultRemoteAuthConfig() *RemoteAuthConfig {
	return &RemoteAuthConfig{
		Enabled:         false,
		SessionDuration: 24 * time.Hour,
		MaxSessions:     10,
	}
}

// RemoteAuthService handles authentication for remote access
type RemoteAuthService struct {
	config   *RemoteAuthConfig
	sessions map[string]*RemoteSession
	byToken  map[string]*RemoteSession
	mu       sync.RWMutex
}

// NewRemoteAuthService creates a new remote authentication service
func NewRemoteAuthService(cfg *RemoteAuthConfig) *RemoteAuthService {
	if cfg == nil {
		cfg = DefaultRemoteAuthConfig()
	}

	return &RemoteAuthService{
		config:   cfg,
		sessions: make(map[string]*RemoteSession),
		byToken:  make(map[string]*RemoteSession),
	}
}

// IsEnabled returns whether remote access is enabled
func (s *RemoteAuthService) IsEnabled() bool {
	return s.config.Enabled
}

// Authenticate validates credentials and creates a new session
func (s *RemoteAuthService) Authenticate(password, clientIP, userAgent string) (*RemoteSession, string, error) {
	if !s.config.Enabled {
		return nil, "", ErrRemoteAccessDisabled
	}

	// Verify password using constant-time comparison
	if !security.VerifyPassword(password, s.config.PasswordHash) {
		// Add a small delay to prevent timing attacks
		time.Sleep(100 * time.Millisecond)
		return nil, "", ErrInvalidCredentials
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check session limit
	if len(s.sessions) >= s.config.MaxSessions {
		// Remove oldest expired session if any
		s.cleanupExpiredLocked()
		if len(s.sessions) >= s.config.MaxSessions {
			return nil, "", ErrConnectionLimitExceeded
		}
	}

	// Generate session ID and token
	sessionID, err := generateRandomString(16)
	if err != nil {
		return nil, "", err
	}

	token, err := generateRandomString(32)
	if err != nil {
		return nil, "", err
	}

	// Hash the token for storage
	tokenHash := security.HashAPIKey(token)

	now := time.Now()
	session := &RemoteSession{
		ID:           sessionID,
		Token:        token,
		TokenHash:    tokenHash,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		CreatedAt:    now,
		ExpiresAt:    now.Add(s.config.SessionDuration),
		LastActivity: now,
	}

	s.sessions[sessionID] = session
	s.byToken[tokenHash] = session

	return session, token, nil
}

// ValidateToken validates a session token and returns the session
func (s *RemoteAuthService) ValidateToken(token string) (*RemoteSession, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	tokenHash := security.HashAPIKey(token)

	s.mu.RLock()
	session, ok := s.byToken[tokenHash]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		s.InvalidateSession(session.ID)
		return nil, ErrSessionExpired
	}

	// Update last activity
	s.mu.Lock()
	session.LastActivity = time.Now()
	s.mu.Unlock()

	return session, nil
}

// InvalidateSession removes a session
func (s *RemoteAuthService) InvalidateSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	delete(s.sessions, sessionID)
	delete(s.byToken, session.TokenHash)

	return nil
}

// GetSession returns a session by ID
func (s *RemoteAuthService) GetSession(sessionID string) (*RemoteSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// GetActiveSessions returns all active sessions
func (s *RemoteAuthService) GetActiveSessions() []*RemoteSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*RemoteSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		if !session.IsExpired() {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// CleanupExpired removes expired sessions
func (s *RemoteAuthService) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked()
}

// cleanupExpiredLocked removes expired sessions (must hold lock)
func (s *RemoteAuthService) cleanupExpiredLocked() {
	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
			delete(s.byToken, session.TokenHash)
		}
	}
}

// SetPassword sets a new password for remote access
func (s *RemoteAuthService) SetPassword(password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.config.PasswordHash = hash
	s.mu.Unlock()

	return nil
}

// Enable enables remote access
func (s *RemoteAuthService) Enable() {
	s.mu.Lock()
	s.config.Enabled = true
	s.mu.Unlock()
}

// Disable disables remote access and invalidates all sessions
func (s *RemoteAuthService) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Enabled = false
	s.sessions = make(map[string]*RemoteSession)
	s.byToken = make(map[string]*RemoteSession)
}

// SessionCount returns the number of active sessions
func (s *RemoteAuthService) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// ValidateTokenConstantTime performs constant-time token comparison
func ValidateTokenConstantTime(provided, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// generateRandomString generates a random hex string of the given byte length
func generateRandomString(byteLen int) (string, error) {
	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
