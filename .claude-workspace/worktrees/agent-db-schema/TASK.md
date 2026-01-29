---
id: agent-db-schema
name: Agent Database Schema & Migration
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- database
---

## Objective

Create the SQLite database schema and migration for the Agent entity to persist AI execution instance data.

## Context

The Prism codebase uses SQLite with raw SQL migrations defined in `/backend/internal/database/sqlite.go`. This task adds the `agents` table following existing patterns from `conversations`, `workspace_todos`, and other tables.

## Implementation

1. **Add migration in `sqlite.go`** - Add new migration to the `migrations` array in the `Migrate()` function

2. **Schema Definition**:
```sql
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    workspace_id TEXT REFERENCES user_workspaces(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'PENDING',
    provider_type TEXT NOT NULL,
    conversation_id TEXT,
    url TEXT,
    github_branch_name TEXT,
    name TEXT NOT NULL,
    model TEXT,
    sandbox_id TEXT,
    is_orchestrator_agent INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agents_workspace_id ON agents(workspace_id);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_created_at ON agents(created_at DESC);
```

3. **Status Values**: `PENDING`, `RUNNING`, `COMPLETED`, `FAILED`

4. **Provider Types**: `PRISM`, `CURSOR`, `JULES`

## Acceptance Criteria

- [ ] Migration added to `sqlite.go` Migrate() function
- [ ] Table created with all required fields from F2.1 spec
- [ ] Foreign key to `user_workspaces` for workspace_id
- [ ] Indexes for common query patterns (workspace_id, status, created_at)
- [ ] Nullable fields use TEXT/NULL properly
- [ ] Boolean `is_orchestrator_agent` stored as INTEGER (0/1)
- [ ] Timestamps default to CURRENT_TIMESTAMP

## Files to Create/Modify

- `backend/internal/database/sqlite.go` - Add migration

## Integration Points

- **Provides**: `agents` table schema for repository layer
- **Consumes**: `user_workspaces` table (foreign key reference)
- **Conflicts**: None - isolated database migration
