package steps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/workflow"
)

// GitStep handles git operations for the workflow
type GitStep struct {
	agentRepo workflow.AgentRepository
}

// NewGitStep creates a new git step handler
func NewGitStep(agentRepo workflow.AgentRepository) *GitStep {
	return &GitStep{
		agentRepo: agentRepo,
	}
}

// CreateBranchIfNeeded creates a branch if this is the first run (Step 5)
func (s *GitStep) CreateBranchIfNeeded(ctx context.Context, sandboxCtx *workflow.SandboxContext, agentID, agentName string) (string, error) {
	if sandboxCtx == nil || sandboxCtx.GitClient == nil {
		return "", nil
	}

	// Check if branch already exists
	if sandboxCtx.BranchName != "" {
		// Try to checkout existing branch
		if err := sandboxCtx.GitClient.Checkout(ctx, sandboxCtx.WorkDir, sandboxCtx.BranchName); err == nil {
			return sandboxCtx.BranchName, nil
		}
		// Branch doesn't exist yet, continue to create it
	}

	// Generate branch name from agent name or ID
	branchName := generateBranchName(agentName, agentID)

	// Create the new branch
	if err := sandboxCtx.GitClient.CreateBranch(ctx, sandboxCtx.WorkDir, branchName); err != nil {
		return "", fmt.Errorf("%w: %v", workflow.ErrBranchCreationFailed, err)
	}

	// Update the sandbox context
	sandboxCtx.BranchName = branchName

	// Update the agent record with the branch name
	if s.agentRepo != nil {
		if err := s.agentRepo.UpdateBranch(ctx, agentID, branchName); err != nil {
			// Log but don't fail on this error
			fmt.Printf("warning: failed to update agent branch: %v\n", err)
		}
	}

	return branchName, nil
}

// CommitChanges commits any changes made by the agent (Step 8)
func (s *GitStep) CommitChanges(ctx context.Context, sandboxCtx *workflow.SandboxContext, message string, prefix string) (string, error) {
	if sandboxCtx == nil || sandboxCtx.GitClient == nil {
		return "", nil
	}

	// Check if there are any changes to commit
	hasChanges, err := sandboxCtx.GitClient.HasChanges(ctx, sandboxCtx.WorkDir)
	if err != nil {
		return "", fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		// No changes to commit
		return "", nil
	}

	// Add all changes
	if err := sandboxCtx.GitClient.Add(ctx, sandboxCtx.WorkDir, "."); err != nil {
		return "", fmt.Errorf("failed to stage changes: %w", err)
	}

	// Build commit message
	commitMessage := message
	if prefix != "" {
		commitMessage = prefix + message
	}

	// Create commit
	if err := sandboxCtx.GitClient.Commit(ctx, sandboxCtx.WorkDir, commitMessage); err != nil {
		return "", fmt.Errorf("%w: %v", workflow.ErrCommitFailed, err)
	}

	// Get the commit SHA (would need to be implemented in GitClient)
	// For now, return empty string
	return "", nil
}

// PushChanges pushes committed changes to the remote (optional step)
func (s *GitStep) PushChanges(ctx context.Context, sandboxCtx *workflow.SandboxContext, token string) error {
	if sandboxCtx == nil || sandboxCtx.GitClient == nil {
		return nil
	}

	if sandboxCtx.BranchName == "" {
		return nil
	}

	if err := sandboxCtx.GitClient.Push(ctx, sandboxCtx.WorkDir, "origin", sandboxCtx.BranchName, token); err != nil {
		return fmt.Errorf("%w: %v", workflow.ErrPushFailed, err)
	}

	return nil
}

// generateBranchName generates a valid git branch name from agent name and ID
func generateBranchName(agentName, agentID string) string {
	// Use agent name if available, otherwise use ID
	base := agentName
	if base == "" {
		base = agentID
	}

	// Sanitize for git branch name
	// Replace spaces with hyphens
	base = strings.ReplaceAll(base, " ", "-")

	// Remove or replace invalid characters
	var result strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	sanitized := strings.ToLower(result.String())

	// Limit length
	if len(sanitized) > 40 {
		sanitized = sanitized[:40]
	}

	// Add prefix and timestamp
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("agent/%s-%s", sanitized, timestamp)
}

// GetCurrentBranch returns the current branch name
func (s *GitStep) GetCurrentBranch(ctx context.Context, sandboxCtx *workflow.SandboxContext) (string, error) {
	if sandboxCtx == nil || sandboxCtx.GitClient == nil {
		return "", nil
	}

	return sandboxCtx.GitClient.GetCurrentBranch(ctx, sandboxCtx.WorkDir)
}
