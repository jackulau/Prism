package approval

import (
	"context"
	"fmt"
	"log"

	"github.com/jacklau/prism/internal/integrations"
)

// ApprovalEventType represents types of approval-specific events for integrations
type ApprovalEventType string

const (
	ApprovalEventApprovalRequested    ApprovalEventType = "approval.requested"
	ApprovalEventApprovalApproved     ApprovalEventType = "approval.approved"
	ApprovalEventApprovalRejected     ApprovalEventType = "approval.rejected"
	ApprovalEventApprovalEscalated    ApprovalEventType = "approval.escalated"
	ApprovalEventApprovalExpired      ApprovalEventType = "approval.expired"
	ApprovalEventApprovalCancelled    ApprovalEventType = "approval.cancelled"
	ApprovalEventApprovalReminder     ApprovalEventType = "approval.reminder"
)

// Notifier handles approval notifications using the integrations manager
type Notifier struct {
	manager *integrations.Manager
}

// NewNotifier creates a new approval notifier
func NewNotifier(manager *integrations.Manager) *Notifier {
	return &Notifier{
		manager: manager,
	}
}

// NotifyApprovers notifies approvers about a pending approval request
func (n *Notifier) NotifyApprovers(ctx context.Context, request *ApprovalRequest, workflow *ApprovalWorkflow, step *ApprovalStep) error {
	if n.manager == nil {
		log.Println("Integrations manager not configured, skipping approval notifications")
		return nil
	}

	// Build approver information
	approvers := make([]string, 0, len(step.ApproverRoles)+len(step.ApproverUserIDs))
	approvers = append(approvers, step.ApproverRoles...)
	approvers = append(approvers, step.ApproverUserIDs...)

	event := &integrations.Event{
		Type:   integrations.EventType(ApprovalEventApprovalRequested),
		UserID: request.RequesterID,
		Data: map[string]interface{}{
			"request_id":        request.ID,
			"workflow_id":       workflow.ID,
			"workflow_name":     workflow.Name,
			"operation_type":    string(request.OperationType),
			"operation_details": request.OperationDetails,
			"step_name":         step.Name,
			"step_order":        step.Order,
			"total_steps":       request.TotalSteps,
			"current_step":      request.CurrentStep,
			"approvers":         approvers,
			"approver_roles":    step.ApproverRoles,
			"approver_user_ids": step.ApproverUserIDs,
			"requester_id":      request.RequesterID,
			"requester_email":   request.RequesterEmail,
			"priority":          request.Priority,
			"organization_id":   request.OrganizationID,
		},
	}

	if request.ExpiresAt != nil {
		event.Data["expires_at"] = request.ExpiresAt.Format("2006-01-02 15:04:05")
	}

	n.manager.TrackAndNotify(event)
	return nil
}

// NotifyRequester notifies the requester about a decision
func (n *Notifier) NotifyRequester(ctx context.Context, request *ApprovalRequest, decision ApprovalStatus, comment string) error {
	if n.manager == nil {
		log.Println("Integrations manager not configured, skipping approval notifications")
		return nil
	}

	var eventType ApprovalEventType
	switch decision {
	case StatusApproved:
		eventType = ApprovalEventApprovalApproved
	case StatusRejected:
		eventType = ApprovalEventApprovalRejected
	case StatusExpired:
		eventType = ApprovalEventApprovalExpired
	case StatusCancelled:
		eventType = ApprovalEventApprovalCancelled
	case StatusEscalated:
		eventType = ApprovalEventApprovalEscalated
	default:
		return fmt.Errorf("unknown decision status: %s", decision)
	}

	event := &integrations.Event{
		Type:   integrations.EventType(eventType),
		UserID: request.RequesterID,
		Data: map[string]interface{}{
			"request_id":        request.ID,
			"workflow_id":       request.WorkflowID,
			"operation_type":    string(request.OperationType),
			"operation_details": request.OperationDetails,
			"decision":          string(decision),
			"comment":           comment,
			"requester_id":      request.RequesterID,
			"requester_email":   request.RequesterEmail,
			"organization_id":   request.OrganizationID,
		},
	}

	n.manager.TrackAndNotify(event)
	return nil
}

// SendReminder sends a reminder to approvers about a pending approval
func (n *Notifier) SendReminder(ctx context.Context, request *ApprovalRequest, step *ApprovalStep) error {
	if n.manager == nil {
		log.Println("Integrations manager not configured, skipping approval reminders")
		return nil
	}

	event := &integrations.Event{
		Type:   integrations.EventType(ApprovalEventApprovalReminder),
		UserID: request.RequesterID,
		Data: map[string]interface{}{
			"request_id":        request.ID,
			"workflow_id":       request.WorkflowID,
			"operation_type":    string(request.OperationType),
			"step_name":         step.Name,
			"step_order":        step.Order,
			"approver_roles":    step.ApproverRoles,
			"approver_user_ids": step.ApproverUserIDs,
			"requester_id":      request.RequesterID,
			"requester_email":   request.RequesterEmail,
			"organization_id":   request.OrganizationID,
		},
	}

	if request.ExpiresAt != nil {
		event.Data["expires_at"] = request.ExpiresAt.Format("2006-01-02 15:04:05")
	}

	n.manager.TrackAndNotify(event)
	return nil
}

// NotifyDecision notifies about a specific decision (approve/reject)
func (n *Notifier) NotifyDecision(ctx context.Context, request *ApprovalRequest, decision *ApprovalDecision) error {
	if n.manager == nil {
		log.Println("Integrations manager not configured, skipping decision notifications")
		return nil
	}

	var eventType ApprovalEventType
	if decision.Decision == StatusApproved {
		eventType = ApprovalEventApprovalApproved
	} else {
		eventType = ApprovalEventApprovalRejected
	}

	event := &integrations.Event{
		Type:   integrations.EventType(eventType),
		UserID: decision.ApproverID,
		Data: map[string]interface{}{
			"request_id":      request.ID,
			"decision_id":     decision.ID,
			"operation_type":  string(request.OperationType),
			"decision":        string(decision.Decision),
			"comment":         decision.Comment,
			"approver_id":     decision.ApproverID,
			"approver_email":  decision.ApproverEmail,
			"step_order":      decision.StepOrder,
			"requester_id":    request.RequesterID,
			"organization_id": request.OrganizationID,
		},
	}

	n.manager.TrackAndNotify(event)
	return nil
}

// NotifyEscalation notifies about an escalated approval request
func (n *Notifier) NotifyEscalation(ctx context.Context, request *ApprovalRequest, escalateToRoles []string, escalateToUsers []string) error {
	if n.manager == nil {
		log.Println("Integrations manager not configured, skipping escalation notifications")
		return nil
	}

	event := &integrations.Event{
		Type:   integrations.EventType(ApprovalEventApprovalEscalated),
		UserID: request.RequesterID,
		Data: map[string]interface{}{
			"request_id":         request.ID,
			"workflow_id":        request.WorkflowID,
			"operation_type":     string(request.OperationType),
			"current_step":       request.CurrentStep,
			"escalate_to_roles":  escalateToRoles,
			"escalate_to_users":  escalateToUsers,
			"requester_id":       request.RequesterID,
			"requester_email":    request.RequesterEmail,
			"organization_id":    request.OrganizationID,
			"new_priority":       request.Priority,
		},
	}

	n.manager.TrackAndNotify(event)
	return nil
}

// BuildApprovalMessage builds a human-readable message for approval notifications
func BuildApprovalMessage(request *ApprovalRequest, workflow *ApprovalWorkflow, step *ApprovalStep) string {
	msg := fmt.Sprintf("**Approval Request** for %s\n\n", request.OperationType)
	msg += fmt.Sprintf("**Workflow:** %s\n", workflow.Name)
	msg += fmt.Sprintf("**Step:** %s (%d of %d)\n", step.Name, request.CurrentStep+1, request.TotalSteps)
	msg += fmt.Sprintf("**Requester:** %s\n", request.RequesterEmail)

	if request.ExpiresAt != nil {
		msg += fmt.Sprintf("**Expires:** %s\n", request.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	if len(request.OperationDetails) > 0 {
		msg += "\n**Details:**\n"
		for key, value := range request.OperationDetails {
			msg += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	return msg
}

// BuildDecisionMessage builds a human-readable message for decision notifications
func BuildDecisionMessage(request *ApprovalRequest, decision *ApprovalDecision) string {
	var status string
	if decision.Decision == StatusApproved {
		status = "✅ Approved"
	} else {
		status = "❌ Rejected"
	}

	msg := fmt.Sprintf("**Approval %s**\n\n", status)
	msg += fmt.Sprintf("**Operation:** %s\n", request.OperationType)
	msg += fmt.Sprintf("**Decision by:** %s\n", decision.ApproverEmail)

	if decision.Comment != "" {
		msg += fmt.Sprintf("**Comment:** %s\n", decision.Comment)
	}

	return msg
}
