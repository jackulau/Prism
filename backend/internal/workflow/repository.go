package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentExecutionRepository handles agent execution database operations
type AgentExecutionRepository struct {
	db *sql.DB
}

// NewAgentExecutionRepository creates a new agent execution repository
func NewAgentExecutionRepository(db *sql.DB) *AgentExecutionRepository {
	return &AgentExecutionRepository{db: db}
}

// Create creates a new agent execution record
func (r *AgentExecutionRepository) Create(ctx context.Context, agentID, userID, conversationID string) (*AgentExecution, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_executions (id, agent_id, user_id, conversation_id, status, started_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, agentID, userID, conversationID, "pending", now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent execution: %w", err)
	}

	return &AgentExecution{
		ID:             id,
		AgentID:        agentID,
		UserID:         userID,
		ConversationID: conversationID,
		Status:         "pending",
		StartedAt:      &now,
	}, nil
}

// GetByID retrieves an agent execution by ID
func (r *AgentExecutionRepository) GetByID(ctx context.Context, id string) (*AgentExecution, error) {
	exec := &AgentExecution{}
	var conversationID, currentStep, error_, branchName, commitSHA sql.NullString
	var startedAt, completedAt sql.NullTime
	var iterations sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, user_id, conversation_id, status, current_step,
		        started_at, completed_at, error, branch_name, commit_sha, iterations
		 FROM agent_executions WHERE id = ?`,
		id,
	).Scan(&exec.ID, &exec.AgentID, &exec.UserID, &conversationID, &exec.Status, &currentStep,
		&startedAt, &completedAt, &error_, &branchName, &commitSHA, &iterations)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent execution: %w", err)
	}

	exec.ConversationID = conversationID.String
	exec.CurrentStep = currentStep.String
	exec.Error = error_.String
	exec.BranchName = branchName.String
	exec.CommitSHA = commitSHA.String
	exec.Iterations = int(iterations.Int64)
	if startedAt.Valid {
		exec.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		exec.CompletedAt = &completedAt.Time
	}

	return exec, nil
}

// UpdateStatus updates the status of an agent execution
func (r *AgentExecutionRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE agent_executions SET status = ? WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent execution status: %w", err)
	}
	return nil
}

// UpdateStep updates the current step of an agent execution
func (r *AgentExecutionRepository) UpdateStep(ctx context.Context, id string, step WorkflowStep) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE agent_executions SET current_step = ?, status = ? WHERE id = ?`,
		string(step), "running", id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent execution step: %w", err)
	}
	return nil
}

// Complete marks an agent execution as complete
func (r *AgentExecutionRepository) Complete(ctx context.Context, id, commitSHA string, iterations int) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE agent_executions
		 SET status = ?, completed_at = ?, commit_sha = ?, iterations = ?
		 WHERE id = ?`,
		"completed", now, commitSHA, iterations, id,
	)
	if err != nil {
		return fmt.Errorf("failed to complete agent execution: %w", err)
	}
	return nil
}

// Fail marks an agent execution as failed
func (r *AgentExecutionRepository) Fail(ctx context.Context, id, errorMsg string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE agent_executions SET status = ?, completed_at = ?, error = ? WHERE id = ?`,
		"failed", now, errorMsg, id,
	)
	if err != nil {
		return fmt.Errorf("failed to fail agent execution: %w", err)
	}
	return nil
}

// UpdateBranch updates the branch name of an agent execution
func (r *AgentExecutionRepository) UpdateBranch(ctx context.Context, id, branchName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE agent_executions SET branch_name = ? WHERE id = ?`,
		branchName, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent execution branch: %w", err)
	}
	return nil
}

// ListByAgentID retrieves all executions for an agent
func (r *AgentExecutionRepository) ListByAgentID(ctx context.Context, agentID string, limit, offset int) ([]*AgentExecution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, user_id, conversation_id, status, current_step,
		        started_at, completed_at, error, branch_name, commit_sha, iterations
		 FROM agent_executions WHERE agent_id = ? ORDER BY started_at DESC LIMIT ? OFFSET ?`,
		agentID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent executions: %w", err)
	}
	defer rows.Close()

	var executions []*AgentExecution
	for rows.Next() {
		exec := &AgentExecution{}
		var conversationID, currentStep, error_, branchName, commitSHA sql.NullString
		var startedAt, completedAt sql.NullTime
		var iterations sql.NullInt64

		err := rows.Scan(&exec.ID, &exec.AgentID, &exec.UserID, &conversationID, &exec.Status, &currentStep,
			&startedAt, &completedAt, &error_, &branchName, &commitSHA, &iterations)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent execution: %w", err)
		}

		exec.ConversationID = conversationID.String
		exec.CurrentStep = currentStep.String
		exec.Error = error_.String
		exec.BranchName = branchName.String
		exec.CommitSHA = commitSHA.String
		exec.Iterations = int(iterations.Int64)
		if startedAt.Valid {
			exec.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			exec.CompletedAt = &completedAt.Time
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// ListByUserID retrieves all executions for a user
func (r *AgentExecutionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*AgentExecution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, user_id, conversation_id, status, current_step,
		        started_at, completed_at, error, branch_name, commit_sha, iterations
		 FROM agent_executions WHERE user_id = ? ORDER BY started_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent executions: %w", err)
	}
	defer rows.Close()

	var executions []*AgentExecution
	for rows.Next() {
		exec := &AgentExecution{}
		var conversationID, currentStep, error_, branchName, commitSHA sql.NullString
		var startedAt, completedAt sql.NullTime
		var iterations sql.NullInt64

		err := rows.Scan(&exec.ID, &exec.AgentID, &exec.UserID, &conversationID, &exec.Status, &currentStep,
			&startedAt, &completedAt, &error_, &branchName, &commitSHA, &iterations)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent execution: %w", err)
		}

		exec.ConversationID = conversationID.String
		exec.CurrentStep = currentStep.String
		exec.Error = error_.String
		exec.BranchName = branchName.String
		exec.CommitSHA = commitSHA.String
		exec.Iterations = int(iterations.Int64)
		if startedAt.Valid {
			exec.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			exec.CompletedAt = &completedAt.Time
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// TokenUsageRepository handles token usage database operations
type TokenUsageRepository struct {
	db *sql.DB
}

// NewTokenUsageRepository creates a new token usage repository
func NewTokenUsageRepository(db *sql.DB) *TokenUsageRepository {
	return &TokenUsageRepository{db: db}
}

// Create creates a new token usage record
func (r *TokenUsageRepository) Create(ctx context.Context, executionID, userID, provider, model string, usage *TokenUsageRecord) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO token_usage (id, execution_id, user_id, provider, model,
		                          prompt_tokens, completion_tokens, total_tokens, cost_usd, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, executionID, userID, provider, model,
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.CostUSD, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create token usage record: %w", err)
	}
	return nil
}

// GetByExecutionID retrieves token usage for an execution
func (r *TokenUsageRepository) GetByExecutionID(ctx context.Context, executionID string) ([]*TokenUsageRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, execution_id, user_id, provider, model,
		        prompt_tokens, completion_tokens, total_tokens, cost_usd, created_at
		 FROM token_usage WHERE execution_id = ? ORDER BY created_at ASC`,
		executionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage: %w", err)
	}
	defer rows.Close()

	var records []*TokenUsageRecord
	for rows.Next() {
		rec := &TokenUsageRecord{}
		err := rows.Scan(&rec.ID, &rec.ExecutionID, &rec.UserID, &rec.Provider, &rec.Model,
			&rec.PromptTokens, &rec.CompletionTokens, &rec.TotalTokens, &rec.CostUSD, &rec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

// GetTotalByUserID retrieves total token usage for a user
func (r *TokenUsageRepository) GetTotalByUserID(ctx context.Context, userID string) (*TokenUsageRecord, error) {
	rec := &TokenUsageRecord{UserID: userID}

	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(prompt_tokens), 0), COALESCE(SUM(completion_tokens), 0),
		        COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost_usd), 0)
		 FROM token_usage WHERE user_id = ?`,
		userID,
	).Scan(&rec.PromptTokens, &rec.CompletionTokens, &rec.TotalTokens, &rec.CostUSD)

	if err != nil {
		return nil, fmt.Errorf("failed to get total token usage: %w", err)
	}

	return rec, nil
}

// GetUsageByDateRange retrieves token usage for a user within a date range
func (r *TokenUsageRepository) GetUsageByDateRange(ctx context.Context, userID string, start, end time.Time) ([]*TokenUsageRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, execution_id, user_id, provider, model,
		        prompt_tokens, completion_tokens, total_tokens, cost_usd, created_at
		 FROM token_usage WHERE user_id = ? AND created_at >= ? AND created_at <= ?
		 ORDER BY created_at ASC`,
		userID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage by date range: %w", err)
	}
	defer rows.Close()

	var records []*TokenUsageRecord
	for rows.Next() {
		rec := &TokenUsageRecord{}
		err := rows.Scan(&rec.ID, &rec.ExecutionID, &rec.UserID, &rec.Provider, &rec.Model,
			&rec.PromptTokens, &rec.CompletionTokens, &rec.TotalTokens, &rec.CostUSD, &rec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}
