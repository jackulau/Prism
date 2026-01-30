package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/types"
)

// FileHistory represents a historical version of a file with attribution
type FileHistory struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	FilePath       string            `json:"file_path"`
	Content        string            `json:"content"`
	Operation      string            `json:"operation"` // "create", "update", "delete", "rename"
	CreatedAt      time.Time         `json:"created_at"`
	// Attribution fields
	AgentID        *string           `json:"agent_id,omitempty"`
	AgentName      *string           `json:"agent_name,omitempty"`
	AgentType      *string           `json:"agent_type,omitempty"`
	ToolName       *string           `json:"tool_name,omitempty"`
	ToolSlug       *string           `json:"tool_slug,omitempty"`
	MessageID      *string           `json:"message_id,omitempty"`
	ConversationID *string           `json:"conversation_id,omitempty"`
	WorkflowID     *string           `json:"workflow_id,omitempty"`
	StepID         *string           `json:"step_id,omitempty"`
	Description    *string           `json:"description,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// FileHistoryRepository handles file history database operations
type FileHistoryRepository struct {
	db *sql.DB
}

// NewFileHistoryRepository creates a new file history repository
func NewFileHistoryRepository(db *sql.DB) *FileHistoryRepository {
	return &FileHistoryRepository{db: db}
}

// Create creates a new file history entry
func (r *FileHistoryRepository) Create(userID, filePath, content, operation string) (*FileHistory, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO file_history (id, user_id, file_path, content, operation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, userID, filePath, content, operation, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file history: %w", err)
	}

	return &FileHistory{
		ID:        id,
		UserID:    userID,
		FilePath:  filePath,
		Content:   content,
		Operation: operation,
		CreatedAt: now,
	}, nil
}

// ListByFilePath retrieves file history for a specific file
func (r *FileHistoryRepository) ListByFilePath(userID, filePath string, limit int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ? AND file_path = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		userID, filePath, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// ListByUserID retrieves all file history entries for a user
func (r *FileHistoryRepository) ListByUserID(userID string, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// GetByID retrieves a specific file history entry
func (r *FileHistoryRepository) GetByID(id string) (*FileHistory, error) {
	h := &FileHistory{}
	err := r.db.QueryRow(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	return h, nil
}

// GetLatestByFilePath gets the most recent history entry for a file
func (r *FileHistoryRepository) GetLatestByFilePath(userID, filePath string) (*FileHistory, error) {
	h := &FileHistory{}
	err := r.db.QueryRow(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ? AND file_path = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID, filePath,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest file history: %w", err)
	}

	return h, nil
}

// DeleteOldEntries removes history entries older than the specified duration
// Keeps at least minKeep entries per file
func (r *FileHistoryRepository) DeleteOldEntries(userID string, olderThan time.Duration, minKeep int) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete old entries but keep at least minKeep per file
	_, err := r.db.Exec(`
		DELETE FROM file_history
		WHERE user_id = ?
		AND created_at < ?
		AND id NOT IN (
			SELECT id FROM (
				SELECT id, file_path,
				ROW_NUMBER() OVER (PARTITION BY file_path ORDER BY created_at DESC) as rn
				FROM file_history WHERE user_id = ?
			) WHERE rn <= ?
		)`,
		userID, cutoff, userID, minKeep,
	)
	if err != nil {
		return fmt.Errorf("failed to delete old file history: %w", err)
	}

	return nil
}

// GetDistinctFiles returns all unique file paths with history for a user
func (r *FileHistoryRepository) GetDistinctFiles(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT file_path FROM file_history WHERE user_id = ? ORDER BY file_path`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct files: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("failed to scan file path: %w", err)
		}
		files = append(files, path)
	}

	return files, nil
}

// CreateWithAttribution creates a new file history entry with attribution information
func (r *FileHistoryRepository) CreateWithAttribution(
	userID, filePath, content, operation string,
	attribution *types.AttributionContext,
) (*FileHistory, error) {
	id := uuid.New().String()
	now := time.Now()

	var metadataJSON *string
	if attribution != nil && len(attribution.Metadata) > 0 {
		data, err := json.Marshal(attribution.Metadata)
		if err == nil {
			s := string(data)
			metadataJSON = &s
		}
	}

	var agentID, agentName, agentType, toolName, toolSlug *string
	var messageID, conversationID, workflowID, stepID, description *string

	if attribution != nil {
		if attribution.AgentID != "" {
			agentID = &attribution.AgentID
		}
		if attribution.AgentName != "" {
			agentName = &attribution.AgentName
		}
		if attribution.AgentType != "" {
			agentType = &attribution.AgentType
		}
		if attribution.ToolName != "" {
			toolName = &attribution.ToolName
		}
		if attribution.ToolSlug != "" {
			toolSlug = &attribution.ToolSlug
		}
		if attribution.MessageID != "" {
			messageID = &attribution.MessageID
		}
		if attribution.ConversationID != "" {
			conversationID = &attribution.ConversationID
		}
		if attribution.WorkflowID != "" {
			workflowID = &attribution.WorkflowID
		}
		if attribution.StepID != "" {
			stepID = &attribution.StepID
		}
		if attribution.Description != "" {
			description = &attribution.Description
		}
	}

	_, err := r.db.Exec(
		`INSERT INTO file_history (
			id, user_id, file_path, content, operation, created_at,
			agent_id, agent_name, agent_type, tool_name, tool_slug,
			message_id, conversation_id, workflow_id, step_id, description, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, userID, filePath, content, operation, now,
		agentID, agentName, agentType, toolName, toolSlug,
		messageID, conversationID, workflowID, stepID, description, metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file history with attribution: %w", err)
	}

	h := &FileHistory{
		ID:             id,
		UserID:         userID,
		FilePath:       filePath,
		Content:        content,
		Operation:      operation,
		CreatedAt:      now,
		AgentID:        agentID,
		AgentName:      agentName,
		AgentType:      agentType,
		ToolName:       toolName,
		ToolSlug:       toolSlug,
		MessageID:      messageID,
		ConversationID: conversationID,
		WorkflowID:     workflowID,
		StepID:         stepID,
		Description:    description,
	}

	if attribution != nil && len(attribution.Metadata) > 0 {
		h.Metadata = attribution.Metadata
	}

	return h, nil
}

// ListByAgent retrieves file history for changes made by a specific agent
func (r *FileHistoryRepository) ListByAgent(userID, agentID string, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
			agent_id, agent_name, agent_type, tool_name, tool_slug,
			message_id, conversation_id, workflow_id, step_id, description, metadata
		 FROM file_history
		 WHERE user_id = ? AND agent_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, agentID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by agent: %w", err)
	}
	defer rows.Close()

	return r.scanHistoryRows(rows)
}

// ListByConversation retrieves all file changes made in a specific conversation
func (r *FileHistoryRepository) ListByConversation(userID, conversationID string) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
			agent_id, agent_name, agent_type, tool_name, tool_slug,
			message_id, conversation_id, workflow_id, step_id, description, metadata
		 FROM file_history
		 WHERE user_id = ? AND conversation_id = ?
		 ORDER BY created_at DESC`,
		userID, conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by conversation: %w", err)
	}
	defer rows.Close()

	return r.scanHistoryRows(rows)
}

// ListByWorkflow retrieves all file changes made in a specific workflow
func (r *FileHistoryRepository) ListByWorkflow(userID, workflowID string) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
			agent_id, agent_name, agent_type, tool_name, tool_slug,
			message_id, conversation_id, workflow_id, step_id, description, metadata
		 FROM file_history
		 WHERE user_id = ? AND workflow_id = ?
		 ORDER BY created_at DESC`,
		userID, workflowID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by workflow: %w", err)
	}
	defer rows.Close()

	return r.scanHistoryRows(rows)
}

// ListByTool retrieves all file changes made by a specific tool
func (r *FileHistoryRepository) ListByTool(userID, toolName string, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
			agent_id, agent_name, agent_type, tool_name, tool_slug,
			message_id, conversation_id, workflow_id, step_id, description, metadata
		 FROM file_history
		 WHERE user_id = ? AND tool_name = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, toolName, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by tool: %w", err)
	}
	defer rows.Close()

	return r.scanHistoryRows(rows)
}

// GetAttributionSummary returns aggregated attribution data for a user within a date range
func (r *FileHistoryRepository) GetAttributionSummary(userID string, startDate, endDate time.Time) (*types.AttributionSummary, error) {
	summary := types.NewAttributionSummary()

	// Get total changes
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM file_history WHERE user_id = ? AND created_at BETWEEN ? AND ?`,
		userID, startDate, endDate,
	).Scan(&summary.TotalChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to get total changes: %w", err)
	}

	// Get changes by agent
	rows, err := r.db.Query(
		`SELECT COALESCE(agent_name, agent_id, 'unknown'), COUNT(*) as cnt
		 FROM file_history
		 WHERE user_id = ? AND created_at BETWEEN ? AND ? AND (agent_id IS NOT NULL OR agent_name IS NOT NULL)
		 GROUP BY COALESCE(agent_name, agent_id)
		 ORDER BY cnt DESC`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes by agent: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var agentKey string
		var count int
		if err := rows.Scan(&agentKey, &count); err != nil {
			return nil, fmt.Errorf("failed to scan agent count: %w", err)
		}
		summary.ByAgent[agentKey] = count
		if summary.MostActiveAgent == "" {
			summary.MostActiveAgent = agentKey
		}
	}

	// Get changes by tool
	rows, err = r.db.Query(
		`SELECT COALESCE(tool_name, 'unknown'), COUNT(*) as cnt
		 FROM file_history
		 WHERE user_id = ? AND created_at BETWEEN ? AND ? AND tool_name IS NOT NULL
		 GROUP BY tool_name
		 ORDER BY cnt DESC`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes by tool: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var toolKey string
		var count int
		if err := rows.Scan(&toolKey, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tool count: %w", err)
		}
		summary.ByTool[toolKey] = count
		if summary.MostUsedTool == "" {
			summary.MostUsedTool = toolKey
		}
	}

	// Get changes by operation
	rows, err = r.db.Query(
		`SELECT operation, COUNT(*) as cnt
		 FROM file_history
		 WHERE user_id = ? AND created_at BETWEEN ? AND ?
		 GROUP BY operation`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes by operation: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var op string
		var count int
		if err := rows.Scan(&op, &count); err != nil {
			return nil, fmt.Errorf("failed to scan operation count: %w", err)
		}
		summary.ByOperation[op] = count
	}

	// Get timeline by day
	rows, err = r.db.Query(
		`SELECT DATE(created_at) as day, COUNT(*) as cnt
		 FROM file_history
		 WHERE user_id = ? AND created_at BETWEEN ? AND ?
		 GROUP BY DATE(created_at)
		 ORDER BY day`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, fmt.Errorf("failed to scan timeline: %w", err)
		}
		summary.TimelineByDay[day] = count
	}

	return summary, nil
}

// UpdateAttribution updates attribution fields on an existing history entry
func (r *FileHistoryRepository) UpdateAttribution(historyID string, attribution *types.AttributionContext) error {
	if attribution == nil {
		return nil
	}

	var metadataJSON *string
	if len(attribution.Metadata) > 0 {
		data, err := json.Marshal(attribution.Metadata)
		if err == nil {
			s := string(data)
			metadataJSON = &s
		}
	}

	_, err := r.db.Exec(
		`UPDATE file_history SET
			agent_id = COALESCE(?, agent_id),
			agent_name = COALESCE(?, agent_name),
			agent_type = COALESCE(?, agent_type),
			tool_name = COALESCE(?, tool_name),
			tool_slug = COALESCE(?, tool_slug),
			message_id = COALESCE(?, message_id),
			conversation_id = COALESCE(?, conversation_id),
			workflow_id = COALESCE(?, workflow_id),
			step_id = COALESCE(?, step_id),
			description = COALESCE(?, description),
			metadata = COALESCE(?, metadata)
		 WHERE id = ?`,
		nullIfEmpty(attribution.AgentID),
		nullIfEmpty(attribution.AgentName),
		nullIfEmpty(attribution.AgentType),
		nullIfEmpty(attribution.ToolName),
		nullIfEmpty(attribution.ToolSlug),
		nullIfEmpty(attribution.MessageID),
		nullIfEmpty(attribution.ConversationID),
		nullIfEmpty(attribution.WorkflowID),
		nullIfEmpty(attribution.StepID),
		nullIfEmpty(attribution.Description),
		metadataJSON,
		historyID,
	)
	if err != nil {
		return fmt.Errorf("failed to update attribution: %w", err)
	}

	return nil
}

// scanHistoryRows scans rows into FileHistory structs with attribution fields
func (r *FileHistoryRepository) scanHistoryRows(rows *sql.Rows) ([]*FileHistory, error) {
	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		var metadataJSON sql.NullString
		err := rows.Scan(
			&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
			&h.AgentID, &h.AgentName, &h.AgentType, &h.ToolName, &h.ToolSlug,
			&h.MessageID, &h.ConversationID, &h.WorkflowID, &h.StepID, &h.Description,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			_ = json.Unmarshal([]byte(metadataJSON.String), &h.Metadata)
		}

		history = append(history, h)
	}
	return history, nil
}

// nullIfEmpty returns nil if the string is empty, otherwise returns a pointer to the string
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
