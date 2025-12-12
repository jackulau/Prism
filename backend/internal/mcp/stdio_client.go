package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// StdioServer represents a local MCP server connected via stdio
type StdioServer struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Command     string     `json:"command"`      // e.g., "npx"
	Args        []string   `json:"args"`         // e.g., ["-y", "@modelcontextprotocol/server-filesystem", "/path"]
	Env         []string   `json:"env"`          // Additional environment variables
	Enabled     bool       `json:"enabled"`
	Tools       []*RemoteTool `json:"tools,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastError   string     `json:"last_error,omitempty"`

	// Runtime state (not persisted)
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	running    bool
	mu         sync.Mutex
	requestID  int64
	pending    map[int64]chan *jsonRPCResponse
	pendingMu  sync.Mutex
}

// JSON-RPC 2.0 structures
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP protocol structures
type mcpInitializeParams struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    mcpCapabilities   `json:"capabilities"`
	ClientInfo      mcpClientInfo     `json:"clientInfo"`
}

type mcpCapabilities struct {
	Roots   *mcpRootsCapability `json:"roots,omitempty"`
	Sampling *struct{}          `json:"sampling,omitempty"`
}

type mcpRootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type mcpClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type mcpInitializeResult struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    mcpServerCapabilities `json:"capabilities"`
	ServerInfo      mcpServerInfo       `json:"serverInfo"`
}

type mcpServerCapabilities struct {
	Tools     *mcpToolsCapability `json:"tools,omitempty"`
	Resources *struct{}           `json:"resources,omitempty"`
	Prompts   *struct{}           `json:"prompts,omitempty"`
}

type mcpToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type mcpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type mcpToolsListResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type mcpToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type mcpToolCallResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// StdioClient manages multiple stdio-based MCP servers
type StdioClient struct {
	servers    map[string]*StdioServer
	mu         sync.RWMutex
}

// NewStdioClient creates a new stdio MCP client
func NewStdioClient() *StdioClient {
	return &StdioClient{
		servers: make(map[string]*StdioServer),
	}
}

// AddServer adds and starts a new stdio MCP server
func (c *StdioClient) AddServer(server *StdioServer) error {
	c.mu.Lock()
	c.servers[server.ID] = server
	c.mu.Unlock()

	// Start the server process
	if err := c.StartServer(server.ID); err != nil {
		server.LastError = err.Error()
		return err
	}

	return nil
}

// RemoveServer stops and removes an MCP server
func (c *StdioClient) RemoveServer(serverID string) error {
	c.mu.Lock()
	server, exists := c.servers[serverID]
	if exists {
		delete(c.servers, serverID)
	}
	c.mu.Unlock()

	if exists && server.running {
		return server.Stop()
	}
	return nil
}

// GetServer returns a server by ID
func (c *StdioClient) GetServer(serverID string) *StdioServer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.servers[serverID]
}

// GetUserServers returns all servers for a user
func (c *StdioClient) GetUserServers(userID string) []*StdioServer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	servers := make([]*StdioServer, 0)
	for _, server := range c.servers {
		if server.UserID == userID {
			servers = append(servers, server)
		}
	}
	return servers
}

// StartServer starts the MCP server process
func (c *StdioClient) StartServer(serverID string) error {
	c.mu.RLock()
	server, exists := c.servers[serverID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	return server.Start()
}

// StopServer stops the MCP server process
func (c *StdioClient) StopServer(serverID string) error {
	c.mu.RLock()
	server, exists := c.servers[serverID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}

	return server.Stop()
}

// GetAllTools returns all tools from all running servers for a user
func (c *StdioClient) GetAllTools(userID string) []*RemoteTool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	allTools := make([]*RemoteTool, 0)
	for _, server := range c.servers {
		if server.UserID == userID && server.Enabled && server.running {
			allTools = append(allTools, server.Tools...)
		}
	}
	return allTools
}

// ExecuteTool executes a tool on a specific server
func (c *StdioClient) ExecuteTool(ctx context.Context, serverID, toolName string, params map[string]interface{}) (interface{}, error) {
	c.mu.RLock()
	server, exists := c.servers[serverID]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	if !server.running {
		return nil, fmt.Errorf("server not running: %s", server.Name)
	}

	return server.CallTool(ctx, toolName, params)
}

// StopAll stops all running servers
func (c *StdioClient) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, server := range c.servers {
		if server.running {
			server.Stop()
		}
	}
}

// StdioServer methods

// Start starts the MCP server process
func (s *StdioServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // Already running
	}

	// Create the command
	s.cmd = exec.Command(s.Command, s.Args...)

	// Set up environment
	s.cmd.Env = append(os.Environ(), s.Env...)

	// Set up pipes
	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Initialize pending requests map
	s.pending = make(map[int64]chan *jsonRPCResponse)

	// Start the process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	s.running = true

	// Start reading responses
	go s.readResponses()
	go s.readStderr()

	// Initialize the MCP connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.initialize(ctx); err != nil {
		s.Stop()
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	// Fetch available tools
	if err := s.refreshTools(ctx); err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch tools: %w", err)
	}

	return nil
}

// Stop stops the MCP server process
func (s *StdioServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	// Close stdin to signal the process to exit
	if s.stdin != nil {
		s.stdin.Close()
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited normally
	case <-time.After(5 * time.Second):
		// Force kill
		s.cmd.Process.Kill()
	}

	// Cancel all pending requests
	s.pendingMu.Lock()
	for _, ch := range s.pending {
		close(ch)
	}
	s.pending = nil
	s.pendingMu.Unlock()

	return nil
}

// readResponses reads JSON-RPC responses from stdout
func (s *StdioServer) readResponses() {
	scanner := bufio.NewScanner(s.stdout)
	// Increase buffer size for large responses
	buf := make([]byte, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		// Dispatch to waiting request
		s.pendingMu.Lock()
		if ch, exists := s.pending[resp.ID]; exists {
			ch <- &resp
			delete(s.pending, resp.ID)
		}
		s.pendingMu.Unlock()
	}
}

// readStderr reads and logs stderr
func (s *StdioServer) readStderr() {
	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		// Log stderr output (could be useful for debugging)
		_ = scanner.Text()
	}
}

// sendRequest sends a JSON-RPC request and waits for response
func (s *StdioServer) sendRequest(ctx context.Context, method string, params interface{}) (*jsonRPCResponse, error) {
	if !s.running {
		return nil, fmt.Errorf("server not running")
	}

	id := atomic.AddInt64(&s.requestID, 1)

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respCh := make(chan *jsonRPCResponse, 1)
	s.pendingMu.Lock()
	s.pending[id] = respCh
	s.pendingMu.Unlock()

	// Serialize and send request
	data, err := json.Marshal(req)
	if err != nil {
		s.pendingMu.Lock()
		delete(s.pending, id)
		s.pendingMu.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	s.mu.Lock()
	_, err = s.stdin.Write(append(data, '\n'))
	s.mu.Unlock()

	if err != nil {
		s.pendingMu.Lock()
		delete(s.pending, id)
		s.pendingMu.Unlock()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, fmt.Errorf("connection closed")
		}
		return resp, nil
	case <-ctx.Done():
		s.pendingMu.Lock()
		delete(s.pending, id)
		s.pendingMu.Unlock()
		return nil, ctx.Err()
	}
}

// initialize performs MCP initialization handshake
func (s *StdioServer) initialize(ctx context.Context) error {
	params := mcpInitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: mcpCapabilities{
			Roots: &mcpRootsCapability{
				ListChanged: true,
			},
		},
		ClientInfo: mcpClientInfo{
			Name:    "Prism",
			Version: "1.0.0",
		},
	}

	resp, err := s.sendRequest(ctx, "initialize", params)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	// Send initialized notification (no response expected, but we'll use a request ID anyway)
	_, err = s.sendRequest(ctx, "notifications/initialized", nil)
	// Ignore error for notification - some servers don't respond to notifications
	_ = err

	return nil
}

// refreshTools fetches available tools from the server
func (s *StdioServer) refreshTools(ctx context.Context) error {
	resp, err := s.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	var result mcpToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to parse tools: %w", err)
	}

	// Convert to RemoteTool format
	tools := make([]*RemoteTool, len(result.Tools))
	for i, t := range result.Tools {
		tool := &RemoteTool{
			ServerID:    s.ID,
			ServerName:  s.Name,
			Name:        t.Name,
			Description: t.Description,
		}

		// Parse input schema
		if t.InputSchema != nil {
			tool.Parameters.Type = "object"
			tool.Parameters.Properties = make(map[string]llm.JSONProperty)

			if props, ok := t.InputSchema["properties"].(map[string]interface{}); ok {
				for name, propRaw := range props {
					if prop, ok := propRaw.(map[string]interface{}); ok {
						tool.Parameters.Properties[name] = llm.JSONProperty{
							Type:        getStringField(prop, "type"),
							Description: getStringField(prop, "description"),
						}
					}
				}
			}
			if required, ok := t.InputSchema["required"].([]interface{}); ok {
				for _, r := range required {
					if s, ok := r.(string); ok {
						tool.Parameters.Required = append(tool.Parameters.Required, s)
					}
				}
			}
		}

		tools[i] = tool
	}

	s.Tools = tools
	return nil
}

// CallTool executes a tool on this server
func (s *StdioServer) CallTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	callParams := mcpToolCallParams{
		Name:      toolName,
		Arguments: params,
	}

	resp, err := s.sendRequest(ctx, "tools/call", callParams)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s", resp.Error.Message)
	}

	var result mcpToolCallResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w", err)
	}

	// Extract text content
	var output string
	for _, content := range result.Content {
		if content.Type == "text" {
			output += content.Text
		}
	}

	if result.IsError {
		return map[string]interface{}{
			"error":  true,
			"output": output,
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"output":  output,
	}, nil
}

// IsRunning returns whether the server process is running
func (s *StdioServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
