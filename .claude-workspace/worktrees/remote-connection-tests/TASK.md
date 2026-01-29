---
id: remote-connection-tests
name: Remote Connection Integration Tests
wave: 4
priority: 4
dependencies:
- remote-auth-service
- tunnel-proxy-server
- remote-connection-handler
estimated_hours: 4
tags:
- testing
- integration
- security
---

## Objective

Create comprehensive integration tests for the remote connection feature including authentication, TLS, proxying, and session management.

## Context

Remote access is a security-critical feature that needs thorough testing:
- Authentication edge cases
- TLS configuration validation
- Proxy behavior under various conditions
- Session lifecycle correctness
- Error handling and recovery

## Implementation

1. **Create auth tests** in `backend/internal/security/remote_auth_test.go`:
   ```go
   func TestRemoteAuthService_Authenticate(t *testing.T)
   func TestRemoteAuthService_FailedLoginTracking(t *testing.T)
   func TestRemoteAuthService_IPBlocking(t *testing.T)
   func TestRemoteAuthService_SessionExpiry(t *testing.T)
   func TestRemoteAuthService_ConcurrentSessions(t *testing.T)
   func TestRemoteAuthService_PasswordValidation(t *testing.T)
   ```

2. **Create tunnel server tests** in `backend/internal/remote/tunnel_server_test.go`:
   ```go
   func TestTunnelServer_TLSHandshake(t *testing.T)
   func TestTunnelServer_HTTPProxy(t *testing.T)
   func TestTunnelServer_WebSocketProxy(t *testing.T)
   func TestTunnelServer_ConnectionLimits(t *testing.T)
   func TestTunnelServer_GracefulShutdown(t *testing.T)
   func TestTunnelServer_InvalidAuth(t *testing.T)
   ```

3. **Create session manager tests** in `backend/internal/remote/session_manager_test.go`:
   ```go
   func TestSessionManager_CreateSession(t *testing.T)
   func TestSessionManager_Heartbeat(t *testing.T)
   func TestSessionManager_Reconnect(t *testing.T)
   func TestSessionManager_IdleTimeout(t *testing.T)
   func TestSessionManager_SessionCleanup(t *testing.T)
   func TestSessionManager_ConcurrentAccess(t *testing.T)
   ```

4. **Create E2E test scenario** in `backend/internal/remote/e2e_test.go`:
   ```go
   func TestE2E_RemoteConnection(t *testing.T) {
       // 1. Start tunnel server with test certs
       // 2. Connect as remote client
       // 3. Authenticate with password
       // 4. Make proxied HTTP request
       // 5. Establish WebSocket connection
       // 6. Send/receive messages
       // 7. Simulate disconnect
       // 8. Reconnect with token
       // 9. Verify session restored
       // 10. Clean disconnect
   }
   ```

5. **Create security tests** in `backend/internal/security/remote_security_test.go`:
   ```go
   func TestSecurity_PasswordEncryption(t *testing.T)
   func TestSecurity_BruteForceProtection(t *testing.T)
   func TestSecurity_TLSCipherSuites(t *testing.T)
   func TestSecurity_SessionTokenEntropy(t *testing.T)
   func TestSecurity_IPWhitelist(t *testing.T)
   ```

6. **Add test utilities** in `backend/internal/remote/testutil_test.go`:
   - Test TLS certificate generation
   - Mock remote client
   - Test fixtures for sessions
   - Helper for concurrent testing

7. **Add Makefile targets**:
   ```makefile
   test-remote:
       cd backend && go test -v ./internal/remote/... ./internal/security/remote_*
   
   test-remote-integration:
       cd backend && go test -v -tags=integration ./internal/remote/...
   ```

## Acceptance Criteria

- [ ] All auth service tests pass
- [ ] All tunnel server tests pass
- [ ] All session manager tests pass
- [ ] E2E test demonstrates full flow
- [ ] Security tests verify protection mechanisms
- [ ] Test coverage > 80% for remote package
- [ ] Tests run in CI (when CI exists)

## Files to Create/Modify

- `backend/internal/security/remote_auth_test.go` - Auth tests
- `backend/internal/remote/tunnel_server_test.go` - Tunnel tests
- `backend/internal/remote/session_manager_test.go` - Session tests
- `backend/internal/remote/e2e_test.go` - Integration tests
- `backend/internal/security/remote_security_test.go` - Security tests
- `backend/internal/remote/testutil_test.go` - Test utilities
- `Makefile` - Add test targets

## Integration Points

- **Provides**: Test coverage for remote access feature
- **Consumes**: All remote access components
- **Conflicts**: None - test files only
