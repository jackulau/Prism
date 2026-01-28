package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentExecution represents an agent execution record
type AgentExecution struct {
	ID               string
	UserID           string
	Provider         string
	LLMProvider      string
	Model            string
	AgentName        string
	Status           string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	InputCost        float64
	OutputCost       float64
	TotalCost        float64
	Currency         string
	Error            string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	Metadata         map[string]interface{}
}

// AgentMessage represents a message within an agent execution
type AgentMessage struct {
	ID               string
	ExecutionID      string
	Role             string
	Content          string
	ToolCalls        []ToolCall
	ToolCallID       string
	PromptTokens     int
	CompletionTokens int
	CreatedAt        time.Time
}

// AgentToolCall represents a tool call within an agent execution
type AgentToolCall struct {
	ID          string
	ExecutionID string
	MessageID   string
	ToolName    string
	Parameters  map[string]interface{}
	Output      string
	Error       string
	Status      string
	DurationMs  int64
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// AgentExecutionRepository handles agent execution database operations
type AgentExecutionRepository struct {
	db *sql.DB
}

// NewAgentExecutionRepository creates a new agent execution repository
func NewAgentExecutionRepository(db *sql.DB) *AgentExecutionRepository {
	return &AgentExecutionRepository{db: db}
}

// Create creates a new agent execution record
func (r *AgentExecutionRepository) Create(userID, provider, llmProvider, model, agentName string, metadata map[string]interface{}) (*AgentExecution, error) {
	id := uuid.New().String()
	now := time.Now()

	var metadataJSON sql.NullString
	if len(metadata) > 0 {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = sql.NullString{String: string(data), Valid: true}
	}

	_, err := r.db.Exec(
		`INSERT INTO agent_executions (id, user_id, provider, llm_provider, model, agent_name, status, created_at, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?)`,
		id, userID, provider, llmProvider, model, agentName, now, metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent execution: %w", err)
	}

	return &AgentExecution{
		ID:          id,
		UserID:      userID,
		Provider:    provider,
		LLMProvider: llmProvider,
		Model:       model,
		AgentName:   agentName,
		Status:      "pending",
		Currency:    "USD",
		CreatedAt:   now,
		Metadata:    metadata,
	}, nil
}

// GetByID retrieves an agent execution by ID
func (r *AgentExecutionRepository) GetByID(id string) (*AgentExecution, error) {
	exec := &AgentExecution{}
	var agentName, errorStr, metadataJSON sql.NullString
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, provider, llm_provider, model, agent_name, status,
		        prompt_tokens, completion_tokens, total_tokens,
		        input_cost, output_cost, total_cost, currency,
		        error, started_at, completed_at, created_at, metadata
		 FROM agent_executions WHERE id = ?`,
		id,
	).Scan(
		&exec.ID, &exec.UserID, &exec.Provider, &exec.LLMProvider, &exec.Model, &agentName, &exec.Status,
		&exec.PromptTokens, &exec.CompletionTokens, &exec.TotalTokens,
		&exec.InputCost, &exec.OutputCost, &exec.TotalCost, &exec.Currency,
		&errorStr, &startedAt, &completedAt, &exec.CreatedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent execution: %w", err)
	}

	exec.AgentName = agentName.String
	exec.Error = errorStr.String
	if startedAt.Valid {
		exec.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		exec.CompletedAt = &completedAt.Time
	}

	if metadataJSON.Valid {
		if err := json.Unmarshal([]byte(metadataJSON.String), &exec.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return exec, nil
}

// ListByUserID retrieves all agent executions for a user
func (r *AgentExecutionRepository) ListByUserID(userID string, limit, offset int) ([]*AgentExecution, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, provider, llm_provider, model, agent_name, status,
		        prompt_tokens, completion_tokens, total_tokens,
		        input_cost, output_cost, total_cost, currency,
		        error, started_at, completed_at, created_at, metadata
		 FROM agent_executions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent executions: %w", err)
	}
	defer rows.Close()

	var executions []*AgentExecution
	for rows.Next() {
		exec := &AgentExecution{}
		var agentName, errorStr, metadataJSON sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&exec.ID, &exec.UserID, &exec.Provider, &exec.LLMProvider, &exec.Model, &agentName, &exec.Status,
			&exec.PromptTokens, &exec.CompletionTokens, &exec.TotalTokens,
			&exec.InputCost, &exec.OutputCost, &exec.TotalCost, &exec.Currency,
			&errorStr, &startedAt, &completedAt, &exec.CreatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent execution: %w", err)
		}

		exec.AgentName = agentName.String
		exec.Error = errorStr.String
		if startedAt.Valid {
			exec.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			exec.CompletedAt = &completedAt.Time
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &exec.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// UpdateStatus updates the status of an agent execution
func (r *AgentExecutionRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(
		`UPDATE agent_executions SET status = ? WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent execution status: %w", err)
	}
	return nil
}

// Start marks an agent execution as started
func (r *AgentExecutionRepository) Start(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_executions SET status = 'running', started_at = ? WHERE id = ?`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to start agent execution: %w", err)
	}
	return nil
}

// Complete marks an agent execution as completed with usage and cost data
func (r *AgentExecutionRepository) Complete(id string, promptTokens, completionTokens, totalTokens int, inputCost, outputCost, totalCost float64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_executions SET
		        status = 'completed',
		        prompt_tokens = ?, completion_tokens = ?, total_tokens = ?,
		        input_cost = ?, output_cost = ?, total_cost = ?,
		        completed_at = ?
		 WHERE id = ?`,
		promptTokens, completionTokens, totalTokens,
		inputCost, outputCost, totalCost,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to complete agent execution: %w", err)
	}
	return nil
}

// Fail marks an agent execution as failed
func (r *AgentExecutionRepository) Fail(id, errorMsg string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_executions SET status = 'failed', error = ?, completed_at = ? WHERE id = ?`,
		errorMsg, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to fail agent execution: %w", err)
	}
	return nil
}

// Cancel marks an agent execution as cancelled
func (r *AgentExecutionRepository) Cancel(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_executions SET status = 'cancelled', completed_at = ? WHERE id = ?`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel agent execution: %w", err)
	}
	return nil
}

// AddUsage adds token usage to an agent execution
func (r *AgentExecutionRepository) AddUsage(id string, promptTokens, completionTokens int, inputCost, outputCost float64) error {
	_, err := r.db.Exec(
		`UPDATE agent_executions SET
		        prompt_tokens = prompt_tokens + ?,
		        completion_tokens = completion_tokens + ?,
		        total_tokens = total_tokens + ?,
		        input_cost = input_cost + ?,
		        output_cost = output_cost + ?,
		        total_cost = total_cost + ?
		 WHERE id = ?`,
		promptTokens, completionTokens, promptTokens+completionTokens,
		inputCost, outputCost, inputCost+outputCost,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to add usage to agent execution: %w", err)
	}
	return nil
}

// GetUsageStats returns aggregated usage statistics for a user
func (r *AgentExecutionRepository) GetUsageStats(userID string, since time.Time) (*UsageStats, error) {
	stats := &UsageStats{}

	err := r.db.QueryRow(
		`SELECT
		        COUNT(*) as total_executions,
		        SUM(prompt_tokens) as total_prompt_tokens,
		        SUM(completion_tokens) as total_completion_tokens,
		        SUM(total_tokens) as total_tokens,
		        SUM(total_cost) as total_cost
		 FROM agent_executions
		 WHERE user_id = ? AND created_at >= ?`,
		userID, since,
	).Scan(&stats.TotalExecutions, &stats.TotalPromptTokens, &stats.TotalCompletionTokens, &stats.TotalTokens, &stats.TotalCost)

	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	return stats, nil
}

// UsageStats contains aggregated usage statistics
type UsageStats struct {
	TotalExecutions       int
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalTokens           int
	TotalCost             float64
}

// Delete deletes an agent execution
func (r *AgentExecutionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM agent_executions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent execution: %w", err)
	}
	return nil
}

// AgentMessageRepository handles agent message database operations
type AgentMessageRepository struct {
	db *sql.DB
}

// NewAgentMessageRepository creates a new agent message repository
func NewAgentMessageRepository(db *sql.DB) *AgentMessageRepository {
	return &AgentMessageRepository{db: db}
}

// Create creates a new agent message
func (r *AgentMessageRepository) Create(executionID, role, content string, toolCalls []ToolCall, toolCallID string, promptTokens, completionTokens int) (*AgentMessage, error) {
	id := uuid.New().String()
	now := time.Now()

	var toolCallsJSON sql.NullString
	if len(toolCalls) > 0 {
		data, err := json.Marshal(toolCalls)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool calls: %w", err)
		}
		toolCallsJSON = sql.NullString{String: string(data), Valid: true}
	}

	var toolCallIDNull sql.NullString
	if toolCallID != "" {
		toolCallIDNull = sql.NullString{String: toolCallID, Valid: true}
	}

	_, err := r.db.Exec(
		`INSERT INTO agent_messages (id, execution_id, role, content, tool_calls, tool_call_id, prompt_tokens, completion_tokens, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, executionID, role, content, toolCallsJSON, toolCallIDNull, promptTokens, completionTokens, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent message: %w", err)
	}

	return &AgentMessage{
		ID:               id,
		ExecutionID:      executionID,
		Role:             role,
		Content:          content,
		ToolCalls:        toolCalls,
		ToolCallID:       toolCallID,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CreatedAt:        now,
	}, nil
}

// ListByExecutionID retrieves all messages for an agent execution
func (r *AgentMessageRepository) ListByExecutionID(executionID string) ([]*AgentMessage, error) {
	rows, err := r.db.Query(
		`SELECT id, execution_id, role, content, tool_calls, tool_call_id, prompt_tokens, completion_tokens, created_at
		 FROM agent_messages WHERE execution_id = ? ORDER BY created_at ASC`,
		executionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent messages: %w", err)
	}
	defer rows.Close()

	var messages []*AgentMessage
	for rows.Next() {
		msg := &AgentMessage{}
		var toolCallsJSON, toolCallID sql.NullString

		err := rows.Scan(&msg.ID, &msg.ExecutionID, &msg.Role, &msg.Content, &toolCallsJSON, &toolCallID, &msg.PromptTokens, &msg.CompletionTokens, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent message: %w", err)
		}

		if toolCallsJSON.Valid {
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		msg.ToolCallID = toolCallID.String
		messages = append(messages, msg)
	}

	return messages, nil
}

// AgentToolCallRepository handles agent tool call database operations
type AgentToolCallRepository struct {
	db *sql.DB
}

// NewAgentToolCallRepository creates a new agent tool call repository
func NewAgentToolCallRepository(db *sql.DB) *AgentToolCallRepository {
	return &AgentToolCallRepository{db: db}
}

// Create creates a new agent tool call record
func (r *AgentToolCallRepository) Create(executionID, messageID, toolName string, parameters map[string]interface{}) (*AgentToolCall, error) {
	id := uuid.New().String()
	now := time.Now()

	paramsJSON, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var messageIDNull sql.NullString
	if messageID != "" {
		messageIDNull = sql.NullString{String: messageID, Valid: true}
	}

	_, err = r.db.Exec(
		`INSERT INTO agent_tool_calls (id, execution_id, message_id, tool_name, parameters, status, created_at)
		 VALUES (?, ?, ?, ?, ?, 'pending', ?)`,
		id, executionID, messageIDNull, toolName, string(paramsJSON), now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent tool call: %w", err)
	}

	return &AgentToolCall{
		ID:          id,
		ExecutionID: executionID,
		MessageID:   messageID,
		ToolName:    toolName,
		Parameters:  parameters,
		Status:      "pending",
		CreatedAt:   now,
	}, nil
}

// Complete marks a tool call as completed
func (r *AgentToolCallRepository) Complete(id, output string, durationMs int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_tool_calls SET status = 'completed', output = ?, duration_ms = ?, completed_at = ? WHERE id = ?`,
		output, durationMs, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to complete agent tool call: %w", err)
	}
	return nil
}

// Fail marks a tool call as failed
func (r *AgentToolCallRepository) Fail(id, errorMsg string, durationMs int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_tool_calls SET status = 'failed', error = ?, duration_ms = ?, completed_at = ? WHERE id = ?`,
		errorMsg, durationMs, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to fail agent tool call: %w", err)
	}
	return nil
}

// ListByExecutionID retrieves all tool calls for an agent execution
func (r *AgentToolCallRepository) ListByExecutionID(executionID string) ([]*AgentToolCall, error) {
	rows, err := r.db.Query(
		`SELECT id, execution_id, message_id, tool_name, parameters, output, error, status, duration_ms, created_at, completed_at
		 FROM agent_tool_calls WHERE execution_id = ? ORDER BY created_at ASC`,
		executionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent tool calls: %w", err)
	}
	defer rows.Close()

	var toolCalls []*AgentToolCall
	for rows.Next() {
		tc := &AgentToolCall{}
		var messageID, output, errorStr sql.NullString
		var durationMs sql.NullInt64
		var completedAt sql.NullTime
		var paramsJSON string

		err := rows.Scan(&tc.ID, &tc.ExecutionID, &messageID, &tc.ToolName, &paramsJSON, &output, &errorStr, &tc.Status, &durationMs, &tc.CreatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent tool call: %w", err)
		}

		tc.MessageID = messageID.String
		tc.Output = output.String
		tc.Error = errorStr.String
		tc.DurationMs = durationMs.Int64
		if completedAt.Valid {
			tc.CompletedAt = &completedAt.Time
		}

		if err := json.Unmarshal([]byte(paramsJSON), &tc.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		toolCalls = append(toolCalls, tc)
	}

	return toolCalls, nil
}
