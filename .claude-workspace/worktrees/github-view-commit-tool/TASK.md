---
id: github-view-commit-tool
name: GitHub Tool - View Commit (github_view_commit)
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

Implement the `github_view_commit` built-in tool that shows the full diff of a specific commit.

## Context

This tool allows AI assistants to view the detailed changes in a specific commit, including the full diff output. It executes `git show {sha}` to display the commit details and file changes.

## Implementation

### 1. Create or extend the tool file

**File:** `backend/internal/tools/builtin/github_commits.go`

Add the view commit tool (can be in same file as list commits):

```go
type GitHubViewCommitTool struct {
	sandbox *sandbox.Service
}

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

	// Validate SHA format (basic check)
	if len(commitSha) < 4 || len(commitSha) > 40 {
		return nil, fmt.Errorf("invalid commit SHA format")
	}

	// Execute git show command
	cmd := exec.CommandContext(ctx, "git", "show", commitSha)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("git show failed: %s - %s", err, string(output)),
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
```

### 2. Register the tool

**File:** `backend/internal/tools/builtin/init.go`

Add registration:

```go
if err := registry.Register(NewGitHubViewCommitTool(sandbox)); err != nil {
    return err
}
```

### 3. Add to read-only tools list

**File:** `backend/internal/tools/approval.go`

Add `"github_view_commit": true` to the `ReadOnlyTools` map.

## Acceptance Criteria

- [ ] Tool is registered and available in tool registry
- [ ] Returns full git show output including diff
- [ ] Validates commitSha parameter is provided
- [ ] Basic validation of SHA format (4-40 characters)
- [ ] Returns structured response with success/error status
- [ ] Marked as read-only (no confirmation required)
- [ ] Works within user's workspace directory
- [ ] Handles invalid SHA gracefully with error message

## Files to Create/Modify

- `backend/internal/tools/builtin/github_commits.go` - **MODIFY** - Add view commit tool
- `backend/internal/tools/builtin/init.go` - **MODIFY** - Register the tool
- `backend/internal/tools/approval.go` - **MODIFY** - Add to read-only tools

## Integration Points

- **Provides**: `github_view_commit` tool for viewing commit diffs
- **Consumes**: Sandbox service for workspace access
- **Conflicts**: Coordinates with `github-list-commits-tool` task on same file
