---
id: agent-persistence
name: Agent Entity Database Persistence
wave: 1
priority: 1
dependencies: []
estimated_hours: 5
tags:
- backend
- database
- agent
---

## Objective

Persist Agent entities to the database for audit trails, resumability, and history tracking.

## Context

Currently, Agent instances exist only in memory during execution:
- **Location**: `backend/internal/agent/agent.go`
- **Current State**: Agents run via goroutines, results published to channels
- **Problem**: No persistence means no audit trail, no resumability, no history

**Related Entities Already Persisted:**
- `conversations` - Chat sessions
- `messages` - Chat messages with tool_calls JSON
- `tool_executions` - Individual tool execution records

## Implementation

### Database Schema

1. **Add Agents Table** (`backend/internal/database/sqlite.go`)
   ```sql
   CREATE TABLE agents (
     id TEXT PRIMARY KEY,
     user_id TEXT NOT NULL,
     conversation_id TEXT,
     name TEXT NOT NULL,
     description TEXT,
     provider TEXT NOT NULL,
     model TEXT NOT NULL,
     system_prompt TEXT,
     status TEXT NOT NULL DEFAULT 'pending',
     config_json TEXT,
     error TEXT,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     started_at DATETIME,
     completed_at DATETIME,
     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
     FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE SET NULL
   );
   CREATE INDEX idx_agents_user_id ON agents(user_id);
   CREATE INDEX idx_agents_conversation_id ON agents(conversation_id);
   CREATE INDEX idx_agents_status ON agents(status);
   ```

2. **Add Agent Results Table**
   ```sql
   CREATE TABLE agent_results (
     id TEXT PRIMARY KEY,
     agent_id TEXT NOT NULL,
     task_id TEXT,
     success INTEGER NOT NULL DEFAULT 0,
     output TEXT,
     error TEXT,
     usage_json TEXT,
     metadata_json TEXT,
     duration_ms INTEGER,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
   );
   CREATE INDEX idx_agent_results_agent_id ON agent_results(agent_id);
   ```

### Repository Layer

3. **Create Agent Repository** (`backend/internal/database/repository/agent.go`)
   - `Create(agent *Agent) error`
   - `GetByID(id string) (*Agent, error)`
   - `GetByUserID(userID string, limit, offset int) ([]*Agent, error)`
   - `GetByConversationID(conversationID string) ([]*Agent, error)`
   - `UpdateStatus(id, status string) error`
   - `UpdateError(id, error string) error`
   - `Complete(id string, completedAt time.Time) error`
   - `SaveResult(result *AgentResult) error`
   - `GetResults(agentID string) ([]*AgentResult, error)`

4. **Define Repository Structs** (`backend/internal/database/repository/agent.go`)
   ```go
   type AgentRecord struct {
     ID             string
     UserID         string
     ConversationID *string
     Name           string
     Description    string
     Provider       string
     Model          string
     SystemPrompt   string
     Status         string  // pending, running, completed, failed, cancelled
     ConfigJSON     string
     Error          string
     CreatedAt      time.Time
     StartedAt      *time.Time
     CompletedAt    *time.Time
   }

   type AgentResultRecord struct {
     ID           string
     AgentID      string
     TaskID       string
     Success      bool
     Output       string
     Error        string
     UsageJSON    string
     MetadataJSON string
     DurationMS   int64
     CreatedAt    time.Time
   }
   ```

### Agent Integration

5. **Update Agent Manager** (`backend/internal/agent/manager.go`)
   - Inject AgentRepository dependency
   - Persist agent on creation
   - Update status on state transitions
   - Save results on completion

6. **Update Agent Struct** (`backend/internal/agent/agent.go`)
   - Add `UserID` field for ownership tracking
   - Add persistence hooks for status changes

### API Endpoints

7. **Add Agent History Endpoints** (`backend/internal/api/handlers/agent.go`)
   - `GET /api/v1/agents` - List user's agents with pagination
   - `GET /api/v1/agents/:id` - Get agent details
   - `GET /api/v1/agents/:id/results` - Get agent execution results
   - `DELETE /api/v1/agents/:id` - Delete agent record

8. **Register Routes** (`backend/internal/api/routes/router.go`)
   - Add agent routes with auth middleware

## Acceptance Criteria

- [ ] Agents are persisted to database on creation
- [ ] Agent status updates are saved
- [ ] Agent results are stored with full details
- [ ] Users can query their agent history
- [ ] Agents can be linked to conversations
- [ ] Existing agent execution flow unchanged
- [ ] Database migrations run cleanly

## Files to Create/Modify

**Create:**
- `backend/internal/database/repository/agent.go` - Agent repository

**Modify:**
- `backend/internal/database/sqlite.go` - Add agents/agent_results tables
- `backend/internal/agent/agent.go` - Add UserID, persistence hooks
- `backend/internal/agent/manager.go` - Inject repository, persist on lifecycle
- `backend/internal/api/handlers/agent.go` - Add history endpoints (or create new)
- `backend/internal/api/routes/router.go` - Register agent routes

## Integration Points

- **Provides**: Agent history API, audit trail for agent executions
- **Consumes**: User repository (ownership), conversation repository (linking)
- **Conflicts**: Minimal - agent.go modifications should be additive
