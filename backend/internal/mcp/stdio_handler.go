package mcp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// StdioHandler handles stdio MCP server management endpoints
type StdioHandler struct {
	client *StdioClient
	repo   *StdioRepository
}

// NewStdioHandler creates a new stdio handler
func NewStdioHandler(client *StdioClient, repo *StdioRepository) *StdioHandler {
	return &StdioHandler{
		client: client,
		repo:   repo,
	}
}

// RegisterRoutes registers stdio MCP routes
func (h *StdioHandler) RegisterRoutes(app fiber.Router) {
	mcp := app.Group("/mcp/stdio")

	mcp.Get("/servers", h.ListServers)
	mcp.Post("/servers", h.AddServer)
	mcp.Get("/servers/:id", h.GetServer)
	mcp.Put("/servers/:id", h.UpdateServer)
	mcp.Delete("/servers/:id", h.DeleteServer)
	mcp.Post("/servers/:id/start", h.StartServer)
	mcp.Post("/servers/:id/stop", h.StopServer)
	mcp.Post("/servers/:id/restart", h.RestartServer)
	mcp.Get("/servers/:id/tools", h.GetServerTools)
}

// AddServerRequest represents a request to add a stdio MCP server
type AddStdioServerRequest struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env,omitempty"`
}

// ListServers returns all stdio MCP servers for the user
func (h *StdioHandler) ListServers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	servers, err := h.repo.GetByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Enrich with runtime status
	result := make([]fiber.Map, len(servers))
	for i, s := range servers {
		// Check if running in client
		clientServer := h.client.GetServer(s.ID)
		running := clientServer != nil && clientServer.IsRunning()

		toolCount := 0
		if clientServer != nil {
			toolCount = len(clientServer.Tools)
		}

		result[i] = fiber.Map{
			"id":         s.ID,
			"name":       s.Name,
			"command":    s.Command,
			"args":       s.Args,
			"env":        s.Env,
			"enabled":    s.Enabled,
			"running":    running,
			"tool_count": toolCount,
			"created_at": s.CreatedAt,
			"updated_at": s.UpdatedAt,
			"last_error": s.LastError,
		}
	}

	return c.JSON(fiber.Map{
		"servers": result,
	})
}

// AddServer adds a new stdio MCP server
func (h *StdioHandler) AddServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req AddStdioServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if req.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Command is required",
		})
	}

	now := time.Now()
	server := &StdioServer{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Command:   req.Command,
		Args:      req.Args,
		Env:       req.Env,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to database
	if err := h.repo.Create(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Try to start the server
	if err := h.client.AddServer(server); err != nil {
		server.LastError = err.Error()
		h.repo.Update(server)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"server": fiber.Map{
				"id":         server.ID,
				"name":       server.Name,
				"command":    server.Command,
				"args":       server.Args,
				"enabled":    server.Enabled,
				"running":    false,
				"last_error": server.LastError,
			},
			"warning": "Server created but failed to start: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"server": fiber.Map{
			"id":         server.ID,
			"name":       server.Name,
			"command":    server.Command,
			"args":       server.Args,
			"enabled":    server.Enabled,
			"running":    true,
			"tool_count": len(server.Tools),
		},
	})
}

// GetServer returns a specific stdio MCP server
func (h *StdioHandler) GetServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	// Check runtime status
	clientServer := h.client.GetServer(serverID)
	running := clientServer != nil && clientServer.IsRunning()

	var tools []fiber.Map
	if clientServer != nil {
		tools = make([]fiber.Map, len(clientServer.Tools))
		for i, t := range clientServer.Tools {
			tools[i] = fiber.Map{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			}
		}
	}

	return c.JSON(fiber.Map{
		"server": fiber.Map{
			"id":         server.ID,
			"name":       server.Name,
			"command":    server.Command,
			"args":       server.Args,
			"env":        server.Env,
			"enabled":    server.Enabled,
			"running":    running,
			"tools":      tools,
			"created_at": server.CreatedAt,
			"updated_at": server.UpdatedAt,
			"last_error": server.LastError,
		},
	})
}

// UpdateServerRequest represents a request to update a stdio MCP server
type UpdateStdioServerRequest struct {
	Name    *string   `json:"name,omitempty"`
	Command *string   `json:"command,omitempty"`
	Args    *[]string `json:"args,omitempty"`
	Env     *[]string `json:"env,omitempty"`
	Enabled *bool     `json:"enabled,omitempty"`
}

// UpdateServer updates a stdio MCP server
func (h *StdioHandler) UpdateServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	var req UpdateStdioServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Command != nil {
		server.Command = *req.Command
	}
	if req.Args != nil {
		server.Args = *req.Args
	}
	if req.Env != nil {
		server.Env = *req.Env
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
	}
	server.UpdatedAt = time.Now()

	if err := h.repo.Update(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// If command/args changed, restart the server
	needsRestart := req.Command != nil || req.Args != nil || req.Env != nil
	if needsRestart {
		h.client.StopServer(serverID)
		if server.Enabled {
			h.client.AddServer(server)
		}
	} else if req.Enabled != nil {
		if *req.Enabled {
			h.client.AddServer(server)
		} else {
			h.client.StopServer(serverID)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// DeleteServer deletes a stdio MCP server
func (h *StdioHandler) DeleteServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	// Stop and remove from client
	h.client.RemoveServer(serverID)

	// Delete from database
	if err := h.repo.Delete(serverID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// StartServer starts a stdio MCP server
func (h *StdioHandler) StartServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	// Check if already in client
	clientServer := h.client.GetServer(serverID)
	if clientServer != nil && clientServer.IsRunning() {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Server already running",
		})
	}

	// Add to client (which starts it)
	if err := h.client.AddServer(server); err != nil {
		server.LastError = err.Error()
		h.repo.Update(server)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start server: " + err.Error(),
		})
	}

	// Clear error on success
	server.LastError = ""
	h.repo.Update(server)

	return c.JSON(fiber.Map{
		"success":    true,
		"tool_count": len(server.Tools),
	})
}

// StopServer stops a stdio MCP server
func (h *StdioHandler) StopServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	if err := h.client.StopServer(serverID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// RestartServer restarts a stdio MCP server
func (h *StdioHandler) RestartServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	// Stop if running
	h.client.StopServer(serverID)
	h.client.RemoveServer(serverID)

	// Start fresh
	if err := h.client.AddServer(server); err != nil {
		server.LastError = err.Error()
		h.repo.Update(server)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restart server: " + err.Error(),
		})
	}

	// Clear error on success
	server.LastError = ""
	h.repo.Update(server)

	return c.JSON(fiber.Map{
		"success":    true,
		"tool_count": len(h.client.GetServer(serverID).Tools),
	})
}

// GetServerTools returns the tools available from a specific server
func (h *StdioHandler) GetServerTools(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	dbServer, err := h.repo.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if dbServer == nil || dbServer.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	server := h.client.GetServer(serverID)
	if server == nil || !server.IsRunning() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Server not running",
			"running": false,
		})
	}

	tools := make([]fiber.Map, len(server.Tools))
	for i, t := range server.Tools {
		tools[i] = fiber.Map{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.Parameters,
		}
	}

	return c.JSON(fiber.Map{
		"tools": tools,
	})
}
