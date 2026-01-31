import { memo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Bot,
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  ArrowRight,
  Zap,
} from 'lucide-react';
import type { AgentExecution } from './types';

interface AgentCardProps {
  agent: AgentExecution;
  onClick?: (agent: AgentExecution) => void;
}

const statusConfig = {
  pending: {
    color: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: Clock,
    spin: false,
    label: 'Pending',
  },
  running: {
    color: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: Loader2,
    spin: true,
    label: 'Running',
  },
  completed: {
    color: 'bg-green-500/20 text-green-400 border-green-500/30',
    icon: CheckCircle,
    spin: false,
    label: 'Completed',
  },
  failed: {
    color: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: XCircle,
    spin: false,
    label: 'Failed',
  },
  cancelled: {
    color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
    icon: XCircle,
    spin: false,
    label: 'Cancelled',
  },
};

function formatRelativeTime(date: Date | string): string {
  const now = new Date();
  const then = new Date(date);
  const diffMs = now.getTime() - then.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return then.toLocaleDateString();
}

function formatDuration(startedAt: Date | string, completedAt?: Date | string | null): string {
  if (!completedAt) return '';
  const start = new Date(startedAt).getTime();
  const end = new Date(completedAt).getTime();
  const durationMs = end - start;
  const seconds = Math.floor(durationMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

function formatTokens(count?: number): string {
  if (!count) return '0';
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
  return count.toString();
}

function formatCost(cost?: number): string {
  if (!cost) return '$0.00';
  if (cost < 0.01) return `$${cost.toFixed(4)}`;
  return `$${cost.toFixed(2)}`;
}

export const AgentCard = memo(function AgentCard({ agent, onClick }: AgentCardProps) {
  const navigate = useNavigate();
  const config = statusConfig[agent.status] || statusConfig.pending;
  const StatusIcon = config.icon;

  const handleClick = () => {
    if (onClick) {
      onClick(agent);
    } else {
      navigate(`/agents/${agent.id}`);
    }
  };

  const totalTokens = (agent.prompt_tokens || 0) + (agent.completion_tokens || 0);

  return (
    <button
      onClick={handleClick}
      className="group w-full bg-editor-surface border border-editor-border rounded-lg p-4 text-left hover:border-editor-accent/50 hover:bg-editor-surface/80 transition-all"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3 min-w-0 flex-1">
          <div className="p-2 bg-editor-accent/10 rounded-lg flex-shrink-0">
            <Bot className="w-5 h-5 text-editor-accent" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h3
                className="font-medium text-editor-text truncate"
                title={agent.name || agent.task || 'Agent Execution'}
              >
                {agent.name || agent.task || 'Agent Execution'}
              </h3>
            </div>
            <div className="flex items-center gap-2 flex-wrap">
              <span
                className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border ${config.color}`}
              >
                <StatusIcon size={12} className={config.spin ? 'animate-spin' : ''} />
                {config.label}
              </span>
              {agent.model && (
                <span className="px-2 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded border border-purple-500/30">
                  {agent.model}
                </span>
              )}
            </div>
          </div>
        </div>
        <ArrowRight
          size={18}
          className="text-editor-muted opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
        />
      </div>

      <div className="mt-3 pt-3 border-t border-editor-border/50">
        <div className="flex items-center justify-between text-xs text-editor-muted">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1" title="Tokens used">
              <Zap size={12} />
              <span>{formatTokens(totalTokens)}</span>
            </div>
            {agent.cost !== undefined && agent.cost > 0 && (
              <span title="Cost">{formatCost(agent.cost)}</span>
            )}
            {agent.status === 'completed' && agent.started_at && agent.completed_at && (
              <span title="Duration">
                {formatDuration(agent.started_at, agent.completed_at)}
              </span>
            )}
          </div>
          <div className="flex items-center gap-1">
            <Clock size={12} />
            <span>{formatRelativeTime(agent.created_at)}</span>
          </div>
        </div>
      </div>
    </button>
  );
});

export default AgentCard;
