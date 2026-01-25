package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/workflow"
)

// SetupStep handles sandbox creation and conversation history loading
type SetupStep struct {
	sandbox      *sandbox.Service
	convRepo     *repository.ConversationRepository
	messageRepo  *repository.MessageRepository
	gitClient    workflow.GitClient
}

// NewSetupStep creates a new setup step handler
func NewSetupStep(
	sandbox *sandbox.Service,
	convRepo *repository.ConversationRepository,
	messageRepo *repository.MessageRepository,
	gitClient workflow.GitClient,
) *SetupStep {
	return &SetupStep{
		sandbox:     sandbox,
		convRepo:    convRepo,
		messageRepo: messageRepo,
		gitClient:   gitClient,
	}
}

// CreateSandbox creates an isolated sandbox environment for the agent (Step 3)
func (s *SetupStep) CreateSandbox(ctx context.Context, userID string, repoURL string, token string, cloneDepth int) (*workflow.SandboxContext, error) {
	// Get or create work directory for the user
	workDir, err := s.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work directory: %w", err)
	}

	sandboxCtx := &workflow.SandboxContext{
		UserID:     userID,
		WorkDir:    workDir,
		RepoURL:    repoURL,
		Service:    s.sandbox,
		GitClient:  s.gitClient,
		CloneDepth: cloneDepth,
	}

	// If a repository URL is provided, clone it
	if repoURL != "" {
		// Create a unique directory for this repo
		repoDir := filepath.Join(workDir, "repo")

		// Remove existing repo directory if it exists
		if _, statErr := os.Stat(repoDir); statErr == nil {
			if removeErr := os.RemoveAll(repoDir); removeErr != nil {
				return nil, fmt.Errorf("failed to clean existing repo directory: %w", removeErr)
			}
		}

		// Clone the repository
		if err := s.gitClient.Clone(ctx, repoURL, repoDir, token); err != nil {
			return nil, fmt.Errorf("%w: %v", workflow.ErrCloneFailed, err)
		}

		sandboxCtx.WorkDir = repoDir
	}

	return sandboxCtx, nil
}

// LoadConversationHistory loads the conversation history for context (Step 4)
func (s *SetupStep) LoadConversationHistory(ctx context.Context, conversationID string) ([]llm.Message, error) {
	if conversationID == "" {
		// No conversation ID, return empty history
		return []llm.Message{}, nil
	}

	// Get conversation details
	conv, err := s.convRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if conv == nil {
		// Conversation not found, return empty history
		return []llm.Message{}, nil
	}

	// Get messages
	messages, err := s.messageRepo.ListByConversationID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert to LLM messages
	llmMessages := make([]llm.Message, 0, len(messages))
	for _, msg := range messages {
		llmMsg := llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

		// Convert tool calls
		if len(msg.ToolCalls) > 0 {
			llmMsg.ToolCalls = make([]llm.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				llmMsg.ToolCalls[i] = llm.ToolCall{
					ID:         tc.ID,
					Name:       tc.Name,
					Parameters: tc.Parameters,
				}
			}
		}

		llmMessages = append(llmMessages, llmMsg)
	}

	return llmMessages, nil
}

// CreateConversation creates a new conversation for the agent
func (s *SetupStep) CreateConversation(ctx context.Context, userID, provider, model, systemPrompt string) (*repository.Conversation, error) {
	conv, err := s.convRepo.Create(userID, provider, model, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	return conv, nil
}

// SaveMessage saves a message to the conversation history
func (s *SetupStep) SaveMessage(ctx context.Context, conversationID, role, content string, toolCalls []repository.ToolCall, toolCallID string) error {
	_, err := s.messageRepo.Create(conversationID, role, content, toolCalls, toolCallID)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}
