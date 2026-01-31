package approval

import (
	"encoding/json"
	"time"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	StatusPending   ApprovalStatus = "pending"
	StatusApproved  ApprovalStatus = "approved"
	StatusRejected  ApprovalStatus = "rejected"
	StatusEscalated ApprovalStatus = "escalated"
	StatusExpired   ApprovalStatus = "expired"
	StatusCancelled ApprovalStatus = "cancelled"
)

// StepRequirement defines how approvals are required for a step
type StepRequirement string

const (
	RequirementAll StepRequirement = "all" // All approvers must approve
	RequirementAny StepRequirement = "any" // Any single approver can approve
)

// OperationType represents types of operations that can require approval
type OperationType string

const (
	OperationToolExecution     OperationType = "tool_execution"
	OperationAgentDeployment   OperationType = "agent_deployment"
	OperationConfigChange      OperationType = "config_change"
	OperationIntegrationSetup  OperationType = "integration_setup"
	OperationSensitiveData     OperationType = "sensitive_data"
	OperationCustom            OperationType = "custom"
)

// ApprovalWorkflow defines a multi-step approval process
type ApprovalWorkflow struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	OperationType  OperationType          `json:"operation_type"`
	Steps          []ApprovalStep         `json:"steps"`
	Conditions     *WorkflowConditions    `json:"conditions,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Active         bool                   `json:"active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      string                 `json:"created_by"`
}

// ApprovalStep defines a single step in an approval workflow
type ApprovalStep struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Order            int                    `json:"order"`
	ApproverRoles    []string               `json:"approver_roles"`    // Role names that can approve
	ApproverUserIDs  []string               `json:"approver_user_ids"` // Specific user IDs that can approve
	Requirement      StepRequirement        `json:"requirement"`       // all or any
	Timeout          time.Duration          `json:"timeout"`           // Duration before timeout
	AutoApprove      *AutoApproveCondition  `json:"auto_approve,omitempty"`
	EscalationConfig *EscalationConfig      `json:"escalation,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AutoApproveCondition defines conditions for automatic approval
type AutoApproveCondition struct {
	Enabled    bool                   `json:"enabled"`
	Expression string                 `json:"expression,omitempty"` // Expression to evaluate
	Conditions map[string]interface{} `json:"conditions,omitempty"` // Key-value conditions
}

// EscalationConfig defines escalation behavior for a step
type EscalationConfig struct {
	Enabled         bool     `json:"enabled"`
	EscalateAfter   time.Duration `json:"escalate_after"`   // Duration before escalating
	EscalateToRoles []string `json:"escalate_to_roles"` // Roles to escalate to
	EscalateToUsers []string `json:"escalate_to_users"` // Specific users to escalate to
	MaxEscalations  int      `json:"max_escalations"`   // Maximum number of escalations
}

// WorkflowConditions defines when a workflow should be triggered
type WorkflowConditions struct {
	ToolNames       []string               `json:"tool_names,omitempty"`       // Specific tool names
	AgentTypes      []string               `json:"agent_types,omitempty"`      // Agent types requiring approval
	ConfigKeys      []string               `json:"config_keys,omitempty"`      // Configuration keys
	MinRiskScore    int                    `json:"min_risk_score,omitempty"`   // Minimum risk score to trigger
	CustomConditions map[string]interface{} `json:"custom_conditions,omitempty"`
}

// ApprovalRequest represents a request for approval
type ApprovalRequest struct {
	ID               string                 `json:"id"`
	WorkflowID       string                 `json:"workflow_id"`
	OrganizationID   string                 `json:"organization_id"`
	RequesterID      string                 `json:"requester_id"`
	RequesterEmail   string                 `json:"requester_email,omitempty"`
	OperationType    OperationType          `json:"operation_type"`
	OperationDetails map[string]interface{} `json:"operation_details"`
	CurrentStep      int                    `json:"current_step"`
	TotalSteps       int                    `json:"total_steps"`
	Status           ApprovalStatus         `json:"status"`
	Priority         int                    `json:"priority"` // Higher = more urgent
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
}

// ApprovalDecision represents a decision made by an approver
type ApprovalDecision struct {
	ID           string         `json:"id"`
	RequestID    string         `json:"request_id"`
	StepOrder    int            `json:"step_order"`
	ApproverID   string         `json:"approver_id"`
	ApproverEmail string        `json:"approver_email,omitempty"`
	Decision     ApprovalStatus `json:"decision"`
	Comment      string         `json:"comment,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StepProgress tracks the progress of a specific step in an approval request
type StepProgress struct {
	RequestID       string             `json:"request_id"`
	StepOrder       int                `json:"step_order"`
	StepName        string             `json:"step_name"`
	Requirement     StepRequirement    `json:"requirement"`
	RequiredCount   int                `json:"required_count"`   // How many approvals needed
	ApprovedCount   int                `json:"approved_count"`
	RejectedCount   int                `json:"rejected_count"`
	PendingApprovers []string          `json:"pending_approvers"` // User IDs who haven't decided
	Decisions       []ApprovalDecision `json:"decisions"`
	Status          ApprovalStatus     `json:"status"`
	StartedAt       time.Time          `json:"started_at"`
	CompletedAt     *time.Time         `json:"completed_at,omitempty"`
}

// ApprovalRequestFilter for listing approval requests
type ApprovalRequestFilter struct {
	OrganizationID string           `json:"organization_id,omitempty"`
	RequesterID    string           `json:"requester_id,omitempty"`
	ApproverID     string           `json:"approver_id,omitempty"` // Requests where user can approve
	Status         []ApprovalStatus `json:"status,omitempty"`
	OperationType  []OperationType  `json:"operation_type,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
}

// WorkflowFilter for listing approval workflows
type WorkflowFilter struct {
	OrganizationID string          `json:"organization_id,omitempty"`
	OperationType  []OperationType `json:"operation_type,omitempty"`
	Active         *bool           `json:"active,omitempty"`
	Limit          int             `json:"limit,omitempty"`
	Offset         int             `json:"offset,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for ApprovalStep (Duration fields)
func (s ApprovalStep) MarshalJSON() ([]byte, error) {
	type Alias ApprovalStep
	return json.Marshal(&struct {
		Timeout int64 `json:"timeout"` // Milliseconds
		*Alias
	}{
		Timeout: int64(s.Timeout / time.Millisecond),
		Alias:   (*Alias)(&s),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for ApprovalStep (Duration fields)
func (s *ApprovalStep) UnmarshalJSON(data []byte) error {
	type Alias ApprovalStep
	aux := &struct {
		Timeout int64 `json:"timeout"` // Milliseconds
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	s.Timeout = time.Duration(aux.Timeout) * time.Millisecond
	return nil
}

// MarshalJSON implements custom JSON marshaling for EscalationConfig (Duration fields)
func (e EscalationConfig) MarshalJSON() ([]byte, error) {
	type Alias EscalationConfig
	return json.Marshal(&struct {
		EscalateAfter int64 `json:"escalate_after"` // Milliseconds
		*Alias
	}{
		EscalateAfter: int64(e.EscalateAfter / time.Millisecond),
		Alias:         (*Alias)(&e),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for EscalationConfig (Duration fields)
func (e *EscalationConfig) UnmarshalJSON(data []byte) error {
	type Alias EscalationConfig
	aux := &struct {
		EscalateAfter int64 `json:"escalate_after"` // Milliseconds
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	e.EscalateAfter = time.Duration(aux.EscalateAfter) * time.Millisecond
	return nil
}

// IsTerminal returns true if the status is a terminal state
func (s ApprovalStatus) IsTerminal() bool {
	return s == StatusApproved || s == StatusRejected || s == StatusExpired || s == StatusCancelled
}

// CanDecide returns true if the request is still pending and can receive decisions
func (r *ApprovalRequest) CanDecide() bool {
	return r.Status == StatusPending
}

// IsExpired returns true if the request has expired
func (r *ApprovalRequest) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// GetStep returns the step at the given order
func (w *ApprovalWorkflow) GetStep(order int) *ApprovalStep {
	for i := range w.Steps {
		if w.Steps[i].Order == order {
			return &w.Steps[i]
		}
	}
	return nil
}

// GetCurrentStep returns the current step for a request
func (r *ApprovalRequest) GetCurrentStepFromWorkflow(w *ApprovalWorkflow) *ApprovalStep {
	return w.GetStep(r.CurrentStep)
}

// Clone creates a deep copy of the approval workflow
func (w *ApprovalWorkflow) Clone() *ApprovalWorkflow {
	if w == nil {
		return nil
	}

	clone := *w
	clone.Steps = make([]ApprovalStep, len(w.Steps))
	copy(clone.Steps, w.Steps)

	if w.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range w.Metadata {
			clone.Metadata[k] = v
		}
	}

	if w.Conditions != nil {
		condClone := *w.Conditions
		clone.Conditions = &condClone
	}

	return &clone
}
