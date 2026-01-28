import { z } from 'zod';

// ==================== Agent Configuration ====================

export const agentConfigSchema = z.object({
  name: z.string().optional(),
  provider: z.string(), // openai, anthropic, google, ollama
  model: z.string(),
  systemPrompt: z.string().optional(),
  temperature: z.number().min(0).max(2).default(0.7),
  maxTokens: z.number().positive().default(4096),
});

// ==================== Task Definition ====================

export const taskPrioritySchema = z.enum(['low', 'normal', 'high', 'urgent']);

export const taskSchema = z.object({
  id: z.string(),
  prompt: z.string(),
  context: z.string().optional(),
  priority: taskPrioritySchema.default('normal'),
  metadata: z.record(z.unknown()).optional(),
  timeout: z.number().positive().optional(), // milliseconds
});

// ==================== Execution Status ====================

export const executionStatusSchema = z.enum([
  'pending',
  'running',
  'completed',
  'partially_completed',
  'failed',
  'cancelled',
]);

export const executionTypeSchema = z.enum([
  'single',
  'parallel',
  'sequential',
]);

// ==================== Agent Result ====================

export const toolResultSchema = z.object({
  toolCallId: z.string(),
  name: z.string(),
  output: z.string(),
  error: z.string().optional(),
});

export const usageSchema = z.object({
  promptTokens: z.number(),
  completionTokens: z.number(),
  totalTokens: z.number(),
});

export const agentResultSchema = z.object({
  agentId: z.string(),
  taskId: z.string(),
  success: z.boolean(),
  output: z.string().optional(),
  toolResults: z.array(toolResultSchema).optional(),
  error: z.string().optional(),
  usage: usageSchema.optional(),
  durationMs: z.number(),
  completedAt: z.date(),
});

// ==================== Execution ====================

export const executionSchema = z.object({
  id: z.string(),
  type: executionTypeSchema,
  status: executionStatusSchema,
  tasks: z.array(taskSchema),
  results: z.array(agentResultSchema).optional(),
  error: z.string().optional(),
  startedAt: z.date(),
  completedAt: z.date().optional(),
});

// ==================== Swarm Types ====================

export const swarmStrategySchema = z.enum([
  'parallel',
  'pipeline',
  'debate',
  'consensus',
  'map_reduce',
  'specialist',
]);

export const agentRoleSchema = z.enum([
  'general',
  'planner',
  'coder',
  'reviewer',
  'researcher',
  'writer',
  'analyst',
  'debugger',
  'tester',
  'synthesizer',
]);

export const agentStatusSchema = z.enum([
  'idle',
  'running',
  'completed',
  'failed',
  'cancelled',
]);

export const swarmStatusSchema = z.enum([
  'pending',
  'running',
  'completed',
  'failed',
  'cancelled',
]);

export const agentRoleConfigSchema = z.object({
  role: agentRoleSchema,
  count: z.number().positive().default(1),
  config: agentConfigSchema.optional(),
  systemPrompt: z.string().optional(),
});

export const swarmAgentSchema = z.object({
  id: z.string(),
  role: agentRoleSchema,
  status: agentStatusSchema,
  input: z.string().optional(),
  output: z.string().optional(),
  startedAt: z.date().optional(),
  completedAt: z.date().optional(),
});

export const swarmMessageSchema = z.object({
  id: z.string(),
  fromAgent: z.string(),
  toAgent: z.string().optional(),
  type: z.string(), // "output", "request", "feedback", "critique"
  content: z.string(),
  timestamp: z.date(),
});

export const swarmResultSchema = z.object({
  agentId: z.string(),
  role: agentRoleSchema,
  output: z.string(),
  success: z.boolean(),
  error: z.string().optional(),
  durationMs: z.number(),
});

export const swarmSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  strategy: swarmStrategySchema,
  status: swarmStatusSchema,
  agents: z.array(swarmAgentSchema),
  messages: z.array(swarmMessageSchema).optional(),
  results: z.array(swarmResultSchema).optional(),
  finalOutput: z.string().optional(),
  error: z.string().optional(),
  createdAt: z.date(),
  startedAt: z.date().optional(),
  completedAt: z.date().optional(),
});

// ==================== Input Schemas ====================

export const runTaskInput = z.object({
  task: taskSchema,
  config: agentConfigSchema,
});

export const runParallelInput = z.object({
  tasks: z.array(taskSchema).min(1),
  config: agentConfigSchema,
});

export const runSequentialInput = z.object({
  tasks: z.array(taskSchema).min(1),
  config: agentConfigSchema,
});

export const runSwarmInput = z.object({
  prompt: z.string(),
  name: z.string().optional(),
  strategy: swarmStrategySchema,
  roles: z.array(agentRoleConfigSchema).min(1),
  config: agentConfigSchema,
  timeout: z.number().positive().optional(), // milliseconds
});

export const executionIdInput = z.object({
  executionId: z.string(),
});

export const swarmIdInput = z.object({
  swarmId: z.string(),
});

// ==================== Stats ====================

export const managerStatsSchema = z.object({
  totalExecutions: z.number(),
  runningExecutions: z.number(),
  completedExecutions: z.number(),
  failedExecutions: z.number(),
  cancelledExecutions: z.number(),
  activeAgents: z.number(),
  queuedTasks: z.number(),
  registeredConfigs: z.number(),
  activeSwarms: z.number(),
});

// ==================== Type Exports ====================

export type AgentConfig = z.infer<typeof agentConfigSchema>;
export type TaskPriority = z.infer<typeof taskPrioritySchema>;
export type Task = z.infer<typeof taskSchema>;
export type ExecutionStatus = z.infer<typeof executionStatusSchema>;
export type ExecutionType = z.infer<typeof executionTypeSchema>;
export type ToolResult = z.infer<typeof toolResultSchema>;
export type Usage = z.infer<typeof usageSchema>;
export type AgentResult = z.infer<typeof agentResultSchema>;
export type Execution = z.infer<typeof executionSchema>;
export type SwarmStrategy = z.infer<typeof swarmStrategySchema>;
export type AgentRole = z.infer<typeof agentRoleSchema>;
export type AgentStatus = z.infer<typeof agentStatusSchema>;
export type SwarmStatus = z.infer<typeof swarmStatusSchema>;
export type AgentRoleConfig = z.infer<typeof agentRoleConfigSchema>;
export type SwarmAgent = z.infer<typeof swarmAgentSchema>;
export type SwarmMessage = z.infer<typeof swarmMessageSchema>;
export type SwarmResult = z.infer<typeof swarmResultSchema>;
export type Swarm = z.infer<typeof swarmSchema>;
export type ManagerStats = z.infer<typeof managerStatsSchema>;
