package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
)

// ToolExecutionsHandler handles tool execution log endpoints
type ToolExecutionsHandler struct {
	execRepo *repository.ToolExecutionRepository
}

// NewToolExecutionsHandler creates a new tool executions handler
func NewToolExecutionsHandler(execRepo *repository.ToolExecutionRepository) *ToolExecutionsHandler {
	return &ToolExecutionsHandler{
		execRepo: execRepo,
	}
}

// ToolExecutionResponse represents a tool execution in API responses
type ToolExecutionResponse struct {
	ID              string `json:"id"`
	MessageID       string `json:"message_id"`
	ToolName        string `json:"tool_name"`
	Parameters      string `json:"parameters"`
	Result          string `json:"result,omitempty"`
	Status          string `json:"status"`
	ExecutionTimeMS int64  `json:"execution_time_ms,omitempty"`
	ContainerID     string `json:"container_id,omitempty"`
	CreatedAt       string `json:"created_at"`
}

// toolExecutionToResponse converts a ToolExecution to ToolExecutionResponse
func toolExecutionToResponse(exec *repository.ToolExecution) ToolExecutionResponse {
	return ToolExecutionResponse{
		ID:              exec.ID,
		MessageID:       exec.MessageID,
		ToolName:        exec.ToolName,
		Parameters:      exec.Parameters,
		Result:          exec.Result,
		Status:          exec.Status,
		ExecutionTimeMS: exec.ExecutionTimeMS,
		ContainerID:     exec.ContainerID,
		CreatedAt:       exec.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListExecutions returns tool executions with optional filtering
// GET /api/v1/tools/executions?tool_name=X&status=Y&limit=N&offset=M
func (h *ToolExecutionsHandler) ListExecutions(c *fiber.Ctx) error {
	filter := repository.ToolExecutionFilter{
		ToolName: c.Query("tool_name"),
		Status:   c.Query("status"),
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid limit parameter",
			})
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // Default limit
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid offset parameter",
			})
		}
		filter.Offset = offset
	}

	executions, err := h.execRepo.List(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list tool executions",
		})
	}

	// Get total count for pagination
	total, err := h.execRepo.Count(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count tool executions",
		})
	}

	response := make([]ToolExecutionResponse, len(executions))
	for i, exec := range executions {
		response[i] = toolExecutionToResponse(exec)
	}

	return c.JSON(fiber.Map{
		"executions": response,
		"total":      total,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	})
}

// GetExecution returns a single tool execution by ID
func (h *ToolExecutionsHandler) GetExecution(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "execution id is required",
		})
	}

	exec, err := h.execRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool execution",
		})
	}

	if exec == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tool execution not found",
		})
	}

	return c.JSON(toolExecutionToResponse(exec))
}

// GetDistinctToolNames returns all distinct tool names from executions
func (h *ToolExecutionsHandler) GetDistinctToolNames(c *fiber.Ctx) error {
	names, err := h.execRepo.GetDistinctToolNames()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tool names",
		})
	}

	return c.JSON(fiber.Map{
		"tool_names": names,
	})
}
