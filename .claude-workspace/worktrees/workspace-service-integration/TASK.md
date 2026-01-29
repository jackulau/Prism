---
id: workspace-service-integration
name: Workspace Service Integration
wave: 4
priority: 4
dependencies:
- workspace-handler
estimated_hours: 2
tags:
- backend
- integration
- services
---

## Objective

Integrate the Workspace entity into the main application by wiring up dependencies in main.go and ensuring end-to-end functionality.

## Context

This is the final integration task that:
- Creates the repository instance in main.go
- Passes it to the Dependencies struct
- Ensures the full request flow works from API to database

This task also prepares for future integration with:
- Agent entity (One-to-Many: Workspace → Agent)
- WorkspaceTool entity (One-to-Many: Workspace → WorkspaceTool)

## Implementation

1. **Update `/backend/cmd/server/main.go`**:

   Add repository instantiation after other repos:

   ```go
   // After line ~65 where other repos are created:
   orgWorkspaceRepo := repository.NewWorkspaceRepository(db.DB)
   ```

   Add to Dependencies struct in routes.Setup call:

   ```go
   OrgWorkspaceRepo: orgWorkspaceRepo,
   ```

2. **Update `/backend/internal/api/routes/router.go`**:

   Add field to Dependencies struct:

   ```go
   type Dependencies struct {
       // ... existing fields
       OrgWorkspaceRepo *repository.WorkspaceRepository
   }
   ```

3. **Verify end-to-end flow**:

   After integration, test manually:

   ```bash
   # Start server
   cd backend && go run cmd/server/main.go

   # Create workspace (requires auth token)
   curl -X POST http://localhost:8080/api/v1/org/workspaces \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>" \
     -d '{"name": "test-workspace", "organization_id": "org-123"}'

   # List workspaces
   curl http://localhost:8080/api/v1/org/workspaces?organization_id=org-123 \
     -H "Authorization: Bearer <token>"
   ```

4. **Add WebSocket message types** (optional, for future agent integration):

   In `/backend/internal/api/routes/router.go`, prepare message type constants:

   ```go
   const (
       // ... existing types
       TypeWorkspaceCreate = "workspace:create"
       TypeWorkspaceUpdate = "workspace:update"
       TypeWorkspaceList   = "workspace:list"
   )
   ```

5. **Document relation points** for future tasks:

   Add comments in org_workspace.go noting where Agent and WorkspaceTool relations will connect:

   ```go
   // Workspace has One-to-Many relation with Agent
   // When Agent entity is created, add:
   // - ListAgentsByWorkspaceID(workspaceID string) ([]*Agent, error)
   // - GetWorkspaceByAgentID(agentID string) (*Workspace, error)

   // Workspace has One-to-Many relation with WorkspaceTool
   // When WorkspaceTool entity is created, add:
   // - ListToolsByWorkspaceID(workspaceID string) ([]*WorkspaceTool, error)
   ```

## Acceptance Criteria

- [ ] OrgWorkspaceRepo instantiated in main.go
- [ ] OrgWorkspaceRepo passed to routes.Dependencies
- [ ] Server compiles and starts without errors
- [ ] POST /api/v1/org/workspaces creates workspace in database
- [ ] GET /api/v1/org/workspaces lists workspaces by organization
- [ ] GET /api/v1/org/workspaces/:id retrieves single workspace
- [ ] PATCH /api/v1/org/workspaces/:id updates workspace
- [ ] DELETE /api/v1/org/workspaces/:id removes workspace
- [ ] All endpoints return appropriate HTTP status codes
- [ ] Database persists data across server restarts
- [ ] Comments document future Agent/WorkspaceTool integration points

## Files to Create/Modify

- `backend/cmd/server/main.go` - Add repository instantiation
- `backend/internal/api/routes/router.go` - Update Dependencies struct

## Integration Points

- **Provides**: Full working Workspace entity with API
- **Consumes**: All previous tasks (schema, repository, handler)
- **Conflicts**: Modifies shared files (main.go, router.go) - minor additions only
