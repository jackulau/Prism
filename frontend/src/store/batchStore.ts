import { create } from 'zustand';
import type {
  BatchTask,
  BatchTaskConfig,
  BatchTaskStatus,
  BatchStatus,
  BatchProgressInfo,
  BatchExecutionState,
  BatchExecutionConfig,
  BatchResult,
  BatchHistoryEntry,
} from '../types/batch';

/**
 * Default batch execution configuration
 */
const DEFAULT_CONFIG: BatchExecutionConfig = {
  maxConcurrency: 3,
  stopOnFirstFailure: false,
  timeoutMs: 3600000, // 1 hour
  taskTimeoutMs: 300000, // 5 minutes
};

/**
 * Initial progress state
 */
const INITIAL_PROGRESS: BatchProgressInfo = {
  total: 0,
  completed: 0,
  succeeded: 0,
  failed: 0,
  running: 0,
  pending: 0,
  percentage: 0,
};

/**
 * Initial batch execution state
 */
const INITIAL_STATE: BatchExecutionState = {
  batchId: null,
  status: 'pending',
  tasks: new Map(),
  taskOrder: [],
  progress: { ...INITIAL_PROGRESS },
  config: { ...DEFAULT_CONFIG },
  createdAt: null,
  startedAt: null,
  completedAt: null,
  error: null,
};

interface BatchStore {
  // State
  execution: BatchExecutionState;
  history: BatchHistoryEntry[];
  isLoadingHistory: boolean;

  // Task Actions
  addTask: (task: BatchTaskConfig) => void;
  addTasks: (tasks: BatchTaskConfig[]) => void;
  removeTask: (taskId: string) => void;
  clearTasks: () => void;
  updateTaskPriority: (taskId: string, priority: number) => void;
  reorderTasks: (taskIds: string[]) => void;

  // Execution Actions
  startBatch: (config?: Partial<BatchExecutionConfig>) => string;
  stopBatch: () => void;
  pauseBatch: () => void;
  resumeBatch: () => void;
  reset: () => void;

  // Task Status Updates (called by WebSocket handlers)
  updateTaskStatus: (taskId: string, status: BatchTaskStatus, error?: string) => void;
  updateTaskOutput: (taskId: string, output: string) => void;
  updateTaskTokenUsage: (taskId: string, usage: { input: number; output: number; total: number }) => void;
  setTaskStarted: (taskId: string) => void;
  setTaskCompleted: (taskId: string, output?: string, error?: string) => void;

  // Batch Status Updates (called by WebSocket handlers)
  updateBatchStatus: (status: BatchStatus, error?: string) => void;
  updateProgress: (progress: Partial<BatchProgressInfo>) => void;
  setBatchStarted: (batchId: string) => void;
  setBatchCompleted: (result: Partial<BatchResult>) => void;

  // Retry Actions
  retryTask: (taskId: string) => void;
  retryFailedTasks: () => void;

  // History Actions
  setHistory: (history: BatchHistoryEntry[]) => void;
  addToHistory: (entry: BatchHistoryEntry) => void;
  clearHistory: () => void;
  setLoadingHistory: (loading: boolean) => void;

  // Computed/Derived
  getTask: (taskId: string) => BatchTask | undefined;
  getTasksByStatus: (status: BatchTaskStatus) => BatchTask[];
  getFailedTasks: () => BatchTask[];
  getPendingTasks: () => BatchTask[];
  getRunningTasks: () => BatchTask[];
  getCompletedTasks: () => BatchTask[];
  getOrderedTasks: () => BatchTask[];
  canStart: () => boolean;
  isRunning: () => boolean;
}

/**
 * Calculate progress from current task states
 */
function calculateProgress(tasks: Map<string, BatchTask>): BatchProgressInfo {
  let total = 0;
  let completed = 0;
  let succeeded = 0;
  let failed = 0;
  let running = 0;
  let pending = 0;

  for (const task of tasks.values()) {
    total++;
    switch (task.status) {
      case 'completed':
        completed++;
        succeeded++;
        break;
      case 'failed':
      case 'cancelled':
        completed++;
        failed++;
        break;
      case 'running':
        running++;
        break;
      case 'pending':
      case 'queued':
        pending++;
        break;
    }
  }

  const percentage = total > 0 ? Math.round((completed / total) * 100) : 0;

  return {
    total,
    completed,
    succeeded,
    failed,
    running,
    pending,
    percentage,
  };
}

/**
 * Generate a unique batch ID
 */
function generateBatchId(): string {
  return `batch-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * Generate a unique task ID if not provided
 */
function generateTaskId(): string {
  return `task-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

export const useBatchStore = create<BatchStore>((set, get) => ({
  // Initial state
  execution: { ...INITIAL_STATE },
  history: [],
  isLoadingHistory: false,

  // Task Actions
  addTask: (taskConfig) => {
    const taskId = taskConfig.id || generateTaskId();
    const task: BatchTask = {
      ...taskConfig,
      id: taskId,
      status: 'pending',
      retryCount: 0,
    };

    set((state) => {
      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, task);

      const newOrder = [...state.execution.taskOrder, taskId];

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          taskOrder: newOrder,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  addTasks: (taskConfigs) => {
    set((state) => {
      const newTasks = new Map(state.execution.tasks);
      const newOrder = [...state.execution.taskOrder];

      for (const taskConfig of taskConfigs) {
        const taskId = taskConfig.id || generateTaskId();
        const task: BatchTask = {
          ...taskConfig,
          id: taskId,
          status: 'pending',
          retryCount: 0,
        };
        newTasks.set(taskId, task);
        newOrder.push(taskId);
      }

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          taskOrder: newOrder,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  removeTask: (taskId) => {
    set((state) => {
      const newTasks = new Map(state.execution.tasks);
      newTasks.delete(taskId);

      const newOrder = state.execution.taskOrder.filter((id) => id !== taskId);

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          taskOrder: newOrder,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  clearTasks: () => {
    set((state) => ({
      execution: {
        ...state.execution,
        tasks: new Map(),
        taskOrder: [],
        progress: { ...INITIAL_PROGRESS },
      },
    }));
  },

  updateTaskPriority: (taskId, priority) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, { ...task, priority });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
        },
      };
    });
  },

  reorderTasks: (taskIds) => {
    set((state) => ({
      execution: {
        ...state.execution,
        taskOrder: taskIds,
      },
    }));
  },

  // Execution Actions
  startBatch: (config) => {
    const batchId = generateBatchId();

    set((state) => {
      // Set all pending tasks to queued
      const newTasks = new Map(state.execution.tasks);
      for (const [id, task] of newTasks) {
        if (task.status === 'pending') {
          newTasks.set(id, { ...task, status: 'queued' });
        }
      }

      return {
        execution: {
          ...state.execution,
          batchId,
          status: 'running',
          tasks: newTasks,
          config: { ...DEFAULT_CONFIG, ...state.execution.config, ...config },
          createdAt: state.execution.createdAt || new Date(),
          startedAt: new Date(),
          progress: calculateProgress(newTasks),
        },
      };
    });

    return batchId;
  },

  stopBatch: () => {
    set((state) => {
      // Cancel all non-completed tasks
      const newTasks = new Map(state.execution.tasks);
      for (const [id, task] of newTasks) {
        if (task.status === 'running' || task.status === 'queued' || task.status === 'pending') {
          newTasks.set(id, { ...task, status: 'cancelled' });
        }
      }

      return {
        execution: {
          ...state.execution,
          status: 'cancelled',
          tasks: newTasks,
          completedAt: new Date(),
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  pauseBatch: () => {
    // Pause is effectively just stopping new tasks from starting
    // Running tasks continue until completion
    set((state) => ({
      execution: {
        ...state.execution,
        status: 'pending', // Use pending as "paused" state
      },
    }));
  },

  resumeBatch: () => {
    set((state) => ({
      execution: {
        ...state.execution,
        status: 'running',
      },
    }));
  },

  reset: () => {
    set({
      execution: { ...INITIAL_STATE },
    });
  },

  // Task Status Updates
  updateTaskStatus: (taskId, status, error) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        status,
        error: error || task.error,
      });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  updateTaskOutput: (taskId, output) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        output: (task.output || '') + output,
      });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
        },
      };
    });
  },

  updateTaskTokenUsage: (taskId, usage) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        tokenUsage: usage,
      });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
        },
      };
    });
  },

  setTaskStarted: (taskId) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        status: 'running',
        startedAt: new Date(),
      });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  setTaskCompleted: (taskId, output, error) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        status: error ? 'failed' : 'completed',
        output: output || task.output,
        error,
        completedAt: new Date(),
      });

      const newProgress = calculateProgress(newTasks);

      // Check if batch is complete
      let newStatus = state.execution.status;
      if (newProgress.pending === 0 && newProgress.running === 0) {
        if (newProgress.failed === 0) {
          newStatus = 'completed';
        } else if (newProgress.succeeded > 0) {
          newStatus = 'partially_completed';
        } else {
          newStatus = 'failed';
        }
      }

      return {
        execution: {
          ...state.execution,
          status: newStatus,
          tasks: newTasks,
          progress: newProgress,
          completedAt: newStatus !== 'running' ? new Date() : state.execution.completedAt,
        },
      };
    });
  },

  // Batch Status Updates
  updateBatchStatus: (status, error) => {
    set((state) => ({
      execution: {
        ...state.execution,
        status,
        error: error || state.execution.error,
      },
    }));
  },

  updateProgress: (progress) => {
    set((state) => ({
      execution: {
        ...state.execution,
        progress: {
          ...state.execution.progress,
          ...progress,
        },
      },
    }));
  },

  setBatchStarted: (batchId) => {
    set((state) => ({
      execution: {
        ...state.execution,
        batchId,
        status: 'running',
        startedAt: new Date(),
      },
    }));
  },

  setBatchCompleted: (result) => {
    set((state) => ({
      execution: {
        ...state.execution,
        status: result.status || 'completed',
        completedAt: new Date(),
        error: result.error ?? null,
      },
    }));
  },

  // Retry Actions
  retryTask: (taskId) => {
    set((state) => {
      const task = state.execution.tasks.get(taskId);
      if (!task || task.status !== 'failed') return state;

      const maxRetries = task.maxRetries ?? 3;
      if (task.retryCount >= maxRetries) return state;

      const newTasks = new Map(state.execution.tasks);
      newTasks.set(taskId, {
        ...task,
        status: 'pending',
        retryCount: task.retryCount + 1,
        error: undefined,
        output: undefined,
        startedAt: undefined,
        completedAt: undefined,
      });

      return {
        execution: {
          ...state.execution,
          tasks: newTasks,
          progress: calculateProgress(newTasks),
        },
      };
    });
  },

  retryFailedTasks: () => {
    set((state) => {
      const newTasks = new Map(state.execution.tasks);
      let hasRetries = false;

      for (const [id, task] of newTasks) {
        if (task.status === 'failed') {
          const maxRetries = task.maxRetries ?? 3;
          if (task.retryCount < maxRetries) {
            newTasks.set(id, {
              ...task,
              status: 'pending',
              retryCount: task.retryCount + 1,
              error: undefined,
              output: undefined,
              startedAt: undefined,
              completedAt: undefined,
            });
            hasRetries = true;
          }
        }
      }

      if (!hasRetries) return state;

      return {
        execution: {
          ...state.execution,
          status: 'running',
          tasks: newTasks,
          progress: calculateProgress(newTasks),
          completedAt: null,
        },
      };
    });
  },

  // History Actions
  setHistory: (history) => {
    set({ history });
  },

  addToHistory: (entry) => {
    set((state) => ({
      history: [entry, ...state.history],
    }));
  },

  clearHistory: () => {
    set({ history: [] });
  },

  setLoadingHistory: (loading) => {
    set({ isLoadingHistory: loading });
  },

  // Computed/Derived
  getTask: (taskId) => {
    return get().execution.tasks.get(taskId);
  },

  getTasksByStatus: (status) => {
    const tasks: BatchTask[] = [];
    for (const task of get().execution.tasks.values()) {
      if (task.status === status) {
        tasks.push(task);
      }
    }
    return tasks;
  },

  getFailedTasks: () => {
    return get().getTasksByStatus('failed');
  },

  getPendingTasks: () => {
    const tasks: BatchTask[] = [];
    for (const task of get().execution.tasks.values()) {
      if (task.status === 'pending' || task.status === 'queued') {
        tasks.push(task);
      }
    }
    return tasks;
  },

  getRunningTasks: () => {
    return get().getTasksByStatus('running');
  },

  getCompletedTasks: () => {
    return get().getTasksByStatus('completed');
  },

  getOrderedTasks: () => {
    const { tasks, taskOrder } = get().execution;
    return taskOrder.map((id) => tasks.get(id)).filter((t): t is BatchTask => t !== undefined);
  },

  canStart: () => {
    const { status, tasks } = get().execution;
    return (status === 'pending' || status === 'cancelled') && tasks.size > 0;
  },

  isRunning: () => {
    return get().execution.status === 'running';
  },
}));
