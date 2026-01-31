/**
 * Zod validation schemas for swarm API communication.
 * These schemas provide runtime validation for swarm API requests and responses.
 */

import { z } from 'zod';

// ============================================================================
// Enum Schemas
// ============================================================================

/** Schema for SwarmStrategy enum */
export const swarmStrategySchema = z.enum([
  'parallel',
  'pipeline',
  'debate',
  'consensus',
  'map_reduce',
  'specialist',
]);

/** Schema for AgentRole enum */
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

/** Schema for SwarmStatus enum */
export const swarmStatusSchema = z.enum([
  'pending',
  'running',
  'completed',
  'failed',
  'cancelled',
]);

/** Schema for AgentStatus enum */
export const agentStatusSchema = z.enum([
  'pending',
  'idle',
  'running',
  'completed',
  'failed',
  'cancelled',
]);

/** Schema for SwarmEventType enum */
export const swarmEventTypeSchema = z.enum([
  'swarm_started',
  'agent_started',
  'agent_output',
  'agent_completed',
  'agent_failed',
  'message',
  'synthesizing',
  'swarm_completed',
  'swarm_failed',
  'swarm_cancelled',
  'progress',
]);

/** Schema for SwarmMessageType enum */
export const swarmMessageTypeSchema = z.enum([
  'output',
  'request',
  'feedback',
  'critique',
]);

// ============================================================================
// Configuration Schemas
// ============================================================================

/** Schema for tool definition */
export const toolDefinitionSchema = z.object({
  name: z.string(),
  description: z.string(),
  parameters: z.record(z.unknown()).optional(),
});

/** Schema for agent configuration */
export const agentConfigSchema = z.object({
  id: z.string().optional(),
  user_id: z.string().optional(),
  conversation_id: z.string().optional(),
  name: z.string().optional(),
  description: z.string().optional(),
  provider: z.string(),
  model: z.string(),
  system_prompt: z.string().optional(),
  temperature: z.number().min(0).max(2).optional(),
  max_tokens: z.number().positive().optional(),
  tools: z.array(toolDefinitionSchema).optional(),
  metadata: z.record(z.string()).optional(),
});

/** Schema for agent role configuration */
export const agentRoleConfigSchema = z.object({
  role: agentRoleSchema,
  config: agentConfigSchema,
  count: z.number().positive().optional(),
  system_prompt: z.string().optional(),
});

/** Schema for swarm configuration */
export const swarmConfigSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Swarm name is required'),
  strategy: swarmStrategySchema,
  max_agents: z.number().positive().optional(),
  timeout: z.number().positive().optional(),
  agent_configs: z.array(agentRoleConfigSchema).min(1, 'At least one agent config is required'),
  synthesizer_config: agentConfigSchema.optional(),
});

// ============================================================================
// Runtime Schemas
// ============================================================================

/** Schema for swarm agent */
export const swarmAgentSchema = z.object({
  id: z.string(),
  role: agentRoleSchema,
  status: agentStatusSchema,
  input: z.string().optional(),
  output: z.string().optional(),
  started_at: z.string().optional(),
  completed_at: z.string().optional(),
});

/** Schema for swarm message */
export const swarmMessageSchema = z.object({
  id: z.string(),
  from_agent: z.string(),
  to_agent: z.string().optional(),
  type: swarmMessageTypeSchema,
  content: z.string(),
  timestamp: z.string(),
});

/** Schema for swarm result */
export const swarmResultSchema = z.object({
  agent_id: z.string(),
  role: agentRoleSchema,
  output: z.string(),
  success: z.boolean(),
  error: z.string().optional(),
  duration: z.number(),
  metadata: z.record(z.unknown()).optional(),
});

/** Schema for swarm event */
export const swarmEventSchema = z.object({
  swarm_id: z.string(),
  type: swarmEventTypeSchema,
  agent_id: z.string().optional(),
  role: agentRoleSchema.optional(),
  data: z.record(z.unknown()).optional(),
  timestamp: z.string(),
});

/** Schema for full swarm state */
export const swarmSchema = z.object({
  id: z.string(),
  config: swarmConfigSchema,
  status: swarmStatusSchema,
  agents: z.array(swarmAgentSchema),
  messages: z.array(swarmMessageSchema),
  results: z.array(swarmResultSchema),
  final_output: z.string().optional(),
  created_at: z.string(),
  started_at: z.string().optional(),
  completed_at: z.string().optional(),
  error: z.string().optional(),
});

// ============================================================================
// Real-time Update Schemas
// ============================================================================

/** Schema for swarm progress info */
export const swarmProgressInfoSchema = z.object({
  swarm_id: z.string(),
  status: swarmStatusSchema,
  total_agents: z.number(),
  completed_agents: z.number(),
  failed_agents: z.number(),
  current_phase: z.string().optional(),
  progress_percentage: z.number().min(0).max(100),
  estimated_time_remaining: z.number().optional(),
});

/** Schema for swarm agent info */
export const swarmAgentInfoSchema = z.object({
  id: z.string(),
  role: agentRoleSchema,
  status: agentStatusSchema,
  current_task: z.string().optional(),
  output_preview: z.string().optional(),
  tokens_used: z.number().optional(),
  started_at: z.string().optional(),
  completed_at: z.string().optional(),
});

// ============================================================================
// API Request/Response Schemas
// ============================================================================

/** Schema for create swarm request */
export const createSwarmRequestSchema = z.object({
  config: swarmConfigSchema,
});

/** Schema for create swarm response */
export const createSwarmResponseSchema = z.object({
  swarm: swarmSchema,
});

/** Schema for run swarm request */
export const runSwarmRequestSchema = z.object({
  swarm_id: z.string(),
  task: z.string().min(1, 'Task is required'),
});

/** Schema for run swarm response */
export const runSwarmResponseSchema = z.object({
  success: z.boolean(),
  error: z.string().optional(),
});

/** Schema for stop swarm request */
export const stopSwarmRequestSchema = z.object({
  swarm_id: z.string(),
});

/** Schema for stop swarm response */
export const stopSwarmResponseSchema = z.object({
  success: z.boolean(),
  error: z.string().optional(),
});

/** Schema for get swarm response */
export const getSwarmResponseSchema = z.object({
  swarm: swarmSchema,
});

/** Schema for list swarms response */
export const listSwarmsResponseSchema = z.object({
  swarms: z.array(swarmSchema),
});

// ============================================================================
// WebSocket Message Schemas
// ============================================================================

/** Schema for swarm WebSocket message types */
export const swarmWSMessageTypeSchema = z.enum([
  'swarm.event',
  'swarm.progress',
  'swarm.complete',
  'swarm.error',
]);

/** Schema for swarm WebSocket message */
export const swarmWSMessageSchema = z.object({
  type: swarmWSMessageTypeSchema,
  swarm_id: z.string(),
  event: swarmEventSchema.optional(),
  progress: swarmProgressInfoSchema.optional(),
  result: swarmSchema.optional(),
  error: z.string().optional(),
});

// ============================================================================
// Type Exports (inferred from schemas)
// ============================================================================

export type SwarmStrategy = z.infer<typeof swarmStrategySchema>;
export type AgentRole = z.infer<typeof agentRoleSchema>;
export type SwarmStatus = z.infer<typeof swarmStatusSchema>;
export type AgentStatus = z.infer<typeof agentStatusSchema>;
export type SwarmEventType = z.infer<typeof swarmEventTypeSchema>;
export type SwarmMessageType = z.infer<typeof swarmMessageTypeSchema>;
export type ToolDefinition = z.infer<typeof toolDefinitionSchema>;
export type AgentConfig = z.infer<typeof agentConfigSchema>;
export type AgentRoleConfig = z.infer<typeof agentRoleConfigSchema>;
export type SwarmConfig = z.infer<typeof swarmConfigSchema>;
export type SwarmAgent = z.infer<typeof swarmAgentSchema>;
export type SwarmMessage = z.infer<typeof swarmMessageSchema>;
export type SwarmResult = z.infer<typeof swarmResultSchema>;
export type SwarmEvent = z.infer<typeof swarmEventSchema>;
export type Swarm = z.infer<typeof swarmSchema>;
export type SwarmProgressInfo = z.infer<typeof swarmProgressInfoSchema>;
export type SwarmAgentInfo = z.infer<typeof swarmAgentInfoSchema>;
export type CreateSwarmRequest = z.infer<typeof createSwarmRequestSchema>;
export type CreateSwarmResponse = z.infer<typeof createSwarmResponseSchema>;
export type RunSwarmRequest = z.infer<typeof runSwarmRequestSchema>;
export type RunSwarmResponse = z.infer<typeof runSwarmResponseSchema>;
export type StopSwarmRequest = z.infer<typeof stopSwarmRequestSchema>;
export type StopSwarmResponse = z.infer<typeof stopSwarmResponseSchema>;
export type GetSwarmResponse = z.infer<typeof getSwarmResponseSchema>;
export type ListSwarmsResponse = z.infer<typeof listSwarmsResponseSchema>;
export type SwarmWSMessageType = z.infer<typeof swarmWSMessageTypeSchema>;
export type SwarmWSMessage = z.infer<typeof swarmWSMessageSchema>;
