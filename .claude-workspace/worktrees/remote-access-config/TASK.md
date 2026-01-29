---
id: remote-access-config
name: Remote Access Configuration & Environment Setup
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- config
- security
---

## Objective

Create the configuration infrastructure for remote access including environment variables, secure defaults, and configuration validation.

## Context

This task establishes the foundation for remote access by defining configuration options for:
- Remote access port binding (separate from main server port)
- Encryption key management for remote passwords
- Connection timeout and session limits
- IP/host binding configuration
- TLS/SSL certificate paths for secure connections

## Implementation

1. **Update configuration struct** in `backend/internal/config/config.go`:
   - Add `RemoteAccessConfig` struct with fields:
     - `Enabled` (bool) - Enable/disable remote access
     - `Port` (int) - Port for remote connections (default: 8443)
     - `Host` (string) - Bind address (default: 0.0.0.0)
     - `PasswordHash` (string) - Argon2id hashed password for remote access
     - `EncryptedPassword` (string) - AES-256-GCM encrypted password (for validation)
     - `TLSCertPath` (string) - Path to TLS certificate
     - `TLSKeyPath` (string) - Path to TLS private key
     - `MaxConnections` (int) - Maximum concurrent remote connections
     - `SessionTimeout` (duration) - Remote session timeout
     - `AllowedIPs` ([]string) - IP whitelist (optional)

2. **Add environment variables** to `.env.example`:
   ```
   # Remote Access Configuration
   REMOTE_ACCESS_ENABLED=false
   REMOTE_ACCESS_PORT=8443
   REMOTE_ACCESS_HOST=0.0.0.0
   REMOTE_ACCESS_PASSWORD=  # Will be hashed on first use
   REMOTE_ACCESS_TLS_CERT=./certs/remote.crt
   REMOTE_ACCESS_TLS_KEY=./certs/remote.key
   REMOTE_ACCESS_MAX_CONNECTIONS=10
   REMOTE_ACCESS_SESSION_TIMEOUT=1h
   REMOTE_ACCESS_ALLOWED_IPS=  # Comma-separated, empty = allow all
   ```

3. **Create password initialization utility** in `backend/internal/security/remote_password.go`:
   - Function to hash password on first setup
   - Function to validate password against stored hash
   - Function to encrypt password for client-side storage

4. **Add configuration validation** in config loading:
   - Require TLS in production when remote access is enabled
   - Validate port is not same as main server port
   - Validate password strength requirements
   - Log warnings for insecure configurations

## Acceptance Criteria

- [ ] RemoteAccessConfig struct defined with all required fields
- [ ] Environment variables documented in .env.example
- [ ] Configuration loads from environment with sensible defaults
- [ ] Password hashing/validation utilities created
- [ ] Validation prevents insecure production configurations
- [ ] TLS requirement enforced in production mode

## Files to Create/Modify

- `backend/internal/config/config.go` - Add RemoteAccessConfig
- `backend/internal/security/remote_password.go` - Create password utilities
- `.env.example` - Add remote access configuration section
- `backend/cmd/server/main.go` - Add validation at startup

## Integration Points

- **Provides**: Configuration for remote-auth-service and tunnel-proxy-server
- **Consumes**: Existing crypto.go encryption utilities
- **Conflicts**: None - new configuration section
