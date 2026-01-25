package steps

import (
	"context"
	"fmt"
	"os"

	"github.com/jacklau/prism/internal/workflow"
)

// CleanupStep handles sandbox cleanup and status updates
type CleanupStep struct {
	agentRepo     workflow.AgentRepository
	executionRepo *workflow.AgentExecutionRepository
}

// NewCleanupStep creates a new cleanup step handler
func NewCleanupStep(
	agentRepo workflow.AgentRepository,
	executionRepo *workflow.AgentExecutionRepository,
) *CleanupStep {
	return &CleanupStep{
		agentRepo:     agentRepo,
		executionRepo: executionRepo,
	}
}

// CleanupSandbox cleans up the sandbox environment (Step 9)
func (s *CleanupStep) CleanupSandbox(ctx context.Context, sandboxCtx *workflow.SandboxContext, keepFiles bool) error {
	if sandboxCtx == nil {
		return nil
	}

	// If we should keep files (for debugging or review), skip cleanup
	if keepFiles {
		return nil
	}

	// Only clean up if the working directory is a repo subdirectory
	// Don't clean up the user's main working directory
	if sandboxCtx.RepoURL != "" && sandboxCtx.WorkDir != "" {
		// Check if directory exists
		if _, err := os.Stat(sandboxCtx.WorkDir); err == nil {
			// Remove the repo directory
			if err := os.RemoveAll(sandboxCtx.WorkDir); err != nil {
				// Log but don't fail on cleanup errors
				fmt.Printf("warning: failed to cleanup sandbox directory %s: %v\n", sandboxCtx.WorkDir, err)
			}
		}
	}

	return nil
}

// MarkComplete marks the agent execution as complete (Step 10)
func (s *CleanupStep) MarkComplete(ctx context.Context, executionID, agentID, status, commitSHA string, errorMsg string, iterations int) error {
	// Update execution status
	if s.executionRepo != nil && executionID != "" {
		if status == "completed" {
			if err := s.executionRepo.Complete(ctx, executionID, commitSHA, iterations); err != nil {
				return fmt.Errorf("failed to mark execution complete: %w", err)
			}
		} else if status == "failed" {
			if err := s.executionRepo.Fail(ctx, executionID, errorMsg); err != nil {
				return fmt.Errorf("failed to mark execution failed: %w", err)
			}
		}
	}

	// Update agent status
	if s.agentRepo != nil && agentID != "" {
		if err := s.agentRepo.UpdateStatus(ctx, agentID, status, errorMsg); err != nil {
			return fmt.Errorf("failed to update agent status: %w", err)
		}
	}

	return nil
}

// MarkFailed is a convenience method to mark the execution as failed
func (s *CleanupStep) MarkFailed(ctx context.Context, executionID, agentID string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	return s.MarkComplete(ctx, executionID, agentID, "failed", "", errorMsg, 0)
}

// MarkCancelled marks the execution as cancelled
func (s *CleanupStep) MarkCancelled(ctx context.Context, executionID, agentID string) error {
	return s.MarkComplete(ctx, executionID, agentID, "cancelled", "", "workflow was cancelled", 0)
}

// UpdateAgentStatus updates only the agent status (without execution record)
func (s *CleanupStep) UpdateAgentStatus(ctx context.Context, agentID, status, errorMsg string) error {
	if s.agentRepo == nil || agentID == "" {
		return nil
	}

	return s.agentRepo.UpdateStatus(ctx, agentID, status, errorMsg)
}
