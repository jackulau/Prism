---
id: tunnel-proxy-server
name: TLS Tunnel Proxy Server for Remote Connections
wave: 2
priority: 2
dependencies:
- remote-access-config
estimated_hours: 5
tags:
- backend
- networking
- security
---

## Objective

Create a TLS-secured proxy server that accepts remote connections on a configurable port and proxies authenticated requests to the main Prism server.

## Context

The tunnel proxy server is the entry point for remote access:
- Listens on a separate port (default 8443) with TLS
- Handles initial authentication handshake
- Proxies authenticated HTTP and WebSocket connections
- Manages connection lifecycle and limits

## Implementation

1. **Create tunnel server** in `backend/internal/remote/tunnel_server.go`:
   ```go
   type TunnelServer struct {
       config       *config.RemoteAccessConfig
       authService  *security.RemoteAuthService
       httpProxy    *httputil.ReverseProxy
       wsProxy      *websocketutil.Proxy
       connections  map[string]*TunnelConnection
       mu           sync.RWMutex
   }
   
   type TunnelConnection struct {
       ID        string
       ClientIP  string
       Session   *security.RemoteSession
       CreatedAt time.Time
       BytesIn   int64
       BytesOut  int64
   }
   
   func NewTunnelServer(cfg *config.RemoteAccessConfig, auth *security.RemoteAuthService) *TunnelServer
   func (t *TunnelServer) Start() error
   func (t *TunnelServer) Stop() error
   func (t *TunnelServer) GetActiveConnections() []*TunnelConnection
   ```

2. **Implement TLS listener**:
   - Load TLS certificate and key from configured paths
   - Support TLS 1.2+ only (no legacy protocols)
   - Strong cipher suite configuration
   - Optional client certificate validation

3. **Create authentication endpoints** on tunnel server:
   ```
   POST /remote/auth       - Authenticate with encrypted password
   POST /remote/logout     - Invalidate session
   GET  /remote/status     - Check connection status
   ```

4. **Implement HTTP reverse proxy**:
   - Proxy authenticated requests to main server
   - Add `X-Remote-Access: true` header
   - Add `X-Forwarded-For` with real client IP
   - Handle request/response streaming

5. **Implement WebSocket proxy**:
   - Upgrade WebSocket connections after auth
   - Proxy messages bidirectionally
   - Handle connection close gracefully
   - Track bytes transferred

6. **Connection management**:
   - Enforce maximum connection limit
   - Track active connections per IP
   - Graceful shutdown with connection draining
   - Health check endpoint

## Acceptance Criteria

- [ ] TLS server starts with valid certificate
- [ ] Authentication endpoint validates credentials
- [ ] HTTP requests proxied to main server correctly
- [ ] WebSocket connections proxied bidirectionally
- [ ] Connection limits enforced
- [ ] Graceful shutdown works properly
- [ ] All connections tracked with metrics

## Files to Create/Modify

- `backend/internal/remote/tunnel_server.go` - Create tunnel server
- `backend/internal/remote/http_proxy.go` - HTTP proxy logic
- `backend/internal/remote/ws_proxy.go` - WebSocket proxy logic
- `backend/internal/remote/connection.go` - Connection management
- `backend/cmd/server/main.go` - Start tunnel server if enabled

## Integration Points

- **Provides**: Remote access entry point for clients
- **Consumes**: remote-access-config, remote-auth-service
- **Conflicts**: None - separate port, new package
