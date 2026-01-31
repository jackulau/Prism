/**
 * React Query hooks for agent and task data fetching and mutations.
 * Provides type-safe data fetching with proper loading/error states.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/trpc-client';
import type {
  Agent,
  AgentResult,
  Task,
  TaskStats,
} from '../types';
import type {
  AgentDTO,
  AgentResultDTO,
  TaskResponse,
} from '../schemas/api';

// ============================================================================
// Query Keys
// ============================================================================

export const agentQueryKeys = {
  all: ['agents'] as const,
  lists: () => [...agentQueryKeys.all, 'list'] as const,
  list: (params?: { limit?: number; offset?: number }) =>
    [...agentQueryKeys.lists(), params] as const,
  details: () => [...agentQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...agentQueryKeys.details(), id] as const,
  results: (id: string) => [...agentQueryKeys.all, 'results', id] as const,
  byConversation: (conversationId: string) =>
    [...agentQueryKeys.all, 'conversation', conversationId] as const,
};

export const taskQueryKeys = {
  all: ['tasks'] as const,
  lists: () => [...taskQueryKeys.all, 'list'] as const,
  list: (params?: { status?: string; limit?: number; offset?: number }) =>
    [...taskQueryKeys.lists(), params] as const,
  details: () => [...taskQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...taskQueryKeys.details(), id] as const,
  stats: () => [...taskQueryKeys.all, 'stats'] as const,
};

// ============================================================================
// Transform Functions
// ============================================================================

/** Transform backend AgentDTO to frontend Agent type */
function transformAgent(dto: AgentDTO): Agent {
  return {
    id: dto.id,
    conversationId: dto.conversation_id ?? undefined,
    name: dto.name,
    description: dto.description,
    provider: dto.provider,
    model: dto.model,
    systemPrompt: dto.system_prompt,
    status: dto.status,
    error: dto.error,
    createdAt: new Date(dto.created_at),
    startedAt: dto.started_at ? new Date(dto.started_at) : undefined,
    completedAt: dto.completed_at ? new Date(dto.completed_at) : undefined,
  };
}

/** Transform backend AgentResultDTO to frontend AgentResult type */
function transformAgentResult(dto: AgentResultDTO): AgentResult {
  return {
    id: dto.id,
    agentId: dto.agent_id,
    taskId: dto.task_id,
    success: dto.success,
    output: dto.output,
    error: dto.error,
    usage: dto.usage
      ? {
          inputTokens: (dto.usage as Record<string, number>).input_tokens ?? 0,
          outputTokens: (dto.usage as Record<string, number>).output_tokens ?? 0,
          totalTokens: (dto.usage as Record<string, number>).total_tokens ?? 0,
        }
      : undefined,
    metadata: dto.metadata as Record<string, unknown> | undefined,
    durationMs: dto.duration_ms,
    createdAt: new Date(dto.created_at),
  };
}

/** Transform backend TaskResponse to frontend Task type */
function transformTask(dto: TaskResponse): Task {
  return {
    id: dto.id,
    userId: dto.user_id,
    prompt: dto.prompt,
    context: dto.context,
    priority: dto.priority as 0 | 1 | 2 | 3,
    status: dto.status,
    agentConfig: dto.agent_config
      ? {
          provider: (dto.agent_config as Record<string, unknown>).provider as string | undefined,
          model: (dto.agent_config as Record<string, unknown>).model as string | undefined,
          temperature: (dto.agent_config as Record<string, unknown>).temperature as number | undefined,
          maxTokens: (dto.agent_config as Record<string, unknown>).max_tokens as number | undefined,
        }
      : undefined,
    metadata: dto.metadata as Record<string, unknown> | undefined,
    result: dto.result as Record<string, unknown> | undefined,
    error: dto.error,
    callbackUrl: dto.callback_url,
    createdAt: dto.created_at,
    startedAt: dto.started_at ?? undefined,
    completedAt: dto.completed_at ?? undefined,
  };
}

// ============================================================================
// Agent Query Hooks
// ============================================================================

export interface UseAgentsOptions {
  limit?: number;
  offset?: number;
  enabled?: boolean;
}

/** Query hook for listing agents with pagination */
export function useAgents(options: UseAgentsOptions = {}) {
  const { limit = 50, offset = 0, enabled = true } = options;

  return useQuery({
    queryKey: agentQueryKeys.list({ limit, offset }),
    queryFn: async () => {
      const result = await api.agents.list({ limit, offset });
      if (!result.success) {
        throw new Error(result.error);
      }
      return {
        agents: result.data.agents.map(transformAgent),
        total: result.data.total,
        limit: result.data.limit,
        offset: result.data.offset,
      };
    },
    enabled,
  });
}

/** Query hook for a single agent */
export function useAgent(id: string | undefined, options: { enabled?: boolean } = {}) {
  const { enabled = true } = options;

  return useQuery({
    queryKey: agentQueryKeys.detail(id ?? ''),
    queryFn: async () => {
      const result = await api.agents.get(id!);
      if (!result.success) {
        throw new Error(result.error);
      }
      return transformAgent(result.data);
    },
    enabled: enabled && !!id,
  });
}

/** Query hook for agent results */
export function useAgentResults(id: string | undefined, options: { enabled?: boolean } = {}) {
  const { enabled = true } = options;

  return useQuery({
    queryKey: agentQueryKeys.results(id ?? ''),
    queryFn: async () => {
      const result = await api.agents.getResults(id!);
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data.results.map(transformAgentResult);
    },
    enabled: enabled && !!id,
  });
}

/** Query hook for agents in a conversation */
export function useAgentsByConversation(
  conversationId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options;

  return useQuery({
    queryKey: agentQueryKeys.byConversation(conversationId ?? ''),
    queryFn: async () => {
      const result = await api.agents.getByConversation(conversationId!);
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data.agents.map(transformAgent);
    },
    enabled: enabled && !!conversationId,
  });
}

// ============================================================================
// Agent Mutation Hooks
// ============================================================================

/** Mutation hook for deleting an agent */
export function useDeleteAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const result = await api.agents.delete(id);
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentQueryKeys.all });
    },
  });
}

// ============================================================================
// Task Query Hooks
// ============================================================================

export interface UseTasksOptions {
  status?: string;
  limit?: number;
  offset?: number;
  enabled?: boolean;
  refetchInterval?: number | false;
}

/** Query hook for listing tasks with filtering and pagination */
export function useTasks(options: UseTasksOptions = {}) {
  const { status, limit = 20, offset = 0, enabled = true, refetchInterval = false } = options;

  return useQuery({
    queryKey: taskQueryKeys.list({ status, limit, offset }),
    queryFn: async () => {
      const result = await api.tasks.list({ status, limit, offset });
      if (!result.success) {
        throw new Error(result.error);
      }
      return {
        tasks: result.data.tasks.map(transformTask),
        total: result.data.total,
        limit: result.data.limit,
        offset: result.data.offset,
      };
    },
    enabled,
    refetchInterval,
  });
}

/** Query hook for a single task */
export function useTask(id: string | undefined, options: { enabled?: boolean; refetchInterval?: number | false } = {}) {
  const { enabled = true, refetchInterval = false } = options;

  return useQuery({
    queryKey: taskQueryKeys.detail(id ?? ''),
    queryFn: async () => {
      const result = await api.tasks.get(id!);
      if (!result.success) {
        throw new Error(result.error);
      }
      return transformTask(result.data);
    },
    enabled: enabled && !!id,
    refetchInterval,
  });
}

/** Query hook for task statistics */
export function useTaskStats(options: { enabled?: boolean; refetchInterval?: number | false } = {}) {
  const { enabled = true, refetchInterval = false } = options;

  return useQuery({
    queryKey: taskQueryKeys.stats(),
    queryFn: async () => {
      const result = await api.tasks.stats();
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data as TaskStats;
    },
    enabled,
    refetchInterval,
  });
}

// ============================================================================
// Task Mutation Hooks
// ============================================================================

/** Mutation hook for cancelling a task */
export function useCancelTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const result = await api.tasks.cancel(id);
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
    },
  });
}

/** Mutation hook for retrying a task */
export function useRetryTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const result = await api.tasks.retry(id);
      if (!result.success) {
        throw new Error(result.error);
      }
      return result.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
    },
  });
}
