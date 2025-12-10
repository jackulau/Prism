package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID           string
	UserID       string
	Title        string
	Provider     string
	Model        string
	SystemPrompt string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Message represents a chat message
type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	ToolCalls      []ToolCall
	ToolCallID     string
	TokensUsed     int
	CreatedAt      time.Time
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ConversationRepository handles conversation database operations
type ConversationRepository struct {
	db *sql.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create creates a new conversation
func (r *ConversationRepository) Create(userID, provider, model, systemPrompt string) (*Conversation, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO conversations (id, user_id, provider, model, system_prompt, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, provider, model, systemPrompt, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &Conversation{
		ID:           id,
		UserID:       userID,
		Provider:     provider,
		Model:        model,
		SystemPrompt: systemPrompt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepository) GetByID(id string) (*Conversation, error) {
	conv := &Conversation{}
	var title, systemPrompt sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, title, provider, model, system_prompt, created_at, updated_at
		 FROM conversations WHERE id = ?`,
		id,
	).Scan(&conv.ID, &conv.UserID, &title, &conv.Provider, &conv.Model, &systemPrompt, &conv.CreatedAt, &conv.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	conv.Title = title.String
	conv.SystemPrompt = systemPrompt.String

	return conv, nil
}

// ListByUserID retrieves all conversations for a user
func (r *ConversationRepository) ListByUserID(userID string, limit, offset int) ([]*Conversation, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, provider, model, system_prompt, created_at, updated_at
		 FROM conversations WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		var title, systemPrompt sql.NullString

		err := rows.Scan(&conv.ID, &conv.UserID, &title, &conv.Provider, &conv.Model, &systemPrompt, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		conv.Title = title.String
		conv.SystemPrompt = systemPrompt.String
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// Update updates a conversation
func (r *ConversationRepository) Update(id, title string) error {
	_, err := r.db.Exec(
		`UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?`,
		title, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// Delete deletes a conversation
func (r *ConversationRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM conversations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// Search searches conversations by title or message content for a user
func (r *ConversationRepository) Search(userID, query string, limit int) ([]*Conversation, error) {
	if limit <= 0 {
		limit = 20
	}

	// Search in conversation titles and message content
	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(
		`SELECT DISTINCT c.id, c.user_id, c.title, c.provider, c.model, c.system_prompt, c.created_at, c.updated_at
		 FROM conversations c
		 LEFT JOIN messages m ON c.id = m.conversation_id
		 WHERE c.user_id = ? AND (c.title LIKE ? OR m.content LIKE ?)
		 ORDER BY c.updated_at DESC
		 LIMIT ?`,
		userID, searchPattern, searchPattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		var title, systemPrompt sql.NullString

		err := rows.Scan(&conv.ID, &conv.UserID, &title, &conv.Provider, &conv.Model, &systemPrompt, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		conv.Title = title.String
		conv.SystemPrompt = systemPrompt.String
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// MessageRepository handles message database operations
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(conversationID, role, content string, toolCalls []ToolCall, toolCallID string) (*Message, error) {
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
		`INSERT INTO messages (id, conversation_id, role, content, tool_calls, tool_call_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, conversationID, role, content, toolCallsJSON, toolCallIDNull, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation updated_at
	_, _ = r.db.Exec(`UPDATE conversations SET updated_at = ? WHERE id = ?`, now, conversationID)

	return &Message{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		ToolCalls:      toolCalls,
		ToolCallID:     toolCallID,
		CreatedAt:      now,
	}, nil
}

// ListByConversationID retrieves all messages for a conversation
func (r *MessageRepository) ListByConversationID(conversationID string) ([]*Message, error) {
	rows, err := r.db.Query(
		`SELECT id, conversation_id, role, content, tool_calls, tool_call_id, tokens_used, created_at
		 FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var toolCallsJSON, toolCallID sql.NullString
		var tokensUsed sql.NullInt64

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &toolCallsJSON, &toolCallID, &tokensUsed, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if toolCallsJSON.Valid {
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		msg.ToolCallID = toolCallID.String
		msg.TokensUsed = int(tokensUsed.Int64)
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id string) (*Message, error) {
	msg := &Message{}
	var toolCallsJSON, toolCallID sql.NullString
	var tokensUsed sql.NullInt64

	err := r.db.QueryRow(
		`SELECT id, conversation_id, role, content, tool_calls, tool_call_id, tokens_used, created_at
		 FROM messages WHERE id = ?`,
		id,
	).Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &toolCallsJSON, &toolCallID, &tokensUsed, &msg.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if toolCallsJSON.Valid {
		if err := json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
		}
	}

	msg.ToolCallID = toolCallID.String
	msg.TokensUsed = int(tokensUsed.Int64)

	return msg, nil
}
