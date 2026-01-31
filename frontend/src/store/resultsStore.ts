import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Result aggregation types
export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface BatchResultSummary {
  id: string;
  name: string;
  status: ExecutionStatus;
  totalExecutions: number;
  completedExecutions: number;
  failedExecutions: number;
  totalTokens: number;
  totalCost: number;
  startedAt: Date;
  completedAt?: Date;
  model: string;
  provider: string;
}

export interface SwarmResultSummary {
  id: string;
  name: string;
  status: ExecutionStatus;
  agentCount: number;
  completedAgents: number;
  failedAgents: number;
  totalTokens: number;
  totalCost: number;
  startedAt: Date;
  completedAt?: Date;
  orchestratorModel: string;
  provider: string;
}

export interface ExecutionDetail {
  id: string;
  type: 'batch' | 'swarm';
  status: ExecutionStatus;
  input: string;
  output?: string;
  error?: string;
  tokens: {
    input: number;
    output: number;
    total: number;
  };
  cost: number;
  model: string;
  provider: string;
  startedAt: Date;
  completedAt?: Date;
  duration?: number;
  metadata?: Record<string, unknown>;
}

export interface AggregatedMetrics {
  totalExecutions: number;
  successfulExecutions: number;
  failedExecutions: number;
  successRate: number;
  totalTokens: {
    input: number;
    output: number;
    total: number;
  };
  averageTokensPerExecution: number;
  totalCost: number;
  averageCostPerExecution: number;
  averageDuration: number;
  byProvider: Record<string, {
    count: number;
    tokens: number;
    cost: number;
  }>;
  byModel: Record<string, {
    count: number;
    tokens: number;
    cost: number;
  }>;
}

export interface TimelineDataPoint {
  timestamp: Date;
  executionCount: number;
  successCount: number;
  failureCount: number;
  tokenUsage: number;
  cost: number;
}

export interface ResultFilters {
  status?: ExecutionStatus[];
  provider?: string[];
  model?: string[];
  type?: ('batch' | 'swarm')[];
  dateRange?: {
    start: Date;
    end: Date;
  };
  search?: string;
}

export interface TimeRange {
  start: Date;
  end: Date;
  granularity?: 'hour' | 'day' | 'week' | 'month';
}

// State interface
interface ResultsState {
  // Data
  batchResults: BatchResultSummary[];
  swarmResults: SwarmResultSummary[];
  selectedExecution: string | null;
  executionDetails: Map<string, ExecutionDetail>;
  aggregatedMetrics: AggregatedMetrics | null;
  timelineData: TimelineDataPoint[];

  // UI State
  isLoading: boolean;
  error: string | null;
  filters: ResultFilters;
  lastUpdated: Date | null;

  // Actions
  fetchBatchResults: (filters?: ResultFilters) => Promise<void>;
  fetchSwarmResults: (filters?: ResultFilters) => Promise<void>;
  fetchExecutionDetails: (executionId: string) => Promise<void>;
  fetchAggregatedMetrics: (timeRange?: TimeRange) => Promise<void>;
  fetchTimeline: (timeRange: TimeRange) => Promise<void>;
  setFilters: (filters: ResultFilters) => void;
  selectExecution: (id: string | null) => void;
  clearResults: () => void;
  setError: (error: string | null) => void;
}

// Initial state values
const initialFilters: ResultFilters = {};

const initialState = {
  batchResults: [],
  swarmResults: [],
  selectedExecution: null,
  executionDetails: new Map<string, ExecutionDetail>(),
  aggregatedMetrics: null,
  timelineData: [],
  isLoading: false,
  error: null,
  filters: initialFilters,
  lastUpdated: null,
};

// API helper for results endpoints
const API_BASE = '/api/v1';

const resultsApi = {
  async getBatchResults(filters?: ResultFilters): Promise<BatchResultSummary[]> {
    const params = new URLSearchParams();
    if (filters?.status?.length) params.set('status', filters.status.join(','));
    if (filters?.provider?.length) params.set('provider', filters.provider.join(','));
    if (filters?.model?.length) params.set('model', filters.model.join(','));
    if (filters?.dateRange) {
      params.set('start_date', filters.dateRange.start.toISOString());
      params.set('end_date', filters.dateRange.end.toISOString());
    }
    if (filters?.search) params.set('search', filters.search);

    const response = await fetch(`${API_BASE}/results/batch?${params}`);
    if (!response.ok) {
      throw new Error('Failed to fetch batch results');
    }
    const data = await response.json();
    return data.results.map((r: Record<string, unknown>) => ({
      ...r,
      startedAt: new Date(r.started_at as string),
      completedAt: r.completed_at ? new Date(r.completed_at as string) : undefined,
    }));
  },

  async getSwarmResults(filters?: ResultFilters): Promise<SwarmResultSummary[]> {
    const params = new URLSearchParams();
    if (filters?.status?.length) params.set('status', filters.status.join(','));
    if (filters?.provider?.length) params.set('provider', filters.provider.join(','));
    if (filters?.model?.length) params.set('model', filters.model.join(','));
    if (filters?.dateRange) {
      params.set('start_date', filters.dateRange.start.toISOString());
      params.set('end_date', filters.dateRange.end.toISOString());
    }
    if (filters?.search) params.set('search', filters.search);

    const response = await fetch(`${API_BASE}/results/swarm?${params}`);
    if (!response.ok) {
      throw new Error('Failed to fetch swarm results');
    }
    const data = await response.json();
    return data.results.map((r: Record<string, unknown>) => ({
      ...r,
      startedAt: new Date(r.started_at as string),
      completedAt: r.completed_at ? new Date(r.completed_at as string) : undefined,
    }));
  },

  async getExecutionDetails(executionId: string): Promise<ExecutionDetail> {
    const response = await fetch(`${API_BASE}/results/execution/${executionId}`);
    if (!response.ok) {
      throw new Error('Failed to fetch execution details');
    }
    const data = await response.json();
    return {
      ...data,
      startedAt: new Date(data.started_at),
      completedAt: data.completed_at ? new Date(data.completed_at) : undefined,
    };
  },

  async getAggregatedMetrics(timeRange?: TimeRange): Promise<AggregatedMetrics> {
    const params = new URLSearchParams();
    if (timeRange) {
      params.set('start_date', timeRange.start.toISOString());
      params.set('end_date', timeRange.end.toISOString());
    }

    const response = await fetch(`${API_BASE}/results/metrics?${params}`);
    if (!response.ok) {
      throw new Error('Failed to fetch aggregated metrics');
    }
    return response.json();
  },

  async getTimeline(timeRange: TimeRange): Promise<TimelineDataPoint[]> {
    const params = new URLSearchParams();
    params.set('start_date', timeRange.start.toISOString());
    params.set('end_date', timeRange.end.toISOString());
    if (timeRange.granularity) params.set('granularity', timeRange.granularity);

    const response = await fetch(`${API_BASE}/results/timeline?${params}`);
    if (!response.ok) {
      throw new Error('Failed to fetch timeline data');
    }
    const data = await response.json();
    return data.points.map((p: Record<string, unknown>) => ({
      ...p,
      timestamp: new Date(p.timestamp as string),
    }));
  },
};

// Create the store with persist middleware
export const useResultsStore = create<ResultsState>()(
  persist(
    (set, get) => ({
      ...initialState,

      fetchBatchResults: async (filters) => {
        set({ isLoading: true, error: null });
        try {
          const appliedFilters = filters || get().filters;
          const results = await resultsApi.getBatchResults(appliedFilters);
          set({
            batchResults: results,
            isLoading: false,
            lastUpdated: new Date(),
          });
        } catch (err) {
          set({
            error: err instanceof Error ? err.message : 'Failed to fetch batch results',
            isLoading: false,
          });
        }
      },

      fetchSwarmResults: async (filters) => {
        set({ isLoading: true, error: null });
        try {
          const appliedFilters = filters || get().filters;
          const results = await resultsApi.getSwarmResults(appliedFilters);
          set({
            swarmResults: results,
            isLoading: false,
            lastUpdated: new Date(),
          });
        } catch (err) {
          set({
            error: err instanceof Error ? err.message : 'Failed to fetch swarm results',
            isLoading: false,
          });
        }
      },

      fetchExecutionDetails: async (executionId) => {
        set({ isLoading: true, error: null });
        try {
          const details = await resultsApi.getExecutionDetails(executionId);
          set((state) => {
            const newDetails = new Map(state.executionDetails);
            newDetails.set(executionId, details);
            return {
              executionDetails: newDetails,
              isLoading: false,
              lastUpdated: new Date(),
            };
          });
        } catch (err) {
          set({
            error: err instanceof Error ? err.message : 'Failed to fetch execution details',
            isLoading: false,
          });
        }
      },

      fetchAggregatedMetrics: async (timeRange) => {
        set({ isLoading: true, error: null });
        try {
          const metrics = await resultsApi.getAggregatedMetrics(timeRange);
          set({
            aggregatedMetrics: metrics,
            isLoading: false,
            lastUpdated: new Date(),
          });
        } catch (err) {
          set({
            error: err instanceof Error ? err.message : 'Failed to fetch metrics',
            isLoading: false,
          });
        }
      },

      fetchTimeline: async (timeRange) => {
        set({ isLoading: true, error: null });
        try {
          const timeline = await resultsApi.getTimeline(timeRange);
          set({
            timelineData: timeline,
            isLoading: false,
            lastUpdated: new Date(),
          });
        } catch (err) {
          set({
            error: err instanceof Error ? err.message : 'Failed to fetch timeline',
            isLoading: false,
          });
        }
      },

      setFilters: (filters) => {
        set({ filters });
      },

      selectExecution: (id) => {
        set({ selectedExecution: id });
      },

      clearResults: () => {
        set({
          ...initialState,
          executionDetails: new Map<string, ExecutionDetail>(),
        });
      },

      setError: (error) => {
        set({ error });
      },
    }),
    {
      name: 'prism-results',
      partialize: (state) => ({
        // Only persist filters and recent data for caching
        filters: state.filters,
        batchResults: state.batchResults.slice(0, 50), // Cache last 50 batch results
        swarmResults: state.swarmResults.slice(0, 50), // Cache last 50 swarm results
        aggregatedMetrics: state.aggregatedMetrics,
        lastUpdated: state.lastUpdated,
      }),
      // Custom storage handlers for Date serialization
      storage: {
        getItem: (name) => {
          const str = localStorage.getItem(name);
          if (!str) return null;
          const parsed = JSON.parse(str);
          // Restore Date objects
          if (parsed.state) {
            if (parsed.state.lastUpdated) {
              parsed.state.lastUpdated = new Date(parsed.state.lastUpdated);
            }
            if (parsed.state.batchResults) {
              parsed.state.batchResults = parsed.state.batchResults.map((r: Record<string, unknown>) => ({
                ...r,
                startedAt: new Date(r.startedAt as string),
                completedAt: r.completedAt ? new Date(r.completedAt as string) : undefined,
              }));
            }
            if (parsed.state.swarmResults) {
              parsed.state.swarmResults = parsed.state.swarmResults.map((r: Record<string, unknown>) => ({
                ...r,
                startedAt: new Date(r.startedAt as string),
                completedAt: r.completedAt ? new Date(r.completedAt as string) : undefined,
              }));
            }
            if (parsed.state.filters?.dateRange) {
              parsed.state.filters.dateRange.start = new Date(parsed.state.filters.dateRange.start);
              parsed.state.filters.dateRange.end = new Date(parsed.state.filters.dateRange.end);
            }
          }
          return parsed;
        },
        setItem: (name, value) => {
          localStorage.setItem(name, JSON.stringify(value));
        },
        removeItem: (name) => {
          localStorage.removeItem(name);
        },
      },
    }
  )
);

// Selectors for computed values
export const selectTotalExecutions = (state: ResultsState): number => {
  return state.batchResults.reduce((sum, b) => sum + b.totalExecutions, 0) +
    state.swarmResults.reduce((sum, s) => sum + s.agentCount, 0);
};

export const selectSuccessRate = (state: ResultsState): number => {
  const metrics = state.aggregatedMetrics;
  if (!metrics || metrics.totalExecutions === 0) {
    // Calculate from results if metrics not loaded
    const totalBatch = state.batchResults.reduce((sum, b) => sum + b.totalExecutions, 0);
    const completedBatch = state.batchResults.reduce((sum, b) => sum + b.completedExecutions, 0);
    const totalSwarm = state.swarmResults.reduce((sum, s) => sum + s.agentCount, 0);
    const completedSwarm = state.swarmResults.reduce((sum, s) => sum + s.completedAgents, 0);
    const total = totalBatch + totalSwarm;
    const completed = completedBatch + completedSwarm;
    return total > 0 ? (completed / total) * 100 : 0;
  }
  return metrics.successRate;
};

export const selectAverageTokenUsage = (state: ResultsState): number => {
  const metrics = state.aggregatedMetrics;
  if (!metrics) {
    // Calculate from results if metrics not loaded
    const totalTokens = state.batchResults.reduce((sum, b) => sum + b.totalTokens, 0) +
      state.swarmResults.reduce((sum, s) => sum + s.totalTokens, 0);
    const totalExecutions = selectTotalExecutions(state);
    return totalExecutions > 0 ? Math.round(totalTokens / totalExecutions) : 0;
  }
  return metrics.averageTokensPerExecution;
};

export const selectTotalCost = (state: ResultsState): number => {
  const metrics = state.aggregatedMetrics;
  if (!metrics) {
    // Calculate from results if metrics not loaded
    return state.batchResults.reduce((sum, b) => sum + b.totalCost, 0) +
      state.swarmResults.reduce((sum, s) => sum + s.totalCost, 0);
  }
  return metrics.totalCost;
};

export const selectByStatus = (state: ResultsState) => (status: ExecutionStatus) => {
  return {
    batch: state.batchResults.filter((b) => b.status === status),
    swarm: state.swarmResults.filter((s) => s.status === status),
  };
};

export const selectRecentExecutions = (state: ResultsState) => (limit: number) => {
  const allResults = [
    ...state.batchResults.map((b) => ({ ...b, type: 'batch' as const })),
    ...state.swarmResults.map((s) => ({ ...s, type: 'swarm' as const })),
  ].sort((a, b) => b.startedAt.getTime() - a.startedAt.getTime());

  return allResults.slice(0, limit);
};

// Export type for external use
export type { ResultsState };
