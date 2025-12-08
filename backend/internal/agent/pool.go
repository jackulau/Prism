package agent

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/llm"
)

// PoolConfig holds configuration for the agent pool
type PoolConfig struct {
	MaxConcurrentAgents int           `json:"max_concurrent_agents"`
	DefaultTimeout      time.Duration `json:"default_timeout"`
	QueueCapacity       int           `json:"queue_capacity"`
}

// DefaultPoolConfig returns the default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConcurrentAgents: 10,
		DefaultTimeout:      5 * time.Minute,
		QueueCapacity:       100,
	}
}

// Pool manages a pool of agents for parallel task execution
type Pool struct {
	config     PoolConfig
	llmManager *llm.Manager

	// Agent tracking
	agents    map[string]*Agent
	agentsMu  sync.RWMutex

	// Task queue
	taskQueue *TaskQueue
	queueMu   sync.Mutex

	// Concurrency control
	semaphore chan struct{}
	wg        sync.WaitGroup

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Event broadcasting
	eventBroadcaster *EventBroadcaster

	// State
	running bool
	runMu   sync.RWMutex
}

// NewPool creates a new agent pool
func NewPool(llmManager *llm.Manager, config PoolConfig) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		config:           config,
		llmManager:       llmManager,
		agents:           make(map[string]*Agent),
		taskQueue:        NewTaskQueue(config.QueueCapacity),
		semaphore:        make(chan struct{}, config.MaxConcurrentAgents),
		ctx:              ctx,
		cancel:           cancel,
		eventBroadcaster: NewEventBroadcaster(),
	}
}

// Start begins the pool's background processing
func (p *Pool) Start() {
	p.runMu.Lock()
	if p.running {
		p.runMu.Unlock()
		return
	}
	p.running = true
	p.runMu.Unlock()

	go p.processQueue()
}

// Stop gracefully shuts down the pool
func (p *Pool) Stop() {
	p.runMu.Lock()
	if !p.running {
		p.runMu.Unlock()
		return
	}
	p.running = false
	p.runMu.Unlock()

	p.cancel()
	p.wg.Wait()
	p.eventBroadcaster.Close()
}

// processQueue continuously processes tasks from the queue
func (p *Pool) processQueue() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			p.queueMu.Lock()
			task := p.taskQueue.Pop()
			p.queueMu.Unlock()

			if task == nil {
				// No tasks, wait a bit before checking again
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Execute the task
			p.executeTask(task)
		}
	}
}

// Submit submits a task to the pool for execution
func (p *Pool) Submit(task *Task, agentConfig AgentConfig) (string, error) {
	p.runMu.RLock()
	if !p.running {
		p.runMu.RUnlock()
		return "", ErrPoolNotRunning
	}
	p.runMu.RUnlock()

	// Store agent config in task if not already set
	if task.AgentConfig == nil {
		task.AgentConfig = &agentConfig
	}

	p.queueMu.Lock()
	if !p.taskQueue.Push(task) {
		p.queueMu.Unlock()
		return "", ErrQueueFull
	}
	p.queueMu.Unlock()

	return task.ID, nil
}

// SubmitImmediate creates an agent and executes a task immediately (bypassing queue)
func (p *Pool) SubmitImmediate(task *Task, agentConfig AgentConfig) (*Agent, error) {
	p.runMu.RLock()
	if !p.running {
		p.runMu.RUnlock()
		return nil, ErrPoolNotRunning
	}
	p.runMu.RUnlock()

	// Create the agent
	agent := NewAgent(agentConfig, p.llmManager)

	// Register the agent
	p.agentsMu.Lock()
	p.agents[agent.ID] = agent
	p.agentsMu.Unlock()

	// Execute asynchronously
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.runAgent(agent, task)
	}()

	return agent, nil
}

// SubmitBatch submits multiple tasks for parallel execution
func (p *Pool) SubmitBatch(batch *BatchTask, defaultConfig AgentConfig) (*BatchExecution, error) {
	p.runMu.RLock()
	if !p.running {
		p.runMu.RUnlock()
		return nil, ErrPoolNotRunning
	}
	p.runMu.RUnlock()

	execution := &BatchExecution{
		ID:          batch.ID,
		Batch:       batch,
		Agents:      make([]*Agent, len(batch.Tasks)),
		Results:     make([]*AgentResult, len(batch.Tasks)),
		resultsChan: make(chan *AgentResult, len(batch.Tasks)),
		done:        make(chan struct{}),
	}

	// Create agents for all tasks
	for i, task := range batch.Tasks {
		config := defaultConfig
		if task.AgentConfig != nil {
			config = *task.AgentConfig
		}
		config.ID = uuid.New().String()

		agent := NewAgent(config, p.llmManager)
		execution.Agents[i] = agent

		p.agentsMu.Lock()
		p.agents[agent.ID] = agent
		p.agentsMu.Unlock()
	}

	// Execute based on parallel flag
	if batch.Parallel {
		go p.executeBatchParallel(execution, batch.MaxParallel)
	} else {
		go p.executeBatchSequential(execution)
	}

	return execution, nil
}

// executeBatchParallel executes tasks in parallel with concurrency control
func (p *Pool) executeBatchParallel(execution *BatchExecution, maxParallel int) {
	defer close(execution.done)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxParallel)
	if maxParallel <= 0 {
		semaphore = make(chan struct{}, len(execution.Batch.Tasks))
	}

	for i, task := range execution.Batch.Tasks {
		wg.Add(1)
		go func(idx int, t *Task, agent *Agent) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Run the agent
			p.runAgent(agent, t)

			// Collect result
			select {
			case result := <-agent.Results():
				execution.mu.Lock()
				execution.Results[idx] = result
				execution.mu.Unlock()
				execution.resultsChan <- result
			case <-p.ctx.Done():
				return
			}
		}(i, task, execution.Agents[i])
	}

	wg.Wait()
	close(execution.resultsChan)

	// Update batch status
	execution.Batch.Status = TaskStatusCompleted
	now := time.Now()
	execution.Batch.CompletedAt = &now
}

// executeBatchSequential executes tasks one after another
func (p *Pool) executeBatchSequential(execution *BatchExecution) {
	defer close(execution.done)

	for i, task := range execution.Batch.Tasks {
		agent := execution.Agents[i]

		// Run the agent
		p.runAgent(agent, task)

		// Collect result
		select {
		case result := <-agent.Results():
			execution.mu.Lock()
			execution.Results[i] = result
			execution.mu.Unlock()
			execution.resultsChan <- result
		case <-p.ctx.Done():
			return
		}
	}

	close(execution.resultsChan)

	// Update batch status
	execution.Batch.Status = TaskStatusCompleted
	now := time.Now()
	execution.Batch.CompletedAt = &now
}

// executeTask executes a single task from the queue
func (p *Pool) executeTask(task *Task) {
	// Acquire semaphore
	select {
	case p.semaphore <- struct{}{}:
	case <-p.ctx.Done():
		return
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.semaphore }()

		// Create agent from task config
		config := AgentConfig{
			Provider: "openai", // Default provider
			Model:    "gpt-4",  // Default model
		}
		if task.AgentConfig != nil {
			config = *task.AgentConfig
		}

		agent := NewAgent(config, p.llmManager)

		// Register the agent
		p.agentsMu.Lock()
		p.agents[agent.ID] = agent
		p.agentsMu.Unlock()

		// Run the agent
		p.runAgent(agent, task)
	}()
}

// runAgent runs an agent with a task
func (p *Pool) runAgent(agent *Agent, task *Task) {
	// Set up timeout
	timeout := p.config.DefaultTimeout
	if task.Timeout > 0 {
		timeout = task.Timeout
	}

	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	// Forward events to broadcaster
	go func() {
		for event := range agent.Events() {
			p.eventBroadcaster.Broadcast(event)
		}
	}()

	// Start the agent
	if err := agent.Start(ctx, task); err != nil {
		return
	}

	// Wait for completion or cancellation
	select {
	case <-agent.Results():
		// Agent completed
	case <-ctx.Done():
		agent.Stop()
	}

	// Cleanup agent after some time
	go func() {
		time.Sleep(5 * time.Minute)
		p.agentsMu.Lock()
		delete(p.agents, agent.ID)
		p.agentsMu.Unlock()
	}()
}

// GetAgent returns an agent by ID
func (p *Pool) GetAgent(agentID string) (*Agent, error) {
	p.agentsMu.RLock()
	defer p.agentsMu.RUnlock()

	agent, ok := p.agents[agentID]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// ListAgents returns all active agents
func (p *Pool) ListAgents() []*Agent {
	p.agentsMu.RLock()
	defer p.agentsMu.RUnlock()

	agents := make([]*Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}
	return agents
}

// StopAgent stops a specific agent by ID
func (p *Pool) StopAgent(agentID string) error {
	p.agentsMu.RLock()
	agent, ok := p.agents[agentID]
	p.agentsMu.RUnlock()

	if !ok {
		return ErrAgentNotFound
	}

	agent.Stop()
	return nil
}

// QueueSize returns the current size of the task queue
func (p *Pool) QueueSize() int {
	p.queueMu.Lock()
	defer p.queueMu.Unlock()
	return p.taskQueue.Len()
}

// ActiveAgents returns the number of currently running agents
func (p *Pool) ActiveAgents() int {
	p.agentsMu.RLock()
	defer p.agentsMu.RUnlock()

	count := 0
	for _, agent := range p.agents {
		if agent.GetStatus() == AgentStatusRunning {
			count++
		}
	}
	return count
}

// Subscribe subscribes to agent events
func (p *Pool) Subscribe() <-chan *AgentEvent {
	return p.eventBroadcaster.Subscribe()
}

// Unsubscribe unsubscribes from agent events
func (p *Pool) Unsubscribe(ch <-chan *AgentEvent) {
	p.eventBroadcaster.Unsubscribe(ch)
}

// BatchExecution represents the execution state of a batch of tasks
type BatchExecution struct {
	ID          string         `json:"id"`
	Batch       *BatchTask     `json:"batch"`
	Agents      []*Agent       `json:"agents"`
	Results     []*AgentResult `json:"results"`
	mu          sync.RWMutex
	resultsChan chan *AgentResult
	done        chan struct{}
}

// ResultsChan returns the channel for receiving results as they complete
func (e *BatchExecution) ResultsChan() <-chan *AgentResult {
	return e.resultsChan
}

// Done returns a channel that's closed when all tasks are complete
func (e *BatchExecution) Done() <-chan struct{} {
	return e.done
}

// Wait blocks until all tasks in the batch are complete
func (e *BatchExecution) Wait() {
	<-e.done
}

// GetResults returns all results (blocks until complete)
func (e *BatchExecution) GetResults() []*AgentResult {
	e.Wait()
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]*AgentResult{}, e.Results...)
}

// EventBroadcaster broadcasts events to multiple subscribers
type EventBroadcaster struct {
	subscribers map[chan *AgentEvent]struct{}
	mu          sync.RWMutex
	closed      bool
}

// NewEventBroadcaster creates a new event broadcaster
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		subscribers: make(map[chan *AgentEvent]struct{}),
	}
}

// Subscribe creates a new subscription channel
func (b *EventBroadcaster) Subscribe() <-chan *AgentEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan *AgentEvent)
		close(ch)
		return ch
	}

	ch := make(chan *AgentEvent, 100)
	b.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscription
func (b *EventBroadcaster) Unsubscribe(ch <-chan *AgentEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Convert to send channel for map lookup
	for sub := range b.subscribers {
		if sub == ch {
			close(sub)
			delete(b.subscribers, sub)
			return
		}
	}
}

// Broadcast sends an event to all subscribers
func (b *EventBroadcaster) Broadcast(event *AgentEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Subscriber channel full, skip
		}
	}
}

// Close closes all subscriber channels
func (b *EventBroadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	for ch := range b.subscribers {
		close(ch)
	}
	b.subscribers = make(map[chan *AgentEvent]struct{})
}
