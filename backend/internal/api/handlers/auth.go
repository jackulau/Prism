package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtService  *security.JWTService
	config      *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, jwtService *security.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
		config:      nil, // Will use defaults
	}
}

// NewAuthHandlerWithConfig creates a new auth handler with configuration
func NewAuthHandlerWithConfig(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, jwtService *security.JWTService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
		config:      cfg,
	}
}

// getRefreshTokenTTL returns the configured refresh token TTL or default
func (h *AuthHandler) getRefreshTokenTTL() time.Duration {
	if h.config != nil && h.config.Auth.RefreshTokenTTL > 0 {
		return h.config.Auth.RefreshTokenTTL
	}
	return 7 * 24 * time.Hour // Default: 7 days
}

// getMaxSessions returns the configured max sessions or default
func (h *AuthHandler) getMaxSessions() int {
	if h.config != nil && h.config.Auth.MaxSessions > 0 {
		return h.config.Auth.MaxSessions
	}
	return 10 // Default: 10 sessions per user
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !isValidEmail(req.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid email format",
		})
	}

	// Validate password
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters",
		})
	}

	// Check if email already exists
	exists, err := h.userRepo.EmailExists(req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check email",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "email already registered",
		})
	}

	// Hash password
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	// Create user
	user, err := h.userRepo.Create(req.Email, passwordHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Extract device info and IP address
	deviceInfo := c.Get("User-Agent")
	ipAddress := c.IP()

	// Create session with device info
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	_, err = h.sessionRepo.CreateSession(user.ID, refreshTokenHash, deviceInfo, ipAddress, time.Now().Add(h.getRefreshTokenTTL()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user by email
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	}

	// Verify password
	if !security.VerifyPassword(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	}

	// Check max sessions limit
	maxSessions := h.getMaxSessions()
	if maxSessions > 0 {
		sessionCount, err := h.sessionRepo.CountUserSessions(user.ID)
		if err == nil && sessionCount >= maxSessions {
			// Revoke oldest sessions to make room
			sessions, _ := h.sessionRepo.GetUserSessions(user.ID)
			if len(sessions) > 0 {
				// Revoke the oldest session (last in the list, sorted by last_used_at DESC)
				h.sessionRepo.RevokeSession(sessions[len(sessions)-1].ID)
			}
		}
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Extract device info and IP address
	deviceInfo := c.Get("User-Agent")
	ipAddress := c.IP()

	// Create session with device info
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	_, err = h.sessionRepo.CreateSession(user.ID, refreshTokenHash, deviceInfo, ipAddress, time.Now().Add(h.getRefreshTokenTTL()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.JSON(AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Delete all sessions for user
	if err := h.sessionRepo.DeleteByUserID(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

// Refresh handles token refresh with token rotation (single-use refresh tokens)
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate refresh token
	claims, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}

	// Check if session exists and is not revoked
	refreshTokenHash := security.HashAPIKey(req.RefreshToken)
	session, err := h.sessionRepo.GetByRefreshTokenHash(refreshTokenHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify session",
		})
	}
	if session == nil {
		// Token was already used or session was revoked (potential replay attack)
		// Revoke all user sessions as a security measure
		h.sessionRepo.RevokeAllUserSessions(claims.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "session not found or already used",
		})
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		h.sessionRepo.RevokeSession(session.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "session expired",
		})
	}

	// Generate new tokens (token rotation - invalidates old token)
	tokens, err := h.jwtService.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Update session with new refresh token hash (token rotation)
	newRefreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	newExpiresAt := time.Now().Add(h.getRefreshTokenTTL())
	if err := h.sessionRepo.UpdateRefreshTokenHash(session.ID, newRefreshTokenHash, newExpiresAt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
		})
	}

	// Get user for response
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	return c.JSON(AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Me returns the current user
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// isValidEmail validates an email address
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// GuestLogin handles guest login - creates a temporary guest account
func (h *AuthHandler) GuestLogin(c *fiber.Ctx) error {
	// Generate a unique guest identifier
	guestID := uuid.New().String()[:8]
	guestEmail := fmt.Sprintf("guest-%s@prism.local", guestID)
	guestPassword := uuid.New().String() // Random password, user won't need it

	// Hash password
	passwordHash, err := security.HashPassword(guestPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create guest account",
		})
	}

	// Create guest user
	user, err := h.userRepo.Create(guestEmail, passwordHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create guest account",
		})
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Extract device info and IP address
	deviceInfo := c.Get("User-Agent")
	ipAddress := c.IP()

	// Create session with device info
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	_, err = h.sessionRepo.CreateSession(user.ID, refreshTokenHash, deviceInfo, ipAddress, time.Now().Add(h.getRefreshTokenTTL()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.JSON(AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// SessionDTO represents a session data transfer object
type SessionDTO struct {
	ID         string    `json:"id"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsCurrent  bool      `json:"is_current"`
}

// ListSessions returns all active sessions for the current user
func (h *AuthHandler) ListSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get current session from the refresh token if available
	currentSessionID := c.Locals("sessionID")

	sessions, err := h.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get sessions",
		})
	}

	sessionDTOs := make([]SessionDTO, len(sessions))
	for i, s := range sessions {
		isCurrent := false
		if currentSessionID != nil {
			isCurrent = s.ID == currentSessionID.(string)
		}
		sessionDTOs[i] = SessionDTO{
			ID:         s.ID,
			DeviceInfo: s.DeviceInfo,
			IPAddress:  s.IPAddress,
			CreatedAt:  s.CreatedAt,
			LastUsedAt: s.LastUsedAt,
			ExpiresAt:  s.ExpiresAt,
			IsCurrent:  isCurrent,
		}
	}

	return c.JSON(fiber.Map{
		"sessions": sessionDTOs,
		"count":    len(sessionDTOs),
	})
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session id is required",
		})
	}

	// Verify the session belongs to the user
	session, err := h.sessionRepo.GetSessionByID(sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get session",
		})
	}
	if session == nil || session.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "session not found",
		})
	}

	// Revoke the session
	if err := h.sessionRepo.RevokeSession(sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to revoke session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions for the current user (logout everywhere)
func (h *AuthHandler) RevokeAllSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Check if user wants to keep the current session
	keepCurrent := c.Query("keep_current") == "true"
	currentSessionID := c.Locals("sessionID")

	if keepCurrent && currentSessionID != nil {
		// Revoke all sessions except the current one
		if err := h.sessionRepo.RevokeOtherSessions(userID, currentSessionID.(string)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to revoke sessions",
			})
		}
	} else {
		// Revoke all sessions
		if err := h.sessionRepo.RevokeAllUserSessions(userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to revoke sessions",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "all sessions revoked successfully",
	})
}
