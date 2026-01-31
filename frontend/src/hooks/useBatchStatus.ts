import { useEffect, useCallback, useRef, useState } from 'react';
import { useBatchStore } from '../store/batchStore';
import type {
  BatchStatus,
  BatchProgressInfo,
  BatchHistoryEntry,
  BatchTask,
} from '../types/batch';

interface UseBatchStatusOptions {
  /** Polling interval in milliseconds (default: 2000) */
  pollingInterval?: number;
  /** Whether to enable polling (default: true when batch is running) */
  enablePolling?: boolean;
  /** API base URL for status requests */
  apiBaseUrl?: string;
  /** Authentication token */
  token?: string;
}

interface UseBatchStatusReturn {
  /** Current batch status */
  status: BatchStatus;
  /** Current progress info */
  progress: BatchProgressInfo;
  /** All tasks in the current batch */
  tasks: BatchTask[];
  /** Failed tasks */
  failedTasks: BatchTask[];
  /** Running tasks */
  runningTasks: BatchTask[];
  /** Completed tasks */
  completedTasks: BatchTask[];
  /** Pending tasks */
  pendingTasks: BatchTask[];
  /** Current batch ID */
  batchId: string | null;
  /** Batch execution history */
  history: BatchHistoryEntry[];
  /** Whether history is loading */
  isLoadingHistory: boolean;
  /** Whether status is being fetched */
  isPolling: boolean;
  /** Last error message */
  error: string | null;
  /** Manually refresh status */
  refreshStatus: () => Promise<void>;
  /** Load batch execution history */
  loadHistory: (limit?: number) => Promise<void>;
  /** Get status of a specific batch */
  getBatchStatus: (batchId: string) => Promise<BatchProgressInfo | null>;
  /** Retry a specific failed task */
  retryTask: (taskId: string) => void;
  /** Retry all failed tasks */
  retryFailedTasks: () => void;
}

/**
 * API response type for batch status
 */
interface BatchStatusResponse {
  batch_id: string;
  status: BatchStatus;
  progress: BatchProgressInfo;
  tasks: Array<{
    id: string;
    name: string;
    status: string;
    error?: string;
    output?: string;
    started_at?: string;
    completed_at?: string;
  }>;
  error?: string;
}

/**
 * API response type for batch history
 */
interface BatchHistoryResponse {
  batches: Array<{
    id: string;
    status: BatchStatus;
    task_count: number;
    success_count: number;
    fail_count: number;
    created_at: string;
    completed_at?: string;
    duration?: number;
  }>;
}

/**
 * Hook for polling batch status and managing batch history
 *
 * This hook provides:
 * - Polling-based status updates as a fallback to WebSocket
 * - Batch execution history
 * - Task retry functionality
 *
 * @example
 * ```tsx
 * const {
 *   status,
 *   progress,
 *   failedTasks,
 *   retryFailedTasks,
 *   loadHistory
 * } = useBatchStatus({
 *   pollingInterval: 3000,
 *   enablePolling: true
 * });
 *
 * // Load history on mount
 * useEffect(() => {
 *   loadHistory(10);
 * }, [loadHistory]);
 *
 * // Retry failed tasks
 * if (failedTasks.length > 0) {
 *   retryFailedTasks();
 * }
 * ```
 */
export function useBatchStatus(options: UseBatchStatusOptions = {}): UseBatchStatusReturn {
  const {
    pollingInterval = 2000,
    enablePolling = true,
    apiBaseUrl = '/api/v1',
    token,
  } = options;

  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Store state
  const store = useBatchStore();
  const {
    execution,
    history,
    isLoadingHistory,
    setHistory,
    setLoadingHistory,
    retryTask: storeRetryTask,
    retryFailedTasks: storeRetryFailedTasks,
    getFailedTasks,
    getRunningTasks,
    getCompletedTasks,
    getPendingTasks,
    getOrderedTasks,
    updateBatchStatus,
    updateProgress,
    setTaskCompleted,
  } = store;

  /**
   * Make an authenticated API request
   */
  const apiRequest = useCallback(
    async <T>(endpoint: string): Promise<T | null> => {
      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
        };

        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${apiBaseUrl}${endpoint}`, {
          method: 'GET',
          headers,
        });

        if (!response.ok) {
          throw new Error(`API request failed: ${response.statusText}`);
        }

        return await response.json();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'API request failed');
        return null;
      }
    },
    [apiBaseUrl, token]
  );

  /**
   * Fetch current batch status from API
   */
  const refreshStatus = useCallback(async () => {
    const { batchId } = execution;
    if (!batchId) return;

    setIsPolling(true);
    setError(null);

    try {
      const response = await apiRequest<BatchStatusResponse>(`/batch/${batchId}/status`);

      if (response) {
        updateBatchStatus(response.status, response.error);
        updateProgress(response.progress);

        // Update individual task states
        for (const task of response.tasks) {
          if (task.status === 'completed' || task.status === 'failed') {
            setTaskCompleted(task.id, task.output, task.error);
          }
        }
      }
    } finally {
      setIsPolling(false);
    }
  }, [execution, apiRequest, updateBatchStatus, updateProgress, setTaskCompleted]);

  /**
   * Get status for a specific batch
   */
  const getBatchStatus = useCallback(
    async (batchId: string): Promise<BatchProgressInfo | null> => {
      const response = await apiRequest<BatchStatusResponse>(`/batch/${batchId}/status`);
      return response?.progress ?? null;
    },
    [apiRequest]
  );

  /**
   * Load batch execution history
   */
  const loadHistory = useCallback(
    async (limit = 20) => {
      setLoadingHistory(true);
      setError(null);

      try {
        const response = await apiRequest<BatchHistoryResponse>(`/batch/history?limit=${limit}`);

        if (response?.batches) {
          const historyEntries: BatchHistoryEntry[] = response.batches.map((b) => ({
            id: b.id,
            status: b.status,
            taskCount: b.task_count,
            successCount: b.success_count,
            failCount: b.fail_count,
            createdAt: new Date(b.created_at),
            completedAt: b.completed_at ? new Date(b.completed_at) : undefined,
            duration: b.duration,
          }));

          setHistory(historyEntries);
        }
      } finally {
        setLoadingHistory(false);
      }
    },
    [apiRequest, setHistory, setLoadingHistory]
  );

  /**
   * Retry a failed task
   */
  const retryTask = useCallback(
    (taskId: string) => {
      storeRetryTask(taskId);
      // Optionally trigger the task execution via API
      // This would be handled by the batch execution system
    },
    [storeRetryTask]
  );

  /**
   * Retry all failed tasks
   */
  const retryFailedTasks = useCallback(() => {
    storeRetryFailedTasks();
    // Optionally trigger the tasks execution via API
  }, [storeRetryFailedTasks]);

  // Set up polling when batch is running
  useEffect(() => {
    const shouldPoll = enablePolling && execution.status === 'running' && execution.batchId;

    if (shouldPoll) {
      // Initial fetch
      refreshStatus();

      // Set up interval
      pollingRef.current = setInterval(refreshStatus, pollingInterval);
    }

    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    };
  }, [enablePolling, execution.status, execution.batchId, pollingInterval, refreshStatus]);

  return {
    status: execution.status,
    progress: execution.progress,
    tasks: getOrderedTasks(),
    failedTasks: getFailedTasks(),
    runningTasks: getRunningTasks(),
    completedTasks: getCompletedTasks(),
    pendingTasks: getPendingTasks(),
    batchId: execution.batchId,
    history,
    isLoadingHistory,
    isPolling,
    error,
    refreshStatus,
    loadHistory,
    getBatchStatus,
    retryTask,
    retryFailedTasks,
  };
}
