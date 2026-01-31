import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Clock,
  Coins,
  Layers,
  Users,
  Zap,
  Network,
} from 'lucide-react';
import { ExecutionResult, ExecutionStatus, ExecutionType } from './types';

interface ResultsListItemProps {
  result: ExecutionResult;
  isSelected?: boolean;
  onClick?: (result: ExecutionResult) => void;
}

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins} ${diffMins === 1 ? 'minute' : 'minutes'} ago`;
  if (diffHours < 24) return `${diffHours} ${diffHours === 1 ? 'hour' : 'hours'} ago`;
  if (diffDays < 7) return `${diffDays} ${diffDays === 1 ? 'day' : 'days'} ago`;
  return date.toLocaleDateString();
}

function formatDuration(ms?: number): string {
  if (!ms) return '--';
  if (ms < 1000) return `${ms}ms`;
  const secs = Math.floor(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const remainingSecs = secs % 60;
  if (mins < 60) return `${mins}m ${remainingSecs}s`;
  const hours = Math.floor(mins / 60);
  const remainingMins = mins % 60;
  return `${hours}h ${remainingMins}m`;
}

function formatTokens(count: number): string {
  if (count < 1000) return count.toString();
  if (count < 1000000) return `${(count / 1000).toFixed(1)}K`;
  return `${(count / 1000000).toFixed(2)}M`;
}

function formatCost(cost?: number): string {
  if (cost === undefined || cost === null) return '--';
  if (cost < 0.01) return '<$0.01';
  return `$${cost.toFixed(2)}`;
}

function getStatusConfig(status: ExecutionStatus): { color: string; bgColor: string; label: string } {
  switch (status) {
    case 'pending':
      return { color: 'text-editor-muted', bgColor: 'bg-editor-muted/10', label: 'Pending' };
    case 'running':
      return { color: 'text-editor-warning', bgColor: 'bg-editor-warning/10', label: 'Running' };
    case 'completed':
      return { color: 'text-editor-success', bgColor: 'bg-editor-success/10', label: 'Completed' };
    case 'partially_completed':
      return { color: 'text-yellow-500', bgColor: 'bg-yellow-500/10', label: 'Partial' };
    case 'failed':
      return { color: 'text-editor-error', bgColor: 'bg-editor-error/10', label: 'Failed' };
    case 'cancelled':
      return { color: 'text-editor-muted', bgColor: 'bg-editor-muted/10', label: 'Cancelled' };
  }
}

function getTypeIcon(type: ExecutionType) {
  switch (type) {
    case 'batch':
      return <Layers size={16} className="text-editor-accent" />;
    case 'swarm':
      return <Network size={16} className="text-purple-400" />;
  }
}

export function ResultsListItem({ result, isSelected, onClick }: ResultsListItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const statusConfig = getStatusConfig(result.status);

  const handleClick = () => {
    onClick?.(result);
  };

  const handleExpandClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsExpanded(!isExpanded);
  };

  return (
    <div
      className={`bg-editor-surface border rounded-lg transition-all ${
        isSelected
          ? 'border-editor-accent ring-1 ring-editor-accent/30'
          : 'border-editor-border hover:border-editor-accent/30'
      }`}
    >
      {/* Main Row */}
      <div
        onClick={handleClick}
        className="p-4 cursor-pointer"
      >
        <div className="flex items-start gap-4">
          {/* Expand/Collapse Toggle */}
          <button
            onClick={handleExpandClick}
            className="mt-1 p-1 text-editor-muted hover:text-editor-text hover:bg-editor-surface/80 rounded transition-colors"
          >
            {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
          </button>

          {/* Type Icon */}
          <div className="p-2 bg-editor-bg rounded-lg">
            {getTypeIcon(result.type)}
          </div>

          {/* Main Content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h3 className="font-medium text-editor-text truncate">
                {result.name || `Execution ${result.id.slice(0, 8)}`}
              </h3>
              <span
                className={`px-2 py-0.5 text-xs rounded-full ${statusConfig.bgColor} ${statusConfig.color}`}
              >
                {statusConfig.label}
              </span>
              <span className="text-xs text-editor-muted px-2 py-0.5 bg-editor-bg rounded capitalize">
                {result.type}
              </span>
            </div>

            <p className="text-sm text-editor-muted truncate">
              ID: {result.id}
            </p>
          </div>

          {/* Stats */}
          <div className="flex items-center gap-6 text-sm">
            <div className="flex items-center gap-1.5 text-editor-muted" title="Started">
              <Clock size={14} />
              <span>{formatRelativeTime(result.startedAt)}</span>
            </div>

            <div className="flex items-center gap-1.5 text-editor-muted" title="Duration">
              <Zap size={14} />
              <span>{formatDuration(result.durationMs)}</span>
            </div>

            <div className="flex items-center gap-1.5 text-editor-muted" title="Tokens">
              <span className="font-mono">{formatTokens(result.totalTokens)}</span>
              <span className="text-xs">tokens</span>
            </div>

            <div className="flex items-center gap-1.5 text-editor-muted" title="Cost">
              <Coins size={14} />
              <span>{formatCost(result.cost)}</span>
            </div>

            <div className="flex items-center gap-1.5 text-editor-muted" title="Agents / Tasks">
              <Users size={14} />
              <span>{result.agentCount} / {result.taskCount}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && (
        <div className="px-4 pb-4 pt-0 border-t border-editor-border mt-2">
          <div className="pt-4 space-y-3">
            {/* Timing Details */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-editor-muted">Started:</span>
                <span className="ml-2 text-editor-text">
                  {result.startedAt.toLocaleString()}
                </span>
              </div>
              {result.completedAt && (
                <div>
                  <span className="text-editor-muted">Completed:</span>
                  <span className="ml-2 text-editor-text">
                    {result.completedAt.toLocaleString()}
                  </span>
                </div>
              )}
            </div>

            {/* Error Message */}
            {result.error && (
              <div className="p-3 bg-editor-error/10 border border-editor-error/20 rounded-lg">
                <p className="text-xs text-editor-error font-medium mb-1">Error</p>
                <pre className="text-sm text-editor-error whitespace-pre-wrap overflow-x-auto">
                  {result.error}
                </pre>
              </div>
            )}

            {/* Agent Results */}
            {result.results && result.results.length > 0 && (
              <div className="space-y-2">
                <h4 className="text-sm font-medium text-editor-text">Agent Results</h4>
                <div className="space-y-2">
                  {result.results.slice(0, 3).map((agentResult) => (
                    <div
                      key={agentResult.agentId}
                      className="p-3 bg-editor-bg rounded-lg text-sm"
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium text-editor-text">
                          Agent {agentResult.agentId.slice(0, 8)}
                        </span>
                        <span
                          className={`text-xs ${
                            agentResult.success ? 'text-editor-success' : 'text-editor-error'
                          }`}
                        >
                          {agentResult.success ? 'Success' : 'Failed'}
                        </span>
                      </div>
                      {agentResult.output && (
                        <p className="text-editor-muted truncate">{agentResult.output}</p>
                      )}
                      {agentResult.error && (
                        <p className="text-editor-error truncate">{agentResult.error}</p>
                      )}
                    </div>
                  ))}
                  {result.results.length > 3 && (
                    <p className="text-xs text-editor-muted text-center">
                      +{result.results.length - 3} more results
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
