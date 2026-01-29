---
id: remote-client-ui
name: Remote Connection Client UI & Settings
wave: 2
priority: 2
dependencies:
- remote-access-config
estimated_hours: 4
tags:
- frontend
- ui
- settings
---

## Objective

Create the frontend UI for configuring remote access, displaying connection information, and managing remote sessions.

## Context

Users need a way to:
- Enable/configure remote access from the settings page
- View connection details (IP, port, password)
- See and manage active remote connections
- Generate/regenerate access passwords
- View connection QR code for easy mobile setup

## Implementation

1. **Create remote access settings component** in `frontend/src/components/settings/RemoteAccessSettings.tsx`:
   ```tsx
   interface RemoteAccessSettingsProps {
     enabled: boolean
     onToggle: (enabled: boolean) => void
   }
   
   // Display:
   // - Enable/disable toggle
   // - Port configuration
   // - Password field with show/hide and copy
   // - Regenerate password button
   // - Connection URL display
   // - QR code for connection details
   // - TLS certificate status
   // - Allowed IPs configuration
   ```

2. **Create connection info display** in `frontend/src/components/settings/ConnectionInfo.tsx`:
   ```tsx
   // Display:
   // - Public IP address (auto-detected)
   // - Local IP addresses
   // - Port number
   // - Full connection URL
   // - One-click copy buttons
   // - QR code with encoded connection details
   ```

3. **Create active sessions view** in `frontend/src/components/settings/RemoteSessions.tsx`:
   ```tsx
   interface RemoteSession {
     id: string
     clientIP: string
     connectedAt: string
     lastActivity: string
     bytesIn: number
     bytesOut: number
   }
   
   // Display:
   // - Table of active connections
   // - Kick/disconnect button per session
   // - Connection duration
   // - Data transferred
   // - Refresh button
   ```

4. **Add remote access API service** in `frontend/src/services/remoteAccess.ts`:
   ```typescript
   export const remoteAccessApi = {
     getStatus: () => api.get('/remote/status'),
     enable: (config: RemoteConfig) => api.post('/remote/enable', config),
     disable: () => api.post('/remote/disable'),
     regeneratePassword: () => api.post('/remote/password/regenerate'),
     getSessions: () => api.get('/remote/sessions'),
     kickSession: (id: string) => api.post(`/remote/sessions/${id}/kick`),
     getConnectionInfo: () => api.get('/remote/connection-info'),
   }
   ```

5. **Add remote access store** in `frontend/src/store/remoteAccessStore.ts`:
   ```typescript
   interface RemoteAccessState {
     enabled: boolean
     port: number
     password: string | null
     connectionUrl: string | null
     sessions: RemoteSession[]
     loading: boolean
     error: string | null
   }
   ```

6. **Integrate into settings page**:
   - Add "Remote Access" section to settings
   - Show status indicator in header when enabled
   - Warning banner for security implications

## Acceptance Criteria

- [ ] Remote access toggle in settings works
- [ ] Connection info displayed clearly
- [ ] Password can be viewed, copied, regenerated
- [ ] QR code generates correctly
- [ ] Active sessions list updates in real-time
- [ ] Kick session functionality works
- [ ] UI properly reflects enable/disable state

## Files to Create/Modify

- `frontend/src/components/settings/RemoteAccessSettings.tsx` - Main settings component
- `frontend/src/components/settings/ConnectionInfo.tsx` - Connection display
- `frontend/src/components/settings/RemoteSessions.tsx` - Session management
- `frontend/src/services/remoteAccess.ts` - API service
- `frontend/src/store/remoteAccessStore.ts` - State management
- `frontend/src/pages/Settings.tsx` - Add remote access section

## Integration Points

- **Provides**: User interface for remote access feature
- **Consumes**: remote-access-config (indirectly via API)
- **Conflicts**: Minor changes to Settings.tsx layout
