---
id: workos-user-schema
name: WorkOS User Schema and Repository Extension
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- database
- workos
---

## Objective

Extend the user database schema and repository to support WorkOS SSO users with organization context.

## Context

The existing user model supports email/password and GitHub OAuth. WorkOS SSO requires additional fields to track the WorkOS user ID and organization membership.

**Current User Model** (`backend/internal/database/repository/user.go`):
```go
type User struct {
    ID                string
    Email             string
    PasswordHash      string
    GitHubToken       string
    GitHubUsername    string
    GitHubConnectedAt *time.Time
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

**Database Schema** (`backend/internal/database/sqlite.go`):
- Users table with email, password_hash, github fields
- Migration pattern: ALTER TABLE in initialization

## Implementation

1. **Add Database Migration** (`backend/internal/database/sqlite.go`)

   Add WorkOS columns to users table:
   ```sql
   ALTER TABLE users ADD COLUMN workos_id TEXT;
   ALTER TABLE users ADD COLUMN organization_id TEXT;
   ALTER TABLE users ADD COLUMN sso_connection_id TEXT;
   ALTER TABLE users ADD COLUMN sso_provider TEXT;
   ```

   Add index for WorkOS ID lookups:
   ```sql
   CREATE INDEX IF NOT EXISTS idx_users_workos_id ON users(workos_id);
   CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);
   ```

2. **Extend User Struct** (`backend/internal/database/repository/user.go`)
   ```go
   type User struct {
       // ... existing fields ...
       WorkOSID        string
       OrganizationID  string
       SSOConnectionID string
       SSOProvider     string  // "saml", "oidc", etc.
   }
   ```

3. **Add Repository Methods** (`backend/internal/database/repository/user.go`)
   ```go
   // GetByWorkOSID finds a user by their WorkOS SSO ID
   func (r *UserRepository) GetByWorkOSID(workosID string) (*User, error)

   // CreateFromSSO creates a new user from SSO profile
   func (r *UserRepository) CreateFromSSO(email, workosID, orgID, connectionID, provider string) (*User, error)

   // LinkWorkOSAccount links an existing user to WorkOS
   func (r *UserRepository) LinkWorkOSAccount(userID, workosID, orgID, connectionID, provider string) error

   // GetByOrganization returns all users in an organization
   func (r *UserRepository) GetByOrganization(orgID string) ([]*User, error)
   ```

4. **Update Existing Scans** (`backend/internal/database/repository/user.go`)
   - Update `scanUser()` helper to include new fields
   - Ensure NULL handling for optional WorkOS fields

## Acceptance Criteria

- [ ] Database migration adds WorkOS columns
- [ ] User struct includes WorkOS fields
- [ ] GetByWorkOSID method works correctly
- [ ] CreateFromSSO creates users with all fields
- [ ] LinkWorkOSAccount updates existing users
- [ ] GetByOrganization returns correct users
- [ ] Existing user queries still work
- [ ] NULL fields are handled properly

## Files to Create/Modify

**Modify:**
- `backend/internal/database/sqlite.go` - Add migration
- `backend/internal/database/repository/user.go` - Extend User struct and add methods

## Integration Points

- **Provides**: User storage with WorkOS fields for SSO handlers
- **Consumes**: Existing database patterns
- **Conflicts**: None - pure addition to existing schema
