---
id: workos-sdk-config
name: WorkOS SDK Integration and Configuration
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- config
- workos
---

## Objective

Add WorkOS SDK dependency and configuration infrastructure to the backend.

## Context

This is the foundation task that sets up WorkOS integration. It adds the Go SDK, extends the configuration system with WorkOS environment variables, and creates the core WorkOS service with client initialization.

**Current Config System:**
- `backend/internal/config/config.go` - Environment variable loading
- Uses `getEnv()` helper functions with defaults
- Has validation for production security settings

## Implementation

1. **Add WorkOS SDK** (`backend/go.mod`)
   ```bash
   cd backend && go get github.com/workos/workos-go/v4
   ```

2. **Update Config** (`backend/internal/config/config.go`)
   - Add struct fields:
     ```go
     // WorkOS Configuration
     WorkOSAPIKey       string
     WorkOSClientID     string
     WorkOSRedirectURI  string
     WorkOSCookiePassword string
     ```
   - Add loading in `Load()` function
   - Add validation in `validateSecurity()` for production

3. **Create WorkOS Service** (`backend/internal/security/workos.go`)
   - Initialize WorkOS client with API key
   - Create `WorkOSService` struct with client reference
   - Add `NewWorkOSService(config *Config) *WorkOSService` constructor
   - Basic health check method to validate configuration

4. **Update .env.example** (`backend/.env.example`)
   - Add WorkOS configuration variables with documentation:
     ```
     # WorkOS SSO Configuration
     WORKOS_API_KEY=
     WORKOS_CLIENT_ID=
     WORKOS_REDIRECT_URI=http://localhost:8080/api/v1/auth/sso/callback
     WORKOS_COOKIE_PASSWORD=
     ```

## Acceptance Criteria

- [ ] WorkOS SDK added to go.mod and go.sum
- [ ] Config struct includes all WorkOS fields
- [ ] Environment variables are loaded correctly
- [ ] WorkOSService can be instantiated with config
- [ ] Production validation warns if WorkOS is incomplete
- [ ] .env.example documents all WorkOS variables

## Files to Create/Modify

**Create:**
- `backend/internal/security/workos.go` - WorkOS service (basic initialization only)

**Modify:**
- `backend/go.mod` - Add WorkOS dependency
- `backend/internal/config/config.go` - Add WorkOS config fields
- `backend/.env.example` - Document WorkOS variables

## Integration Points

- **Provides**: WorkOS SDK and configuration for other auth tasks
- **Consumes**: Existing config loading pattern
- **Conflicts**: None - pure addition
