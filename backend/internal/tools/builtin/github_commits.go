package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// GitHubViewCommitTool displays the full diff and details of a specific commit
type GitHubViewCommitTool struct {
	sandbox *sandbox.Service
}

// NewGitHubViewCommitTool creates a new GitHub view commit tool
func NewGitHubViewCommitTool(sandbox *sandbox.Service) *GitHubViewCommitTool {
	return &GitHubViewCommitTool{sandbox: sandbox}
}

func (t *GitHubViewCommitTool) Name() string {
	return "github_view_commit"
}

func (t *GitHubViewCommitTool) Description() string {
	return "Show the full diff and details of a specific commit. Displays commit message, author, date, and all file changes."
}

func (t *GitHubViewCommitTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"commitSha": {
				Type:        "string",
				Description: "The commit SHA (full or abbreviated) to view",
			},
		},
		Required: []string{"commitSha"},
	}
}

// shaPattern validates commit SHA format (4-40 hex characters)
var shaPattern = regexp.MustCompile(`^[0-9a-fA-F]{4,40}$`)

func (t *GitHubViewCommitTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get commit SHA parameter
	commitSha, ok := params["commitSha"].(string)
	if !ok || commitSha == "" {
		return nil, fmt.Errorf("commitSha parameter is required")
	}

	// Validate SHA format (4-40 hex characters)
	if !shaPattern.MatchString(commitSha) {
		return map[string]interface{}{
			"success": false,
			"error":   "invalid commit SHA format: must be 4-40 hexadecimal characters",
		}, nil
	}

	// Execute git show command
	cmd := exec.CommandContext(ctx, "git", "show", commitSha)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("git show failed: %s", string(output)),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"sha":     commitSha,
		"diff":    string(output),
	}, nil
}

func (t *GitHubViewCommitTool) RequiresConfirmation() bool {
	return false // Read-only operation
}
