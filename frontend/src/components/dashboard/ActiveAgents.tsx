import { useMemo } from 'react';
import { Bot, Users, Zap } from 'lucide-react';
import { useAgentProgressStore } from '../../store/agentProgressStore';
import { AgentProgressCard, SwarmProgressCard } from '../agent';
import type { AgentProgress, SwarmProgress } from '../../types';

/**
 * ActiveAgents - Dashboard component showing all active agents and swarms
 *
 * Displays a grid of agent progress cards and swarm progress cards,
 * providing a real-time overview of all running AI agents.
 */
export function ActiveAgents() {
  const activeAgents = useAgentProgressStore((state) => state.activeAgents);
  const activeSwarms = useAgentProgressStore((state) => state.activeSwarms);
  const clearAgent = useAgentProgressStore((state) => state.clearAgent);
  const clearSwarm = useAgentProgressStore((state) => state.clearSwarm);
  const historicalMetrics = useAgentProgressStore((state) => state.historicalMetrics);

  // Convert Maps to arrays for rendering
  const agentsList = useMemo(
    () => Array.from(activeAgents.values()),
    [activeAgents]
  );
  const swarmsList = useMemo(
    () => Array.from(activeSwarms.values()),
    [activeSwarms]
  );

  // Filter for active agents (exclude ones in swarms to avoid duplication)
  const swarmAgentIds = useMemo(() => {
    const ids = new Set<string>();
    swarmsList.forEach((swarm) => {
      swarm.agents.forEach((_, agentId) => ids.add(agentId));
    });
    return ids;
  }, [swarmsList]);

  const standaloneAgents = useMemo(
    () => agentsList.filter((agent) => !swarmAgentIds.has(agent.agentId)),
    [agentsList, swarmAgentIds]
  );

  // Calculate summary stats
  const stats = useMemo(() => {
    const running = agentsList.filter(
      (a) => a.status === 'running' || a.status === 'thinking'
    ).length;
    const completed = agentsList.filter((a) => a.status === 'completed').length;
    const failed = agentsList.filter(
      (a) => a.status === 'failed' || a.status === 'cancelled'
    ).length;
    const totalTokens = historicalMetrics.reduce(
      (sum, m) => sum + m.tokensGenerated,
      0
    );
    return { running, completed, failed, totalTokens };
  }, [agentsList, historicalMetrics]);

  // If no agents at all, show empty state
  if (agentsList.length === 0 && swarmsList.length === 0) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Active Agents</h2>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-editor-accent/10 flex items-center justify-center">
            <Bot size={24} className="text-editor-accent" />
          </div>
          <h3 className="text-editor-text font-medium mb-2">No Active Agents</h3>
          <p className="text-sm text-editor-muted max-w-sm mx-auto">
            When you run AI agents or workflows, their progress will appear here in real-time.
          </p>
        </div>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      {/* Header with stats */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Active Agents</h2>
        <div className="flex items-center gap-4 text-sm">
          {stats.running > 0 && (
            <span className="flex items-center gap-1.5 text-blue-400">
              <span className="w-2 h-2 rounded-full bg-blue-400 animate-pulse" />
              {stats.running} running
            </span>
          )}
          {stats.completed > 0 && (
            <span className="flex items-center gap-1.5 text-editor-success">
              <span className="w-2 h-2 rounded-full bg-editor-success" />
              {stats.completed} completed
            </span>
          )}
          {stats.totalTokens > 0 && (
            <span className="flex items-center gap-1.5 text-editor-muted">
              <Zap size={14} />
              {stats.totalTokens >= 1000
                ? `${(stats.totalTokens / 1000).toFixed(1)}k`
                : stats.totalTokens}{' '}
              tokens
            </span>
          )}
        </div>
      </div>

      {/* Swarms */}
      {swarmsList.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm text-editor-muted">
            <Users size={14} />
            <span>Swarms ({swarmsList.length})</span>
          </div>
          <div className="space-y-4">
            {swarmsList.map((swarm: SwarmProgress) => (
              <SwarmProgressCard
                key={swarm.swarmId}
                {...swarm}
                showAgentCards
                defaultExpanded={swarmsList.length === 1}
                onCancelSwarm={() => {
                  // Cancel all agents in swarm
                  swarm.agents.forEach((_, agentId) => clearAgent(agentId));
                  clearSwarm(swarm.swarmId);
                }}
                onCancelAgent={(agentId) => clearAgent(agentId)}
              />
            ))}
          </div>
        </div>
      )}

      {/* Standalone Agents */}
      {standaloneAgents.length > 0 && (
        <div className="space-y-3">
          {swarmsList.length > 0 && (
            <div className="flex items-center gap-2 text-sm text-editor-muted">
              <Bot size={14} />
              <span>Standalone Agents ({standaloneAgents.length})</span>
            </div>
          )}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {standaloneAgents.map((agent: AgentProgress) => (
              <AgentProgressCard
                key={agent.agentId}
                agentId={agent.agentId}
                name={agent.name}
                status={agent.status}
                percentComplete={agent.percentComplete}
                currentStep={agent.currentStep}
                totalSteps={agent.totalSteps}
                stepName={agent.stepName}
                message={agent.message}
                isThinking={agent.isThinking}
                estimatedTimeRemaining={agent.estimatedTimeRemaining ?? undefined}
                tokensGenerated={agent.tokensGenerated}
                elapsedTime={agent.startedAt ? Date.now() - agent.startedAt : undefined}
                showMetrics
                onCancel={
                  agent.status === 'running' || agent.status === 'thinking'
                    ? () => clearAgent(agent.agentId)
                    : undefined
                }
              />
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
