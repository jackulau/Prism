package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// GitHubListCommitsTool lists recent commits from a git repository
type GitHubListCommitsTool struct {
	sandbox *sandbox.Service
}

// NewGitHubListCommitsTool creates a new GitHub list commits tool
func NewGitHubListCommitsTool(sandbox *sandbox.Service) *GitHubListCommitsTool {
	return &GitHubListCommitsTool{sandbox: sandbox}
}

func (t *GitHubListCommitsTool) Name() string {
	return "github_list_commits"
}

func (t *GitHubListCommitsTool) Description() string {
	return "List recent commits from the git repository in the workspace. Returns commit SHA, author, date, and message."
}

func (t *GitHubListCommitsTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"commitLimit": {
				Type:        "number",
				Description: "Number of commits to list (1-10, default 5)",
			},
		},
		Required: []string{},
	}
}

func (t *GitHubListCommitsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Get user workspace
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Parse limit parameter (default 5, max 10, min 1)
	limit := 5
	if limitVal, ok := params["commitLimit"].(float64); ok {
		limit = int(limitVal)
		if limit < 1 {
			limit = 1
		}
		if limit > 10 {
			limit = 10
		}
	}

	// Execute git log command
	cmd := exec.CommandContext(ctx, "git", "log",
		fmt.Sprintf("-n%d", limit),
		"--pretty=format:%H - %an, %ad : %s",
		"--date=iso")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("git log failed: %s - %s", err, string(output)),
		}, nil
	}

	// Parse output into structured format
	commits := []map[string]string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		commits = append(commits, map[string]string{
			"raw": line,
		})
	}

	return map[string]interface{}{
		"success": true,
		"commits": commits,
		"count":   len(commits),
		"raw":     string(output),
	}, nil
}

func (t *GitHubListCommitsTool) RequiresConfirmation() bool {
	return false // Read-only operation
}
