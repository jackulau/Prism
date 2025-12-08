package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// OAuthHandler handles OAuth-related endpoints
type OAuthHandler struct {
	userRepo      *repository.UserRepository
	encryptionSvc *security.EncryptionService
	config        *config.Config
	stateStore    *stateStore
	httpClient    *http.Client
}

// stateStore stores OAuth state tokens for CSRF protection
type stateStore struct {
	states map[string]stateEntry
	mu     sync.RWMutex
}

type stateEntry struct {
	userID    string
	expiresAt time.Time
}

func newStateStore() *stateStore {
	s := &stateStore{
		states: make(map[string]stateEntry),
	}
	// Clean up expired states periodically
	go s.cleanup()
	return s
}

func (s *stateStore) set(state, userID string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateEntry{
		userID:    userID,
		expiresAt: time.Now().Add(ttl),
	}
}

func (s *stateStore) get(state string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.states[state]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.states, state)
		return "", false
	}
	delete(s.states, state) // One-time use
	return entry.userID, true
}

func (s *stateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, entry := range s.states {
			if now.After(entry.expiresAt) {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	userRepo *repository.UserRepository,
	encryptionSvc *security.EncryptionService,
	cfg *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:      userRepo,
		encryptionSvc: encryptionSvc,
		config:        cfg,
		stateStore:    newStateStore(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// generateState generates a random state token for CSRF protection
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GitHubAuthorize returns the GitHub OAuth authorization URL
func (h *OAuthHandler) GitHubAuthorize(c *fiber.Ctx) error {
	if h.config.GitHubClientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub OAuth not configured",
		})
	}

	userID := c.Locals("userID").(string)

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate state",
		})
	}

	// Store state with user ID (expires in 10 minutes)
	h.stateStore.set(state, userID, 10*time.Minute)

	// Build authorization URL
	redirectURL := h.config.GitHubRedirectURL
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("%s/api/v1/oauth/github/callback", h.config.BaseURL)
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		h.config.GitHubClientID,
		url.QueryEscape(redirectURL),
		url.QueryEscape("repo read:user user:email"),
		state,
	)

	return c.JSON(fiber.Map{
		"url": authURL,
	})
}

// GitHubCallback handles the OAuth callback from GitHub
func (h *OAuthHandler) GitHubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDesc := c.Query("error_description")

	if errorParam != "" {
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=%s", h.config.FrontendURL, url.QueryEscape(errorDesc)))
	}

	if code == "" {
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=missing_code", h.config.FrontendURL))
	}

	// Verify state
	userID, ok := h.stateStore.get(state)
	if !ok {
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=invalid_state", h.config.FrontendURL))
	}

	// Exchange code for access token
	token, err := h.exchangeGitHubCode(code)
	if err != nil {
		log.Printf("GitHub OAuth token exchange failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=token_exchange_failed", h.config.FrontendURL))
	}

	// Get user info from GitHub
	ghUser, err := h.getGitHubUser(token)
	if err != nil {
		log.Printf("GitHub OAuth user fetch failed: %v", err)
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=user_fetch_failed", h.config.FrontendURL))
	}

	// Encrypt the token
	encryptedToken, nonce, err := h.encryptionSvc.Encrypt([]byte(token))
	if err != nil {
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=encryption_failed", h.config.FrontendURL))
	}

	// Encode nonce and ciphertext as hex, separated by ':'
	encryptedTokenStr := hex.EncodeToString(nonce) + ":" + hex.EncodeToString(encryptedToken)

	// Save to database
	if err := h.userRepo.SaveGitHubConnection(userID, encryptedTokenStr, ghUser.Login); err != nil {
		return c.Redirect(fmt.Sprintf("%s/settings?github=error&message=save_failed", h.config.FrontendURL))
	}

	return c.Redirect(fmt.Sprintf("%s/settings?github=connected", h.config.FrontendURL))
}

// GitHubStatus returns the GitHub connection status
func (h *OAuthHandler) GitHubStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

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

	connected := user.GitHubUsername != ""

	return c.JSON(fiber.Map{
		"connected":    connected,
		"username":     user.GitHubUsername,
		"connected_at": user.GitHubConnectedAt,
	})
}

// DisconnectGitHub removes the GitHub connection
func (h *OAuthHandler) DisconnectGitHub(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	if err := h.userRepo.RemoveGitHubConnection(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to disconnect",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// ListGitHubRepos lists the user's GitHub repositories
func (h *OAuthHandler) ListGitHubRepos(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	if user.GitHubToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub not connected",
		})
	}

	// Decrypt token (stored as "nonce_hex:ciphertext_hex")
	token, err := h.decryptGitHubToken(user.GitHubToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to decrypt token",
		})
	}

	// Fetch repos from GitHub
	repos, err := h.fetchGitHubRepos(token)
	if err != nil {
		log.Printf("Failed to fetch GitHub repos: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to fetch repositories from GitHub",
		})
	}

	return c.JSON(fiber.Map{
		"repos": repos,
	})
}

// exchangeGitHubCode exchanges an authorization code for an access token
func (h *OAuthHandler) exchangeGitHubCode(code string) (string, error) {
	data := url.Values{
		"client_id":     {h.config.GitHubClientID},
		"client_secret": {h.config.GitHubClientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("%s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return tokenResp.AccessToken, nil
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// getGitHubUser fetches the authenticated user's info from GitHub
func (h *OAuthHandler) getGitHubUser(token string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub API user fetch error (%d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("GitHub API error (status %d)", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	UpdatedAt   string `json:"updated_at"`
}

// decryptGitHubToken decrypts a GitHub token stored as "nonce_hex:ciphertext_hex"
func (h *OAuthHandler) decryptGitHubToken(encryptedToken string) (string, error) {
	// Split into nonce and ciphertext
	parts := make([]string, 2)
	colonIdx := -1
	for i, c := range encryptedToken {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return "", fmt.Errorf("invalid encrypted token format")
	}
	parts[0] = encryptedToken[:colonIdx]
	parts[1] = encryptedToken[colonIdx+1:]

	// Decode hex
	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}
	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt
	plaintext, err := h.encryptionSvc.Decrypt(ciphertext, nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// fetchGitHubRepos fetches the user's repositories from GitHub
func (h *OAuthHandler) fetchGitHubRepos(token string) ([]GitHubRepo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/repos?sort=updated&per_page=100", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("GitHub API repos fetch error (%d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("GitHub API error (status %d)", resp.StatusCode)
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}

	return repos, nil
}
