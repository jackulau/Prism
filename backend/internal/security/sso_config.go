package security

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// SSOProviderType represents the type of SSO provider
type SSOProviderType string

const (
	SSOProviderTypeSAML  SSOProviderType = "saml"
	SSOProviderTypeOIDC  SSOProviderType = "oidc"
	SSOProviderTypeOAuth SSOProviderType = "oauth2"
)

// SSOProviderStatus represents the status of an SSO provider
type SSOProviderStatus string

const (
	SSOProviderStatusPending  SSOProviderStatus = "pending"
	SSOProviderStatusActive   SSOProviderStatus = "active"
	SSOProviderStatusInactive SSOProviderStatus = "inactive"
	SSOProviderStatusError    SSOProviderStatus = "error"
)

// SSOProviderConfig represents the configuration for an SSO provider
type SSOProviderConfig struct {
	// Common fields
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	Name           string            `json:"name"`
	Type           SSOProviderType   `json:"type"`
	Status         SSOProviderStatus `json:"status"`
	Priority       int               `json:"priority"` // Display order on login page
	Enabled        bool              `json:"enabled"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`

	// SAML-specific configuration
	SAMLConfig *SAMLConfiguration `json:"saml_config,omitempty"`

	// OIDC-specific configuration
	OIDCConfig *OIDCConfiguration `json:"oidc_config,omitempty"`

	// OAuth2-specific configuration
	OAuth2Config *OAuth2Configuration `json:"oauth2_config,omitempty"`

	// Attribute mappings (for custom claim/attribute to user field mapping)
	AttributeMappings []AttributeMapping `json:"attribute_mappings,omitempty"`

	// WorkOS connection ID (if using WorkOS as the backend)
	WorkOSConnectionID string `json:"workos_connection_id,omitempty"`

	// Last error message (if status is error)
	LastError string `json:"last_error,omitempty"`
}

// SAMLConfiguration contains SAML 2.0 specific settings
type SAMLConfiguration struct {
	// Identity Provider (IdP) metadata URL for automatic configuration
	MetadataURL string `json:"metadata_url,omitempty"`

	// Manual IdP configuration (used if MetadataURL is empty)
	EntityID          string `json:"entity_id,omitempty"`
	SSOURL            string `json:"sso_url,omitempty"`
	SLOUrl            string `json:"slo_url,omitempty"`
	X509Certificate   string `json:"x509_certificate,omitempty"`
	X509CertificateSHA256 string `json:"x509_certificate_sha256,omitempty"`

	// Service Provider (SP) configuration
	SPEntityID     string `json:"sp_entity_id,omitempty"`
	SPACSURL       string `json:"sp_acs_url,omitempty"`      // Assertion Consumer Service URL
	SPMetadataURL  string `json:"sp_metadata_url,omitempty"`

	// SAML options
	SignRequest         bool   `json:"sign_request"`
	SignatureAlgorithm  string `json:"signature_algorithm,omitempty"` // e.g., "RSA-SHA256"
	DigestAlgorithm     string `json:"digest_algorithm,omitempty"`    // e.g., "SHA256"
	RequestedAuthnContext string `json:"requested_authn_context,omitempty"`
	AllowUnsolicitedResponse bool `json:"allow_unsolicited_response"`
}

// OIDCConfiguration contains OIDC specific settings
type OIDCConfiguration struct {
	// Discovery URL for automatic configuration
	DiscoveryURL string `json:"discovery_url,omitempty"`

	// Manual configuration (used if DiscoveryURL is empty)
	Issuer               string `json:"issuer,omitempty"`
	AuthorizationURL     string `json:"authorization_url,omitempty"`
	TokenURL             string `json:"token_url,omitempty"`
	UserInfoURL          string `json:"user_info_url,omitempty"`
	JWKSURL              string `json:"jwks_url,omitempty"`
	EndSessionURL        string `json:"end_session_url,omitempty"`

	// Client credentials (client_secret is encrypted in storage)
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"` // Only present when creating/updating

	// Scopes to request
	Scopes []string `json:"scopes,omitempty"`

	// Response type (code, token, id_token)
	ResponseType string `json:"response_type,omitempty"`

	// Response mode (query, fragment, form_post)
	ResponseMode string `json:"response_mode,omitempty"`

	// PKCE support
	UsePKCE bool `json:"use_pkce"`

}

// OAuth2Configuration contains generic OAuth2 settings for custom providers
type OAuth2Configuration struct {
	// Provider display name
	DisplayName string `json:"display_name"`

	// Authorization endpoints
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	UserInfoURL      string `json:"user_info_url,omitempty"`
	RevokeURL        string `json:"revoke_url,omitempty"`

	// Client credentials (client_secret is encrypted in storage)
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"` // Only present when creating/updating

	// Scopes to request
	Scopes []string `json:"scopes,omitempty"`

	// Response type
	ResponseType string `json:"response_type,omitempty"`

	// Token authentication method (client_secret_basic, client_secret_post)
	TokenAuthMethod string `json:"token_auth_method,omitempty"`

	// User info claim mappings
	ClaimMappings UserClaimMappings `json:"claim_mappings,omitempty"`

	// PKCE support
	UsePKCE bool `json:"use_pkce"`

	// Icon URL for display
	IconURL string `json:"icon_url,omitempty"`

	// Button color (hex)
	ButtonColor string `json:"button_color,omitempty"`
}

// UserClaimMappings defines how to extract user info from OAuth2 responses
type UserClaimMappings struct {
	ID        string `json:"id,omitempty"`         // Path to user ID claim
	Email     string `json:"email,omitempty"`      // Path to email claim
	FirstName string `json:"first_name,omitempty"` // Path to first name claim
	LastName  string `json:"last_name,omitempty"`  // Path to last name claim
	Name      string `json:"name,omitempty"`       // Path to full name claim
	Picture   string `json:"picture,omitempty"`    // Path to avatar/picture claim
	Groups    string `json:"groups,omitempty"`     // Path to groups/roles claim
}

// AttributeMapping defines a mapping from SSO attribute/claim to user field
type AttributeMapping struct {
	ID               string `json:"id"`
	SSOProviderID    string `json:"sso_provider_id"`
	SourceAttribute  string `json:"source_attribute"`  // Attribute name from IdP
	TargetField      string `json:"target_field"`      // Field in our user model
	TransformType    string `json:"transform_type,omitempty"` // e.g., "lowercase", "split", "regex"
	TransformPattern string `json:"transform_pattern,omitempty"`
}

// Validate validates the SSO provider configuration
func (c *SSOProviderConfig) Validate() error {
	if c.OrganizationID == "" {
		return errors.New("organization_id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Type == "" {
		return errors.New("type is required")
	}

	switch c.Type {
	case SSOProviderTypeSAML:
		return c.validateSAML()
	case SSOProviderTypeOIDC:
		return c.validateOIDC()
	case SSOProviderTypeOAuth:
		return c.validateOAuth2()
	default:
		return errors.New("invalid provider type: must be saml, oidc, or oauth2")
	}
}

func (c *SSOProviderConfig) validateSAML() error {
	if c.SAMLConfig == nil {
		return errors.New("saml_config is required for SAML providers")
	}

	cfg := c.SAMLConfig

	// Either metadata URL or manual configuration is required
	if cfg.MetadataURL == "" {
		if cfg.EntityID == "" {
			return errors.New("entity_id is required when metadata_url is not provided")
		}
		if cfg.SSOURL == "" {
			return errors.New("sso_url is required when metadata_url is not provided")
		}
		if !isValidURL(cfg.SSOURL) {
			return errors.New("sso_url is not a valid URL")
		}
	} else if !isValidURL(cfg.MetadataURL) {
		return errors.New("metadata_url is not a valid URL")
	}

	return nil
}

func (c *SSOProviderConfig) validateOIDC() error {
	if c.OIDCConfig == nil {
		return errors.New("oidc_config is required for OIDC providers")
	}

	cfg := c.OIDCConfig

	if cfg.ClientID == "" {
		return errors.New("client_id is required for OIDC")
	}

	// Either discovery URL or manual configuration is required
	if cfg.DiscoveryURL == "" {
		if cfg.Issuer == "" {
			return errors.New("issuer is required when discovery_url is not provided")
		}
		if cfg.AuthorizationURL == "" {
			return errors.New("authorization_url is required when discovery_url is not provided")
		}
		if cfg.TokenURL == "" {
			return errors.New("token_url is required when discovery_url is not provided")
		}
		if !isValidURL(cfg.AuthorizationURL) {
			return errors.New("authorization_url is not a valid URL")
		}
		if !isValidURL(cfg.TokenURL) {
			return errors.New("token_url is not a valid URL")
		}
	} else if !isValidURL(cfg.DiscoveryURL) {
		return errors.New("discovery_url is not a valid URL")
	}

	return nil
}

func (c *SSOProviderConfig) validateOAuth2() error {
	if c.OAuth2Config == nil {
		return errors.New("oauth2_config is required for OAuth2 providers")
	}

	cfg := c.OAuth2Config

	if cfg.ClientID == "" {
		return errors.New("client_id is required for OAuth2")
	}
	if cfg.AuthorizationURL == "" {
		return errors.New("authorization_url is required for OAuth2")
	}
	if cfg.TokenURL == "" {
		return errors.New("token_url is required for OAuth2")
	}
	if !isValidURL(cfg.AuthorizationURL) {
		return errors.New("authorization_url is not a valid URL")
	}
	if !isValidURL(cfg.TokenURL) {
		return errors.New("token_url is not a valid URL")
	}

	return nil
}

// isValidURL checks if a string is a valid URL
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// GetDefaultScopes returns default scopes for a provider type
func GetDefaultScopes(providerType SSOProviderType) []string {
	switch providerType {
	case SSOProviderTypeOIDC:
		return []string{"openid", "profile", "email"}
	case SSOProviderTypeOAuth:
		return []string{"email", "profile"}
	default:
		return nil
	}
}

// SSOTestResult represents the result of testing an SSO connection
type SSOTestResult struct {
	Success     bool              `json:"success"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details,omitempty"`
	TestedAt    time.Time         `json:"tested_at"`
	Latency     time.Duration     `json:"latency_ms"`
}

// CreateSSOProviderRequest represents a request to create an SSO provider
type CreateSSOProviderRequest struct {
	Name              string              `json:"name"`
	Type              SSOProviderType     `json:"type"`
	Priority          int                 `json:"priority"`
	SAMLConfig        *SAMLConfiguration  `json:"saml_config,omitempty"`
	OIDCConfig        *OIDCConfiguration  `json:"oidc_config,omitempty"`
	OAuth2Config      *OAuth2Configuration `json:"oauth2_config,omitempty"`
	AttributeMappings []AttributeMapping   `json:"attribute_mappings,omitempty"`
}

// UpdateSSOProviderRequest represents a request to update an SSO provider
type UpdateSSOProviderRequest struct {
	Name              *string             `json:"name,omitempty"`
	Priority          *int                `json:"priority,omitempty"`
	Enabled           *bool               `json:"enabled,omitempty"`
	SAMLConfig        *SAMLConfiguration  `json:"saml_config,omitempty"`
	OIDCConfig        *OIDCConfiguration  `json:"oidc_config,omitempty"`
	OAuth2Config      *OAuth2Configuration `json:"oauth2_config,omitempty"`
	AttributeMappings []AttributeMapping   `json:"attribute_mappings,omitempty"`
}

// SSOProviderListItem represents a summary of an SSO provider for listing
type SSOProviderListItem struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      SSOProviderType   `json:"type"`
	Status    SSOProviderStatus `json:"status"`
	Priority  int               `json:"priority"`
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ToListItem converts a full config to a list item
func (c *SSOProviderConfig) ToListItem() SSOProviderListItem {
	return SSOProviderListItem{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		Status:    c.Status,
		Priority:  c.Priority,
		Enabled:   c.Enabled,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// SanitizeForResponse removes sensitive data before sending to client
func (c *SSOProviderConfig) SanitizeForResponse() {
	// Remove client secrets
	if c.OIDCConfig != nil {
		c.OIDCConfig.ClientSecret = ""
	}
	if c.OAuth2Config != nil {
		c.OAuth2Config.ClientSecret = ""
	}
	// Remove SAML certificates (keep fingerprint)
	if c.SAMLConfig != nil {
		c.SAMLConfig.X509Certificate = ""
	}
}

// NormalizeType normalizes the provider type to lowercase
func NormalizeType(t string) SSOProviderType {
	switch strings.ToLower(t) {
	case "saml", "saml2", "saml2.0":
		return SSOProviderTypeSAML
	case "oidc", "openid", "openid-connect":
		return SSOProviderTypeOIDC
	case "oauth", "oauth2", "oauth2.0":
		return SSOProviderTypeOAuth
	default:
		return SSOProviderType(strings.ToLower(t))
	}
}
