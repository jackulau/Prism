import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAgentProgressStore } from './agentProgressStore';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(global, 'localStorage', { value: localStorageMock });

describe('AgentProgressStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAgentProgressStore.setState({
      activeAgents: new Map(),
      activeSwarms: new Map(),
      historicalMetrics: [],
    });
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe('Agent Actions', () => {
    it('should start an agent with initial state', () => {
      const { startAgent, getAgentProgress } = useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent', 5);

      const agent = getAgentProgress('agent-1');
      expect(agent).toBeDefined();
      expect(agent?.agentId).toBe('agent-1');
      expect(agent?.name).toBe('Test Agent');
      expect(agent?.status).toBe('pending');
      expect(agent?.totalSteps).toBe(5);
      expect(agent?.currentStep).toBe(0);
      expect(agent?.percentComplete).toBe(0);
      expect(agent?.isThinking).toBe(false);
    });

    it('should update agent progress', () => {
      const { startAgent, updateProgress, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent', 4);
      updateProgress('agent-1', {
        currentStep: 2,
        stepName: 'Processing',
        message: 'Working...',
        status: 'running',
      });

      const agent = getAgentProgress('agent-1');
      expect(agent?.currentStep).toBe(2);
      expect(agent?.percentComplete).toBe(50);
      expect(agent?.stepName).toBe('Processing');
      expect(agent?.message).toBe('Working...');
      expect(agent?.status).toBe('running');
    });

    it('should calculate percent complete correctly', () => {
      const { startAgent, updateProgress, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent', 10);
      updateProgress('agent-1', { currentStep: 7 });

      const agent = getAgentProgress('agent-1');
      expect(agent?.percentComplete).toBe(70);
    });

    it('should cap percent at 100', () => {
      const { startAgent, updateProgress, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent', 5);
      updateProgress('agent-1', { currentStep: 10 });

      const agent = getAgentProgress('agent-1');
      expect(agent?.percentComplete).toBe(100);
    });

    it('should set thinking state', () => {
      const { startAgent, setThinking, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent');

      setThinking('agent-1', true);
      let agent = getAgentProgress('agent-1');
      expect(agent?.isThinking).toBe(true);
      expect(agent?.status).toBe('thinking');
      expect(agent?.thinkingStartedAt).toBeDefined();

      setThinking('agent-1', false);
      agent = getAgentProgress('agent-1');
      expect(agent?.isThinking).toBe(false);
      expect(agent?.status).toBe('running');
      expect(agent?.thinkingStartedAt).toBe(null);
    });

    it('should complete an agent and add to historical metrics', () => {
      const { startAgent, completeAgent, getAgentProgress, historicalMetrics } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent', 5);

      // Get fresh state after mutation
      useAgentProgressStore.getState().updateProgress('agent-1', {
        currentStep: 5,
        tokensGenerated: 100,
      });

      useAgentProgressStore.getState().completeAgent('agent-1', 'completed');

      const state = useAgentProgressStore.getState();
      const agent = state.getAgentProgress('agent-1');

      expect(agent?.status).toBe('completed');
      expect(agent?.percentComplete).toBe(100);
      expect(state.historicalMetrics.length).toBe(1);
      expect(state.historicalMetrics[0].agentId).toBe('agent-1');
      expect(state.historicalMetrics[0].tokensGenerated).toBe(100);
    });

    it('should clear an agent', () => {
      const { startAgent, clearAgent, getAgentProgress, getActiveAgentCount } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent');
      expect(getActiveAgentCount()).toBe(1);

      clearAgent('agent-1');
      expect(getAgentProgress('agent-1')).toBeUndefined();
      expect(useAgentProgressStore.getState().getActiveAgentCount()).toBe(0);
    });

    it('should add progress events', () => {
      const { startAgent, addProgressEvent, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Test Agent');
      addProgressEvent('agent-1', { type: 'step_start', data: { step: 1 } });

      const agent = useAgentProgressStore.getState().getAgentProgress('agent-1');
      expect(agent?.events.length).toBe(1);
      expect(agent?.events[0].type).toBe('step_start');
      expect(agent?.events[0].data).toEqual({ step: 1 });
      expect(agent?.events[0].timestamp).toBeDefined();
    });

    it('should not update non-existent agent', () => {
      const { updateProgress, getAgentProgress } =
        useAgentProgressStore.getState();

      updateProgress('non-existent', { currentStep: 5 });
      expect(getAgentProgress('non-existent')).toBeUndefined();
    });
  });

  describe('Swarm Actions', () => {
    it('should start a swarm', () => {
      const { startAgent, startSwarm, getSwarmProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1');
      startAgent('agent-2', 'Agent 2');

      useAgentProgressStore.getState().startSwarm('swarm-1', ['agent-1', 'agent-2']);

      const swarm = useAgentProgressStore.getState().getSwarmProgress('swarm-1');
      expect(swarm).toBeDefined();
      expect(swarm?.swarmId).toBe('swarm-1');
      expect(swarm?.totalAgents).toBe(2);
      expect(swarm?.status).toBe('running');
      expect(swarm?.overallPercent).toBe(0);
    });

    it('should update swarm progress based on agent progress', () => {
      const { startAgent, startSwarm, updateProgress, updateSwarmProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1', 10);
      startAgent('agent-2', 'Agent 2', 10);

      useAgentProgressStore.getState().startSwarm('swarm-1', ['agent-1', 'agent-2']);
      useAgentProgressStore.getState().updateProgress('agent-1', { currentStep: 5 });
      useAgentProgressStore.getState().updateProgress('agent-2', { currentStep: 10 });
      useAgentProgressStore.getState().updateSwarmProgress('swarm-1');

      const swarm = useAgentProgressStore.getState().getSwarmProgress('swarm-1');
      expect(swarm?.overallPercent).toBe(75); // (50 + 100) / 2
    });

    it('should complete a swarm', () => {
      const { startSwarm, completeSwarm, getSwarmProgress } =
        useAgentProgressStore.getState();

      startSwarm('swarm-1', []);
      completeSwarm('swarm-1', 'completed');

      const swarm = useAgentProgressStore.getState().getSwarmProgress('swarm-1');
      expect(swarm?.status).toBe('completed');
      expect(swarm?.overallPercent).toBe(100);
    });

    it('should clear a swarm', () => {
      const { startSwarm, clearSwarm, getSwarmProgress, getActiveSwarmCount } =
        useAgentProgressStore.getState();

      startSwarm('swarm-1', []);
      expect(getActiveSwarmCount()).toBe(1);

      clearSwarm('swarm-1');
      expect(useAgentProgressStore.getState().getSwarmProgress('swarm-1')).toBeUndefined();
      expect(useAgentProgressStore.getState().getActiveSwarmCount()).toBe(0);
    });

    it('should track completed agents count', () => {
      const { startAgent, startSwarm, completeAgent, updateSwarmProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1');
      startAgent('agent-2', 'Agent 2');

      useAgentProgressStore.getState().startSwarm('swarm-1', ['agent-1', 'agent-2']);
      useAgentProgressStore.getState().completeAgent('agent-1', 'completed');
      useAgentProgressStore.getState().updateSwarmProgress('swarm-1');

      const swarm = useAgentProgressStore.getState().getSwarmProgress('swarm-1');
      expect(swarm?.completedAgents).toBe(1);
    });
  });

  describe('Selectors', () => {
    it('should get active agent count', () => {
      const { startAgent, getActiveAgentCount } =
        useAgentProgressStore.getState();

      expect(getActiveAgentCount()).toBe(0);

      startAgent('agent-1', 'Agent 1');
      expect(useAgentProgressStore.getState().getActiveAgentCount()).toBe(1);

      useAgentProgressStore.getState().startAgent('agent-2', 'Agent 2');
      expect(useAgentProgressStore.getState().getActiveAgentCount()).toBe(2);
    });

    it('should get active swarm count', () => {
      const { startSwarm, getActiveSwarmCount } =
        useAgentProgressStore.getState();

      expect(getActiveSwarmCount()).toBe(0);

      startSwarm('swarm-1', []);
      expect(useAgentProgressStore.getState().getActiveSwarmCount()).toBe(1);
    });
  });

  describe('Utility Actions', () => {
    it('should clear all active agents and swarms', () => {
      const { startAgent, startSwarm, clearAllActive, getActiveAgentCount, getActiveSwarmCount } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1');
      startAgent('agent-2', 'Agent 2');
      useAgentProgressStore.getState().startSwarm('swarm-1', []);

      useAgentProgressStore.getState().clearAllActive();

      const state = useAgentProgressStore.getState();
      expect(state.getActiveAgentCount()).toBe(0);
      expect(state.getActiveSwarmCount()).toBe(0);
    });

    it('should clear historical metrics', () => {
      const { startAgent, completeAgent, clearHistoricalMetrics } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1');
      useAgentProgressStore.getState().completeAgent('agent-1', 'completed');

      expect(useAgentProgressStore.getState().historicalMetrics.length).toBe(1);

      useAgentProgressStore.getState().clearHistoricalMetrics();
      expect(useAgentProgressStore.getState().historicalMetrics.length).toBe(0);
    });

    it('should limit historical metrics to max 100', () => {
      const { startAgent, completeAgent } = useAgentProgressStore.getState();

      // Create 105 agents and complete them
      for (let i = 0; i < 105; i++) {
        useAgentProgressStore.getState().startAgent(`agent-${i}`, `Agent ${i}`);
        useAgentProgressStore.getState().completeAgent(`agent-${i}`, 'completed');
      }

      const state = useAgentProgressStore.getState();
      expect(state.historicalMetrics.length).toBe(100);
    });
  });

  describe('Concurrent Agents', () => {
    it('should handle multiple concurrent agents', () => {
      const { startAgent, updateProgress, getAgentProgress } =
        useAgentProgressStore.getState();

      startAgent('agent-1', 'Agent 1', 10);
      startAgent('agent-2', 'Agent 2', 20);
      startAgent('agent-3', 'Agent 3', 5);

      useAgentProgressStore.getState().updateProgress('agent-1', { currentStep: 5 });
      useAgentProgressStore.getState().updateProgress('agent-2', { currentStep: 10 });
      useAgentProgressStore.getState().updateProgress('agent-3', { currentStep: 5 });

      const state = useAgentProgressStore.getState();
      expect(state.getAgentProgress('agent-1')?.percentComplete).toBe(50);
      expect(state.getAgentProgress('agent-2')?.percentComplete).toBe(50);
      expect(state.getAgentProgress('agent-3')?.percentComplete).toBe(100);
    });
  });
});
