package prism

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/providers"
	"github.com/jacklau/prism/internal/sandbox"
)

// Provider errors
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrAgentAlreadyExists = errors.New("agent already exists")
	ErrProviderNotStarted = errors.New("provider not started")
)

// Config holds configuration for the Prism provider
type Config struct {
	// PersistExecutions controls whether to persist execution records to the database
	PersistExecutions bool

	// TrackCosts controls whether to track and calculate costs
	TrackCosts bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		PersistExecutions: true,
		TrackCosts:        true,
	}
}

// Provider implements the AgentProvider interface for Prism's native agent execution
type Provider struct {
	config Config

	// Dependencies
	agentManager *agent.Manager
	llmManager   *llm.Manager
	sandbox      *sandbox.Service

	// Repositories for persistence
	execRepo    *repository.AgentExecutionRepository
	messageRepo *repository.AgentMessageRepository
	toolRepo    *repository.AgentToolCallRepository

	// Active agents managed by this provider
	agents   map[string]*providerAgent
	agentsMu sync.RWMutex

	// State
	running bool
	mu      sync.RWMutex
}

// providerAgent wraps an agent execution with provider-specific state
type providerAgent struct {
	agent       *providers.Agent
	internalID  string // ID in agent.Manager
	execution   *agent.Execution
	messages    []providers.Message
	totalUsage  *providers.Usage
	totalCost   *providers.Cost
	streamChans []chan providers.StreamChunk
	mu          sync.RWMutex
}

// NewProvider creates a new Prism provider
func NewProvider(
	agentManager *agent.Manager,
	llmManager *llm.Manager,
	sandboxService *sandbox.Service,
	config Config,
) *Provider {
	return &Provider{
		config:       config,
		agentManager: agentManager,
		llmManager:   llmManager,
		sandbox:      sandboxService,
		agents:       make(map[string]*providerAgent),
	}
}

// SetRepositories sets the database repositories for persistence
func (p *Provider) SetRepositories(
	execRepo *repository.AgentExecutionRepository,
	messageRepo *repository.AgentMessageRepository,
	toolRepo *repository.AgentToolCallRepository,
) {
	p.execRepo = execRepo
	p.messageRepo = messageRepo
	p.toolRepo = toolRepo
}

// Start starts the provider
func (p *Provider) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return
	}

	p.running = true
	p.agentManager.Start()
}

// Stop stops the provider
func (p *Provider) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
	p.agentManager.Stop()
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "prism"
}

// CreateAgent creates a new agent
func (p *Provider) CreateAgent(ctx context.Context, req providers.CreateAgentRequest) (*providers.Agent, error) {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return nil, ErrProviderNotStarted
	}
	p.mu.RUnlock()

	// Generate agent ID
	agentID := uuid.New().String()
	now := time.Now()

	// Create the provider agent
	agentRecord := &providers.Agent{
		ID:           agentID,
		ProviderName: "prism",
		UserID:       req.UserID,
		Name:         req.Name,
		Description:  req.Description,
		Status:       providers.AgentStatusIdle,
		LLMProvider:  req.Provider,
		Model:        req.Model,
		CreatedAt:    now,
		Metadata:     req.Metadata,
	}

	// Create internal wrapper
	pa := &providerAgent{
		agent:      agentRecord,
		totalUsage: &providers.Usage{},
		totalCost:  &providers.Cost{Currency: "USD"},
		messages:   make([]providers.Message, 0),
	}

	// Store the agent
	p.agentsMu.Lock()
	p.agents[agentID] = pa
	p.agentsMu.Unlock()

	// Persist to database if enabled
	if p.config.PersistExecutions && p.execRepo != nil {
		_, err := p.execRepo.Create(
			req.UserID,
			"prism",
			req.Provider,
			req.Model,
			req.Name,
			req.Metadata,
		)
		if err != nil {
			// Log but don't fail - persistence is optional
			// TODO: Add proper logging
		}
	}

	return agentRecord, nil
}

// GetAgent retrieves an agent by ID
func (p *Provider) GetAgent(ctx context.Context, agentID string) (*providers.Agent, error) {
	p.agentsMu.RLock()
	pa, ok := p.agents[agentID]
	p.agentsMu.RUnlock()

	if !ok {
		return nil, ErrAgentNotFound
	}

	pa.mu.RLock()
	defer pa.mu.RUnlock()

	// Return a copy with current state
	agentCopy := *pa.agent
	agentCopy.Usage = pa.totalUsage
	agentCopy.Cost = pa.totalCost

	return &agentCopy, nil
}

// SendMessage sends a message to an agent and returns a streaming response
func (p *Provider) SendMessage(ctx context.Context, agentID string, message string) (<-chan providers.StreamChunk, error) {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return nil, ErrProviderNotStarted
	}
	p.mu.RUnlock()

	p.agentsMu.RLock()
	pa, ok := p.agents[agentID]
	p.agentsMu.RUnlock()

	if !ok {
		return nil, ErrAgentNotFound
	}

	// Create output channel
	outChan := make(chan providers.StreamChunk, 100)

	// Track the channel for cleanup
	pa.mu.Lock()
	pa.streamChans = append(pa.streamChans, outChan)
	pa.mu.Unlock()

	// Start async execution
	go p.executeMessage(ctx, pa, message, outChan)

	return outChan, nil
}

// executeMessage handles the actual message execution
func (p *Provider) executeMessage(ctx context.Context, pa *providerAgent, message string, outChan chan providers.StreamChunk) {
	defer close(outChan)

	// Update agent status
	pa.mu.Lock()
	pa.agent.Status = providers.AgentStatusRunning
	now := time.Now()
	pa.agent.StartedAt = &now
	pa.mu.Unlock()

	// Add user message to history
	userMsg := providers.Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	pa.mu.Lock()
	pa.messages = append(pa.messages, userMsg)
	pa.mu.Unlock()

	// Persist user message if enabled
	if p.config.PersistExecutions && p.messageRepo != nil {
		_, _ = p.messageRepo.Create(pa.agent.ID, "user", message, nil, "", 0, 0)
	}

	// Build the chat request
	llmMessages := p.buildLLMMessages(pa)

	req := &llm.ChatRequest{
		Model:       pa.agent.Model,
		Messages:    llmMessages,
		Stream:      true,
		Temperature: 0.7, // Default temperature
		MaxTokens:   4096,
	}

	// Execute the chat
	stream, err := p.llmManager.Chat(ctx, pa.agent.LLMProvider, req)
	if err != nil {
		p.handleError(pa, outChan, err)
		return
	}

	// Create stream adapter for converting chunks
	adapter := NewStreamAdapter(pa.agent.LLMProvider, pa.agent.Model)

	// Process the stream
	var fullResponse string
	var usage *llm.Usage

	for chunk := range stream {
		select {
		case <-ctx.Done():
			p.handleCancellation(pa, outChan)
			return
		default:
		}

		if chunk.Error != nil {
			p.handleError(pa, outChan, chunk.Error)
			return
		}

		// Convert and send the chunk
		providerChunk := adapter.ConvertLLMChunk(chunk)
		if providerChunk != nil {
			select {
			case outChan <- *providerChunk:
			default:
				// Channel full, skip
			}
		}

		// Accumulate response
		if chunk.Delta != "" {
			fullResponse += chunk.Delta
		}

		// Capture usage
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
	}

	// Add assistant message to history
	assistantMsg := providers.Message{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   fullResponse,
		Timestamp: time.Now(),
	}

	if usage != nil {
		assistantMsg.Usage = &providers.Usage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}

	pa.mu.Lock()
	pa.messages = append(pa.messages, assistantMsg)

	// Update totals
	if usage != nil {
		pa.totalUsage.PromptTokens += usage.PromptTokens
		pa.totalUsage.CompletionTokens += usage.CompletionTokens
		pa.totalUsage.TotalTokens += usage.TotalTokens

		if p.config.TrackCosts {
			cost := CalculateCost(pa.agent.LLMProvider, pa.agent.Model, usage)
			if cost != nil {
				pa.totalCost.InputCost += cost.InputCost
				pa.totalCost.OutputCost += cost.OutputCost
				pa.totalCost.TotalCost += cost.TotalCost
			}
		}
	}

	// Update agent status
	pa.agent.Status = providers.AgentStatusCompleted
	completedNow := time.Now()
	pa.agent.CompletedAt = &completedNow
	pa.agent.Usage = pa.totalUsage
	pa.agent.Cost = pa.totalCost
	pa.mu.Unlock()

	// Persist assistant message if enabled
	if p.config.PersistExecutions && p.messageRepo != nil {
		promptTokens := 0
		completionTokens := 0
		if usage != nil {
			promptTokens = usage.PromptTokens
			completionTokens = usage.CompletionTokens
		}
		_, _ = p.messageRepo.Create(pa.agent.ID, "assistant", fullResponse, nil, "", promptTokens, completionTokens)

		// Update execution record with usage
		if p.execRepo != nil && usage != nil {
			cost := CalculateCost(pa.agent.LLMProvider, pa.agent.Model, usage)
			inputCost, outputCost := 0.0, 0.0
			if cost != nil {
				inputCost = cost.InputCost
				outputCost = cost.OutputCost
			}
			_ = p.execRepo.AddUsage(pa.agent.ID, usage.PromptTokens, usage.CompletionTokens, inputCost, outputCost)
		}
	}

	// Send final done chunk
	finalChunk := providers.StreamChunk{
		Type:  providers.StreamChunkTypeDone,
		Done:  true,
		Usage: pa.totalUsage,
		Cost:  pa.totalCost,
	}

	select {
	case outChan <- finalChunk:
	default:
	}
}

// buildLLMMessages converts provider messages to LLM messages
func (p *Provider) buildLLMMessages(pa *providerAgent) []llm.Message {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	messages := make([]llm.Message, 0, len(pa.messages))
	for _, msg := range pa.messages {
		llmMsg := llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

		// Convert tool calls
		if len(msg.ToolCalls) > 0 {
			llmMsg.ToolCalls = make([]llm.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				llmMsg.ToolCalls[i] = llm.ToolCall{
					ID:         tc.ID,
					Name:       tc.Name,
					Parameters: tc.Parameters,
				}
			}
		}

		messages = append(messages, llmMsg)
	}

	return messages
}

// handleError handles errors during message execution
func (p *Provider) handleError(pa *providerAgent, outChan chan providers.StreamChunk, err error) {
	pa.mu.Lock()
	pa.agent.Status = providers.AgentStatusFailed
	pa.agent.Error = err.Error()
	now := time.Now()
	pa.agent.CompletedAt = &now
	pa.mu.Unlock()

	// Persist failure if enabled
	if p.config.PersistExecutions && p.execRepo != nil {
		_ = p.execRepo.Fail(pa.agent.ID, err.Error())
	}

	// Send error chunk
	errChunk := providers.StreamChunk{
		Type:  providers.StreamChunkTypeError,
		Error: err,
	}

	select {
	case outChan <- errChunk:
	default:
	}
}

// handleCancellation handles cancellation during message execution
func (p *Provider) handleCancellation(pa *providerAgent, outChan chan providers.StreamChunk) {
	pa.mu.Lock()
	pa.agent.Status = providers.AgentStatusCancelled
	now := time.Now()
	pa.agent.CompletedAt = &now
	pa.mu.Unlock()

	// Persist cancellation if enabled
	if p.config.PersistExecutions && p.execRepo != nil {
		_ = p.execRepo.Cancel(pa.agent.ID)
	}

	// Send error chunk
	errChunk := providers.StreamChunk{
		Type:  providers.StreamChunkTypeError,
		Error: errors.New("cancelled"),
	}

	select {
	case outChan <- errChunk:
	default:
	}
}

// GetMessages retrieves the message history for an agent
func (p *Provider) GetMessages(ctx context.Context, agentID string) ([]providers.Message, error) {
	p.agentsMu.RLock()
	pa, ok := p.agents[agentID]
	p.agentsMu.RUnlock()

	if !ok {
		return nil, ErrAgentNotFound
	}

	pa.mu.RLock()
	defer pa.mu.RUnlock()

	// Return a copy of the messages
	messages := make([]providers.Message, len(pa.messages))
	copy(messages, pa.messages)

	return messages, nil
}

// StopAgent stops a running agent
func (p *Provider) StopAgent(ctx context.Context, agentID string) error {
	p.agentsMu.RLock()
	pa, ok := p.agents[agentID]
	p.agentsMu.RUnlock()

	if !ok {
		return ErrAgentNotFound
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	// If there's an internal execution, cancel it
	if pa.execution != nil {
		_ = p.agentManager.CancelExecution(pa.internalID)
	}

	pa.agent.Status = providers.AgentStatusCancelled
	now := time.Now()
	pa.agent.CompletedAt = &now

	// Persist cancellation if enabled
	if p.config.PersistExecutions && p.execRepo != nil {
		_ = p.execRepo.Cancel(pa.agent.ID)
	}

	return nil
}

// SupportsStreaming returns whether the provider supports streaming
func (p *Provider) SupportsStreaming() bool {
	return true
}

// Capabilities returns the provider's capabilities
func (p *Provider) Capabilities() providers.ProviderCapabilities {
	return providers.ProviderCapabilities{
		Streaming:    true,
		Tools:        true,
		Vision:       true,
		MultiAgent:   true,
		Sandbox:      true,
		CostTracking: p.config.TrackCosts,
		Persistence:  p.config.PersistExecutions && p.execRepo != nil,
	}
}

// ListAgents returns all agents for a user
func (p *Provider) ListAgents(ctx context.Context, userID string) ([]*providers.Agent, error) {
	p.agentsMu.RLock()
	defer p.agentsMu.RUnlock()

	agents := make([]*providers.Agent, 0)
	for _, pa := range p.agents {
		pa.mu.RLock()
		if pa.agent.UserID == userID {
			agentCopy := *pa.agent
			agentCopy.Usage = pa.totalUsage
			agentCopy.Cost = pa.totalCost
			agents = append(agents, &agentCopy)
		}
		pa.mu.RUnlock()
	}

	return agents, nil
}

// DeleteAgent deletes an agent
func (p *Provider) DeleteAgent(ctx context.Context, agentID string) error {
	p.agentsMu.Lock()
	pa, ok := p.agents[agentID]
	if ok {
		// Stop if running
		pa.mu.Lock()
		if pa.agent.Status == providers.AgentStatusRunning && pa.execution != nil {
			_ = p.agentManager.CancelExecution(pa.internalID)
		}
		pa.mu.Unlock()

		delete(p.agents, agentID)
	}
	p.agentsMu.Unlock()

	if !ok {
		return ErrAgentNotFound
	}

	// Delete from database if enabled
	if p.config.PersistExecutions && p.execRepo != nil {
		_ = p.execRepo.Delete(agentID)
	}

	return nil
}

// GetUsageStats returns usage statistics for a user
func (p *Provider) GetUsageStats(ctx context.Context, userID string, since time.Time) (*repository.UsageStats, error) {
	if p.execRepo == nil {
		return nil, errors.New("persistence not configured")
	}

	return p.execRepo.GetUsageStats(userID, since)
}
