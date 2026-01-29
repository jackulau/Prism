---
id: message-streaming
name: Enhanced Message Entity & Streaming
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- database
- messages
- streaming
---

## Objective

Enhance the Message entity with better metadata tracking and improve streaming architecture for real-time updates.

## Context

**Current Message Entity** (`backend/internal/database/repository/conversation.go`):
```go
type Message struct {
    ID             string
    ConversationID string
    Role           string      // "user", "assistant", "system"
    Content        string
    ToolCalls      []ToolCall  // JSON marshaled
    ToolCallID     string
    TokensUsed     int
    CreatedAt      time.Time
}
```

**Issues to Address:**
1. No tracking of streaming state (partial vs complete)
2. No parent message linking for conversation threading
3. Limited metadata for model/provider info per message
4. No tracking of extended thinking content
5. Tool results stored separately from tool calls

## Implementation

### Database Schema Updates

1. **Enhance Messages Table** (`backend/internal/database/sqlite.go`)
   ```sql
   ALTER TABLE messages ADD COLUMN provider TEXT;
   ALTER TABLE messages ADD COLUMN model TEXT;
   ALTER TABLE messages ADD COLUMN parent_id TEXT REFERENCES messages(id);
   ALTER TABLE messages ADD COLUMN status TEXT DEFAULT 'complete';
   ALTER TABLE messages ADD COLUMN thinking_content TEXT;
   ALTER TABLE messages ADD COLUMN metadata_json TEXT;
   ALTER TABLE messages ADD COLUMN input_tokens INTEGER DEFAULT 0;
   ALTER TABLE messages ADD COLUMN output_tokens INTEGER DEFAULT 0;
   ALTER TABLE messages ADD COLUMN finish_reason TEXT;

   CREATE INDEX idx_messages_parent_id ON messages(parent_id);
   CREATE INDEX idx_messages_status ON messages(status);
   ```

2. **Add Message Chunks Table** (for streaming reconstruction)
   ```sql
   CREATE TABLE message_chunks (
     id TEXT PRIMARY KEY,
     message_id TEXT NOT NULL,
     chunk_index INTEGER NOT NULL,
     content TEXT NOT NULL,
     chunk_type TEXT DEFAULT 'content',
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
   );
   CREATE INDEX idx_message_chunks_message_id ON message_chunks(message_id);
   ```

### Repository Updates

3. **Update Message Struct** (`backend/internal/database/repository/conversation.go`)
   ```go
   type Message struct {
       ID              string
       ConversationID  string
       ParentID        *string    // For threading
       Role            string
       Content         string
       ThinkingContent string     // Extended thinking
       ToolCalls       []ToolCall
       ToolCallID      string
       Provider        string
       Model           string
       Status          string     // streaming, complete, error
       InputTokens     int
       OutputTokens    int
       FinishReason    string
       MetadataJSON    string
       CreatedAt       time.Time
   }
   ```

4. **Add New Repository Methods**
   - `UpdateStatus(id, status string) error`
   - `UpdateContent(id, content string) error`
   - `AppendContent(id, delta string) error`
   - `SetThinkingContent(id, thinking string) error`
   - `GetThread(parentID string) ([]*Message, error)`
   - `SaveChunk(messageID string, index int, content, chunkType string) error`
   - `GetChunks(messageID string) ([]*MessageChunk, error)`

### Streaming Improvements

5. **Create Message Builder** (`backend/internal/api/routes/message_builder.go`)
   - Accumulates streaming chunks
   - Handles content + thinking content separately
   - Persists incrementally or on completion
   - Tracks streaming state

6. **Update Chat Handler** (`backend/internal/api/routes/chat_handler.go`)
   - Use MessageBuilder for streaming responses
   - Save message status transitions
   - Include provider/model in message record
   - Track thinking content from extended thinking

### Frontend Updates

7. **Update Message Types** (`frontend/src/types/index.ts`)
   ```typescript
   interface Message {
     id: string;
     conversation_id: string;
     parent_id?: string;
     role: 'user' | 'assistant' | 'system' | 'tool';
     content: string;
     thinking_content?: string;
     tool_calls?: ToolCall[];
     provider?: string;
     model?: string;
     status: 'streaming' | 'complete' | 'error';
     input_tokens?: number;
     output_tokens?: number;
     finish_reason?: string;
     created_at: string;
   }
   ```

8. **Update Message Display** (`frontend/src/components/chat/`)
   - Show streaming indicator based on status
   - Display thinking content in collapsible section
   - Show token counts in message metadata
   - Support threaded message display

## Acceptance Criteria

- [ ] Messages track provider and model used
- [ ] Streaming state is persisted and recoverable
- [ ] Extended thinking content is stored separately
- [ ] Token usage tracked per message (input/output)
- [ ] Parent-child message threading works
- [ ] Existing message functionality unchanged
- [ ] Frontend displays enhanced message metadata

## Files to Create/Modify

**Create:**
- `backend/internal/api/routes/message_builder.go` - Streaming message builder

**Modify:**
- `backend/internal/database/sqlite.go` - Schema updates
- `backend/internal/database/repository/conversation.go` - Enhanced Message struct
- `backend/internal/api/routes/chat_handler.go` - Use MessageBuilder
- `frontend/src/types/index.ts` - Updated Message interface
- `frontend/src/components/chat/MessageItem.tsx` - Display enhancements

## Integration Points

- **Provides**: Enhanced message metadata, streaming recovery, conversation threading
- **Consumes**: Existing conversation repository, WebSocket streaming
- **Conflicts**: Minimal - schema changes are additive (ALTER TABLE ADD COLUMN)
