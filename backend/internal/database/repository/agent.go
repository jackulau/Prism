package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentRecord represents an agent record in the database
type AgentRecord struct {
	ID             string
	UserID         string
	ConversationID *string
	Name           string
	Description    string
	Provider       string
	Model          string
	SystemPrompt   string
	Status         string // pending, running, completed, failed, cancelled
	ConfigJSON     string
	Error          string
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
}

// AgentResultRecord represents an agent result record in the database
type AgentResultRecord struct {
	ID           string
	AgentID      string
	TaskID       string
	Success      bool
	Output       string
	Error        string
	UsageJSON    string
	MetadataJSON string
	DurationMS   int64
	CreatedAt    time.Time
}

// AgentRepository handles agent database operations
type AgentRepository struct {
	db *sql.DB
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create creates a new agent record
func (r *AgentRepository) Create(agent *AgentRecord) error {
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}
	if agent.Status == "" {
		agent.Status = "pending"
	}

	var conversationID sql.NullString
	if agent.ConversationID != nil && *agent.ConversationID != "" {
		conversationID = sql.NullString{String: *agent.ConversationID, Valid: true}
	}

	var startedAt sql.NullTime
	if agent.StartedAt != nil {
		startedAt = sql.NullTime{Time: *agent.StartedAt, Valid: true}
	}

	var completedAt sql.NullTime
	if agent.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *agent.CompletedAt, Valid: true}
	}

	_, err := r.db.Exec(
		`INSERT INTO agents (id, user_id, conversation_id, name, description, provider, model, system_prompt, status, config_json, error, created_at, started_at, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		agent.ID, agent.UserID, conversationID, agent.Name, agent.Description, agent.Provider, agent.Model,
		agent.SystemPrompt, agent.Status, agent.ConfigJSON, agent.Error, agent.CreatedAt, startedAt, completedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetByID retrieves an agent by ID
func (r *AgentRepository) GetByID(id string) (*AgentRecord, error) {
	agent := &AgentRecord{}
	var conversationID, description, systemPrompt, configJSON, errStr sql.NullString
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, conversation_id, name, description, provider, model, system_prompt, status, config_json, error, created_at, started_at, completed_at
		 FROM agents WHERE id = ?`,
		id,
	).Scan(&agent.ID, &agent.UserID, &conversationID, &agent.Name, &description, &agent.Provider, &agent.Model,
		&systemPrompt, &agent.Status, &configJSON, &errStr, &agent.CreatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if conversationID.Valid {
		agent.ConversationID = &conversationID.String
	}
	agent.Description = description.String
	agent.SystemPrompt = systemPrompt.String
	agent.ConfigJSON = configJSON.String
	agent.Error = errStr.String
	if startedAt.Valid {
		agent.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		agent.CompletedAt = &completedAt.Time
	}

	return agent, nil
}

// GetByUserID retrieves all agents for a user with pagination
func (r *AgentRepository) GetByUserID(userID string, limit, offset int) ([]*AgentRecord, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, conversation_id, name, description, provider, model, system_prompt, status, config_json, error, created_at, started_at, completed_at
		 FROM agents WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*AgentRecord
	for rows.Next() {
		agent := &AgentRecord{}
		var conversationID, description, systemPrompt, configJSON, errStr sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&agent.ID, &agent.UserID, &conversationID, &agent.Name, &description, &agent.Provider, &agent.Model,
			&systemPrompt, &agent.Status, &configJSON, &errStr, &agent.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		if conversationID.Valid {
			agent.ConversationID = &conversationID.String
		}
		agent.Description = description.String
		agent.SystemPrompt = systemPrompt.String
		agent.ConfigJSON = configJSON.String
		agent.Error = errStr.String
		if startedAt.Valid {
			agent.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			agent.CompletedAt = &completedAt.Time
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// GetByConversationID retrieves all agents for a conversation
func (r *AgentRepository) GetByConversationID(conversationID string) ([]*AgentRecord, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, conversation_id, name, description, provider, model, system_prompt, status, config_json, error, created_at, started_at, completed_at
		 FROM agents WHERE conversation_id = ? ORDER BY created_at DESC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents by conversation: %w", err)
	}
	defer rows.Close()

	var agents []*AgentRecord
	for rows.Next() {
		agent := &AgentRecord{}
		var convID, description, systemPrompt, configJSON, errStr sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&agent.ID, &agent.UserID, &convID, &agent.Name, &description, &agent.Provider, &agent.Model,
			&systemPrompt, &agent.Status, &configJSON, &errStr, &agent.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		if convID.Valid {
			agent.ConversationID = &convID.String
		}
		agent.Description = description.String
		agent.SystemPrompt = systemPrompt.String
		agent.ConfigJSON = configJSON.String
		agent.Error = errStr.String
		if startedAt.Valid {
			agent.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			agent.CompletedAt = &completedAt.Time
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateStatus updates the status of an agent
func (r *AgentRepository) UpdateStatus(id, status string) error {
	var startedAt sql.NullTime
	if status == "running" {
		startedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := r.db.Exec(
		`UPDATE agents SET status = ?, started_at = COALESCE(?, started_at) WHERE id = ?`,
		status, startedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	return nil
}

// UpdateError updates the error field of an agent and sets status to failed
func (r *AgentRepository) UpdateError(id, errorMsg string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agents SET status = 'failed', error = ?, completed_at = ? WHERE id = ?`,
		errorMsg, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent error: %w", err)
	}
	return nil
}

// Complete marks an agent as completed
func (r *AgentRepository) Complete(id string, completedAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE agents SET status = 'completed', completed_at = ? WHERE id = ?`,
		completedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to complete agent: %w", err)
	}
	return nil
}

// Cancel marks an agent as cancelled
func (r *AgentRepository) Cancel(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agents SET status = 'cancelled', completed_at = ? WHERE id = ?`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel agent: %w", err)
	}
	return nil
}

// Delete deletes an agent record
func (r *AgentRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

// SaveResult saves an agent execution result
func (r *AgentRepository) SaveResult(result *AgentResultRecord) error {
	if result.ID == "" {
		result.ID = uuid.New().String()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(
		`INSERT INTO agent_results (id, agent_id, task_id, success, output, error, usage_json, metadata_json, duration_ms, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID, result.AgentID, result.TaskID, result.Success, result.Output, result.Error,
		result.UsageJSON, result.MetadataJSON, result.DurationMS, result.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save agent result: %w", err)
	}

	return nil
}

// GetResults retrieves all results for an agent
func (r *AgentRepository) GetResults(agentID string) ([]*AgentResultRecord, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, task_id, success, output, error, usage_json, metadata_json, duration_ms, created_at
		 FROM agent_results WHERE agent_id = ? ORDER BY created_at DESC`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent results: %w", err)
	}
	defer rows.Close()

	var results []*AgentResultRecord
	for rows.Next() {
		result := &AgentResultRecord{}
		var taskID, output, errStr, usageJSON, metadataJSON sql.NullString
		var durationMS sql.NullInt64

		err := rows.Scan(&result.ID, &result.AgentID, &taskID, &result.Success, &output, &errStr,
			&usageJSON, &metadataJSON, &durationMS, &result.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent result: %w", err)
		}

		result.TaskID = taskID.String
		result.Output = output.String
		result.Error = errStr.String
		result.UsageJSON = usageJSON.String
		result.MetadataJSON = metadataJSON.String
		result.DurationMS = durationMS.Int64

		results = append(results, result)
	}

	return results, nil
}

// Count returns the total count of agents for a user
func (r *AgentRepository) Count(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM agents WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count agents: %w", err)
	}
	return count, nil
}

// AgentUsage represents token usage information
type AgentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// MarshalUsage marshals usage to JSON
func MarshalUsage(usage *AgentUsage) string {
	if usage == nil {
		return ""
	}
	data, _ := json.Marshal(usage)
	return string(data)
}

// UnmarshalUsage unmarshals usage from JSON
func UnmarshalUsage(data string) *AgentUsage {
	if data == "" {
		return nil
	}
	var usage AgentUsage
	if err := json.Unmarshal([]byte(data), &usage); err != nil {
		return nil
	}
	return &usage
}
