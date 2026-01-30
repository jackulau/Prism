import { create } from 'zustand';
import type {
  BatchTask,
  BatchConfig,
  BatchExecution,
  BatchExecutionStatus,
  BatchTaskStatus,
  BatchResult,
} from '../types/batch';

interface BatchState {
  // Tasks
  tasks: BatchTask[];
  addTask: (prompt: string, systemPrompt?: string) => void;
  removeTask: (id: string) => void;
  updateTask: (id: string, updates: Partial<BatchTask>) => void;
  clearTasks: () => void;
  reorderTasks: (fromIndex: number, toIndex: number) => void;

  // Configuration
  config: BatchConfig;
  setConfig: (config: Partial<BatchConfig>) => void;

  // Execution
  execution: BatchExecution | null;
  isRunning: boolean;
  startBatch: () => void;
  stopBatch: () => void;
  pauseBatch: () => void;
  resumeBatch: () => void;

  // Results
  results: BatchResult[];
  clearResults: () => void;
  exportResults: (format: 'json' | 'csv') => string;

  // Internal
  _simulateBatchExecution: () => Promise<void>;
}

const defaultConfig: BatchConfig = {
  provider: 'ollama',
  model: '',
  maxConcurrent: 3,
  timeout: 120000,
  temperature: 0.7,
  maxTokens: 4096,
};

export const useBatchStore = create<BatchState>((set, get) => ({
  // Tasks
  tasks: [],
  addTask: (prompt, systemPrompt) => {
    const task: BatchTask = {
      id: crypto.randomUUID(),
      prompt,
      systemPrompt,
      status: 'pending',
    };
    set((state) => ({ tasks: [...state.tasks, task] }));
  },
  removeTask: (id) => {
    set((state) => ({
      tasks: state.tasks.filter((t) => t.id !== id),
    }));
  },
  updateTask: (id, updates) => {
    set((state) => ({
      tasks: state.tasks.map((t) =>
        t.id === id ? { ...t, ...updates } : t
      ),
    }));
  },
  clearTasks: () => set({ tasks: [] }),
  reorderTasks: (fromIndex, toIndex) => {
    set((state) => {
      const newTasks = [...state.tasks];
      const [removed] = newTasks.splice(fromIndex, 1);
      newTasks.splice(toIndex, 0, removed);
      return { tasks: newTasks };
    });
  },

  // Configuration
  config: defaultConfig,
  setConfig: (config) => {
    set((state) => ({
      config: { ...state.config, ...config },
    }));
  },

  // Execution
  execution: null,
  isRunning: false,
  startBatch: () => {
    const { tasks, config } = get();
    if (tasks.length === 0) return;

    const execution: BatchExecution = {
      id: crypto.randomUUID(),
      status: 'running',
      config: { ...config },
      tasks: tasks.map((t) => ({ ...t, status: 'pending' as BatchTaskStatus })),
      startedAt: new Date(),
      totalTasks: tasks.length,
      completedTasks: 0,
      failedTasks: 0,
    };

    set({
      execution,
      isRunning: true,
      tasks: tasks.map((t) => ({ ...t, status: 'pending' as BatchTaskStatus })),
    });

    // Simulate batch execution (in real implementation, this would call the API)
    get()._simulateBatchExecution();
  },
  stopBatch: () => {
    set((state) => ({
      isRunning: false,
      execution: state.execution
        ? { ...state.execution, status: 'cancelled' as BatchExecutionStatus }
        : null,
      tasks: state.tasks.map((t) =>
        t.status === 'running' || t.status === 'pending'
          ? { ...t, status: 'cancelled' as BatchTaskStatus }
          : t
      ),
    }));
  },
  pauseBatch: () => {
    set((state) => ({
      isRunning: false,
      execution: state.execution
        ? { ...state.execution, status: 'paused' as BatchExecutionStatus }
        : null,
    }));
  },
  resumeBatch: () => {
    const { execution } = get();
    if (!execution || execution.status !== 'paused') return;

    set({
      isRunning: true,
      execution: { ...execution, status: 'running' },
    });
    get()._simulateBatchExecution();
  },

  // Results
  results: [],
  clearResults: () => set({ results: [] }),
  exportResults: (format) => {
    const { results } = get();
    if (format === 'json') {
      return JSON.stringify(results, null, 2);
    }
    // CSV format
    const headers = ['Task ID', 'Prompt', 'Status', 'Result', 'Error', 'Duration (ms)', 'Tokens'];
    const rows = results.map((r) => [
      r.taskId,
      `"${r.prompt.replace(/"/g, '""')}"`,
      r.status,
      r.result ? `"${r.result.replace(/"/g, '""')}"` : '',
      r.error || '',
      r.duration?.toString() || '',
      r.tokensUsed?.toString() || '',
    ]);
    return [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
  },

  // Internal: Helper method to simulate batch execution
  _simulateBatchExecution: async () => {
    const { tasks, config, isRunning, updateTask, execution } = get();
    if (!isRunning || !execution) return;

    const pendingTasks = tasks.filter((t) => t.status === 'pending');
    const runningCount = tasks.filter((t) => t.status === 'running').length;
    const availableSlots = config.maxConcurrent - runningCount;

    // Start new tasks up to max concurrent
    const tasksToStart = pendingTasks.slice(0, availableSlots);

    for (const task of tasksToStart) {
      if (!get().isRunning) return;

      updateTask(task.id, {
        status: 'running',
        startedAt: new Date(),
        progress: 0,
      });

      // Simulate async task execution
      simulateTaskExecution(task.id, get, set);
    }
  },
}));

// Simulate a single task execution
async function simulateTaskExecution(
  taskId: string,
  get: () => BatchState,
  set: (state: Partial<BatchState> | ((state: BatchState) => Partial<BatchState>)) => void
) {
  const duration = 2000 + Math.random() * 3000; // 2-5 seconds
  const steps = 10;
  const stepDuration = duration / steps;

  for (let i = 1; i <= steps; i++) {
    await new Promise((resolve) => setTimeout(resolve, stepDuration));

    if (!get().isRunning) return;

    get().updateTask(taskId, { progress: i * 10 });
  }

  const success = Math.random() > 0.1; // 90% success rate

  const task = get().tasks.find((t) => t.id === taskId);
  const completedAt = new Date();
  const startedAt = task?.startedAt || new Date();
  const actualDuration = completedAt.getTime() - startedAt.getTime();

  if (success) {
    get().updateTask(taskId, {
      status: 'completed',
      completedAt,
      duration: actualDuration,
      result: `Simulated response for: "${task?.prompt?.slice(0, 50)}..."`,
      tokensUsed: Math.floor(100 + Math.random() * 500),
      progress: 100,
    });
  } else {
    get().updateTask(taskId, {
      status: 'failed',
      completedAt,
      duration: actualDuration,
      error: 'Simulated error: Task failed due to timeout',
      progress: 100,
    });
  }

  // Update execution stats
  set((state) => {
    const completedTasks = state.tasks.filter((t) => t.status === 'completed').length;
    const failedTasks = state.tasks.filter((t) => t.status === 'failed').length;
    const allDone = completedTasks + failedTasks === state.tasks.length;

    // Add to results
    const updatedTask = state.tasks.find((t) => t.id === taskId);
    const newResult: BatchResult = {
      taskId,
      prompt: updatedTask?.prompt || '',
      status: updatedTask?.status || 'failed',
      result: updatedTask?.result,
      error: updatedTask?.error,
      duration: updatedTask?.duration,
      tokensUsed: updatedTask?.tokensUsed,
    };

    return {
      execution: state.execution
        ? {
            ...state.execution,
            completedTasks,
            failedTasks,
            status: allDone ? 'completed' : state.execution.status,
            completedAt: allDone ? new Date() : undefined,
          }
        : null,
      isRunning: allDone ? false : state.isRunning,
      results: [...state.results, newResult],
    };
  });

  // Continue processing pending tasks
  if (get().isRunning) {
    get()._simulateBatchExecution();
  }
}
