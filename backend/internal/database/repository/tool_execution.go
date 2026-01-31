package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ToolExecution represents a tool execution log entry
type ToolExecution struct {
	ID              string
	MessageID       string
	ToolName        string
	Parameters      string
	Result          string
	Status          string
	ExecutionTimeMS int64
	ContainerID     string
	CreatedAt       time.Time
}

// ToolExecutionFilter holds filter options for listing executions
type ToolExecutionFilter struct {
	ToolName string
	Status   string
	Limit    int
	Offset   int
}

// ToolExecutionRepository handles tool execution database operations
type ToolExecutionRepository struct {
	db *sql.DB
}

// NewToolExecutionRepository creates a new tool execution repository
func NewToolExecutionRepository(db *sql.DB) *ToolExecutionRepository {
	return &ToolExecutionRepository{db: db}
}

// Create creates a new tool execution record
func (r *ToolExecutionRepository) Create(exec *ToolExecution) error {
	if exec.ID == "" {
		exec.ID = uuid.New().String()
	}
	if exec.CreatedAt.IsZero() {
		exec.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(`
		INSERT INTO tool_executions (id, message_id, tool_name, parameters, result, status, execution_time_ms, container_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, exec.MessageID, exec.ToolName, exec.Parameters, exec.Result, exec.Status, exec.ExecutionTimeMS, exec.ContainerID, exec.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tool execution: %w", err)
	}
	return nil
}

// GetByID retrieves a tool execution by ID
func (r *ToolExecutionRepository) GetByID(id string) (*ToolExecution, error) {
	exec := &ToolExecution{}
	var result, containerID sql.NullString
	var executionTimeMS sql.NullInt64

	err := r.db.QueryRow(`
		SELECT id, message_id, tool_name, parameters, result, status, execution_time_ms, container_id, created_at
		FROM tool_executions
		WHERE id = ?
	`, id).Scan(&exec.ID, &exec.MessageID, &exec.ToolName, &exec.Parameters, &result, &exec.Status, &executionTimeMS, &containerID, &exec.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tool execution: %w", err)
	}

	exec.Result = result.String
	exec.ContainerID = containerID.String
	exec.ExecutionTimeMS = executionTimeMS.Int64

	return exec, nil
}

// List retrieves tool executions with optional filtering
func (r *ToolExecutionRepository) List(filter ToolExecutionFilter) ([]*ToolExecution, error) {
	query := `
		SELECT id, message_id, tool_name, parameters, result, status, execution_time_ms, container_id, created_at
		FROM tool_executions
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.ToolName != "" {
		query += " AND tool_name = ?"
		args = append(args, filter.ToolName)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	} else {
		query += " LIMIT 100"
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tool executions: %w", err)
	}
	defer rows.Close()

	return r.scanExecutions(rows)
}

// ListByMessageID retrieves all tool executions for a specific message
func (r *ToolExecutionRepository) ListByMessageID(messageID string) ([]*ToolExecution, error) {
	rows, err := r.db.Query(`
		SELECT id, message_id, tool_name, parameters, result, status, execution_time_ms, container_id, created_at
		FROM tool_executions
		WHERE message_id = ?
		ORDER BY created_at ASC
	`, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tool executions by message: %w", err)
	}
	defer rows.Close()

	return r.scanExecutions(rows)
}

// Update updates an existing tool execution
func (r *ToolExecutionRepository) Update(exec *ToolExecution) error {
	result, err := r.db.Exec(`
		UPDATE tool_executions
		SET result = ?, status = ?, execution_time_ms = ?, container_id = ?
		WHERE id = ?
	`, exec.Result, exec.Status, exec.ExecutionTimeMS, exec.ContainerID, exec.ID)

	if err != nil {
		return fmt.Errorf("failed to update tool execution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool execution not found")
	}

	return nil
}

// Count returns the total count of executions matching the filter
func (r *ToolExecutionRepository) Count(filter ToolExecutionFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM tool_executions WHERE 1=1"
	args := []interface{}{}

	if filter.ToolName != "" {
		query += " AND tool_name = ?"
		args = append(args, filter.ToolName)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	var count int64
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tool executions: %w", err)
	}

	return count, nil
}

// GetDistinctToolNames returns all distinct tool names from executions
func (r *ToolExecutionRepository) GetDistinctToolNames() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT tool_name
		FROM tool_executions
		ORDER BY tool_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct tool names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan tool name: %w", err)
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tool names: %w", err)
	}

	return names, nil
}

// scanExecutions scans rows into a slice of ToolExecution pointers
func (r *ToolExecutionRepository) scanExecutions(rows *sql.Rows) ([]*ToolExecution, error) {
	var executions []*ToolExecution

	for rows.Next() {
		exec := &ToolExecution{}
		var result, containerID sql.NullString
		var executionTimeMS sql.NullInt64

		err := rows.Scan(&exec.ID, &exec.MessageID, &exec.ToolName, &exec.Parameters, &result, &exec.Status, &executionTimeMS, &containerID, &exec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool execution: %w", err)
		}

		exec.Result = result.String
		exec.ContainerID = containerID.String
		exec.ExecutionTimeMS = executionTimeMS.Int64

		executions = append(executions, exec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tool executions: %w", err)
	}

	return executions, nil
}
