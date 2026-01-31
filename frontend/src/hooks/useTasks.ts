import { useState, useCallback, useEffect } from 'react';
import { apiService } from '../services/api';

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface Task {
  id: string;
  user_id: string;
  prompt: string;
  context?: string;
  priority: number;
  status: TaskStatus;
  agent_config?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  result?: Record<string, unknown>;
  error?: string;
  callback_url?: string;
  created_at: number;
  started_at?: number;
  completed_at?: number;
}

export interface TaskFiltersState {
  status?: TaskStatus | '';
  search?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface PaginationState {
  page: number;
  limit: number;
  total: number;
}

export interface TaskStats {
  total: number;
  pending: number;
  running: number;
  completed: number;
  failed: number;
}

export function useTasks(filters: TaskFiltersState = {}) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<PaginationState>({
    page: 1,
    limit: 20,
    total: 0,
  });

  const fetchTasks = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    const offset = (pagination.page - 1) * pagination.limit;
    const response = await apiService.getTasks({
      status: filters.status || undefined,
      limit: pagination.limit,
      offset,
    });

    if (response.error) {
      setError(response.error);
      setTasks([]);
    } else if (response.data) {
      let filteredTasks = response.data.tasks;

      // Client-side search filter
      if (filters.search) {
        const searchLower = filters.search.toLowerCase();
        filteredTasks = filteredTasks.filter(
          (task) =>
            task.prompt.toLowerCase().includes(searchLower) ||
            task.id.toLowerCase().includes(searchLower)
        );
      }

      // Client-side date filters
      if (filters.dateFrom) {
        const fromDate = new Date(filters.dateFrom).getTime();
        filteredTasks = filteredTasks.filter((task) => task.created_at >= fromDate);
      }
      if (filters.dateTo) {
        const toDate = new Date(filters.dateTo).getTime() + 24 * 60 * 60 * 1000; // End of day
        filteredTasks = filteredTasks.filter((task) => task.created_at <= toDate);
      }

      setTasks(filteredTasks);
      setPagination((prev) => ({
        ...prev,
        total: response.data?.total ?? 0,
      }));
    }

    setIsLoading(false);
  }, [pagination.page, pagination.limit, filters.status, filters.search, filters.dateFrom, filters.dateTo]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const setPage = useCallback((page: number) => {
    setPagination((prev) => ({ ...prev, page }));
  }, []);

  const refetch = useCallback(() => {
    fetchTasks();
  }, [fetchTasks]);

  return {
    tasks,
    isLoading,
    error,
    pagination,
    setPage,
    refetch,
  };
}

export function useTaskStats() {
  const [stats, setStats] = useState<TaskStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    const response = await apiService.getTaskStats();

    if (response.error) {
      setError(response.error);
      setStats(null);
    } else if (response.data) {
      setStats(response.data);
    }

    setIsLoading(false);
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  return {
    stats,
    isLoading,
    error,
    refetch: fetchStats,
  };
}

export function useCancelTask() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const cancelTask = useCallback(async (taskId: string): Promise<boolean> => {
    setIsLoading(true);
    setError(null);

    const response = await apiService.cancelTask(taskId);

    setIsLoading(false);

    if (response.error) {
      setError(response.error);
      return false;
    }

    return response.data?.success ?? false;
  }, []);

  return {
    cancelTask,
    isLoading,
    error,
  };
}

export function useRetryTask() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const retryTask = useCallback(async (taskId: string): Promise<boolean> => {
    setIsLoading(true);
    setError(null);

    const response = await apiService.retryTask(taskId);

    setIsLoading(false);

    if (response.error) {
      setError(response.error);
      return false;
    }

    return response.data?.success ?? false;
  }, []);

  return {
    retryTask,
    isLoading,
    error,
  };
}

export function useTaskFilters() {
  const [filters, setFilters] = useState<TaskFiltersState>({
    status: '',
    search: '',
    dateFrom: '',
    dateTo: '',
  });

  const updateFilter = useCallback(<K extends keyof TaskFiltersState>(
    key: K,
    value: TaskFiltersState[K]
  ) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  }, []);

  const clearFilters = useCallback(() => {
    setFilters({
      status: '',
      search: '',
      dateFrom: '',
      dateTo: '',
    });
  }, []);

  const hasActiveFilters = filters.status !== '' ||
    filters.search !== '' ||
    filters.dateFrom !== '' ||
    filters.dateTo !== '';

  return {
    filters,
    setFilters,
    updateFilter,
    clearFilters,
    hasActiveFilters,
  };
}
