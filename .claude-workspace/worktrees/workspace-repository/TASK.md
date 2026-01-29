---
id: workspace-repository
name: Workspace Repository Implementation
wave: 2
priority: 2
dependencies:
- workspace-entity-schema
estimated_hours: 3
tags:
- backend
- database
- repository
---

## Objective

Create the WorkspaceRepository with full CRUD operations and query methods for the organization-scoped Workspace entity.

## Context

This repository follows the existing patterns in `/backend/internal/database/repository/`. It will provide data access for the new Workspace entity, supporting operations needed by agents, API handlers, and integrations.

The repository must handle:
- Basic CRUD operations
- Querying by organization
- Querying by GitHub repository
- Querying by branch
- Optional field handling (Slack integration, worker template)

## Implementation

1. **Create `/backend/internal/database/repository/org_workspace.go`**:

   ```go
   package repository

   import (
       "database/sql"
       "fmt"
       "time"

       "github.com/google/uuid"
   )

   // Workspace represents an organization-scoped workspace for agent sessions
   type Workspace struct {
       ID                   string
       Name                 string
       OrganizationID       string
       GitHubRepositoryName string     // Optional
       WorkerID             string     // Optional
       CurrentBranch        string     // Optional
       SlackChannelID       string     // Optional
       SlackMessageTs       string     // Optional
       CreatedAt            time.Time
   }

   type WorkspaceRepository struct {
       db *sql.DB
   }

   func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
       return &WorkspaceRepository{db: db}
   }
   ```

2. **Implement CRUD methods**:

   - `Create(workspace *Workspace) (*Workspace, error)` - Insert new workspace
   - `GetByID(id string) (*Workspace, error)` - Retrieve by primary key
   - `Update(workspace *Workspace) error` - Update all fields
   - `Delete(id string) error` - Delete by ID

3. **Implement query methods**:

   - `ListByOrganizationID(orgID string) ([]*Workspace, error)` - List all workspaces for an org
   - `GetByGitHubRepo(orgID, repoName string) (*Workspace, error)` - Find by repo name within org
   - `ListByBranch(branch string) ([]*Workspace, error)` - Find workspaces on a specific branch
   - `UpdateBranch(id, branch string) error` - Update only the current_branch field
   - `UpdateSlackInfo(id, channelID, messageTs string) error` - Update Slack integration fields

4. **Handle nullable fields properly**:

   ```go
   // Use sql.NullString for scanning optional fields
   var githubRepo, workerID, currentBranch sql.NullString
   var slackChannelID, slackMessageTs sql.NullString

   err := row.Scan(&ws.ID, &ws.Name, &ws.OrganizationID,
       &githubRepo, &workerID, &currentBranch,
       &slackChannelID, &slackMessageTs, &ws.CreatedAt)

   if githubRepo.Valid {
       ws.GitHubRepositoryName = githubRepo.String
   }
   // ... handle other nullable fields
   ```

5. **Follow existing patterns from**:
   - `workspace.go` - For list/query patterns
   - `user.go` - For CRUD structure
   - `conversation.go` - For nullable field handling

## Acceptance Criteria

- [ ] WorkspaceRepository struct created with db field
- [ ] Constructor function NewWorkspaceRepository implemented
- [ ] Create method generates UUID and inserts workspace
- [ ] GetByID retrieves workspace or returns nil if not found
- [ ] Update method updates all mutable fields
- [ ] Delete method removes workspace by ID
- [ ] ListByOrganizationID returns all workspaces for org
- [ ] GetByGitHubRepo finds workspace by repo name within org
- [ ] ListByBranch returns workspaces filtered by branch
- [ ] UpdateBranch updates only the branch field
- [ ] UpdateSlackInfo updates Slack-related fields
- [ ] All nullable fields handled with sql.NullString
- [ ] Error handling follows existing patterns (wrap with context)
- [ ] No SQL injection vulnerabilities (parameterized queries)

## Files to Create/Modify

- `backend/internal/database/repository/org_workspace.go` - Create new file

## Integration Points

- **Provides**: WorkspaceRepository for handlers and services
- **Consumes**: `workspaces` table from workspace-entity-schema task
- **Conflicts**: None - new file
