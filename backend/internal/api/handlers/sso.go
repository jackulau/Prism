package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/security"
)

// SSOHandler handles SSO configuration endpoints
type SSOHandler struct {
	registry      *security.SSORegistry
	workosService *security.WorkOSService
	oauth2Service *security.OAuth2Service
}

// NewSSOHandler creates a new SSO handler
func NewSSOHandler(
	registry *security.SSORegistry,
	workosService *security.WorkOSService,
	oauth2Service *security.OAuth2Service,
) *SSOHandler {
	return &SSOHandler{
		registry:      registry,
		workosService: workosService,
		oauth2Service: oauth2Service,
	}
}

// ListProviders lists all SSO providers for an organization
// GET /api/v1/org/sso/providers
func (h *SSOHandler) ListProviders(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providers, err := h.registry.ListProviders(context.Background(), orgID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list providers",
		})
	}

	// Convert to list items (sanitized)
	items := make([]security.SSOProviderListItem, len(providers))
	for i, p := range providers {
		items[i] = p.ToListItem()
	}

	return c.JSON(fiber.Map{
		"providers": items,
	})
}

// GetProvider retrieves a single SSO provider
// GET /api/v1/org/sso/providers/:id
func (h *SSOHandler) GetProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	// Verify organization ownership
	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Sanitize before returning
	provider.SanitizeForResponse()

	return c.JSON(fiber.Map{
		"provider": provider,
	})
}

// CreateProvider creates a new SSO provider
// POST /api/v1/org/sso/providers
func (h *SSOHandler) CreateProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	var req security.CreateSSOProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Build provider config
	config := &security.SSOProviderConfig{
		OrganizationID:    orgID.(string),
		Name:              req.Name,
		Type:              security.NormalizeType(string(req.Type)),
		Priority:          req.Priority,
		Enabled:           false, // Start disabled until tested
		Status:            security.SSOProviderStatusPending,
		SAMLConfig:        req.SAMLConfig,
		OIDCConfig:        req.OIDCConfig,
		OAuth2Config:      req.OAuth2Config,
		AttributeMappings: req.AttributeMappings,
	}

	// Validate
	if err := config.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Register provider
	if err := h.registry.RegisterProvider(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create provider",
		})
	}

	// Sanitize before returning
	config.SanitizeForResponse()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"provider": config,
	})
}

// UpdateProvider updates an SSO provider
// PUT /api/v1/org/sso/providers/:id
func (h *SSOHandler) UpdateProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	// Get existing provider
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	// Verify organization ownership
	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req security.UpdateSSOProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Apply updates
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Priority != nil {
		provider.Priority = *req.Priority
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
		if *req.Enabled {
			provider.Status = security.SSOProviderStatusActive
		} else {
			provider.Status = security.SSOProviderStatusInactive
		}
	}
	if req.SAMLConfig != nil {
		provider.SAMLConfig = req.SAMLConfig
	}
	if req.OIDCConfig != nil {
		provider.OIDCConfig = req.OIDCConfig
	}
	if req.OAuth2Config != nil {
		provider.OAuth2Config = req.OAuth2Config
	}
	if req.AttributeMappings != nil {
		provider.AttributeMappings = req.AttributeMappings
	}

	// Validate
	if err := provider.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Update
	if err := h.registry.UpdateProvider(context.Background(), provider); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update provider",
		})
	}

	// Sanitize before returning
	provider.SanitizeForResponse()

	return c.JSON(fiber.Map{
		"provider": provider,
	})
}

// DeleteProvider deletes an SSO provider
// DELETE /api/v1/org/sso/providers/:id
func (h *SSOHandler) DeleteProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	// Get existing provider to verify ownership
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	// Verify organization ownership
	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Delete
	if err := h.registry.DeleteProvider(context.Background(), providerID, orgID.(string)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete provider",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// EnableProvider enables an SSO provider
// POST /api/v1/org/sso/providers/:id/enable
func (h *SSOHandler) EnableProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	// Get existing provider to verify ownership
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.registry.EnableProvider(context.Background(), providerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enable provider",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// DisableProvider disables an SSO provider
// POST /api/v1/org/sso/providers/:id/disable
func (h *SSOHandler) DisableProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	// Get existing provider to verify ownership
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.registry.DisableProvider(context.Background(), providerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to disable provider",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// TestProvider tests an SSO provider connection
// POST /api/v1/org/sso/providers/:id/test
func (h *SSOHandler) TestProvider(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	// Get provider
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	if provider.OrganizationID != orgID.(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var result *security.SSOTestResult

	switch provider.Type {
	case security.SSOProviderTypeSAML, security.SSOProviderTypeOIDC:
		// Test WorkOS connection
		if provider.WorkOSConnectionID != "" && h.workosService != nil {
			testResult, err := h.workosService.TestConnection(context.Background(), provider.WorkOSConnectionID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to test connection",
				})
			}
			result = &security.SSOTestResult{
				Success:  testResult.Success,
				Message:  testResult.Message,
				TestedAt: testResult.TestedAt,
				Latency:  time.Duration(testResult.Latency) * time.Millisecond,
			}
		} else {
			result = &security.SSOTestResult{
				Success: false,
				Message: "WorkOS connection not configured",
			}
		}

	case security.SSOProviderTypeOAuth:
		// Test OAuth2 provider
		if h.oauth2Service != nil && provider.OAuth2Config != nil {
			result, err = h.oauth2Service.TestProvider(context.Background(), provider.OAuth2Config)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to test provider",
				})
			}
		} else {
			result = &security.SSOTestResult{
				Success: false,
				Message: "OAuth2 configuration not found",
			}
		}

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported provider type",
		})
	}

	// Update provider status based on test result
	if result.Success {
		if err := h.registry.ClearProviderError(context.Background(), providerID); err != nil {
			// Log but don't fail
		}
	} else {
		if err := h.registry.SetProviderError(context.Background(), providerID, result.Message); err != nil {
			// Log but don't fail
		}
	}

	return c.JSON(fiber.Map{
		"result": result,
	})
}

// GetProviderSummary returns a summary of SSO providers for an organization
// GET /api/v1/org/sso/summary
func (h *SSOHandler) GetProviderSummary(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	summary, err := h.registry.GetProviderSummary(context.Background(), orgID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get summary",
		})
	}

	return c.JSON(fiber.Map{
		"summary": summary,
	})
}

// ReorderProviders reorders SSO providers by priority
// POST /api/v1/org/sso/providers/reorder
func (h *SSOHandler) ReorderProviders(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization context required",
		})
	}

	var req struct {
		ProviderIDs []string `json:"provider_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.ProviderIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider_ids required",
		})
	}

	if err := h.registry.ReorderProviders(context.Background(), orgID.(string), req.ProviderIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to reorder providers",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// GetActiveProviders returns active SSO providers for the login page
// GET /api/v1/org/sso/active
func (h *SSOHandler) GetActiveProviders(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID")
	if orgID == nil || orgID.(string) == "" {
		// Try to get org from query param for public access
		orgID = c.Query("organization_id")
		if orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "organization_id required",
			})
		}
	}

	providers, err := h.registry.GetActiveProviders(context.Background(), orgID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get providers",
		})
	}

	// Return simplified list for login page
	items := make([]fiber.Map, len(providers))
	for i, p := range providers {
		item := fiber.Map{
			"id":   p.ID,
			"name": p.Name,
			"type": p.Type,
		}

		// Add display info for OAuth2 providers
		if p.Type == security.SSOProviderTypeOAuth && p.OAuth2Config != nil {
			item["display_name"] = p.OAuth2Config.DisplayName
			item["icon_url"] = p.OAuth2Config.IconURL
			item["button_color"] = p.OAuth2Config.ButtonColor
		}

		items[i] = item
	}

	return c.JSON(fiber.Map{
		"providers": items,
	})
}

// InitiateSSO initiates SSO login for a provider
// POST /api/v1/org/sso/providers/:id/initiate
func (h *SSOHandler) InitiateSSO(c *fiber.Ctx) error {
	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider ID required",
		})
	}

	var req struct {
		RedirectTo string `json:"redirect_to"`
	}
	c.BodyParser(&req)

	// Get provider
	provider, err := h.registry.GetProvider(context.Background(), providerID)
	if err != nil || provider == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider not found",
		})
	}

	if !provider.Enabled || provider.Status != security.SSOProviderStatusActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider is not active",
		})
	}

	var authURL string

	switch provider.Type {
	case security.SSOProviderTypeSAML, security.SSOProviderTypeOIDC:
		// Use WorkOS for SAML/OIDC
		if h.workosService == nil || provider.WorkOSConnectionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "WorkOS not configured for this provider",
			})
		}

		url, _, err := h.workosService.GenerateAuthorizationURL(security.AuthorizationOptions{
			ConnectionID: provider.WorkOSConnectionID,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate authorization URL",
			})
		}
		authURL = url

	case security.SSOProviderTypeOAuth:
		// Use OAuth2 service
		if h.oauth2Service == nil || provider.OAuth2Config == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "OAuth2 not configured for this provider",
			})
		}

		url, err := h.oauth2Service.GenerateAuthorizationURL(security.OAuth2AuthorizeOptions{
			ProviderConfig: provider.OAuth2Config,
			ProviderID:     provider.ID,
			OrganizationID: provider.OrganizationID,
			RedirectTo:     req.RedirectTo,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate authorization URL",
			})
		}
		authURL = url

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported provider type",
		})
	}

	return c.JSON(fiber.Map{
		"authorization_url": authURL,
	})
}
