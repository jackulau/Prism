package security

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/config"
)

// WorkOSService handles WorkOS SSO authentication
type WorkOSService struct {
	apiKey       string
	clientID     string
	redirectURL  string
	stateStore   *ssoStateStore
	httpClient   *http.Client
}

// SSOProfile contains user info from WorkOS SSO
type SSOProfile struct {
	ID             string                 `json:"id"`
	Email          string                 `json:"email"`
	FirstName      string                 `json:"first_name"`
	LastName       string                 `json:"last_name"`
	OrganizationID string                 `json:"organization_id"`
	ConnectionID   string                 `json:"connection_id"`
	ConnectionType string                 `json:"connection_type"`
	IdpID          string                 `json:"idp_id"`
	RawAttributes  map[string]interface{} `json:"raw_attributes"`
}

// AuthorizationOptions contains options for generating an authorization URL
type AuthorizationOptions struct {
	// Organization domain or WorkOS organization ID
	Organization string
	// Connection ID for direct SSO (optional, overrides organization)
	ConnectionID string
	// Redirect URI for the callback
	RedirectURI string
	// State token for CSRF protection (auto-generated if empty)
	State string
}

// ssoStateStore stores SSO state tokens for CSRF protection
type ssoStateStore struct {
	states map[string]ssoStateEntry
	mu     sync.RWMutex
}

type ssoStateEntry struct {
	organization string
	expiresAt    time.Time
}

func newSSOStateStore() *ssoStateStore {
	s := &ssoStateStore{
		states: make(map[string]ssoStateEntry),
	}
	go s.cleanup()
	return s
}

func (s *ssoStateStore) set(state, organization string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = ssoStateEntry{
		organization: organization,
		expiresAt:    time.Now().Add(ttl),
	}
}

func (s *ssoStateStore) get(state string) (string, bool) {
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
	return entry.organization, true
}

func (s *ssoStateStore) cleanup() {
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

// NewWorkOSService creates a new WorkOS service
func NewWorkOSService(cfg *config.Config) *WorkOSService {
	redirectURL := cfg.WorkOSRedirectURL
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("%s/api/v1/auth/sso/callback", cfg.BaseURL)
	}

	return &WorkOSService{
		apiKey:      cfg.WorkOSAPIKey,
		clientID:    cfg.WorkOSClientID,
		redirectURL: redirectURL,
		stateStore:  newSSOStateStore(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if WorkOS is configured
func (s *WorkOSService) IsConfigured() bool {
	return s.apiKey != "" && s.clientID != ""
}

// GenerateState generates a random state token for CSRF protection
func (s *WorkOSService) GenerateState(organization string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(bytes)
	s.stateStore.set(state, organization, 10*time.Minute)
	return state, nil
}

// ValidateState validates a state token and returns the associated organization
func (s *WorkOSService) ValidateState(state string) (string, bool) {
	return s.stateStore.get(state)
}

// GenerateAuthorizationURL creates an SSO authorization URL
func (s *WorkOSService) GenerateAuthorizationURL(opts AuthorizationOptions) (string, error) {
	if !s.IsConfigured() {
		return "", fmt.Errorf("WorkOS is not configured")
	}

	baseURL := "https://api.workos.com/sso/authorize"

	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("response_type", "code")

	// Set redirect URI
	redirectURI := opts.RedirectURI
	if redirectURI == "" {
		redirectURI = s.redirectURL
	}
	params.Set("redirect_uri", redirectURI)

	// Set organization or connection
	if opts.ConnectionID != "" {
		params.Set("connection", opts.ConnectionID)
	} else if opts.Organization != "" {
		// WorkOS accepts either organization ID or domain
		params.Set("organization", opts.Organization)
	} else {
		return "", fmt.Errorf("organization or connection_id is required")
	}

	// Set state for CSRF protection
	state := opts.State
	if state == "" {
		var err error
		state, err = s.GenerateState(opts.Organization)
		if err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
	}
	params.Set("state", state)

	return baseURL + "?" + params.Encode(), nil
}

// HandleCallback processes the SSO callback and returns user profile
func (s *WorkOSService) HandleCallback(code string) (*SSOProfile, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("WorkOS is not configured")
	}

	// Exchange code for profile using WorkOS API
	tokenURL := "https://api.workos.com/sso/token"

	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.apiKey)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Profile     struct {
			ID             string                 `json:"id"`
			Email          string                 `json:"email"`
			FirstName      string                 `json:"first_name"`
			LastName       string                 `json:"last_name"`
			OrganizationID string                 `json:"organization_id"`
			ConnectionID   string                 `json:"connection_id"`
			ConnectionType string                 `json:"connection_type"`
			IdpID          string                 `json:"idp_id"`
			RawAttributes  map[string]interface{} `json:"raw_attributes"`
		} `json:"profile"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s: %s", tokenResp.Error, tokenResp.ErrorDescription)
	}

	return &SSOProfile{
		ID:             tokenResp.Profile.ID,
		Email:          tokenResp.Profile.Email,
		FirstName:      tokenResp.Profile.FirstName,
		LastName:       tokenResp.Profile.LastName,
		OrganizationID: tokenResp.Profile.OrganizationID,
		ConnectionID:   tokenResp.Profile.ConnectionID,
		ConnectionType: tokenResp.Profile.ConnectionType,
		IdpID:          tokenResp.Profile.IdpID,
		RawAttributes:  tokenResp.Profile.RawAttributes,
	}, nil
}

// SSOConnection represents an SSO connection
type SSOConnection struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ConnectionType   string    `json:"connection_type"`
	OrganizationID   string    `json:"organization_id"`
	State            string    `json:"state"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListConnections returns all SSO connections for an organization
func (s *WorkOSService) ListConnections(organizationID string) ([]SSOConnection, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("WorkOS is not configured")
	}

	listURL := fmt.Sprintf("https://api.workos.com/connections?organization_id=%s", url.QueryEscape(organizationID))

	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WorkOS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var listResp struct {
		Data []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			ConnectionType   string `json:"connection_type"`
			OrganizationID   string `json:"organization_id"`
			State            string `json:"state"`
			CreatedAt        string `json:"created_at"`
			UpdatedAt        string `json:"updated_at"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	connections := make([]SSOConnection, len(listResp.Data))
	for i, c := range listResp.Data {
		createdAt, _ := time.Parse(time.RFC3339, c.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, c.UpdatedAt)
		connections[i] = SSOConnection{
			ID:             c.ID,
			Name:           c.Name,
			ConnectionType: c.ConnectionType,
			OrganizationID: c.OrganizationID,
			State:          c.State,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	}

	return connections, nil
}
