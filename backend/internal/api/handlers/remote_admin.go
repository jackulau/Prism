package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/remote"
)

// RemoteAdminHandler handles remote session administration endpoints
type RemoteAdminHandler struct {
	sessionManager   *remote.SessionManager
	heartbeatHandler *remote.HeartbeatHandler
	reconnectHandler *remote.ReconnectHandler
}

// NewRemoteAdminHandler creates a new remote admin handler
func NewRemoteAdminHandler(
	sessionManager *remote.SessionManager,
	heartbeatHandler *remote.HeartbeatHandler,
	reconnectHandler *remote.ReconnectHandler,
) *RemoteAdminHandler {
	return &RemoteAdminHandler{
		sessionManager:   sessionManager,
		heartbeatHandler: heartbeatHandler,
		reconnectHandler: reconnectHandler,
	}
}

// ListSessions returns all active remote sessions
// GET /api/v1/remote/sessions
func (h *RemoteAdminHandler) ListSessions(c *fiber.Ctx) error {
	sessions := h.sessionManager.ListSessions()

	infos := make([]remote.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		infos = append(infos, session.GetInfo())
	}

	return c.JSON(fiber.Map{
		"sessions": infos,
		"count":    len(infos),
	})
}

// GetSession returns details of a specific session
// GET /api/v1/remote/sessions/:id
func (h *RemoteAdminHandler) GetSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session ID is required",
		})
	}

	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		if err == remote.ErrSessionNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	info := session.GetInfo()

	// Get heartbeat status if available
	var heartbeatStatus *remote.HeartbeatStatus
	if h.heartbeatHandler != nil {
		heartbeatStatus = h.heartbeatHandler.GetHeartbeatStatus(sessionID)
	}

	// Get reconnect attempt info if available
	var attemptInfo *remote.AttemptInfo
	if h.reconnectHandler != nil {
		attemptInfo = h.reconnectHandler.GetAttemptInfo(sessionID)
	}

	return c.JSON(fiber.Map{
		"session":          info,
		"heartbeat_status": heartbeatStatus,
		"reconnect_info":   attemptInfo,
	})
}

// KickSessionRequest represents a request to kick a session
type KickSessionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// KickSession forcefully disconnects a session
// POST /api/v1/remote/sessions/:id/kick
func (h *RemoteAdminHandler) KickSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session ID is required",
		})
	}

	var req KickSessionRequest
	if err := c.BodyParser(&req); err != nil {
		// Body is optional, use default reason
		req.Reason = "kicked by admin"
	}

	err := h.sessionManager.CloseSession(sessionID)
	if err != nil {
		if err == remote.ErrSessionNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "session kicked",
		"reason":  req.Reason,
	})
}

// GetStats returns session statistics
// GET /api/v1/remote/sessions/stats
func (h *RemoteAdminHandler) GetStats(c *fiber.Ctx) error {
	stats := h.sessionManager.GetStats()

	response := fiber.Map{
		"total_sessions":        stats.TotalSessions,
		"active_sessions":       stats.ActiveSessions,
		"disconnected_sessions": stats.DisconnectedSessions,
		"reconnecting_sessions": stats.ReconnectingSessions,
		"total_users":           stats.TotalUsers,
	}

	// Add heartbeat config if available
	if h.heartbeatHandler != nil {
		config := h.heartbeatHandler.GetConfig()
		response["heartbeat_config"] = fiber.Map{
			"interval_ms":     config.Interval.Milliseconds(),
			"timeout_ms":      config.Timeout.Milliseconds(),
			"grace_period_ms": config.GracePeriod.Milliseconds(),
		}
	}

	return c.JSON(response)
}

// ListUserSessions returns all sessions for a specific user
// GET /api/v1/remote/users/:user_id/sessions
func (h *RemoteAdminHandler) ListUserSessions(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}

	sessions := h.sessionManager.GetUserSessions(userID)

	infos := make([]remote.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		infos = append(infos, session.GetInfo())
	}

	return c.JSON(fiber.Map{
		"user_id":  userID,
		"sessions": infos,
		"count":    len(infos),
	})
}

// GetMySessions returns all sessions for the current authenticated user
// GET /api/v1/remote/sessions/me
func (h *RemoteAdminHandler) GetMySessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessions := h.sessionManager.GetUserSessions(userID)

	infos := make([]remote.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		infos = append(infos, session.GetInfo())
	}

	return c.JSON(fiber.Map{
		"sessions": infos,
		"count":    len(infos),
	})
}
