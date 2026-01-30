package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// OrgWorkspaceHandler handles organization workspace endpoints
type OrgWorkspaceHandler struct {
	repo    *repository.OrgWorkspaceRepository
	orgRepo *repository.OrganizationRepository
}

// NewOrgWorkspaceHandler creates a new organization workspace handler
func NewOrgWorkspaceHandler(repo *repository.OrgWorkspaceRepository, orgRepo *repository.OrganizationRepository) *OrgWorkspaceHandler {
	return &OrgWorkspaceHandler{repo: repo, orgRepo: orgRepo}
}

// checkMembership verifies the user is a member of the organization
func (h *OrgWorkspaceHandler) checkMembership(c *fiber.Ctx, orgID, userID string) (bool, error) {
	if h.orgRepo == nil {
		// Skip membership check if org repo not provided
		return true, nil
	}
	return h.orgRepo.IsMember(orgID, userID)
}

// checkAdminRole verifies the user has admin or owner role in the organization
func (h *OrgWorkspaceHandler) checkAdminRole(c *fiber.Ctx, orgID, userID string) (bool, error) {
	if h.orgRepo == nil {
		// Skip role check if org repo not provided
		return true, nil
	}
	role, err := h.orgRepo.GetMemberRole(orgID, userID)
	if err != nil {
		return false, err
	}
	return role == "admin" || role == "owner", nil
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name                 string `json:"name"`
	GitHubRepositoryName string `json:"github_repository_name,omitempty"`
	WorkerID             string `json:"worker_id,omitempty"`
	CurrentBranch        string `json:"current_branch,omitempty"`
	SlackChannelID       string `json:"slack_channel_id,omitempty"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name                 string `json:"name,omitempty"`
	GitHubRepositoryName string `json:"github_repository_name,omitempty"`
	WorkerID             string `json:"worker_id,omitempty"`
	CurrentBranch        string `json:"current_branch,omitempty"`
	SlackChannelID       string `json:"slack_channel_id,omitempty"`
	SlackMessageTs       string `json:"slack_message_ts,omitempty"`
}

// UpdateBranchRequest represents a request to update workspace branch
type UpdateBranchRequest struct {
	Branch string `json:"branch"`
}

// WorkspaceResponse represents a workspace in API responses
type WorkspaceResponse struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	OrganizationID       string `json:"organization_id"`
	GitHubRepositoryName string `json:"github_repository_name,omitempty"`
	WorkerID             string `json:"worker_id,omitempty"`
	CurrentBranch        string `json:"current_branch,omitempty"`
	SlackChannelID       string `json:"slack_channel_id,omitempty"`
	SlackMessageTs       string `json:"slack_message_ts,omitempty"`
	CreatedAt            int64  `json:"created_at"`
}

// toResponse converts OrgWorkspace to WorkspaceResponse
func toResponse(ws *repository.OrgWorkspace) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:                   ws.ID,
		Name:                 ws.Name,
		OrganizationID:       ws.OrganizationID,
		GitHubRepositoryName: ws.GitHubRepositoryName,
		WorkerID:             ws.WorkerID,
		CurrentBranch:        ws.CurrentBranch,
		SlackChannelID:       ws.SlackChannelID,
		SlackMessageTs:       ws.SlackMessageTs,
		CreatedAt:            ws.CreatedAt.Unix(),
	}
}

// Create creates a new organization workspace
func (h *OrgWorkspaceHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check admin role for create operations
	isAdmin, err := h.checkAdminRole(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin or owner role required",
		})
	}

	var req CreateWorkspaceRequest
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

	// Check name uniqueness within organization
	existing, err := h.repo.GetByName(orgID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check workspace name",
		})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "workspace with this name already exists in the organization",
		})
	}

	ws := &repository.OrgWorkspace{
		Name:                 req.Name,
		OrganizationID:       orgID,
		GitHubRepositoryName: req.GitHubRepositoryName,
		WorkerID:             req.WorkerID,
		CurrentBranch:        req.CurrentBranch,
		SlackChannelID:       req.SlackChannelID,
	}

	created, err := h.repo.Create(ws)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create workspace",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toResponse(created))
}

// Get retrieves a workspace by ID
func (h *OrgWorkspaceHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check membership for read operations
	isMember, err := h.checkMembership(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isMember {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "organization membership required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	ws, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get workspace",
		})
	}
	if ws == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// Verify workspace belongs to the organization
	if ws.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	return c.JSON(toResponse(ws))
}

// List retrieves all workspaces for an organization with pagination
func (h *OrgWorkspaceHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check membership for read operations
	isMember, err := h.checkMembership(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isMember {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "organization membership required",
		})
	}

	// Parse pagination parameters
	limit := 20 // default
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	workspaces, err := h.repo.ListByOrganizationIDWithPagination(orgID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list workspaces",
		})
	}

	total, err := h.repo.CountByOrganizationID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count workspaces",
		})
	}

	responses := make([]*WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		responses[i] = toResponse(ws)
	}

	return c.JSON(fiber.Map{
		"workspaces": responses,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// Update updates a workspace
func (h *OrgWorkspaceHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check admin role for update operations
	isAdmin, err := h.checkAdminRole(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin or owner role required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	var req UpdateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	ws, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get workspace",
		})
	}
	if ws == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// Verify workspace belongs to the organization
	if ws.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// If updating name, check uniqueness
	if req.Name != "" && req.Name != ws.Name {
		existing, err := h.repo.GetByName(orgID, req.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check workspace name",
			})
		}
		if existing != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "workspace with this name already exists in the organization",
			})
		}
	}

	// Apply partial updates
	if req.Name != "" {
		ws.Name = req.Name
	}
	if req.GitHubRepositoryName != "" {
		ws.GitHubRepositoryName = req.GitHubRepositoryName
	}
	if req.WorkerID != "" {
		ws.WorkerID = req.WorkerID
	}
	if req.CurrentBranch != "" {
		ws.CurrentBranch = req.CurrentBranch
	}
	if req.SlackChannelID != "" {
		ws.SlackChannelID = req.SlackChannelID
	}
	if req.SlackMessageTs != "" {
		ws.SlackMessageTs = req.SlackMessageTs
	}

	if err := h.repo.Update(ws); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update workspace",
		})
	}

	return c.JSON(toResponse(ws))
}

// Delete removes a workspace
func (h *OrgWorkspaceHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check admin role for delete operations
	isAdmin, err := h.checkAdminRole(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin or owner role required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	ws, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get workspace",
		})
	}
	if ws == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// Verify workspace belongs to the organization
	if ws.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	if err := h.repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete workspace",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateBranch updates only the branch field
func (h *OrgWorkspaceHandler) UpdateBranch(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Params("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Check admin role for update operations
	isAdmin, err := h.checkAdminRole(c, orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check permissions",
		})
	}
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin or owner role required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	var req UpdateBranchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	ws, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get workspace",
		})
	}
	if ws == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// Verify workspace belongs to the organization
	if ws.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	if err := h.repo.UpdateBranch(id, req.Branch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update workspace branch",
		})
	}

	ws.CurrentBranch = req.Branch
	return c.JSON(toResponse(ws))
}
