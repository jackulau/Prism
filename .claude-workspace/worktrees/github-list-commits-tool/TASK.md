---
id: github-list-commits-tool
name: GitHub Tool - List Commits (github_list_commits)
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- tools
- github
---

## Objective

Implement the `github_list_commits` built-in tool that lists recent commits from a git repository.

## Context

This tool provides AI assistants with the ability to view recent commit history in the user's workspace. It executes `git log` with a configurable limit and returns formatted commit information including SHA, author, date, and message.

## Implementation

### 1. Create the tool file

**File:** `backend/internal/tools/builtin/github_commits.go`

Follow the existing tool pattern from `shell_exec.go`:

```go
package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

type GitHubListCommitsTool struct {
	sandbox *sandbox.Service
}

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
```

### 2. Register the tool

**File:** `backend/internal/tools/builtin/init.go`

Add registration in `RegisterAll()`:

```go
// GitHub tools for repository operations
if err := registry.Register(NewGitHubListCommitsTool(sandbox)); err != nil {
    return err
}
```

### 3. Add to read-only tools list

**File:** `backend/internal/tools/approval.go`

Add `"github_list_commits": true` to the `ReadOnlyTools` map for auto-approval.

## Acceptance Criteria

- [ ] Tool is registered and available in tool registry
- [ ] Returns formatted commit list with SHA, author, date, message
- [ ] Respects commitLimit parameter (1-10 range)
- [ ] Defaults to 5 commits when no limit specified
- [ ] Returns structured response with success/error status
- [ ] Marked as read-only (no confirmation required)
- [ ] Works within user's workspace directory
- [ ] Handles non-git directories gracefully with error message

## Files to Create/Modify

- `backend/internal/tools/builtin/github_commits.go` - **CREATE** - Main tool implementation
- `backend/internal/tools/builtin/init.go` - **MODIFY** - Register the tool
- `backend/internal/tools/approval.go` - **MODIFY** - Add to read-only tools

## Integration Points

- **Provides**: `github_list_commits` tool for viewing commit history
- **Consumes**: Sandbox service for workspace access
- **Conflicts**: None - new file creation
