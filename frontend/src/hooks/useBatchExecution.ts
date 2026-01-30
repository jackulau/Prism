import { useEffect, useCallback, useRef } from 'react';
import { useBatchStore } from '../store/batchStore';
import type {
  BatchTaskConfig,
  BatchExecutionConfig,
  BatchWSMessageIncoming,
  BatchStatus,
} from '../types/batch';

/**
 * WebSocket message types for batch operations
 */
const BATCH_MESSAGE_TYPES = {
  RUN_PARALLEL: 'agent.run_parallel',
  BATCH_PROGRESS: 'agent.batch_progress',
  BATCH_COMPLETED: 'agent.batch_completed',
  TASK_STARTED: 'agent.task_started',
  TASK_PROGRESS: 'agent.task_progress',
  TASK_COMPLETED: 'agent.task_completed',
  BATCH_STOP: 'agent.batch_stop',
} as const;

interface UseBatchExecutionOptions {
  /** WebSocket URL (defaults to current host) */
  wsUrl?: string;
  /** Authentication token for WebSocket */
  token?: string;
  /** Callback when batch starts */
  onBatchStart?: (batchId: string) => void;
  /** Callback when batch completes */
  onBatchComplete?: (batchId: string, status: BatchStatus) => void;
  /** Callback when a task completes */
  onTaskComplete?: (taskId: string, success: boolean) => void;
  /** Callback on error */
  onError?: (error: string) => void;
}

interface UseBatchExecutionReturn {
  /** Start batch execution with configured tasks */
  startBatch: (tasks?: BatchTaskConfig[], config?: BatchExecutionConfig) => void;
  /** Stop the current batch execution */
  stopBatch: () => void;
  /** Add a task to the batch (before starting) */
  addTask: (task: BatchTaskConfig) => void;
  /** Add multiple tasks to the batch */
  addTasks: (tasks: BatchTaskConfig[]) => void;
  /** Remove a task from the batch */
  removeTask: (taskId: string) => void;
  /** Clear all tasks */
  clearTasks: () => void;
  /** Retry a failed task */
  retryTask: (taskId: string) => void;
  /** Retry all failed tasks */
  retryFailedTasks: () => void;
  /** Reset the batch state */
  reset: () => void;
  /** Whether the batch is currently running */
  isRunning: boolean;
  /** Whether tasks can be started (has tasks, not running) */
  canStart: boolean;
  /** Whether WebSocket is connected */
  isConnected: boolean;
  /** Current batch ID */
  batchId: string | null;
  /** Current batch status */
  status: BatchStatus;
}

/**
 * Hook for managing batch execution via WebSocket
 *
 * This hook provides real-time batch execution capabilities including:
 * - Starting parallel agent executions
 * - Receiving progress updates
 * - Handling task completions
 * - Managing batch lifecycle
 *
 * @example
 * ```tsx
 * const {
 *   startBatch,
 *   stopBatch,
 *   addTask,
 *   isRunning,
 *   status
 * } = useBatchExecution({
 *   onBatchComplete: (batchId, status) => {
 *     console.log(`Batch ${batchId} completed with status: ${status}`);
 *   }
 * });
 *
 * // Add tasks
 * addTask({ id: '1', name: 'Task 1', prompt: 'Do something' });
 * addTask({ id: '2', name: 'Task 2', prompt: 'Do something else' });
 *
 * // Start batch
 * startBatch();
 * ```
 */
export function useBatchExecution(options: UseBatchExecutionOptions = {}): UseBatchExecutionReturn {
  const {
    wsUrl,
    token,
    onBatchStart,
    onBatchComplete,
    onTaskComplete,
    onError,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const isConnectedRef = useRef(false);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;

  // Store actions
  const store = useBatchStore();
  const {
    execution,
    addTask: storeAddTask,
    addTasks: storeAddTasks,
    removeTask: storeRemoveTask,
    clearTasks: storeClearTasks,
    startBatch: storeStartBatch,
    stopBatch: storeStopBatch,
    reset: storeReset,
    retryTask: storeRetryTask,
    retryFailedTasks: storeRetryFailedTasks,
    setTaskStarted,
    setTaskCompleted,
    updateTaskOutput,
    updateTaskTokenUsage,
    updateBatchStatus,
    updateProgress,
    setBatchStarted,
    setBatchCompleted,
  } = store;

  /**
   * Handle incoming WebSocket messages
   */
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: BatchWSMessageIncoming = JSON.parse(event.data);

        // Only process batch-related messages
        if (!message.type?.startsWith('agent.')) {
          return;
        }

        switch (message.type) {
          case BATCH_MESSAGE_TYPES.BATCH_PROGRESS:
            if (message.progress) {
              updateProgress(message.progress);
            }
            if (message.status) {
              updateBatchStatus(message.status as BatchStatus);
            }
            break;

          case BATCH_MESSAGE_TYPES.BATCH_COMPLETED:
            setBatchCompleted({
              status: message.status as BatchStatus,
              error: message.error,
            });
            if (message.batch_id) {
              onBatchComplete?.(message.batch_id, message.status as BatchStatus);
            }
            break;

          case BATCH_MESSAGE_TYPES.TASK_STARTED:
            if (message.task_id) {
              setTaskStarted(message.task_id);
            }
            break;

          case BATCH_MESSAGE_TYPES.TASK_PROGRESS:
            if (message.task_id) {
              if (message.output) {
                updateTaskOutput(message.task_id, message.output);
              }
              if (message.token_usage) {
                updateTaskTokenUsage(message.task_id, message.token_usage);
              }
            }
            break;

          case BATCH_MESSAGE_TYPES.TASK_COMPLETED:
            if (message.task_id) {
              const isSuccess = message.status === 'completed';
              setTaskCompleted(
                message.task_id,
                message.output,
                isSuccess ? undefined : message.error
              );
              onTaskComplete?.(message.task_id, isSuccess);
            }
            break;

          default:
            // Unknown message type - ignore
            break;
        }
      } catch {
        // Failed to parse message - ignore
      }
    },
    [
      updateProgress,
      updateBatchStatus,
      setBatchCompleted,
      setTaskStarted,
      setTaskCompleted,
      updateTaskOutput,
      updateTaskTokenUsage,
      onBatchComplete,
      onTaskComplete,
    ]
  );

  /**
   * Connect to WebSocket
   */
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const url = wsUrl || `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws`;

    try {
      wsRef.current = token
        ? new WebSocket(url, ['auth', token])
        : new WebSocket(url);

      wsRef.current.onopen = () => {
        isConnectedRef.current = true;
        reconnectAttemptsRef.current = 0;
      };

      wsRef.current.onmessage = handleMessage;

      wsRef.current.onclose = () => {
        isConnectedRef.current = false;
        attemptReconnect();
      };

      wsRef.current.onerror = () => {
        isConnectedRef.current = false;
        onError?.('WebSocket connection error');
      };
    } catch (err) {
      onError?.(`Failed to connect: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  }, [wsUrl, token, handleMessage, onError]);

  /**
   * Attempt to reconnect with exponential backoff
   */
  const attemptReconnect = useCallback(() => {
    if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
      onError?.('Max reconnection attempts reached');
      return;
    }

    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
    reconnectAttemptsRef.current++;

    reconnectTimeoutRef.current = setTimeout(() => {
      connect();
    }, delay);
  }, [connect, onError]);

  /**
   * Send a message via WebSocket
   */
  const sendMessage = useCallback((message: Record<string, unknown>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      onError?.('WebSocket not connected');
    }
  }, [onError]);

  /**
   * Start batch execution
   */
  const startBatch = useCallback(
    (tasks?: BatchTaskConfig[], config?: BatchExecutionConfig) => {
      // Add tasks if provided
      if (tasks && tasks.length > 0) {
        storeAddTasks(tasks);
      }

      // Start the batch in the store
      const batchId = storeStartBatch(config);

      // Send start message via WebSocket
      sendMessage({
        type: BATCH_MESSAGE_TYPES.RUN_PARALLEL,
        batch_id: batchId,
        tasks: store.getOrderedTasks().map((t) => ({
          id: t.id,
          name: t.name,
          prompt: t.prompt,
          conversationId: t.conversationId,
          provider: t.provider || config?.defaultProvider,
          model: t.model || config?.defaultModel,
          metadata: t.metadata,
        })),
        config: {
          maxConcurrency: config?.maxConcurrency ?? execution.config.maxConcurrency,
          stopOnFirstFailure: config?.stopOnFirstFailure ?? execution.config.stopOnFirstFailure,
          timeoutMs: config?.timeoutMs ?? execution.config.timeoutMs,
          taskTimeoutMs: config?.taskTimeoutMs ?? execution.config.taskTimeoutMs,
        },
      });

      setBatchStarted(batchId);
      onBatchStart?.(batchId);
    },
    [
      storeAddTasks,
      storeStartBatch,
      sendMessage,
      store,
      execution.config,
      setBatchStarted,
      onBatchStart,
    ]
  );

  /**
   * Stop batch execution
   */
  const stopBatch = useCallback(() => {
    const { batchId } = execution;

    if (batchId) {
      sendMessage({
        type: BATCH_MESSAGE_TYPES.BATCH_STOP,
        batch_id: batchId,
      });
    }

    storeStopBatch();
  }, [execution, sendMessage, storeStopBatch]);

  /**
   * Reset and cleanup
   */
  const reset = useCallback(() => {
    storeReset();
  }, [storeReset]);

  // Connect on mount
  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  return {
    startBatch,
    stopBatch,
    addTask: storeAddTask,
    addTasks: storeAddTasks,
    removeTask: storeRemoveTask,
    clearTasks: storeClearTasks,
    retryTask: storeRetryTask,
    retryFailedTasks: storeRetryFailedTasks,
    reset,
    isRunning: execution.status === 'running',
    canStart: store.canStart(),
    isConnected: isConnectedRef.current,
    batchId: execution.batchId,
    status: execution.status,
  };
}
