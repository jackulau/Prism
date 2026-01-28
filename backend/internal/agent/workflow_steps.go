package agent

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// WorkflowContext holds dependencies needed by workflow step functions
type WorkflowContext struct {
	SandboxService   *sandbox.Service
	MessageRepo      *repository.MessageRepository
	ConversationRepo *repository.ConversationRepository
	WorkspaceRepo    *repository.WorkspaceRepository
}

// NewWorkflowContext creates a new workflow context with required dependencies
func NewWorkflowContext(
	sandboxService *sandbox.Service,
	messageRepo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
	workspaceRepo *repository.WorkspaceRepository,
) *WorkflowContext {
	return &WorkflowContext{
		SandboxService:   sandboxService,
		MessageRepo:      messageRepo,
		ConversationRepo: conversationRepo,
		WorkspaceRepo:    workspaceRepo,
	}
}

// LoadAgent retrieves agent configuration by ID from storage
// Returns the agent if found, or an error if not found
func LoadAgent(manager *Manager, agentID string) (*Agent, error) {
	log.Printf("[workflow] Loading agent: %s", agentID)

	// Check if agent exists in any active execution
	executions := manager.ListExecutions()
	for _, exec := range executions {
		for _, agent := range exec.Agents {
			if agent.ID == agentID {
				log.Printf("[workflow] Found agent %s in execution %s", agentID, exec.ID)
				return agent, nil
			}
		}
	}

	log.Printf("[workflow] Agent %s not found", agentID)
	return nil, ErrAgentNotFound
}

// ValidateAndGetToken validates repository access and returns GitHub token
// This is a placeholder for GitHub OAuth integration
func ValidateAndGetToken(ctx context.Context, repoName string) (string, error) {
	log.Printf("[workflow] Validating repository access: %s", repoName)

	if repoName == "" {
		return "", ErrInvalidRepository
	}

	// Validate repository name format (owner/repo)
	if !isValidRepoFormat(repoName) {
		log.Printf("[workflow] Invalid repository format: %s", repoName)
		return "", fmt.Errorf("%w: invalid format, expected owner/repo", ErrInvalidRepository)
	}

	// In a full implementation, this would:
	// 1. Check if user has GitHub OAuth token stored
	// 2. Validate token has access to the repository
	// 3. Return the token for git operations
	//
	// For now, return empty string (no token needed for public repos)
	log.Printf("[workflow] Repository %s validated successfully", repoName)
	return "", nil
}

// PrepareSandbox creates and configures a sandbox for agent execution
// Returns the sandbox service configured for the agent's workspace
func PrepareSandbox(ctx context.Context, wfCtx *WorkflowContext, agentID string, repoName string, token string) (*sandbox.Service, string, error) {
	log.Printf("[workflow] Preparing sandbox for agent: %s, repo: %s", agentID, repoName)

	if wfCtx.SandboxService == nil {
		return nil, "", ErrSandboxNotReady
	}

	// Get or create working directory for the agent
	workDir, err := wfCtx.SandboxService.GetOrCreateWorkDir(agentID)
	if err != nil {
		log.Printf("[workflow] Failed to get/create work directory: %v", err)
		return nil, "", fmt.Errorf("%w: %v", ErrSandboxNotReady, err)
	}

	log.Printf("[workflow] Sandbox ready at: %s", workDir)
	return wfCtx.SandboxService, workDir, nil
}

// LoadPreviousMessages retrieves conversation history for an agent
// Returns the messages in LLM format for context continuity
func LoadPreviousMessages(ctx context.Context, wfCtx *WorkflowContext, conversationID string) ([]llm.Message, error) {
	log.Printf("[workflow] Loading previous messages for conversation: %s", conversationID)

	if conversationID == "" {
		// No conversation ID means this is a new conversation
		log.Printf("[workflow] No conversation ID, returning empty message history")
		return []llm.Message{}, nil
	}

	if wfCtx.MessageRepo == nil {
		return nil, ErrMessageLoadFailed
	}

	// Load messages from repository
	repoMessages, err := wfCtx.MessageRepo.ListByConversationID(conversationID)
	if err != nil {
		log.Printf("[workflow] Failed to load messages: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrMessageLoadFailed, err)
	}

	// Convert repository messages to LLM messages
	messages := make([]llm.Message, 0, len(repoMessages))
	for _, msg := range repoMessages {
		llmMsg := llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

		// Convert tool calls if present
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

		messages = append(messages, llmMsg)
	}

	log.Printf("[workflow] Loaded %d messages for conversation %s", len(messages), conversationID)
	return messages, nil
}

// CreateBranchIfNeeded creates a feature branch for agent work
// Branch naming follows: prism/{agent-id}-{sanitized-prompt}
func CreateBranchIfNeeded(ctx context.Context, sandboxSvc *sandbox.Service, workDir string, agentID string, prompt string) (string, error) {
	log.Printf("[workflow] Checking/creating branch for agent: %s", agentID)

	// Generate branch name
	branchName := generateBranchName(agentID, prompt)
	log.Printf("[workflow] Generated branch name: %s", branchName)

	// Check if we're in a git repository
	checkCmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	checkCmd.Dir = workDir
	if err := checkCmd.Run(); err != nil {
		log.Printf("[workflow] Not a git repository, skipping branch creation")
		return "", nil // Not a git repo, nothing to do
	}

	// Check if branch already exists
	listCmd := exec.CommandContext(ctx, "git", "branch", "--list", branchName)
	listCmd.Dir = workDir
	output, err := listCmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		log.Printf("[workflow] Branch %s already exists, checking out", branchName)
		// Branch exists, just checkout
		checkoutCmd := exec.CommandContext(ctx, "git", "checkout", branchName)
		checkoutCmd.Dir = workDir
		if err := checkoutCmd.Run(); err != nil {
			return "", fmt.Errorf("%w: failed to checkout existing branch: %v", ErrBranchCreationFailed, err)
		}
		return branchName, nil
	}

	// Create and checkout new branch
	createCmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	createCmd.Dir = workDir
	if err := createCmd.Run(); err != nil {
		log.Printf("[workflow] Failed to create branch: %v", err)
		return "", fmt.Errorf("%w: %v", ErrBranchCreationFailed, err)
	}

	log.Printf("[workflow] Created and checked out branch: %s", branchName)
	return branchName, nil
}

// SaveAgentResponse persists agent response and usage stats to the database
func SaveAgentResponse(ctx context.Context, wfCtx *WorkflowContext, conversationID string, role string, text string, toolCalls []llm.ToolCall, usage *llm.Usage) error {
	log.Printf("[workflow] Saving agent response for conversation: %s", conversationID)

	if wfCtx.MessageRepo == nil {
		return fmt.Errorf("message repository not available")
	}

	if conversationID == "" {
		return fmt.Errorf("conversation ID is required")
	}

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

	// Create message in repository
	_, err := wfCtx.MessageRepo.Create(conversationID, role, text, repoToolCalls, "")
	if err != nil {
		log.Printf("[workflow] Failed to save response: %v", err)
		return fmt.Errorf("failed to save agent response: %w", err)
	}

	// Log usage if available
	if usage != nil {
		log.Printf("[workflow] Token usage - prompt: %d, completion: %d, total: %d",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
	}

	log.Printf("[workflow] Saved agent response successfully")
	return nil
}

// CommitChangesIfNeeded commits sandbox changes with a generated message
// Only commits if there are actual changes in the working directory
func CommitChangesIfNeeded(ctx context.Context, workDir string, agentID string, prompt string) (string, error) {
	log.Printf("[workflow] Checking for changes to commit in: %s", workDir)

	// Check if we're in a git repository
	checkCmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	checkCmd.Dir = workDir
	if err := checkCmd.Run(); err != nil {
		log.Printf("[workflow] Not a git repository, skipping commit")
		return "", nil
	}

	// Check for changes
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = workDir
	output, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: failed to check git status: %v", ErrCommitFailed, err)
	}

	if strings.TrimSpace(string(output)) == "" {
		log.Printf("[workflow] No changes to commit")
		return "", ErrNoChangesToCommit
	}

	// Stage all changes
	addCmd := exec.CommandContext(ctx, "git", "add", "-A")
	addCmd.Dir = workDir
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("%w: failed to stage changes: %v", ErrCommitFailed, err)
	}

	// Generate commit message (truncated to 50 chars)
	commitMsg := generateCommitMessage(prompt)
	log.Printf("[workflow] Committing with message: %s", commitMsg)

	// Commit changes
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	commitCmd.Dir = workDir
	if err := commitCmd.Run(); err != nil {
		return "", fmt.Errorf("%w: failed to commit changes: %v", ErrCommitFailed, err)
	}

	// Get commit SHA
	shaCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	shaCmd.Dir = workDir
	shaOutput, err := shaCmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: failed to get commit SHA: %v", ErrCommitFailed, err)
	}

	commitSHA := strings.TrimSpace(string(shaOutput))
	log.Printf("[workflow] Committed changes: %s", commitSHA)
	return commitSHA, nil
}

// CleanupSandbox releases sandbox resources
func CleanupSandbox(ctx context.Context, sandboxSvc *sandbox.Service, buildID string) error {
	log.Printf("[workflow] Cleaning up sandbox for build: %s", buildID)

	if sandboxSvc == nil {
		return nil
	}

	if buildID == "" {
		log.Printf("[workflow] No build ID provided, nothing to cleanup")
		return nil
	}

	// Stop any running builds
	if err := sandboxSvc.StopBuild(buildID); err != nil {
		// Build might already be stopped, just log it
		log.Printf("[workflow] Note: Could not stop build %s: %v", buildID, err)
	}

	log.Printf("[workflow] Sandbox cleanup complete")
	return nil
}

// Helper functions

// isValidRepoFormat checks if the repository name follows owner/repo format
func isValidRepoFormat(repoName string) bool {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return false
	}
	// Check that both owner and repo are non-empty and contain valid characters
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	return validPattern.MatchString(parts[0]) && validPattern.MatchString(parts[1])
}

// generateBranchName creates a branch name from agent ID and prompt
// Format: prism/{agent-id}-{sanitized-prompt}
func generateBranchName(agentID string, prompt string) string {
	// Sanitize prompt: lowercase, replace spaces with hyphens, remove special chars
	sanitized := strings.ToLower(prompt)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(sanitized, "")

	// Truncate to reasonable length
	maxPromptLen := 30
	if len(sanitized) > maxPromptLen {
		sanitized = sanitized[:maxPromptLen]
	}

	// Remove trailing hyphens
	sanitized = strings.TrimRight(sanitized, "-")

	// Use short agent ID (first 8 chars)
	shortID := agentID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	if sanitized == "" {
		sanitized = fmt.Sprintf("task-%d", time.Now().Unix())
	}

	return fmt.Sprintf("prism/%s-%s", shortID, sanitized)
}

// generateCommitMessage creates a commit message from the prompt
// Truncated to 50 characters as per acceptance criteria
func generateCommitMessage(prompt string) string {
	// Clean up the prompt
	msg := strings.TrimSpace(prompt)

	// Truncate to 50 chars (with ellipsis if needed)
	maxLen := 50
	if len(msg) > maxLen {
		msg = msg[:maxLen-3] + "..."
	}

	// Ensure it starts with capital letter
	if len(msg) > 0 && msg[0] >= 'a' && msg[0] <= 'z' {
		msg = strings.ToUpper(string(msg[0])) + msg[1:]
	}

	return msg
}
