import React, { memo, useMemo } from 'react';
import { Users, CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { ProgressBar } from '../ui/ProgressBar';
import { AgentProgressCard } from './AgentProgressCard';
import type { SwarmProgress, AgentProgress } from '../../types';

export interface SwarmProgressCardProps extends SwarmProgress {
  /** Whether to show individual agent cards */
  showAgentCards?: boolean;
  /** Whether to start collapsed */
  defaultExpanded?: boolean;
  /** Handler for canceling the entire swarm */
  onCancelSwarm?: () => void;
  /** Handler for canceling a specific agent */
  onCancelAgent?: (agentId: string) => void;
  /** Handler for retrying a failed agent */
  onRetryAgent?: (agentId: string) => void;
  /** Custom class name */
  className?: string;
}

const statusColors = {
  running: 'text-blue-400',
  synthesizing: 'text-purple-400',
  completed: 'text-editor-success',
  failed: 'text-editor-error',
};

const statusLabels = {
  running: 'Running',
  synthesizing: 'Synthesizing',
  completed: 'Completed',
  failed: 'Failed',
};

/**
 * SwarmProgressCard - Displays progress for a multi-agent swarm
 *
 * Shows overall swarm progress with a breakdown of individual agents.
 * Useful for visualizing parallel agent execution.
 *
 * @example
 * ```tsx
 * const swarm = useAgentProgressStore(state => state.activeSwarms.get('swarm-1'));
 *
 * {swarm && (
 *   <SwarmProgressCard
 *     {...swarm}
 *     showAgentCards
 *     onCancelSwarm={() => cancelSwarm('swarm-1')}
 *   />
 * )}
 * ```
 */
export const SwarmProgressCard: React.FC<SwarmProgressCardProps> = memo(
  ({
    swarmId,
    agents,
    overallPercent,
    completedAgents,
    totalAgents,
    status,
    showAgentCards = true,
    defaultExpanded = true,
    onCancelSwarm,
    onCancelAgent,
    onRetryAgent,
    className = '',
  }) => {
    const [isExpanded, setIsExpanded] = React.useState(defaultExpanded);

    // Convert Map to array for rendering
    const agentList = useMemo(() => Array.from(agents.values()), [agents]);

    // Count agents by status
    const statusCounts = useMemo(() => {
      const counts = { running: 0, thinking: 0, completed: 0, failed: 0, pending: 0 };
      agentList.forEach((agent) => {
        if (agent.status in counts) {
          counts[agent.status as keyof typeof counts]++;
        }
      });
      return counts;
    }, [agentList]);

    const isActive = status === 'running' || status === 'synthesizing';

    return (
      <div
        className={`
          rounded-xl border border-editor-border bg-editor-surface/50 overflow-hidden
          ${className}
        `}
      >
        {/* Header */}
        <div className="px-4 py-3 bg-editor-bg/30">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div
                className={`p-2 rounded-lg border ${
                  isActive
                    ? 'bg-purple-500/20 text-purple-400 border-purple-500/30'
                    : status === 'completed'
                    ? 'bg-editor-success/20 text-editor-success border-editor-success/30'
                    : 'bg-editor-error/20 text-editor-error border-editor-error/30'
                }`}
              >
                <Users size={18} />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium text-editor-text">Swarm</span>
                  <span className="text-xs text-editor-muted font-mono">{swarmId.slice(0, 8)}</span>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className={statusColors[status]}>{statusLabels[status]}</span>
                  {isActive && (
                    <Loader2 size={12} className="animate-spin text-purple-400" />
                  )}
                </div>
              </div>
            </div>

            {/* Status counts */}
            <div className="flex items-center gap-3 text-sm">
              {statusCounts.completed > 0 && (
                <span className="flex items-center gap-1 text-editor-success">
                  <CheckCircle2 size={14} />
                  {statusCounts.completed}
                </span>
              )}
              {statusCounts.failed > 0 && (
                <span className="flex items-center gap-1 text-editor-error">
                  <XCircle size={14} />
                  {statusCounts.failed}
                </span>
              )}
              {(statusCounts.running > 0 || statusCounts.thinking > 0) && (
                <span className="flex items-center gap-1 text-blue-400">
                  <Loader2 size={14} className="animate-spin" />
                  {statusCounts.running + statusCounts.thinking}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Progress bar */}
        <div className="px-4 py-3">
          <ProgressBar
            value={overallPercent}
            showPercentage
            animated={isActive}
            variant={
              status === 'completed'
                ? 'success'
                : status === 'failed'
                ? 'error'
                : 'default'
            }
            size="md"
          />
          <div className="flex items-center justify-between mt-2 text-sm text-editor-muted">
            <span>
              {completedAgents}/{totalAgents} agents completed
            </span>
            <span>{Math.round(overallPercent)}% overall</span>
          </div>
        </div>

        {/* Agent cards (expandable) */}
        {showAgentCards && agentList.length > 0 && (
          <>
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="w-full px-4 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface/50 transition-colors border-t border-editor-border flex items-center justify-between"
            >
              <span>
                {isExpanded ? 'Hide' : 'Show'} agent details ({agentList.length})
              </span>
              <span>{isExpanded ? '▼' : '▶'}</span>
            </button>

            {isExpanded && (
              <div className="px-4 pb-4 pt-2 grid grid-cols-1 md:grid-cols-2 gap-2 border-t border-editor-border/50">
                {agentList.map((agent: AgentProgress) => (
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
                    compact
                    showMetrics={false}
                    onCancel={
                      onCancelAgent &&
                      (agent.status === 'running' || agent.status === 'thinking')
                        ? () => onCancelAgent(agent.agentId)
                        : undefined
                    }
                    onRetry={
                      onRetryAgent &&
                      (agent.status === 'failed' || agent.status === 'cancelled')
                        ? () => onRetryAgent(agent.agentId)
                        : undefined
                    }
                  />
                ))}
              </div>
            )}
          </>
        )}

        {/* Cancel swarm button */}
        {isActive && onCancelSwarm && (
          <div className="px-4 py-3 bg-editor-bg/30 border-t border-editor-border">
            <button
              onClick={onCancelSwarm}
              className="w-full py-2 px-4 bg-editor-error/10 text-editor-error rounded-lg hover:bg-editor-error/20 transition-colors text-sm font-medium"
            >
              Cancel All Agents
            </button>
          </div>
        )}
      </div>
    );
  }
);

SwarmProgressCard.displayName = 'SwarmProgressCard';

export default SwarmProgressCard;
