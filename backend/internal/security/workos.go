package security

import (
	"errors"

	"github.com/workos/workos-go/v4/pkg/sso"
)

// WorkOSService handles WorkOS SSO authentication
type WorkOSService struct {
	apiKey         string
	clientID       string
	redirectURI    string
	cookiePassword string
	configured     bool
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
