package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/llm"
)

// SwarmStrategy defines how agents in a swarm coordinate
type SwarmStrategy string

const (
	// StrategyParallel - All agents work independently in parallel
	StrategyParallel SwarmStrategy = "parallel"
	// StrategyPipeline - Agents work in sequence, each building on previous output
	StrategyPipeline SwarmStrategy = "pipeline"
	// StrategyDebate - Agents debate/critique each other's outputs
	StrategyDebate SwarmStrategy = "debate"
	// StrategyConsensus - Agents work to reach consensus on a solution
	StrategyConsensus SwarmStrategy = "consensus"
	// StrategyMapReduce - Split task, parallel execution, then combine results
	StrategyMapReduce SwarmStrategy = "map_reduce"
	// StrategySpecialist - Route tasks to specialized agents
	StrategySpecialist SwarmStrategy = "specialist"
)

// AgentRole defines the specialization of an agent
type AgentRole string

const (
	RoleGeneral    AgentRole = "general"
	RolePlanner    AgentRole = "planner"
	RoleCoder      AgentRole = "coder"
	RoleReviewer   AgentRole = "reviewer"
	RoleResearcher AgentRole = "researcher"
	RoleWriter     AgentRole = "writer"
	RoleAnalyst    AgentRole = "analyst"
	RoleDebugger   AgentRole = "debugger"
	RoleTester     AgentRole = "tester"
	RoleSynthesizer AgentRole = "synthesizer"
)

// SwarmConfig configures a multi-agent swarm
type SwarmConfig struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Strategy        SwarmStrategy     `json:"strategy"`
	MaxAgents       int               `json:"max_agents"`
	Timeout         time.Duration     `json:"timeout"`
	AgentConfigs    []AgentRoleConfig `json:"agent_configs"`
	SynthesizerConfig *AgentConfig    `json:"synthesizer_config,omitempty"` // For combining results
}

// AgentRoleConfig configures an agent with a specific role
type AgentRoleConfig struct {
	Role        AgentRole   `json:"role"`
	Config      AgentConfig `json:"config"`
	Count       int         `json:"count"`       // Number of agents with this role
	SystemPrompt string     `json:"system_prompt,omitempty"` // Override system prompt
}

// Swarm represents a coordinated group of agents working together
type Swarm struct {
	ID          string        `json:"id"`
	Config      SwarmConfig   `json:"config"`
	Status      SwarmStatus   `json:"status"`
	Agents      []*SwarmAgent `json:"agents"`
	Messages    []SwarmMessage `json:"messages"` // Inter-agent communication
	Results     []SwarmResult `json:"results"`
	FinalOutput string        `json:"final_output,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty"`

	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	llmManager  *llm.Manager
	events      chan *SwarmEvent
}

// SwarmStatus represents the status of a swarm
type SwarmStatus string

const (
	SwarmStatusPending   SwarmStatus = "pending"
	SwarmStatusRunning   SwarmStatus = "running"
	SwarmStatusCompleted SwarmStatus = "completed"
	SwarmStatusFailed    SwarmStatus = "failed"
	SwarmStatusCancelled SwarmStatus = "cancelled"
)

// SwarmAgent represents an agent within a swarm
type SwarmAgent struct {
	ID        string      `json:"id"`
	Role      AgentRole   `json:"role"`
	Agent     *Agent      `json:"-"`
	Status    AgentStatus `json:"status"`
	Input     string      `json:"input,omitempty"`
	Output    string      `json:"output,omitempty"`
	StartedAt *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SwarmMessage represents communication between agents
type SwarmMessage struct {
	ID        string    `json:"id"`
	FromAgent string    `json:"from_agent"`
	ToAgent   string    `json:"to_agent,omitempty"` // Empty means broadcast
	Type      string    `json:"type"` // "output", "request", "feedback", "critique"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SwarmResult represents an individual agent's result
type SwarmResult struct {
	AgentID   string        `json:"agent_id"`
	Role      AgentRole     `json:"role"`
	Output    string        `json:"output"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SwarmEvent represents an event from a swarm
type SwarmEvent struct {
	SwarmID   string                 `json:"swarm_id"`
	Type      SwarmEventType         `json:"type"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Role      AgentRole              `json:"role,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SwarmEventType represents types of swarm events
type SwarmEventType string

const (
	SwarmEventStarted         SwarmEventType = "swarm_started"
	SwarmEventAgentStarted    SwarmEventType = "agent_started"
	SwarmEventAgentOutput     SwarmEventType = "agent_output"
	SwarmEventAgentCompleted  SwarmEventType = "agent_completed"
	SwarmEventAgentFailed     SwarmEventType = "agent_failed"
	SwarmEventMessage         SwarmEventType = "message"
	SwarmEventSynthesizing    SwarmEventType = "synthesizing"
	SwarmEventCompleted       SwarmEventType = "swarm_completed"
	SwarmEventFailed          SwarmEventType = "swarm_failed"
	SwarmEventCancelled       SwarmEventType = "swarm_cancelled"
	SwarmEventProgress        SwarmEventType = "progress"
)

// Orchestrator manages multi-agent swarms
type Orchestrator struct {
	llmManager *llm.Manager
	swarms     map[string]*Swarm
	swarmsMu   sync.RWMutex

	// Default role prompts
	rolePrompts map[AgentRole]string
}

// NewOrchestrator creates a new multi-agent orchestrator
func NewOrchestrator(llmManager *llm.Manager) *Orchestrator {
	o := &Orchestrator{
		llmManager:  llmManager,
		swarms:      make(map[string]*Swarm),
		rolePrompts: make(map[AgentRole]string),
	}
	o.initDefaultRolePrompts()
	return o
}

// initDefaultRolePrompts sets up default system prompts for each role
func (o *Orchestrator) initDefaultRolePrompts() {
	o.rolePrompts[RoleGeneral] = "You are a helpful AI assistant."

	o.rolePrompts[RolePlanner] = `You are a planning specialist. Your role is to:
- Break down complex tasks into actionable steps
- Identify dependencies between tasks
- Create clear, structured plans
- Prioritize tasks effectively
Focus on creating practical, implementable plans.`

	o.rolePrompts[RoleCoder] = `You are an expert software developer. Your role is to:
- Write clean, efficient, well-documented code
- Follow best practices and design patterns
- Consider edge cases and error handling
- Write code that is maintainable and testable
Focus on producing high-quality, working code.`

	o.rolePrompts[RoleReviewer] = `You are a code review specialist. Your role is to:
- Identify bugs, security issues, and code smells
- Suggest improvements for readability and performance
- Verify adherence to best practices
- Provide constructive, actionable feedback
Be thorough but constructive in your reviews.`

	o.rolePrompts[RoleResearcher] = `You are a research specialist. Your role is to:
- Gather and synthesize information on topics
- Identify relevant sources and references
- Provide comprehensive analysis
- Present findings clearly and objectively
Focus on accuracy and thoroughness.`

	o.rolePrompts[RoleWriter] = `You are a technical writer. Your role is to:
- Create clear, well-structured documentation
- Explain complex concepts simply
- Write for the target audience
- Ensure consistency and accuracy
Focus on clarity and readability.`

	o.rolePrompts[RoleAnalyst] = `You are an analytical specialist. Your role is to:
- Analyze data and identify patterns
- Evaluate options and trade-offs
- Provide data-driven recommendations
- Present analysis clearly
Focus on insights and actionable conclusions.`

	o.rolePrompts[RoleDebugger] = `You are a debugging specialist. Your role is to:
- Identify root causes of issues
- Trace execution flow and state
- Propose targeted fixes
- Verify fixes don't introduce new issues
Be systematic and thorough in debugging.`

	o.rolePrompts[RoleTester] = `You are a testing specialist. Your role is to:
- Design comprehensive test cases
- Identify edge cases and failure modes
- Write clear test specifications
- Verify functionality meets requirements
Focus on coverage and reliability.`

	o.rolePrompts[RoleSynthesizer] = `You are a synthesis specialist. Your role is to:
- Combine multiple inputs into coherent output
- Identify common themes and key points
- Resolve conflicts between different perspectives
- Create a unified, comprehensive response
Focus on creating a cohesive final result.`
}

// CreateSwarm creates a new swarm with the given configuration
func (o *Orchestrator) CreateSwarm(config SwarmConfig) *Swarm {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.MaxAgents == 0 {
		config.MaxAgents = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Minute
	}

	swarm := &Swarm{
		ID:         config.ID,
		Config:     config,
		Status:     SwarmStatusPending,
		Agents:     make([]*SwarmAgent, 0),
		Messages:   make([]SwarmMessage, 0),
		Results:    make([]SwarmResult, 0),
		CreatedAt:  time.Now(),
		llmManager: o.llmManager,
		events:     make(chan *SwarmEvent, 100),
	}

	o.swarmsMu.Lock()
	o.swarms[swarm.ID] = swarm
	o.swarmsMu.Unlock()

	return swarm
}

// RunSwarm executes a swarm with the given task
func (o *Orchestrator) RunSwarm(ctx context.Context, swarmID string, task string) error {
	o.swarmsMu.RLock()
	swarm, ok := o.swarms[swarmID]
	o.swarmsMu.RUnlock()

	if !ok {
		return fmt.Errorf("swarm not found: %s", swarmID)
	}

	swarm.mu.Lock()
	if swarm.Status == SwarmStatusRunning {
		swarm.mu.Unlock()
		return fmt.Errorf("swarm already running")
	}

	swarm.ctx, swarm.cancel = context.WithTimeout(ctx, swarm.Config.Timeout)
	swarm.Status = SwarmStatusRunning
	now := time.Now()
	swarm.StartedAt = &now
	swarm.mu.Unlock()

	// Emit started event
	swarm.emitEvent(SwarmEventStarted, "", "", nil)

	// Run based on strategy
	go func() {
		defer close(swarm.events)

		var err error
		switch swarm.Config.Strategy {
		case StrategyParallel:
			err = o.runParallelStrategy(swarm, task)
		case StrategyPipeline:
			err = o.runPipelineStrategy(swarm, task)
		case StrategyDebate:
			err = o.runDebateStrategy(swarm, task)
		case StrategyMapReduce:
			err = o.runMapReduceStrategy(swarm, task)
		case StrategySpecialist:
			err = o.runSpecialistStrategy(swarm, task)
		default:
			err = o.runParallelStrategy(swarm, task)
		}

		swarm.mu.Lock()
		now := time.Now()
		swarm.CompletedAt = &now
		if err != nil {
			swarm.Status = SwarmStatusFailed
			swarm.Error = err.Error()
			swarm.emitEvent(SwarmEventFailed, "", "", map[string]interface{}{"error": err.Error()})
		} else {
			swarm.Status = SwarmStatusCompleted
			swarm.emitEvent(SwarmEventCompleted, "", "", map[string]interface{}{"output": swarm.FinalOutput})
		}
		swarm.mu.Unlock()
	}()

	return nil
}

// runParallelStrategy runs all agents in parallel
func (o *Orchestrator) runParallelStrategy(swarm *Swarm, task string) error {
	// Create agents for each role config
	agents := o.createAgentsFromConfig(swarm)

	var wg sync.WaitGroup
	resultsChan := make(chan SwarmResult, len(agents))

	// Run all agents in parallel
	for _, sa := range agents {
		wg.Add(1)
		go func(agent *SwarmAgent) {
			defer wg.Done()
			result := o.runAgent(swarm, agent, task)
			resultsChan <- result
		}(sa)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		swarm.mu.Lock()
		swarm.Results = append(swarm.Results, result)
		swarm.mu.Unlock()
	}

	// Synthesize results
	return o.synthesizeResults(swarm)
}

// runPipelineStrategy runs agents in sequence, passing output to next agent
func (o *Orchestrator) runPipelineStrategy(swarm *Swarm, task string) error {
	agents := o.createAgentsFromConfig(swarm)

	currentInput := task

	for _, sa := range agents {
		select {
		case <-swarm.ctx.Done():
			return swarm.ctx.Err()
		default:
		}

		sa.Input = currentInput
		result := o.runAgent(swarm, sa, currentInput)

		swarm.mu.Lock()
		swarm.Results = append(swarm.Results, result)
		swarm.mu.Unlock()

		if !result.Success {
			return fmt.Errorf("agent %s failed: %s", sa.ID, result.Error)
		}

		// Pass output to next agent
		currentInput = result.Output

		// Record inter-agent message
		swarm.mu.Lock()
		swarm.Messages = append(swarm.Messages, SwarmMessage{
			ID:        uuid.New().String(),
			FromAgent: sa.ID,
			Type:      "output",
			Content:   result.Output,
			Timestamp: time.Now(),
		})
		swarm.mu.Unlock()
	}

	// Final output is the last agent's output
	swarm.mu.Lock()
	if len(swarm.Results) > 0 {
		swarm.FinalOutput = swarm.Results[len(swarm.Results)-1].Output
	}
	swarm.mu.Unlock()

	return nil
}

// runDebateStrategy has agents critique and improve each other's work
func (o *Orchestrator) runDebateStrategy(swarm *Swarm, task string) error {
	agents := o.createAgentsFromConfig(swarm)
	if len(agents) < 2 {
		return fmt.Errorf("debate strategy requires at least 2 agents")
	}

	// Round 1: All agents produce initial responses
	var wg sync.WaitGroup
	for _, sa := range agents {
		wg.Add(1)
		go func(agent *SwarmAgent) {
			defer wg.Done()
			result := o.runAgent(swarm, agent, task)
			swarm.mu.Lock()
			swarm.Results = append(swarm.Results, result)
			swarm.mu.Unlock()
		}(sa)
	}
	wg.Wait()

	// Round 2: Each agent critiques others and refines
	critiques := make([]string, 0)
	for i, sa := range agents {
		// Gather other agents' outputs for critique
		otherOutputs := ""
		for j, result := range swarm.Results {
			if i != j {
				otherOutputs += fmt.Sprintf("\n--- Agent %d's response ---\n%s\n", j+1, result.Output)
			}
		}

		critiquePrompt := fmt.Sprintf(`Original task: %s

Your initial response: %s

Other agents' responses: %s

Please:
1. Critique the other responses - identify strengths and weaknesses
2. Reflect on your own response considering the other perspectives
3. Provide an improved final response that incorporates the best ideas`,
			task, swarm.Results[i].Output, otherOutputs)

		result := o.runAgent(swarm, sa, critiquePrompt)
		critiques = append(critiques, result.Output)
	}

	// Synthesize final answer
	return o.synthesizeResults(swarm)
}

// runMapReduceStrategy splits task, runs in parallel, then combines
func (o *Orchestrator) runMapReduceStrategy(swarm *Swarm, task string) error {
	// First, use a planner agent to split the task
	plannerConfig := AgentConfig{
		Provider:     swarm.Config.AgentConfigs[0].Config.Provider,
		Model:        swarm.Config.AgentConfigs[0].Config.Model,
		SystemPrompt: o.rolePrompts[RolePlanner],
	}

	plannerAgent := NewAgent(plannerConfig, swarm.llmManager)
	planTask := NewTask(fmt.Sprintf(`Break down this task into %d independent subtasks that can be worked on in parallel:

Task: %s

Output a numbered list of subtasks, one per line.`, len(swarm.Config.AgentConfigs), task))

	if err := plannerAgent.Start(swarm.ctx, planTask); err != nil {
		return err
	}

	// Wait for planner result
	var subtasks string
	select {
	case result := <-plannerAgent.Results():
		if !result.Success {
			return fmt.Errorf("planning failed: %s", result.Error)
		}
		subtasks = result.Output
	case <-swarm.ctx.Done():
		return swarm.ctx.Err()
	}

	// Create agents and assign subtasks
	agents := o.createAgentsFromConfig(swarm)

	// Simple subtask distribution (in practice, would parse the planner output)
	var wg sync.WaitGroup
	for i, sa := range agents {
		wg.Add(1)
		agentTask := fmt.Sprintf("As part of a larger task, work on this subtask:\n\nOverall task: %s\n\nSubtasks identified:\n%s\n\nYou are agent %d. Focus on your portion of the work.",
			task, subtasks, i+1)

		go func(agent *SwarmAgent, t string) {
			defer wg.Done()
			result := o.runAgent(swarm, agent, t)
			swarm.mu.Lock()
			swarm.Results = append(swarm.Results, result)
			swarm.mu.Unlock()
		}(sa, agentTask)
	}
	wg.Wait()

	// Reduce: Synthesize all results
	return o.synthesizeResults(swarm)
}

// runSpecialistStrategy routes tasks to specialized agents
func (o *Orchestrator) runSpecialistStrategy(swarm *Swarm, task string) error {
	// First, analyze the task to determine which specialists to involve
	analyzerConfig := AgentConfig{
		Provider:     swarm.Config.AgentConfigs[0].Config.Provider,
		Model:        swarm.Config.AgentConfigs[0].Config.Model,
		SystemPrompt: "You are a task analyzer. Given a task, identify which specialist roles would be most helpful.",
	}

	analyzerAgent := NewAgent(analyzerConfig, swarm.llmManager)

	// Build list of available roles
	availableRoles := make([]string, 0)
	for _, rc := range swarm.Config.AgentConfigs {
		availableRoles = append(availableRoles, string(rc.Role))
	}

	analyzeTask := NewTask(fmt.Sprintf(`Analyze this task and determine which specialists should work on it:

Task: %s

Available specialists: %v

For each specialist that should be involved, explain what aspect of the task they should handle.`, task, availableRoles))

	if err := analyzerAgent.Start(swarm.ctx, analyzeTask); err != nil {
		return err
	}

	// Wait for analysis
	select {
	case result := <-analyzerAgent.Results():
		if !result.Success {
			return fmt.Errorf("analysis failed: %s", result.Error)
		}
	case <-swarm.ctx.Done():
		return swarm.ctx.Err()
	}

	// Run all configured specialists in parallel
	agents := o.createAgentsFromConfig(swarm)

	var wg sync.WaitGroup
	for _, sa := range agents {
		wg.Add(1)
		specialistTask := fmt.Sprintf(`You are a %s specialist.

Task: %s

Apply your expertise to this task. Focus on the aspects most relevant to your specialty.`, sa.Role, task)

		go func(agent *SwarmAgent, t string) {
			defer wg.Done()
			result := o.runAgent(swarm, agent, t)
			swarm.mu.Lock()
			swarm.Results = append(swarm.Results, result)
			swarm.mu.Unlock()
		}(sa, specialistTask)
	}
	wg.Wait()

	return o.synthesizeResults(swarm)
}

// createAgentsFromConfig creates SwarmAgents from the config
func (o *Orchestrator) createAgentsFromConfig(swarm *Swarm) []*SwarmAgent {
	agents := make([]*SwarmAgent, 0)

	for _, rc := range swarm.Config.AgentConfigs {
		count := rc.Count
		if count == 0 {
			count = 1
		}

		for i := 0; i < count; i++ {
			config := rc.Config
			config.ID = uuid.New().String()

			// Set role-specific system prompt if not overridden
			if config.SystemPrompt == "" {
				if rc.SystemPrompt != "" {
					config.SystemPrompt = rc.SystemPrompt
				} else if prompt, ok := o.rolePrompts[rc.Role]; ok {
					config.SystemPrompt = prompt
				}
			}

			agent := NewAgent(config, swarm.llmManager)

			sa := &SwarmAgent{
				ID:     agent.ID,
				Role:   rc.Role,
				Agent:  agent,
				Status: AgentStatusIdle,
			}

			agents = append(agents, sa)
			swarm.Agents = append(swarm.Agents, sa)
		}
	}

	return agents
}

// runAgent runs a single agent and returns the result
func (o *Orchestrator) runAgent(swarm *Swarm, sa *SwarmAgent, input string) SwarmResult {
	startTime := time.Now()
	sa.Input = input
	sa.Status = AgentStatusRunning
	now := time.Now()
	sa.StartedAt = &now

	swarm.emitEvent(SwarmEventAgentStarted, sa.ID, sa.Role, map[string]interface{}{
		"input": input,
	})

	task := NewTask(input)
	if err := sa.Agent.Start(swarm.ctx, task); err != nil {
		sa.Status = AgentStatusFailed
		swarm.emitEvent(SwarmEventAgentFailed, sa.ID, sa.Role, map[string]interface{}{
			"error": err.Error(),
		})
		return SwarmResult{
			AgentID:  sa.ID,
			Role:     sa.Role,
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}
	}

	// Forward streaming events
	go func() {
		for event := range sa.Agent.Events() {
			if event.Type == AgentEventStreamChunk {
				if delta, ok := event.Data["delta"].(string); ok {
					swarm.emitEvent(SwarmEventAgentOutput, sa.ID, sa.Role, map[string]interface{}{
						"delta": delta,
					})
				}
			}
		}
	}()

	// Wait for result
	select {
	case result := <-sa.Agent.Results():
		now := time.Now()
		sa.CompletedAt = &now
		sa.Output = result.Output

		if result.Success {
			sa.Status = AgentStatusCompleted
			swarm.emitEvent(SwarmEventAgentCompleted, sa.ID, sa.Role, map[string]interface{}{
				"output":   result.Output,
				"duration": result.Duration.Milliseconds(),
			})
		} else {
			sa.Status = AgentStatusFailed
			swarm.emitEvent(SwarmEventAgentFailed, sa.ID, sa.Role, map[string]interface{}{
				"error": result.Error,
			})
		}

		return SwarmResult{
			AgentID:  sa.ID,
			Role:     sa.Role,
			Output:   result.Output,
			Success:  result.Success,
			Error:    result.Error,
			Duration: result.Duration,
		}

	case <-swarm.ctx.Done():
		sa.Status = AgentStatusCancelled
		return SwarmResult{
			AgentID:  sa.ID,
			Role:     sa.Role,
			Success:  false,
			Error:    "cancelled",
			Duration: time.Since(startTime),
		}
	}
}

// synthesizeResults combines all agent results into a final output
func (o *Orchestrator) synthesizeResults(swarm *Swarm) error {
	swarm.emitEvent(SwarmEventSynthesizing, "", "", nil)

	// If only one result, use it directly
	if len(swarm.Results) == 1 {
		swarm.mu.Lock()
		swarm.FinalOutput = swarm.Results[0].Output
		swarm.mu.Unlock()
		return nil
	}

	// Use synthesizer config if provided, otherwise use first agent's config
	var synthConfig AgentConfig
	if swarm.Config.SynthesizerConfig != nil {
		synthConfig = *swarm.Config.SynthesizerConfig
	} else {
		synthConfig = swarm.Config.AgentConfigs[0].Config
	}
	synthConfig.SystemPrompt = o.rolePrompts[RoleSynthesizer]
	synthConfig.ID = uuid.New().String()

	// Build synthesis prompt
	resultsText := ""
	for i, result := range swarm.Results {
		resultsText += fmt.Sprintf("\n--- Agent %d (%s) ---\n%s\n", i+1, result.Role, result.Output)
	}

	synthPrompt := fmt.Sprintf(`You have received outputs from multiple specialized agents working on a task.

Agent outputs:%s

Your task is to:
1. Identify the key insights and contributions from each agent
2. Resolve any conflicts or contradictions
3. Synthesize a comprehensive, coherent final response
4. Ensure nothing important is lost in the synthesis

Provide the synthesized final response:`, resultsText)

	synthesizer := NewAgent(synthConfig, swarm.llmManager)
	synthTask := NewTask(synthPrompt)

	if err := synthesizer.Start(swarm.ctx, synthTask); err != nil {
		return err
	}

	select {
	case result := <-synthesizer.Results():
		if !result.Success {
			return fmt.Errorf("synthesis failed: %s", result.Error)
		}
		swarm.mu.Lock()
		swarm.FinalOutput = result.Output
		swarm.mu.Unlock()
		return nil

	case <-swarm.ctx.Done():
		return swarm.ctx.Err()
	}
}

// emitEvent sends an event from the swarm
func (s *Swarm) emitEvent(eventType SwarmEventType, agentID string, role AgentRole, data map[string]interface{}) {
	select {
	case s.events <- &SwarmEvent{
		SwarmID:   s.ID,
		Type:      eventType,
		AgentID:   agentID,
		Role:      role,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip
	}
}

// Events returns the swarm's event channel
func (s *Swarm) Events() <-chan *SwarmEvent {
	return s.events
}

// Stop cancels the swarm execution
func (s *Swarm) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}
	s.Status = SwarmStatusCancelled
}

// GetSwarm returns a swarm by ID
func (o *Orchestrator) GetSwarm(id string) (*Swarm, error) {
	o.swarmsMu.RLock()
	defer o.swarmsMu.RUnlock()

	swarm, ok := o.swarms[id]
	if !ok {
		return nil, fmt.Errorf("swarm not found: %s", id)
	}
	return swarm, nil
}

// ListSwarms returns all swarms
func (o *Orchestrator) ListSwarms() []*Swarm {
	o.swarmsMu.RLock()
	defer o.swarmsMu.RUnlock()

	swarms := make([]*Swarm, 0, len(o.swarms))
	for _, s := range o.swarms {
		swarms = append(swarms, s)
	}
	return swarms
}

// CancelSwarm cancels a running swarm
func (o *Orchestrator) CancelSwarm(id string) error {
	swarm, err := o.GetSwarm(id)
	if err != nil {
		return err
	}
	swarm.Stop()
	return nil
}
