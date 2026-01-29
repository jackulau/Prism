package handlers

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// WorkOSHandler handles WorkOS SSO endpoints
type WorkOSHandler struct {
	workosService *security.WorkOSService
	userRepo      *repository.UserRepository
	jwtService    *security.JWTService
	sessionRepo   *repository.SessionRepository
	config        *config.Config
}

// NewWorkOSHandler creates a new WorkOS handler
func NewWorkOSHandler(
	workosService *security.WorkOSService,
	userRepo *repository.UserRepository,
	jwtService *security.JWTService,
	sessionRepo *repository.SessionRepository,
	cfg *config.Config,
) *WorkOSHandler {
	return &WorkOSHandler{
		workosService: workosService,
		userRepo:      userRepo,
		jwtService:    jwtService,
		sessionRepo:   sessionRepo,
		config:        cfg,
	}
}

// AuthorizeRequest represents the SSO authorize request
type AuthorizeRequest struct {
	Organization string `query:"organization"`
	ConnectionID string `query:"connection_id"`
}

// AuthorizeResponse represents the SSO authorize response
type AuthorizeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
}

// Authorize handles SSO authorization requests
// GET /api/v1/auth/sso/authorize
// Query params: organization (domain or org ID), connection_id (optional)
// Returns: { authorization_url: string }
func (h *WorkOSHandler) Authorize(c *fiber.Ctx) error {
	if !h.workosService.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "SSO is not configured",
		})
	}

	organization := c.Query("organization")
	connectionID := c.Query("connection_id")

	if organization == "" && connectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization or connection_id is required",
		})
	}

	authURL, _, err := h.workosService.GenerateAuthorizationURL(security.AuthorizationOptions{
		OrganizationID: organization,
		ConnectionID:   connectionID,
	})
	if err != nil {
		log.Printf("WorkOS SSO authorization failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate authorization URL",
		})
	}

	return c.JSON(AuthorizeResponse{
		AuthorizationURL: authURL,
	})
}

// Callback handles the SSO callback from WorkOS
// GET /api/v1/auth/sso/callback
// Query params: code, state, error, error_description
// Redirects to frontend with tokens or error
func (h *WorkOSHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDesc := c.Query("error_description")

	frontendURL := h.config.FrontendURL

	// Handle error from WorkOS
	if errorParam != "" {
		log.Printf("WorkOS SSO error: %s - %s", errorParam, errorDesc)
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=%s",
			frontendURL, url.QueryEscape(errorDesc)))
	}

	// Validate code
	if code == "" {
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=missing_code", frontendURL))
	}

	// Validate state token (CSRF protection)
	if state != "" {
		_, ok := h.workosService.ValidateState(state)
		if !ok {
			return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=invalid_state", frontendURL))
		}
	}

	// Exchange code for profile
	profile, err := h.workosService.HandleCallback(c.Context(), code)
	if err != nil {
		log.Printf("WorkOS SSO callback failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=authentication_failed", frontendURL))
	}

	// Find or create user
	user, err := h.findOrCreateUser(profile)
	if err != nil {
		log.Printf("WorkOS SSO user creation failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=user_creation_failed", frontendURL))
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		log.Printf("WorkOS SSO token generation failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=token_generation_failed", frontendURL))
	}

	// Create session
	refreshTokenHash := security.HashAPIKey(tokens.RefreshToken)
	_, err = h.sessionRepo.Create(user.ID, refreshTokenHash, time.Now().Add(7*24*time.Hour))
	if err != nil {
		log.Printf("WorkOS SSO session creation failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/login?sso=error&message=session_creation_failed", frontendURL))
	}

	// Set wos-session cookie (optional, for additional session tracking)
	c.Cookie(&fiber.Cookie{
		Name:     "wos-session",
		Value:    tokens.AccessToken,
		Expires:  tokens.ExpiresAt,
		HTTPOnly: true,
		Secure:   h.config.Environment == "production",
		SameSite: "Lax",
	})

	// Redirect to frontend with tokens
	// The frontend will store these tokens and complete the login
	return c.Redirect(fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&expires_at=%d",
		frontendURL,
		url.QueryEscape(tokens.AccessToken),
		url.QueryEscape(tokens.RefreshToken),
		tokens.ExpiresAt.Unix()))
}

// findOrCreateUser finds an existing user by WorkOS ID or email, or creates a new one
func (h *WorkOSHandler) findOrCreateUser(profile *security.SSOProfile) (*repository.User, error) {
	// First, try to find by WorkOS ID
	user, err := h.userRepo.GetByWorkOSID(profile.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by WorkOS ID: %w", err)
	}
	if user != nil {
		return user, nil
	}

	// Try to find by email
	user, err = h.userRepo.GetByEmail(profile.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	if user != nil {
		// Link existing user to WorkOS
		err = h.userRepo.LinkWorkOSAccount(
			user.ID,
			profile.ID,
			profile.OrganizationID,
			profile.ConnectionID,
			profile.ConnectionType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to link WorkOS account: %w", err)
		}
		user.WorkOSID = profile.ID
		user.OrganizationID = profile.OrganizationID
		user.SSOConnectionID = profile.ConnectionID
		user.SSOProvider = profile.ConnectionType
		return user, nil
	}

	// Create new user from SSO profile
	user, err = h.userRepo.CreateFromSSO(
		profile.Email,
		profile.ID,
		profile.OrganizationID,
		profile.ConnectionID,
		profile.ConnectionType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from SSO: %w", err)
	}

	return user, nil
}

// SSOConnectionResponse represents an SSO connection in the API response
type SSOConnectionResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ConnectionType string `json:"connection_type"`
	State          string `json:"state"`
}

// GetConnections returns the SSO connections for the user's organization
// GET /api/v1/auth/sso/connections
// Returns: list of SSO connections
func (h *WorkOSHandler) GetConnections(c *fiber.Ctx) error {
	if !h.workosService.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "SSO is not configured",
		})
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get user to find their organization
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

	// Check if user has an organization
	if user.OrganizationID == "" {
		return c.JSON(fiber.Map{
			"connections": []SSOConnectionResponse{},
		})
	}

	// List connections for the organization
	connections, err := h.workosService.ListConnections(user.OrganizationID)
	if err != nil {
		log.Printf("Failed to list SSO connections: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list connections",
		})
	}

	// Convert to response format
	response := make([]SSOConnectionResponse, len(connections))
	for i, conn := range connections {
		response[i] = SSOConnectionResponse{
			ID:             conn.ID,
			Name:           conn.Name,
			ConnectionType: conn.ConnectionType,
			State:          conn.State,
		}
	}

	return c.JSON(fiber.Map{
		"connections": response,
	})
}

// SSOStatusResponse represents the SSO status for a user
type SSOStatusResponse struct {
	Enabled        bool   `json:"enabled"`
	Connected      bool   `json:"connected"`
	OrganizationID string `json:"organization_id,omitempty"`
	ConnectionType string `json:"connection_type,omitempty"`
}

// GetStatus returns the SSO status for the current user
// GET /api/v1/auth/sso/status
func (h *WorkOSHandler) GetStatus(c *fiber.Ctx) error {
	// Check if SSO is configured at all
	if !h.workosService.IsConfigured() {
		return c.JSON(SSOStatusResponse{
			Enabled:   false,
			Connected: false,
		})
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.JSON(SSOStatusResponse{
			Enabled:   true,
			Connected: false,
		})
	}

	// Get user to check their SSO status
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.JSON(SSOStatusResponse{
			Enabled:   true,
			Connected: false,
		})
	}

	return c.JSON(SSOStatusResponse{
		Enabled:        true,
		Connected:      user.WorkOSID != "",
		OrganizationID: user.OrganizationID,
		ConnectionType: user.SSOProvider,
	})
}
