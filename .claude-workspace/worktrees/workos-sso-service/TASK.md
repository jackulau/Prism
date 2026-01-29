---
id: workos-sso-service
name: WorkOS SSO Authorization and Callback Logic
wave: 2
priority: 2
dependencies:
- workos-sdk-config
estimated_hours: 3
tags:
- backend
- auth
- workos
- sso
---

## Objective

Implement the core WorkOS SSO authorization flow: URL generation, callback handling, and profile extraction.

## Context

This task implements the SSO business logic in the WorkOS service. It handles generating authorization URLs for SSO providers and processing callbacks to extract user profiles.

**Reference Implementation Pattern:**
The existing GitHub OAuth implementation in `backend/internal/api/handlers/oauth.go` provides a similar pattern:
- State token generation for CSRF protection
- Authorization URL construction
- Callback processing with code exchange
- User profile extraction

## Implementation

1. **Extend WorkOS Service** (`backend/internal/security/workos.go`)

   Add SSO methods:
   ```go
   // GenerateAuthorizationURL creates an SSO authorization URL
   // Parameters: organizationID or connectionID, redirectURI, state
   func (s *WorkOSService) GenerateAuthorizationURL(opts AuthorizationOptions) (string, error)

   // HandleCallback processes the SSO callback and returns user profile
   func (s *WorkOSService) HandleCallback(code string) (*SSOProfile, error)

   // SSOProfile contains user info from WorkOS SSO
   type SSOProfile struct {
       ID             string
       Email          string
       FirstName      string
       LastName       string
       OrganizationID string
       ConnectionID   string
       IdpID          string
       RawAttributes  map[string]interface{}
   }
   ```

2. **Add State Management** (`backend/internal/security/workos.go`)
   - Create state store for CSRF protection (similar to OAuth handler pattern)
   - Generate random state tokens (16 bytes)
   - 10-minute TTL for state tokens
   - One-time use verification
   - Background cleanup goroutine

3. **Session Cookie Support** (`backend/internal/security/workos.go`)
   - Implement `wos-session` cookie encryption/decryption
   - Use WORKOS_COOKIE_PASSWORD for cookie encryption
   - Cookie contains: user_id, organization_id, expires_at
   - Secure, HttpOnly, SameSite=Lax attributes

4. **Error Handling**
   - Handle WorkOS API errors gracefully
   - Map WorkOS errors to appropriate HTTP status codes
   - Log errors with context for debugging

## Acceptance Criteria

- [ ] SSO authorization URLs can be generated
- [ ] State tokens are created and validated
- [ ] SSO callbacks are processed successfully
- [ ] User profiles are extracted from WorkOS response
- [ ] Session cookies can be created and read
- [ ] CSRF protection is properly implemented
- [ ] Error cases are handled gracefully

## Files to Create/Modify

**Modify:**
- `backend/internal/security/workos.go` - Add SSO methods

## Integration Points

- **Provides**: SSO authorization and callback logic for handlers
- **Consumes**: WorkOS SDK, config from workos-sdk-config
- **Conflicts**: None - extends workos.go file created in dependency
