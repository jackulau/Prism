package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// RemoteServer represents a connected MCP server
type RemoteServer struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	APIKey    string     `json:"api_key,omitempty"` // Stored encrypted
	Enabled   bool       `json:"enabled"`
	Manifest  *Manifest  `json:"manifest,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastError string     `json:"last_error,omitempty"`
	LastSync  *time.Time `json:"last_sync,omitempty"`
}

// RemoteTool represents a tool from a remote MCP server
type RemoteTool struct {
	ServerID    string
	ServerName  string
	Name        string
	Description string
	Parameters  llm.JSONSchema
}

// Client connects to external MCP servers
type Client struct {
	servers    map[string]*RemoteServer // serverID -> server
	toolCache  map[string][]*RemoteTool // serverID -> tools
	httpClient *http.Client
	mu         sync.RWMutex
}

// NewClient creates a new MCP client
func NewClient() *Client {
	return &Client{
		servers:   make(map[string]*RemoteServer),
		toolCache: make(map[string][]*RemoteTool),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AddServer adds a new MCP server connection
func (c *Client) AddServer(server *RemoteServer) error {
	c.mu.Lock()
	c.servers[server.ID] = server
	c.mu.Unlock()

	// Fetch manifest
	if err := c.RefreshManifest(server.ID); err != nil {
		server.LastError = err.Error()
		return err
	}

	return nil
}

// RemoveServer removes an MCP server connection
func (c *Client) RemoveServer(serverID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.servers, serverID)
	delete(c.toolCache, serverID)
}

// GetServer returns a server by ID
func (c *Client) GetServer(serverID string) *RemoteServer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.servers[serverID]
}

// GetUserServers returns all servers for a user
func (c *Client) GetUserServers(userID string) []*RemoteServer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	servers := make([]*RemoteServer, 0)
	for _, server := range c.servers {
		if server.UserID == userID {
			servers = append(servers, server)
		}
	}
	return servers
}

// RefreshManifest fetches the manifest from a remote MCP server
func (c *Client) RefreshManifest(serverID string) error {
	c.mu.RLock()
	server, exists := c.servers[serverID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/manifest", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if server.APIKey != "" {
		req.Header.Set("X-MCP-API-Key", server.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("manifest fetch failed: %s - %s", resp.Status, string(body))
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	// Update server and cache tools
	c.mu.Lock()
	server.Manifest = &manifest
	now := time.Now()
	server.LastSync = &now
	server.LastError = ""

	// Convert to RemoteTools
	tools := make([]*RemoteTool, 0, len(manifest.Tools))
	for _, toolDef := range manifest.Tools {
		tool := &RemoteTool{
			ServerID:    serverID,
			ServerName:  server.Name,
			Name:        toolDef.Name,
			Description: toolDef.Description,
		}

		// Parse parameters
		if params, ok := toolDef.Parameters["properties"].(map[string]interface{}); ok {
			tool.Parameters.Type = "object"
			tool.Parameters.Properties = make(map[string]llm.JSONProperty)
			for name, propRaw := range params {
				if prop, ok := propRaw.(map[string]interface{}); ok {
					tool.Parameters.Properties[name] = llm.JSONProperty{
						Type:        getStringField(prop, "type"),
						Description: getStringField(prop, "description"),
					}
				}
			}
		}
		if required, ok := toolDef.Parameters["required"].([]interface{}); ok {
			for _, r := range required {
				if s, ok := r.(string); ok {
					tool.Parameters.Required = append(tool.Parameters.Required, s)
				}
			}
		}

		tools = append(tools, tool)
	}
	c.toolCache[serverID] = tools
	c.mu.Unlock()

	return nil
}

// GetAllTools returns all tools from all enabled servers for a user
func (c *Client) GetAllTools(userID string) []*RemoteTool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	allTools := make([]*RemoteTool, 0)
	for serverID, server := range c.servers {
		if server.UserID == userID && server.Enabled {
			if tools, exists := c.toolCache[serverID]; exists {
				allTools = append(allTools, tools...)
			}
		}
	}
	return allTools
}

// ExecuteTool executes a tool on a remote MCP server
func (c *Client) ExecuteTool(ctx context.Context, serverID, toolName string, params map[string]interface{}) (interface{}, error) {
	c.mu.RLock()
	server, exists := c.servers[serverID]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	if !server.Enabled {
		return nil, fmt.Errorf("server is disabled: %s", server.Name)
	}

	// Prepare request body
	reqBody, err := json.Marshal(ToolExecutionRequest{Parameters: params})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/tools/"+toolName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if server.APIKey != "" {
		req.Header.Set("X-MCP-API-Key", server.APIKey)
	}

	// Set body
	req.Body = io.NopCloser(jsonReader(reqBody))
	req.ContentLength = int64(len(reqBody))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	defer resp.Body.Close()

	var result ToolExecutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("tool execution error: %s", result.Error)
	}

	return result.Result, nil
}

// TestConnection tests the connection to an MCP server
func (c *Client) TestConnection(url, apiKey string) (*Manifest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url+"/manifest", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("X-MCP-API-Key", apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("connection test failed: %s - %s", resp.Status, string(body))
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest response: %w", err)
	}

	return &manifest, nil
}

// EnableServer enables an MCP server
func (c *Client) EnableServer(serverID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	server, exists := c.servers[serverID]
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	server.Enabled = true
	server.UpdatedAt = time.Now()
	return nil
}

// DisableServer disables an MCP server
func (c *Client) DisableServer(serverID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	server, exists := c.servers[serverID]
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	server.Enabled = false
	server.UpdatedAt = time.Now()
	return nil
}

// Helper function to get string field from map
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// jsonReader returns an io.Reader for JSON bytes
type jsonReaderType struct {
	data []byte
	pos  int
}

func jsonReader(data []byte) *jsonReaderType {
	return &jsonReaderType{data: data}
}

func (r *jsonReaderType) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
