package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/approval"
)

// ApprovalRepository handles approval workflow persistence
type ApprovalRepository struct {
	db *sql.DB
}

// NewApprovalRepository creates a new approval repository
func NewApprovalRepository(db *sql.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// CreateWorkflow creates a new approval workflow
func (r *ApprovalRepository) CreateWorkflow(w *approval.ApprovalWorkflow) error {
	stepsJSON, err := json.Marshal(w.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	var conditionsJSON []byte
	if w.Conditions != nil {
		conditionsJSON, err = json.Marshal(w.Conditions)
		if err != nil {
			return fmt.Errorf("failed to marshal conditions: %w", err)
		}
	}

	var metadataJSON []byte
	if w.Metadata != nil {
		metadataJSON, err = json.Marshal(w.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO approval_workflows (
			id, organization_id, name, description, operation_type, steps,
			conditions, metadata, active, created_at, updated_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		w.ID,
		w.OrganizationID,
		w.Name,
		w.Description,
		string(w.OperationType),
		string(stepsJSON),
		string(conditionsJSON),
		string(metadataJSON),
		w.Active,
		w.CreatedAt,
		w.UpdatedAt,
		w.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert workflow: %w", err)
	}

	return nil
}

// GetWorkflowByID retrieves a workflow by ID
func (r *ApprovalRepository) GetWorkflowByID(id string) (*approval.ApprovalWorkflow, error) {
	query := `
		SELECT id, organization_id, name, description, operation_type, steps,
			conditions, metadata, active, created_at, updated_at, created_by
		FROM approval_workflows
		WHERE id = ?
	`

	var w approval.ApprovalWorkflow
	var stepsJSON, conditionsJSON, metadataJSON sql.NullString
	var operationType string

	err := r.db.QueryRow(query, id).Scan(
		&w.ID,
		&w.OrganizationID,
		&w.Name,
		&w.Description,
		&operationType,
		&stepsJSON,
		&conditionsJSON,
		&metadataJSON,
		&w.Active,
		&w.CreatedAt,
		&w.UpdatedAt,
		&w.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	w.OperationType = approval.OperationType(operationType)

	if stepsJSON.Valid && stepsJSON.String != "" {
		if err := json.Unmarshal([]byte(stepsJSON.String), &w.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
		}
	}

	if conditionsJSON.Valid && conditionsJSON.String != "" {
		if err := json.Unmarshal([]byte(conditionsJSON.String), &w.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &w.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &w, nil
}

// UpdateWorkflow updates an existing workflow
func (r *ApprovalRepository) UpdateWorkflow(w *approval.ApprovalWorkflow) error {
	stepsJSON, err := json.Marshal(w.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	var conditionsJSON []byte
	if w.Conditions != nil {
		conditionsJSON, err = json.Marshal(w.Conditions)
		if err != nil {
			return fmt.Errorf("failed to marshal conditions: %w", err)
		}
	}

	var metadataJSON []byte
	if w.Metadata != nil {
		metadataJSON, err = json.Marshal(w.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE approval_workflows
		SET name = ?, description = ?, operation_type = ?, steps = ?,
			conditions = ?, metadata = ?, active = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query,
		w.Name,
		w.Description,
		string(w.OperationType),
		string(stepsJSON),
		string(conditionsJSON),
		string(metadataJSON),
		w.Active,
		time.Now(),
		w.ID,
	)
	return err
}

// DeleteWorkflow deletes a workflow by ID
func (r *ApprovalRepository) DeleteWorkflow(id string) error {
	_, err := r.db.Exec("DELETE FROM approval_workflows WHERE id = ?", id)
	return err
}

// ListWorkflows returns workflows matching the filter
func (r *ApprovalRepository) ListWorkflows(filter *approval.WorkflowFilter) ([]*approval.ApprovalWorkflow, error) {
	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.OrganizationID != "" {
			conditions = append(conditions, "organization_id = ?")
			args = append(args, filter.OrganizationID)
		}
		if len(filter.OperationType) > 0 {
			placeholders := make([]string, len(filter.OperationType))
			for i, op := range filter.OperationType {
				placeholders[i] = "?"
				args = append(args, string(op))
			}
			conditions = append(conditions, fmt.Sprintf("operation_type IN (%s)", strings.Join(placeholders, ",")))
		}
		if filter.Active != nil {
			conditions = append(conditions, "active = ?")
			args = append(args, *filter.Active)
		}
	}

	query := `
		SELECT id, organization_id, name, description, operation_type, steps,
			conditions, metadata, active, created_at, updated_at, created_by
		FROM approval_workflows
	`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*approval.ApprovalWorkflow
	for rows.Next() {
		var w approval.ApprovalWorkflow
		var stepsJSON, conditionsJSON, metadataJSON sql.NullString
		var operationType string

		err := rows.Scan(
			&w.ID,
			&w.OrganizationID,
			&w.Name,
			&w.Description,
			&operationType,
			&stepsJSON,
			&conditionsJSON,
			&metadataJSON,
			&w.Active,
			&w.CreatedAt,
			&w.UpdatedAt,
			&w.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		w.OperationType = approval.OperationType(operationType)

		if stepsJSON.Valid && stepsJSON.String != "" {
			if err := json.Unmarshal([]byte(stepsJSON.String), &w.Steps); err != nil {
				return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
			}
		}

		if conditionsJSON.Valid && conditionsJSON.String != "" {
			if err := json.Unmarshal([]byte(conditionsJSON.String), &w.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &w.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		workflows = append(workflows, &w)
	}

	return workflows, nil
}

// GetActiveWorkflowForOperation finds an active workflow for a given operation type
func (r *ApprovalRepository) GetActiveWorkflowForOperation(orgID string, opType approval.OperationType) (*approval.ApprovalWorkflow, error) {
	active := true
	workflows, err := r.ListWorkflows(&approval.WorkflowFilter{
		OrganizationID: orgID,
		OperationType:  []approval.OperationType{opType},
		Active:         &active,
		Limit:          1,
	})
	if err != nil {
		return nil, err
	}
	if len(workflows) == 0 {
		return nil, nil
	}
	return workflows[0], nil
}

// CreateRequest creates a new approval request
func (r *ApprovalRepository) CreateRequest(req *approval.ApprovalRequest) error {
	detailsJSON, err := json.Marshal(req.OperationDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal operation details: %w", err)
	}

	var metadataJSON []byte
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO approval_requests (
			id, workflow_id, organization_id, requester_id, requester_email,
			operation_type, operation_details, current_step, total_steps,
			status, priority, expires_at, metadata, created_at, updated_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		req.ID,
		req.WorkflowID,
		req.OrganizationID,
		req.RequesterID,
		req.RequesterEmail,
		string(req.OperationType),
		string(detailsJSON),
		req.CurrentStep,
		req.TotalSteps,
		string(req.Status),
		req.Priority,
		req.ExpiresAt,
		string(metadataJSON),
		req.CreatedAt,
		req.UpdatedAt,
		req.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert request: %w", err)
	}

	return nil
}

// GetRequestByID retrieves a request by ID
func (r *ApprovalRepository) GetRequestByID(id string) (*approval.ApprovalRequest, error) {
	query := `
		SELECT id, workflow_id, organization_id, requester_id, requester_email,
			operation_type, operation_details, current_step, total_steps,
			status, priority, expires_at, metadata, created_at, updated_at, completed_at
		FROM approval_requests
		WHERE id = ?
	`

	var req approval.ApprovalRequest
	var detailsJSON, metadataJSON sql.NullString
	var operationType, status string
	var expiresAt, completedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&req.ID,
		&req.WorkflowID,
		&req.OrganizationID,
		&req.RequesterID,
		&req.RequesterEmail,
		&operationType,
		&detailsJSON,
		&req.CurrentStep,
		&req.TotalSteps,
		&status,
		&req.Priority,
		&expiresAt,
		&metadataJSON,
		&req.CreatedAt,
		&req.UpdatedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %w", err)
	}

	req.OperationType = approval.OperationType(operationType)
	req.Status = approval.ApprovalStatus(status)

	if expiresAt.Valid {
		req.ExpiresAt = &expiresAt.Time
	}
	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}

	if detailsJSON.Valid && detailsJSON.String != "" {
		if err := json.Unmarshal([]byte(detailsJSON.String), &req.OperationDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal operation details: %w", err)
		}
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &req.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &req, nil
}

// UpdateRequest updates an existing request
func (r *ApprovalRepository) UpdateRequest(req *approval.ApprovalRequest) error {
	detailsJSON, err := json.Marshal(req.OperationDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal operation details: %w", err)
	}

	var metadataJSON []byte
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE approval_requests
		SET current_step = ?, status = ?, priority = ?, expires_at = ?,
			metadata = ?, operation_details = ?, updated_at = ?, completed_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query,
		req.CurrentStep,
		string(req.Status),
		req.Priority,
		req.ExpiresAt,
		string(metadataJSON),
		string(detailsJSON),
		time.Now(),
		req.CompletedAt,
		req.ID,
	)
	return err
}

// ListRequests returns requests matching the filter
func (r *ApprovalRepository) ListRequests(filter *approval.ApprovalRequestFilter) ([]*approval.ApprovalRequest, error) {
	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.OrganizationID != "" {
			conditions = append(conditions, "organization_id = ?")
			args = append(args, filter.OrganizationID)
		}
		if filter.RequesterID != "" {
			conditions = append(conditions, "requester_id = ?")
			args = append(args, filter.RequesterID)
		}
		if len(filter.Status) > 0 {
			placeholders := make([]string, len(filter.Status))
			for i, s := range filter.Status {
				placeholders[i] = "?"
				args = append(args, string(s))
			}
			conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}
		if len(filter.OperationType) > 0 {
			placeholders := make([]string, len(filter.OperationType))
			for i, op := range filter.OperationType {
				placeholders[i] = "?"
				args = append(args, string(op))
			}
			conditions = append(conditions, fmt.Sprintf("operation_type IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	query := `
		SELECT id, workflow_id, organization_id, requester_id, requester_email,
			operation_type, operation_details, current_step, total_steps,
			status, priority, expires_at, metadata, created_at, updated_at, completed_at
		FROM approval_requests
	`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY priority DESC, created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list requests: %w", err)
	}
	defer rows.Close()

	var requests []*approval.ApprovalRequest
	for rows.Next() {
		var req approval.ApprovalRequest
		var detailsJSON, metadataJSON sql.NullString
		var operationType, status string
		var expiresAt, completedAt sql.NullTime

		err := rows.Scan(
			&req.ID,
			&req.WorkflowID,
			&req.OrganizationID,
			&req.RequesterID,
			&req.RequesterEmail,
			&operationType,
			&detailsJSON,
			&req.CurrentStep,
			&req.TotalSteps,
			&status,
			&req.Priority,
			&expiresAt,
			&metadataJSON,
			&req.CreatedAt,
			&req.UpdatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		req.OperationType = approval.OperationType(operationType)
		req.Status = approval.ApprovalStatus(status)

		if expiresAt.Valid {
			req.ExpiresAt = &expiresAt.Time
		}
		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}

		if detailsJSON.Valid && detailsJSON.String != "" {
			if err := json.Unmarshal([]byte(detailsJSON.String), &req.OperationDetails); err != nil {
				return nil, fmt.Errorf("failed to unmarshal operation details: %w", err)
			}
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &req.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

// GetPendingRequestsForApprover returns pending requests where the user can approve
func (r *ApprovalRepository) GetPendingRequestsForApprover(orgID, userID string, userRoles []string) ([]*approval.ApprovalRequest, error) {
	// Get all pending requests for the organization
	pendingRequests, err := r.ListRequests(&approval.ApprovalRequestFilter{
		OrganizationID: orgID,
		Status:         []approval.ApprovalStatus{approval.StatusPending},
	})
	if err != nil {
		return nil, err
	}

	var result []*approval.ApprovalRequest
	for _, req := range pendingRequests {
		// Get the workflow to check the current step's approvers
		workflow, err := r.GetWorkflowByID(req.WorkflowID)
		if err != nil || workflow == nil {
			continue
		}

		step := workflow.GetStep(req.CurrentStep)
		if step == nil {
			continue
		}

		// Check if user is an approver for this step
		isApprover := false

		// Check specific user IDs
		for _, id := range step.ApproverUserIDs {
			if id == userID {
				isApprover = true
				break
			}
		}

		// Check roles
		if !isApprover {
			for _, approverRole := range step.ApproverRoles {
				for _, userRole := range userRoles {
					if approverRole == userRole {
						isApprover = true
						break
					}
				}
				if isApprover {
					break
				}
			}
		}

		// Check if user already decided
		if isApprover {
			decisions, _ := r.GetDecisionsForRequest(req.ID)
			alreadyDecided := false
			for _, d := range decisions {
				if d.ApproverID == userID && d.StepOrder == req.CurrentStep {
					alreadyDecided = true
					break
				}
			}
			if !alreadyDecided {
				result = append(result, req)
			}
		}
	}

	return result, nil
}

// CreateDecision creates a new approval decision
func (r *ApprovalRepository) CreateDecision(d *approval.ApprovalDecision) error {
	var metadataJSON []byte
	var err error
	if d.Metadata != nil {
		metadataJSON, err = json.Marshal(d.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO approval_decisions (
			id, request_id, step_order, approver_id, approver_email,
			decision, comment, created_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		d.ID,
		d.RequestID,
		d.StepOrder,
		d.ApproverID,
		d.ApproverEmail,
		string(d.Decision),
		d.Comment,
		d.CreatedAt,
		string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert decision: %w", err)
	}

	return nil
}

// GetDecisionsForRequest returns all decisions for a request
func (r *ApprovalRepository) GetDecisionsForRequest(requestID string) ([]*approval.ApprovalDecision, error) {
	query := `
		SELECT id, request_id, step_order, approver_id, approver_email,
			decision, comment, created_at, metadata
		FROM approval_decisions
		WHERE request_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*approval.ApprovalDecision
	for rows.Next() {
		var d approval.ApprovalDecision
		var metadataJSON sql.NullString
		var decision string

		err := rows.Scan(
			&d.ID,
			&d.RequestID,
			&d.StepOrder,
			&d.ApproverID,
			&d.ApproverEmail,
			&decision,
			&d.Comment,
			&d.CreatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}

		d.Decision = approval.ApprovalStatus(decision)

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &d.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		decisions = append(decisions, &d)
	}

	return decisions, nil
}

// GetDecisionsForStep returns decisions for a specific step of a request
func (r *ApprovalRepository) GetDecisionsForStep(requestID string, stepOrder int) ([]*approval.ApprovalDecision, error) {
	query := `
		SELECT id, request_id, step_order, approver_id, approver_email,
			decision, comment, created_at, metadata
		FROM approval_decisions
		WHERE request_id = ? AND step_order = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, requestID, stepOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*approval.ApprovalDecision
	for rows.Next() {
		var d approval.ApprovalDecision
		var metadataJSON sql.NullString
		var decision string

		err := rows.Scan(
			&d.ID,
			&d.RequestID,
			&d.StepOrder,
			&d.ApproverID,
			&d.ApproverEmail,
			&decision,
			&d.Comment,
			&d.CreatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}

		d.Decision = approval.ApprovalStatus(decision)

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &d.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		decisions = append(decisions, &d)
	}

	return decisions, nil
}

// GetExpiredRequests returns requests that have expired but are still pending
func (r *ApprovalRepository) GetExpiredRequests() ([]*approval.ApprovalRequest, error) {
	query := `
		SELECT id, workflow_id, organization_id, requester_id, requester_email,
			operation_type, operation_details, current_step, total_steps,
			status, priority, expires_at, metadata, created_at, updated_at, completed_at
		FROM approval_requests
		WHERE status = ? AND expires_at IS NOT NULL AND expires_at < ?
	`

	rows, err := r.db.Query(query, string(approval.StatusPending), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired requests: %w", err)
	}
	defer rows.Close()

	var requests []*approval.ApprovalRequest
	for rows.Next() {
		var req approval.ApprovalRequest
		var detailsJSON, metadataJSON sql.NullString
		var operationType, status string
		var expiresAt, completedAt sql.NullTime

		err := rows.Scan(
			&req.ID,
			&req.WorkflowID,
			&req.OrganizationID,
			&req.RequesterID,
			&req.RequesterEmail,
			&operationType,
			&detailsJSON,
			&req.CurrentStep,
			&req.TotalSteps,
			&status,
			&req.Priority,
			&expiresAt,
			&metadataJSON,
			&req.CreatedAt,
			&req.UpdatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		req.OperationType = approval.OperationType(operationType)
		req.Status = approval.ApprovalStatus(status)

		if expiresAt.Valid {
			req.ExpiresAt = &expiresAt.Time
		}
		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}

		if detailsJSON.Valid && detailsJSON.String != "" {
			if err := json.Unmarshal([]byte(detailsJSON.String), &req.OperationDetails); err != nil {
				return nil, fmt.Errorf("failed to unmarshal operation details: %w", err)
			}
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &req.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

// CountPendingByOrganization returns the count of pending requests for an organization
func (r *ApprovalRepository) CountPendingByOrganization(orgID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM approval_requests WHERE organization_id = ? AND status = ?",
		orgID,
		string(approval.StatusPending),
	).Scan(&count)
	return count, err
}
