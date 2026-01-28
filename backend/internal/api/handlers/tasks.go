package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/database/repository"
)

// TasksHandler handles task-related HTTP requests
type TasksHandler struct {
	taskRepo     *repository.AgentTaskRepository
	agentManager *agent.Manager
}

// NewTasksHandler creates a new tasks handler
func NewTasksHandler(taskRepo *repository.AgentTaskRepository, agentManager *agent.Manager) *TasksHandler {
	return &TasksHandler{
		taskRepo:     taskRepo,
		agentManager: agentManager,
	}
}

// TaskResponse represents a task in API responses
type TaskResponse struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Prompt       string                 `json:"prompt"`
	Context      string                 `json:"context,omitempty"`
	Priority     int                    `json:"priority"`
	Status       string                 `json:"status"`
	AgentConfig  map[string]interface{} `json:"agent_config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	CallbackURL  string                 `json:"callback_url,omitempty"`
	CreatedAt    int64                  `json:"created_at"`
	StartedAt    *int64                 `json:"started_at,omitempty"`
	CompletedAt  *int64                 `json:"completed_at,omitempty"`
}

// toResponse converts a repository task to API response
func toTaskResponse(task *repository.AgentTask) TaskResponse {
	resp := TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Prompt:      task.Prompt,
		Context:     task.Context,
		Priority:    task.Priority,
		Status:      task.Status,
		AgentConfig: task.AgentConfig,
		Metadata:    task.Metadata,
		Result:      task.Result,
		Error:       task.Error,
		CallbackURL: task.CallbackURL,
		CreatedAt:   task.CreatedAt.UnixMilli(),
	}
	if task.StartedAt != nil {
		ts := task.StartedAt.UnixMilli()
		resp.StartedAt = &ts
	}
	if task.CompletedAt != nil {
		ts := task.CompletedAt.UnixMilli()
		resp.CompletedAt = &ts
	}
	return resp
}

// ListTasks lists tasks for the authenticated user
// GET /api/v1/tasks?status=pending&limit=20&offset=0
func (h *TasksHandler) ListTasks(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	status := c.Query("status")
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var tasks []*repository.AgentTask
	if status != "" {
		tasks, err = h.taskRepo.ListByUserIDAndStatus(userID, status, limit, offset)
	} else {
		tasks, err = h.taskRepo.ListByUserID(userID, limit, offset)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list tasks",
		})
	}

	// Get total count
	total, _ := h.taskRepo.CountByUserID(userID)

	response := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = toTaskResponse(task)
	}

	return c.JSON(fiber.Map{
		"tasks":  response,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetTask gets a single task by ID
// GET /api/v1/tasks/:id
func (h *TasksHandler) GetTask(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	taskID := c.Params("id")

	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "task id is required",
		})
	}

	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get task",
		})
	}
	if task == nil || task.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.JSON(toTaskResponse(task))
}

// CancelTask cancels a pending task
// DELETE /api/v1/tasks/:id
func (h *TasksHandler) CancelTask(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	taskID := c.Params("id")

	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "task id is required",
		})
	}

	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get task",
		})
	}
	if task == nil || task.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	// Can only cancel pending or running tasks
	if task.Status != "pending" && task.Status != "running" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "can only cancel pending or running tasks",
		})
	}

	// Update status to cancelled
	if err := h.taskRepo.UpdateStatus(taskID, "cancelled"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to cancel task",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "task cancelled",
	})
}

// RetryTask retries a failed task
// POST /api/v1/tasks/:id/retry
func (h *TasksHandler) RetryTask(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	taskID := c.Params("id")

	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "task id is required",
		})
	}

	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get task",
		})
	}
	if task == nil || task.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	// Can only retry failed or cancelled tasks
	if task.Status != "failed" && task.Status != "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "can only retry failed or cancelled tasks",
		})
	}

	// Reset task status to pending
	task.Status = "pending"
	task.Error = ""
	task.Result = nil
	task.StartedAt = nil
	task.CompletedAt = nil

	if err := h.taskRepo.Update(task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retry task",
		})
	}

	// Re-submit to agent pool if available
	if h.agentManager != nil {
		agentTask := &agent.Task{
			ID:       task.ID,
			UserID:   task.UserID,
			Prompt:   task.Prompt,
			Context:  task.Context,
			Priority: agent.TaskPriority(task.Priority),
			Status:   agent.TaskStatusPending,
		}

		// Convert agent config back
		if task.AgentConfig != nil {
			config := agent.AgentConfig{}
			if provider, ok := task.AgentConfig["provider"].(string); ok {
				config.Provider = provider
			}
			if model, ok := task.AgentConfig["model"].(string); ok {
				config.Model = model
			}
			if temp, ok := task.AgentConfig["temperature"].(float64); ok {
				config.Temperature = temp
			}
			if maxTokens, ok := task.AgentConfig["max_tokens"].(float64); ok {
				config.MaxTokens = int(maxTokens)
			}
			agentTask.AgentConfig = &config
		}

		// Note: In a real implementation, you'd re-queue this through the manager
		// For now, just update the database status
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "task requeued for retry",
		"task":    toTaskResponse(task),
	})
}

// GetTaskStats gets task statistics for the authenticated user
// GET /api/v1/tasks/stats
func (h *TasksHandler) GetTaskStats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	total, _ := h.taskRepo.CountByUserID(userID)
	pending, _ := h.taskRepo.CountByStatus("pending")
	running, _ := h.taskRepo.CountByStatus("running")
	completed, _ := h.taskRepo.CountByStatus("completed")
	failed, _ := h.taskRepo.CountByStatus("failed")

	return c.JSON(fiber.Map{
		"total":     total,
		"pending":   pending,
		"running":   running,
		"completed": completed,
		"failed":    failed,
	})
}
