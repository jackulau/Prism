---
id: tool-entity
name: Tool Entity Database Model
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- database
- models
---

## Objective

Create the Tool entity database model for the tool catalog system with proper schema, repository, and API endpoints.

## Context

The Tool entity serves as a catalog of available tools in the system. It supports both model configurations (like LLM settings) and actual tools. Each tool has a unique slug identifier and can optionally be linked to an integration provider.

## Implementation

1. **Database Schema** - Add to `backend/internal/database/sqlite.go`:
   ```sql
   CREATE TABLE IF NOT EXISTS tools (
       id TEXT PRIMARY KEY,
       display_name TEXT NOT NULL,
       slug_name TEXT NOT NULL,
       description TEXT,
       is_model BOOLEAN DEFAULT FALSE,
       provider_id TEXT,
       parameters_schema TEXT,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       UNIQUE(display_name, provider_id, slug_name)
   );
   CREATE INDEX idx_tools_slug ON tools(slug_name);
   CREATE INDEX idx_tools_provider ON tools(provider_id);
   ```

2. **Repository** - Create `backend/internal/database/repository/tool.go`:
   - `Tool` struct with all fields
   - `ToolRepository` with CRUD operations:
     - `Create(tool *Tool) error`
     - `GetByID(id string) (*Tool, error)`
     - `GetBySlug(slug string) (*Tool, error)`
     - `List() ([]*Tool, error)`
     - `ListByProvider(providerID string) ([]*Tool, error)`
     - `ListModels() ([]*Tool, error)` - where is_model = true
     - `Update(tool *Tool) error`
     - `Delete(id string) error`

3. **API Handler** - Create `backend/internal/api/handlers/tools_catalog.go`:
   - `GET /api/tools` - List all tools
   - `GET /api/tools/:id` - Get tool by ID
   - `POST /api/tools` - Create tool (admin)
   - `PUT /api/tools/:id` - Update tool (admin)
   - `DELETE /api/tools/:id` - Delete tool (admin)

4. **Routes** - Add to `backend/internal/api/routes/routes.go`:
   - Register tool catalog routes with authentication middleware

## Acceptance Criteria

- [ ] Database migration creates tools table with unique constraint
- [ ] Repository implements all CRUD operations
- [ ] API endpoints work correctly with proper authentication
- [ ] Unique constraint on [displayName, provider, slugName] is enforced
- [ ] Index on slug_name for efficient lookups
- [ ] No breaking changes to existing functionality

## Files to Create/Modify

- `backend/internal/database/sqlite.go` - Add tools table migration
- `backend/internal/database/repository/tool.go` - Create (new file)
- `backend/internal/api/handlers/tools_catalog.go` - Create (new file)
- `backend/internal/api/routes/routes.go` - Add tool routes

## Integration Points

- **Provides**: Tool entity for dynamic tool building and tool management
- **Consumes**: Database connection, authentication middleware
- **Conflicts**: None - new entity, avoid modifying existing tool registry
