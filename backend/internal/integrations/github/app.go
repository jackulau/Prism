package github

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GitHubApp handles GitHub App authentication and token management
type GitHubApp struct {
	appID      int64
	privateKey *rsa.PrivateKey
	httpClient *http.Client

	// Token cache
	tokenCache   map[int64]*cachedToken
	tokenCacheMu sync.RWMutex
}

// cachedToken represents a cached installation access token
type cachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// InstallationToken represents a GitHub App installation access token response
type InstallationToken struct {
	Token       string       `json:"token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

// Permissions represents GitHub App permissions
type Permissions struct {
	Contents     string `json:"contents,omitempty"`
	Issues       string `json:"issues,omitempty"`
	Metadata     string `json:"metadata,omitempty"`
	PullRequests string `json:"pull_requests,omitempty"`
}

// Installation represents a GitHub App installation
type Installation struct {
	ID                  int64        `json:"id"`
	Account             *Account     `json:"account"`
	AppID               int64        `json:"app_id"`
	TargetType          string       `json:"target_type"` // "User" or "Organization"
	Permissions         *Permissions `json:"permissions"`
	Events              []string     `json:"events"`
	RepositorySelection string       `json:"repository_selection"` // "all" or "selected"
	SingleFileName      string       `json:"single_file_name,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

// Account represents a GitHub account (user or organization)
type Account struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Type      string `json:"type"` // "User" or "Organization"
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// InstallationEvent represents a GitHub App installation webhook event
type InstallationEvent struct {
	Action       string        `json:"action"` // "created", "deleted", "suspend", "unsuspend"
	Installation *Installation `json:"installation"`
	Repositories []Repository  `json:"repositories,omitempty"`
	Sender       *User         `json:"sender"`
}

// InstallationRepositoriesEvent represents changes to repository access
type InstallationRepositoriesEvent struct {
	Action              string        `json:"action"` // "added" or "removed"
	Installation        *Installation `json:"installation"`
	RepositoriesAdded   []Repository  `json:"repositories_added,omitempty"`
	RepositoriesRemoved []Repository  `json:"repositories_removed,omitempty"`
	RepositorySelection string        `json:"repository_selection"`
	Sender              *User         `json:"sender"`
}

// NewGitHubApp creates a new GitHub App instance
func NewGitHubApp(appID int64, privateKeyPEM string) (*GitHubApp, error) {
	if appID == 0 {
		return nil, fmt.Errorf("GitHub App ID is required")
	}

	if privateKeyPEM == "" {
		return nil, fmt.Errorf("GitHub App private key is required")
	}

	// Parse the PEM-encoded private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var privateKey *rsa.PrivateKey
	var err error

	// Try PKCS#1 format first
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS#8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an RSA key")
		}
	}

	return &GitHubApp{
		appID:      appID,
		privateKey: privateKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenCache: make(map[int64]*cachedToken),
	}, nil
}

// GenerateJWT generates a JWT for authenticating as the GitHub App
func (app *GitHubApp) GenerateJWT() (string, error) {
	now := time.Now()

	// JWT is valid for 10 minutes (GitHub's maximum)
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": app.appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(app.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

// GetInstallationToken retrieves an access token for a specific installation
// Tokens are cached and automatically refreshed when expired
func (app *GitHubApp) GetInstallationToken(installationID int64) (string, error) {
	// Check cache first
	app.tokenCacheMu.RLock()
	cached, ok := app.tokenCache[installationID]
	app.tokenCacheMu.RUnlock()

	// Return cached token if still valid (with 5 minute buffer)
	if ok && time.Now().Add(5*time.Minute).Before(cached.ExpiresAt) {
		return cached.Token, nil
	}

	// Generate new JWT for authentication
	jwt, err := app.GenerateJWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Request new installation token
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := app.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request installation token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp InstallationToken
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// Cache the token
	app.tokenCacheMu.Lock()
	app.tokenCache[installationID] = &cachedToken{
		Token:     tokenResp.Token,
		ExpiresAt: tokenResp.ExpiresAt,
	}
	app.tokenCacheMu.Unlock()

	return tokenResp.Token, nil
}

// GetInstallation retrieves information about a specific installation
func (app *GitHubApp) GetInstallation(installationID int64) (*Installation, error) {
	jwt, err := app.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d", installationID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := app.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch installation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var installation Installation
	if err := json.NewDecoder(resp.Body).Decode(&installation); err != nil {
		return nil, fmt.Errorf("failed to parse installation: %w", err)
	}

	return &installation, nil
}

// ListInstallations lists all installations for this GitHub App
func (app *GitHubApp) ListInstallations() ([]*Installation, error) {
	jwt, err := app.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	req, err := http.NewRequest("GET", "https://api.github.com/app/installations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := app.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch installations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var installations []*Installation
	if err := json.NewDecoder(resp.Body).Decode(&installations); err != nil {
		return nil, fmt.Errorf("failed to parse installations: %w", err)
	}

	return installations, nil
}

// GetRepositoryInstallation finds the installation ID for a specific repository
func (app *GitHubApp) GetRepositoryInstallation(owner, repo string) (*Installation, error) {
	jwt, err := app.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/installation", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := app.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository installation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("GitHub App not installed on repository %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var installation Installation
	if err := json.NewDecoder(resp.Body).Decode(&installation); err != nil {
		return nil, fmt.Errorf("failed to parse installation: %w", err)
	}

	return &installation, nil
}

// InvalidateToken removes a cached token for an installation
func (app *GitHubApp) InvalidateToken(installationID int64) {
	app.tokenCacheMu.Lock()
	delete(app.tokenCache, installationID)
	app.tokenCacheMu.Unlock()
}

// IsConfigured returns true if the GitHub App is configured
func (app *GitHubApp) IsConfigured() bool {
	return app != nil && app.appID != 0 && app.privateKey != nil
}

// AppID returns the GitHub App ID
func (app *GitHubApp) AppID() int64 {
	return app.appID
}
