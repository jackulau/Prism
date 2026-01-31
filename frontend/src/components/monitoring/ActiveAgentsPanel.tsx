import { useState } from 'react';
import { Bot, ChevronDown, ChevronUp, Loader2, CheckCircle2, XCircle, Circle } from 'lucide-react';
import { useActiveAgents, useActiveAgentCount } from '../../store/monitoringStore';
import type { AgentStatus } from '../../types/monitoring';

interface AgentStatusIconProps {
  status: AgentStatus;
}

function AgentStatusIcon({ status }: AgentStatusIconProps) {
  switch (status) {
    case 'running':
    case 'starting':
      return <Loader2 size={14} className="text-editor-accent animate-spin" />;
    case 'completed':
      return <CheckCircle2 size={14} className="text-editor-success" />;
    case 'failed':
      return <XCircle size={14} className="text-editor-error" />;
    case 'cancelled':
      return <Circle size={14} className="text-editor-muted" />;
    default:
      return <Circle size={14} className="text-editor-muted" />;
  }
}

function getStatusLabel(status: AgentStatus): string {
  const labels: Record<AgentStatus, string> = {
    starting: 'Starting',
    running: 'Running',
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',
  };
  return labels[status] || status;
}

function formatDuration(startedAt: number): string {
  const seconds = Math.floor((Date.now() - startedAt) / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

export function ActiveAgentsPanel() {
  const [isExpanded, setIsExpanded] = useState(false);
  const agentCount = useActiveAgentCount();
  const agents = useActiveAgents();

  const runningAgents = agents.filter((a) => a.status === 'running' || a.status === 'starting');
  const displayAgents = isExpanded ? agents : agents.slice(0, 3);

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <div className="p-1.5 bg-editor-success/10 rounded-lg">
            <Bot size={16} className="text-editor-success" />
          </div>
          <h3 className="font-medium text-editor-text">Active Agents</h3>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-2xl font-semibold text-editor-text">{agentCount}</span>
          {runningAgents.length > 0 && (
            <span className="px-2 py-0.5 text-xs font-medium bg-editor-accent/20 text-editor-accent rounded-full">
              {runningAgents.length} running
            </span>
          )}
        </div>
      </div>

      {agents.length === 0 ? (
        <p className="text-sm text-editor-muted text-center py-4">No active agents</p>
      ) : (
        <div className="space-y-2">
          {displayAgents.map((agent) => (
            <div
              key={agent.id}
              className="flex items-center justify-between p-2 bg-editor-bg rounded-lg"
            >
              <div className="flex items-center gap-2 min-w-0">
                <AgentStatusIcon status={agent.status} />
                <div className="min-w-0">
                  <p className="text-sm font-medium text-editor-text truncate">{agent.name}</p>
                  {agent.taskDescription && (
                    <p className="text-xs text-editor-muted truncate">{agent.taskDescription}</p>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2 shrink-0 ml-2">
                <span className="text-xs text-editor-muted">{getStatusLabel(agent.status)}</span>
                <span className="text-xs text-editor-muted">{formatDuration(agent.startedAt)}</span>
              </div>
            </div>
          ))}

          {agents.length > 3 && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="flex items-center justify-center gap-1 w-full py-2 text-sm text-editor-accent hover:text-editor-accent/80 transition-colors"
              aria-expanded={isExpanded}
              aria-label={isExpanded ? 'Show fewer agents' : `Show ${agents.length - 3} more agents`}
            >
              {isExpanded ? (
                <>
                  <ChevronUp size={14} />
                  Show less
                </>
              ) : (
                <>
                  <ChevronDown size={14} />
                  Show {agents.length - 3} more
                </>
              )}
            </button>
          )}
        </div>
      )}
    </div>
  );
}
