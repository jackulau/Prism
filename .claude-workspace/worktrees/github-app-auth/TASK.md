---
id: github-app-auth
name: GitHub App Authentication System
wave: 1
priority: 1
dependencies: []
estimated_hours: 5
tags:
- backend
- auth
- github
---

## Objective

Implement full GitHub App authentication to enable repository access without requiring users to manually configure OAuth. GitHub Apps provide better security, finer-grained permissions, and automatic token refresh.

## Context

The codebase currently has:
- OAuth-based GitHub integration in `backend/internal/api/handlers/oauth.go`
- Webhook handling with signature verification in `backend/internal/integrations/github/`
- Security services for encryption in `backend/internal/security/`

GitHub Apps authenticate via:
1. JWT tokens signed with the App's private key (for App-level API calls)
2. Installation access tokens (for repository-level API calls)

## Implementation

1. Create `backend/internal/integrations/github/app.go`:
   - `GitHubApp` struct with App ID and private key
   - `GenerateJWT()` - Create JWT signed with RS256
   - `GetInstallationToken(installationID)` - Exchange JWT for installation token
   - `GetRepositoryInstallation(owner, repo)` - Find installation for a repo
   - Token caching with expiration handling

2. Create `backend/internal/integrations/github/app_client.go`:
   - `AppClient` struct wrapping authenticated HTTP client
   - Methods: `ListRepositories()`, `GetRepository()`, `GetCommits()`, `GetPullRequests()`
   - Automatic token refresh on 401

3. Create database migration for GitHub App installations:
   - `backend/internal/database/repository/github_installation.go`
   - Store installation_id, account type, account login, permissions, repositories

4. Create API endpoints in `backend/internal/api/handlers/github_app.go`:
   - `GET /api/v1/github/app/installations` - List user's installations
   - `POST /api/v1/github/app/setup` - Handle GitHub App installation callback
   - `GET /api/v1/github/app/repos` - List repos accessible via App

5. Add configuration in `backend/internal/config/config.go`:
   - `GitHubAppID`, `GitHubAppPrivateKey`, `GitHubAppClientID`, `GitHubAppClientSecret`

## Acceptance Criteria

- [ ] JWT generation with RS256 signing using App private key
- [ ] Installation token retrieval and caching (tokens expire after 1 hour)
- [ ] Automatic token refresh on expiration
- [ ] Installation webhook handler for app installed/uninstalled events
- [ ] Database persistence of installations and their repository access
- [ ] API endpoints for listing installations and accessible repos
- [ ] Graceful fallback to OAuth when App not configured

## Files to Create/Modify

- `backend/internal/integrations/github/app.go` - **Create**: JWT and token generation
- `backend/internal/integrations/github/app_client.go` - **Create**: Authenticated API client
- `backend/internal/database/repository/github_installation.go` - **Create**: Installation persistence
- `backend/internal/database/migrations/` - **Create**: Installation table migration
- `backend/internal/api/handlers/github_app.go` - **Create**: API endpoints
- `backend/internal/api/routes/router.go` - **Modify**: Register new routes
- `backend/internal/config/config.go` - **Modify**: Add App configuration
- `backend/internal/integrations/github/webhook.go` - **Modify**: Handle installation events

## Integration Points

- **Provides**: GitHub App authentication for all GitHub API calls
- **Consumes**: Config (App credentials), EncryptionService, Database
- **Conflicts**: Avoid modifying OAuth handler logic - this is complementary

## GitHub App JWT Format

```go
// Header
{
  "alg": "RS256",
  "typ": "JWT"
}
// Payload
{
  "iat": <issued_at>,
  "exp": <expires_at>,  // Max 10 minutes
  "iss": <app_id>
}
```

## API Reference

```
# Get installation token
POST /app/installations/{installation_id}/access_tokens
Authorization: Bearer {jwt}

# List installations
GET /user/installations
Authorization: Bearer {user_token}

# List repos for installation
GET /installation/repositories
Authorization: Bearer {installation_token}
```
