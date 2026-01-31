package handlers

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// ToolsCatalogHandler handles tool catalog endpoints
type ToolsCatalogHandler struct {
	toolRepo *repository.ToolRepository
}

// NewToolsCatalogHandler creates a new tools catalog handler
func NewToolsCatalogHandler(toolRepo *repository.ToolRepository) *ToolsCatalogHandler {
	return &ToolsCatalogHandler{
		toolRepo: toolRepo,
	}
}

// ToolResponse represents a tool in API responses
type ToolResponse struct {
	ID               string `json:"id"`
	DisplayName      string `json:"display_name"`
	SlugName         string `json:"slug_name"`
	Description      string `json:"description,omitempty"`
	IsModel          bool   `json:"is_model"`
	IsBuiltin        bool   `json:"is_builtin"`
	ProviderID       string `json:"provider_id,omitempty"`
	ParametersSchema string `json:"parameters_schema,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CreateToolRequest represents a request to create a tool
type CreateToolRequest struct {
	DisplayName      string `json:"display_name"`
	SlugName         string `json:"slug_name"`
	Description      string `json:"description,omitempty"`
	IsModel          bool   `json:"is_model"`
	ProviderID       string `json:"provider_id,omitempty"`
	ParametersSchema string `json:"parameters_schema,omitempty"`
}

// UpdateToolRequest represents a request to update a tool
type UpdateToolRequest struct {
	DisplayName      *string `json:"display_name,omitempty"`
	SlugName         *string `json:"slug_name,omitempty"`
	Description      *string `json:"description,omitempty"`
	IsModel          *bool   `json:"is_model,omitempty"`
	ProviderID       *string `json:"provider_id,omitempty"`
	ParametersSchema *string `json:"parameters_schema,omitempty"`
}

// toolToResponse converts a Tool to ToolResponse
func toolToResponse(tool *repository.Tool) ToolResponse {
	return ToolResponse{
		ID:               tool.ID,
		DisplayName:      tool.DisplayName,
		SlugName:         tool.SlugName,
		Description:      tool.Description,
		IsModel:          tool.IsModel,
		IsBuiltin:        tool.IsBuiltin,
		ProviderID:       tool.ProviderID,
		ParametersSchema: tool.ParametersSchema,
		CreatedAt:        tool.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        tool.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListTools returns all tools
func (h *ToolsCatalogHandler) ListTools(c *fiber.Ctx) error {
	// Optional query parameters for filtering
	providerID := c.Query("provider_id")
	modelsOnly := c.Query("models_only") == "true"

	var tools []*repository.Tool
	var err error

	if modelsOnly {
		tools, err = h.toolRepo.ListModels()
	} else if providerID != "" {
		tools, err = h.toolRepo.ListByProvider(providerID)
	} else {
		tools, err = h.toolRepo.List()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list tools",
		})
	}

	response := make([]ToolResponse, len(tools))
	for i, tool := range tools {
		response[i] = toolToResponse(tool)
	}

	return c.JSON(fiber.Map{
		"tools": response,
	})
}

// GetTool returns a tool by ID
func (h *ToolsCatalogHandler) GetTool(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tool id is required",
		})
	}

	tool, err := h.toolRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool",
		})
	}

	if tool == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tool not found",
		})
	}

	return c.JSON(toolToResponse(tool))
}

// GetToolBySlug returns a tool by slug name
func (h *ToolsCatalogHandler) GetToolBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tool slug is required",
		})
	}

	tool, err := h.toolRepo.GetBySlug(slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool",
		})
	}

	if tool == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tool not found",
		})
	}

	return c.JSON(toolToResponse(tool))
}

// CreateTool creates a new tool (admin only)
func (h *ToolsCatalogHandler) CreateTool(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateToolRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "display_name is required",
		})
	}

	if req.SlugName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug_name is required",
		})
	}

	// Validate slug format
	if !isValidSlug(req.SlugName) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug_name must contain only lowercase letters, numbers, and hyphens",
		})
	}

	// Validate JSON schema if provided
	if req.ParametersSchema != "" {
		if err := validateJSONSchema(req.ParametersSchema); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid parameters_schema: " + err.Error(),
			})
		}
	}

	// Check for duplicate slug
	existing, err := h.toolRepo.GetBySlug(req.SlugName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check slug uniqueness",
		})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "a tool with this slug_name already exists",
		})
	}

	tool := &repository.Tool{
		DisplayName:      req.DisplayName,
		SlugName:         req.SlugName,
		Description:      req.Description,
		IsModel:          req.IsModel,
		ProviderID:       req.ProviderID,
		ParametersSchema: req.ParametersSchema,
	}

	if err := h.toolRepo.Create(tool); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create tool",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toolToResponse(tool))
}

// UpdateTool updates an existing tool (admin only)
func (h *ToolsCatalogHandler) UpdateTool(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tool id is required",
		})
	}

	// Get existing tool
	tool, err := h.toolRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool",
		})
	}

	if tool == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tool not found",
		})
	}

	// Prevent editing builtin tools
	if tool.IsBuiltin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "cannot edit builtin tools",
		})
	}

	var req UpdateToolRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate slug if being updated
	if req.SlugName != nil && *req.SlugName != tool.SlugName {
		if !isValidSlug(*req.SlugName) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "slug_name must contain only lowercase letters, numbers, and hyphens",
			})
		}
		// Check for duplicate slug
		existing, err := h.toolRepo.GetBySlug(*req.SlugName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check slug uniqueness",
			})
		}
		if existing != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "a tool with this slug_name already exists",
			})
		}
	}

	// Validate JSON schema if being updated
	if req.ParametersSchema != nil && *req.ParametersSchema != "" {
		if err := validateJSONSchema(*req.ParametersSchema); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid parameters_schema: " + err.Error(),
			})
		}
	}

	// Apply updates
	if req.DisplayName != nil {
		tool.DisplayName = *req.DisplayName
	}
	if req.SlugName != nil {
		tool.SlugName = *req.SlugName
	}
	if req.Description != nil {
		tool.Description = *req.Description
	}
	if req.IsModel != nil {
		tool.IsModel = *req.IsModel
	}
	if req.ProviderID != nil {
		tool.ProviderID = *req.ProviderID
	}
	if req.ParametersSchema != nil {
		tool.ParametersSchema = *req.ParametersSchema
	}

	if err := h.toolRepo.Update(tool); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update tool",
		})
	}

	return c.JSON(toolToResponse(tool))
}

// DeleteTool deletes a tool (admin only)
func (h *ToolsCatalogHandler) DeleteTool(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tool id is required",
		})
	}

	// Get the tool to check if it's builtin
	tool, err := h.toolRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool",
		})
	}

	if tool == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tool not found",
		})
	}

	// Prevent deleting builtin tools
	if tool.IsBuiltin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "cannot delete builtin tools",
		})
	}

	if err := h.toolRepo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete tool",
		})
	}

	return c.JSON(fiber.Map{
		"message": "tool deleted successfully",
	})
}

// isValidSlug checks if a slug contains only lowercase letters, numbers, and hyphens
func isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// validateJSONSchema validates that the provided string is valid JSON
func validateJSONSchema(schema string) error {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return nil
	}
	var js map[string]interface{}
	return json.Unmarshal([]byte(schema), &js)
}
