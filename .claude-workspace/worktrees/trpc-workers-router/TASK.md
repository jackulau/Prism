---
id: trpc-workers-router
name: Workers Router Implementation
wave: 2
priority: 2
dependencies:
- trpc-core-setup
estimated_hours: 4
tags:
- backend
- api
- trpc
- workers
- agents
---

## Objective

Implement the tRPC workers router for agent/worker management, providing type-safe procedures for running, stopping, and monitoring AI agent executions.

## Context

The existing Go backend has an Agent Manager (`/backend/internal/agent/manager.go`) that handles:
- Single agent execution (`RunTask`)
- Parallel agent execution (`RunParallel`)
- Multi-agent swarms (`RunMultiAgent`)
- Execution cancellation and status

Currently these are controlled via WebSocket messages. The tRPC router provides a REST-like alternative with full type safety.

## Implementation

### 1. Define Zod Schemas

**File: `packages/trpc/src/routers/workers/schemas.ts`**
```typescript
import { z } from 'zod';

// Agent configuration
export const agentConfigSchema = z.object({
  name: z.string().optional(),
  provider: z.string(), // openai, anthropic, google, ollama
  model: z.string(),
  systemPrompt: z.string().optional(),
  temperature: z.number().min(0).max(2).default(0.7),
  maxTokens: z.number().positive().default(4096),
});

// Task definition
export const taskSchema = z.object({
  id: z.string(),
  prompt: z.string(),
  conversationId: z.string().optional(),
  workspaceDir: z.string().optional(),
});

// Execution status
export const executionStatusSchema = z.enum([
  'pending',
  'running',
  'completed',
  'failed',
  'cancelled',
]);

// Execution result
export const executionSchema = z.object({
  id: z.string(),
  status: executionStatusSchema,
  task: taskSchema,
  result: z.string().nullable(),
  error: z.string().nullable(),
  startedAt: z.date(),
  completedAt: z.date().nullable(),
  tokensUsed: z.number().nullable(),
});

// Swarm role configuration
export const agentRoleSchema = z.enum(['coder', 'reviewer', 'tester']);

export const swarmRoleConfigSchema = z.object({
  role: agentRoleSchema,
  count: z.number().positive().default(1),
  config: agentConfigSchema.optional(),
});

// Swarm strategy
export const swarmStrategySchema = z.enum([
  'parallel',
  'sequential',
  'hierarchical',
]);

// Swarm definition
export const swarmSchema = z.object({
  id: z.string(),
  status: executionStatusSchema,
  strategy: swarmStrategySchema,
  agents: z.array(z.object({
    id: z.string(),
    role: agentRoleSchema,
    status: executionStatusSchema,
  })),
  result: z.string().nullable(),
  startedAt: z.date(),
  completedAt: z.date().nullable(),
});

// Input schemas
export const runTaskInput = z.object({
  task: taskSchema,
  config: agentConfigSchema,
});

export const runParallelInput = z.object({
  tasks: z.array(taskSchema),
  config: agentConfigSchema,
});

export const runSwarmInput = z.object({
  prompt: z.string(),
  strategy: swarmStrategySchema,
  roles: z.array(swarmRoleConfigSchema),
  config: agentConfigSchema,
});

export const executionIdInput = z.object({
  executionId: z.string(),
});

export const swarmIdInput = z.object({
  swarmId: z.string(),
});

export type AgentConfig = z.infer<typeof agentConfigSchema>;
export type Task = z.infer<typeof taskSchema>;
export type Execution = z.infer<typeof executionSchema>;
export type Swarm = z.infer<typeof swarmSchema>;
```

### 2. Implement Workers Router

**File: `packages/trpc/src/routers/workers/index.ts`**
```typescript
import { router, protectedProcedure } from '../../trpc';
import { TRPCError } from '@trpc/server';
import * as schemas from './schemas';

export const workersRouter = router({
  // Run a single agent task
  runTask: protectedProcedure
    .input(schemas.runTaskInput)
    .output(schemas.executionSchema)
    .mutation(async ({ ctx, input }) => {
      const execution = await workerService.runTask(
        ctx.session.userId,
        input.task,
        input.config
      );
      return execution;
    }),

  // Run parallel agent tasks
  runParallel: protectedProcedure
    .input(schemas.runParallelInput)
    .output(schemas.executionSchema)
    .mutation(async ({ ctx, input }) => {
      const execution = await workerService.runParallel(
        ctx.session.userId,
        input.tasks,
        input.config
      );
      return execution;
    }),

  // Run multi-agent swarm
  runSwarm: protectedProcedure
    .input(schemas.runSwarmInput)
    .output(schemas.swarmSchema)
    .mutation(async ({ ctx, input }) => {
      const swarm = await workerService.runSwarm(
        ctx.session.userId,
        input.prompt,
        input.strategy,
        input.roles,
        input.config
      );
      return swarm;
    }),

  // Get execution status
  getExecution: protectedProcedure
    .input(schemas.executionIdInput)
    .output(schemas.executionSchema.nullable())
    .query(async ({ ctx, input }) => {
      const execution = await workerService.getExecution(input.executionId);
      return execution;
    }),

  // List active executions
  listExecutions: protectedProcedure
    .output(z.array(schemas.executionSchema))
    .query(async ({ ctx }) => {
      const executions = await workerService.listExecutions(ctx.session.userId);
      return executions;
    }),

  // Cancel execution
  cancelExecution: protectedProcedure
    .input(schemas.executionIdInput)
    .mutation(async ({ ctx, input }) => {
      await workerService.cancelExecution(input.executionId);
      return { success: true };
    }),

  // Get swarm status
  getSwarm: protectedProcedure
    .input(schemas.swarmIdInput)
    .output(schemas.swarmSchema.nullable())
    .query(async ({ ctx, input }) => {
      const swarm = await workerService.getSwarm(input.swarmId);
      return swarm;
    }),

  // List active swarms
  listSwarms: protectedProcedure
    .output(z.array(schemas.swarmSchema))
    .query(async ({ ctx }) => {
      const swarms = await workerService.listSwarms(ctx.session.userId);
      return swarms;
    }),

  // Cancel swarm
  cancelSwarm: protectedProcedure
    .input(schemas.swarmIdInput)
    .mutation(async ({ ctx, input }) => {
      await workerService.cancelSwarm(input.swarmId);
      return { success: true };
    }),
});
```

### 3. Create Worker Service

**File: `packages/trpc/src/services/worker.ts`**
```typescript
import type { Task, AgentConfig, Execution, Swarm, SwarmRoleConfig, SwarmStrategy } from '../routers/workers/schemas';

export const workerService = {
  async runTask(userId: string, task: Task, config: AgentConfig): Promise<Execution> {
    // Call Go backend agent manager or implement directly
    // POST to /api/v1/ws or use direct agent invocation
  },

  async runParallel(userId: string, tasks: Task[], config: AgentConfig): Promise<Execution> {
    // Parallel task execution
  },

  async runSwarm(
    userId: string,
    prompt: string,
    strategy: SwarmStrategy,
    roles: SwarmRoleConfig[],
    config: AgentConfig
  ): Promise<Swarm> {
    // Multi-agent swarm execution
  },

  async getExecution(executionId: string): Promise<Execution | null> {
    // Get execution by ID
  },

  async listExecutions(userId: string): Promise<Execution[]> {
    // List user's active executions
  },

  async cancelExecution(executionId: string): Promise<void> {
    // Cancel running execution
  },

  async getSwarm(swarmId: string): Promise<Swarm | null> {
    // Get swarm by ID
  },

  async listSwarms(userId: string): Promise<Swarm[]> {
    // List user's active swarms
  },

  async cancelSwarm(swarmId: string): Promise<void> {
    // Cancel running swarm
  },
};
```

## Acceptance Criteria

- [ ] All agent/worker Zod schemas defined
- [ ] Task execution procedures implemented
- [ ] Swarm management procedures implemented
- [ ] Execution status tracking
- [ ] Cancellation support
- [ ] User isolation (users can only manage their own workers)
- [ ] Proper error handling for failed executions

## Files to Create/Modify

- `packages/trpc/src/routers/workers/schemas.ts` - Zod schemas
- `packages/trpc/src/routers/workers/index.ts` - Router implementation
- `packages/trpc/src/services/worker.ts` - Worker service
- `packages/trpc/src/router.ts` - Add workers router (modify)

## Integration Points

- **Provides**: Worker/agent management via tRPC
- **Consumes**: trpc-core-setup (protectedProcedure, router, context)
- **Conflicts**: Avoid modifying Go agent manager directly

## Notes

- Existing Go agent types in `/backend/internal/agent/`
- WebSocket messages: TypeAgentRun, TypeAgentStop, TypeAgentStatus, etc.
- Consider using WebSocket for streaming results, tRPC for control
- Swarm strategies: parallel, sequential, hierarchical
