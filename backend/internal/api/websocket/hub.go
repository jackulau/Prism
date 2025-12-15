package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("Client registered: user=%s", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients[client.UserID], client)
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: user=%s", client.UserID)
		}
	}
}

// SendToUser sends a message to all clients of a user
func (h *Hub) SendToUser(userID string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				// Client buffer is full, skip
				log.Printf("Client buffer full, skipping message for user=%s", userID)
			}
		}
	}
}

// Register registers a client with the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Message types
const (
	TypeChatMessage   = "chat.message"
	TypeChatChunk     = "chat.chunk"
	TypeChatComplete  = "chat.complete"
	TypeToolStarted   = "tool.started"
	TypeToolCompleted = "tool.completed"
	TypeToolConfirm   = "tool.confirm"
	TypeError         = "error"
	TypeChatStop      = "chat.stop"

	// Agent message types
	TypeAgentRun             = "agent.run"
	TypeAgentRunParallel     = "agent.run_parallel"
	TypeAgentStarted         = "agent.started"
	TypeAgentThinking        = "agent.thinking"
	TypeAgentStreamChunk     = "agent.stream_chunk"
	TypeAgentToolCall        = "agent.tool_call"
	TypeAgentToolResult      = "agent.tool_result"
	TypeAgentCompleted       = "agent.completed"
	TypeAgentFailed          = "agent.failed"
	TypeAgentCancelled       = "agent.cancelled"
	TypeAgentStop            = "agent.stop"
	TypeAgentStatus          = "agent.status"
	TypeAgentList            = "agent.list"
	TypeAgentBatchProgress   = "agent.batch_progress"
	TypeAgentBatchCompleted  = "agent.batch_completed"
	TypeAgentCheckIn         = "agent.check_in"      // Pause point for user confirmation after max iterations
	TypeAgentContinue        = "agent.continue"      // User confirms to continue the agentic loop

	// Preview/Sandbox message types
	TypePreviewReady    = "preview.ready"
	TypePreviewContent  = "preview.content"
	TypePreviewError    = "preview.error"
	TypeBuildStart      = "build.start"
	TypeBuildStarted    = "build.started"
	TypeBuildOutput     = "build.output"
	TypeBuildCompleted  = "build.completed"
	TypeBuildStop       = "build.stop"

	// Shell execution message types
	TypeShellStart     = "shell.start"
	TypeShellOutput    = "shell.output"
	TypeShellCompleted = "shell.completed"
	TypeShellFailed    = "shell.failed"
	TypeShellStop      = "shell.stop"
	TypeFilesUpdated       = "files.updated"
	TypeFileContent        = "file.content"
	TypeFileRequest        = "file.request"
	TypeFileHistoryRequest = "file.history_request"
	TypeFileHistoryList    = "file.history_list"
	TypeFileHistoryContent = "file.history_content"

	// Swarm/Multi-agent message types
	TypeSwarmCreate          = "swarm.create"
	TypeSwarmRun             = "swarm.run"
	TypeSwarmStarted         = "swarm.started"
	TypeSwarmAgentStarted    = "swarm.agent_started"
	TypeSwarmAgentOutput     = "swarm.agent_output"
	TypeSwarmAgentCompleted  = "swarm.agent_completed"
	TypeSwarmAgentFailed     = "swarm.agent_failed"
	TypeSwarmMessage         = "swarm.message"
	TypeSwarmSynthesizing    = "swarm.synthesizing"
	TypeSwarmProgress        = "swarm.progress"
	TypeSwarmCompleted       = "swarm.completed"
	TypeSwarmFailed          = "swarm.failed"
	TypeSwarmCancelled       = "swarm.cancelled"
	TypeSwarmStop            = "swarm.stop"
	TypeSwarmStatus          = "swarm.status"
	TypeSwarmList            = "swarm.list"
)

// IncomingMessage represents a message from the client
type IncomingMessage struct {
	Type           string                 `json:"type"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Attachments    []string               `json:"attachments,omitempty"`
	ExecutionID    string                 `json:"execution_id,omitempty"`
	Approved       bool                   `json:"approved,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`

	// Agent-related fields
	AgentID     string        `json:"agent_id,omitempty"`
	Tasks       []AgentTask   `json:"tasks,omitempty"`       // For parallel execution
	AgentConfig *AgentConfig  `json:"agent_config,omitempty"`
	Context     string        `json:"context,omitempty"`
	Priority    int           `json:"priority,omitempty"`

	// Swarm/Multi-agent fields
	SwarmID      string            `json:"swarm_id,omitempty"`
	SwarmConfig  *SwarmConfig      `json:"swarm_config,omitempty"`
	Strategy     string            `json:"strategy,omitempty"`     // parallel, pipeline, debate, consensus, map_reduce, specialist
	AgentRoles   []AgentRoleConfig `json:"agent_roles,omitempty"`  // Roles for multi-agent swarm
}

// SwarmConfig represents configuration for a multi-agent swarm
type SwarmConfig struct {
	Name         string            `json:"name,omitempty"`
	Strategy     string            `json:"strategy"`       // parallel, pipeline, debate, consensus, map_reduce, specialist
	MaxAgents    int               `json:"max_agents,omitempty"`
	TimeoutSecs  int               `json:"timeout_secs,omitempty"`
	AgentRoles   []AgentRoleConfig `json:"agent_roles"`
	Synthesizer  *AgentConfig      `json:"synthesizer,omitempty"` // Config for result synthesizer
}

// AgentRoleConfig represents an agent with a specific role
type AgentRoleConfig struct {
	Role         string       `json:"role"`         // general, planner, coder, reviewer, researcher, writer, analyst, debugger, tester, synthesizer
	Config       *AgentConfig `json:"config,omitempty"`
	Count        int          `json:"count,omitempty"`        // Number of agents with this role
	SystemPrompt string       `json:"system_prompt,omitempty"` // Override default role prompt
}

// AgentTask represents a task in a batch request
type AgentTask struct {
	ID       string                 `json:"id,omitempty"`
	Prompt   string                 `json:"prompt"`
	Context  string                 `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentConfig represents configuration for an agent
type AgentConfig struct {
	Name         string   `json:"name,omitempty"`
	Provider     string   `json:"provider"`
	Model        string   `json:"model"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	Tools        []string `json:"tools,omitempty"` // Tool names to enable
}

// OutgoingMessage represents a message to the client
type OutgoingMessage struct {
	Type           string      `json:"type"`
	ConversationID string      `json:"conversation_id,omitempty"`
	MessageID      string      `json:"message_id,omitempty"`
	Delta          string      `json:"delta,omitempty"`
	FinishReason   string      `json:"finish_reason,omitempty"`
	ExecutionID    string      `json:"execution_id,omitempty"`
	ToolName       string      `json:"tool_name,omitempty"`
	Parameters     interface{} `json:"parameters,omitempty"`
	Result         interface{} `json:"result,omitempty"`
	Status         string      `json:"status,omitempty"`
	Code           string      `json:"code,omitempty"`
	Message        string      `json:"message,omitempty"`

	// Agent-related fields
	AgentID       string                 `json:"agent_id,omitempty"`
	TaskID        string                 `json:"task_id,omitempty"`
	Output        string                 `json:"output,omitempty"`
	Agents        []AgentInfo            `json:"agents,omitempty"`       // For agent list
	Results       []AgentResultInfo      `json:"results,omitempty"`      // For batch results
	Progress      *BatchProgressInfo     `json:"progress,omitempty"`     // For batch progress
	Error         string                 `json:"error,omitempty"`
	Duration      int64                  `json:"duration,omitempty"`     // Duration in milliseconds
	Metadata      map[string]interface{} `json:"metadata,omitempty"`

	// Preview/Sandbox-related fields
	URL         string     `json:"url,omitempty"`
	Content     string     `json:"content,omitempty"`
	FilePath    string     `json:"file_path,omitempty"`
	Files       []FileInfo `json:"files,omitempty"`
	Stream      string     `json:"stream,omitempty"`       // "stdout" or "stderr"
	Success     bool       `json:"success,omitempty"`
	PreviewURL  string     `json:"preview_url,omitempty"`
	BuildID     string     `json:"build_id,omitempty"`

	// Swarm/Multi-agent fields
	SwarmID       string           `json:"swarm_id,omitempty"`
	SwarmStatus   string           `json:"swarm_status,omitempty"`
	FinalOutput   string           `json:"final_output,omitempty"`
	SwarmAgents   []SwarmAgentInfo `json:"swarm_agents,omitempty"`
	SwarmProgress *SwarmProgressInfo `json:"swarm_progress,omitempty"`
	AgentRole     string           `json:"agent_role,omitempty"`
	Input         string           `json:"input,omitempty"`

	// MCP-related fields
	IsMCPTool     bool   `json:"is_mcp_tool,omitempty"`
	MCPServerName string `json:"mcp_server_name,omitempty"`
	MCPServerID   string `json:"mcp_server_id,omitempty"`

	// Iteration tracking for agentic loops
	IterationCount int `json:"iteration_count,omitempty"`
}

// SwarmAgentInfo represents information about an agent in a swarm
type SwarmAgentInfo struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Input       string `json:"input,omitempty"`
	Output      string `json:"output,omitempty"`
	StartedAt   *int64 `json:"started_at,omitempty"`
	CompletedAt *int64 `json:"completed_at,omitempty"`
}

// SwarmProgressInfo represents the progress of a swarm
type SwarmProgressInfo struct {
	TotalAgents     int    `json:"total_agents"`
	RunningAgents   int    `json:"running_agents"`
	CompletedAgents int    `json:"completed_agents"`
	FailedAgents    int    `json:"failed_agents"`
	Phase           string `json:"phase"` // "initializing", "running", "synthesizing", "completed"
}

// FileInfo represents a file in the sandbox
type FileInfo struct {
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	IsDirectory bool       `json:"is_directory"`
	Children    []FileInfo `json:"children,omitempty"`
	Content     string     `json:"content,omitempty"`
	Size        int64      `json:"size,omitempty"`
	Modified    int64      `json:"modified,omitempty"`
}

// AgentInfo represents information about an agent
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	CreatedAt   int64  `json:"created_at"`
	StartedAt   *int64 `json:"started_at,omitempty"`
	CompletedAt *int64 `json:"completed_at,omitempty"`
}

// AgentResultInfo represents the result of an agent execution
type AgentResultInfo struct {
	AgentID     string                 `json:"agent_id"`
	TaskID      string                 `json:"task_id"`
	Success     bool                   `json:"success"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration"` // Duration in milliseconds
	CompletedAt int64                  `json:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchProgressInfo represents the progress of a batch execution
type BatchProgressInfo struct {
	TotalTasks     int `json:"total_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
	RunningTasks   int `json:"running_tasks"`
	PendingTasks   int `json:"pending_tasks"`
}

// NewChatChunk creates a new chat chunk message
func NewChatChunk(conversationID, messageID, delta string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:           TypeChatChunk,
		ConversationID: conversationID,
		MessageID:      messageID,
		Delta:          delta,
	}
}

// NewChatComplete creates a new chat complete message
func NewChatComplete(conversationID, messageID, finishReason string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:           TypeChatComplete,
		ConversationID: conversationID,
		MessageID:      messageID,
		FinishReason:   finishReason,
	}
}

// NewToolStarted creates a new tool started message
func NewToolStarted(conversationID, executionID, toolName string, parameters interface{}) *OutgoingMessage {
	return &OutgoingMessage{
		Type:           TypeToolStarted,
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       toolName,
		Parameters:     parameters,
	}
}

// NewToolCompleted creates a new tool completed message
func NewToolCompleted(conversationID, executionID string, result interface{}, status string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:           TypeToolCompleted,
		ConversationID: conversationID,
		ExecutionID:    executionID,
		Result:         result,
		Status:         status,
	}
}

// NewError creates a new error message
func NewError(code, message string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeError,
		Code:    code,
		Message: message,
		Error:   message, // Include Error field for consistency with other error constructors
	}
}

// NewAgentStarted creates a new agent started message
func NewAgentStarted(agentID, taskID string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeAgentStarted,
		AgentID: agentID,
		TaskID:  taskID,
		Status:  "running",
	}
}

// NewAgentStreamChunk creates a new agent stream chunk message
func NewAgentStreamChunk(agentID, taskID, delta string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeAgentStreamChunk,
		AgentID: agentID,
		TaskID:  taskID,
		Delta:   delta,
	}
}

// NewAgentToolCall creates a new agent tool call message
func NewAgentToolCall(agentID, taskID, toolName string, parameters interface{}) *OutgoingMessage {
	return &OutgoingMessage{
		Type:       TypeAgentToolCall,
		AgentID:    agentID,
		TaskID:     taskID,
		ToolName:   toolName,
		Parameters: parameters,
	}
}

// NewAgentCompleted creates a new agent completed message
func NewAgentCompleted(agentID, taskID, output string, durationMs int64) *OutgoingMessage {
	return &OutgoingMessage{
		Type:     TypeAgentCompleted,
		AgentID:  agentID,
		TaskID:   taskID,
		Output:   output,
		Status:   "completed",
		Duration: durationMs,
	}
}

// NewAgentFailed creates a new agent failed message
func NewAgentFailed(agentID, taskID, errorMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeAgentFailed,
		AgentID: agentID,
		TaskID:  taskID,
		Error:   errorMsg,
		Status:  "failed",
	}
}

// NewAgentCancelled creates a new agent cancelled message
func NewAgentCancelled(agentID, taskID string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeAgentCancelled,
		AgentID: agentID,
		TaskID:  taskID,
		Status:  "cancelled",
	}
}

// NewAgentCheckIn creates a new agent check-in message
// This is sent when the agent has reached the max iterations and needs user confirmation to continue
func NewAgentCheckIn(conversationID string, iterationCount int, message string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:           TypeAgentCheckIn,
		ConversationID: conversationID,
		IterationCount: iterationCount,
		Message:        message,
	}
}

// NewAgentList creates a new agent list message
func NewAgentList(agents []AgentInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:   TypeAgentList,
		Agents: agents,
	}
}

// NewAgentStatus creates a new agent status message
func NewAgentStatus(agent AgentInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:   TypeAgentStatus,
		Agents: []AgentInfo{agent},
	}
}

// NewAgentBatchProgress creates a new batch progress message
func NewAgentBatchProgress(executionID string, progress *BatchProgressInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeAgentBatchProgress,
		ExecutionID: executionID,
		Progress:    progress,
	}
}

// NewAgentBatchCompleted creates a new batch completed message
func NewAgentBatchCompleted(executionID string, results []AgentResultInfo, durationMs int64) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeAgentBatchCompleted,
		ExecutionID: executionID,
		Results:     results,
		Status:      "completed",
		Duration:    durationMs,
	}
}

// Preview/Sandbox message constructors

// NewPreviewReady creates a new preview ready message
func NewPreviewReady(url string) *OutgoingMessage {
	return &OutgoingMessage{
		Type: TypePreviewReady,
		URL:  url,
	}
}

// NewPreviewContent creates a new preview content message (for inline HTML)
func NewPreviewContent(content string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypePreviewContent,
		Content: content,
	}
}

// NewPreviewError creates a new preview error message
func NewPreviewError(errorMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:  TypePreviewError,
		Error: errorMsg,
	}
}

// NewBuildStarted creates a new build started message
func NewBuildStarted(buildID string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeBuildStarted,
		BuildID: buildID,
		Status:  "building",
	}
}

// NewBuildOutput creates a new build output message
func NewBuildOutput(buildID, content, stream string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeBuildOutput,
		BuildID: buildID,
		Content: content,
		Stream:  stream,
	}
}

// NewBuildCompleted creates a new build completed message
func NewBuildCompleted(buildID string, success bool, previewURL string, durationMs int64) *OutgoingMessage {
	status := "success"
	if !success {
		status = "error"
	}
	return &OutgoingMessage{
		Type:       TypeBuildCompleted,
		BuildID:    buildID,
		Success:    success,
		PreviewURL: previewURL,
		Status:     status,
		Duration:   durationMs,
	}
}

// Shell command message constructors

// NewShellStarted creates a new shell command started message
func NewShellStarted(commandID, command string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeShellStart,
		BuildID: commandID, // Reusing BuildID for command ID
		Content: command,
		Status:  "running",
	}
}

// NewShellOutput creates a new shell output message
func NewShellOutput(commandID, content, stream string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeShellOutput,
		BuildID: commandID,
		Content: content,
		Stream:  stream, // "stdout" or "stderr"
	}
}

// NewShellCompleted creates a new shell command completed message
func NewShellCompleted(commandID string, exitCode int, success bool, durationMs int64) *OutgoingMessage {
	status := "success"
	if !success {
		status = "error"
	}
	return &OutgoingMessage{
		Type:     TypeShellCompleted,
		BuildID:  commandID,
		Success:  success,
		Status:   status,
		Duration: durationMs,
		Metadata: map[string]interface{}{
			"exit_code": exitCode,
		},
	}
}

// NewShellFailed creates a new shell command failed message
func NewShellFailed(commandID, errorMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:    TypeShellFailed,
		BuildID: commandID,
		Error:   errorMsg,
		Status:  "failed",
	}
}

// NewFilesUpdated creates a new files updated message
func NewFilesUpdated(files []FileInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:  TypeFilesUpdated,
		Files: files,
	}
}

// NewFileContent creates a new file content message
func NewFileContent(filePath, content string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:     TypeFileContent,
		FilePath: filePath,
		Content:  content,
	}
}

// Swarm/Multi-agent message constructors

// NewSwarmStarted creates a new swarm started message
func NewSwarmStarted(swarmID string, agents []SwarmAgentInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmStarted,
		SwarmID:     swarmID,
		SwarmStatus: "running",
		SwarmAgents: agents,
	}
}

// NewSwarmAgentStarted creates a message when an agent in a swarm starts
func NewSwarmAgentStarted(swarmID, agentID, role, input string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      TypeSwarmAgentStarted,
		SwarmID:   swarmID,
		AgentID:   agentID,
		AgentRole: role,
		Input:     input,
		Status:    "running",
	}
}

// NewSwarmAgentOutput creates a message for agent output in a swarm
func NewSwarmAgentOutput(swarmID, agentID, role, delta string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      TypeSwarmAgentOutput,
		SwarmID:   swarmID,
		AgentID:   agentID,
		AgentRole: role,
		Delta:     delta,
	}
}

// NewSwarmAgentCompleted creates a message when an agent in a swarm completes
func NewSwarmAgentCompleted(swarmID, agentID, role, output string, durationMs int64) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      TypeSwarmAgentCompleted,
		SwarmID:   swarmID,
		AgentID:   agentID,
		AgentRole: role,
		Output:    output,
		Status:    "completed",
		Duration:  durationMs,
	}
}

// NewSwarmAgentFailed creates a message when an agent in a swarm fails
func NewSwarmAgentFailed(swarmID, agentID, role, errorMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      TypeSwarmAgentFailed,
		SwarmID:   swarmID,
		AgentID:   agentID,
		AgentRole: role,
		Error:     errorMsg,
		Status:    "failed",
	}
}

// NewSwarmProgress creates a swarm progress message
func NewSwarmProgress(swarmID string, progress *SwarmProgressInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:          TypeSwarmProgress,
		SwarmID:       swarmID,
		SwarmProgress: progress,
	}
}

// NewSwarmSynthesizing creates a message when swarm is synthesizing results
func NewSwarmSynthesizing(swarmID string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmSynthesizing,
		SwarmID:     swarmID,
		SwarmStatus: "synthesizing",
	}
}

// NewSwarmCompleted creates a swarm completed message
func NewSwarmCompleted(swarmID, finalOutput string, agents []SwarmAgentInfo, durationMs int64) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmCompleted,
		SwarmID:     swarmID,
		SwarmStatus: "completed",
		FinalOutput: finalOutput,
		SwarmAgents: agents,
		Duration:    durationMs,
	}
}

// NewSwarmFailed creates a swarm failed message
func NewSwarmFailed(swarmID, errorMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmFailed,
		SwarmID:     swarmID,
		SwarmStatus: "failed",
		Error:       errorMsg,
	}
}

// NewSwarmCancelled creates a swarm cancelled message
func NewSwarmCancelled(swarmID string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmCancelled,
		SwarmID:     swarmID,
		SwarmStatus: "cancelled",
	}
}

// NewSwarmStatus creates a swarm status message
func NewSwarmStatus(swarmID, status string, agents []SwarmAgentInfo, progress *SwarmProgressInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:          TypeSwarmStatus,
		SwarmID:       swarmID,
		SwarmStatus:   status,
		SwarmAgents:   agents,
		SwarmProgress: progress,
	}
}

// NewSwarmList creates a swarm list message
func NewSwarmList(swarms []SwarmAgentInfo) *OutgoingMessage {
	return &OutgoingMessage{
		Type:        TypeSwarmList,
		SwarmAgents: swarms,
	}
}

// FileHistoryEntry represents a file history entry for WebSocket messages
type FileHistoryEntry struct {
	ID        string `json:"id"`
	FilePath  string `json:"file_path"`
	Operation string `json:"operation"`
	Size      int    `json:"size"`
	CreatedAt string `json:"created_at"`
}

// NewFileHistoryList creates a new file history list message
func NewFileHistoryList(filePath string, entries []FileHistoryEntry) *OutgoingMessage {
	return &OutgoingMessage{
		Type:     TypeFileHistoryList,
		FilePath: filePath,
		Metadata: map[string]interface{}{
			"entries": entries,
			"count":   len(entries),
		},
	}
}

// NewFileHistoryContent creates a new file history content message
func NewFileHistoryContent(historyID, filePath, content, operation, createdAt string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:     TypeFileHistoryContent,
		FilePath: filePath,
		Content:  content,
		Metadata: map[string]interface{}{
			"history_id": historyID,
			"operation":  operation,
			"created_at": createdAt,
		},
	}
}
