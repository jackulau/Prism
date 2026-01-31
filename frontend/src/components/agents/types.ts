import type { Message, ToolCall } from '../../types';

export type AgentStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface AgentExecutionMetrics {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  inputCost?: number;
  outputCost?: number;
  totalCost?: number;
  duration?: number;
  iterationCount?: number;
}

export interface AgentExecution {
  id: string;
  name?: string;
  task?: string;
  status: AgentStatus;
  model?: string;
  provider?: string;
  systemPrompt?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  cost?: number;
  error?: string;
  result?: unknown;
  messages?: Message[];
  toolCalls?: ToolCall[];
  metrics?: AgentExecutionMetrics;
  created_at: string | Date;
  started_at?: string | Date | null;
  completed_at?: string | Date | null;
  updated_at?: string | Date;
}

export interface AgentFiltersState {
  status: AgentStatus | 'all';
  model: string | 'all';
  dateRange: 'all' | 'today' | 'week' | 'month';
  search: string;
}

export interface UseAgentsOptions {
  filters?: Partial<AgentFiltersState>;
  limit?: number;
  offset?: number;
}

export interface UseAgentsResult {
  agents: AgentExecution[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  refetch: () => void;
  hasMore: boolean;
}
