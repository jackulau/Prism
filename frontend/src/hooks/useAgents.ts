import { useState, useEffect, useCallback, useRef } from 'react';
import { useAuthStore } from '../store/authStore';
import type { AgentExecution, AgentStatus, UseAgentsOptions, UseAgentsResult } from '../components/agents/types';

interface TaskResponse {
  id: string;
  user_id: string;
  prompt: string;
  context?: string;
  priority: number;
  status: string;
  agent_config?: {
    provider?: string;
    model?: string;
    name?: string;
  };
  metadata?: Record<string, unknown>;
  result?: Record<string, unknown>;
  error?: string;
  created_at: number;
  started_at?: number;
  completed_at?: number;
}

interface TasksListResponse {
  tasks: TaskResponse[];
  total: number;
  limit: number;
  offset: number;
}

interface TaskStatsResponse {
  total: number;
  pending: number;
  running: number;
  completed: number;
  failed: number;
}

function toAgentExecution(task: TaskResponse): AgentExecution {
  return {
    id: task.id,
    name: task.agent_config?.name || (task.metadata?.name as string) || undefined,
    task: task.prompt,
    status: task.status as AgentStatus,
    model: task.agent_config?.model,
    provider: task.agent_config?.provider,
    error: task.error,
    result: task.result,
    created_at: new Date(task.created_at),
    started_at: task.started_at ? new Date(task.started_at) : null,
    completed_at: task.completed_at ? new Date(task.completed_at) : null,
  };
}

function useApiClient() {
  const { accessToken } = useAuthStore();

  const fetchWithAuth = useCallback(async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    };

    if (accessToken) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${accessToken}`;
    }

    return fetch(url, {
      ...options,
      headers,
    });
  }, [accessToken]);

  return { fetchWithAuth };
}

export function useAgents(options: UseAgentsOptions = {}): UseAgentsResult {
  const { filters, limit = 20, offset = 0 } = options;
  const { fetchWithAuth } = useApiClient();

  const [agents, setAgents] = useState<AgentExecution[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const lastFetchRef = useRef<number>(0);

  const fetchAgents = useCallback(async () => {
    const now = Date.now();
    if (now - lastFetchRef.current < 500) return;
    lastFetchRef.current = now;

    setIsLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      params.set('limit', limit.toString());
      params.set('offset', offset.toString());

      if (filters?.status && filters.status !== 'all') {
        params.set('status', filters.status);
      }

      const response = await fetchWithAuth(`/api/v1/tasks?${params.toString()}`);

      if (!response.ok) {
        throw new Error(`Failed to fetch agents: ${response.statusText}`);
      }

      const data: TasksListResponse = await response.json();
      setAgents(data.tasks.map(toAgentExecution));
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch agents'));
    } finally {
      setIsLoading(false);
    }
  }, [fetchWithAuth, limit, offset, filters?.status]);

  useEffect(() => {
    fetchAgents();
  }, [fetchAgents]);

  const refetch = useCallback(() => {
    lastFetchRef.current = 0;
    fetchAgents();
  }, [fetchAgents]);

  return {
    agents,
    total,
    isLoading,
    error,
    refetch,
    hasMore: offset + agents.length < total,
  };
}

export function useAgentStats() {
  const { fetchWithAuth } = useApiClient();
  const [stats, setStats] = useState<TaskStatsResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchStats = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetchWithAuth('/api/v1/tasks/stats');

      if (!response.ok) {
        throw new Error(`Failed to fetch stats: ${response.statusText}`);
      }

      const data: TaskStatsResponse = await response.json();
      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch stats'));
    } finally {
      setIsLoading(false);
    }
  }, [fetchWithAuth]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  return { stats, isLoading, error, refetch: fetchStats };
}

export function useAgentActions() {
  const { fetchWithAuth } = useApiClient();

  const cancelAgent = useCallback(async (id: string) => {
    const response = await fetchWithAuth(`/api/v1/tasks/${id}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error('Failed to cancel agent');
    }

    return response.json();
  }, [fetchWithAuth]);

  const retryAgent = useCallback(async (id: string) => {
    const response = await fetchWithAuth(`/api/v1/tasks/${id}/retry`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error('Failed to retry agent');
    }

    return response.json();
  }, [fetchWithAuth]);

  const getAgent = useCallback(async (id: string): Promise<AgentExecution> => {
    const response = await fetchWithAuth(`/api/v1/tasks/${id}`);

    if (!response.ok) {
      throw new Error('Failed to get agent');
    }

    const data: TaskResponse = await response.json();
    return toAgentExecution(data);
  }, [fetchWithAuth]);

  return { cancelAgent, retryAgent, getAgent };
}
