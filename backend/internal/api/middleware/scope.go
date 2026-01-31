package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// ScopeContext holds the extracted scope information
type ScopeContext struct {
	UserID   string
	TeamID   string
	OrgID    string
	TeamIDs  []string // All teams user belongs to
	OrgIDs   []string // All organizations user belongs to
}

// TeamScopeMiddleware extracts team context from request and verifies membership
func TeamScopeMiddleware(
	teamRepo *repository.TeamRepository,
	orgRepo *repository.OrganizationRepository,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		// Extract team ID from header, query, or path
		teamID := c.Get("X-Team-ID")
		if teamID == "" {
			teamID = c.Query("team_id")
		}
		if teamID == "" {
			teamID = c.Params("teamId")
		}

		// Extract org ID from header, query, or path
		orgID := c.Get("X-Organization-ID")
		if orgID == "" {
			orgID = c.Query("organization_id")
		}
		if orgID == "" {
			orgID = c.Params("orgId")
		}

		// Build scope context
		scopeCtx := &ScopeContext{
			UserID: userID,
			TeamID: teamID,
			OrgID:  orgID,
		}

		// Get all user team IDs for filtering
		if teamRepo != nil {
			teamIDs, err := teamRepo.GetTeamIDsForUser(userID)
			if err == nil {
				scopeCtx.TeamIDs = teamIDs
			}
		}

		// Get all user org IDs for filtering
		if orgRepo != nil {
			orgs, err := orgRepo.GetUserOrganizations(userID)
			if err == nil {
				orgIDs := make([]string, len(orgs))
				for i, org := range orgs {
					orgIDs[i] = org.ID
				}
				scopeCtx.OrgIDs = orgIDs
			}
		}

		// Verify team membership if team ID is provided
		if teamID != "" && teamRepo != nil {
			isMember, err := teamRepo.IsMember(teamID, userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to verify team membership",
				})
			}
			if !isMember {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "not a member of this team",
				})
			}
		}

		// Verify org membership if org ID is provided
		if orgID != "" && orgRepo != nil {
			isMember, err := orgRepo.IsMember(orgID, userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to verify organization membership",
				})
			}
			if !isMember {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "not a member of this organization",
				})
			}
		}

		// Store scope context
		c.Locals("scopeContext", scopeCtx)
		c.Locals("teamID", teamID)
		c.Locals("orgID", orgID)

		return c.Next()
	}
}

// RequireTeamMiddleware requires a team ID to be present and verified
func RequireTeamMiddleware(teamRepo *repository.TeamRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		// Extract team ID
		teamID := c.Get("X-Team-ID")
		if teamID == "" {
			teamID = c.Query("team_id")
		}
		if teamID == "" {
			teamID = c.Params("teamId")
		}

		if teamID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "team_id is required",
			})
		}

		// Verify membership
		isMember, err := teamRepo.IsMember(teamID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to verify team membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "not a member of this team",
			})
		}

		c.Locals("teamID", teamID)
		return c.Next()
	}
}

// RequireOrgMiddleware requires an organization ID to be present and verified
func RequireOrgMiddleware(orgRepo *repository.OrganizationRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		// Extract org ID
		orgID := c.Get("X-Organization-ID")
		if orgID == "" {
			orgID = c.Query("organization_id")
		}
		if orgID == "" {
			orgID = c.Params("orgId")
		}

		if orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "organization_id is required",
			})
		}

		// Verify membership
		isMember, err := orgRepo.IsMember(orgID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to verify organization membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "not a member of this organization",
			})
		}

		c.Locals("orgID", orgID)
		return c.Next()
	}
}

// PermissionMiddleware checks permissions for a specific resource and action
func PermissionMiddleware(
	permService *security.PermissionService,
	resource security.ResourceType,
	action security.Action,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		teamID := GetTeamID(c)
		orgID := GetOrgID(c)

		// Build access context
		ctx := security.AccessContext{
			UserID:   userID,
			TeamID:   teamID,
			OrgID:    orgID,
			Resource: resource,
			Action:   action,
		}

		// Check team-level permission if team context exists
		if teamID != "" {
			if err := permService.RequireTeamPermission(ctx); err != nil {
				if security.IsPermissionError(err) {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "permission denied",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to check permissions",
				})
			}
			return c.Next()
		}

		// Check org-level permission if org context exists
		if orgID != "" {
			if err := permService.RequireOrgPermission(ctx); err != nil {
				if security.IsPermissionError(err) {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "permission denied",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to check permissions",
				})
			}
			return c.Next()
		}

		// Check resource-specific permission
		resourceID := c.Params("id")
		if resourceID == "" {
			resourceID = c.Params("resourceId")
		}

		if resourceID != "" {
			if err := permService.RequirePermission(ctx, resourceID); err != nil {
				if security.IsPermissionError(err) {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "permission denied",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to check permissions",
				})
			}
		}

		return c.Next()
	}
}

// TeamAdminMiddleware requires team admin permissions
func TeamAdminMiddleware(
	teamRepo *repository.TeamRepository,
	roleRepo *repository.RoleRepository,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		teamID := GetTeamID(c)
		if teamID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "team_id is required",
			})
		}

		// Get member role in team
		roleID, roleName, err := teamRepo.GetMemberRole(teamID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get member role",
			})
		}
		if roleID == "" && roleName == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "not a member of this team",
			})
		}

		// Check if role has admin permission
		if roleName == string(security.RoleTeamAdmin) ||
			roleName == string(security.RoleOrgAdmin) ||
			roleName == string(security.RoleOrgOwner) {
			return c.Next()
		}

		// Check custom role permissions
		if roleRepo != nil && roleID != "" {
			permissions, err := roleRepo.GetPermissionsForRole(roleID)
			if err == nil && security.HasPermission(permissions, security.ResourceTeam, security.ActionAdmin) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "team admin permission required",
		})
	}
}

// OrgAdminMiddleware requires organization admin permissions
func OrgAdminMiddleware(
	orgRepo *repository.OrganizationRepository,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		orgID := GetOrgID(c)
		if orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "organization_id is required",
			})
		}

		// Get member role in org
		role, err := orgRepo.GetMemberRole(orgID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get member role",
			})
		}
		if role == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "not a member of this organization",
			})
		}

		// Check for admin role
		if role == "admin" || role == "owner" ||
			role == string(security.RoleOrgAdmin) ||
			role == string(security.RoleOrgOwner) {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "organization admin permission required",
		})
	}
}

// GetTeamID gets the team ID from the context
func GetTeamID(c *fiber.Ctx) string {
	teamID, ok := c.Locals("teamID").(string)
	if !ok {
		return ""
	}
	return teamID
}

// GetOrgID gets the organization ID from the context
func GetOrgID(c *fiber.Ctx) string {
	orgID, ok := c.Locals("orgID").(string)
	if !ok {
		return ""
	}
	return orgID
}

// GetScopeContext gets the full scope context from the request
func GetScopeContext(c *fiber.Ctx) *ScopeContext {
	ctx, ok := c.Locals("scopeContext").(*ScopeContext)
	if !ok {
		return nil
	}
	return ctx
}
