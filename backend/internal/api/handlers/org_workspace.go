package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// OrgWorkspaceHandler handles organization workspace endpoints
type OrgWorkspaceHandler struct {
	repo *repository.OrgWorkspaceRepository
}

// NewOrgWorkspaceHandler creates a new organization workspace handler
func NewOrgWorkspaceHandler(repo *repository.OrgWorkspaceRepository) *OrgWorkspaceHandler {
	return &OrgWorkspaceHandler{repo: repo}
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name                 string `json:"name"`
	OrganizationID       string `json:"organization_id"`
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
	if req.OrganizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	ws := &repository.OrgWorkspace{
		Name:                 req.Name,
		OrganizationID:       req.OrganizationID,
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

	return c.JSON(toResponse(ws))
}

// List retrieves all workspaces for an organization
func (h *OrgWorkspaceHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	orgID := c.Query("organization_id")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id query parameter is required",
		})
	}

	workspaces, err := h.repo.ListByOrganizationID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list workspaces",
		})
	}

	responses := make([]*WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		responses[i] = toResponse(ws)
	}

	return c.JSON(fiber.Map{
		"workspaces": responses,
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

	if err := h.repo.UpdateBranch(id, req.Branch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update workspace branch",
		})
	}

	ws.CurrentBranch = req.Branch
	return c.JSON(toResponse(ws))
}
