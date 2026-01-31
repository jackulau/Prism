package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// TeamHandler handles team-related endpoints
type TeamHandler struct {
	teamRepo *repository.TeamRepository
	roleRepo *repository.RoleRepository
	orgRepo  *repository.OrganizationRepository
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(
	teamRepo *repository.TeamRepository,
	roleRepo *repository.RoleRepository,
	orgRepo *repository.OrganizationRepository,
) *TeamHandler {
	return &TeamHandler{
		teamRepo: teamRepo,
		roleRepo: roleRepo,
		orgRepo:  orgRepo,
	}
}

// CreateTeam creates a new team within an organization
func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
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

	// Create the team
	team, err := h.teamRepo.Create(orgID, req.Name, req.Description)
	if err != nil {
		log.Printf("Failed to create team: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create team",
		})
	}

	// Add the creator as team admin
	_, err = h.teamRepo.AddMember(team.ID, userID, string(security.RoleTeamAdmin), string(security.RoleTeamAdmin))
	if err != nil {
		log.Printf("Failed to add creator as team admin: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":              team.ID,
		"organization_id": team.OrganizationID,
		"name":            team.Name,
		"description":     team.Description,
		"created_at":      team.CreatedAt,
	})
}

// GetTeam retrieves a team by ID
func (h *TeamHandler) GetTeam(c *fiber.Ctx) error {
	teamID := c.Params("id")

	team, err := h.teamRepo.GetByID(teamID)
	if err != nil || team == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "team not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":              team.ID,
		"organization_id": team.OrganizationID,
		"name":            team.Name,
		"description":     team.Description,
		"created_at":      team.CreatedAt,
		"updated_at":      team.UpdatedAt,
	})
}

// UpdateTeam updates a team's name and description
func (h *TeamHandler) UpdateTeam(c *fiber.Ctx) error {
	teamID := c.Params("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
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

	if err := h.teamRepo.Update(teamID, req.Name, req.Description); err != nil {
		log.Printf("Failed to update team: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update team",
		})
	}

	return c.JSON(fiber.Map{
		"id":          teamID,
		"name":        req.Name,
		"description": req.Description,
	})
}

// DeleteTeam deletes a team
func (h *TeamHandler) DeleteTeam(c *fiber.Ctx) error {
	teamID := c.Params("id")

	if err := h.teamRepo.Delete(teamID); err != nil {
		log.Printf("Failed to delete team: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete team",
		})
	}

	return c.JSON(fiber.Map{
		"message": "team deleted",
	})
}

// ListTeams lists teams for an organization
func (h *TeamHandler) ListTeams(c *fiber.Ctx) error {
	orgID := middleware.GetOrgID(c)

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	teams, err := h.teamRepo.ListByOrganization(orgID, limit, offset)
	if err != nil {
		log.Printf("Failed to list teams: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list teams",
		})
	}

	result := make([]fiber.Map, len(teams))
	for i, team := range teams {
		result[i] = fiber.Map{
			"id":              team.ID,
			"organization_id": team.OrganizationID,
			"name":            team.Name,
			"description":     team.Description,
			"member_count":    team.MemberCount,
			"created_at":      team.CreatedAt,
		}
	}

	total, _ := h.teamRepo.CountByOrganization(orgID)

	return c.JSON(fiber.Map{
		"teams": result,
		"total": total,
	})
}

// ListUserTeams lists teams the current user belongs to
func (h *TeamHandler) ListUserTeams(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	teams, err := h.teamRepo.GetUserTeams(userID)
	if err != nil {
		log.Printf("Failed to list user teams: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list teams",
		})
	}

	result := make([]fiber.Map, len(teams))
	for i, team := range teams {
		result[i] = fiber.Map{
			"id":              team.ID,
			"organization_id": team.OrganizationID,
			"name":            team.Name,
			"description":     team.Description,
			"created_at":      team.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"teams": result,
	})
}

// GetMembers retrieves members of a team
func (h *TeamHandler) GetMembers(c *fiber.Ctx) error {
	teamID := c.Params("id")

	members, err := h.teamRepo.GetMembers(teamID)
	if err != nil {
		log.Printf("Failed to get team members: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get members",
		})
	}

	result := make([]fiber.Map, len(members))
	for i, m := range members {
		result[i] = fiber.Map{
			"id":         m.ID,
			"user_id":    m.UserID,
			"role_id":    m.RoleID,
			"role":       m.Role,
			"created_at": m.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"members": result,
	})
}

// AddMember adds a member to a team
func (h *TeamHandler) AddMember(c *fiber.Ctx) error {
	teamID := c.Params("id")

	var req struct {
		UserID string `json:"user_id"`
		RoleID string `json:"role_id"`
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

	// Default to team_member role if not specified
	if req.Role == "" {
		req.Role = string(security.RoleTeamMember)
	}
	if req.RoleID == "" {
		req.RoleID = req.Role
	}

	// Verify the team exists
	team, err := h.teamRepo.GetByID(teamID)
	if err != nil || team == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "team not found",
		})
	}

	// Verify the user is a member of the organization
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(team.OrganizationID, req.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to verify organization membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "user must be a member of the organization",
			})
		}
	}

	member, err := h.teamRepo.AddMember(teamID, req.UserID, req.RoleID, req.Role)
	if err != nil {
		log.Printf("Failed to add team member: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to add member",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         member.ID,
		"team_id":    member.TeamID,
		"user_id":    member.UserID,
		"role_id":    member.RoleID,
		"role":       member.Role,
		"created_at": member.CreatedAt,
	})
}

// RemoveMember removes a member from a team
func (h *TeamHandler) RemoveMember(c *fiber.Ctx) error {
	teamID := c.Params("id")
	userIDToRemove := c.Params("userId")

	if err := h.teamRepo.RemoveMember(teamID, userIDToRemove); err != nil {
		log.Printf("Failed to remove team member: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove member",
		})
	}

	return c.JSON(fiber.Map{
		"message": "member removed",
	})
}

// UpdateMemberRole updates a member's role in a team
func (h *TeamHandler) UpdateMemberRole(c *fiber.Ctx) error {
	teamID := c.Params("id")
	userIDToUpdate := c.Params("userId")

	var req struct {
		RoleID string `json:"role_id"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Role == "" && req.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "role or role_id is required",
		})
	}

	if req.RoleID == "" {
		req.RoleID = req.Role
	}

	if err := h.teamRepo.UpdateMemberRole(teamID, userIDToUpdate, req.RoleID, req.Role); err != nil {
		log.Printf("Failed to update member role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update member role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "role updated",
		"user_id": userIDToUpdate,
		"role_id": req.RoleID,
		"role":    req.Role,
	})
}

// ListRoles lists available roles for a team
func (h *TeamHandler) ListRoles(c *fiber.Ctx) error {
	orgID := middleware.GetOrgID(c)

	if h.roleRepo == nil {
		// Return predefined roles if no role repo
		predefined := security.PredefinedRoles()
		roles := make([]fiber.Map, 0)

		for roleType, permissions := range predefined {
			permList := make([]fiber.Map, len(permissions))
			for i, p := range permissions {
				permList[i] = fiber.Map{
					"resource": p.Resource,
					"action":   p.Action,
					"scope":    p.Scope,
				}
			}
			roles = append(roles, fiber.Map{
				"id":          string(roleType),
				"name":        string(roleType),
				"type":        string(roleType),
				"is_system":   true,
				"permissions": permList,
			})
		}

		return c.JSON(fiber.Map{
			"roles": roles,
		})
	}

	// Get system roles + custom org roles
	allRoles, err := h.roleRepo.ListByOrganization(orgID)
	if err != nil {
		log.Printf("Failed to list roles: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list roles",
		})
	}

	result := make([]fiber.Map, len(allRoles))
	for i, role := range allRoles {
		permList := make([]fiber.Map, len(role.Permissions))
		for j, p := range role.Permissions {
			permList[j] = fiber.Map{
				"resource": p.Resource,
				"action":   p.Action,
				"scope":    p.Scope,
			}
		}
		result[i] = fiber.Map{
			"id":              role.ID,
			"name":            role.Name,
			"description":     role.Description,
			"type":            role.Type,
			"organization_id": role.OrganizationID,
			"is_system":       role.IsSystem,
			"permissions":     permList,
		}
	}

	return c.JSON(fiber.Map{
		"roles": result,
	})
}

// CreateRole creates a custom role for an organization
func (h *TeamHandler) CreateRole(c *fiber.Ctx) error {
	orgID := middleware.GetOrgID(c)

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	if h.roleRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "role management not available",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions []struct {
			Resource string `json:"resource"`
			Action   string `json:"action"`
			Scope    string `json:"scope"`
		} `json:"permissions"`
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

	// Convert permissions
	permissions := make([]security.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		permissions[i] = security.Permission{
			Resource: security.ResourceType(p.Resource),
			Action:   security.Action(p.Action),
			Scope:    security.Scope(p.Scope),
		}
	}

	role, err := h.roleRepo.Create(orgID, req.Name, req.Description, permissions)
	if err != nil {
		log.Printf("Failed to create role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create role",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":              role.ID,
		"name":            role.Name,
		"description":     role.Description,
		"type":            role.Type,
		"organization_id": role.OrganizationID,
		"is_system":       role.IsSystem,
	})
}

// DeleteRole deletes a custom role
func (h *TeamHandler) DeleteRole(c *fiber.Ctx) error {
	roleID := c.Params("roleId")

	if h.roleRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "role management not available",
		})
	}

	if err := h.roleRepo.Delete(roleID); err != nil {
		log.Printf("Failed to delete role: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "role deleted",
	})
}
