package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/workflowtypes"
)

// WorkflowRepository handles workflow persistence
type WorkflowRepository struct {
	db *sql.DB
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(db *sql.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Create creates a new workflow
func (r *WorkflowRepository) Create(w *workflowtypes.Workflow) error {
	// Serialize steps
	stepsJSON, err := json.Marshal(w.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	// Serialize state
	var stateJSON []byte
	if w.State != nil {
		stateJSON, err = json.Marshal(w.State)
		if err != nil {
			return fmt.Errorf("failed to marshal state: %w", err)
		}
	}

	// Insert workflow
	query := `
		INSERT INTO workflows (id, user_id, name, description, definition, status, current_step, state, error, created_at, updated_at, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		w.ID,
		w.UserID,
		w.Name,
		w.Description,
		string(stepsJSON),
		string(w.Status),
		w.CurrentStep,
		string(stateJSON),
		w.Error,
		w.CreatedAt,
		w.UpdatedAt,
		w.StartedAt,
		w.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert workflow: %w", err)
	}

	// Insert individual steps
	for i, step := range w.Steps {
		if err := r.createStep(w.ID, i, &step); err != nil {
			return fmt.Errorf("failed to create step %d: %w", i, err)
		}
	}

	return nil
}

// createStep inserts a workflow step
func (r *WorkflowRepository) createStep(workflowID string, index int, step *workflowtypes.Step) error {
	configJSON, err := json.Marshal(step.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal step config: %w", err)
	}

	query := `
		INSERT INTO workflow_steps (id, workflow_id, step_index, name, type, config, status, result, started_at, completed_at, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL, NULL, NULL, NULL)
	`

	_, err = r.db.Exec(query,
		step.ID,
		workflowID,
		index,
		step.Name,
		string(step.Type),
		string(configJSON),
		string(workflowtypes.StepStatusPending),
	)
	return err
}

// GetByID retrieves a workflow by ID
func (r *WorkflowRepository) GetByID(id string) (*workflowtypes.Workflow, error) {
	query := `
		SELECT id, user_id, name, description, definition, status, current_step, state, error, created_at, updated_at, started_at, completed_at
		FROM workflows
		WHERE id = ?
	`

	var w workflowtypes.Workflow
	var stepsJSON, stateJSON sql.NullString
	var startedAt, completedAt sql.NullTime
	var status string

	err := r.db.QueryRow(query, id).Scan(
		&w.ID,
		&w.UserID,
		&w.Name,
		&w.Description,
		&stepsJSON,
		&status,
		&w.CurrentStep,
		&stateJSON,
		&w.Error,
		&w.CreatedAt,
		&w.UpdatedAt,
		&startedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	w.Status = workflowtypes.WorkflowStatus(status)

	if startedAt.Valid {
		w.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		w.CompletedAt = &completedAt.Time
	}

	// Parse steps
	if stepsJSON.Valid && stepsJSON.String != "" {
		if err := json.Unmarshal([]byte(stepsJSON.String), &w.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
		}
	}

	// Parse state
	if stateJSON.Valid && stateJSON.String != "" {
		if err := json.Unmarshal([]byte(stateJSON.String), &w.State); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	return &w, nil
}

// Update updates a workflow
func (r *WorkflowRepository) Update(w *workflowtypes.Workflow) error {
	// Serialize steps
	stepsJSON, err := json.Marshal(w.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	// Serialize state
	var stateJSON []byte
	if w.State != nil {
		stateJSON, err = json.Marshal(w.State)
		if err != nil {
			return fmt.Errorf("failed to marshal state: %w", err)
		}
	}

	query := `
		UPDATE workflows
		SET name = ?, description = ?, definition = ?, status = ?, current_step = ?, state = ?, error = ?, updated_at = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query,
		w.Name,
		w.Description,
		string(stepsJSON),
		string(w.Status),
		w.CurrentStep,
		string(stateJSON),
		w.Error,
		time.Now(),
		w.StartedAt,
		w.CompletedAt,
		w.ID,
	)
	return err
}

// UpdateState updates just the workflow state and current step
func (r *WorkflowRepository) UpdateState(id string, state map[string]interface{}, currentStep int, status workflowtypes.WorkflowStatus) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	query := `
		UPDATE workflows
		SET state = ?, current_step = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query, string(stateJSON), currentStep, string(status), time.Now(), id)
	return err
}

// List returns workflows matching the filter
func (r *WorkflowRepository) List(filter *workflowtypes.WorkflowFilter) ([]*workflowtypes.Workflow, error) {
	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.UserID != "" {
			conditions = append(conditions, "user_id = ?")
			args = append(args, filter.UserID)
		}
		if len(filter.Status) > 0 {
			placeholders := make([]string, len(filter.Status))
			for i, s := range filter.Status {
				placeholders[i] = "?"
				args = append(args, string(s))
			}
			conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}
		if filter.Name != "" {
			conditions = append(conditions, "name LIKE ?")
			args = append(args, "%"+filter.Name+"%")
		}
	}

	query := "SELECT id, user_id, name, description, definition, status, current_step, state, error, created_at, updated_at, started_at, completed_at FROM workflows"
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

	var workflows []*workflowtypes.Workflow

	for rows.Next() {
		var w workflowtypes.Workflow
		var stepsJSON, stateJSON sql.NullString
		var startedAt, completedAt sql.NullTime
		var status string

		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Name,
			&w.Description,
			&stepsJSON,
			&status,
			&w.CurrentStep,
			&stateJSON,
			&w.Error,
			&w.CreatedAt,
			&w.UpdatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		w.Status = workflowtypes.WorkflowStatus(status)

		if startedAt.Valid {
			w.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			w.CompletedAt = &completedAt.Time
		}

		// Parse steps
		if stepsJSON.Valid && stepsJSON.String != "" {
			if err := json.Unmarshal([]byte(stepsJSON.String), &w.Steps); err != nil {
				return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
			}
		}

		// Parse state
		if stateJSON.Valid && stateJSON.String != "" {
			if err := json.Unmarshal([]byte(stateJSON.String), &w.State); err != nil {
				return nil, fmt.Errorf("failed to unmarshal state: %w", err)
			}
		}

		workflows = append(workflows, &w)
	}

	return workflows, nil
}

// Delete deletes a workflow by ID
func (r *WorkflowRepository) Delete(id string) error {
	// Delete steps first (foreign key constraint)
	_, err := r.db.Exec("DELETE FROM workflow_steps WHERE workflow_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow steps: %w", err)
	}

	// Delete workflow
	_, err = r.db.Exec("DELETE FROM workflows WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// UpdateStepStatus updates the status of a specific step
func (r *WorkflowRepository) UpdateStepStatus(stepID string, status workflowtypes.StepStatus, result interface{}, errMsg string) error {
	var resultJSON []byte
	var err error

	if result != nil {
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	query := `
		UPDATE workflow_steps
		SET status = ?, result = ?, error = ?, completed_at = ?
		WHERE id = ?
	`

	var completedAt interface{}
	if status == workflowtypes.StepStatusCompleted || status == workflowtypes.StepStatusFailed || status == workflowtypes.StepStatusSkipped {
		now := time.Now()
		completedAt = now
	}

	_, err = r.db.Exec(query, string(status), string(resultJSON), errMsg, completedAt, stepID)
	return err
}

// GetWorkflowSteps retrieves all steps for a workflow
func (r *WorkflowRepository) GetWorkflowSteps(workflowID string) ([]workflowtypes.Step, error) {
	query := `
		SELECT id, name, type, config, status, result, started_at, completed_at, error
		FROM workflow_steps
		WHERE workflow_id = ?
		ORDER BY step_index
	`

	rows, err := r.db.Query(query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow steps: %w", err)
	}
	defer rows.Close()

	var steps []workflowtypes.Step

	for rows.Next() {
		var step workflowtypes.Step
		var configJSON, resultJSON sql.NullString
		var status string
		var startedAt, completedAt sql.NullTime
		var errMsg sql.NullString

		err := rows.Scan(
			&step.ID,
			&step.Name,
			&status,
			&configJSON,
			&status,
			&resultJSON,
			&startedAt,
			&completedAt,
			&errMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}

		step.Type = workflowtypes.StepType(status)

		// Parse config
		if configJSON.Valid && configJSON.String != "" {
			if err := json.Unmarshal([]byte(configJSON.String), &step.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal step config: %w", err)
			}
		}

		steps = append(steps, step)
	}

	return steps, nil
}

// CountByUserID returns the count of workflows for a user
func (r *WorkflowRepository) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM workflows WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// GetByUserIDAndStatus returns workflows for a user with specific status
func (r *WorkflowRepository) GetByUserIDAndStatus(userID string, status workflowtypes.WorkflowStatus) ([]*workflowtypes.Workflow, error) {
	return r.List(&workflowtypes.WorkflowFilter{
		UserID: userID,
		Status: []workflowtypes.WorkflowStatus{status},
	})
}

// GetRunningWorkflows returns all running workflows
func (r *WorkflowRepository) GetRunningWorkflows() ([]*workflowtypes.Workflow, error) {
	return r.List(&workflowtypes.WorkflowFilter{
		Status: []workflowtypes.WorkflowStatus{workflowtypes.StatusRunning, workflowtypes.StatusPaused},
	})
}
