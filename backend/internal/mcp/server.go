package mcp

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/tools"
)

// ToolDefinition represents an MCP tool definition
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Manifest represents the MCP server manifest
type Manifest struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Description string           `json:"description"`
	Tools       []ToolDefinition `json:"tools"`
}

// ToolExecutionRequest represents a tool execution request
type ToolExecutionRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolExecutionResponse represents a tool execution response
type ToolExecutionResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIKey represents an MCP API key
type APIKey struct {
	ID          string
	UserID      string
	Key         string
	Name        string
	Permissions []string
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}

// Server exposes Prism tools via MCP protocol
type Server struct {
	toolRegistry *tools.Registry
	apiKeys      map[string]*APIKey // key -> APIKey
	mu           sync.RWMutex
}

// NewServer creates a new MCP server
func NewServer(toolRegistry *tools.Registry) *Server {
	return &Server{
		toolRegistry: toolRegistry,
		apiKeys:      make(map[string]*APIKey),
	}
}

// RegisterRoutes registers MCP server routes
func (s *Server) RegisterRoutes(app fiber.Router) {
	mcp := app.Group("/mcp")

	// Public manifest endpoint
	mcp.Get("/manifest", s.GetManifest)

	// Protected tool execution
	mcp.Post("/tools/:name", s.AuthMiddleware, s.ExecuteTool)

	// API key management (requires user auth)
	mcp.Get("/api-keys", s.ListAPIKeys)
	mcp.Post("/api-keys", s.CreateAPIKey)
	mcp.Delete("/api-keys/:id", s.DeleteAPIKey)
}

// GetManifest returns the MCP server manifest
func (s *Server) GetManifest(c *fiber.Ctx) error {
	allTools := s.toolRegistry.List()

	toolDefs := make([]ToolDefinition, 0, len(allTools))
	for _, tool := range allTools {
		params := tool.Parameters()
		toolDefs = append(toolDefs, ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: map[string]interface{}{
				"type":       params.Type,
				"properties": params.Properties,
				"required":   params.Required,
			},
		})
	}

	manifest := Manifest{
		Name:        "Prism MCP Server",
		Version:     "1.0.0",
		Description: "Prism AI agent tools exposed via MCP protocol",
		Tools:       toolDefs,
	}

	return c.JSON(manifest)
}

// AuthMiddleware validates MCP API keys
func (s *Server) AuthMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-MCP-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key required",
		})
	}

	s.mu.RLock()
	key, exists := s.apiKeys[apiKey]
	s.mu.RUnlock()

	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	// Update last used timestamp
	s.mu.Lock()
	now := time.Now()
	key.LastUsedAt = &now
	s.mu.Unlock()

	c.Locals("mcp_api_key", key)
	return c.Next()
}

// ExecuteTool executes a tool via MCP
func (s *Server) ExecuteTool(c *fiber.Ctx) error {
	toolName := c.Params("name")

	tool, ok := s.toolRegistry.Get(toolName)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(ToolExecutionResponse{
			Success: false,
			Error:   "Tool not found: " + toolName,
		})
	}

	var req ToolExecutionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ToolExecutionResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
	}

	// Check if tool requires confirmation
	if tool.RequiresConfirmation() {
		return c.Status(fiber.StatusForbidden).JSON(ToolExecutionResponse{
			Success: false,
			Error:   "Tool requires user confirmation and cannot be executed via MCP",
		})
	}

	// Execute tool
	result, err := tool.Execute(c.Context(), req.Parameters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ToolExecutionResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(ToolExecutionResponse{
		Success: true,
		Result:  result,
	})
}

// ListAPIKeys lists all API keys for a user
func (s *Server) ListAPIKeys(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]fiber.Map, 0)
	for _, key := range s.apiKeys {
		if key.UserID == userID {
			keys = append(keys, fiber.Map{
				"id":          key.ID,
				"name":        key.Name,
				"permissions": key.Permissions,
				"created_at":  key.CreatedAt,
				"last_used":   key.LastUsedAt,
				// Don't expose the actual key
				"key_preview": key.Key[:8] + "..." + key.Key[len(key.Key)-4:],
			})
		}
	}

	return c.JSON(fiber.Map{"api_keys": keys})
}

// CreateAPIKeyRequest represents an API key creation request
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// CreateAPIKey creates a new MCP API key
func (s *Server) CreateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateAPIKeyRequest
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

	// Generate API key
	keyID := uuid.New().String()
	apiKey := "mcp_" + uuid.New().String()

	key := &APIKey{
		ID:          keyID,
		UserID:      userID,
		Key:         apiKey,
		Name:        req.Name,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.apiKeys[apiKey] = key
	s.mu.Unlock()

	// Return the key only once - user must save it
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":          key.ID,
		"name":        key.Name,
		"key":         apiKey, // Only shown once
		"permissions": key.Permissions,
		"created_at":  key.CreatedAt,
	})
}

// DeleteAPIKey deletes an MCP API key
func (s *Server) DeleteAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	keyID := c.Params("id")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and delete the key
	for apiKey, key := range s.apiKeys {
		if key.ID == keyID && key.UserID == userID {
			delete(s.apiKeys, apiKey)
			return c.JSON(fiber.Map{"success": true})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "API key not found",
	})
}

// LoadAPIKeysFromJSON loads API keys from JSON (for persistence)
func (s *Server) LoadAPIKeysFromJSON(data []byte) error {
	var keys []*APIKey
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.apiKeys[key.Key] = key
	}

	return nil
}

// ExportAPIKeysToJSON exports API keys to JSON (for persistence)
func (s *Server) ExportAPIKeysToJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*APIKey, 0, len(s.apiKeys))
	for _, key := range s.apiKeys {
		keys = append(keys, key)
	}

	return json.Marshal(keys)
}
