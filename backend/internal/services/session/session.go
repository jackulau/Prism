package session

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database/repository"
)

// Info represents session information for API responses
type Info struct {
	ID             string    `json:"id"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	DeviceName     string    `json:"device_name"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsCurrent      bool      `json:"is_current"`
}

// Service manages user sessions with timeout handling
type Service struct {
	repo   *repository.SessionRepository
	config config.SessionConfig

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewService creates a new session service
func NewService(repo *repository.SessionRepository, cfg config.SessionConfig) *Service {
	return &Service{
		repo:   repo,
		config: cfg,
		stopCh: make(chan struct{}),
	}
}

// Start begins the background cleanup goroutine
func (s *Service) Start() {
	s.wg.Add(1)
	go s.cleanupLoop()
	log.Printf("Session service started (idle timeout: %v, cleanup interval: %v, max per user: %d)",
		s.config.IdleTimeout, s.config.CleanupInterval, s.config.MaxPerUser)
}

// Stop stops the background cleanup goroutine
func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Println("Session service stopped")
}

// Create creates a new session with device metadata, enforcing max sessions per user
func (s *Service) Create(userID, refreshTokenHash string, expiresAt time.Time, ipAddress, userAgent string) (*repository.Session, error) {
	deviceName := ParseDeviceName(userAgent)

	// Enforce max sessions per user
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	if count >= s.config.MaxPerUser {
		// Delete oldest sessions to make room
		if err := s.repo.DeleteOldestByUserID(userID, s.config.MaxPerUser-1); err != nil {
			log.Printf("Warning: failed to delete oldest sessions for user %s: %v", userID, err)
		}
	}

	return s.repo.CreateWithMetadata(userID, refreshTokenHash, expiresAt, ipAddress, userAgent, deviceName)
}

// RecordActivity updates the last activity timestamp for a session
func (s *Service) RecordActivity(sessionID string) error {
	return s.repo.UpdateActivity(sessionID)
}

// IsValid checks if a session is valid (not expired, not idle)
func (s *Service) IsValid(sess *repository.Session) bool {
	if sess == nil {
		return false
	}

	now := time.Now()

	// Check if session has expired
	if now.After(sess.ExpiresAt) {
		return false
	}

	// Check idle timeout
	if sess.LastActivityAt != nil {
		idleTime := now.Sub(*sess.LastActivityAt)
		if idleTime > s.config.IdleTimeout {
			return false
		}
	}

	return true
}

// IsSessionIdle checks if a session has exceeded the idle timeout
func (s *Service) IsSessionIdle(sess *repository.Session) bool {
	if sess == nil || sess.LastActivityAt == nil {
		return false
	}

	idleTime := time.Since(*sess.LastActivityAt)
	return idleTime > s.config.IdleTimeout
}

// ListUserSessions returns all sessions for a user with current session marked
func (s *Service) ListUserSessions(userID, currentSessionID string) ([]Info, error) {
	sessions, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	infos := make([]Info, 0, len(sessions))
	for _, sess := range sessions {
		info := Info{
			ID:         sess.ID,
			IPAddress:  sess.IPAddress,
			UserAgent:  sess.UserAgent,
			DeviceName: sess.DeviceName,
			CreatedAt:  sess.CreatedAt,
			ExpiresAt:  sess.ExpiresAt,
			IsCurrent:  sess.ID == currentSessionID,
		}
		if sess.LastActivityAt != nil {
			info.LastActivityAt = *sess.LastActivityAt
		} else {
			info.LastActivityAt = sess.CreatedAt
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// Terminate terminates a specific session for a user
func (s *Service) Terminate(userID, sessionID string) error {
	return s.repo.DeleteByIDAndUserID(sessionID, userID)
}

// TerminateOthers terminates all sessions except the current one
func (s *Service) TerminateOthers(userID, currentSessionID string) error {
	return s.repo.DeleteOthers(userID, currentSessionID)
}

// TerminateAll terminates all sessions for a user
func (s *Service) TerminateAll(userID string) error {
	return s.repo.DeleteByUserID(userID)
}

// GetByTokenHash retrieves a session by its refresh token hash
func (s *Service) GetByTokenHash(tokenHash string) (*repository.Session, error) {
	return s.repo.GetByRefreshTokenHash(tokenHash)
}

// GetByID retrieves a session by its ID
func (s *Service) GetByID(sessionID string) (*repository.Session, error) {
	return s.repo.GetByID(sessionID)
}

// Delete deletes a session by its ID
func (s *Service) Delete(sessionID string) error {
	return s.repo.Delete(sessionID)
}

// UpdateTokenHash updates the refresh token hash for a session
func (s *Service) UpdateTokenHash(sessionID, tokenHash string) error {
	return s.repo.UpdateTokenHash(sessionID, tokenHash)
}

// cleanupLoop runs periodically to clean up expired and idle sessions
func (s *Service) cleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

// performCleanup removes expired and idle sessions
func (s *Service) performCleanup() {
	// Delete expired sessions
	expiredCount, err := s.repo.DeleteExpired()
	if err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
	}

	// Delete idle sessions
	idleCount, err := s.repo.DeleteIdle(s.config.IdleTimeout)
	if err != nil {
		log.Printf("Error cleaning up idle sessions: %v", err)
	}

	if expiredCount > 0 || idleCount > 0 {
		log.Printf("Session cleanup: removed %d expired, %d idle sessions", expiredCount, idleCount)
	}
}

// ParseDeviceName extracts a friendly device name from User-Agent
func ParseDeviceName(userAgent string) string {
	if userAgent == "" {
		return "Unknown Device"
	}

	ua := strings.ToLower(userAgent)

	// Detect OS
	var os string
	switch {
	case strings.Contains(ua, "iphone"):
		os = "iPhone"
	case strings.Contains(ua, "ipad"):
		os = "iPad"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		os = "macOS"
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	case strings.Contains(ua, "chromeos"):
		os = "ChromeOS"
	default:
		os = "Unknown"
	}

	// Detect browser
	var browser string
	switch {
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg/"):
		browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		browser = "Opera"
	case strings.Contains(ua, "brave"):
		browser = "Brave"
	case strings.Contains(ua, "vivaldi"):
		browser = "Vivaldi"
	case strings.Contains(ua, "arc"):
		browser = "Arc"
	default:
		// Try to extract app name for API clients
		if match := regexp.MustCompile(`^([A-Za-z0-9-]+)/`).FindStringSubmatch(userAgent); len(match) > 1 {
			browser = match[1]
		} else {
			browser = "Unknown"
		}
	}

	if os == "Unknown" && browser == "Unknown" {
		return "Unknown Device"
	}

	return browser + " on " + os
}
