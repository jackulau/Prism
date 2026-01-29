---
id: workspace-handler
name: Workspace API Handler
wave: 3
priority: 3
dependencies:
- workspace-repository
estimated_hours: 3
tags:
- backend
- api
- handlers
---

## Objective

Create HTTP API endpoints for managing organization-scoped Workspaces, following existing handler patterns.

## Context

This handler provides REST API endpoints for the Workspace entity. It will be used by:
- Frontend dashboard for workspace management
- Agent orchestration for creating/updating workspace state
- Integration services for Slack/GitHub coordination

**Note**: This is separate from the existing `workspace.go` handler which manages user filesystem workspaces. This handler manages organization-scoped workspaces for agent sessions.

## Implementation

1. **Create `/backend/internal/api/handlers/org_workspace.go`**:

   ```go
   package handlers

   import (
       "github.com/gofiber/fiber/v2"
       "github.com/jacklau/prism/internal/database/repository"
   )

   type OrgWorkspaceHandler struct {
       repo *repository.WorkspaceRepository
   }

   func NewOrgWorkspaceHandler(repo *repository.WorkspaceRepository) *OrgWorkspaceHandler {
       return &OrgWorkspaceHandler{repo: repo}
   }
   ```

2. **Implement request/response DTOs**:

   ```go
   type CreateWorkspaceRequest struct {
       Name                 string `json:"name"`
       OrganizationID       string `json:"organization_id"`
       GitHubRepositoryName string `json:"github_repository_name,omitempty"`
       WorkerID             string `json:"worker_id,omitempty"`
       CurrentBranch        string `json:"current_branch,omitempty"`
       SlackChannelID       string `json:"slack_channel_id,omitempty"`
   }

   type UpdateWorkspaceRequest struct {
       Name                 string `json:"name,omitempty"`
       GitHubRepositoryName string `json:"github_repository_name,omitempty"`
       WorkerID             string `json:"worker_id,omitempty"`
       CurrentBranch        string `json:"current_branch,omitempty"`
       SlackChannelID       string `json:"slack_channel_id,omitempty"`
       SlackMessageTs       string `json:"slack_message_ts,omitempty"`
   }

   type WorkspaceResponse struct {
       ID                   string `json:"id"`
       Name                 string `json:"name"`
       OrganizationID       string `json:"organization_id"`
       GitHubRepositoryName string `json:"github_repository_name,omitempty"`
       WorkerID             string `json:"worker_id,omitempty"`
       CurrentBranch        string `json:"current_branch,omitempty"`
       SlackChannelID       string `json:"slack_channel_id,omitempty"`
       SlackMessageTs       string `json:"slack_message_ts,omitempty"`
       CreatedAt            int64  `json:"created_at"`
   }
   ```

3. **Implement handler methods**:

   - `Create(c *fiber.Ctx) error` - POST /api/v1/org/workspaces
   - `Get(c *fiber.Ctx) error` - GET /api/v1/org/workspaces/:id
   - `List(c *fiber.Ctx) error` - GET /api/v1/org/workspaces?organization_id=...
   - `Update(c *fiber.Ctx) error` - PATCH /api/v1/org/workspaces/:id
   - `Delete(c *fiber.Ctx) error` - DELETE /api/v1/org/workspaces/:id
   - `UpdateBranch(c *fiber.Ctx) error` - PATCH /api/v1/org/workspaces/:id/branch

4. **Add validation**:

   ```go
   // In Create handler
   if req.Name == "" {
       return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
           "error": "name is required",
       })
   }
   if req.OrganizationID == "" {
       return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
           "error": "organization_id is required",
       })
   }
   ```

5. **Register routes in `/backend/internal/api/routes/router.go`**:

   ```go
   // In Setup function, add after other route groups
   orgWorkspaceHandler := handlers.NewOrgWorkspaceHandler(deps.WorkspaceRepo)

   orgWorkspaces := api.Group("/org/workspaces")
   orgWorkspaces.Use(authMiddleware)
   orgWorkspaces.Post("/", orgWorkspaceHandler.Create)
   orgWorkspaces.Get("/", orgWorkspaceHandler.List)
   orgWorkspaces.Get("/:id", orgWorkspaceHandler.Get)
   orgWorkspaces.Patch("/:id", orgWorkspaceHandler.Update)
   orgWorkspaces.Delete("/:id", orgWorkspaceHandler.Delete)
   orgWorkspaces.Patch("/:id/branch", orgWorkspaceHandler.UpdateBranch)
   ```

6. **Update Dependencies struct** in router.go to include WorkspaceRepository

## Acceptance Criteria

- [ ] OrgWorkspaceHandler struct created with repository dependency
- [ ] Constructor function NewOrgWorkspaceHandler implemented
- [ ] Create endpoint validates required fields and returns 201 with workspace
- [ ] Get endpoint returns workspace by ID or 404 if not found
- [ ] List endpoint filters by organization_id query param
- [ ] Update endpoint allows partial updates via PATCH
- [ ] Delete endpoint removes workspace and returns 204
- [ ] UpdateBranch endpoint updates only the branch field
- [ ] All endpoints require authentication (authMiddleware)
- [ ] Response DTOs use Unix timestamps for dates (consistency with existing handlers)
- [ ] Error responses follow existing JSON format
- [ ] Routes registered under /api/v1/org/workspaces prefix

## Files to Create/Modify

- `backend/internal/api/handlers/org_workspace.go` - Create new handler
- `backend/internal/api/routes/router.go` - Register routes and update Dependencies

## Integration Points

- **Provides**: REST API for workspace management
- **Consumes**: WorkspaceRepository from workspace-repository task
- **Conflicts**: None - new file and additive route changes
