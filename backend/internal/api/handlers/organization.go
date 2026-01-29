package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	orgRepo *repository.OrganizationRepository
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgRepo *repository.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{orgRepo: orgRepo}
}

// CreateOrganizationRequest represents a request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	Name                       *string `json:"name,omitempty"`
	SubscriptionTier           *string `json:"subscription_tier,omitempty"`
	SubscriptionStatus         *string `json:"subscription_status,omitempty"`
	CancelAtPeriodEnd          *bool   `json:"cancel_at_period_end,omitempty"`
	TokenCostLimitMicrodollars *int64  `json:"token_cost_limit_microdollars,omitempty"`
	SandboxTimeLimitSeconds    *int64  `json:"sandbox_time_limit_seconds,omitempty"`
}

// OrganizationResponse represents an organization response
type OrganizationResponse struct {
	ID                         string  `json:"id"`
	WorkOSOrganizationID       *string `json:"workos_organization_id,omitempty"`
	Name                       string  `json:"name"`
	StripeCustomerID           *string `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID       *string `json:"stripe_subscription_id,omitempty"`
	SubscriptionTier           string  `json:"subscription_tier"`
	SubscriptionStatus         string  `json:"subscription_status"`
	CancelAtPeriodEnd          bool    `json:"cancel_at_period_end"`
	TokenCostUsedMicrodollars  int64   `json:"token_cost_used_microdollars"`
	TokenCostLimitMicrodollars int64   `json:"token_cost_limit_microdollars"`
	SandboxTimeUsedSeconds     int64   `json:"sandbox_time_used_seconds"`
	SandboxTimeLimitSeconds    int64   `json:"sandbox_time_limit_seconds"`
	BillingPeriodStart         *string `json:"billing_period_start,omitempty"`
	BillingPeriodEnd           *string `json:"billing_period_end,omitempty"`
	CreatedAt                  string  `json:"created_at"`
	UpdatedAt                  string  `json:"updated_at"`
}

// Valid subscription tiers
var validSubscriptionTiers = map[string]bool{
	"FREE":       true,
	"PAID":       true,
	"ENTERPRISE": true,
}

// Valid subscription statuses
var validSubscriptionStatuses = map[string]bool{
	"ACTIVE":     true,
	"CANCELED":   true,
	"PAST_DUE":   true,
	"INCOMPLETE": true,
}

// toOrganizationResponse converts an Organization entity to an OrganizationResponse
func toOrganizationResponse(org *repository.Organization) *OrganizationResponse {
	resp := &OrganizationResponse{
		ID:                         org.ID,
		WorkOSOrganizationID:       org.WorkOSOrganizationID,
		Name:                       org.Name,
		StripeCustomerID:           org.StripeCustomerID,
		StripeSubscriptionID:       org.StripeSubscriptionID,
		SubscriptionTier:           org.SubscriptionTier,
		SubscriptionStatus:         org.SubscriptionStatus,
		CancelAtPeriodEnd:          org.CancelAtPeriodEnd,
		TokenCostUsedMicrodollars:  org.TokenCostUsedMicrodollars,
		TokenCostLimitMicrodollars: org.TokenCostLimitMicrodollars,
		SandboxTimeUsedSeconds:     org.SandboxTimeUsedSeconds,
		SandboxTimeLimitSeconds:    org.SandboxTimeLimitSeconds,
		CreatedAt:                  org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                  org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if org.BillingPeriodStart != nil {
		formatted := org.BillingPeriodStart.Format("2006-01-02T15:04:05Z")
		resp.BillingPeriodStart = &formatted
	}
	if org.BillingPeriodEnd != nil {
		formatted := org.BillingPeriodEnd.Format("2006-01-02T15:04:05Z")
		resp.BillingPeriodEnd = &formatted
	}

	return resp
}

// Create creates a new organization
// POST /organizations
func (h *OrganizationHandler) Create(c *fiber.Ctx) error {
	var req CreateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}
	if len(name) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name must be at most 255 characters",
		})
	}

	org, err := h.orgRepo.Create(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create organization",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toOrganizationResponse(org))
}

// GetByID retrieves an organization by ID
// GET /organizations/:id
func (h *OrganizationHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization id is required",
		})
	}

	org, err := h.orgRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get organization",
		})
	}
	if org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "organization not found",
		})
	}

	return c.JSON(toOrganizationResponse(org))
}

// Update updates an organization
// PATCH /organizations/:id
func (h *OrganizationHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization id is required",
		})
	}

	var req UpdateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Get existing organization
	org, err := h.orgRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get organization",
		})
	}
	if org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "organization not found",
		})
	}

	// Apply partial updates
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "name cannot be empty",
			})
		}
		if len(name) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "name must be at most 255 characters",
			})
		}
		org.Name = name
	}

	if req.SubscriptionTier != nil {
		tier := strings.ToUpper(*req.SubscriptionTier)
		if !validSubscriptionTiers[tier] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid subscription_tier, must be one of: FREE, PAID, ENTERPRISE",
			})
		}
		org.SubscriptionTier = tier
	}

	if req.SubscriptionStatus != nil {
		status := strings.ToUpper(*req.SubscriptionStatus)
		if !validSubscriptionStatuses[status] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid subscription_status, must be one of: ACTIVE, CANCELED, PAST_DUE, INCOMPLETE",
			})
		}
		org.SubscriptionStatus = status
	}

	if req.CancelAtPeriodEnd != nil {
		org.CancelAtPeriodEnd = *req.CancelAtPeriodEnd
	}

	if req.TokenCostLimitMicrodollars != nil {
		if *req.TokenCostLimitMicrodollars < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "token_cost_limit_microdollars must be non-negative",
			})
		}
		org.TokenCostLimitMicrodollars = *req.TokenCostLimitMicrodollars
	}

	if req.SandboxTimeLimitSeconds != nil {
		if *req.SandboxTimeLimitSeconds < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "sandbox_time_limit_seconds must be non-negative",
			})
		}
		org.SandboxTimeLimitSeconds = *req.SandboxTimeLimitSeconds
	}

	if err := h.orgRepo.Update(org); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update organization",
		})
	}

	return c.JSON(toOrganizationResponse(org))
}

// Delete deletes an organization
// DELETE /organizations/:id
func (h *OrganizationHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization id is required",
		})
	}

	err := h.orgRepo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "organization not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete organization",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// List lists organizations with pagination
// GET /organizations
func (h *OrganizationHandler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	// Validate pagination params
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	organizations, err := h.orgRepo.List(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list organizations",
		})
	}

	responses := make([]*OrganizationResponse, len(organizations))
	for i, org := range organizations {
		responses[i] = toOrganizationResponse(org)
	}

	return c.JSON(fiber.Map{
		"organizations": responses,
		"limit":         limit,
		"offset":        offset,
	})
}
