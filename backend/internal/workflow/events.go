package workflow

import (
	"sync"
	"time"
)

// EventBus manages workflow event subscriptions and broadcasting
type EventBus struct {
	subscribers map[string][]chan WorkflowEvent
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan WorkflowEvent),
	}
}

// Subscribe creates a subscription for events related to an execution
// Returns a channel that receives events and an unsubscribe function
func (b *EventBus) Subscribe(executionID string) (<-chan WorkflowEvent, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan WorkflowEvent, 100)
	b.subscribers[executionID] = append(b.subscribers[executionID], ch)

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs := b.subscribers[executionID]
		for i, sub := range subs {
			if sub == ch {
				b.subscribers[executionID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}

		// Clean up empty subscription lists
		if len(b.subscribers[executionID]) == 0 {
			delete(b.subscribers, executionID)
		}
	}

	return ch, unsubscribe
}

// SubscribeAll creates a subscription for all events
// Returns a channel that receives all events and an unsubscribe function
func (b *EventBus) SubscribeAll() (<-chan WorkflowEvent, func()) {
	return b.Subscribe("*")
}

// Publish sends an event to all subscribers
func (b *EventBus) Publish(event WorkflowEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Send to execution-specific subscribers
	if subs, ok := b.subscribers[event.ExecutionID]; ok {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}

	// Send to wildcard subscribers
	if subs, ok := b.subscribers["*"]; ok {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

// Close closes all subscriptions
func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, subs := range b.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	b.subscribers = make(map[string][]chan WorkflowEvent)
}

// EventEmitter provides a convenient way to emit events during workflow execution
type EventEmitter struct {
	bus         *EventBus
	executionID string
	agentID     string
	handlers    []WorkflowEventHandler
}

// NewEventEmitter creates a new event emitter for a specific execution
func NewEventEmitter(bus *EventBus, executionID, agentID string) *EventEmitter {
	return &EventEmitter{
		bus:         bus,
		executionID: executionID,
		agentID:     agentID,
		handlers:    make([]WorkflowEventHandler, 0),
	}
}

// AddHandler adds an event handler that will be called for all events
func (e *EventEmitter) AddHandler(handler WorkflowEventHandler) {
	e.handlers = append(e.handlers, handler)
}

// EmitStepStarted emits an event indicating a step has started
func (e *EventEmitter) EmitStepStarted(step WorkflowStep, data map[string]interface{}) {
	e.emit(step, "started", data, "")
}

// EmitStepCompleted emits an event indicating a step has completed
func (e *EventEmitter) EmitStepCompleted(step WorkflowStep, data map[string]interface{}) {
	e.emit(step, "completed", data, "")
}

// EmitStepFailed emits an event indicating a step has failed
func (e *EventEmitter) EmitStepFailed(step WorkflowStep, err error, data map[string]interface{}) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	e.emit(step, "failed", data, errMsg)
}

// EmitProgress emits a progress event for long-running steps
func (e *EventEmitter) EmitProgress(step WorkflowStep, progress float64, message string) {
	e.emit(step, "progress", map[string]interface{}{
		"progress": progress,
		"message":  message,
	}, "")
}

// EmitLLMChunk emits an LLM streaming chunk event
func (e *EventEmitter) EmitLLMChunk(delta string, toolCalls []ToolCallResult) {
	data := map[string]interface{}{
		"delta": delta,
	}
	if len(toolCalls) > 0 {
		data["tool_calls"] = toolCalls
	}
	e.emit(StepRunLLM, "chunk", data, "")
}

// EmitToolExecution emits a tool execution event
func (e *EventEmitter) EmitToolExecution(toolName string, status string, output string, err error) {
	data := map[string]interface{}{
		"tool_name": toolName,
		"output":    output,
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	e.emit(StepRunLLM, "tool_"+status, data, errMsg)
}

func (e *EventEmitter) emit(step WorkflowStep, status string, data map[string]interface{}, errMsg string) {
	event := WorkflowEvent{
		ExecutionID: e.executionID,
		AgentID:     e.agentID,
		Step:        step,
		Status:      status,
		Data:        data,
		Error:       errMsg,
		Timestamp:   time.Now(),
	}

	// Call registered handlers
	for _, handler := range e.handlers {
		handler(event)
	}

	// Publish to event bus if available
	if e.bus != nil {
		e.bus.Publish(event)
	}
}

// WorkflowEventType constants for common event types
const (
	EventStatusStarted   = "started"
	EventStatusCompleted = "completed"
	EventStatusFailed    = "failed"
	EventStatusProgress  = "progress"
	EventStatusChunk     = "chunk"
	EventStatusToolStart = "tool_started"
	EventStatusToolEnd   = "tool_completed"
)
