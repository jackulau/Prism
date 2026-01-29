---
id: agent-frontend-types
name: Agent Frontend Types & API Service
wave: 1
priority: 2
dependencies: []
estimated_hours: 2
tags:
- frontend
- types
---

## Objective

Create TypeScript types for the Agent entity and API service functions in the frontend, following existing patterns.

## Context

The frontend uses TypeScript types in `/frontend/src/types/index.ts` and API services in `/frontend/src/services/`. This task adds Agent types and service functions.

## Implementation

### 1. Add Agent types in `/frontend/src/types/index.ts`

```typescript
// Agent status enum
export type AgentStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';

// Agent provider type enum
export type AgentProviderType = 'PRISM' | 'CURSOR' | 'JULES';

// Agent interface
export interface Agent {
  id: string;
  workspace_id: string | null;
  status: AgentStatus;
  provider_type: AgentProviderType;
  conversation_id: string | null;
  url: string | null;
  github_branch_name: string | null;
  name: string;
  model: string | null;
  sandbox_id: string | null;
  is_orchestrator_agent: boolean;
  created_at: string;
  updated_at: string;
}

// Create agent request
export interface CreateAgentRequest {
  workspace_id?: string;
  name: string;
  provider_type: AgentProviderType;
  model?: string;
  is_orchestrator_agent?: boolean;
}

// Update agent request
export interface UpdateAgentRequest {
  name?: string;
  status?: AgentStatus;
  url?: string;
  github_branch_name?: string;
  conversation_id?: string;
  sandbox_id?: string;
  model?: string;
}
```

### 2. Create Agent API service in `/frontend/src/services/agentService.ts`

```typescript
import { Agent, CreateAgentRequest, UpdateAgentRequest } from '../types';

const API_BASE = '/api/v1';

export const agentService = {
  async list(workspaceId?: string): Promise<Agent[]> {
    const params = workspaceId ? `?workspace_id=${workspaceId}` : '';
    const response = await fetch(`${API_BASE}/agents${params}`, {
      headers: { Authorization: `Bearer ${getToken()}` },
    });
    if (!response.ok) throw new Error('Failed to fetch agents');
    return response.json();
  },

  async get(id: string): Promise<Agent> {
    const response = await fetch(`${API_BASE}/agents/${id}`, {
      headers: { Authorization: `Bearer ${getToken()}` },
    });
    if (!response.ok) throw new Error('Failed to fetch agent');
    return response.json();
  },

  async create(data: CreateAgentRequest): Promise<Agent> {
    const response = await fetch(`${API_BASE}/agents`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getToken()}`,
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('Failed to create agent');
    return response.json();
  },

  async update(id: string, data: UpdateAgentRequest): Promise<Agent> {
    const response = await fetch(`${API_BASE}/agents/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getToken()}`,
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('Failed to update agent');
    return response.json();
  },

  async updateStatus(id: string, status: AgentStatus): Promise<Agent> {
    const response = await fetch(`${API_BASE}/agents/${id}/status`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getToken()}`,
      },
      body: JSON.stringify({ status }),
    });
    if (!response.ok) throw new Error('Failed to update agent status');
    return response.json();
  },

  async delete(id: string): Promise<void> {
    const response = await fetch(`${API_BASE}/agents/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${getToken()}` },
    });
    if (!response.ok) throw new Error('Failed to delete agent');
  },
};
```

### 3. Export from services index

Add to `/frontend/src/services/index.ts`:
```typescript
export * from './agentService';
```

## Acceptance Criteria

- [ ] Agent types added to types/index.ts
- [ ] AgentStatus and AgentProviderType enums defined
- [ ] CreateAgentRequest and UpdateAgentRequest interfaces
- [ ] agentService.ts created with all CRUD methods
- [ ] Proper TypeScript typing throughout
- [ ] Error handling for failed requests
- [ ] Authorization header included in all requests
- [ ] Service exported from services/index.ts

## Files to Create/Modify

- `frontend/src/types/index.ts` - Add Agent types
- `frontend/src/services/agentService.ts` - Create new service
- `frontend/src/services/index.ts` - Export new service (if exists)

## Integration Points

- **Provides**: TypeScript types and API service for Agent entity
- **Consumes**: Backend REST API (can be developed in parallel)
- **Conflicts**: Modifies types/index.ts - coordinate with other frontend tasks
