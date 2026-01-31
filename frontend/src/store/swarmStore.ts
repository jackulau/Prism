import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// ============================================================================
// Types
// ============================================================================

/**
 * Orchestration strategies for swarm coordination
 */
export type SwarmStrategy =
  | 'sequential'      // Agents execute in order
  | 'parallel'        // All agents execute simultaneously
  | 'hierarchical'    // Manager agent delegates to workers
  | 'collaborative'   // Agents work together with shared context
  | 'competitive';    // Multiple agents compete, best result wins

/**
 * Possible statuses for a swarm agent
 */
export type SwarmAgentStatus =
  | 'idle'
  | 'running'
  | 'paused'
  | 'completed'
  | 'failed'
  | 'waiting';

/**
 * Possible statuses for a swarm execution
 */
export type SwarmExecutionStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled';

/**
 * Individual agent configuration within a swarm
 */
export interface SwarmAgent {
  id: string;
  name: string;
  role: string;
  description?: string;
  model: string;
  provider: string;
  status: SwarmAgentStatus;
  systemPrompt?: string;
  tools?: string[];
  maxTokens?: number;
  temperature?: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Configuration for a swarm
 */
export interface Swarm {
  id: string;
  name: string;
  description?: string;
  strategy: SwarmStrategy;
  agents: SwarmAgent[];
  maxConcurrency: number;
  timeout?: number;          // Execution timeout in seconds
  retryAttempts?: number;    // Number of retry attempts on failure
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Result from an individual agent execution
 */
export interface AgentExecutionResult {
  agentId: string;
  status: SwarmAgentStatus;
  output?: string;
  error?: string;
  startedAt: string;
  completedAt?: string;
  tokensUsed?: number;
}

/**
 * Record of a swarm execution
 */
export interface SwarmExecution {
  id: string;
  swarmId: string;
  status: SwarmExecutionStatus;
  input: string;
  output?: string;
  error?: string;
  progress: number;          // 0-100
  agentResults: AgentExecutionResult[];
  startedAt: string;
  completedAt?: string;
  totalTokensUsed?: number;
}

// ============================================================================
// Store Interface
// ============================================================================

interface SwarmState {
  // State
  swarms: Swarm[];
  activeSwarm: Swarm | null;
  swarmHistory: SwarmExecution[];
  isConfiguring: boolean;
  isExecuting: boolean;

  // Swarm CRUD actions
  createSwarm: (swarm: Omit<Swarm, 'id' | 'createdAt' | 'updatedAt'>) => Swarm;
  updateSwarm: (id: string, updates: Partial<Omit<Swarm, 'id' | 'createdAt'>>) => void;
  deleteSwarm: (id: string) => void;
  setActiveSwarm: (id: string | null) => void;

  // Agent actions
  addAgent: (swarmId: string, agent: Omit<SwarmAgent, 'id' | 'createdAt' | 'updatedAt'>) => SwarmAgent | null;
  updateAgent: (swarmId: string, agentId: string, updates: Partial<Omit<SwarmAgent, 'id' | 'createdAt'>>) => void;
  removeAgent: (swarmId: string, agentId: string) => void;

  // Strategy action
  setStrategy: (swarmId: string, strategy: SwarmStrategy) => void;

  // UI state actions
  setIsConfiguring: (isConfiguring: boolean) => void;
  setIsExecuting: (isExecuting: boolean) => void;

  // Execution actions
  startExecution: (swarmId: string, input: string) => SwarmExecution | null;
  updateExecution: (executionId: string, updates: Partial<Omit<SwarmExecution, 'id' | 'swarmId' | 'startedAt'>>) => void;
  cancelExecution: (executionId: string) => void;
  updateAgentResult: (executionId: string, agentId: string, result: Partial<AgentExecutionResult>) => void;

  // History actions
  clearHistory: () => void;
  clearSwarmHistory: (swarmId: string) => void;
}

// ============================================================================
// Utility Functions
// ============================================================================

const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
};

const getTimestamp = (): string => {
  return new Date().toISOString();
};

// ============================================================================
// Store Implementation
// ============================================================================

export const useSwarmStore = create<SwarmState>()(
  persist(
    (set, get) => ({
      // Initial state
      swarms: [],
      activeSwarm: null,
      swarmHistory: [],
      isConfiguring: false,
      isExecuting: false,

      // Swarm CRUD actions
      createSwarm: (swarmData) => {
        const timestamp = getTimestamp();
        const newSwarm: Swarm = {
          ...swarmData,
          id: generateId(),
          createdAt: timestamp,
          updatedAt: timestamp,
        };

        set((state) => ({
          swarms: [...state.swarms, newSwarm],
        }));

        return newSwarm;
      },

      updateSwarm: (id, updates) => {
        set((state) => ({
          swarms: state.swarms.map((swarm) =>
            swarm.id === id
              ? { ...swarm, ...updates, updatedAt: getTimestamp() }
              : swarm
          ),
          activeSwarm:
            state.activeSwarm?.id === id
              ? { ...state.activeSwarm, ...updates, updatedAt: getTimestamp() }
              : state.activeSwarm,
        }));
      },

      deleteSwarm: (id) => {
        set((state) => ({
          swarms: state.swarms.filter((swarm) => swarm.id !== id),
          activeSwarm: state.activeSwarm?.id === id ? null : state.activeSwarm,
          swarmHistory: state.swarmHistory.filter((exec) => exec.swarmId !== id),
        }));
      },

      setActiveSwarm: (id) => {
        const { swarms } = get();
        const swarm = id ? swarms.find((s) => s.id === id) || null : null;
        set({ activeSwarm: swarm });
      },

      // Agent actions
      addAgent: (swarmId, agentData) => {
        const timestamp = getTimestamp();
        const newAgent: SwarmAgent = {
          ...agentData,
          id: generateId(),
          createdAt: timestamp,
          updatedAt: timestamp,
        };

        let added = false;
        set((state) => {
          const swarms = state.swarms.map((swarm) => {
            if (swarm.id === swarmId) {
              added = true;
              return {
                ...swarm,
                agents: [...swarm.agents, newAgent],
                updatedAt: timestamp,
              };
            }
            return swarm;
          });

          const activeSwarm =
            state.activeSwarm?.id === swarmId
              ? {
                  ...state.activeSwarm,
                  agents: [...state.activeSwarm.agents, newAgent],
                  updatedAt: timestamp,
                }
              : state.activeSwarm;

          return { swarms, activeSwarm };
        });

        return added ? newAgent : null;
      },

      updateAgent: (swarmId, agentId, updates) => {
        const timestamp = getTimestamp();
        set((state) => {
          const swarms = state.swarms.map((swarm) => {
            if (swarm.id === swarmId) {
              return {
                ...swarm,
                agents: swarm.agents.map((agent) =>
                  agent.id === agentId
                    ? { ...agent, ...updates, updatedAt: timestamp }
                    : agent
                ),
                updatedAt: timestamp,
              };
            }
            return swarm;
          });

          const activeSwarm =
            state.activeSwarm?.id === swarmId
              ? {
                  ...state.activeSwarm,
                  agents: state.activeSwarm.agents.map((agent) =>
                    agent.id === agentId
                      ? { ...agent, ...updates, updatedAt: timestamp }
                      : agent
                  ),
                  updatedAt: timestamp,
                }
              : state.activeSwarm;

          return { swarms, activeSwarm };
        });
      },

      removeAgent: (swarmId, agentId) => {
        const timestamp = getTimestamp();
        set((state) => {
          const swarms = state.swarms.map((swarm) => {
            if (swarm.id === swarmId) {
              return {
                ...swarm,
                agents: swarm.agents.filter((agent) => agent.id !== agentId),
                updatedAt: timestamp,
              };
            }
            return swarm;
          });

          const activeSwarm =
            state.activeSwarm?.id === swarmId
              ? {
                  ...state.activeSwarm,
                  agents: state.activeSwarm.agents.filter(
                    (agent) => agent.id !== agentId
                  ),
                  updatedAt: timestamp,
                }
              : state.activeSwarm;

          return { swarms, activeSwarm };
        });
      },

      // Strategy action
      setStrategy: (swarmId, strategy) => {
        const timestamp = getTimestamp();
        set((state) => ({
          swarms: state.swarms.map((swarm) =>
            swarm.id === swarmId
              ? { ...swarm, strategy, updatedAt: timestamp }
              : swarm
          ),
          activeSwarm:
            state.activeSwarm?.id === swarmId
              ? { ...state.activeSwarm, strategy, updatedAt: timestamp }
              : state.activeSwarm,
        }));
      },

      // UI state actions
      setIsConfiguring: (isConfiguring) => set({ isConfiguring }),
      setIsExecuting: (isExecuting) => set({ isExecuting }),

      // Execution actions
      startExecution: (swarmId, input) => {
        const { swarms } = get();
        const swarm = swarms.find((s) => s.id === swarmId);
        if (!swarm) return null;

        const execution: SwarmExecution = {
          id: generateId(),
          swarmId,
          status: 'running',
          input,
          progress: 0,
          agentResults: swarm.agents.map((agent) => ({
            agentId: agent.id,
            status: 'waiting',
            startedAt: getTimestamp(),
          })),
          startedAt: getTimestamp(),
        };

        set((state) => ({
          swarmHistory: [execution, ...state.swarmHistory],
          isExecuting: true,
        }));

        return execution;
      },

      updateExecution: (executionId, updates) => {
        set((state) => ({
          swarmHistory: state.swarmHistory.map((exec) =>
            exec.id === executionId ? { ...exec, ...updates } : exec
          ),
          isExecuting:
            updates.status && ['completed', 'failed', 'cancelled'].includes(updates.status)
              ? false
              : state.isExecuting,
        }));
      },

      cancelExecution: (executionId) => {
        set((state) => ({
          swarmHistory: state.swarmHistory.map((exec) =>
            exec.id === executionId
              ? {
                  ...exec,
                  status: 'cancelled' as SwarmExecutionStatus,
                  completedAt: getTimestamp(),
                }
              : exec
          ),
          isExecuting: false,
        }));
      },

      updateAgentResult: (executionId, agentId, result) => {
        set((state) => ({
          swarmHistory: state.swarmHistory.map((exec) => {
            if (exec.id === executionId) {
              return {
                ...exec,
                agentResults: exec.agentResults.map((ar) =>
                  ar.agentId === agentId ? { ...ar, ...result } : ar
                ),
              };
            }
            return exec;
          }),
        }));
      },

      // History actions
      clearHistory: () => set({ swarmHistory: [] }),
      clearSwarmHistory: (swarmId) =>
        set((state) => ({
          swarmHistory: state.swarmHistory.filter(
            (exec) => exec.swarmId !== swarmId
          ),
        })),
    }),
    {
      name: 'prism-swarm',
      partialize: (state) => ({
        swarms: state.swarms,
        swarmHistory: state.swarmHistory.slice(0, 50), // Keep last 50 executions
      }),
    }
  )
);

// ============================================================================
// Selectors
// ============================================================================

/**
 * Get all agents from the active swarm
 */
export const selectActiveSwarmAgents = (): SwarmAgent[] => {
  const { activeSwarm } = useSwarmStore.getState();
  return activeSwarm?.agents || [];
};

/**
 * Get all running swarms (swarms with isActive = true)
 */
export const selectRunningSwarms = (): Swarm[] => {
  const { swarms } = useSwarmStore.getState();
  return swarms.filter((swarm) => swarm.isActive);
};

/**
 * Get the current execution (most recent running execution)
 */
export const selectCurrentExecution = (): SwarmExecution | null => {
  const { swarmHistory } = useSwarmStore.getState();
  return swarmHistory.find((exec) => exec.status === 'running') || null;
};

/**
 * Get executions for a specific swarm
 */
export const selectSwarmExecutions = (swarmId: string): SwarmExecution[] => {
  const { swarmHistory } = useSwarmStore.getState();
  return swarmHistory.filter((exec) => exec.swarmId === swarmId);
};

/**
 * Get a swarm by ID
 */
export const selectSwarmById = (id: string): Swarm | null => {
  const { swarms } = useSwarmStore.getState();
  return swarms.find((swarm) => swarm.id === id) || null;
};

/**
 * Get agents by status from active swarm
 */
export const selectAgentsByStatus = (status: SwarmAgentStatus): SwarmAgent[] => {
  const { activeSwarm } = useSwarmStore.getState();
  return activeSwarm?.agents.filter((agent) => agent.status === status) || [];
};

/**
 * Get total agent count across all swarms
 */
export const selectTotalAgentCount = (): number => {
  const { swarms } = useSwarmStore.getState();
  return swarms.reduce((total, swarm) => total + swarm.agents.length, 0);
};

/**
 * Check if any swarm is currently executing
 */
export const selectIsAnySwarmExecuting = (): boolean => {
  const { swarmHistory } = useSwarmStore.getState();
  return swarmHistory.some((exec) => exec.status === 'running');
};
