package workos

import "time"

// Organization represents a WorkOS organization
type Organization struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	AllowProfilesOutsideOrg bool     `json:"allow_profiles_outside_organization"`
	Domains                 []Domain `json:"domains"`
	CreatedAt               string   `json:"created_at"`
	UpdatedAt               string   `json:"updated_at"`
}

// Domain represents a domain associated with an organization
type Domain struct {
	ID             string `json:"id"`
	Domain         string `json:"domain"`
	State          string `json:"state"`
	VerificationType string `json:"verification_type"`
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name                    string   `json:"name"`
	AllowProfilesOutsideOrg bool     `json:"allow_profiles_outside_organization,omitempty"`
	DomainData              []Domain `json:"domain_data,omitempty"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name                    string   `json:"name,omitempty"`
	AllowProfilesOutsideOrg *bool    `json:"allow_profiles_outside_organization,omitempty"`
	DomainData              []Domain `json:"domain_data,omitempty"`
}

// ListOrganizationsResponse represents the paginated response for listing organizations
type ListOrganizationsResponse struct {
	Data     []Organization `json:"data"`
	ListMeta ListMeta       `json:"list_metadata"`
}

// ListMeta contains pagination metadata
type ListMeta struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// WebhookEvent represents a WorkOS webhook event
type WebhookEvent struct {
	ID        string          `json:"id"`
	Event     string          `json:"event"`
	Data      WebhookData     `json:"data"`
	CreatedAt string          `json:"created_at"`
}

// WebhookData contains the event-specific data
type WebhookData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ErrorResponse represents a WorkOS API error
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// OrganizationEvent types for webhooks
const (
	EventOrganizationCreated = "organization.created"
	EventOrganizationUpdated = "organization.updated"
	EventOrganizationDeleted = "organization.deleted"
)

// LocalOrganization represents an organization in the local database
type LocalOrganization struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	WorkOSOrganizationID string   `json:"workos_organization_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
