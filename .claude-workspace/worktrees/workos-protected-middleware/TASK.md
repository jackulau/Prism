---
id: workos-protected-middleware
name: WorkOS Session Middleware for Protected Routes
wave: 2
priority: 2
dependencies:
- workos-sdk-config
estimated_hours: 2
tags:
- backend
- middleware
- workos
- auth
---

## Objective

Create middleware that validates WorkOS session cookies and extracts organization context for protected routes.

## Context

The existing auth middleware (`backend/internal/api/middleware/auth.go`) validates JWT tokens and sets user context. WorkOS SSO adds the `wos-session` cookie which contains organization-level context.

**Current Auth Middleware Pattern:**
```go
func AuthMiddleware(jwtService *security.JWTService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Extract token from Authorization header
        // Validate JWT
        // Set userID, email in context
        // Call c.Next()
    }
}
```

## Implementation

1. **Create WorkOS Session Middleware** (`backend/internal/api/middleware/workos.go`)

   ```go
   // WorkOSSessionMiddleware validates wos-session cookie and sets org context
   func WorkOSSessionMiddleware(workosService *security.WorkOSService) fiber.Handler {
       return func(c *fiber.Ctx) error {
           // Check for wos-session cookie
           cookie := c.Cookies("wos-session")
           if cookie == "" {
               return c.Next() // No session, continue
           }

           // Decrypt and validate session
           session, err := workosService.DecryptSession(cookie)
           if err != nil {
               // Invalid session, clear cookie and continue
               c.Cookie(&fiber.Cookie{
                   Name:     "wos-session",
                   Value:    "",
                   Expires:  time.Now().Add(-time.Hour),
                   HTTPOnly: true,
               })
               return c.Next()
           }

           // Check expiration
           if session.ExpiresAt.Before(time.Now()) {
               return c.Next()
           }

           // Set organization context
           c.Locals("organizationID", session.OrganizationID)
           c.Locals("ssoSessionID", session.ID)

           return c.Next()
       }
   }
   ```

2. **Add Session Types** (`backend/internal/security/workos.go`)

   ```go
   type WorkOSSession struct {
       ID             string
       UserID         string
       OrganizationID string
       ConnectionID   string
       ExpiresAt      time.Time
       CreatedAt      time.Time
   }

   // CreateSession creates an encrypted session cookie value
   func (s *WorkOSService) CreateSession(session *WorkOSSession) (string, error)

   // DecryptSession decrypts and parses a session cookie
   func (s *WorkOSService) DecryptSession(cookieValue string) (*WorkOSSession, error)
   ```

3. **Add Context Helper Functions** (`backend/internal/api/middleware/workos.go`)

   ```go
   // GetOrganizationID returns the organization ID from context
   func GetOrganizationID(c *fiber.Ctx) string {
       if orgID, ok := c.Locals("organizationID").(string); ok {
           return orgID
       }
       return ""
   }

   // HasOrganizationContext checks if request has org context
   func HasOrganizationContext(c *fiber.Ctx) bool {
       return GetOrganizationID(c) != ""
   }
   ```

4. **Create Organization-Required Middleware** (`backend/internal/api/middleware/workos.go`)

   ```go
   // RequireOrganization ensures request has organization context
   func RequireOrganization() fiber.Handler {
       return func(c *fiber.Ctx) error {
           if !HasOrganizationContext(c) {
               return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                   "error": "Organization context required",
               })
           }
           return c.Next()
       }
   }
   ```

5. **Update Router** (`backend/internal/api/routes/router.go`)

   Apply middleware globally (after auth middleware):
   ```go
   // Add after SecurityHeaders middleware
   if deps.WorkOSService != nil {
       app.Use(middleware.WorkOSSessionMiddleware(deps.WorkOSService))
   }
   ```

## Acceptance Criteria

- [ ] Middleware extracts session from wos-session cookie
- [ ] Invalid sessions are cleared and ignored
- [ ] Expired sessions are ignored
- [ ] Organization context is set in request locals
- [ ] Helper functions retrieve org context correctly
- [ ] RequireOrganization middleware blocks requests without org
- [ ] Existing JWT auth continues to work
- [ ] Middleware is registered in router

## Files to Create/Modify

**Create:**
- `backend/internal/api/middleware/workos.go` - WorkOS session middleware

**Modify:**
- `backend/internal/security/workos.go` - Add session encryption methods
- `backend/internal/api/routes/router.go` - Register middleware

## Integration Points

- **Provides**: Organization context for routes that need it
- **Consumes**: WorkOS service for session decryption
- **Conflicts**: None - additive middleware that works alongside existing auth
