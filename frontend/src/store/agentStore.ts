import { create } from 'zustand';

export type AgentExecutionStatus = 'idle' | 'running' | 'complete' | 'error' | 'stopped';

export interface AgentConfig {
  name: string;
  systemPrompt: string;
  provider: string;
  model: string;
  temperature: number;
  maxTokens: number;
}

export interface AgentOutput {
  id: string;
  type: 'text' | 'thinking' | 'tool_call' | 'tool_result' | 'error';
  content: string;
  timestamp: Date;
  toolName?: string;
  toolParams?: Record<string, unknown>;
  toolResult?: unknown;
}

export interface AgentExecution {
  id: string;
  config: AgentConfig;
  status: AgentExecutionStatus;
  outputs: AgentOutput[];
  startedAt: Date;
  completedAt?: Date;
  error?: string;
}

interface AgentState {
  // Current execution
  currentExecution: AgentExecution | null;
  executionStatus: AgentExecutionStatus;
  outputs: AgentOutput[];

  // Configuration
  config: AgentConfig;
  setConfig: (config: Partial<AgentConfig>) => void;
  resetConfig: () => void;

  // Execution history
  recentExecutions: AgentExecution[];

  // Actions
  startExecution: () => string;
  addOutput: (output: Omit<AgentOutput, 'id' | 'timestamp'>) => void;
  appendToLastOutput: (delta: string) => void;
  completeExecution: () => void;
  stopExecution: () => void;
  setExecutionError: (error: string) => void;
  clearOutputs: () => void;
  clearHistory: () => void;
}

const defaultConfig: AgentConfig = {
  name: 'My Agent',
  systemPrompt: 'You are a helpful AI assistant.',
  provider: 'ollama',
  model: '',
  temperature: 0.7,
  maxTokens: 4096,
};

const MAX_HISTORY_SIZE = 10;

export const useAgentStore = create<AgentState>((set, get) => ({
  // Current execution
  currentExecution: null,
  executionStatus: 'idle',
  outputs: [],

  // Configuration
  config: defaultConfig,
  setConfig: (updates) => set((state) => ({
    config: { ...state.config, ...updates },
  })),
  resetConfig: () => set({ config: defaultConfig }),

  // Execution history
  recentExecutions: [],

  // Actions
  startExecution: () => {
    const id = crypto.randomUUID();
    const { config } = get();
    const execution: AgentExecution = {
      id,
      config: { ...config },
      status: 'running',
      outputs: [],
      startedAt: new Date(),
    };

    set({
      currentExecution: execution,
      executionStatus: 'running',
      outputs: [],
    });

    return id;
  },

  addOutput: (output) => {
    const newOutput: AgentOutput = {
      ...output,
      id: crypto.randomUUID(),
      timestamp: new Date(),
    };

    set((state) => ({
      outputs: [...state.outputs, newOutput],
      currentExecution: state.currentExecution
        ? {
            ...state.currentExecution,
            outputs: [...state.currentExecution.outputs, newOutput],
          }
        : null,
    }));
  },

  appendToLastOutput: (delta) => {
    set((state) => {
      if (state.outputs.length === 0) return state;

      const lastOutput = state.outputs[state.outputs.length - 1];
      const updatedOutput = {
        ...lastOutput,
        content: lastOutput.content + delta,
      };

      const newOutputs = [...state.outputs.slice(0, -1), updatedOutput];

      return {
        outputs: newOutputs,
        currentExecution: state.currentExecution
          ? {
              ...state.currentExecution,
              outputs: newOutputs,
            }
          : null,
      };
    });
  },

  completeExecution: () => {
    const { currentExecution, outputs, recentExecutions } = get();

    if (currentExecution) {
      const completedExecution: AgentExecution = {
        ...currentExecution,
        status: 'complete',
        outputs,
        completedAt: new Date(),
      };

      // Add to history, keeping only the most recent
      const newHistory = [completedExecution, ...recentExecutions].slice(0, MAX_HISTORY_SIZE);

      set({
        currentExecution: completedExecution,
        executionStatus: 'complete',
        recentExecutions: newHistory,
      });
    }
  },

  stopExecution: () => {
    const { currentExecution, outputs, recentExecutions } = get();

    if (currentExecution) {
      const stoppedExecution: AgentExecution = {
        ...currentExecution,
        status: 'stopped',
        outputs,
        completedAt: new Date(),
      };

      const newHistory = [stoppedExecution, ...recentExecutions].slice(0, MAX_HISTORY_SIZE);

      set({
        currentExecution: stoppedExecution,
        executionStatus: 'stopped',
        recentExecutions: newHistory,
      });
    }
  },

  setExecutionError: (error) => {
    const { currentExecution, outputs, recentExecutions } = get();

    if (currentExecution) {
      const errorExecution: AgentExecution = {
        ...currentExecution,
        status: 'error',
        outputs,
        completedAt: new Date(),
        error,
      };

      const newHistory = [errorExecution, ...recentExecutions].slice(0, MAX_HISTORY_SIZE);

      set({
        currentExecution: errorExecution,
        executionStatus: 'error',
        recentExecutions: newHistory,
      });
    }
  },

  clearOutputs: () => set({
    currentExecution: null,
    executionStatus: 'idle',
    outputs: [],
  }),

  clearHistory: () => set({ recentExecutions: [] }),
}));
