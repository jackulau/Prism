package mcp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles MCP-related API requests
type Handler struct {
	client     *Client
	repository *Repository
}

// NewHandler creates a new MCP handler
func NewHandler(client *Client, repository *Repository) *Handler {
	return &Handler{
		client:     client,
		repository: repository,
	}
}

// RegisterRoutes registers MCP client routes
func (h *Handler) RegisterRoutes(app fiber.Router) {
	servers := app.Group("/mcp/servers")
	servers.Get("/", h.ListServers)
	servers.Post("/", h.AddServer)
	servers.Get("/:id", h.GetServer)
	servers.Put("/:id", h.UpdateServer)
	servers.Delete("/:id", h.RemoveServer)
	servers.Post("/:id/test", h.TestServer)
	servers.Post("/:id/refresh", h.RefreshServer)
	servers.Post("/:id/enable", h.EnableServer)
	servers.Post("/:id/disable", h.DisableServer)

	// List all available tools
	app.Get("/mcp/tools", h.ListTools)
}

// ListServers lists all MCP servers for the current user
func (h *Handler) ListServers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	servers, err := h.repository.GetByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch servers",
		})
	}

	// Redact API keys
	result := make([]fiber.Map, len(servers))
	for i, server := range servers {
		result[i] = fiber.Map{
			"id":         server.ID,
			"name":       server.Name,
			"url":        server.URL,
			"enabled":    server.Enabled,
			"has_api_key": server.APIKey != "",
			"created_at": server.CreatedAt,
			"updated_at": server.UpdatedAt,
			"last_sync":  server.LastSync,
			"last_error": server.LastError,
		}

		// Include manifest info if available
		if server.Manifest != nil {
			result[i]["manifest"] = fiber.Map{
				"name":        server.Manifest.Name,
				"version":     server.Manifest.Version,
				"description": server.Manifest.Description,
				"tool_count":  len(server.Manifest.Tools),
			}
		}
	}

	return c.JSON(fiber.Map{"servers": result})
}

// AddServerRequest represents a request to add an MCP server
type AddServerRequest struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	APIKey string `json:"api_key,omitempty"`
}

// AddServer adds a new MCP server connection
func (h *Handler) AddServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req AddServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and URL are required",
		})
	}

	// Test connection first
	manifest, err := h.client.TestConnection(req.URL, req.APIKey)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to connect to MCP server: " + err.Error(),
		})
	}

	now := time.Now()
	server := &RemoteServer{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		URL:       req.URL,
		APIKey:    req.APIKey,
		Enabled:   true,
		Manifest:  manifest,
		CreatedAt: now,
		UpdatedAt: now,
		LastSync:  &now,
	}

	// Save to database
	if err := h.repository.Create(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save server",
		})
	}

	// Add to client
	if err := h.client.AddServer(server); err != nil {
		// Clean up database entry
		h.repository.Delete(server.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize server connection",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         server.ID,
		"name":       server.Name,
		"url":        server.URL,
		"enabled":    server.Enabled,
		"manifest": fiber.Map{
			"name":        manifest.Name,
			"version":     manifest.Version,
			"description": manifest.Description,
			"tool_count":  len(manifest.Tools),
		},
		"created_at": server.CreatedAt,
	})
}

// GetServer returns a specific MCP server
func (h *Handler) GetServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	result := fiber.Map{
		"id":         server.ID,
		"name":       server.Name,
		"url":        server.URL,
		"enabled":    server.Enabled,
		"has_api_key": server.APIKey != "",
		"created_at": server.CreatedAt,
		"updated_at": server.UpdatedAt,
		"last_sync":  server.LastSync,
		"last_error": server.LastError,
	}

	if server.Manifest != nil {
		result["manifest"] = server.Manifest
	}

	return c.JSON(result)
}

// UpdateServerRequest represents a request to update an MCP server
type UpdateServerRequest struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	APIKey string `json:"api_key,omitempty"`
}

// UpdateServer updates an MCP server
func (h *Handler) UpdateServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	var req UpdateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.URL != "" {
		server.URL = req.URL
	}
	if req.APIKey != "" {
		server.APIKey = req.APIKey
	}
	server.UpdatedAt = time.Now()

	if err := h.repository.Update(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update server",
		})
	}

	// Refresh manifest if URL changed
	if req.URL != "" {
		h.client.RefreshManifest(serverID)
	}

	return c.JSON(fiber.Map{"success": true})
}

// RemoveServer removes an MCP server connection
func (h *Handler) RemoveServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	h.client.RemoveServer(serverID)
	if err := h.repository.Delete(serverID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete server",
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// TestServer tests connection to an MCP server
func (h *Handler) TestServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	manifest, err := h.client.TestConnection(server.URL, server.APIKey)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"manifest": fiber.Map{
			"name":        manifest.Name,
			"version":     manifest.Version,
			"description": manifest.Description,
			"tool_count":  len(manifest.Tools),
		},
	})
}

// RefreshServer refreshes the manifest for an MCP server
func (h *Handler) RefreshServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	if err := h.client.RefreshManifest(serverID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get updated server info
	updatedServer := h.client.GetServer(serverID)
	if updatedServer == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get updated server",
		})
	}

	// Save updated manifest to database
	server.Manifest = updatedServer.Manifest
	server.LastSync = updatedServer.LastSync
	server.LastError = updatedServer.LastError
	h.repository.Update(server)

	return c.JSON(fiber.Map{
		"success":   true,
		"last_sync": server.LastSync,
		"manifest": fiber.Map{
			"name":        server.Manifest.Name,
			"version":     server.Manifest.Version,
			"description": server.Manifest.Description,
			"tool_count":  len(server.Manifest.Tools),
		},
	})
}

// EnableServer enables an MCP server
func (h *Handler) EnableServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	server.Enabled = true
	server.UpdatedAt = time.Now()

	if err := h.repository.Update(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update server",
		})
	}

	h.client.EnableServer(serverID)

	return c.JSON(fiber.Map{"success": true, "enabled": true})
}

// DisableServer disables an MCP server
func (h *Handler) DisableServer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	serverID := c.Params("id")

	server, err := h.repository.GetByID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch server",
		})
	}

	if server == nil || server.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Server not found",
		})
	}

	server.Enabled = false
	server.UpdatedAt = time.Now()

	if err := h.repository.Update(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update server",
		})
	}

	h.client.DisableServer(serverID)

	return c.JSON(fiber.Map{"success": true, "enabled": false})
}

// ListTools lists all available tools from all enabled MCP servers
func (h *Handler) ListTools(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	tools := h.client.GetAllTools(userID)

	result := make([]fiber.Map, len(tools))
	for i, tool := range tools {
		result[i] = fiber.Map{
			"server_id":   tool.ServerID,
			"server_name": tool.ServerName,
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		}
	}

	return c.JSON(fiber.Map{"tools": result})
}
