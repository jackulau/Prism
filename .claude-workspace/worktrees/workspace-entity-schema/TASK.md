---
id: workspace-entity-schema
name: Workspace Entity Database Schema
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- database
- schema
---

## Objective

Create the database schema for the new organization-scoped Workspace entity with all required fields, indexes, and foreign key constraints.

## Context

The Workspace entity is a container for agent sessions linked to repositories. This differs from the existing `user_workspaces` table which is user-scoped for file system workspace tracking. This new Workspace entity:

- Is organization-scoped (linked to Organization via `organizationId`)
- Contains GitHub repository references
- Supports optional Slack integration
- Has One-to-Many relations with Agent and WorkspaceTool

**Important**: The Organization entity may not exist yet. The schema should use a TEXT foreign key field that can be constrained later when Organization is created.

## Implementation

1. **Add migration to `/backend/internal/database/sqlite.go`**:

   Add the following migration statements to the `migrations` slice in `Migrate()`:

   ```sql
   -- Create workspaces table (organization-scoped)
   CREATE TABLE IF NOT EXISTS workspaces (
       id TEXT PRIMARY KEY,
       name TEXT NOT NULL,
       organization_id TEXT NOT NULL,
       github_repository_name TEXT,
       worker_id TEXT,
       current_branch TEXT,
       slack_channel_id TEXT,
       slack_message_ts TEXT,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );

   -- Index on organization_id for listing workspaces by org
   CREATE INDEX IF NOT EXISTS idx_workspaces_organization_id ON workspaces(organization_id);

   -- Index on current_branch for branch-based queries
   CREATE INDEX IF NOT EXISTS idx_workspaces_current_branch ON workspaces(current_branch);

   -- Index on github_repository_name for repo lookups
   CREATE INDEX IF NOT EXISTS idx_workspaces_github_repo ON workspaces(github_repository_name);
   ```

2. **Field mapping from spec**:
   - `id` - TEXT PRIMARY KEY (UUID)
   - `name` - TEXT NOT NULL
   - `organizationId` → `organization_id` - TEXT NOT NULL (FK to organizations when created)
   - `githubRepositoryName` → `github_repository_name` - TEXT (nullable)
   - `workerId` → `worker_id` - TEXT (nullable, for worker template link)
   - `currentBranch` → `current_branch` - TEXT (nullable, indexed)
   - `slackChannelId` → `slack_channel_id` - TEXT (nullable)
   - `slackMessageTs` → `slack_message_ts` - TEXT (nullable)
   - `createdAt` → `created_at` - DATETIME DEFAULT CURRENT_TIMESTAMP

3. **Foreign key notes**:
   - `organization_id` will reference `organizations(id)` when Organization entity is created
   - For now, leave as unconstrained TEXT field
   - Add `ON DELETE CASCADE` constraint when Organization table exists

## Acceptance Criteria

- [ ] Migration added to sqlite.go in correct position (after existing migrations)
- [ ] Table created with all 9 fields matching spec
- [ ] 3 indexes created: organization_id, current_branch, github_repository_name
- [ ] Server starts successfully with new migration
- [ ] Migration is idempotent (can run multiple times without error)
- [ ] Field names follow existing snake_case convention

## Files to Create/Modify

- `backend/internal/database/sqlite.go` - Add migration statements

## Integration Points

- **Provides**: `workspaces` table for WorkspaceRepository
- **Consumes**: Will need `organizations` table FK when Organization entity created
- **Conflicts**: None - new table, no overlap with existing code
