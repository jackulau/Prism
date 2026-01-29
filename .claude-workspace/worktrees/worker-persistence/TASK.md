---
id: worker-persistence
name: Persistent Worker Task Queue
wave: 1
priority: 2
dependencies: []
estimated_hours: 4
tags:
- backend
- workers
- database
---

## Objective

Add database persistence to the worker/agent task queue system, ensuring tasks survive server restarts and can be monitored/managed externally.

## Context

The current implementation in `backend/internal/agent/`:
- `pool.go` - In-memory agent pool with task queue
- `task.go` - Task and BatchTask types with priority queue
- Tasks are lost on server restart
- No external visibility into queued/running tasks

This enhancement enables:
- Task recovery after server restart
- Task history and audit trail
- External task monitoring dashboard
- Webhook callbacks on task completion

## Implementation

1. Create `backend/internal/database/repository/task.go`:
   - `TaskRepository` with CRUD operations
   - `Create(task)`, `Update(task)`, `GetByID(id)`, `Delete(id)`
   - `ListByStatus(status, limit, offset)` - Get tasks by status
   - `ListByUserID(userID, limit, offset)` - Get user's tasks
   - `CleanupOld(before time.Time)` - Remove completed tasks older than N days

2. Create database migration `backend/internal/database/migrations/xxx_create_tasks.sql`:
   ```sql
   CREATE TABLE tasks (
     id TEXT PRIMARY KEY,
     user_id TEXT NOT NULL,
     prompt TEXT NOT NULL,
     context TEXT,
     priority INTEGER DEFAULT 1,
     status TEXT DEFAULT 'pending',
     agent_config TEXT,  -- JSON
     metadata TEXT,      -- JSON
     result TEXT,        -- JSON
     error TEXT,
     callback_url TEXT,
     callback_data TEXT, -- JSON
     created_at DATETIME,
     started_at DATETIME,
     completed_at DATETIME,
     FOREIGN KEY (user_id) REFERENCES users(id)
   );
   CREATE INDEX idx_tasks_status ON tasks(status);
   CREATE INDEX idx_tasks_user_id ON tasks(user_id);
   ```

3. Update `backend/internal/agent/pool.go`:
   - Add `TaskRepository` dependency
   - Persist task on `Submit()`
   - Update status on start/complete/fail
   - Load pending tasks on `Start()`
   - Add `RecoverTasks()` method for restart recovery

4. Create `backend/internal/api/handlers/tasks.go`:
   - `GET /api/v1/tasks` - List user's tasks with filters
   - `GET /api/v1/tasks/:id` - Get task details
   - `DELETE /api/v1/tasks/:id` - Cancel pending task
   - `POST /api/v1/tasks/:id/retry` - Retry failed task

5. Add WebSocket messages for task status updates:
   - `task.queued` - Task added to queue
   - `task.started` - Task execution began
   - `task.progress` - Task progress update
   - `task.completed` - Task finished successfully
   - `task.failed` - Task failed with error

## Acceptance Criteria

- [ ] Tasks persist to database on submission
- [ ] Task status updates on start/complete/fail
- [ ] Pending tasks recovered on server restart
- [ ] Tasks filterable by status and user
- [ ] Tasks can be cancelled while pending
- [ ] Failed tasks can be retried
- [ ] Completed tasks cleaned up after configurable retention period
- [ ] WebSocket notifications for task lifecycle events

## Files to Create/Modify

- `backend/internal/database/repository/task.go` - **Create**: Task persistence
- `backend/internal/database/migrations/xxx_create_tasks.sql` - **Create**: Migration
- `backend/internal/agent/pool.go` - **Modify**: Add persistence hooks
- `backend/internal/agent/task.go` - **Modify**: Add database fields
- `backend/internal/api/handlers/tasks.go` - **Create**: REST API
- `backend/internal/api/routes/router.go` - **Modify**: Register routes
- `backend/internal/api/websocket/hub.go` - **Modify**: Add message types

## Integration Points

- **Provides**: Persistent task queue, task management API
- **Consumes**: Database, WebSocket hub
- **Conflicts**: Careful with pool.go modifications - maintain backward compatibility

## Database Schema

```
tasks
├── id (TEXT, PK)
├── user_id (TEXT, FK → users)
├── prompt (TEXT)
├── context (TEXT, nullable)
├── priority (INTEGER)
├── status (TEXT: pending|running|completed|failed|cancelled)
├── agent_config (TEXT, JSON)
├── metadata (TEXT, JSON)
├── result (TEXT, JSON, nullable)
├── error (TEXT, nullable)
├── callback_url (TEXT, nullable)
├── callback_data (TEXT, JSON, nullable)
├── created_at (DATETIME)
├── started_at (DATETIME, nullable)
└── completed_at (DATETIME, nullable)
```
