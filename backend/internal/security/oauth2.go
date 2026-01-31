package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuth2Service handles generic OAuth2 authentication for custom providers
type OAuth2Service struct {
	httpClient     *http.Client
	encryptionSvc  *EncryptionService
	stateStore     *oauth2StateStore
	callbackURL    string
}

// oauth2StateStore stores OAuth2 state tokens for CSRF protection
type oauth2StateStore struct {
	states map[string]oauth2StateEntry
	mu     sync.RWMutex
}

type oauth2StateEntry struct {
	createdAt  time.Time
	providerID string
	orgID      string
	userID     string
	redirectTo string
	pkceVerifier string // For PKCE support
}

// NewOAuth2Service creates a new OAuth2 service
func NewOAuth2Service(encryptionSvc *EncryptionService, callbackURL string) *OAuth2Service {
	s := &OAuth2Service{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		encryptionSvc: encryptionSvc,
		stateStore:    newOAuth2StateStore(),
		callbackURL:   callbackURL,
	}
	return s
}

func newOAuth2StateStore() *oauth2StateStore {
	s := &oauth2StateStore{
		states: make(map[string]oauth2StateEntry),
	}
	go s.cleanup()
	return s
}

func (s *oauth2StateStore) set(state string, entry oauth2StateEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry.createdAt = time.Now()
	s.states[state] = entry
}

func (s *oauth2StateStore) get(state string) (oauth2StateEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.states[state]
	if !ok {
		return oauth2StateEntry{}, false
	}

	// Check if expired (10 minute TTL)
	if time.Since(entry.createdAt) > 10*time.Minute {
		delete(s.states, state)
		return oauth2StateEntry{}, false
	}

	// One-time use - delete after retrieval
	delete(s.states, state)
	return entry, true
}

func (s *oauth2StateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, entry := range s.states {
			if now.Sub(entry.createdAt) > 10*time.Minute {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}

// OAuth2AuthorizeOptions contains options for generating an OAuth2 authorization URL
type OAuth2AuthorizeOptions struct {
	ProviderConfig *OAuth2Configuration
	ProviderID     string
	OrganizationID string
	UserID         string
	RedirectTo     string // Where to redirect after auth
	State          string // Optional custom state
}

// OAuth2TokenResponse represents the response from an OAuth2 token endpoint
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// OAuth2UserInfo represents user information from an OAuth2 provider
type OAuth2UserInfo struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Picture   string                 `json:"picture"`
	Groups    []string               `json:"groups,omitempty"`
	Raw       map[string]interface{} `json:"raw"`
}

// GenerateAuthorizationURL creates an OAuth2 authorization URL
func (s *OAuth2Service) GenerateAuthorizationURL(opts OAuth2AuthorizeOptions) (string, error) {
	if opts.ProviderConfig == nil {
		return "", errors.New("oauth2: provider configuration required")
	}

	cfg := opts.ProviderConfig

	// Generate state
	state := opts.State
	if state == "" {
		stateBytes := make([]byte, 16)
		if _, err := rand.Read(stateBytes); err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
		state = hex.EncodeToString(stateBytes)
	}

	// Generate PKCE code verifier if enabled
	var pkceVerifier, pkceChallenge string
	if cfg.UsePKCE {
		verifierBytes := make([]byte, 32)
		if _, err := rand.Read(verifierBytes); err != nil {
			return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
		}
		pkceVerifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

		// Generate code challenge (S256)
		hash := sha256.Sum256([]byte(pkceVerifier))
		pkceChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	// Store state with metadata
	s.stateStore.set(state, oauth2StateEntry{
		providerID:   opts.ProviderID,
		orgID:        opts.OrganizationID,
		userID:       opts.UserID,
		redirectTo:   opts.RedirectTo,
		pkceVerifier: pkceVerifier,
	})

	// Build authorization URL
	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", s.callbackURL)
	params.Set("response_type", getResponseType(cfg.ResponseType))
	params.Set("state", state)

	// Add scopes
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = GetDefaultScopes(SSOProviderTypeOAuth)
	}
	params.Set("scope", strings.Join(scopes, " "))

	// Add PKCE if enabled
	if cfg.UsePKCE {
		params.Set("code_challenge", pkceChallenge)
		params.Set("code_challenge_method", "S256")
	}

	authURL := cfg.AuthorizationURL + "?" + params.Encode()
	return authURL, nil
}

// ValidateCallback validates the OAuth2 callback and returns state information
func (s *OAuth2Service) ValidateCallback(state string) (oauth2StateEntry, error) {
	entry, ok := s.stateStore.get(state)
	if !ok {
		return oauth2StateEntry{}, errors.New("oauth2: invalid or expired state")
	}
	return entry, nil
}

// ExchangeCode exchanges an authorization code for tokens
func (s *OAuth2Service) ExchangeCode(ctx context.Context, cfg *OAuth2Configuration, code string, pkceVerifier string) (*OAuth2TokenResponse, error) {
	if cfg == nil {
		return nil, errors.New("oauth2: provider configuration required")
	}

	// Build token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.callbackURL)

	// Add PKCE verifier if present
	if pkceVerifier != "" {
		data.Set("code_verifier", pkceVerifier)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add client authentication
	authMethod := cfg.TokenAuthMethod
	if authMethod == "" {
		authMethod = "client_secret_basic"
	}

	switch authMethod {
	case "client_secret_basic":
		req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)
	case "client_secret_post":
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user information from the provider's userinfo endpoint
func (s *OAuth2Service) GetUserInfo(ctx context.Context, cfg *OAuth2Configuration, accessToken string) (*OAuth2UserInfo, error) {
	if cfg == nil || cfg.UserInfoURL == "" {
		return nil, errors.New("oauth2: userinfo URL not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cfg.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var rawInfo map[string]interface{}
	if err := json.Unmarshal(body, &rawInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	// Extract user info using claim mappings
	userInfo := s.extractUserInfoFromClaims(rawInfo, cfg.ClaimMappings)
	userInfo.Raw = rawInfo

	return userInfo, nil
}

// extractUserInfoFromClaims extracts user information using claim mappings
func (s *OAuth2Service) extractUserInfoFromClaims(claims map[string]interface{}, mappings UserClaimMappings) *OAuth2UserInfo {
	info := &OAuth2UserInfo{}

	// Helper function to extract string value from claims
	getString := func(path string, defaults ...string) string {
		if path != "" {
			if val, ok := extractNestedValue(claims, path); ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
		}
		// Try defaults
		for _, d := range defaults {
			if val, ok := extractNestedValue(claims, d); ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
		}
		return ""
	}

	// Extract fields
	info.ID = getString(mappings.ID, "sub", "id", "user_id")
	info.Email = getString(mappings.Email, "email")
	info.Name = getString(mappings.Name, "name")
	info.FirstName = getString(mappings.FirstName, "given_name", "first_name")
	info.LastName = getString(mappings.LastName, "family_name", "last_name")
	info.Picture = getString(mappings.Picture, "picture", "avatar_url")

	// Extract groups
	groupsPath := mappings.Groups
	if groupsPath == "" {
		groupsPath = "groups"
	}
	if val, ok := extractNestedValue(claims, groupsPath); ok {
		switch v := val.(type) {
		case []interface{}:
			for _, g := range v {
				if str, ok := g.(string); ok {
					info.Groups = append(info.Groups, str)
				}
			}
		case []string:
			info.Groups = v
		}
	}

	return info
}

// extractNestedValue extracts a value from nested map using dot notation
func extractNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// RevokeToken revokes an OAuth2 token
func (s *OAuth2Service) RevokeToken(ctx context.Context, cfg *OAuth2Configuration, token string) error {
	if cfg == nil || cfg.RevokeURL == "" {
		return errors.New("oauth2: revoke URL not configured")
	}

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.RevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// RefreshToken refreshes an OAuth2 access token
func (s *OAuth2Service) RefreshToken(ctx context.Context, cfg *OAuth2Configuration, refreshToken string) (*OAuth2TokenResponse, error) {
	if cfg == nil {
		return nil, errors.New("oauth2: provider configuration required")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &tokenResp, nil
}

// TestProvider tests connectivity to an OAuth2 provider
func (s *OAuth2Service) TestProvider(ctx context.Context, cfg *OAuth2Configuration) (*SSOTestResult, error) {
	if cfg == nil {
		return nil, errors.New("oauth2: provider configuration required")
	}

	start := time.Now()
	result := &SSOTestResult{
		TestedAt: start,
		Details:  make(map[string]string),
	}

	// Test authorization endpoint is reachable
	req, err := http.NewRequestWithContext(ctx, "HEAD", cfg.AuthorizationURL, nil)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Invalid authorization URL: %v", err)
		result.Latency = time.Since(start)
		return result, nil
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to connect to authorization endpoint: %v", err)
		result.Latency = time.Since(start)
		return result, nil
	}
	resp.Body.Close()
	result.Details["authorization_endpoint"] = "reachable"

	// Test token endpoint is reachable
	req, err = http.NewRequestWithContext(ctx, "HEAD", cfg.TokenURL, nil)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Invalid token URL: %v", err)
		result.Latency = time.Since(start)
		return result, nil
	}

	resp, err = s.httpClient.Do(req)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to connect to token endpoint: %v", err)
		result.Latency = time.Since(start)
		return result, nil
	}
	resp.Body.Close()
	result.Details["token_endpoint"] = "reachable"

	// Test userinfo endpoint if configured
	if cfg.UserInfoURL != "" {
		req, err = http.NewRequestWithContext(ctx, "HEAD", cfg.UserInfoURL, nil)
		if err == nil {
			resp, err = s.httpClient.Do(req)
			if err == nil {
				resp.Body.Close()
				result.Details["userinfo_endpoint"] = "reachable"
			} else {
				result.Details["userinfo_endpoint"] = "unreachable"
			}
		}
	}

	result.Success = true
	result.Message = "OAuth2 provider endpoints are reachable"
	result.Latency = time.Since(start)

	return result, nil
}

// getResponseType returns the response type, defaulting to "code"
func getResponseType(responseType string) string {
	if responseType == "" {
		return "code"
	}
	return responseType
}

// GetCallbackURL returns the configured callback URL
func (s *OAuth2Service) GetCallbackURL() string {
	return s.callbackURL
}
