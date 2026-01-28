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
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
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
