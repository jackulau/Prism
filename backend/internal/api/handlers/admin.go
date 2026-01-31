package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// AdminHandler handles admin-only endpoints
type AdminHandler struct {
	userRepo *repository.UserRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
	}
}

// AdminUserResponse represents a user in admin responses (without sensitive data)
type AdminUserResponse struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	GitHubUsername *string `json:"github_username,omitempty"`
	OrganizationID *string `json:"organization_id,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ListUsersResponse is the response for listing users
type ListUsersResponse struct {
	Users      []AdminUserResponse `json:"users"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

// UpdateRoleRequest is the request body for updating a user's role
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// ListUsers returns a paginated list of all users (admin only)
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	// Parse pagination params
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	roleFilter := c.Query("role")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	var users []*repository.User
	var total int
	var err error

	if roleFilter != "" {
		// Filter by role
		if !security.IsValidRole(roleFilter) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid role filter",
			})
		}
		users, err = h.userRepo.GetUsersByRole(roleFilter)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get users",
			})
		}
		total = len(users)
		// Apply pagination manually for filtered results
		start := offset
		end := offset + perPage
		if start > len(users) {
			users = []*repository.User{}
		} else {
			if end > len(users) {
				end = len(users)
			}
			users = users[start:end]
		}
	} else {
		// Get all users with pagination
		users, err = h.userRepo.GetAllUsers(perPage, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get users",
			})
		}
		total, err = h.userRepo.CountAllUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to count users",
			})
		}
	}

	// Convert to response format
	userResponses := make([]AdminUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = toAdminUserResponse(user)
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return c.JSON(ListUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// GetUser returns details for a specific user (admin only)
// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(toAdminUserResponse(user))
}

// UpdateUserRole updates a user's role (admin only)
// PATCH /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}

	// Get current user ID from context
	currentUserID := middleware.GetUserID(c)

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate role
	if !security.IsValidRole(req.Role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "invalid role",
			"valid_roles": []string{string(security.RoleUser), string(security.RoleAdmin)},
		})
	}

	// Get the target user
	targetUser, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	if targetUser == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	// Safety check: prevent demoting the last admin
	if targetUser.Role == string(security.RoleAdmin) && req.Role != string(security.RoleAdmin) {
		adminCount, err := h.userRepo.CountUsersByRole(string(security.RoleAdmin))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check admin count",
			})
		}
		if adminCount <= 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cannot demote the last admin",
			})
		}
	}

	// Prevent self-demotion (optional safety feature)
	if userID == currentUserID && targetUser.Role == string(security.RoleAdmin) && req.Role != string(security.RoleAdmin) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot demote yourself",
		})
	}

	// Update the role
	if err := h.userRepo.SetUserRole(userID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user role",
		})
	}

	// Return updated user
	updatedUser, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get updated user",
		})
	}

	return c.JSON(toAdminUserResponse(updatedUser))
}

// GetStats returns admin statistics
// GET /api/v1/admin/stats
func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	totalUsers, err := h.userRepo.CountAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count users",
		})
	}

	adminCount, err := h.userRepo.CountUsersByRole(string(security.RoleAdmin))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count admins",
		})
	}

	userCount, err := h.userRepo.CountUsersByRole(string(security.RoleUser))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count regular users",
		})
	}

	return c.JSON(fiber.Map{
		"total_users": totalUsers,
		"admin_count": adminCount,
		"user_count":  userCount,
	})
}

// toAdminUserResponse converts a User to AdminUserResponse
func toAdminUserResponse(user *repository.User) AdminUserResponse {
	resp := AdminUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if user.GitHubUsername != "" {
		resp.GitHubUsername = &user.GitHubUsername
	}
	if user.OrganizationID != "" {
		resp.OrganizationID = &user.OrganizationID
	}

	return resp
}
