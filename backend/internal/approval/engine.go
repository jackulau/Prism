package approval

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for approval persistence
type Repository interface {
	CreateWorkflow(w *ApprovalWorkflow) error
	GetWorkflowByID(id string) (*ApprovalWorkflow, error)
	UpdateWorkflow(w *ApprovalWorkflow) error
	DeleteWorkflow(id string) error
	ListWorkflows(filter *WorkflowFilter) ([]*ApprovalWorkflow, error)
	GetActiveWorkflowForOperation(orgID string, opType OperationType) (*ApprovalWorkflow, error)

	CreateRequest(req *ApprovalRequest) error
	GetRequestByID(id string) (*ApprovalRequest, error)
	UpdateRequest(req *ApprovalRequest) error
	ListRequests(filter *ApprovalRequestFilter) ([]*ApprovalRequest, error)
	GetPendingRequestsForApprover(orgID, userID string, userRoles []string) ([]*ApprovalRequest, error)
	GetExpiredRequests() ([]*ApprovalRequest, error)

	CreateDecision(d *ApprovalDecision) error
	GetDecisionsForRequest(requestID string) ([]*ApprovalDecision, error)
	GetDecisionsForStep(requestID string, stepOrder int) ([]*ApprovalDecision, error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	NotifyApprovers(ctx context.Context, request *ApprovalRequest, workflow *ApprovalWorkflow, step *ApprovalStep) error
	NotifyRequester(ctx context.Context, request *ApprovalRequest, decision ApprovalStatus, comment string) error
	SendReminder(ctx context.Context, request *ApprovalRequest, step *ApprovalStep) error
}

// EventType represents types of approval events
type EventType string

const (
	EventRequestCreated    EventType = "request_created"
	EventRequestApproved   EventType = "request_approved"
	EventRequestRejected   EventType = "request_rejected"
	EventRequestEscalated  EventType = "request_escalated"
	EventRequestExpired    EventType = "request_expired"
	EventRequestCancelled  EventType = "request_cancelled"
	EventStepApproved      EventType = "step_approved"
	EventStepRejected      EventType = "step_rejected"
	EventDecisionRecorded  EventType = "decision_recorded"
)

// Event represents an approval system event
type Event struct {
	Type       EventType              `json:"type"`
	RequestID  string                 `json:"request_id"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	StepOrder  int                    `json:"step_order,omitempty"`
	ApproverID string                 `json:"approver_id,omitempty"`
	Decision   ApprovalStatus         `json:"decision,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Engine processes approval requests through workflow steps
type Engine struct {
	repo         Repository
	notifier     NotificationService
	subscribers  map[chan *Event]struct{}
	mu           sync.RWMutex
	stopChan     chan struct{}
	checkTicker  *time.Ticker
}

// NewEngine creates a new approval engine
func NewEngine(repo Repository, notifier NotificationService) *Engine {
	return &Engine{
		repo:        repo,
		notifier:    notifier,
		subscribers: make(map[chan *Event]struct{}),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the approval engine background processes
func (e *Engine) Start(ctx context.Context) {
	// Start expiration checker (runs every minute)
	e.checkTicker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-e.checkTicker.C:
				e.checkExpiredRequests(ctx)
			case <-e.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("Approval engine started")
}

// Stop stops the approval engine
func (e *Engine) Stop() {
	if e.checkTicker != nil {
		e.checkTicker.Stop()
	}
	close(e.stopChan)
	log.Println("Approval engine stopped")
}

// Subscribe returns a channel that receives approval events
func (e *Engine) Subscribe() chan *Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := make(chan *Event, 100)
	e.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber
func (e *Engine) Unsubscribe(ch chan *Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.subscribers, ch)
	close(ch)
}

// emit sends an event to all subscribers
func (e *Engine) emit(event *Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for ch := range e.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// CreateRequest creates a new approval request
func (e *Engine) CreateRequest(ctx context.Context, orgID, requesterID, requesterEmail string, opType OperationType, details map[string]interface{}) (*ApprovalRequest, error) {
	// Find active workflow for this operation type
	workflow, err := e.repo.GetActiveWorkflowForOperation(orgID, opType)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("no active approval workflow found for operation type: %s", opType)
	}

	// Check if conditions match (if specified)
	if workflow.Conditions != nil {
		if !e.matchesConditions(workflow.Conditions, details) {
			return nil, fmt.Errorf("operation does not match workflow conditions")
		}
	}

	// Calculate expiration based on first step timeout
	var expiresAt *time.Time
	if len(workflow.Steps) > 0 && workflow.Steps[0].Timeout > 0 {
		exp := time.Now().Add(workflow.Steps[0].Timeout)
		expiresAt = &exp
	}

	request := &ApprovalRequest{
		ID:               uuid.New().String(),
		WorkflowID:       workflow.ID,
		OrganizationID:   orgID,
		RequesterID:      requesterID,
		RequesterEmail:   requesterEmail,
		OperationType:    opType,
		OperationDetails: details,
		CurrentStep:      0,
		TotalSteps:       len(workflow.Steps),
		Status:           StatusPending,
		Priority:         0,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := e.repo.CreateRequest(request); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Check auto-approve for first step
	if len(workflow.Steps) > 0 {
		step := &workflow.Steps[0]
		if step.AutoApprove != nil && step.AutoApprove.Enabled {
			if e.checkAutoApprove(step.AutoApprove, details) {
				// Auto-approve this step
				if err := e.autoApproveStep(ctx, request, workflow, step); err != nil {
					log.Printf("Auto-approve failed for request %s: %v", request.ID, err)
				}
			}
		}
	}

	// Notify approvers if still pending
	if request.Status == StatusPending && e.notifier != nil {
		step := workflow.GetStep(request.CurrentStep)
		if step != nil {
			if err := e.notifier.NotifyApprovers(ctx, request, workflow, step); err != nil {
				log.Printf("Failed to notify approvers for request %s: %v", request.ID, err)
			}
		}
	}

	e.emit(&Event{
		Type:       EventRequestCreated,
		RequestID:  request.ID,
		WorkflowID: workflow.ID,
		Data: map[string]interface{}{
			"operation_type": opType,
			"requester_id":   requesterID,
		},
		Timestamp: time.Now(),
	})

	return request, nil
}

// RecordDecision records an approver's decision
func (e *Engine) RecordDecision(ctx context.Context, requestID, approverID, approverEmail string, decision ApprovalStatus, comment string) error {
	request, err := e.repo.GetRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}
	if request == nil {
		return fmt.Errorf("request not found: %s", requestID)
	}

	if !request.CanDecide() {
		return fmt.Errorf("request is not pending, current status: %s", request.Status)
	}

	if request.IsExpired() {
		// Mark as expired
		request.Status = StatusExpired
		now := time.Now()
		request.CompletedAt = &now
		request.UpdatedAt = now
		if err := e.repo.UpdateRequest(request); err != nil {
			return fmt.Errorf("failed to update expired request: %w", err)
		}
		return fmt.Errorf("request has expired")
	}

	workflow, err := e.repo.GetWorkflowByID(request.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}
	if workflow == nil {
		return fmt.Errorf("workflow not found: %s", request.WorkflowID)
	}

	step := workflow.GetStep(request.CurrentStep)
	if step == nil {
		return fmt.Errorf("step not found: %d", request.CurrentStep)
	}

	// Validate decision
	if decision != StatusApproved && decision != StatusRejected {
		return fmt.Errorf("invalid decision: %s, must be approved or rejected", decision)
	}

	// Record the decision
	decisionRecord := &ApprovalDecision{
		ID:            uuid.New().String(),
		RequestID:     requestID,
		StepOrder:     request.CurrentStep,
		ApproverID:    approverID,
		ApproverEmail: approverEmail,
		Decision:      decision,
		Comment:       comment,
		CreatedAt:     time.Now(),
	}

	if err := e.repo.CreateDecision(decisionRecord); err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}

	e.emit(&Event{
		Type:       EventDecisionRecorded,
		RequestID:  requestID,
		WorkflowID: workflow.ID,
		StepOrder:  request.CurrentStep,
		ApproverID: approverID,
		Decision:   decision,
		Data: map[string]interface{}{
			"comment": comment,
		},
		Timestamp: time.Now(),
	})

	// Process step completion
	return e.processStepCompletion(ctx, request, workflow, step)
}

// processStepCompletion checks if a step is complete and advances the workflow
func (e *Engine) processStepCompletion(ctx context.Context, request *ApprovalRequest, workflow *ApprovalWorkflow, step *ApprovalStep) error {
	decisions, err := e.repo.GetDecisionsForStep(request.ID, request.CurrentStep)
	if err != nil {
		return fmt.Errorf("failed to get decisions: %w", err)
	}

	approvedCount := 0
	rejectedCount := 0
	for _, d := range decisions {
		if d.Decision == StatusApproved {
			approvedCount++
		} else if d.Decision == StatusRejected {
			rejectedCount++
		}
	}

	// Determine required approvals based on requirement type
	requiredApprovals := len(step.ApproverRoles) + len(step.ApproverUserIDs)
	if step.Requirement == RequirementAny {
		requiredApprovals = 1
	}

	// Check for rejection first
	if rejectedCount > 0 {
		// Single rejection fails the step (and request)
		request.Status = StatusRejected
		now := time.Now()
		request.CompletedAt = &now
		request.UpdatedAt = now

		if err := e.repo.UpdateRequest(request); err != nil {
			return fmt.Errorf("failed to update request: %w", err)
		}

		e.emit(&Event{
			Type:       EventStepRejected,
			RequestID:  request.ID,
			WorkflowID: workflow.ID,
			StepOrder:  request.CurrentStep,
			Decision:   StatusRejected,
			Timestamp:  time.Now(),
		})

		e.emit(&Event{
			Type:       EventRequestRejected,
			RequestID:  request.ID,
			WorkflowID: workflow.ID,
			Decision:   StatusRejected,
			Timestamp:  time.Now(),
		})

		// Notify requester
		if e.notifier != nil {
			if err := e.notifier.NotifyRequester(ctx, request, StatusRejected, ""); err != nil {
				log.Printf("Failed to notify requester: %v", err)
			}
		}

		return nil
	}

	// Check if step is approved
	if approvedCount >= requiredApprovals {
		e.emit(&Event{
			Type:       EventStepApproved,
			RequestID:  request.ID,
			WorkflowID: workflow.ID,
			StepOrder:  request.CurrentStep,
			Decision:   StatusApproved,
			Timestamp:  time.Now(),
		})

		// Advance to next step or complete
		if request.CurrentStep >= len(workflow.Steps)-1 {
			// All steps complete
			request.Status = StatusApproved
			now := time.Now()
			request.CompletedAt = &now
			request.UpdatedAt = now

			if err := e.repo.UpdateRequest(request); err != nil {
				return fmt.Errorf("failed to update request: %w", err)
			}

			e.emit(&Event{
				Type:       EventRequestApproved,
				RequestID:  request.ID,
				WorkflowID: workflow.ID,
				Decision:   StatusApproved,
				Timestamp:  time.Now(),
			})

			// Notify requester
			if e.notifier != nil {
				if err := e.notifier.NotifyRequester(ctx, request, StatusApproved, ""); err != nil {
					log.Printf("Failed to notify requester: %v", err)
				}
			}
		} else {
			// Advance to next step
			request.CurrentStep++
			request.UpdatedAt = time.Now()

			// Update expiration based on new step timeout
			nextStep := workflow.GetStep(request.CurrentStep)
			if nextStep != nil && nextStep.Timeout > 0 {
				exp := time.Now().Add(nextStep.Timeout)
				request.ExpiresAt = &exp
			}

			if err := e.repo.UpdateRequest(request); err != nil {
				return fmt.Errorf("failed to update request: %w", err)
			}

			// Notify approvers for next step
			if e.notifier != nil && nextStep != nil {
				if err := e.notifier.NotifyApprovers(ctx, request, workflow, nextStep); err != nil {
					log.Printf("Failed to notify approvers: %v", err)
				}
			}
		}
	}

	return nil
}

// autoApproveStep automatically approves a step
func (e *Engine) autoApproveStep(ctx context.Context, request *ApprovalRequest, workflow *ApprovalWorkflow, step *ApprovalStep) error {
	decision := &ApprovalDecision{
		ID:            uuid.New().String(),
		RequestID:     request.ID,
		StepOrder:     request.CurrentStep,
		ApproverID:    "system",
		ApproverEmail: "system@auto-approve",
		Decision:      StatusApproved,
		Comment:       "Auto-approved by system",
		CreatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"auto_approved": true,
		},
	}

	if err := e.repo.CreateDecision(decision); err != nil {
		return fmt.Errorf("failed to create auto-approve decision: %w", err)
	}

	return e.processStepCompletion(ctx, request, workflow, step)
}

// CancelRequest cancels an approval request
func (e *Engine) CancelRequest(ctx context.Context, requestID, cancellerID string) error {
	request, err := e.repo.GetRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}
	if request == nil {
		return fmt.Errorf("request not found: %s", requestID)
	}

	if request.Status.IsTerminal() {
		return fmt.Errorf("request is already in terminal state: %s", request.Status)
	}

	request.Status = StatusCancelled
	now := time.Now()
	request.CompletedAt = &now
	request.UpdatedAt = now

	if err := e.repo.UpdateRequest(request); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	e.emit(&Event{
		Type:      EventRequestCancelled,
		RequestID: requestID,
		Data: map[string]interface{}{
			"cancelled_by": cancellerID,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// EscalateRequest escalates a request to the next level
func (e *Engine) EscalateRequest(ctx context.Context, requestID string) error {
	request, err := e.repo.GetRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}
	if request == nil {
		return fmt.Errorf("request not found: %s", requestID)
	}

	if !request.CanDecide() {
		return fmt.Errorf("request cannot be escalated, status: %s", request.Status)
	}

	workflow, err := e.repo.GetWorkflowByID(request.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	step := workflow.GetStep(request.CurrentStep)
	if step == nil || step.EscalationConfig == nil || !step.EscalationConfig.Enabled {
		return fmt.Errorf("escalation not configured for current step")
	}

	request.Status = StatusEscalated
	request.UpdatedAt = time.Now()

	// Increase priority on escalation
	request.Priority++

	// Extend expiration
	if step.EscalationConfig.EscalateAfter > 0 {
		exp := time.Now().Add(step.EscalationConfig.EscalateAfter)
		request.ExpiresAt = &exp
	}

	if err := e.repo.UpdateRequest(request); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	e.emit(&Event{
		Type:      EventRequestEscalated,
		RequestID: requestID,
		StepOrder: request.CurrentStep,
		Data: map[string]interface{}{
			"escalate_to_roles": step.EscalationConfig.EscalateToRoles,
			"escalate_to_users": step.EscalationConfig.EscalateToUsers,
		},
		Timestamp: time.Now(),
	})

	// TODO: Notify escalation targets

	return nil
}

// GetRequest returns a request by ID
func (e *Engine) GetRequest(requestID string) (*ApprovalRequest, error) {
	return e.repo.GetRequestByID(requestID)
}

// GetWorkflow returns a workflow by ID
func (e *Engine) GetWorkflow(workflowID string) (*ApprovalWorkflow, error) {
	return e.repo.GetWorkflowByID(workflowID)
}

// GetStepProgress returns the progress of the current step
func (e *Engine) GetStepProgress(requestID string) (*StepProgress, error) {
	request, err := e.repo.GetRequestByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %w", err)
	}
	if request == nil {
		return nil, fmt.Errorf("request not found: %s", requestID)
	}

	workflow, err := e.repo.GetWorkflowByID(request.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	step := workflow.GetStep(request.CurrentStep)
	if step == nil {
		return nil, fmt.Errorf("step not found: %d", request.CurrentStep)
	}

	decisions, err := e.repo.GetDecisionsForStep(requestID, request.CurrentStep)
	if err != nil {
		return nil, fmt.Errorf("failed to get decisions: %w", err)
	}

	approvedCount := 0
	rejectedCount := 0
	decidedUsers := make(map[string]bool)

	approvalDecisions := make([]ApprovalDecision, 0, len(decisions))
	for _, d := range decisions {
		if d.Decision == StatusApproved {
			approvedCount++
		} else if d.Decision == StatusRejected {
			rejectedCount++
		}
		decidedUsers[d.ApproverID] = true
		approvalDecisions = append(approvalDecisions, *d)
	}

	// Calculate required count
	requiredCount := len(step.ApproverRoles) + len(step.ApproverUserIDs)
	if step.Requirement == RequirementAny {
		requiredCount = 1
	}

	// Find pending approvers
	var pendingApprovers []string
	for _, userID := range step.ApproverUserIDs {
		if !decidedUsers[userID] {
			pendingApprovers = append(pendingApprovers, userID)
		}
	}

	progress := &StepProgress{
		RequestID:        requestID,
		StepOrder:        request.CurrentStep,
		StepName:         step.Name,
		Requirement:      step.Requirement,
		RequiredCount:    requiredCount,
		ApprovedCount:    approvedCount,
		RejectedCount:    rejectedCount,
		PendingApprovers: pendingApprovers,
		Decisions:        approvalDecisions,
		Status:           request.Status,
		StartedAt:        request.CreatedAt,
	}

	return progress, nil
}

// checkExpiredRequests processes expired requests
func (e *Engine) checkExpiredRequests(ctx context.Context) {
	expired, err := e.repo.GetExpiredRequests()
	if err != nil {
		log.Printf("Failed to get expired requests: %v", err)
		return
	}

	for _, request := range expired {
		request.Status = StatusExpired
		now := time.Now()
		request.CompletedAt = &now
		request.UpdatedAt = now

		if err := e.repo.UpdateRequest(request); err != nil {
			log.Printf("Failed to update expired request %s: %v", request.ID, err)
			continue
		}

		e.emit(&Event{
			Type:      EventRequestExpired,
			RequestID: request.ID,
			Timestamp: time.Now(),
		})

		// Notify requester
		if e.notifier != nil {
			if err := e.notifier.NotifyRequester(ctx, request, StatusExpired, "Request expired"); err != nil {
				log.Printf("Failed to notify requester: %v", err)
			}
		}
	}
}

// matchesConditions checks if operation details match workflow conditions
func (e *Engine) matchesConditions(conditions *WorkflowConditions, details map[string]interface{}) bool {
	// Check tool names
	if len(conditions.ToolNames) > 0 {
		toolName, ok := details["tool_name"].(string)
		if !ok {
			return false
		}
		found := false
		for _, name := range conditions.ToolNames {
			if name == toolName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check agent types
	if len(conditions.AgentTypes) > 0 {
		agentType, ok := details["agent_type"].(string)
		if !ok {
			return false
		}
		found := false
		for _, t := range conditions.AgentTypes {
			if t == agentType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check config keys
	if len(conditions.ConfigKeys) > 0 {
		configKey, ok := details["config_key"].(string)
		if !ok {
			return false
		}
		found := false
		for _, key := range conditions.ConfigKeys {
			if key == configKey {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check minimum risk score
	if conditions.MinRiskScore > 0 {
		riskScore, ok := details["risk_score"].(int)
		if !ok {
			return false
		}
		if riskScore < conditions.MinRiskScore {
			return false
		}
	}

	return true
}

// checkAutoApprove checks if auto-approve conditions are met
func (e *Engine) checkAutoApprove(config *AutoApproveCondition, details map[string]interface{}) bool {
	if config == nil || !config.Enabled {
		return false
	}

	// Check condition key-value matches
	if len(config.Conditions) > 0 {
		for key, expectedValue := range config.Conditions {
			actualValue, ok := details[key]
			if !ok || actualValue != expectedValue {
				return false
			}
		}
	}

	// Expression evaluation could be added here

	return true
}
