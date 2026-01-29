---
id: sse-streaming
name: SSE Streaming Infrastructure
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- frontend
- streaming
---

## Objective

Implement Server-Sent Events (SSE) as an alternative/complement to WebSocket for streaming LLM responses, providing better compatibility with proxies, load balancers, and simpler client implementations.

## Context

The codebase currently uses WebSocket exclusively for streaming in:
- `backend/internal/api/websocket/` - Hub and client management
- `backend/internal/api/routes/chat_handler.go` - Message streaming
- `frontend/src/services/websocket.ts` - Client-side WebSocket service

While WebSocket works well, SSE offers advantages:
- Works through HTTP/2 proxies without special configuration
- Simpler to implement and debug
- Native browser EventSource API
- Better for one-way server-to-client streaming
- Automatic reconnection built-in

## Implementation

### 1. Create SSE Service

**File**: `backend/internal/api/sse/service.go`

```go
package sse

import (
    "github.com/gofiber/fiber/v2"
)

type Service struct {
    clients map[string]*Client
    mu      sync.RWMutex
}

type Client struct {
    ID       string
    UserID   string
    Writer   *bufio.Writer
    Done     chan struct{}
}

func NewService() *Service
func (s *Service) RegisterClient(userID string, c *fiber.Ctx) (*Client, error)
func (s *Service) UnregisterClient(clientID string)
func (s *Service) SendEvent(clientID string, event *Event) error
func (s *Service) BroadcastToUser(userID string, event *Event) error
```

### 2. Create SSE Event Types

**File**: `backend/internal/api/sse/events.go`

```go
package sse

type Event struct {
    ID    string      `json:"id,omitempty"`
    Type  EventType   `json:"event"`
    Data  interface{} `json:"data"`
    Retry int         `json:"retry,omitempty"` // Reconnection time in ms
}

type EventType string
const (
    EventChatChunk     EventType = "chat.chunk"
    EventChatComplete  EventType = "chat.complete"
    EventToolStarted   EventType = "tool.started"
    EventToolCompleted EventType = "tool.completed"
    EventToolConfirm   EventType = "tool.confirm"
    EventError         EventType = "error"
    EventHeartbeat     EventType = "heartbeat"
)

func (e *Event) Format() string {
    var buf bytes.Buffer
    if e.ID != "" {
        fmt.Fprintf(&buf, "id: %s\n", e.ID)
    }
    fmt.Fprintf(&buf, "event: %s\n", e.Type)
    data, _ := json.Marshal(e.Data)
    fmt.Fprintf(&buf, "data: %s\n\n", data)
    return buf.String()
}
```

### 3. Create SSE Handler

**File**: `backend/internal/api/handlers/sse.go`

```go
package handlers

func SSEHandler(sseService *sse.Service, deps *Dependencies) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Set SSE headers
        c.Set("Content-Type", "text/event-stream")
        c.Set("Cache-Control", "no-cache")
        c.Set("Connection", "keep-alive")
        c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

        // Get user from auth
        userID := c.Locals("userID").(string)

        // Register client
        client, err := sseService.RegisterClient(userID, c)
        if err != nil {
            return err
        }
        defer sseService.UnregisterClient(client.ID)

        // Keep connection alive with heartbeats
        // Stream events to client
        // ...
    }
}
```

### 4. Create SSE Chat Handler

**File**: `backend/internal/api/routes/sse_chat_handler.go`

```go
package routes

func HandleSSEChat(deps *Dependencies, sseService *sse.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse request
        var req ChatRequest
        if err := c.BodyParser(&req); err != nil {
            return err
        }

        // Start streaming response
        stream, err := deps.LLMManager.Chat(ctx, provider, llmReq)
        if err != nil {
            return err
        }

        // Stream chunks via SSE
        for chunk := range stream {
            sseService.SendEvent(clientID, &sse.Event{
                Type: sse.EventChatChunk,
                Data: map[string]interface{}{
                    "delta": chunk.Delta,
                    "conversationId": conversationID,
                    "messageId": messageID,
                },
            })
        }

        // Send completion
        sseService.SendEvent(clientID, &sse.Event{
            Type: sse.EventChatComplete,
            Data: map[string]interface{}{
                "messageId": messageID,
            },
        })
    }
}
```

### 5. Create Frontend SSE Service

**File**: `frontend/src/services/sse.ts`

```typescript
export class SSEService {
    private eventSource: EventSource | null = null;
    private handlers: Map<string, EventHandler[]> = new Map();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;

    connect(token: string): void {
        const url = `${API_BASE}/api/v1/sse?token=${token}`;
        this.eventSource = new EventSource(url);

        this.eventSource.onopen = () => {
            this.reconnectAttempts = 0;
        };

        this.eventSource.onerror = (error) => {
            this.handleReconnect();
        };

        // Register event handlers
        this.eventSource.addEventListener('chat.chunk', (e) => {
            const data = JSON.parse(e.data);
            this.emit('chat.chunk', data);
        });

        // ... other event types
    }

    on(event: string, handler: EventHandler): void {
        if (!this.handlers.has(event)) {
            this.handlers.set(event, []);
        }
        this.handlers.get(event)!.push(handler);
    }

    disconnect(): void {
        this.eventSource?.close();
        this.eventSource = null;
    }
}
```

### 6. Add SSE Routes

**File**: `backend/internal/api/routes/router.go`

```go
// SSE routes
api.Get("/sse", middleware.AuthMiddleware(deps.JWTService), handlers.SSEHandler(sseService, deps))
api.Post("/sse/chat", middleware.AuthMiddleware(deps.JWTService), HandleSSEChat(deps, sseService))
```

### 7. Update Frontend to Support Both

**File**: `frontend/src/services/streaming.ts`

```typescript
export class StreamingService {
    private transport: 'websocket' | 'sse';
    private wsService?: WebSocketService;
    private sseService?: SSEService;

    constructor(preferSSE: boolean = false) {
        this.transport = preferSSE ? 'sse' : 'websocket';
    }

    connect(token: string): void {
        if (this.transport === 'sse') {
            this.sseService = new SSEService();
            this.sseService.connect(token);
        } else {
            this.wsService = new WebSocketService();
            this.wsService.connect(token);
        }
    }

    on(event: string, handler: EventHandler): void {
        if (this.transport === 'sse') {
            this.sseService?.on(event, handler);
        } else {
            this.wsService?.on(event, handler);
        }
    }
}
```

## Acceptance Criteria

- [ ] SSE service with client registration and event streaming
- [ ] SSE event types matching WebSocket message types
- [ ] SSE endpoint with proper headers (Content-Type, Cache-Control, etc.)
- [ ] SSE chat handler for LLM streaming
- [ ] Heartbeat mechanism to keep connections alive
- [ ] Frontend SSEService with EventSource
- [ ] Automatic reconnection with exponential backoff
- [ ] StreamingService abstraction supporting both transports
- [ ] Tool confirmation events via SSE

## Files to Create/Modify

- `backend/internal/api/sse/service.go` - SSE service
- `backend/internal/api/sse/events.go` - Event types
- `backend/internal/api/handlers/sse.go` - SSE connection handler
- `backend/internal/api/routes/sse_chat_handler.go` - Chat streaming
- `frontend/src/services/sse.ts` - SSE client
- `frontend/src/services/streaming.ts` - Transport abstraction
- `backend/internal/api/routes/router.go` - Add SSE routes

## Integration Points

- **Provides**: SSE streaming as alternative to WebSocket
- **Provides**: Transport abstraction for frontend
- **Consumes**: LLM manager for chat streaming
- **Consumes**: Auth middleware for authentication
- **Conflicts**: None - additive alongside WebSocket
