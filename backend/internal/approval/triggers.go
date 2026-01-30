package approval

import (
	"context"
	"fmt"
	"log"
)

// TriggerResult represents the result of checking if approval is required
type TriggerResult struct {
	RequiresApproval bool              `json:"requires_approval"`
	Workflow         *ApprovalWorkflow `json:"workflow,omitempty"`
	Reason           string            `json:"reason,omitempty"`
}

// OperationContext provides context for an operation being checked
type OperationContext struct {
	OrganizationID string
	UserID         string
	UserEmail      string
	OperationType  OperationType
	Details        map[string]interface{}
}

// Triggers handles checking if operations require approval
type Triggers struct {
	repo   Repository
	engine *Engine
}

// NewTriggers creates a new triggers handler
func NewTriggers(repo Repository, engine *Engine) *Triggers {
	return &Triggers{
		repo:   repo,
		engine: engine,
	}
}

// CheckApprovalRequired checks if an operation requires approval
func (t *Triggers) CheckApprovalRequired(ctx context.Context, opCtx *OperationContext) (*TriggerResult, error) {
	// Find active workflow for this operation type
	workflow, err := t.repo.GetActiveWorkflowForOperation(opCtx.OrganizationID, opCtx.OperationType)
	if err != nil {
		return nil, fmt.Errorf("failed to check for workflow: %w", err)
	}

	if workflow == nil {
		return &TriggerResult{
			RequiresApproval: false,
			Reason:           "no active workflow for operation type",
		}, nil
	}

	// Check if conditions match
	if workflow.Conditions != nil {
		if !t.matchesConditions(workflow.Conditions, opCtx.Details) {
			return &TriggerResult{
				RequiresApproval: false,
				Reason:           "operation does not match workflow conditions",
			}, nil
		}
	}

	return &TriggerResult{
		RequiresApproval: true,
		Workflow:         workflow,
		Reason:           fmt.Sprintf("workflow '%s' requires approval", workflow.Name),
	}, nil
}

// TriggerApproval creates an approval request if required
func (t *Triggers) TriggerApproval(ctx context.Context, opCtx *OperationContext) (*ApprovalRequest, error) {
	result, err := t.CheckApprovalRequired(ctx, opCtx)
	if err != nil {
		return nil, err
	}

	if !result.RequiresApproval {
		return nil, nil // No approval needed
	}

	request, err := t.engine.CreateRequest(
		ctx,
		opCtx.OrganizationID,
		opCtx.UserID,
		opCtx.UserEmail,
		opCtx.OperationType,
		opCtx.Details,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	return request, nil
}

// matchesConditions checks if operation details match workflow conditions
func (t *Triggers) matchesConditions(conditions *WorkflowConditions, details map[string]interface{}) bool {
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
			riskScoreFloat, ok := details["risk_score"].(float64)
			if !ok {
				return false
			}
			riskScore = int(riskScoreFloat)
		}
		if riskScore < conditions.MinRiskScore {
			return false
		}
	}

	return true
}

// ToolExecutionTrigger checks if a tool execution requires approval
func (t *Triggers) ToolExecutionTrigger(ctx context.Context, orgID, userID, userEmail, toolName string, params map[string]interface{}) (*TriggerResult, error) {
	details := map[string]interface{}{
		"tool_name":  toolName,
		"parameters": params,
	}

	return t.CheckApprovalRequired(ctx, &OperationContext{
		OrganizationID: orgID,
		UserID:         userID,
		UserEmail:      userEmail,
		OperationType:  OperationToolExecution,
		Details:        details,
	})
}

// AgentDeploymentTrigger checks if an agent deployment requires approval
func (t *Triggers) AgentDeploymentTrigger(ctx context.Context, orgID, userID, userEmail, agentType, agentName string, config map[string]interface{}) (*TriggerResult, error) {
	details := map[string]interface{}{
		"agent_type": agentType,
		"agent_name": agentName,
		"config":     config,
	}

	return t.CheckApprovalRequired(ctx, &OperationContext{
		OrganizationID: orgID,
		UserID:         userID,
		UserEmail:      userEmail,
		OperationType:  OperationAgentDeployment,
		Details:        details,
	})
}

// ConfigChangeTrigger checks if a configuration change requires approval
func (t *Triggers) ConfigChangeTrigger(ctx context.Context, orgID, userID, userEmail, configKey string, oldValue, newValue interface{}) (*TriggerResult, error) {
	details := map[string]interface{}{
		"config_key": configKey,
		"old_value":  oldValue,
		"new_value":  newValue,
	}

	return t.CheckApprovalRequired(ctx, &OperationContext{
		OrganizationID: orgID,
		UserID:         userID,
		UserEmail:      userEmail,
		OperationType:  OperationConfigChange,
		Details:        details,
	})
}

// IntegrationSetupTrigger checks if an integration setup requires approval
func (t *Triggers) IntegrationSetupTrigger(ctx context.Context, orgID, userID, userEmail, integrationType string, config map[string]interface{}) (*TriggerResult, error) {
	details := map[string]interface{}{
		"integration_type": integrationType,
		"config":           config,
	}

	return t.CheckApprovalRequired(ctx, &OperationContext{
		OrganizationID: orgID,
		UserID:         userID,
		UserEmail:      userEmail,
		OperationType:  OperationIntegrationSetup,
		Details:        details,
	})
}

// SensitiveDataTrigger checks if a sensitive data operation requires approval
func (t *Triggers) SensitiveDataTrigger(ctx context.Context, orgID, userID, userEmail, dataType, operation string, riskScore int) (*TriggerResult, error) {
	details := map[string]interface{}{
		"data_type":  dataType,
		"operation":  operation,
		"risk_score": riskScore,
	}

	return t.CheckApprovalRequired(ctx, &OperationContext{
		OrganizationID: orgID,
		UserID:         userID,
		UserEmail:      userEmail,
		OperationType:  OperationSensitiveData,
		Details:        details,
	})
}

// WaitForApproval waits for an approval request to complete and returns the result
func (t *Triggers) WaitForApproval(ctx context.Context, requestID string) (ApprovalStatus, error) {
	eventChan := t.engine.Subscribe()
	defer t.engine.Unsubscribe(eventChan)

	for {
		select {
		case event := <-eventChan:
			if event.RequestID != requestID {
				continue
			}

			switch event.Type {
			case EventRequestApproved:
				return StatusApproved, nil
			case EventRequestRejected:
				return StatusRejected, nil
			case EventRequestExpired:
				return StatusExpired, nil
			case EventRequestCancelled:
				return StatusCancelled, nil
			}

		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// ExecuteWithApproval executes an operation that may require approval
// If approval is required, it creates a request and waits for it
// The operationFunc is called only if approved (or no approval needed)
func (t *Triggers) ExecuteWithApproval(
	ctx context.Context,
	opCtx *OperationContext,
	operationFunc func(ctx context.Context) error,
) error {
	result, err := t.CheckApprovalRequired(ctx, opCtx)
	if err != nil {
		return err
	}

	if !result.RequiresApproval {
		// No approval needed, execute directly
		return operationFunc(ctx)
	}

	// Create approval request
	request, err := t.engine.CreateRequest(
		ctx,
		opCtx.OrganizationID,
		opCtx.UserID,
		opCtx.UserEmail,
		opCtx.OperationType,
		opCtx.Details,
	)
	if err != nil {
		return fmt.Errorf("failed to create approval request: %w", err)
	}

	log.Printf("Approval required for operation %s, request ID: %s", opCtx.OperationType, request.ID)

	// Wait for approval
	status, err := t.WaitForApproval(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("error waiting for approval: %w", err)
	}

	switch status {
	case StatusApproved:
		log.Printf("Approval granted for request %s, executing operation", request.ID)
		return operationFunc(ctx)
	case StatusRejected:
		return fmt.Errorf("operation rejected by approver")
	case StatusExpired:
		return fmt.Errorf("approval request expired")
	case StatusCancelled:
		return fmt.Errorf("approval request cancelled")
	default:
		return fmt.Errorf("unexpected approval status: %s", status)
	}
}

// ExecuteWithApprovalAsync creates an approval request and returns immediately
// The caller must handle the approval workflow separately
func (t *Triggers) ExecuteWithApprovalAsync(
	ctx context.Context,
	opCtx *OperationContext,
) (*ApprovalRequest, bool, error) {
	result, err := t.CheckApprovalRequired(ctx, opCtx)
	if err != nil {
		return nil, false, err
	}

	if !result.RequiresApproval {
		return nil, false, nil // No approval needed
	}

	// Create approval request
	request, err := t.engine.CreateRequest(
		ctx,
		opCtx.OrganizationID,
		opCtx.UserID,
		opCtx.UserEmail,
		opCtx.OperationType,
		opCtx.Details,
	)
	if err != nil {
		return nil, true, fmt.Errorf("failed to create approval request: %w", err)
	}

	return request, true, nil
}
