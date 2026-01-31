import React, { useState, memo, useMemo } from 'react';
import { ChevronDown, ChevronUp, Bot } from 'lucide-react';
import { AgentProgressCard } from './AgentProgressCard';
import type { AgentProgress } from '../../types';

export interface AgentProgressPanelProps {
  /** List of agents to display */
  agents: AgentProgress[];
  /** Position of the panel */
  position?: 'bottom' | 'sidebar' | 'inline';
  /** Whether the panel can be collapsed */
  collapsible?: boolean;
  /** Maximum number of agents to show before "show more" */
  maxVisible?: number;
  /** Handler for canceling an agent */
  onCancelAgent?: (agentId: string) => void;
  /** Handler for retrying a failed agent */
  onRetryAgent?: (agentId: string) => void;
  /** Custom class name */
  className?: string;
}

const positionStyles = {
  bottom: 'fixed bottom-4 left-1/2 -translate-x-1/2 w-full max-w-2xl px-4',
  sidebar: 'w-full',
  inline: 'w-full',
};

/**
 * AgentProgressPanel - Displays a collection of agent progress cards
 *
 * This component renders a panel showing progress for multiple agents,
 * with collapsible behavior and configurable display options.
 *
 * @example
 * ```tsx
 * const activeAgents = useAgentProgressStore(state =>
 *   Array.from(state.activeAgents.values()).filter(a => a.status === 'running')
 * );
 *
 * <AgentProgressPanel
 *   agents={activeAgents}
 *   position="bottom"
 *   collapsible
 *   onCancelAgent={(id) => cancelAgent(id)}
 * />
 * ```
 */
export const AgentProgressPanel: React.FC<AgentProgressPanelProps> = memo(
  ({
    agents,
    position = 'bottom',
    collapsible = true,
    maxVisible = 3,
    onCancelAgent,
    onRetryAgent,
    className = '',
  }) => {
    const [isCollapsed, setIsCollapsed] = useState(false);
    const [showAll, setShowAll] = useState(false);

    // Filter to only running/thinking agents for active count
    const activeAgents = useMemo(
      () => agents.filter((a) => a.status === 'running' || a.status === 'thinking'),
      [agents]
    );

    // Determine which agents to display
    const displayedAgents = useMemo(() => {
      if (showAll) return agents;
      return agents.slice(0, maxVisible);
    }, [agents, showAll, maxVisible]);

    const hiddenCount = agents.length - displayedAgents.length;

    // Don't render if no agents
    if (agents.length === 0) {
      return null;
    }

    const panelContent = (
      <div
        className={`
          rounded-xl border border-editor-border bg-editor-surface/95 backdrop-blur-sm
          shadow-lg shadow-black/20 overflow-hidden
          transition-all duration-300 ease-in-out
          ${positionStyles[position]}
          ${className}
        `}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 bg-editor-bg/50 border-b border-editor-border">
          <div className="flex items-center gap-3">
            <div className="p-1.5 rounded-lg bg-editor-accent/20 text-editor-accent">
              <Bot size={16} />
            </div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-editor-text">
                {activeAgents.length > 0
                  ? `${activeAgents.length} agent${activeAgents.length !== 1 ? 's' : ''} running`
                  : `${agents.length} agent${agents.length !== 1 ? 's' : ''}`}
              </span>
              {activeAgents.length > 0 && (
                <span className="flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-editor-accent animate-pulse" />
                </span>
              )}
            </div>
          </div>

          {collapsible && (
            <button
              onClick={() => setIsCollapsed(!isCollapsed)}
              className="p-1.5 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors"
              aria-label={isCollapsed ? 'Expand panel' : 'Collapse panel'}
            >
              {isCollapsed ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
            </button>
          )}
        </div>

        {/* Content */}
        <div
          className={`
            transition-all duration-300 ease-in-out overflow-hidden
            ${isCollapsed ? 'max-h-0 opacity-0' : 'max-h-[500px] opacity-100'}
          `}
        >
          <div className="p-3 space-y-2 max-h-[400px] overflow-y-auto">
            {displayedAgents.map((agent) => (
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
                compact={position === 'inline'}
                showMetrics={position !== 'inline'}
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

            {/* Show more/less toggle */}
            {hiddenCount > 0 && !showAll && (
              <button
                onClick={() => setShowAll(true)}
                className="w-full py-2 text-sm text-editor-muted hover:text-editor-accent transition-colors"
              >
                + {hiddenCount} more agent{hiddenCount !== 1 ? 's' : ''}
              </button>
            )}
            {showAll && agents.length > maxVisible && (
              <button
                onClick={() => setShowAll(false)}
                className="w-full py-2 text-sm text-editor-muted hover:text-editor-accent transition-colors"
              >
                Show less
              </button>
            )}
          </div>
        </div>
      </div>
    );

    // For bottom position, we need a portal-like wrapper
    if (position === 'bottom') {
      return <div className="fixed inset-x-0 bottom-0 z-50 pointer-events-none">
        <div className="pointer-events-auto">{panelContent}</div>
      </div>;
    }

    return panelContent;
  }
);

AgentProgressPanel.displayName = 'AgentProgressPanel';

export default AgentProgressPanel;
