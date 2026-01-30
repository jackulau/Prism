package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
)

// BuildHistoryHandler handles build history related HTTP requests
type BuildHistoryHandler struct {
	repo *repository.BuildHistoryRepository
}

// NewBuildHistoryHandler creates a new build history handler
func NewBuildHistoryHandler(repo *repository.BuildHistoryRepository) *BuildHistoryHandler {
	return &BuildHistoryHandler{
		repo: repo,
	}
}

// BuildHistoryResponse represents a build history record in API responses
type BuildHistoryResponse struct {
	ID             string  `json:"id"`
	WorkspaceID    *string `json:"workspace_id,omitempty"`
	OrgWorkspaceID *string `json:"org_workspace_id,omitempty"`
	UserID         string  `json:"user_id"`
	Command        string  `json:"command"`
	Status         string  `json:"status"`
	ExitCode       *int    `json:"exit_code,omitempty"`
	StartedAt      int64   `json:"started_at"`
	CompletedAt    *int64  `json:"completed_at,omitempty"`
	DurationMs     *int64  `json:"duration_ms,omitempty"`
	CreatedAt      int64   `json:"created_at"`
}

// BuildLogResponse represents a build log entry in API responses
type BuildLogResponse struct {
	ID        string `json:"id"`
	BuildID   string `json:"build_id"`
	Stream    string `json:"stream"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// toBuildHistoryResponse converts a repository build history to API response
func toBuildHistoryResponse(build *repository.BuildHistory) BuildHistoryResponse {
	resp := BuildHistoryResponse{
		ID:             build.ID,
		WorkspaceID:    build.WorkspaceID,
		OrgWorkspaceID: build.OrgWorkspaceID,
		UserID:         build.UserID,
		Command:        build.Command,
		Status:         string(build.Status),
		ExitCode:       build.ExitCode,
		StartedAt:      build.StartedAt.UnixMilli(),
		DurationMs:     build.DurationMs,
		CreatedAt:      build.CreatedAt.UnixMilli(),
	}
	if build.CompletedAt != nil {
		ts := build.CompletedAt.UnixMilli()
		resp.CompletedAt = &ts
	}
	return resp
}

// toBuildLogResponse converts a repository build log to API response
func toBuildLogResponse(log *repository.BuildLog) BuildLogResponse {
	return BuildLogResponse{
		ID:        log.ID,
		BuildID:   log.BuildID,
		Stream:    string(log.Stream),
		Content:   log.Content,
		Timestamp: log.Timestamp.UnixMilli(),
	}
}

// List lists build history for the authenticated user
// GET /api/v1/builds?workspace_id=xxx&limit=20&offset=0
func (h *BuildHistoryHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	workspaceID := c.Query("workspace_id")
	orgWorkspaceID := c.Query("org_workspace_id")
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

	var builds []*repository.BuildHistory

	if orgWorkspaceID != "" {
		builds, err = h.repo.ListByOrgWorkspaceID(orgWorkspaceID, limit, offset)
	} else if workspaceID != "" {
		builds, err = h.repo.ListByWorkspaceID(workspaceID, limit, offset)
	} else {
		builds, err = h.repo.ListByUserID(userID, limit, offset)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list builds",
		})
	}

	// Get total count
	total, _ := h.repo.CountByUserID(userID)

	response := make([]BuildHistoryResponse, len(builds))
	for i, build := range builds {
		response[i] = toBuildHistoryResponse(build)
	}

	return c.JSON(fiber.Map{
		"builds": response,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Get retrieves a single build history record by ID
// GET /api/v1/builds/:id
func (h *BuildHistoryHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	buildID := c.Params("id")

	if buildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "build id is required",
		})
	}

	build, err := h.repo.GetByID(buildID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build",
		})
	}
	if build == nil || build.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build not found",
		})
	}

	return c.JSON(toBuildHistoryResponse(build))
}

// GetLogs retrieves logs for a specific build
// GET /api/v1/builds/:id/logs?since=<timestamp>
func (h *BuildHistoryHandler) GetLogs(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	buildID := c.Params("id")

	if buildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "build id is required",
		})
	}

	// Verify the build exists and belongs to the user
	build, err := h.repo.GetByID(buildID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build",
		})
	}
	if build == nil || build.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build not found",
		})
	}

	// Check for since parameter (for streaming/polling)
	sinceStr := c.Query("since")
	var logs []*repository.BuildLog

	if sinceStr != "" {
		sinceMs, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid since parameter",
			})
		}
		since := time.UnixMilli(sinceMs)
		logs, err = h.repo.GetLogsSince(buildID, since)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get build logs",
			})
		}
	} else {
		logs, err = h.repo.GetLogs(buildID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get build logs",
			})
		}
	}

	response := make([]BuildLogResponse, len(logs))
	for i, log := range logs {
		response[i] = toBuildLogResponse(log)
	}

	return c.JSON(fiber.Map{
		"logs":    response,
		"build":   toBuildHistoryResponse(build),
		"running": build.Status == repository.BuildStatusRunning,
	})
}

// Delete removes a build history record and its logs
// DELETE /api/v1/builds/:id
func (h *BuildHistoryHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	buildID := c.Params("id")

	if buildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "build id is required",
		})
	}

	// Verify the build exists and belongs to the user
	build, err := h.repo.GetByID(buildID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build",
		})
	}
	if build == nil || build.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build not found",
		})
	}

	// Cannot delete running builds
	if build.Status == repository.BuildStatusRunning {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot delete a running build, cancel it first",
		})
	}

	if err := h.repo.Delete(buildID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete build",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "build deleted",
	})
}

// Cancel cancels a running build
// POST /api/v1/builds/:id/cancel
func (h *BuildHistoryHandler) Cancel(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	buildID := c.Params("id")

	if buildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "build id is required",
		})
	}

	// Verify the build exists and belongs to the user
	build, err := h.repo.GetByID(buildID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build",
		})
	}
	if build == nil || build.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build not found",
		})
	}

	// Can only cancel pending or running builds
	if build.Status != repository.BuildStatusPending && build.Status != repository.BuildStatusRunning {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "can only cancel pending or running builds",
		})
	}

	// Update status to cancelled
	now := time.Now()
	exitCode := -1
	var durationMs *int64
	if build.Status == repository.BuildStatusRunning {
		duration := now.Sub(build.StartedAt).Milliseconds()
		durationMs = &duration
	}

	if err := h.repo.UpdateStatus(buildID, repository.BuildStatusCancelled, &exitCode, &now, durationMs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to cancel build",
		})
	}

	// TODO: In a real implementation, you'd also need to signal the sandbox
	// to actually stop the running process

	return c.JSON(fiber.Map{
		"success": true,
		"message": "build cancelled",
	})
}
