package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/sso"
)

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

// AuthorizationOptions contains options for generating an SSO authorization URL
type AuthorizationOptions struct {
	// OrganizationID is the WorkOS organization ID (use this OR ConnectionID)
	OrganizationID string
	// ConnectionID is the WorkOS connection ID (use this OR OrganizationID)
	ConnectionID string
	// RedirectURI overrides the default redirect URI
	RedirectURI string
	// State is an optional state parameter (if empty, one will be generated)
	State string
	// DomainHint is the email domain for the user
	DomainHint string
	// LoginHint is the email address for the user
	LoginHint string
}

// SessionData contains the data stored in the wos-session cookie
type SessionData struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	ConnectionID   string    `json:"connection_id"`
	Email          string    `json:"email"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// stateEntry stores state token data
type stateEntry struct {
	createdAt time.Time
	data      string // optional associated data
}

// stateStore stores SSO state tokens for CSRF protection
type stateStore struct {
	states map[string]stateEntry
	mu     sync.RWMutex
}

// WorkOSService handles WorkOS SSO authentication
type WorkOSService struct {
	apiKey         string
	clientID       string
	redirectURI    string
	cookiePassword string
	configured     bool
	stateStore     *stateStore
}

// newStateStore creates a new state store with background cleanup
func newStateStore() *stateStore {
	s := &stateStore{
		states: make(map[string]stateEntry),
	}
	go s.cleanup()
	return s
}

// set stores a state token with associated data
func (s *stateStore) set(state, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateEntry{
		createdAt: time.Now(),
		data:      data,
	}
}

// get retrieves and deletes a state token (one-time use)
func (s *stateStore) get(state string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.states[state]
	if !ok {
		return "", false
	}

	// Check if expired (10 minute TTL)
	if time.Since(entry.createdAt) > 10*time.Minute {
		delete(s.states, state)
		return "", false
	}

	// One-time use - delete after retrieval
	delete(s.states, state)
	return entry.data, true
}

// cleanup removes expired states periodically
func (s *stateStore) cleanup() {
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

// NewWorkOSService creates a new WorkOS service instance
func NewWorkOSService(apiKey, clientID, redirectURI, cookiePassword string) *WorkOSService {
	configured := apiKey != "" && clientID != "" && cookiePassword != ""

	if configured {
		// Initialize the WorkOS SSO client with API key and client ID
		sso.Configure(apiKey, clientID)
	}

	return &WorkOSService{
		apiKey:         apiKey,
		clientID:       clientID,
		redirectURI:    redirectURI,
		cookiePassword: cookiePassword,
		configured:     configured,
		stateStore:     newStateStore(),
	}
}

// IsConfigured returns whether WorkOS SSO is fully configured
func (s *WorkOSService) IsConfigured() bool {
	return s.configured
}

// HealthCheck validates that the WorkOS service is properly configured
func (s *WorkOSService) HealthCheck() error {
	if !s.configured {
		return errors.New("workos: not configured - missing WORKOS_API_KEY, WORKOS_CLIENT_ID, or WORKOS_COOKIE_PASSWORD")
	}
	return nil
}

// GetClientID returns the WorkOS client ID
func (s *WorkOSService) GetClientID() string {
	return s.clientID
}

// GetRedirectURI returns the configured redirect URI
func (s *WorkOSService) GetRedirectURI() string {
	return s.redirectURI
}

// GetCookiePassword returns the cookie encryption password
func (s *WorkOSService) GetCookiePassword() string {
	return s.cookiePassword
}

// GenerateState generates a random state token for CSRF protection
func (s *WorkOSService) GenerateState(data string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	state := hex.EncodeToString(bytes)
	s.stateStore.set(state, data)
	return state, nil
}

// ValidateState validates a state token and returns associated data
func (s *WorkOSService) ValidateState(state string) (string, bool) {
	return s.stateStore.get(state)
}

// GenerateAuthorizationURL creates an SSO authorization URL
func (s *WorkOSService) GenerateAuthorizationURL(opts AuthorizationOptions) (string, string, error) {
	if !s.configured {
		return "", "", errors.New("workos: not configured")
	}

	// Use provided redirect URI or default
	redirectURI := opts.RedirectURI
	if redirectURI == "" {
		redirectURI = s.redirectURI
	}

	// Generate state if not provided
	state := opts.State
	if state == "" {
		var err error
		state, err = s.GenerateState("")
		if err != nil {
			return "", "", fmt.Errorf("failed to generate state: %w", err)
		}
	} else {
		// Store provided state
		s.stateStore.set(state, "")
	}

	// Build authorization URL options
	authOpts := sso.GetAuthorizationURLOpts{
		RedirectURI: redirectURI,
		State:       state,
	}

	// Set organization or connection
	if opts.OrganizationID != "" {
		authOpts.Organization = opts.OrganizationID
	} else if opts.ConnectionID != "" {
		authOpts.Connection = opts.ConnectionID
	}

	// Set hints if provided
	if opts.DomainHint != "" {
		authOpts.DomainHint = opts.DomainHint
	}
	if opts.LoginHint != "" {
		authOpts.LoginHint = opts.LoginHint
	}

	url, err := sso.GetAuthorizationURL(authOpts)
	if err != nil {
		return "", "", fmt.Errorf("failed to get authorization URL: %w", err)
	}

	return url.String(), state, nil
}

// HandleCallback processes the SSO callback and returns user profile
func (s *WorkOSService) HandleCallback(ctx context.Context, code string) (*SSOProfile, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	if code == "" {
		return nil, errors.New("workos: missing authorization code")
	}

	// Exchange code for profile
	profileResponse, err := sso.GetProfileAndToken(ctx, sso.GetProfileAndTokenOpts{
		Code: code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Map WorkOS profile to our SSOProfile
	profile := &SSOProfile{
		ID:             profileResponse.Profile.ID,
		Email:          profileResponse.Profile.Email,
		FirstName:      profileResponse.Profile.FirstName,
		LastName:       profileResponse.Profile.LastName,
		OrganizationID: profileResponse.Profile.OrganizationID,
		ConnectionID:   profileResponse.Profile.ConnectionID,
		ConnectionType: string(profileResponse.Profile.ConnectionType),
		IdpID:          profileResponse.Profile.IdpID,
		RawAttributes:  profileResponse.Profile.RawAttributes,
	}

	return profile, nil
}

// deriveKey derives a 32-byte key from the cookie password using SHA-256
func (s *WorkOSService) deriveKey() []byte {
	hash := sha256.Sum256([]byte(s.cookiePassword))
	return hash[:]
}

// EncryptSessionCookie encrypts session data for the wos-session cookie
func (s *WorkOSService) EncryptSessionCookie(data *SessionData) (string, error) {
	if s.cookiePassword == "" {
		return "", errors.New("workos: cookie password not configured")
	}

	// Serialize session data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Create cipher
	key := s.deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	// Encode as base64
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptSessionCookie decrypts the wos-session cookie
func (s *WorkOSService) DecryptSessionCookie(encryptedData string) (*SessionData, error) {
	if s.cookiePassword == "" {
		return nil, errors.New("workos: cookie password not configured")
	}

	// Decode from base64
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cookie: %w", err)
	}

	// Create cipher
	key := s.deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Validate ciphertext length
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("invalid cookie data: too short")
	}

	// Extract nonce and decrypt
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt cookie: %w", err)
	}

	// Parse session data
	var data SessionData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	// Check expiration
	if time.Now().After(data.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return &data, nil
}

// CreateSessionData creates a new session data object
func (s *WorkOSService) CreateSessionData(profile *SSOProfile, expiryDuration time.Duration) *SessionData {
	return &SessionData{
		UserID:         profile.ID,
		OrganizationID: profile.OrganizationID,
		Email:          profile.Email,
		ExpiresAt:      time.Now().Add(expiryDuration),
	}
}

// SSOConnection represents an SSO connection from WorkOS
type SSOConnection struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ConnectionType string `json:"connection_type"`
	State          string `json:"state"`
}

// ListConnections lists SSO connections for an organization
func (s *WorkOSService) ListConnections(organizationID string) ([]SSOConnection, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	if organizationID == "" {
		return nil, errors.New("workos: organization ID required")
	}

	// List connections from WorkOS
	resp, err := sso.ListConnections(context.Background(), sso.ListConnectionsOpts{
		OrganizationID: organizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}

	connections := make([]SSOConnection, len(resp.Data))
	for i, conn := range resp.Data {
		connections[i] = SSOConnection{
			ID:             conn.ID,
			Name:           conn.Name,
			ConnectionType: string(conn.ConnectionType),
			State:          string(conn.State),
		}
	}

	return connections, nil
}

// GetOrganization retrieves an organization by ID
func (s *WorkOSService) GetOrganization(ctx context.Context, organizationID string) (*organizations.Organization, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	org, err := organizations.GetOrganization(ctx, organizations.GetOrganizationOpts{
		Organization: organizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// ==================== Extended Connection Management ====================

// SSOConnectionDetails contains detailed information about an SSO connection
type SSOConnectionDetails struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ConnectionType   string   `json:"connection_type"`
	State            string   `json:"state"`
	OrganizationID   string   `json:"organization_id"`
	Domains          []string `json:"domains,omitempty"`
	SAMLEntityID     string   `json:"saml_entity_id,omitempty"`
	SAMLSSOURL       string   `json:"saml_sso_url,omitempty"`
	OIDCClientID     string   `json:"oidc_client_id,omitempty"`
	OIDCDiscoveryURL string   `json:"oidc_discovery_url,omitempty"`
}

// GetConnection retrieves a specific SSO connection by ID
func (s *WorkOSService) GetConnection(ctx context.Context, connectionID string) (*SSOConnectionDetails, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	if connectionID == "" {
		return nil, errors.New("workos: connection ID required")
	}

	conn, err := sso.GetConnection(ctx, sso.GetConnectionOpts{
		Connection: connectionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	details := &SSOConnectionDetails{
		ID:             conn.ID,
		Name:           conn.Name,
		ConnectionType: string(conn.ConnectionType),
		State:          string(conn.State),
		OrganizationID: conn.OrganizationID,
	}

	// Extract domains
	for _, domain := range conn.Domains {
		details.Domains = append(details.Domains, domain.Domain)
	}

	return details, nil
}

// DeleteConnection deletes an SSO connection
func (s *WorkOSService) DeleteConnection(ctx context.Context, connectionID string) error {
	if !s.configured {
		return errors.New("workos: not configured")
	}

	if connectionID == "" {
		return errors.New("workos: connection ID required")
	}

	if err := sso.DeleteConnection(ctx, sso.DeleteConnectionOpts{
		Connection: connectionID,
	}); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	return nil
}

// ConnectionTestResult contains the result of testing an SSO connection
type ConnectionTestResult struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Connection *SSOConnectionDetails `json:"connection,omitempty"`
	AuthURL    string            `json:"auth_url,omitempty"`
	TestedAt   time.Time         `json:"tested_at"`
	Latency    int64             `json:"latency_ms"`
}

// TestConnection validates that an SSO connection is properly configured
func (s *WorkOSService) TestConnection(ctx context.Context, connectionID string) (*ConnectionTestResult, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	start := time.Now()
	result := &ConnectionTestResult{
		TestedAt: start,
	}

	// Get connection details
	conn, err := s.GetConnection(ctx, connectionID)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to retrieve connection: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, nil
	}

	result.Connection = conn

	// Check connection state
	if conn.State != "active" {
		result.Success = false
		result.Message = fmt.Sprintf("Connection is not active (state: %s)", conn.State)
		result.Latency = time.Since(start).Milliseconds()
		return result, nil
	}

	// Try to generate an authorization URL
	authURL, _, err := s.GenerateAuthorizationURL(AuthorizationOptions{
		ConnectionID: connectionID,
	})
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to generate authorization URL: %v", err)
		result.Latency = time.Since(start).Milliseconds()
		return result, nil
	}

	result.Success = true
	result.Message = "Connection is active and properly configured"
	result.AuthURL = authURL
	result.Latency = time.Since(start).Milliseconds()

	return result, nil
}

// ==================== Attribute Mapping ====================

// AttributeMappingConfig defines how to map SSO attributes to user fields
type AttributeMappingConfig struct {
	// Standard attribute names to look for
	EmailAttributes     []string `json:"email_attributes"`
	FirstNameAttributes []string `json:"first_name_attributes"`
	LastNameAttributes  []string `json:"last_name_attributes"`
	GroupsAttributes    []string `json:"groups_attributes"`

	// Custom mappings (source -> target)
	CustomMappings map[string]string `json:"custom_mappings,omitempty"`
}

// DefaultAttributeMapping returns the default attribute mapping configuration
func DefaultAttributeMapping() *AttributeMappingConfig {
	return &AttributeMappingConfig{
		EmailAttributes:     []string{"email", "Email", "mail", "emailAddress"},
		FirstNameAttributes: []string{"firstName", "first_name", "givenName", "given_name"},
		LastNameAttributes:  []string{"lastName", "last_name", "surname", "sn", "familyName"},
		GroupsAttributes:    []string{"groups", "memberOf", "roles"},
	}
}

// ExtractUserInfo extracts user information from raw SSO attributes using mappings
func (s *WorkOSService) ExtractUserInfo(rawAttributes map[string]interface{}, mappings *AttributeMappingConfig) map[string]string {
	if mappings == nil {
		mappings = DefaultAttributeMapping()
	}

	result := make(map[string]string)

	// Helper to find first matching attribute
	findAttribute := func(attributes []string) string {
		for _, attr := range attributes {
			if val, ok := rawAttributes[attr]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
		}
		return ""
	}

	result["email"] = findAttribute(mappings.EmailAttributes)
	result["first_name"] = findAttribute(mappings.FirstNameAttributes)
	result["last_name"] = findAttribute(mappings.LastNameAttributes)
	result["groups"] = findAttribute(mappings.GroupsAttributes)

	// Apply custom mappings
	for source, target := range mappings.CustomMappings {
		if val, ok := rawAttributes[source]; ok {
			if str, ok := val.(string); ok {
				result[target] = str
			}
		}
	}

	return result
}

// ==================== Organization Management ====================

// CreateOrganization creates a new organization in WorkOS
func (s *WorkOSService) CreateOrganization(ctx context.Context, name string, domains []string) (*organizations.Organization, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	org, err := organizations.CreateOrganization(ctx, organizations.CreateOrganizationOpts{
		Name:    name,
		Domains: domains,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &org, nil
}

// UpdateOrganization updates an organization in WorkOS
func (s *WorkOSService) UpdateOrganization(ctx context.Context, orgID, name string, domains []string) (*organizations.Organization, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	org, err := organizations.UpdateOrganization(ctx, organizations.UpdateOrganizationOpts{
		Organization: orgID,
		Name:         name,
		Domains:      domains,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &org, nil
}

// DeleteOrganization deletes an organization from WorkOS
func (s *WorkOSService) DeleteOrganization(ctx context.Context, orgID string) error {
	if !s.configured {
		return errors.New("workos: not configured")
	}

	if err := organizations.DeleteOrganization(ctx, organizations.DeleteOrganizationOpts{
		Organization: orgID,
	}); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// ListOrganizations lists all organizations
func (s *WorkOSService) ListOrganizations(ctx context.Context, limit int) ([]organizations.Organization, error) {
	if !s.configured {
		return nil, errors.New("workos: not configured")
	}

	resp, err := organizations.ListOrganizations(ctx, organizations.ListOrganizationsOpts{
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return resp.Data, nil
}

// ==================== Portal Links ====================

// GenerateAdminPortalLink generates a link for admins to configure SSO
func (s *WorkOSService) GenerateAdminPortalLink(ctx context.Context, orgID string, returnURL string) (string, error) {
	if !s.configured {
		return "", errors.New("workos: not configured")
	}

	// WorkOS Admin Portal is typically generated via their dashboard or specific portal link endpoints
	// This is a placeholder for when WorkOS provides a Go SDK method
	// For now, return a formatted URL that can be used with the WorkOS Admin Portal
	return fmt.Sprintf("https://api.workos.com/portal/organizations/%s/setup?return_url=%s", orgID, returnURL), nil
}
