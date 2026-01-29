package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"
)

// RemoteAccessConfig holds configuration for remote authentication
type RemoteAccessConfig struct {
	Enabled             bool
	PasswordHash        string        // Argon2id hash of the remote access password
	SessionTimeout      time.Duration // How long sessions remain valid
	MaxConcurrentSessions int         // Maximum number of concurrent sessions
	MaxFailedAttempts   int           // Failed attempts before IP is blocked
	BlockDuration       time.Duration // Initial block duration
	MaxBlockDuration    time.Duration // Maximum block duration (exponential backoff cap)
}

// DefaultRemoteAccessConfig returns a config with secure defaults
func DefaultRemoteAccessConfig() *RemoteAccessConfig {
	return &RemoteAccessConfig{
		Enabled:             false,
		PasswordHash:        "",
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:   5,
		BlockDuration:       15 * time.Minute,
		MaxBlockDuration:    24 * time.Hour,
	}
}

// RemoteSession represents an authenticated remote session
type RemoteSession struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	ClientIP  string
	IsActive  bool
}

// failedLoginInfo tracks failed login attempts for an IP
type failedLoginInfo struct {
	Count       int
	LastAttempt time.Time
	BlockUntil  time.Time
	BlockCount  int // Number of times this IP has been blocked (for exponential backoff)
}

// RemoteAuthService handles authentication for remote connections
type RemoteAuthService struct {
	config       *RemoteAccessConfig
	crypto       *EncryptionService
	sessions     map[string]*RemoteSession // token -> session
	failedLogins map[string]*failedLoginInfo // ip -> info
	mu           sync.RWMutex
	stopCleanup  chan struct{}
}

// NewRemoteAuthService creates a new remote authentication service
func NewRemoteAuthService(config *RemoteAccessConfig, crypto *EncryptionService) *RemoteAuthService {
	if config == nil {
		config = DefaultRemoteAccessConfig()
	}

	s := &RemoteAuthService{
		config:       config,
		crypto:       crypto,
		sessions:     make(map[string]*RemoteSession),
		failedLogins: make(map[string]*failedLoginInfo),
		stopCleanup:  make(chan struct{}),
	}

	// Start background cleanup goroutine
	go s.cleanupLoop()

	return s
}

// Stop stops the background cleanup goroutine
func (s *RemoteAuthService) Stop() {
	close(s.stopCleanup)
}

// cleanupLoop periodically cleans up expired sessions and stale failed login records
func (s *RemoteAuthService) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CleanupExpiredSessions()
			s.cleanupExpiredBlocks()
		case <-s.stopCleanup:
			return
		}
	}
}

// IsEnabled returns whether remote access is enabled
func (s *RemoteAuthService) IsEnabled() bool {
	return s.config.Enabled && s.config.PasswordHash != ""
}

// Authenticate validates a password and creates a session if successful
func (s *RemoteAuthService) Authenticate(password string, clientIP string) (*RemoteSession, error) {
	if !s.IsEnabled() {
		return nil, errors.New("remote access is not enabled")
	}

	// Check if IP is blocked
	if s.IsIPBlocked(clientIP) {
		s.logAuthAttempt(clientIP, false, "IP is blocked")
		return nil, errors.New("too many failed attempts, please try again later")
	}

	// Verify password
	if !VerifyPassword(password, s.config.PasswordHash) {
		s.recordFailedAttempt(clientIP)
		s.logAuthAttempt(clientIP, false, "invalid password")
		return nil, errors.New("invalid password")
	}

	// Password valid - clear any failed attempts for this IP
	s.clearFailedAttempts(clientIP)

	// Check concurrent session limit
	s.mu.Lock()
	defer s.mu.Unlock()

	activeCount := 0
	for _, session := range s.sessions {
		if session.IsActive && time.Now().Before(session.ExpiresAt) {
			activeCount++
		}
	}

	if activeCount >= s.config.MaxConcurrentSessions {
		s.logAuthAttempt(clientIP, false, "max concurrent sessions reached")
		return nil, errors.New("maximum concurrent sessions reached")
	}

	// Generate secure session token
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, errors.New("failed to generate session token")
	}

	now := time.Now()
	session := &RemoteSession{
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(s.config.SessionTimeout),
		ClientIP:  clientIP,
		IsActive:  true,
	}

	s.sessions[token] = session
	s.logAuthAttempt(clientIP, true, "session created")

	return session, nil
}

// ValidateSession validates a session token and returns the session if valid
func (s *RemoteAuthService) ValidateSession(token string) (*RemoteSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[token]
	if !exists {
		return nil, errors.New("session not found")
	}

	if !session.IsActive {
		return nil, errors.New("session has been invalidated")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}

	return session, nil
}

// ValidateSessionWithIP validates a session and checks that the client IP matches
func (s *RemoteAuthService) ValidateSessionWithIP(token string, clientIP string) (*RemoteSession, error) {
	session, err := s.ValidateSession(token)
	if err != nil {
		return nil, err
	}

	// Use constant-time comparison for IP to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(session.ClientIP), []byte(clientIP)) != 1 {
		return nil, errors.New("session IP mismatch")
	}

	return session, nil
}

// InvalidateSession invalidates a session by token
func (s *RemoteAuthService) InvalidateSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[token]
	if !exists {
		return errors.New("session not found")
	}

	session.IsActive = false
	log.Printf("[REMOTE AUTH] Session invalidated for IP %s", session.ClientIP)

	return nil
}

// CleanupExpiredSessions removes expired sessions from memory
func (s *RemoteAuthService) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) || !session.IsActive {
			delete(s.sessions, token)
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Printf("[REMOTE AUTH] Cleaned up %d expired sessions", cleaned)
	}
}

// cleanupExpiredBlocks removes stale failed login records
func (s *RemoteAuthService) cleanupExpiredBlocks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for ip, info := range s.failedLogins {
		// Remove entries where block has expired and no recent attempts
		if now.After(info.BlockUntil) && now.Sub(info.LastAttempt) > s.config.MaxBlockDuration {
			delete(s.failedLogins, ip)
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Printf("[REMOTE AUTH] Cleaned up %d stale failed login records", cleaned)
	}
}

// IsIPBlocked checks if an IP is currently blocked due to failed attempts
func (s *RemoteAuthService) IsIPBlocked(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.failedLogins[ip]
	if !exists {
		return false
	}

	return time.Now().Before(info.BlockUntil)
}

// GetBlockTimeRemaining returns how long until an IP is unblocked
func (s *RemoteAuthService) GetBlockTimeRemaining(ip string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.failedLogins[ip]
	if !exists {
		return 0
	}

	remaining := time.Until(info.BlockUntil)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// recordFailedAttempt records a failed login attempt for an IP
func (s *RemoteAuthService) recordFailedAttempt(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	info, exists := s.failedLogins[ip]

	if !exists {
		info = &failedLoginInfo{
			Count:       0,
			LastAttempt: now,
			BlockCount:  0,
		}
		s.failedLogins[ip] = info
	}

	info.Count++
	info.LastAttempt = now

	// Check if we need to block this IP
	if info.Count >= s.config.MaxFailedAttempts {
		// Calculate block duration with exponential backoff
		blockDuration := s.config.BlockDuration
		for i := 0; i < info.BlockCount; i++ {
			blockDuration *= 2
			if blockDuration > s.config.MaxBlockDuration {
				blockDuration = s.config.MaxBlockDuration
				break
			}
		}

		info.BlockUntil = now.Add(blockDuration)
		info.BlockCount++
		info.Count = 0 // Reset count for next block period

		log.Printf("[REMOTE AUTH] IP %s blocked for %v (block #%d)", ip, blockDuration, info.BlockCount)
	}
}

// clearFailedAttempts clears failed login attempts for an IP after successful auth
func (s *RemoteAuthService) clearFailedAttempts(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't completely delete - keep the block count for persistent bad actors
	if info, exists := s.failedLogins[ip]; exists {
		info.Count = 0
		// Gradually reduce block count over time (successful auths reduce severity)
		if info.BlockCount > 0 {
			info.BlockCount--
		}
	}
}

// logAuthAttempt logs an authentication attempt for audit purposes
func (s *RemoteAuthService) logAuthAttempt(ip string, success bool, message string) {
	status := "FAILED"
	if success {
		status = "SUCCESS"
	}
	log.Printf("[REMOTE AUTH] %s - IP: %s - %s", status, ip, message)
}

// GetActiveSessions returns the count of active sessions
func (s *RemoteAuthService) GetActiveSessions() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, session := range s.sessions {
		if session.IsActive && now.Before(session.ExpiresAt) {
			count++
		}
	}
	return count
}

// RefreshSession extends the expiry of an existing session
func (s *RemoteAuthService) RefreshSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[token]
	if !exists {
		return errors.New("session not found")
	}

	if !session.IsActive {
		return errors.New("session has been invalidated")
	}

	if time.Now().After(session.ExpiresAt) {
		return errors.New("session has expired")
	}

	session.ExpiresAt = time.Now().Add(s.config.SessionTimeout)
	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
