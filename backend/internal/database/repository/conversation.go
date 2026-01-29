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
	ID              string
	ConversationID  string
	ParentID        *string // For threading conversations
	Role            string
	Content         string
	ThinkingContent string // Extended thinking content
	ToolCalls       []ToolCall
	ToolCallID      string
	Provider        string
	Model           string
	Status          string // streaming, complete, error
	InputTokens     int
	OutputTokens    int
	FinishReason    string
	MetadataJSON    string
	TokensUsed      int // Deprecated: use InputTokens + OutputTokens
	CreatedAt       time.Time
}

// MessageChunk represents a streaming chunk for reconstruction
type MessageChunk struct {
	ID         string
	MessageID  string
	ChunkIndex int
	Content    string
	ChunkType  string // content, thinking
	CreatedAt  time.Time
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

// MessageCreateOptions contains optional fields for message creation
type MessageCreateOptions struct {
	ParentID        *string
	Provider        string
	Model           string
	Status          string
	ThinkingContent string
	InputTokens     int
	OutputTokens    int
	FinishReason    string
	MetadataJSON    string
}

// Create creates a new message
func (r *MessageRepository) Create(conversationID, role, content string, toolCalls []ToolCall, toolCallID string) (*Message, error) {
	return r.CreateWithOptions(conversationID, role, content, toolCalls, toolCallID, nil)
}

// CreateWithOptions creates a new message with additional options
func (r *MessageRepository) CreateWithOptions(conversationID, role, content string, toolCalls []ToolCall, toolCallID string, opts *MessageCreateOptions) (*Message, error) {
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

	// Set defaults for optional fields
	status := "complete"
	var parentID *string
	var provider, model, thinkingContent, finishReason, metadataJSON string
	var inputTokens, outputTokens int

	if opts != nil {
		if opts.Status != "" {
			status = opts.Status
		}
		parentID = opts.ParentID
		provider = opts.Provider
		model = opts.Model
		thinkingContent = opts.ThinkingContent
		inputTokens = opts.InputTokens
		outputTokens = opts.OutputTokens
		finishReason = opts.FinishReason
		metadataJSON = opts.MetadataJSON
	}

	var parentIDNull sql.NullString
	if parentID != nil {
		parentIDNull = sql.NullString{String: *parentID, Valid: true}
	}

	_, err := r.db.Exec(
		`INSERT INTO messages (id, conversation_id, parent_id, role, content, thinking_content, tool_calls, tool_call_id, provider, model, status, input_tokens, output_tokens, finish_reason, metadata_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, conversationID, parentIDNull, role, content, thinkingContent, toolCallsJSON, toolCallIDNull, provider, model, status, inputTokens, outputTokens, finishReason, metadataJSON, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation updated_at
	_, _ = r.db.Exec(`UPDATE conversations SET updated_at = ? WHERE id = ?`, now, conversationID)

	return &Message{
		ID:              id,
		ConversationID:  conversationID,
		ParentID:        parentID,
		Role:            role,
		Content:         content,
		ThinkingContent: thinkingContent,
		ToolCalls:       toolCalls,
		ToolCallID:      toolCallID,
		Provider:        provider,
		Model:           model,
		Status:          status,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		FinishReason:    finishReason,
		MetadataJSON:    metadataJSON,
		CreatedAt:       now,
	}, nil
}

// ListByConversationID retrieves all messages for a conversation
func (r *MessageRepository) ListByConversationID(conversationID string) ([]*Message, error) {
	rows, err := r.db.Query(
		`SELECT id, conversation_id, parent_id, role, content, thinking_content, tool_calls, tool_call_id,
		        provider, model, status, input_tokens, output_tokens, finish_reason, metadata_json, tokens_used, created_at
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
		var parentID, toolCallsJSON, toolCallID, thinkingContent, provider, model, status, finishReason, metadataJSON sql.NullString
		var tokensUsed, inputTokens, outputTokens sql.NullInt64

		err := rows.Scan(&msg.ID, &msg.ConversationID, &parentID, &msg.Role, &msg.Content, &thinkingContent,
			&toolCallsJSON, &toolCallID, &provider, &model, &status, &inputTokens, &outputTokens, &finishReason, &metadataJSON, &tokensUsed, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if toolCallsJSON.Valid {
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		if parentID.Valid {
			msg.ParentID = &parentID.String
		}
		msg.ToolCallID = toolCallID.String
		msg.ThinkingContent = thinkingContent.String
		msg.Provider = provider.String
		msg.Model = model.String
		msg.Status = status.String
		if msg.Status == "" {
			msg.Status = "complete" // Default for older messages
		}
		msg.InputTokens = int(inputTokens.Int64)
		msg.OutputTokens = int(outputTokens.Int64)
		msg.FinishReason = finishReason.String
		msg.MetadataJSON = metadataJSON.String
		msg.TokensUsed = int(tokensUsed.Int64)
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id string) (*Message, error) {
	msg := &Message{}
	var parentID, toolCallsJSON, toolCallID, thinkingContent, provider, model, status, finishReason, metadataJSON sql.NullString
	var tokensUsed, inputTokens, outputTokens sql.NullInt64

	err := r.db.QueryRow(
		`SELECT id, conversation_id, parent_id, role, content, thinking_content, tool_calls, tool_call_id,
		        provider, model, status, input_tokens, output_tokens, finish_reason, metadata_json, tokens_used, created_at
		 FROM messages WHERE id = ?`,
		id,
	).Scan(&msg.ID, &msg.ConversationID, &parentID, &msg.Role, &msg.Content, &thinkingContent,
		&toolCallsJSON, &toolCallID, &provider, &model, &status, &inputTokens, &outputTokens, &finishReason, &metadataJSON, &tokensUsed, &msg.CreatedAt)

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

	if parentID.Valid {
		msg.ParentID = &parentID.String
	}
	msg.ToolCallID = toolCallID.String
	msg.ThinkingContent = thinkingContent.String
	msg.Provider = provider.String
	msg.Model = model.String
	msg.Status = status.String
	if msg.Status == "" {
		msg.Status = "complete"
	}
	msg.InputTokens = int(inputTokens.Int64)
	msg.OutputTokens = int(outputTokens.Int64)
	msg.FinishReason = finishReason.String
	msg.MetadataJSON = metadataJSON.String
	msg.TokensUsed = int(tokensUsed.Int64)

	return msg, nil
}

// UpdateStatus updates the status of a message
func (r *MessageRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(`UPDATE messages SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}

// UpdateContent updates the content of a message
func (r *MessageRepository) UpdateContent(id, content string) error {
	_, err := r.db.Exec(`UPDATE messages SET content = ? WHERE id = ?`, content, id)
	if err != nil {
		return fmt.Errorf("failed to update message content: %w", err)
	}
	return nil
}

// AppendContent appends to the content of a message
func (r *MessageRepository) AppendContent(id, delta string) error {
	_, err := r.db.Exec(`UPDATE messages SET content = content || ? WHERE id = ?`, delta, id)
	if err != nil {
		return fmt.Errorf("failed to append message content: %w", err)
	}
	return nil
}

// SetThinkingContent sets the thinking content of a message
func (r *MessageRepository) SetThinkingContent(id, thinking string) error {
	_, err := r.db.Exec(`UPDATE messages SET thinking_content = ? WHERE id = ?`, thinking, id)
	if err != nil {
		return fmt.Errorf("failed to set thinking content: %w", err)
	}
	return nil
}

// SetFinishReason sets the finish reason and marks message as complete
func (r *MessageRepository) SetFinishReason(id, finishReason string) error {
	_, err := r.db.Exec(`UPDATE messages SET finish_reason = ?, status = 'complete' WHERE id = ?`, finishReason, id)
	if err != nil {
		return fmt.Errorf("failed to set finish reason: %w", err)
	}
	return nil
}

// SetTokenCounts updates the token counts for a message
func (r *MessageRepository) SetTokenCounts(id string, inputTokens, outputTokens int) error {
	_, err := r.db.Exec(`UPDATE messages SET input_tokens = ?, output_tokens = ?, tokens_used = ? WHERE id = ?`,
		inputTokens, outputTokens, inputTokens+outputTokens, id)
	if err != nil {
		return fmt.Errorf("failed to set token counts: %w", err)
	}
	return nil
}

// GetThread retrieves all messages in a thread starting from a parent
func (r *MessageRepository) GetThread(parentID string) ([]*Message, error) {
	rows, err := r.db.Query(
		`WITH RECURSIVE thread AS (
			SELECT id FROM messages WHERE id = ?
			UNION ALL
			SELECT m.id FROM messages m INNER JOIN thread t ON m.parent_id = t.id
		)
		SELECT m.id, m.conversation_id, m.parent_id, m.role, m.content, m.thinking_content, m.tool_calls, m.tool_call_id,
		       m.provider, m.model, m.status, m.input_tokens, m.output_tokens, m.finish_reason, m.metadata_json, m.tokens_used, m.created_at
		FROM messages m INNER JOIN thread t ON m.id = t.id ORDER BY m.created_at ASC`,
		parentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var parentIDVal, toolCallsJSON, toolCallID, thinkingContent, provider, model, status, finishReason, metadataJSON sql.NullString
		var tokensUsed, inputTokens, outputTokens sql.NullInt64

		err := rows.Scan(&msg.ID, &msg.ConversationID, &parentIDVal, &msg.Role, &msg.Content, &thinkingContent,
			&toolCallsJSON, &toolCallID, &provider, &model, &status, &inputTokens, &outputTokens, &finishReason, &metadataJSON, &tokensUsed, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if toolCallsJSON.Valid {
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		if parentIDVal.Valid {
			msg.ParentID = &parentIDVal.String
		}
		msg.ToolCallID = toolCallID.String
		msg.ThinkingContent = thinkingContent.String
		msg.Provider = provider.String
		msg.Model = model.String
		msg.Status = status.String
		if msg.Status == "" {
			msg.Status = "complete"
		}
		msg.InputTokens = int(inputTokens.Int64)
		msg.OutputTokens = int(outputTokens.Int64)
		msg.FinishReason = finishReason.String
		msg.MetadataJSON = metadataJSON.String
		msg.TokensUsed = int(tokensUsed.Int64)
		messages = append(messages, msg)
	}

	return messages, nil
}

// SaveChunk saves a streaming chunk for later reconstruction
func (r *MessageRepository) SaveChunk(messageID string, index int, content, chunkType string) error {
	id := uuid.New().String()
	_, err := r.db.Exec(
		`INSERT INTO message_chunks (id, message_id, chunk_index, content, chunk_type) VALUES (?, ?, ?, ?, ?)`,
		id, messageID, index, content, chunkType,
	)
	if err != nil {
		return fmt.Errorf("failed to save chunk: %w", err)
	}
	return nil
}

// GetChunks retrieves all chunks for a message
func (r *MessageRepository) GetChunks(messageID string) ([]*MessageChunk, error) {
	rows, err := r.db.Query(
		`SELECT id, message_id, chunk_index, content, chunk_type, created_at
		 FROM message_chunks WHERE message_id = ? ORDER BY chunk_index ASC`,
		messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*MessageChunk
	for rows.Next() {
		chunk := &MessageChunk{}
		err := rows.Scan(&chunk.ID, &chunk.MessageID, &chunk.ChunkIndex, &chunk.Content, &chunk.ChunkType, &chunk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// DeleteChunks deletes all chunks for a message
func (r *MessageRepository) DeleteChunks(messageID string) error {
	_, err := r.db.Exec(`DELETE FROM message_chunks WHERE message_id = ?`, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}
	return nil
}
