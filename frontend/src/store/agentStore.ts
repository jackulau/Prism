import { create } from 'zustand';
import type {
  AgentConfig,
  AgentExecutionStatus,
  AgentExecutionState,
  AgentToolCall,
  AgentResult,
} from '../types/agent';

interface AgentStoreState {
  // Current agent execution state
  currentExecution: AgentExecutionState;

  // History of completed executions
  executionHistory: AgentResult[];

  // Active agents (for multi-agent scenarios)
  activeAgents: Map<string, AgentExecutionState>;

  // Actions - Agent Lifecycle
  startAgent: (config: AgentConfig, agentId: string) => void;
  stopAgent: (agentId?: string) => void;
  resetAgent: () => void;

  // Actions - Output
  appendOutput: (chunk: string, agentId?: string) => void;
  clearOutput: (agentId?: string) => void;

  // Actions - Status Updates
  updateStatus: (status: AgentExecutionStatus, agentId?: string) => void;
  setError: (error: string | null, agentId?: string) => void;

  // Actions - Tool Calls
  addToolCall: (toolCall: Omit<AgentToolCall, 'status'>, agentId?: string) => void;
  updateToolCallStatus: (
    toolCallId: string,
    status: AgentToolCall['status'],
    result?: unknown,
    agentId?: string
  ) => void;

  // Actions - Iteration
  incrementIteration: (agentId?: string) => void;

  // Actions - Completion
  completeExecution: (result: AgentResult) => void;
  failExecution: (error: string, agentId?: string) => void;
  cancelExecution: (agentId?: string) => void;

  // Getters
  getExecution: (agentId: string) => AgentExecutionState | undefined;
  isRunning: (agentId?: string) => boolean;
}

const initialExecutionState: AgentExecutionState = {
  agentId: null,
  config: null,
  status: 'idle',
  output: [],
  error: null,
  startedAt: null,
  completedAt: null,
  iterationCount: 0,
  lastToolCall: undefined,
};

export const useAgentStore = create<AgentStoreState>((set, get) => ({
  // Initial state
  currentExecution: { ...initialExecutionState },
  executionHistory: [],
  activeAgents: new Map(),

  // Agent Lifecycle
  startAgent: (config, agentId) => {
    const newExecution: AgentExecutionState = {
      agentId,
      config,
      status: 'running',
      output: [],
      error: null,
      startedAt: new Date(),
      completedAt: null,
      iterationCount: 0,
      lastToolCall: undefined,
    };

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      newActiveAgents.set(agentId, newExecution);
      return {
        currentExecution: newExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  stopAgent: (agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          status: 'cancelled',
          completedAt: new Date(),
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, status: 'cancelled' as const, completedAt: new Date() }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  resetAgent: () => {
    set({
      currentExecution: { ...initialExecutionState },
    });
  },

  // Output
  appendOutput: (chunk, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          output: [...execution.output, chunk],
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, output: [...state.currentExecution.output, chunk] }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  clearOutput: (agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          output: [],
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, output: [] }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  // Status Updates
  updateStatus: (status, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          status,
          completedAt: ['completed', 'failed', 'cancelled'].includes(status) ? new Date() : null,
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? {
              ...state.currentExecution,
              status,
              completedAt: ['completed', 'failed', 'cancelled'].includes(status)
                ? new Date()
                : null,
            }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  setError: (error, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          error,
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, error }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  // Tool Calls
  addToolCall: (toolCall, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    const newToolCall: AgentToolCall = {
      ...toolCall,
      status: 'running',
      startedAt: new Date(),
    };

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          lastToolCall: newToolCall,
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, lastToolCall: newToolCall }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  updateToolCallStatus: (toolCallId, status, result, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution && execution.lastToolCall?.id === toolCallId) {
        newActiveAgents.set(targetId, {
          ...execution,
          lastToolCall: {
            ...execution.lastToolCall,
            status,
            result,
            completedAt: ['completed', 'failed'].includes(status) ? new Date() : undefined,
          },
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId &&
        state.currentExecution.lastToolCall?.id === toolCallId
          ? {
              ...state.currentExecution,
              lastToolCall: {
                ...state.currentExecution.lastToolCall,
                status,
                result,
                completedAt: ['completed', 'failed'].includes(status) ? new Date() : undefined,
              },
            }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  // Iteration
  incrementIteration: (agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          iterationCount: execution.iterationCount + 1,
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? { ...state.currentExecution, iterationCount: state.currentExecution.iterationCount + 1 }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  // Completion
  completeExecution: (result) => {
    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      newActiveAgents.delete(result.agentId);

      const newCurrentExecution =
        state.currentExecution.agentId === result.agentId
          ? { ...state.currentExecution, status: 'completed' as const, completedAt: new Date() }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
        executionHistory: [...state.executionHistory, result],
      };
    });
  },

  failExecution: (error, agentId) => {
    const targetId = agentId || get().currentExecution.agentId;
    if (!targetId) return;

    set((state) => {
      const newActiveAgents = new Map(state.activeAgents);
      const execution = newActiveAgents.get(targetId);
      if (execution) {
        newActiveAgents.set(targetId, {
          ...execution,
          status: 'failed',
          error,
          completedAt: new Date(),
        });
      }

      const newCurrentExecution =
        state.currentExecution.agentId === targetId
          ? {
              ...state.currentExecution,
              status: 'failed' as const,
              error,
              completedAt: new Date(),
            }
          : state.currentExecution;

      return {
        currentExecution: newCurrentExecution,
        activeAgents: newActiveAgents,
      };
    });
  },

  cancelExecution: (agentId) => {
    get().stopAgent(agentId);
  },

  // Getters
  getExecution: (agentId) => {
    return get().activeAgents.get(agentId);
  },

  isRunning: (agentId) => {
    if (agentId) {
      const execution = get().activeAgents.get(agentId);
      return execution?.status === 'running';
    }
    return get().currentExecution.status === 'running';
  },
}));
