package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations/workos"
)

// OrganizationHandler handles organization-related endpoints
type OrganizationHandler struct {
	orgRepo      *repository.OrganizationRepository
	workosClient *workos.Client
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(
	orgRepo *repository.OrganizationRepository,
	workosClient *workos.Client,
) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo:      orgRepo,
		workosClient: workosClient,
	}
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		Name       string `json:"name"`
		SyncWorkOS bool   `json:"sync_workos"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	var workosOrgID string

	// Optionally create in WorkOS
	if req.SyncWorkOS && h.workosClient != nil && h.workosClient.Enabled() {
		workosOrg, err := h.workosClient.CreateOrganization(req.Name)
		if err != nil {
			log.Printf("Failed to create organization in WorkOS: %v", err)
			// Continue without WorkOS - graceful degradation
		} else {
			workosOrgID = workosOrg.ID
		}
	}

	// Create in local database
	org, err := h.orgRepo.Create(req.Name, workosOrgID)
	if err != nil {
		log.Printf("Failed to create organization: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create organization",
		})
	}

	// Add the creator as an admin member
	_, err = h.orgRepo.AddMember(org.ID, userID, "admin")
	if err != nil {
		log.Printf("Failed to add creator as admin: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":                     org.ID,
		"name":                   org.Name,
		"workos_organization_id": workosOrgID,
		"created_at":             org.CreatedAt,
	})
}

// GetOrganization retrieves an organization by ID
func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")

	org, err := h.orgRepo.GetByID(orgID)
	if err != nil || org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "organization not found",
		})
	}

	// Check membership
	isMember, err := h.orgRepo.IsMember(orgID, userID)
	if err != nil || !isMember {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	workosOrgID := ""
	if org.WorkOSOrganizationID.Valid {
		workosOrgID = org.WorkOSOrganizationID.String
	}

	return c.JSON(fiber.Map{
		"id":                     org.ID,
		"name":                   org.Name,
		"workos_organization_id": workosOrgID,
		"created_at":             org.CreatedAt,
		"updated_at":             org.UpdatedAt,
	})
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")

	// Check admin role
	role, err := h.orgRepo.GetMemberRole(orgID, userID)
	if err != nil || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	org, err := h.orgRepo.GetByID(orgID)
	if err != nil || org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "organization not found",
		})
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	// Update in WorkOS if linked
	if org.WorkOSOrganizationID.Valid && h.workosClient != nil && h.workosClient.Enabled() {
		if err := h.workosClient.UpdateOrganization(org.WorkOSOrganizationID.String, req.Name); err != nil {
			log.Printf("Failed to update organization in WorkOS: %v", err)
			// Continue without WorkOS update - graceful degradation
		}
	}

	workosOrgID := ""
	if org.WorkOSOrganizationID.Valid {
		workosOrgID = org.WorkOSOrganizationID.String
	}

	// Update in local database
	if err := h.orgRepo.Update(orgID, req.Name, workosOrgID); err != nil {
		log.Printf("Failed to update organization: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update organization",
		})
	}

	return c.JSON(fiber.Map{
		"id":   orgID,
		"name": req.Name,
	})
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")

	// Check admin role
	role, err := h.orgRepo.GetMemberRole(orgID, userID)
	if err != nil || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	org, err := h.orgRepo.GetByID(orgID)
	if err != nil || org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "organization not found",
		})
	}

	// Delete from WorkOS if linked
	if org.WorkOSOrganizationID.Valid && h.workosClient != nil && h.workosClient.Enabled() {
		if err := h.workosClient.DeleteOrganization(org.WorkOSOrganizationID.String); err != nil {
			log.Printf("Failed to delete organization from WorkOS: %v", err)
			// Continue with local delete - graceful degradation
		}
	}

	// Delete from local database
	if err := h.orgRepo.Delete(orgID); err != nil {
		log.Printf("Failed to delete organization: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete organization",
		})
	}

	return c.JSON(fiber.Map{
		"message": "organization deleted",
	})
}

// ListOrganizations lists organizations for the current user
func (h *OrganizationHandler) ListOrganizations(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	orgs, err := h.orgRepo.GetUserOrganizations(userID)
	if err != nil {
		log.Printf("Failed to list organizations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list organizations",
		})
	}

	result := make([]fiber.Map, len(orgs))
	for i, org := range orgs {
		workosOrgID := ""
		if org.WorkOSOrganizationID.Valid {
			workosOrgID = org.WorkOSOrganizationID.String
		}
		result[i] = fiber.Map{
			"id":                     org.ID,
			"name":                   org.Name,
			"workos_organization_id": workosOrgID,
			"created_at":             org.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"organizations": result,
	})
}

// GetMembers retrieves members of an organization
func (h *OrganizationHandler) GetMembers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")

	// Check membership
	isMember, err := h.orgRepo.IsMember(orgID, userID)
	if err != nil || !isMember {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	members, err := h.orgRepo.GetMembers(orgID)
	if err != nil {
		log.Printf("Failed to get organization members: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get members",
		})
	}

	result := make([]fiber.Map, len(members))
	for i, m := range members {
		result[i] = fiber.Map{
			"id":         m.ID,
			"user_id":    m.UserID,
			"role":       m.Role,
			"created_at": m.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"members": result,
	})
}

// AddMember adds a member to an organization
func (h *OrganizationHandler) AddMember(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")

	// Check admin role
	role, err := h.orgRepo.GetMemberRole(orgID, userID)
	if err != nil || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}

	if req.Role == "" {
		req.Role = "member"
	}

	member, err := h.orgRepo.AddMember(orgID, req.UserID, req.Role)
	if err != nil {
		log.Printf("Failed to add organization member: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to add member",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         member.ID,
		"user_id":    member.UserID,
		"role":       member.Role,
		"created_at": member.CreatedAt,
	})
}

// RemoveMember removes a member from an organization
func (h *OrganizationHandler) RemoveMember(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Params("id")
	memberUserID := c.Params("userId")

	// Check admin role
	role, err := h.orgRepo.GetMemberRole(orgID, userID)
	if err != nil || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.orgRepo.RemoveMember(orgID, memberUserID); err != nil {
		log.Printf("Failed to remove organization member: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove member",
		})
	}

	return c.JSON(fiber.Map{
		"message": "member removed",
	})
}

// HandleWorkOSWebhook handles incoming WorkOS webhooks
func (h *OrganizationHandler) HandleWorkOSWebhook(c *fiber.Ctx) error {
	if h.workosClient == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "workos integration not configured",
		})
	}

	// Get signature and timestamp headers
	signature := c.Get("WorkOS-Signature")
	timestamp := c.Get("WorkOS-Timestamp")

	if signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing WorkOS-Signature header",
		})
	}

	// Read the raw body
	body := c.Body()

	// Verify signature
	if !h.workosClient.VerifyWebhookSignature(body, signature, timestamp) {
		log.Printf("WorkOS webhook signature verification failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid signature",
		})
	}

	// Parse the event
	event, err := h.workosClient.ParseWebhookEvent(body)
	if err != nil {
		log.Printf("Failed to parse WorkOS webhook event: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse event",
		})
	}

	log.Printf("Received WorkOS webhook: event=%s, id=%s", event.Event, event.ID)

	// Process the event based on type
	switch event.Event {
	case workos.EventOrganizationCreated:
		if err := h.handleOrganizationCreated(event); err != nil {
			log.Printf("Failed to handle organization.created: %v", err)
		}

	case workos.EventOrganizationUpdated:
		if err := h.handleOrganizationUpdated(event); err != nil {
			log.Printf("Failed to handle organization.updated: %v", err)
		}

	case workos.EventOrganizationDeleted:
		if err := h.handleOrganizationDeleted(event); err != nil {
			log.Printf("Failed to handle organization.deleted: %v", err)
		}

	default:
		log.Printf("Unhandled WorkOS event type: %s", event.Event)
	}

	return c.JSON(fiber.Map{
		"message": "webhook received",
		"id":      event.ID,
	})
}

// handleOrganizationCreated handles organization.created webhook events
func (h *OrganizationHandler) handleOrganizationCreated(event *workos.WebhookEvent) error {
	// Check if we already have this organization
	existing, err := h.orgRepo.GetByWorkOSID(event.Data.ID)
	if err != nil {
		return err
	}

	if existing != nil {
		log.Printf("Organization %s already exists locally", event.Data.ID)
		return nil
	}

	// Create the organization locally
	_, err = h.orgRepo.Create(event.Data.Name, event.Data.ID)
	if err != nil {
		return err
	}

	log.Printf("Created local organization from WorkOS: %s (%s)", event.Data.Name, event.Data.ID)
	return nil
}

// handleOrganizationUpdated handles organization.updated webhook events
func (h *OrganizationHandler) handleOrganizationUpdated(event *workos.WebhookEvent) error {
	// Find the local organization
	org, err := h.orgRepo.GetByWorkOSID(event.Data.ID)
	if err != nil {
		return err
	}

	if org == nil {
		log.Printf("No local organization found for WorkOS ID %s", event.Data.ID)
		return nil
	}

	// Update the organization name
	if err := h.orgRepo.Update(org.ID, event.Data.Name, event.Data.ID); err != nil {
		return err
	}

	log.Printf("Updated local organization from WorkOS: %s", event.Data.ID)
	return nil
}

// handleOrganizationDeleted handles organization.deleted webhook events
func (h *OrganizationHandler) handleOrganizationDeleted(event *workos.WebhookEvent) error {
	// Delete the local organization
	if err := h.orgRepo.DeleteByWorkOSID(event.Data.ID); err != nil {
		return err
	}

	log.Printf("Deleted local organization from WorkOS: %s", event.Data.ID)
	return nil
}

// SyncFromWorkOS syncs all organizations from WorkOS
func (h *OrganizationHandler) SyncFromWorkOS(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	// This is an admin-only operation - you may want to add additional checks here
	_ = userID

	if h.workosClient == nil || !h.workosClient.Enabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "workos integration not configured",
		})
	}

	var synced, created, updated int
	var cursor string

	for {
		orgs, nextCursor, err := h.workosClient.ListOrganizations(100, cursor)
		if err != nil {
			log.Printf("Failed to list organizations from WorkOS: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to sync from workos",
			})
		}

		for _, org := range orgs {
			synced++

			existing, err := h.orgRepo.GetByWorkOSID(org.ID)
			if err != nil {
				log.Printf("Error checking organization %s: %v", org.ID, err)
				continue
			}

			if existing == nil {
				// Create new organization
				if _, err := h.orgRepo.Create(org.Name, org.ID); err != nil {
					log.Printf("Error creating organization %s: %v", org.ID, err)
				} else {
					created++
				}
			} else {
				// Update existing organization
				if err := h.orgRepo.Update(existing.ID, org.Name, org.ID); err != nil {
					log.Printf("Error updating organization %s: %v", org.ID, err)
				} else {
					updated++
				}
			}
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	log.Printf("WorkOS sync complete: synced=%d, created=%d, updated=%d", synced, created, updated)

	return c.JSON(fiber.Map{
		"message": "sync complete",
		"synced":  synced,
		"created": created,
		"updated": updated,
	})
}
