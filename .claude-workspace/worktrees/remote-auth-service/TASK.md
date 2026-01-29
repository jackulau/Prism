---
id: remote-auth-service
name: Remote Authentication Service with Encrypted Password
wave: 2
priority: 2
dependencies:
- remote-access-config
estimated_hours: 4
tags:
- backend
- security
- authentication
---

## Objective

Implement a secure authentication service for remote connections using encrypted passwords with Argon2id hashing and AES-256-GCM encryption.

## Context

Remote access requires a separate authentication mechanism from the main user authentication:
- Single shared password for remote access (not per-user)
- Password encrypted in transit and at rest
- Time-limited authentication tokens for active sessions
- Brute-force protection with rate limiting and lockout

## Implementation

1. **Create remote auth service** in `backend/internal/security/remote_auth.go`:
   ```go
   type RemoteAuthService struct {
       config       *config.RemoteAccessConfig
       crypto       *CryptoService
       sessions     map[string]*RemoteSession  // token -> session
       failedLogins map[string]int             // ip -> count
       mu           sync.RWMutex
   }
   
   type RemoteSession struct {
       Token     string
       CreatedAt time.Time
       ExpiresAt time.Time
       ClientIP  string
       IsActive  bool
   }
   
   func (s *RemoteAuthService) Authenticate(password string, clientIP string) (*RemoteSession, error)
   func (s *RemoteAuthService) ValidateSession(token string) (*RemoteSession, error)
   func (s *RemoteAuthService) InvalidateSession(token string) error
   func (s *RemoteAuthService) CleanupExpiredSessions()
   func (s *RemoteAuthService) IsIPBlocked(ip string) bool
   ```

2. **Implement password validation flow**:
   - Client sends encrypted password (AES-256-GCM with shared nonce)
   - Server decrypts and validates against stored Argon2id hash
   - On success: generate secure session token (32 bytes random)
   - On failure: increment failed login counter for IP

3. **Add brute-force protection**:
   - Track failed login attempts per IP
   - Block IP after 5 failed attempts for 15 minutes
   - Exponential backoff for repeated blocks
   - Log all authentication attempts

4. **Session management**:
   - Generate cryptographically secure session tokens
   - Store sessions in memory with configurable timeout
   - Background goroutine for expired session cleanup
   - Maximum concurrent sessions limit

5. **Create remote auth middleware** in `backend/internal/api/middleware/remote_auth.go`:
   - Validate session token from header
   - Check session expiry and IP match
   - Rate limit authenticated requests

## Acceptance Criteria

- [ ] RemoteAuthService correctly validates passwords
- [ ] Session tokens are cryptographically secure
- [ ] Failed login tracking prevents brute force
- [ ] Sessions expire after configured timeout
- [ ] IP blocking works with exponential backoff
- [ ] All auth attempts are logged for audit
- [ ] Middleware properly protects remote endpoints

## Files to Create/Modify

- `backend/internal/security/remote_auth.go` - Create RemoteAuthService
- `backend/internal/api/middleware/remote_auth.go` - Create middleware
- `backend/internal/security/remote_auth_test.go` - Unit tests

## Integration Points

- **Provides**: Authentication for tunnel-proxy-server
- **Consumes**: remote-access-config, existing crypto.go
- **Conflicts**: None - new service
