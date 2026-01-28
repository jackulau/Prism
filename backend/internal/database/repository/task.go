package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentTask represents a persistent agent task in the database
type AgentTask struct {
	ID           string
	UserID       string
	Prompt       string
	Context      string
	Priority     int
	Status       string
	AgentConfig  map[string]interface{}
	Metadata     map[string]interface{}
	Result       map[string]interface{}
	Error        string
	CallbackURL  string
	CallbackData map[string]string
	CreatedAt    time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
}

// AgentTaskRepository handles agent task database operations
type AgentTaskRepository struct {
	db *sql.DB
}

// NewAgentTaskRepository creates a new agent task repository
func NewAgentTaskRepository(db *sql.DB) *AgentTaskRepository {
	return &AgentTaskRepository{db: db}
}

// Create creates a new agent task
func (r *AgentTaskRepository) Create(task *AgentTask) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Status == "" {
		task.Status = "pending"
	}

	agentConfigJSON, err := json.Marshal(task.AgentConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal agent config: %w", err)
	}

	metadataJSON, err := json.Marshal(task.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	callbackDataJSON, err := json.Marshal(task.CallbackData)
	if err != nil {
		return fmt.Errorf("failed to marshal callback data: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO agent_tasks (id, user_id, prompt, context, priority, status, agent_config, metadata, callback_url, callback_data, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.UserID, task.Prompt, task.Context, task.Priority, task.Status,
		string(agentConfigJSON), string(metadataJSON), task.CallbackURL, string(callbackDataJSON), task.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent task: %w", err)
	}

	return nil
}

// Update updates an existing agent task
func (r *AgentTaskRepository) Update(task *AgentTask) error {
	var resultJSON []byte
	var err error
	if task.Result != nil {
		resultJSON, err = json.Marshal(task.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	metadataJSON, err := json.Marshal(task.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.Exec(
		`UPDATE agent_tasks SET status = ?, result = ?, error = ?, metadata = ?, started_at = ?, completed_at = ?
		 WHERE id = ?`,
		task.Status, string(resultJSON), task.Error, string(metadataJSON), task.StartedAt, task.CompletedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent task: %w", err)
	}

	return nil
}

// UpdateStatus updates only the status field of a task
func (r *AgentTaskRepository) UpdateStatus(id string, status string) error {
	var err error
	if status == "running" {
		now := time.Now()
		_, err = r.db.Exec(`UPDATE agent_tasks SET status = ?, started_at = ? WHERE id = ?`, status, now, id)
	} else if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		_, err = r.db.Exec(`UPDATE agent_tasks SET status = ?, completed_at = ? WHERE id = ?`, status, now, id)
	} else {
		_, err = r.db.Exec(`UPDATE agent_tasks SET status = ? WHERE id = ?`, status, id)
	}
	if err != nil {
		return fmt.Errorf("failed to update agent task status: %w", err)
	}
	return nil
}

// SetResult sets the result of a completed task
func (r *AgentTaskRepository) SetResult(id string, result map[string]interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = r.db.Exec(`UPDATE agent_tasks SET result = ? WHERE id = ?`, string(resultJSON), id)
	if err != nil {
		return fmt.Errorf("failed to set agent task result: %w", err)
	}
	return nil
}

// SetError sets the error message for a failed task
func (r *AgentTaskRepository) SetError(id string, errorMsg string) error {
	_, err := r.db.Exec(`UPDATE agent_tasks SET error = ? WHERE id = ?`, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to set agent task error: %w", err)
	}
	return nil
}

// GetByID retrieves an agent task by ID
func (r *AgentTaskRepository) GetByID(id string) (*AgentTask, error) {
	task := &AgentTask{}
	var context, agentConfigJSON, metadataJSON, resultJSON, errorStr, callbackURL, callbackDataJSON sql.NullString
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, prompt, context, priority, status, agent_config, metadata, result, error, callback_url, callback_data, created_at, started_at, completed_at
		 FROM agent_tasks WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.UserID, &task.Prompt, &context, &task.Priority, &task.Status,
		&agentConfigJSON, &metadataJSON, &resultJSON, &errorStr, &callbackURL, &callbackDataJSON,
		&task.CreatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent task: %w", err)
	}

	task.Context = context.String
	task.Error = errorStr.String
	task.CallbackURL = callbackURL.String

	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	if agentConfigJSON.Valid && agentConfigJSON.String != "" {
		if err := json.Unmarshal([]byte(agentConfigJSON.String), &task.AgentConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal agent config: %w", err)
		}
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if resultJSON.Valid && resultJSON.String != "" {
		if err := json.Unmarshal([]byte(resultJSON.String), &task.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	if callbackDataJSON.Valid && callbackDataJSON.String != "" {
		if err := json.Unmarshal([]byte(callbackDataJSON.String), &task.CallbackData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
		}
	}

	return task, nil
}

// Delete deletes an agent task by ID
func (r *AgentTaskRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM agent_tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent task: %w", err)
	}
	return nil
}

// ListByStatus retrieves tasks by status with pagination
func (r *AgentTaskRepository) ListByStatus(status string, limit, offset int) ([]*AgentTask, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, prompt, context, priority, status, agent_config, metadata, result, error, callback_url, callback_data, created_at, started_at, completed_at
		 FROM agent_tasks WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		status, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tasks by status: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// ListByUserID retrieves tasks by user ID with pagination
func (r *AgentTaskRepository) ListByUserID(userID string, limit, offset int) ([]*AgentTask, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, prompt, context, priority, status, agent_config, metadata, result, error, callback_url, callback_data, created_at, started_at, completed_at
		 FROM agent_tasks WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tasks by user: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// ListByUserIDAndStatus retrieves tasks by user ID and status with pagination
func (r *AgentTaskRepository) ListByUserIDAndStatus(userID, status string, limit, offset int) ([]*AgentTask, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, prompt, context, priority, status, agent_config, metadata, result, error, callback_url, callback_data, created_at, started_at, completed_at
		 FROM agent_tasks WHERE user_id = ? AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, status, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tasks by user and status: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// ListPending retrieves pending tasks ordered by priority (descending) and created_at (ascending)
func (r *AgentTaskRepository) ListPending(limit int) ([]*AgentTask, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, prompt, context, priority, status, agent_config, metadata, result, error, callback_url, callback_data, created_at, started_at, completed_at
		 FROM agent_tasks WHERE status = 'pending' ORDER BY priority DESC, created_at ASC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending agent tasks: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// CleanupOld removes completed tasks older than the specified time
func (r *AgentTaskRepository) CleanupOld(before time.Time) (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM agent_tasks WHERE status IN ('completed', 'failed', 'cancelled') AND completed_at < ?`,
		before,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old agent tasks: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// CountByStatus returns the count of tasks with the specified status
func (r *AgentTaskRepository) CountByStatus(status string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM agent_tasks WHERE status = ?`, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count agent tasks by status: %w", err)
	}
	return count, nil
}

// CountByUserID returns the count of tasks for a user
func (r *AgentTaskRepository) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM agent_tasks WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count agent tasks by user: %w", err)
	}
	return count, nil
}

// scanTasks scans rows into AgentTask slice
func (r *AgentTaskRepository) scanTasks(rows *sql.Rows) ([]*AgentTask, error) {
	var tasks []*AgentTask

	for rows.Next() {
		task := &AgentTask{}
		var context, agentConfigJSON, metadataJSON, resultJSON, errorStr, callbackURL, callbackDataJSON sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&task.ID, &task.UserID, &task.Prompt, &context, &task.Priority, &task.Status,
			&agentConfigJSON, &metadataJSON, &resultJSON, &errorStr, &callbackURL, &callbackDataJSON,
			&task.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent task: %w", err)
		}

		task.Context = context.String
		task.Error = errorStr.String
		task.CallbackURL = callbackURL.String

		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		if agentConfigJSON.Valid && agentConfigJSON.String != "" {
			if err := json.Unmarshal([]byte(agentConfigJSON.String), &task.AgentConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal agent config: %w", err)
			}
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if resultJSON.Valid && resultJSON.String != "" {
			if err := json.Unmarshal([]byte(resultJSON.String), &task.Result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		if callbackDataJSON.Valid && callbackDataJSON.String != "" {
			if err := json.Unmarshal([]byte(callbackDataJSON.String), &task.CallbackData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
			}
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agent tasks: %w", err)
	}

	return tasks, nil
}
