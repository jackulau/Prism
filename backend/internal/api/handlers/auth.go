package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
	"github.com/jacklau/prism/internal/services/session"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo       *repository.UserRepository
	sessionService *session.Service
	jwtService     *security.JWTService
	refreshExpiry  time.Duration
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, sessionService *session.Service, jwtService *security.JWTService, refreshExpiry time.Duration) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		sessionService: sessionService,
		jwtService:     jwtService,
		refreshExpiry:  refreshExpiry,
	}
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

	// Create session with metadata first to get session ID
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	// Use a placeholder token hash initially, we'll update it after generating the real tokens
	session, err := h.sessionService.Create(user.ID, "placeholder", time.Now().Add(h.refreshExpiry), ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Generate tokens with session ID
	tokens, err := h.jwtService.GenerateTokenPairWithSession(user.ID, user.Email, session.ID)
	if err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Update session with actual refresh token hash
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	if err := h.sessionService.UpdateTokenHash(session.ID, refreshTokenHash); err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
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

	// Create session with metadata first to get session ID
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	session, err := h.sessionService.Create(user.ID, "placeholder", time.Now().Add(h.refreshExpiry), ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Generate tokens with session ID
	tokens, err := h.jwtService.GenerateTokenPairWithSession(user.ID, user.Email, session.ID)
	if err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Update session with actual refresh token hash
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	if err := h.sessionService.UpdateTokenHash(session.ID, refreshTokenHash); err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
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

	// Delete the current session only (not all sessions)
	sessionID := middleware.GetSessionID(c)
	if sessionID != "" {
		if err := h.sessionService.Terminate(userID, sessionID); err != nil {
			// Log but don't fail - session might already be deleted
		}
	} else {
		// Fallback: delete all sessions for user if no session ID
		if err := h.sessionService.TerminateAll(userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to logout",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

// RefreshResponse extends AuthResponse with session expiration info
type RefreshResponse struct {
	AuthResponse
	SessionExpired bool `json:"session_expired,omitempty"`
}

// Refresh handles token refresh
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

	// Check if session exists
	refreshTokenHash := security.HashAPIKey(req.RefreshToken)
	session, err := h.sessionService.GetByTokenHash(refreshTokenHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify session",
		})
	}
	if session == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "session not found",
		})
	}

	// Check if session is idle (exceeded idle timeout)
	if h.sessionService.IsSessionIdle(session) {
		// Delete the idle session
		_ = h.sessionService.Delete(session.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":           "session expired due to inactivity",
			"session_expired": true,
		})
	}

	// Check if session is valid (not expired)
	if !h.sessionService.IsValid(session) {
		_ = h.sessionService.Delete(session.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":           "session expired",
			"session_expired": true,
		})
	}

	// Delete old session
	if err := h.sessionService.Delete(session.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to invalidate old session",
		})
	}

	// Create new session first to get session ID
	ipAddress := c.IP()
	userAgent := session.UserAgent // Keep original user agent
	if userAgent == "" {
		userAgent = c.Get("User-Agent")
	}
	newSession, err := h.sessionService.Create(claims.UserID, "placeholder", time.Now().Add(h.refreshExpiry), ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Generate new tokens with session ID
	tokens, err := h.jwtService.GenerateTokenPairWithSession(claims.UserID, claims.Email, newSession.ID)
	if err != nil {
		_ = h.sessionService.Terminate(claims.UserID, newSession.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Update session with actual refresh token hash
	newRefreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	if err := h.sessionService.UpdateTokenHash(newSession.ID, newRefreshTokenHash); err != nil {
		_ = h.sessionService.Terminate(claims.UserID, newSession.ID)
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

	// Create session with metadata first to get session ID
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	session, err := h.sessionService.Create(user.ID, "placeholder", time.Now().Add(h.refreshExpiry), ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	// Generate tokens with session ID
	tokens, err := h.jwtService.GenerateTokenPairWithSession(user.ID, user.Email, session.ID)
	if err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Update session with actual refresh token hash
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	if err := h.sessionService.UpdateTokenHash(session.ID, refreshTokenHash); err != nil {
		_ = h.sessionService.Terminate(user.ID, session.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update session",
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
