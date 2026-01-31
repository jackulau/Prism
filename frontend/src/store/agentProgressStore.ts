import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type {
  AgentProgress,
  SwarmProgress,
  ProgressMetrics,
  ProgressEvent,
} from '../types';

interface AgentProgressState {
  // State
  activeAgents: Map<string, AgentProgress>;
  activeSwarms: Map<string, SwarmProgress>;
  historicalMetrics: ProgressMetrics[];

  // Agent Actions
  startAgent: (agentId: string, name: string, totalSteps?: number) => void;
  updateProgress: (agentId: string, progress: Partial<AgentProgress>) => void;
  setThinking: (agentId: string, isThinking: boolean) => void;
  completeAgent: (agentId: string, status: 'completed' | 'failed' | 'cancelled') => void;
  clearAgent: (agentId: string) => void;
  addProgressEvent: (agentId: string, event: Omit<ProgressEvent, 'timestamp'>) => void;

  // Swarm Actions
  startSwarm: (swarmId: string, agentIds: string[]) => void;
  updateSwarmProgress: (swarmId: string) => void;
  completeSwarm: (swarmId: string, status: 'completed' | 'failed') => void;
  clearSwarm: (swarmId: string) => void;

  // Selectors
  getAgentProgress: (agentId: string) => AgentProgress | undefined;
  getSwarmProgress: (swarmId: string) => SwarmProgress | undefined;
  getActiveAgentCount: () => number;
  getActiveSwarmCount: () => number;

  // Utility
  clearAllActive: () => void;
  clearHistoricalMetrics: () => void;
}

const MAX_HISTORICAL_METRICS = 100;

const createInitialAgentProgress = (
  agentId: string,
  name: string,
  totalSteps: number
): AgentProgress => ({
  agentId,
  name,
  status: 'pending',
  currentStep: 0,
  totalSteps,
  percentComplete: 0,
  stepName: '',
  message: '',
  startedAt: Date.now(),
  estimatedTimeRemaining: null,
  estimatedTokensRemaining: null,
  isThinking: false,
  thinkingStartedAt: null,
  tokensGenerated: 0,
  events: [],
});

export const useAgentProgressStore = create<AgentProgressState>()(
  persist(
    (set, get) => ({
      // Initial State
      activeAgents: new Map<string, AgentProgress>(),
      activeSwarms: new Map<string, SwarmProgress>(),
      historicalMetrics: [],

      // Agent Actions
      startAgent: (agentId, name, totalSteps = 0) => {
        set((state) => {
          const newAgents = new Map(state.activeAgents);
          newAgents.set(agentId, createInitialAgentProgress(agentId, name, totalSteps));
          return { activeAgents: newAgents };
        });
      },

      updateProgress: (agentId, progress) => {
        set((state) => {
          const agent = state.activeAgents.get(agentId);
          if (!agent) return state;

          const newAgents = new Map(state.activeAgents);
          const updatedAgent = { ...agent, ...progress };

          // Recalculate percent if steps changed
          if (
            (progress.currentStep !== undefined || progress.totalSteps !== undefined) &&
            updatedAgent.totalSteps > 0
          ) {
            updatedAgent.percentComplete = Math.min(
              100,
              Math.round((updatedAgent.currentStep / updatedAgent.totalSteps) * 100)
            );
          }

          newAgents.set(agentId, updatedAgent);
          return { activeAgents: newAgents };
        });
      },

      setThinking: (agentId, isThinking) => {
        set((state) => {
          const agent = state.activeAgents.get(agentId);
          if (!agent) return state;

          const newAgents = new Map(state.activeAgents);
          newAgents.set(agentId, {
            ...agent,
            isThinking,
            thinkingStartedAt: isThinking ? Date.now() : null,
            status: isThinking ? 'thinking' : (agent.status === 'thinking' ? 'running' : agent.status),
          });
          return { activeAgents: newAgents };
        });
      },

      completeAgent: (agentId, status) => {
        set((state) => {
          const agent = state.activeAgents.get(agentId);
          if (!agent) return state;

          const newAgents = new Map(state.activeAgents);
          const completedAgent: AgentProgress = {
            ...agent,
            status,
            isThinking: false,
            thinkingStartedAt: null,
            percentComplete: status === 'completed' ? 100 : agent.percentComplete,
          };
          newAgents.set(agentId, completedAgent);

          // Add to historical metrics
          const metric: ProgressMetrics = {
            agentId,
            duration: Date.now() - agent.startedAt,
            tokensGenerated: agent.tokensGenerated,
            stepsCompleted: agent.currentStep,
            completedAt: Date.now(),
          };

          const newMetrics = [metric, ...state.historicalMetrics].slice(0, MAX_HISTORICAL_METRICS);

          return {
            activeAgents: newAgents,
            historicalMetrics: newMetrics,
          };
        });
      },

      clearAgent: (agentId) => {
        set((state) => {
          const newAgents = new Map(state.activeAgents);
          newAgents.delete(agentId);
          return { activeAgents: newAgents };
        });
      },

      addProgressEvent: (agentId, event) => {
        set((state) => {
          const agent = state.activeAgents.get(agentId);
          if (!agent) return state;

          const newAgents = new Map(state.activeAgents);
          newAgents.set(agentId, {
            ...agent,
            events: [...agent.events, { ...event, timestamp: Date.now() }],
          });
          return { activeAgents: newAgents };
        });
      },

      // Swarm Actions
      startSwarm: (swarmId, agentIds) => {
        set((state) => {
          const newSwarms = new Map(state.activeSwarms);
          const agents = new Map<string, AgentProgress>();

          // Create agent entries in the swarm (agents must be started separately)
          agentIds.forEach((agentId) => {
            const existingAgent = state.activeAgents.get(agentId);
            if (existingAgent) {
              agents.set(agentId, existingAgent);
            }
          });

          newSwarms.set(swarmId, {
            swarmId,
            agents,
            overallPercent: 0,
            completedAgents: 0,
            totalAgents: agentIds.length,
            status: 'running',
          });

          return { activeSwarms: newSwarms };
        });
      },

      updateSwarmProgress: (swarmId) => {
        set((state) => {
          const swarm = state.activeSwarms.get(swarmId);
          if (!swarm) return state;

          const newSwarms = new Map(state.activeSwarms);
          const agents = new Map<string, AgentProgress>();
          let totalPercent = 0;
          let completedCount = 0;

          // Update agent references and calculate progress
          swarm.agents.forEach((_, agentId) => {
            const currentAgent = state.activeAgents.get(agentId);
            if (currentAgent) {
              agents.set(agentId, currentAgent);
              totalPercent += currentAgent.percentComplete;
              if (
                currentAgent.status === 'completed' ||
                currentAgent.status === 'failed' ||
                currentAgent.status === 'cancelled'
              ) {
                completedCount++;
              }
            }
          });

          const overallPercent =
            agents.size > 0 ? Math.round(totalPercent / agents.size) : 0;

          newSwarms.set(swarmId, {
            ...swarm,
            agents,
            overallPercent,
            completedAgents: completedCount,
          });

          return { activeSwarms: newSwarms };
        });
      },

      completeSwarm: (swarmId, status) => {
        set((state) => {
          const swarm = state.activeSwarms.get(swarmId);
          if (!swarm) return state;

          const newSwarms = new Map(state.activeSwarms);
          newSwarms.set(swarmId, {
            ...swarm,
            status,
            overallPercent: status === 'completed' ? 100 : swarm.overallPercent,
          });

          return { activeSwarms: newSwarms };
        });
      },

      clearSwarm: (swarmId) => {
        set((state) => {
          const newSwarms = new Map(state.activeSwarms);
          newSwarms.delete(swarmId);
          return { activeSwarms: newSwarms };
        });
      },

      // Selectors
      getAgentProgress: (agentId) => {
        return get().activeAgents.get(agentId);
      },

      getSwarmProgress: (swarmId) => {
        return get().activeSwarms.get(swarmId);
      },

      getActiveAgentCount: () => {
        return get().activeAgents.size;
      },

      getActiveSwarmCount: () => {
        return get().activeSwarms.size;
      },

      // Utility
      clearAllActive: () => {
        set({
          activeAgents: new Map(),
          activeSwarms: new Map(),
        });
      },

      clearHistoricalMetrics: () => {
        set({ historicalMetrics: [] });
      },
    }),
    {
      name: 'agent-progress-storage',
      // Only persist historical metrics, not active state
      partialize: (state) => ({
        historicalMetrics: state.historicalMetrics,
      }),
      // Custom storage serialization for the persisted data
      storage: {
        getItem: (name) => {
          const str = localStorage.getItem(name);
          if (!str) return null;
          const parsed = JSON.parse(str);
          return {
            state: {
              ...parsed.state,
              // Initialize Maps (not persisted)
              activeAgents: new Map(),
              activeSwarms: new Map(),
            },
            version: parsed.version,
          };
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
