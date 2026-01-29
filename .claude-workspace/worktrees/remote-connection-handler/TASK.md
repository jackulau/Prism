---
id: remote-connection-handler
name: Remote Connection Handler & Session Manager
wave: 3
priority: 3
dependencies:
- remote-auth-service
- tunnel-proxy-server
estimated_hours: 4
tags:
- backend
- networking
- session
---

## Objective

Create the connection handler that manages the lifecycle of remote sessions, including heartbeats, reconnection, and session persistence.

## Context

Remote connections need robust session management to handle:
- Intermittent connectivity (mobile, unstable networks)
- Session resumption without re-authentication
- Heartbeat monitoring for connection health
- Activity tracking for idle timeout

## Implementation

1. **Create session manager** in `backend/internal/remote/session_manager.go`:
   ```go
   type SessionManager struct {
       sessions       map[string]*RemoteSession
       reconnectTokens map[string]string  // reconnect token -> session ID
       mu             sync.RWMutex
   }
   
   type RemoteSession struct {
       ID              string
       AuthSession     *security.RemoteSession
       Connection      net.Conn
       LastActivity    time.Time
       LastHeartbeat   time.Time
       ReconnectToken  string
       ReconnectExpiry time.Time
       State           SessionState
   }
   
   type SessionState int
   const (
       StateActive SessionState = iota
       StateDisconnected
       StateReconnecting
       StateClosed
   )
   ```

2. **Implement heartbeat protocol**:
   - Client sends heartbeat every 30 seconds
   - Server responds with acknowledgment
   - Detect stale connections after 90 seconds
   - Mark session as disconnected (not closed) on timeout

3. **Implement reconnection flow**:
   - Generate reconnect token on initial connection
   - Token valid for 5 minutes after disconnect
   - Resume session without re-authentication
   - Merge pending messages on reconnect

4. **Activity tracking**:
   - Track last activity timestamp
   - Separate idle timeout from heartbeat timeout
   - Send idle warning before timeout
   - Graceful session close on idle timeout

5. **Create management API**:
   ```
   GET  /api/v1/remote/sessions        - List active sessions (admin)
   POST /api/v1/remote/sessions/:id/kick - Force disconnect session
   GET  /api/v1/remote/sessions/stats  - Connection statistics
   ```

6. **Add WebSocket message types** for session management:
   ```go
   // Message types
   RemoteHeartbeat       = "remote_heartbeat"
   RemoteHeartbeatAck    = "remote_heartbeat_ack"
   RemoteIdleWarning     = "remote_idle_warning"
   RemoteSessionExpiring = "remote_session_expiring"
   RemoteDisconnect      = "remote_disconnect"
   ```

## Acceptance Criteria

- [ ] Sessions tracked with proper lifecycle states
- [ ] Heartbeat protocol detects dead connections
- [ ] Reconnection works within token validity window
- [ ] Activity tracking triggers idle warnings
- [ ] Management API allows session inspection
- [ ] Graceful cleanup of abandoned sessions

## Files to Create/Modify

- `backend/internal/remote/session_manager.go` - Session lifecycle management
- `backend/internal/remote/heartbeat.go` - Heartbeat protocol
- `backend/internal/remote/reconnect.go` - Reconnection logic
- `backend/internal/api/handlers/remote_admin.go` - Admin API endpoints

## Integration Points

- **Provides**: Session management for WebSocket connections
- **Consumes**: remote-auth-service, tunnel-proxy-server
- **Conflicts**: Minor changes to ws hub for message types
