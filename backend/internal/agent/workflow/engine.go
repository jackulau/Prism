package workflow

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/tools"
)

// Repository defines the interface for workflow persistence
type Repository interface {
	Create(workflow *Workflow) error
	GetByID(id string) (*Workflow, error)
	Update(workflow *Workflow) error
	UpdateState(id string, state map[string]interface{}, currentStep int, status WorkflowStatus) error
	List(filter *WorkflowFilter) ([]*Workflow, error)
	Delete(id string) error
}

// Engine manages workflow execution
type Engine struct {
	repository   Repository
	agentMgr     *agent.Manager
	llmManager   *llm.Manager
	toolRegistry *tools.Registry
	executors    map[StepType]StepExecutor

	// Running workflows
	running   map[string]*runningWorkflow
	runningMu sync.RWMutex

	// Event subscribers
	subscribers   []chan *WorkflowEvent
	subscribersMu sync.RWMutex

	mu sync.RWMutex
}

// runningWorkflow tracks a workflow's execution state
type runningWorkflow struct {
	workflow *Workflow
	ctx      context.Context
	cancel   context.CancelFunc
	events   chan *WorkflowEvent
	pauseCh  chan struct{}
	paused   bool
	pausedMu sync.RWMutex
}

// NewEngine creates a new workflow engine
func NewEngine(repo Repository, agentMgr *agent.Manager, llmManager *llm.Manager, toolRegistry *tools.Registry) *Engine {
	e := &Engine{
		repository:   repo,
		agentMgr:     agentMgr,
		llmManager:   llmManager,
		toolRegistry: toolRegistry,
		running:      make(map[string]*runningWorkflow),
		subscribers:  make([]chan *WorkflowEvent, 0),
	}

	// Initialize executors
	e.executors = map[StepType]StepExecutor{
		StepTypeAgent:     NewAgentExecutor(agentMgr, llmManager),
		StepTypeTool:      NewToolExecutor(toolRegistry),
		StepTypeCondition: NewConditionExecutor(),
		StepTypeParallel:  NewParallelExecutor(e),
		StepTypeWait:      NewWaitExecutor(),
		StepTypeTransform: NewTransformExecutor(),
	}

	return e
}

// CreateWorkflow creates a new workflow from a definition
func (e *Engine) CreateWorkflow(userID string, def *WorkflowDefinition) (*Workflow, error) {
	if def.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(def.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	// Assign IDs to steps if not provided
	for i := range def.Steps {
		if def.Steps[i].ID == "" {
			def.Steps[i].ID = fmt.Sprintf("step_%d_%s", i, uuid.New().String()[:8])
		}
	}

	now := time.Now()
	workflow := &Workflow{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        def.Name,
		Description: def.Description,
		Steps:       def.Steps,
		Status:      StatusPending,
		CurrentStep: 0,
		State:       def.InitialState,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if workflow.State == nil {
		workflow.State = make(map[string]interface{})
	}

	if e.repository != nil {
		if err := e.repository.Create(workflow); err != nil {
			return nil, fmt.Errorf("failed to create workflow: %w", err)
		}
	}

	return workflow, nil
}

// StartWorkflow begins executing a workflow
func (e *Engine) StartWorkflow(ctx context.Context, workflowID string) error {
	e.runningMu.Lock()
	if _, exists := e.running[workflowID]; exists {
		e.runningMu.Unlock()
		return fmt.Errorf("workflow already running")
	}

	var workflow *Workflow
	var err error

	if e.repository != nil {
		workflow, err = e.repository.GetByID(workflowID)
		if err != nil {
			e.runningMu.Unlock()
			return fmt.Errorf("failed to get workflow: %w", err)
		}
	}

	if workflow == nil {
		e.runningMu.Unlock()
		return fmt.Errorf("workflow not found")
	}

	if !workflow.CanStart() {
		e.runningMu.Unlock()
		return fmt.Errorf("workflow cannot be started in status: %s", workflow.Status)
	}

	// Create running context
	runCtx, cancel := context.WithCancel(ctx)
	rw := &runningWorkflow{
		workflow: workflow,
		ctx:      runCtx,
		cancel:   cancel,
		events:   make(chan *WorkflowEvent, 100),
		pauseCh:  make(chan struct{}),
	}

	e.running[workflowID] = rw
	e.runningMu.Unlock()

	// Update status
	now := time.Now()
	workflow.Status = StatusRunning
	workflow.StartedAt = &now
	workflow.UpdatedAt = now

	if e.repository != nil {
		if err := e.repository.Update(workflow); err != nil {
			return fmt.Errorf("failed to update workflow: %w", err)
		}
	}

	// Emit started event
	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflowID,
		Type:       WorkflowEventStarted,
		Timestamp:  now,
	})

	// Run workflow in background
	go e.runWorkflow(rw)

	return nil
}

// PauseWorkflow pauses a running workflow
func (e *Engine) PauseWorkflow(ctx context.Context, workflowID string) error {
	e.runningMu.RLock()
	rw, exists := e.running[workflowID]
	e.runningMu.RUnlock()

	if !exists {
		return fmt.Errorf("workflow not running")
	}

	rw.pausedMu.Lock()
	if rw.paused {
		rw.pausedMu.Unlock()
		return fmt.Errorf("workflow already paused")
	}
	rw.paused = true
	rw.pausedMu.Unlock()

	// Update status
	rw.workflow.Status = StatusPaused
	rw.workflow.UpdatedAt = time.Now()

	if e.repository != nil {
		if err := e.repository.Update(rw.workflow); err != nil {
			return fmt.Errorf("failed to update workflow: %w", err)
		}
	}

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflowID,
		Type:       WorkflowEventPaused,
		Timestamp:  time.Now(),
	})

	return nil
}

// ResumeWorkflow resumes a paused workflow
func (e *Engine) ResumeWorkflow(ctx context.Context, workflowID string) error {
	e.runningMu.RLock()
	rw, exists := e.running[workflowID]
	e.runningMu.RUnlock()

	if !exists {
		// Try to load and restart from persisted state
		if e.repository != nil {
			workflow, err := e.repository.GetByID(workflowID)
			if err != nil {
				return fmt.Errorf("failed to get workflow: %w", err)
			}
			if workflow != nil && workflow.CanResume() {
				return e.StartWorkflow(ctx, workflowID)
			}
		}
		return fmt.Errorf("workflow not found or not resumable")
	}

	rw.pausedMu.Lock()
	if !rw.paused {
		rw.pausedMu.Unlock()
		return fmt.Errorf("workflow not paused")
	}
	rw.paused = false
	rw.pausedMu.Unlock()

	// Signal resume
	select {
	case rw.pauseCh <- struct{}{}:
	default:
	}

	// Update status
	rw.workflow.Status = StatusRunning
	rw.workflow.UpdatedAt = time.Now()

	if e.repository != nil {
		if err := e.repository.Update(rw.workflow); err != nil {
			return fmt.Errorf("failed to update workflow: %w", err)
		}
	}

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflowID,
		Type:       WorkflowEventResumed,
		Timestamp:  time.Now(),
	})

	return nil
}

// CancelWorkflow cancels a workflow
func (e *Engine) CancelWorkflow(ctx context.Context, workflowID string) error {
	e.runningMu.Lock()
	rw, exists := e.running[workflowID]
	if exists {
		rw.cancel()
		delete(e.running, workflowID)
	}
	e.runningMu.Unlock()

	// Update status in repository
	if e.repository != nil {
		workflow, err := e.repository.GetByID(workflowID)
		if err != nil {
			return fmt.Errorf("failed to get workflow: %w", err)
		}
		if workflow != nil {
			now := time.Now()
			workflow.Status = StatusCancelled
			workflow.CompletedAt = &now
			workflow.UpdatedAt = now
			if err := e.repository.Update(workflow); err != nil {
				return fmt.Errorf("failed to update workflow: %w", err)
			}
		}
	}

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflowID,
		Type:       WorkflowEventCancelled,
		Timestamp:  time.Now(),
	})

	return nil
}

// GetWorkflow returns a workflow by ID
func (e *Engine) GetWorkflow(workflowID string) (*Workflow, error) {
	// Check running workflows first
	e.runningMu.RLock()
	if rw, exists := e.running[workflowID]; exists {
		e.runningMu.RUnlock()
		return rw.workflow.Clone(), nil
	}
	e.runningMu.RUnlock()

	// Check repository
	if e.repository != nil {
		return e.repository.GetByID(workflowID)
	}

	return nil, fmt.Errorf("workflow not found")
}

// ListWorkflows returns workflows matching the filter
func (e *Engine) ListWorkflows(filter *WorkflowFilter) ([]*Workflow, error) {
	if e.repository == nil {
		return nil, nil
	}
	return e.repository.List(filter)
}

// Subscribe returns a channel for receiving workflow events
func (e *Engine) Subscribe() chan *WorkflowEvent {
	ch := make(chan *WorkflowEvent, 100)
	e.subscribersMu.Lock()
	e.subscribers = append(e.subscribers, ch)
	e.subscribersMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber
func (e *Engine) Unsubscribe(ch chan *WorkflowEvent) {
	e.subscribersMu.Lock()
	defer e.subscribersMu.Unlock()

	for i, sub := range e.subscribers {
		if sub == ch {
			e.subscribers = append(e.subscribers[:i], e.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

// runWorkflow executes the workflow steps
func (e *Engine) runWorkflow(rw *runningWorkflow) {
	workflow := rw.workflow
	defer func() {
		// Cleanup
		e.runningMu.Lock()
		delete(e.running, workflow.ID)
		e.runningMu.Unlock()
		close(rw.events)

		// Recover from panics
		if r := recover(); r != nil {
			workflow.Status = StatusFailed
			workflow.Error = fmt.Sprintf("panic: %v", r)
			now := time.Now()
			workflow.CompletedAt = &now
			workflow.UpdatedAt = now

			if e.repository != nil {
				e.repository.Update(workflow)
			}

			e.emitEvent(&WorkflowEvent{
				WorkflowID: workflow.ID,
				Type:       WorkflowEventFailed,
				Data:       map[string]interface{}{"error": workflow.Error},
				Timestamp:  time.Now(),
			})
		}
	}()

	for workflow.CurrentStep < len(workflow.Steps) {
		// Check for cancellation
		select {
		case <-rw.ctx.Done():
			return
		default:
		}

		// Check for pause
		rw.pausedMu.RLock()
		isPaused := rw.paused
		rw.pausedMu.RUnlock()

		if isPaused {
			// Wait for resume
			select {
			case <-rw.pauseCh:
				// Resumed
			case <-rw.ctx.Done():
				return
			}
		}

		step := &workflow.Steps[workflow.CurrentStep]

		// Check step condition
		if step.Condition != nil {
			shouldExecute, err := e.evaluateCondition(step.Condition, workflow.State)
			if err != nil {
				e.failWorkflow(rw, fmt.Sprintf("condition evaluation failed: %v", err))
				return
			}
			if !shouldExecute {
				e.emitEvent(&WorkflowEvent{
					WorkflowID: workflow.ID,
					Type:       WorkflowEventStepSkipped,
					StepID:     step.ID,
					StepName:   step.Name,
					Timestamp:  time.Now(),
				})
				workflow.CurrentStep++
				continue
			}
		}

		// Execute step
		result, err := e.executeStep(rw.ctx, workflow, step)
		if err != nil {
			// Handle retry
			if step.RetryPolicy != nil && result != nil && result.RetryCount < step.RetryPolicy.MaxRetries {
				result.RetryCount++
				e.emitEvent(&WorkflowEvent{
					WorkflowID: workflow.ID,
					Type:       WorkflowEventStepRetrying,
					StepID:     step.ID,
					StepName:   step.Name,
					Data: map[string]interface{}{
						"retry_count": result.RetryCount,
						"error":       err.Error(),
					},
					Timestamp: time.Now(),
				})

				// Wait before retry
				delay := step.RetryPolicy.Delay
				if step.RetryPolicy.BackoffType == "exponential" {
					delay = delay * time.Duration(1<<result.RetryCount)
					if step.RetryPolicy.MaxDelay > 0 && delay > step.RetryPolicy.MaxDelay {
						delay = step.RetryPolicy.MaxDelay
					}
				}

				select {
				case <-time.After(delay):
				case <-rw.ctx.Done():
					return
				}
				continue
			}

			// Check for failure handler
			if step.OnFailure != "" {
				nextIdx := workflow.GetStepIndex(step.OnFailure)
				if nextIdx >= 0 {
					workflow.CurrentStep = nextIdx
					continue
				}
			}

			// Fail workflow
			e.failWorkflow(rw, fmt.Sprintf("step %s failed: %v", step.Name, err))
			return
		}

		// Store result in state if output key specified
		if result != nil && result.Output != nil {
			var outputKey string
			switch step.Type {
			case StepTypeAgent:
				if step.Config.AgentConfig != nil {
					outputKey = step.Config.AgentConfig.OutputKey
				}
			case StepTypeTool:
				if step.Config.ToolConfig != nil {
					outputKey = step.Config.ToolConfig.OutputKey
				}
			case StepTypeWait:
				if step.Config.WaitConfig != nil {
					outputKey = step.Config.WaitConfig.OutputKey
				}
			case StepTypeTransform:
				if step.Config.TransformConfig != nil {
					outputKey = step.Config.TransformConfig.OutputKey
				}
			}

			if outputKey != "" {
				workflow.SetStateValue(outputKey, result.Output)
			}

			// Always store in step-specific key
			workflow.SetStateValue(fmt.Sprintf("step_%s_output", step.ID), result.Output)
		}

		// Persist state
		if e.repository != nil {
			if err := e.repository.UpdateState(workflow.ID, workflow.State, workflow.CurrentStep, workflow.Status); err != nil {
				e.failWorkflow(rw, fmt.Sprintf("failed to persist state: %v", err))
				return
			}
		}

		// Determine next step
		if step.OnSuccess != "" {
			nextIdx := workflow.GetStepIndex(step.OnSuccess)
			if nextIdx >= 0 {
				workflow.CurrentStep = nextIdx
			} else {
				workflow.CurrentStep++
			}
		} else {
			workflow.CurrentStep++
		}
	}

	// Workflow completed successfully
	now := time.Now()
	workflow.Status = StatusCompleted
	workflow.CompletedAt = &now
	workflow.UpdatedAt = now

	if e.repository != nil {
		e.repository.Update(workflow)
	}

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflow.ID,
		Type:       WorkflowEventCompleted,
		Data: map[string]interface{}{
			"state": workflow.State,
		},
		Timestamp: now,
	})
}

// executeStep executes a single workflow step
func (e *Engine) executeStep(ctx context.Context, workflow *Workflow, step *Step) (*StepResult, error) {
	executor, ok := e.executors[step.Type]
	if !ok {
		return nil, fmt.Errorf("no executor for step type: %s", step.Type)
	}

	// Apply timeout if specified
	execCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Emit step started event
	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflow.ID,
		Type:       WorkflowEventStepStarted,
		StepID:     step.ID,
		StepName:   step.Name,
		Data: map[string]interface{}{
			"step_type": step.Type,
		},
		Timestamp: time.Now(),
	})

	startTime := time.Now()
	result, err := executor.Execute(execCtx, step, workflow.State)

	if result == nil {
		result = &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
		}
	}
	result.StartedAt = startTime
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(startTime)

	if err != nil {
		result.Status = StepStatusFailed
		result.Error = err.Error()

		e.emitEvent(&WorkflowEvent{
			WorkflowID: workflow.ID,
			Type:       WorkflowEventStepFailed,
			StepID:     step.ID,
			StepName:   step.Name,
			Data: map[string]interface{}{
				"error":    err.Error(),
				"duration": result.Duration.Milliseconds(),
			},
			Timestamp: time.Now(),
		})

		return result, err
	}

	result.Status = StepStatusCompleted

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflow.ID,
		Type:       WorkflowEventStepCompleted,
		StepID:     step.ID,
		StepName:   step.Name,
		Data: map[string]interface{}{
			"output":   result.Output,
			"duration": result.Duration.Milliseconds(),
		},
		Timestamp: time.Now(),
	})

	return result, nil
}

// evaluateCondition evaluates a step condition
func (e *Engine) evaluateCondition(cond *Condition, state map[string]interface{}) (bool, error) {
	if cond == nil {
		return true, nil
	}

	switch cond.Type {
	case "state_check":
		val, exists := state[cond.StateKey]
		switch cond.Operator {
		case "exists":
			return exists, nil
		case "equals":
			return fmt.Sprintf("%v", val) == cond.Value, nil
		case "not_equals":
			return fmt.Sprintf("%v", val) != cond.Value, nil
		case "contains":
			return strings.Contains(fmt.Sprintf("%v", val), cond.Value), nil
		default:
			return false, fmt.Errorf("unknown operator: %s", cond.Operator)
		}

	case "expression":
		// Simple expression evaluation
		// For now, support basic checks like "state.key == value"
		return e.evaluateExpression(cond.Expression, state)

	default:
		return false, fmt.Errorf("unknown condition type: %s", cond.Type)
	}
}

// evaluateExpression evaluates a simple expression
func (e *Engine) evaluateExpression(expr string, state map[string]interface{}) (bool, error) {
	// Replace state references
	expr = e.interpolateState(expr, state)

	// Basic true/false check
	expr = strings.TrimSpace(strings.ToLower(expr))
	return expr == "true" || expr == "1", nil
}

// interpolateState replaces {{state.key}} placeholders with actual values
func (e *Engine) interpolateState(s string, state map[string]interface{}) string {
	re := regexp.MustCompile(`\{\{state\.([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := re.FindStringSubmatch(match)[1]
		if val, ok := state[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

// failWorkflow marks a workflow as failed
func (e *Engine) failWorkflow(rw *runningWorkflow, errMsg string) {
	workflow := rw.workflow
	now := time.Now()
	workflow.Status = StatusFailed
	workflow.Error = errMsg
	workflow.CompletedAt = &now
	workflow.UpdatedAt = now

	if e.repository != nil {
		e.repository.Update(workflow)
	}

	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflow.ID,
		Type:       WorkflowEventFailed,
		Data:       map[string]interface{}{"error": errMsg},
		Timestamp:  now,
	})
}

// emitEvent sends an event to all subscribers
func (e *Engine) emitEvent(event *WorkflowEvent) {
	e.subscribersMu.RLock()
	defer e.subscribersMu.RUnlock()

	for _, ch := range e.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// ProvideInput provides external input to a waiting workflow step
func (e *Engine) ProvideInput(workflowID, stepID string, input interface{}) error {
	e.runningMu.RLock()
	rw, exists := e.running[workflowID]
	e.runningMu.RUnlock()

	if !exists {
		return fmt.Errorf("workflow not running")
	}

	// Store input in state
	rw.workflow.SetStateValue(fmt.Sprintf("input_%s", stepID), input)

	// Signal the wait executor if needed
	e.emitEvent(&WorkflowEvent{
		WorkflowID: workflowID,
		Type:       WorkflowEventStateUpdated,
		StepID:     stepID,
		Data: map[string]interface{}{
			"input": input,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// GetState returns the current state of a workflow
func (e *Engine) GetState(workflowID string) (map[string]interface{}, error) {
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found")
	}
	return workflow.State, nil
}

// InterpolateString replaces state placeholders in a string
func (e *Engine) InterpolateString(s string, state map[string]interface{}) string {
	return e.interpolateState(s, state)
}
