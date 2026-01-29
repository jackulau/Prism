---
id: workos-api-handlers
name: WorkOS SSO HTTP Handlers and Routes
wave: 3
priority: 3
dependencies:
- workos-sso-service
- workos-user-schema
estimated_hours: 3
tags:
- backend
- api
- workos
- handlers
---

## Objective

Create HTTP handlers for WorkOS SSO endpoints and register them in the router.

## Context

This task implements the HTTP layer for WorkOS SSO authentication. It creates handlers that use the WorkOS service for SSO logic and the user repository for persistence.

**Existing Pattern** (`backend/internal/api/handlers/auth.go`, `oauth.go`):
- Handler struct with dependencies injected
- NewXxxHandler() constructor
- Methods for each endpoint
- Fiber context handling

**Route Structure** (`backend/internal/api/routes/router.go`):
- Routes grouped by feature
- Auth middleware applied per group
- Dependencies injected via Dependencies struct

## Implementation

1. **Create SSO Handler** (`backend/internal/api/handlers/workos.go`)

   ```go
   type WorkOSHandler struct {
       workosService *security.WorkOSService
       userRepo      *repository.UserRepository
       jwtService    *security.JWTService
       sessionRepo   *repository.SessionRepository
   }

   func NewWorkOSHandler(
       workosService *security.WorkOSService,
       userRepo *repository.UserRepository,
       jwtService *security.JWTService,
       sessionRepo *repository.SessionRepository,
   ) *WorkOSHandler
   ```

2. **Implement SSO Authorize Endpoint**
   ```go
   // GET /api/v1/auth/sso/authorize
   // Query params: organization (domain or org ID)
   // Returns: { authorization_url: string }
   func (h *WorkOSHandler) Authorize(c *fiber.Ctx) error
   ```
   - Accept organization domain or ID
   - Generate state token
   - Build authorization URL
   - Return URL to client

3. **Implement SSO Callback Endpoint**
   ```go
   // GET /api/v1/auth/sso/callback
   // Query params: code, state
   // Redirects to frontend with tokens or error
   func (h *WorkOSHandler) Callback(c *fiber.Ctx) error
   ```
   - Validate state token (CSRF)
   - Exchange code for profile
   - Find or create user
   - Generate JWT tokens
   - Create session
   - Set wos-session cookie
   - Redirect to frontend

4. **Implement Get Connections Endpoint** (Optional, protected)
   ```go
   // GET /api/v1/auth/sso/connections
   // Returns: list of SSO connections for user's organization
   func (h *WorkOSHandler) GetConnections(c *fiber.Ctx) error
   ```

5. **Register Routes** (`backend/internal/api/routes/router.go`)

   Add to Dependencies struct:
   ```go
   WorkOSService *security.WorkOSService
   ```

   Add routes:
   ```go
   // SSO routes (public for authorize/callback)
   sso := auth.Group("/sso")
   sso.Get("/authorize", workosHandler.Authorize)
   sso.Get("/callback", workosHandler.Callback)

   // SSO management (protected)
   ssoProtected := auth.Group("/sso", middleware.AuthMiddleware(deps.JWTService))
   ssoProtected.Get("/connections", workosHandler.GetConnections)
   ```

6. **Update Server Initialization** (`backend/cmd/server/main.go`)
   - Create WorkOSService with config
   - Add to Dependencies struct
   - Pass to router setup

## Acceptance Criteria

- [ ] SSO authorize endpoint generates valid URLs
- [ ] SSO callback processes codes and creates sessions
- [ ] State token validation prevents CSRF
- [ ] Users are created or linked on first SSO login
- [ ] JWT tokens are issued after successful SSO
- [ ] wos-session cookie is set correctly
- [ ] Frontend redirect works after callback
- [ ] Error handling returns appropriate responses
- [ ] Routes are registered correctly

## Files to Create/Modify

**Create:**
- `backend/internal/api/handlers/workos.go` - SSO HTTP handlers

**Modify:**
- `backend/internal/api/routes/router.go` - Register SSO routes
- `backend/cmd/server/main.go` - Initialize WorkOS service

## Integration Points

- **Provides**: HTTP endpoints for SSO authentication
- **Consumes**: WorkOS service, user repository, JWT service
- **Conflicts**: Avoid modifying existing auth handlers (separate file)
