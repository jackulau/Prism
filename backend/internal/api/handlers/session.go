package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/services/session"
)

// SessionHandler handles session management endpoints
type SessionHandler struct {
	sessionService *session.Service
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *session.Service) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// ListSessions returns all sessions for the current user
// GET /api/v1/sessions
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	currentSessionID := middleware.GetSessionID(c)

	sessions, err := h.sessionService.ListUserSessions(userID, currentSessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list sessions",
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}

// TerminateSession terminates a specific session
// DELETE /api/v1/sessions/:id
func (h *SessionHandler) TerminateSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "session id is required",
		})
	}

	// Don't allow terminating the current session via this endpoint
	currentSessionID := middleware.GetSessionID(c)
	if sessionID == currentSessionID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot terminate current session, use logout instead",
		})
	}

	if err := h.sessionService.Terminate(userID, sessionID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "session not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "session terminated successfully",
	})
}

// TerminateOtherSessions terminates all sessions except the current one
// DELETE /api/v1/sessions/others
func (h *SessionHandler) TerminateOtherSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	currentSessionID := middleware.GetSessionID(c)
	if currentSessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "current session not found",
		})
	}

	if err := h.sessionService.TerminateOthers(userID, currentSessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to terminate sessions",
		})
	}

	return c.JSON(fiber.Map{
		"message": "all other sessions terminated successfully",
	})
}

// TerminateAllSessions terminates all sessions (logout everywhere)
// DELETE /api/v1/sessions
func (h *SessionHandler) TerminateAllSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.sessionService.TerminateAll(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to terminate sessions",
		})
	}

	return c.JSON(fiber.Map{
		"message": "all sessions terminated successfully",
	})
}
