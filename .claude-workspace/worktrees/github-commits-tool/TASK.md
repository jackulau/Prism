---
id: github-commits-tool
name: GitHub Commits Tool for LLM
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

Create a new LLM tool that allows the AI to fetch commit history from GitHub repositories using the user's OAuth token.

## Context

The codebase already has:
- Tool interface pattern in `backend/internal/tools/registry.go`
- GitHub OAuth integration storing tokens per user
- GitHub types including `Commit` struct in `backend/internal/integrations/github/types.go`
- Existing tools registered in `backend/internal/tools/builtin/init.go`

This tool enables the AI assistant to retrieve commit information for debugging, code review, and understanding project history.

## Implementation

1. Create `backend/internal/tools/builtin/github_commits.go`:
   - Implement `GitHubCommitsTool` struct
   - Accept parameters: `repo` (owner/name), `branch` (optional), `limit` (optional, default 10)
   - Use GitHub REST API: `GET /repos/{owner}/{repo}/commits`
   - Extract OAuth token from user's stored integration

2. Update `backend/internal/tools/builtin/init.go`:
   - Add `GitHubCommitsTool` to `RegisterAll()` function
   - Inject required dependencies (IntegrationRepository, EncryptionService)

3. Add to auto-approval list in `backend/internal/tools/approval.go`:
   - Add `"github_commits"` to read-only tools list

## Acceptance Criteria

- [ ] Tool accepts `repo`, `branch`, `limit` parameters
- [ ] Tool fetches commits from GitHub API using stored OAuth token
- [ ] Tool returns commit SHA, message, author, date, and files changed
- [ ] Tool handles errors gracefully (no token, repo not found, rate limit)
- [ ] Tool is auto-approved as read-only operation
- [ ] Tool respects rate limits and returns helpful error on 403

## Files to Create/Modify

- `backend/internal/tools/builtin/github_commits.go` - **Create**: New tool implementation
- `backend/internal/tools/builtin/init.go` - **Modify**: Register the new tool
- `backend/internal/tools/approval.go` - **Modify**: Add to read-only list

## Integration Points

- **Provides**: GitHub commit history access for LLM
- **Consumes**: IntegrationRepository (OAuth tokens), EncryptionService (decrypt tokens)
- **Conflicts**: None - new file, minimal changes to existing files

## Example Usage

```json
{
  "tool": "github_commits",
  "params": {
    "repo": "anthropics/claude-code",
    "branch": "main",
    "limit": 5
  }
}
```

## API Reference

GitHub REST API endpoint:
```
GET /repos/{owner}/{repo}/commits
Headers: Authorization: Bearer {token}
Query params: sha (branch), per_page (limit)
```
