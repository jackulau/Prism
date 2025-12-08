package agent

import (
	"context"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// ManagerConfig holds configuration for the agent manager
type ManagerConfig struct {
	Pool PoolConfig `json:"pool"`
}

// DefaultManagerConfig returns the default manager configuration
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		Pool: DefaultPoolConfig(),
	}
}

// Manager coordinates agent operations and provides a high-level API
type Manager struct {
	config     ManagerConfig
	llmManager *llm.Manager
	pool       *Pool

	// Multi-agent orchestrator for swarm operations
	orchestrator *Orchestrator

	// Agent configs registry (named/reusable agent configurations)
	configs   map[string]AgentConfig
	configsMu sync.RWMutex

	// Execution tracking
	executions   map[string]*Execution
	executionsMu sync.RWMutex

	// State
	running bool
	mu      sync.RWMutex
}

// NewManager creates a new agent manager
func NewManager(llmManager *llm.Manager, config ManagerConfig) *Manager {
	pool := NewPool(llmManager, config.Pool)
	orchestrator := NewOrchestrator(llmManager)

	return &Manager{
		config:       config,
		llmManager:   llmManager,
		pool:         pool,
		orchestrator: orchestrator,
		configs:      make(map[string]AgentConfig),
		executions:   make(map[string]*Execution),
	}
}

// Start initializes and starts the manager
func (m *Manager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.pool.Start()
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	m.pool.Stop()
}

// RegisterConfig registers a named agent configuration for reuse
func (m *Manager) RegisterConfig(name string, config AgentConfig) {
	m.configsMu.Lock()
	defer m.configsMu.Unlock()
	m.configs[name] = config
}

// GetConfig retrieves a named agent configuration
func (m *Manager) GetConfig(name string) (AgentConfig, bool) {
	m.configsMu.RLock()
	defer m.configsMu.RUnlock()
	config, ok := m.configs[name]
	return config, ok
}

// ListConfigs returns all registered agent configurations
func (m *Manager) ListConfigs() map[string]AgentConfig {
	m.configsMu.RLock()
	defer m.configsMu.RUnlock()

	configs := make(map[string]AgentConfig, len(m.configs))
	for name, config := range m.configs {
		configs[name] = config
	}
	return configs
}

// RunTask executes a single task with a new agent
func (m *Manager) RunTask(ctx context.Context, task *Task, config AgentConfig) (*Execution, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotInitialized
	}
	m.mu.RUnlock()

	// Validate config
	if config.Provider == "" || config.Model == "" {
		return nil, ErrInvalidAgentConfig
	}

	// Create agent and execute
	agent, err := m.pool.SubmitImmediate(task, config)
	if err != nil {
		return nil, err
	}

	execution := &Execution{
		ID:        task.ID,
		Type:      ExecutionTypeSingle,
		Tasks:     []*Task{task},
		Agents:    []*Agent{agent},
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
	}

	m.executionsMu.Lock()
	m.executions[execution.ID] = execution
	m.executionsMu.Unlock()

	// Monitor execution in background
	go m.monitorExecution(execution)

	return execution, nil
}

// RunTaskWithConfig executes a task using a registered configuration
func (m *Manager) RunTaskWithConfig(ctx context.Context, task *Task, configName string) (*Execution, error) {
	config, ok := m.GetConfig(configName)
	if !ok {
		return nil, ErrInvalidAgentConfig
	}
	return m.RunTask(ctx, task, config)
}

// RunParallel executes multiple tasks in parallel
func (m *Manager) RunParallel(ctx context.Context, tasks []*Task, config AgentConfig) (*Execution, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotInitialized
	}
	m.mu.RUnlock()

	// Validate config
	if config.Provider == "" || config.Model == "" {
		return nil, ErrInvalidAgentConfig
	}

	// Create batch task
	batch := NewBatchTask(tasks, true, m.config.Pool.MaxConcurrentAgents)

	// Submit batch
	batchExec, err := m.pool.SubmitBatch(batch, config)
	if err != nil {
		return nil, err
	}

	execution := &Execution{
		ID:             batch.ID,
		Type:           ExecutionTypeParallel,
		Tasks:          tasks,
		Agents:         batchExec.Agents,
		Status:         ExecutionStatusRunning,
		StartedAt:      time.Now(),
		batchExecution: batchExec,
	}

	m.executionsMu.Lock()
	m.executions[execution.ID] = execution
	m.executionsMu.Unlock()

	// Monitor execution in background
	go m.monitorBatchExecution(execution, batchExec)

	return execution, nil
}

// RunSequential executes multiple tasks sequentially
func (m *Manager) RunSequential(ctx context.Context, tasks []*Task, config AgentConfig) (*Execution, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotInitialized
	}
	m.mu.RUnlock()

	// Validate config
	if config.Provider == "" || config.Model == "" {
		return nil, ErrInvalidAgentConfig
	}

	// Create batch task (sequential)
	batch := NewBatchTask(tasks, false, 1)

	// Submit batch
	batchExec, err := m.pool.SubmitBatch(batch, config)
	if err != nil {
		return nil, err
	}

	execution := &Execution{
		ID:             batch.ID,
		Type:           ExecutionTypeSequential,
		Tasks:          tasks,
		Agents:         batchExec.Agents,
		Status:         ExecutionStatusRunning,
		StartedAt:      time.Now(),
		batchExecution: batchExec,
	}

	m.executionsMu.Lock()
	m.executions[execution.ID] = execution
	m.executionsMu.Unlock()

	// Monitor execution in background
	go m.monitorBatchExecution(execution, batchExec)

	return execution, nil
}

// monitorExecution monitors a single task execution
func (m *Manager) monitorExecution(execution *Execution) {
	if len(execution.Agents) == 0 {
		return
	}

	agent := execution.Agents[0]

	// Wait for result
	select {
	case result := <-agent.Results():
		execution.mu.Lock()
		execution.Results = []*AgentResult{result}
		if result.Success {
			execution.Status = ExecutionStatusCompleted
		} else {
			execution.Status = ExecutionStatusFailed
			execution.Error = result.Error
		}
		now := time.Now()
		execution.CompletedAt = &now
		execution.mu.Unlock()
	}
}

// monitorBatchExecution monitors a batch execution
func (m *Manager) monitorBatchExecution(execution *Execution, batchExec *BatchExecution) {
	// Collect results as they come in
	results := make([]*AgentResult, 0, len(execution.Tasks))
	for result := range batchExec.ResultsChan() {
		results = append(results, result)
	}

	// Update execution status
	execution.mu.Lock()
	execution.Results = results
	execution.Status = ExecutionStatusCompleted

	// Check if any failed
	for _, result := range results {
		if !result.Success {
			execution.Status = ExecutionStatusPartiallyCompleted
			break
		}
	}

	now := time.Now()
	execution.CompletedAt = &now
	execution.mu.Unlock()
}

// GetExecution retrieves an execution by ID
func (m *Manager) GetExecution(id string) (*Execution, error) {
	m.executionsMu.RLock()
	defer m.executionsMu.RUnlock()

	execution, ok := m.executions[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return execution, nil
}

// ListExecutions returns all executions
func (m *Manager) ListExecutions() []*Execution {
	m.executionsMu.RLock()
	defer m.executionsMu.RUnlock()

	executions := make([]*Execution, 0, len(m.executions))
	for _, exec := range m.executions {
		executions = append(executions, exec)
	}
	return executions
}

// CancelExecution cancels an execution by ID
func (m *Manager) CancelExecution(id string) error {
	m.executionsMu.RLock()
	execution, ok := m.executions[id]
	m.executionsMu.RUnlock()

	if !ok {
		return ErrTaskNotFound
	}

	// Stop all agents in the execution
	for _, agent := range execution.Agents {
		agent.Stop()
	}

	execution.mu.Lock()
	execution.Status = ExecutionStatusCancelled
	now := time.Now()
	execution.CompletedAt = &now
	execution.mu.Unlock()

	return nil
}

// Subscribe subscribes to agent events from all agents
func (m *Manager) Subscribe() <-chan *AgentEvent {
	return m.pool.Subscribe()
}

// Unsubscribe unsubscribes from agent events
func (m *Manager) Unsubscribe(ch <-chan *AgentEvent) {
	m.pool.Unsubscribe(ch)
}

// Stats returns current statistics about the agent manager
func (m *Manager) Stats() *ManagerStats {
	m.executionsMu.RLock()
	defer m.executionsMu.RUnlock()

	stats := &ManagerStats{
		TotalExecutions:  len(m.executions),
		ActiveAgents:     m.pool.ActiveAgents(),
		QueuedTasks:      m.pool.QueueSize(),
		RegisteredConfigs: len(m.configs),
	}

	for _, exec := range m.executions {
		switch exec.Status {
		case ExecutionStatusRunning:
			stats.RunningExecutions++
		case ExecutionStatusCompleted:
			stats.CompletedExecutions++
		case ExecutionStatusFailed:
			stats.FailedExecutions++
		case ExecutionStatusCancelled:
			stats.CancelledExecutions++
		}
	}

	return stats
}

// ExecutionType represents the type of execution
type ExecutionType string

const (
	ExecutionTypeSingle     ExecutionType = "single"
	ExecutionTypeParallel   ExecutionType = "parallel"
	ExecutionTypeSequential ExecutionType = "sequential"
)

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending            ExecutionStatus = "pending"
	ExecutionStatusRunning            ExecutionStatus = "running"
	ExecutionStatusCompleted          ExecutionStatus = "completed"
	ExecutionStatusPartiallyCompleted ExecutionStatus = "partially_completed"
	ExecutionStatusFailed             ExecutionStatus = "failed"
	ExecutionStatusCancelled          ExecutionStatus = "cancelled"
)

// Execution represents the execution of one or more tasks
type Execution struct {
	ID          string          `json:"id"`
	Type        ExecutionType   `json:"type"`
	Tasks       []*Task         `json:"tasks"`
	Agents      []*Agent        `json:"-"` // Don't expose internal agents
	Results     []*AgentResult  `json:"results,omitempty"`
	Status      ExecutionStatus `json:"status"`
	Error       string          `json:"error,omitempty"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`

	mu             sync.RWMutex
	batchExecution *BatchExecution
}

// GetStatus returns the current execution status (thread-safe)
func (e *Execution) GetStatus() ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Status
}

// GetResults returns the execution results (thread-safe)
func (e *Execution) GetResults() []*AgentResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]*AgentResult{}, e.Results...)
}

// Wait blocks until the execution is complete
func (e *Execution) Wait() {
	if e.batchExecution != nil {
		e.batchExecution.Wait()
		return
	}

	// For single execution, poll status
	for {
		status := e.GetStatus()
		if status != ExecutionStatusRunning && status != ExecutionStatusPending {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// ManagerStats contains statistics about the agent manager
type ManagerStats struct {
	TotalExecutions     int `json:"total_executions"`
	RunningExecutions   int `json:"running_executions"`
	CompletedExecutions int `json:"completed_executions"`
	FailedExecutions    int `json:"failed_executions"`
	CancelledExecutions int `json:"cancelled_executions"`
	ActiveAgents        int `json:"active_agents"`
	QueuedTasks         int `json:"queued_tasks"`
	RegisteredConfigs   int `json:"registered_configs"`
	ActiveSwarms        int `json:"active_swarms"`
}

// ==================== Swarm/Multi-Agent Operations ====================

// CreateSwarm creates a new multi-agent swarm
func (m *Manager) CreateSwarm(config SwarmConfig) *Swarm {
	return m.orchestrator.CreateSwarm(config)
}

// RunSwarm starts executing a swarm with the given task
func (m *Manager) RunSwarm(ctx context.Context, swarmID string, task string) error {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return ErrManagerNotInitialized
	}
	m.mu.RUnlock()

	return m.orchestrator.RunSwarm(ctx, swarmID, task)
}

// GetSwarm retrieves a swarm by ID
func (m *Manager) GetSwarm(id string) (*Swarm, error) {
	return m.orchestrator.GetSwarm(id)
}

// ListSwarms returns all swarms
func (m *Manager) ListSwarms() []*Swarm {
	return m.orchestrator.ListSwarms()
}

// CancelSwarm cancels a running swarm
func (m *Manager) CancelSwarm(id string) error {
	return m.orchestrator.CancelSwarm(id)
}

// RunMultiAgent is a convenience method to run multiple agents with different roles on a task
func (m *Manager) RunMultiAgent(ctx context.Context, task string, strategy SwarmStrategy, agentConfigs []AgentRoleConfig, baseConfig AgentConfig) (*Swarm, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotInitialized
	}
	m.mu.RUnlock()

	// Set defaults for agent configs from base
	for i := range agentConfigs {
		if agentConfigs[i].Config.Provider == "" {
			agentConfigs[i].Config.Provider = baseConfig.Provider
		}
		if agentConfigs[i].Config.Model == "" {
			agentConfigs[i].Config.Model = baseConfig.Model
		}
		if agentConfigs[i].Config.Temperature == 0 {
			agentConfigs[i].Config.Temperature = baseConfig.Temperature
		}
		if agentConfigs[i].Config.MaxTokens == 0 {
			agentConfigs[i].Config.MaxTokens = baseConfig.MaxTokens
		}
	}

	swarmConfig := SwarmConfig{
		Name:         "multi-agent-" + string(strategy),
		Strategy:     strategy,
		AgentConfigs: agentConfigs,
		SynthesizerConfig: &baseConfig,
	}

	swarm := m.CreateSwarm(swarmConfig)
	if err := m.RunSwarm(ctx, swarm.ID, task); err != nil {
		return nil, err
	}

	return swarm, nil
}

// QuickSwarm is a convenience method to quickly run a multi-agent swarm with preset roles
func (m *Manager) QuickSwarm(ctx context.Context, task string, roles []AgentRole, provider, model string) (*Swarm, error) {
	agentConfigs := make([]AgentRoleConfig, len(roles))
	for i, role := range roles {
		agentConfigs[i] = AgentRoleConfig{
			Role:  role,
			Count: 1,
			Config: AgentConfig{
				Provider: provider,
				Model:    model,
			},
		}
	}

	baseConfig := AgentConfig{
		Provider: provider,
		Model:    model,
	}

	return m.RunMultiAgent(ctx, task, StrategyParallel, agentConfigs, baseConfig)
}
