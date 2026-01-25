package routes

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
)

// MessageBuilder accumulates streaming chunks and manages message state
type MessageBuilder struct {
	mu sync.Mutex

	// Message identity
	ID             string
	ConversationID string
	ParentID       *string
	Provider       string
	Model          string

	// Content accumulators
	content         strings.Builder
	thinkingContent strings.Builder

	// Tool calls collected during streaming
	toolCalls []llm.ToolCall

	// State tracking
	status       string // streaming, complete, error
	finishReason string
	inputTokens  int
	outputTokens int
	chunkIndex   int

	// Repository for persistence
	messageRepo *repository.MessageRepository

	// Chunk persistence options
	persistChunks bool
	createdAt     time.Time
}

// NewMessageBuilder creates a new MessageBuilder for streaming
func NewMessageBuilder(conversationID, provider, model string, messageRepo *repository.MessageRepository) *MessageBuilder {
	return &MessageBuilder{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		Provider:       provider,
		Model:          model,
		status:         "streaming",
		messageRepo:    messageRepo,
		persistChunks:  false, // Set to true if you need chunk reconstruction
		createdAt:      time.Now(),
	}
}

// WithParent sets the parent message ID for threading
func (b *MessageBuilder) WithParent(parentID string) *MessageBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ParentID = &parentID
	return b
}

// WithChunkPersistence enables saving individual chunks for reconstruction
func (b *MessageBuilder) WithChunkPersistence(enabled bool) *MessageBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.persistChunks = enabled
	return b
}

// AppendContent adds content delta from streaming
func (b *MessageBuilder) AppendContent(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.content.WriteString(delta)

	// Optionally persist chunk for reconstruction
	if b.persistChunks && b.messageRepo != nil && delta != "" {
		_ = b.messageRepo.SaveChunk(b.ID, b.chunkIndex, delta, "content")
		b.chunkIndex++
	}
}

// AppendThinkingContent adds thinking content delta from extended thinking
func (b *MessageBuilder) AppendThinkingContent(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.thinkingContent.WriteString(delta)

	// Optionally persist chunk for reconstruction
	if b.persistChunks && b.messageRepo != nil && delta != "" {
		_ = b.messageRepo.SaveChunk(b.ID, b.chunkIndex, delta, "thinking")
		b.chunkIndex++
	}
}

// AddToolCall adds a tool call from the stream
func (b *MessageBuilder) AddToolCall(tc llm.ToolCall) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.toolCalls = append(b.toolCalls, tc)
}

// SetInputTokens sets the input token count
func (b *MessageBuilder) SetInputTokens(count int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inputTokens = count
}

// SetOutputTokens sets the output token count
func (b *MessageBuilder) SetOutputTokens(count int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.outputTokens = count
}

// SetTokenCounts sets both input and output token counts
func (b *MessageBuilder) SetTokenCounts(input, output int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inputTokens = input
	b.outputTokens = output
}

// SetFinishReason sets the finish reason and marks streaming complete
func (b *MessageBuilder) SetFinishReason(reason string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.finishReason = reason
	if reason == "error" {
		b.status = "error"
	} else {
		b.status = "complete"
	}
}

// SetError marks the message as having an error
func (b *MessageBuilder) SetError() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = "error"
	b.finishReason = "error"
}

// GetContent returns the accumulated content
func (b *MessageBuilder) GetContent() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.content.String()
}

// GetThinkingContent returns the accumulated thinking content
func (b *MessageBuilder) GetThinkingContent() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.thinkingContent.String()
}

// GetToolCalls returns the collected tool calls
func (b *MessageBuilder) GetToolCalls() []llm.ToolCall {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.toolCalls
}

// GetStatus returns the current status
func (b *MessageBuilder) GetStatus() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.status
}

// IsEmpty returns true if no content has been accumulated
func (b *MessageBuilder) IsEmpty() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.content.Len() == 0 && len(b.toolCalls) == 0
}

// HasContent returns true if there is content to save
func (b *MessageBuilder) HasContent() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.content.Len() > 0 || len(b.toolCalls) > 0
}

// Build creates the final Message without saving to database
func (b *MessageBuilder) Build() *repository.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Convert LLM tool calls to repository format
	var repoToolCalls []repository.ToolCall
	if len(b.toolCalls) > 0 {
		repoToolCalls = make([]repository.ToolCall, len(b.toolCalls))
		for i, tc := range b.toolCalls {
			repoToolCalls[i] = repository.ToolCall{
				ID:         tc.ID,
				Name:       tc.Name,
				Parameters: tc.Parameters,
			}
		}
	}

	return &repository.Message{
		ID:              b.ID,
		ConversationID:  b.ConversationID,
		ParentID:        b.ParentID,
		Role:            "assistant",
		Content:         b.content.String(),
		ThinkingContent: b.thinkingContent.String(),
		ToolCalls:       repoToolCalls,
		Provider:        b.Provider,
		Model:           b.Model,
		Status:          b.status,
		InputTokens:     b.inputTokens,
		OutputTokens:    b.outputTokens,
		FinishReason:    b.finishReason,
		TokensUsed:      b.inputTokens + b.outputTokens,
		CreatedAt:       b.createdAt,
	}
}

// Save persists the message to the database
func (b *MessageBuilder) Save() (*repository.Message, error) {
	b.mu.Lock()
	content := b.content.String()
	thinkingContent := b.thinkingContent.String()
	toolCalls := b.toolCalls
	status := b.status
	inputTokens := b.inputTokens
	outputTokens := b.outputTokens
	finishReason := b.finishReason
	provider := b.Provider
	model := b.Model
	parentID := b.ParentID
	b.mu.Unlock()

	// Convert LLM tool calls to repository format
	var repoToolCalls []repository.ToolCall
	if len(toolCalls) > 0 {
		repoToolCalls = make([]repository.ToolCall, len(toolCalls))
		for i, tc := range toolCalls {
			repoToolCalls[i] = repository.ToolCall{
				ID:         tc.ID,
				Name:       tc.Name,
				Parameters: tc.Parameters,
			}
		}
	}

	opts := &repository.MessageCreateOptions{
		ParentID:        parentID,
		Provider:        provider,
		Model:           model,
		Status:          status,
		ThinkingContent: thinkingContent,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		FinishReason:    finishReason,
	}

	msg, err := b.messageRepo.CreateWithOptions(b.ConversationID, "assistant", content, repoToolCalls, "", opts)
	if err != nil {
		return nil, err
	}

	// Update the ID to match the saved message
	b.mu.Lock()
	b.ID = msg.ID
	b.mu.Unlock()

	// Clean up chunks if they were persisted
	if b.persistChunks {
		// Chunks are now represented in the final message, we could delete them
		// but keeping them allows for audit/reconstruction
	}

	return msg, nil
}

// SaveWithID persists the message with a pre-determined ID
func (b *MessageBuilder) SaveWithID(id string) (*repository.Message, error) {
	b.mu.Lock()
	b.ID = id
	b.mu.Unlock()
	return b.Save()
}
