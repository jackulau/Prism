package agent

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/llm"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "idle"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusFailed    AgentStatus = "failed"
	AgentStatusCancelled AgentStatus = "cancelled"
)

// AgentConfig holds configuration for an agent
type AgentConfig struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
	SystemPrompt  string            `json:"system_prompt,omitempty"`
	Temperature   float64           `json:"temperature,omitempty"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	Tools         []llm.ToolDefinition `json:"tools,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Agent represents an autonomous agent that can execute tasks
type Agent struct {
	ID          string            `json:"id"`
	Config      AgentConfig       `json:"config"`
	Status      AgentStatus       `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`

	// Internal state
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	llmManager  *llm.Manager
	messages    []llm.Message
	results     chan *AgentResult
	events      chan *AgentEvent
}

// AgentResult represents the result of an agent's execution
type AgentResult struct {
	AgentID     string                 `json:"agent_id"`
	TaskID      string                 `json:"task_id"`
	Success     bool                   `json:"success"`
	Output      string                 `json:"output,omitempty"`
	ToolResults []ToolResult           `json:"tool_results,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Usage       *llm.Usage             `json:"usage,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Duration    time.Duration          `json:"duration"`
	CompletedAt time.Time              `json:"completed_at"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Name       string                 `json:"name"`
	Output     string                 `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AgentEvent represents an event from an agent during execution
type AgentEvent struct {
	AgentID   string                 `json:"agent_id"`
	Type      AgentEventType         `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AgentEventType represents the type of agent event
type AgentEventType string

const (
	AgentEventStarted       AgentEventType = "started"
	AgentEventThinking      AgentEventType = "thinking"
	AgentEventStreamChunk   AgentEventType = "stream_chunk"
	AgentEventToolCall      AgentEventType = "tool_call"
	AgentEventToolResult    AgentEventType = "tool_result"
	AgentEventCompleted     AgentEventType = "completed"
	AgentEventFailed        AgentEventType = "failed"
	AgentEventCancelled     AgentEventType = "cancelled"
)

// NewAgent creates a new agent instance
func NewAgent(config AgentConfig, llmManager *llm.Manager) *Agent {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	return &Agent{
		ID:         config.ID,
		Config:     config,
		Status:     AgentStatusIdle,
		CreatedAt:  time.Now(),
		llmManager: llmManager,
		messages:   make([]llm.Message, 0),
		results:    make(chan *AgentResult, 1),
		events:     make(chan *AgentEvent, 100),
	}
}

// Start begins the agent's execution with the given task
func (a *Agent) Start(ctx context.Context, task *Task) error {
	a.mu.Lock()
	if a.Status == AgentStatusRunning {
		a.mu.Unlock()
		return ErrAgentAlreadyRunning
	}

	a.ctx, a.cancel = context.WithCancel(ctx)
	a.Status = AgentStatusRunning
	now := time.Now()
	a.StartedAt = &now
	a.mu.Unlock()

	// Emit started event
	a.emitEvent(AgentEventStarted, map[string]interface{}{
		"task_id": task.ID,
		"prompt":  task.Prompt,
	})

	// Run the agent asynchronously
	go a.run(task)

	return nil
}

// run executes the agent's main loop
func (a *Agent) run(task *Task) {
	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			a.fail(ErrAgentPanicked.Error())
		}
		close(a.events)
	}()

	// Build initial messages
	messages := a.buildMessages(task)

	// Create chat request
	req := &llm.ChatRequest{
		Model:       a.Config.Model,
		Messages:    messages,
		Tools:       a.Config.Tools,
		Temperature: a.Config.Temperature,
		MaxTokens:   a.Config.MaxTokens,
		Stream:      true,
	}

	// Execute chat
	stream, err := a.llmManager.Chat(a.ctx, a.Config.Provider, req)
	if err != nil {
		a.fail(err.Error())
		a.results <- &AgentResult{
			AgentID:     a.ID,
			TaskID:      task.ID,
			Success:     false,
			Error:       err.Error(),
			Duration:    time.Since(startTime),
			CompletedAt: time.Now(),
		}
		return
	}

	// Process stream
	var fullResponse string
	var toolCalls []llm.ToolCall
	var usage *llm.Usage

	for chunk := range stream {
		select {
		case <-a.ctx.Done():
			a.cancel()
			a.mu.Lock()
			a.Status = AgentStatusCancelled
			now := time.Now()
			a.CompletedAt = &now
			a.mu.Unlock()

			a.emitEvent(AgentEventCancelled, nil)
			a.results <- &AgentResult{
				AgentID:     a.ID,
				TaskID:      task.ID,
				Success:     false,
				Error:       "cancelled",
				Output:      fullResponse,
				Duration:    time.Since(startTime),
				CompletedAt: time.Now(),
			}
			return
		default:
		}

		if chunk.Error != nil {
			a.fail(chunk.Error.Error())
			a.results <- &AgentResult{
				AgentID:     a.ID,
				TaskID:      task.ID,
				Success:     false,
				Error:       chunk.Error.Error(),
				Output:      fullResponse,
				Duration:    time.Since(startTime),
				CompletedAt: time.Now(),
			}
			return
		}

		if chunk.Delta != "" {
			fullResponse += chunk.Delta
			a.emitEvent(AgentEventStreamChunk, map[string]interface{}{
				"delta": chunk.Delta,
			})
		}

		if len(chunk.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.ToolCalls...)
			for _, tc := range chunk.ToolCalls {
				a.emitEvent(AgentEventToolCall, map[string]interface{}{
					"tool_call_id": tc.ID,
					"name":         tc.Name,
					"parameters":   tc.Parameters,
				})
			}
		}

		if chunk.Usage != nil {
			usage = chunk.Usage
		}
	}

	// Mark as completed
	a.mu.Lock()
	a.Status = AgentStatusCompleted
	now := time.Now()
	a.CompletedAt = &now
	a.mu.Unlock()

	a.emitEvent(AgentEventCompleted, map[string]interface{}{
		"output": fullResponse,
	})

	// Send result
	result := &AgentResult{
		AgentID:     a.ID,
		TaskID:      task.ID,
		Success:     true,
		Output:      fullResponse,
		Usage:       usage,
		Duration:    time.Since(startTime),
		CompletedAt: time.Now(),
	}

	// Convert tool calls to tool results placeholder
	if len(toolCalls) > 0 {
		result.ToolResults = make([]ToolResult, len(toolCalls))
		for i, tc := range toolCalls {
			result.ToolResults[i] = ToolResult{
				ToolCallID: tc.ID,
				Name:       tc.Name,
			}
		}
	}

	a.results <- result
}

// buildMessages constructs the message history for the LLM
func (a *Agent) buildMessages(task *Task) []llm.Message {
	messages := make([]llm.Message, 0)

	// Add system prompt if configured
	if a.Config.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: a.Config.SystemPrompt,
		})
	}

	// Add task context if provided
	if task.Context != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "Context: " + task.Context,
		})
	}

	// Add existing conversation history
	messages = append(messages, a.messages...)

	// Add the task prompt
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: task.Prompt,
	})

	return messages
}

// Stop cancels the agent's execution
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
}

// fail marks the agent as failed
func (a *Agent) fail(errMsg string) {
	a.mu.Lock()
	a.Status = AgentStatusFailed
	a.Error = errMsg
	now := time.Now()
	a.CompletedAt = &now
	a.mu.Unlock()

	a.emitEvent(AgentEventFailed, map[string]interface{}{
		"error": errMsg,
	})
}

// emitEvent sends an event to the events channel
func (a *Agent) emitEvent(eventType AgentEventType, data map[string]interface{}) {
	select {
	case a.events <- &AgentEvent{
		AgentID:   a.ID,
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		// Event channel full, skip
	}
}

// GetStatus returns the current status of the agent
func (a *Agent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// Results returns the results channel
func (a *Agent) Results() <-chan *AgentResult {
	return a.results
}

// Events returns the events channel
func (a *Agent) Events() <-chan *AgentEvent {
	return a.events
}

// AddMessage adds a message to the agent's conversation history
func (a *Agent) AddMessage(msg llm.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = append(a.messages, msg)
}

// GetMessages returns the agent's conversation history
func (a *Agent) GetMessages() []llm.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]llm.Message{}, a.messages...)
}
