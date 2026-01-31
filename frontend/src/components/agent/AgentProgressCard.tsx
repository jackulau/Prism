import React, { memo } from 'react';
import { Bot, X, RotateCcw } from 'lucide-react';
import { ProgressBar } from '../ui/ProgressBar';
import { ThinkingIndicator } from '../ui/ThinkingIndicator';
import { StatusBadge, AgentStatus } from './StatusBadge';

export interface AgentProgressCardProps {
  agentId: string;
  name: string;
  status: AgentStatus;

  // Progress
  percentComplete: number;
  currentStep?: number;
  totalSteps?: number;
  stepName?: string;
  message?: string;

  // Thinking
  isThinking?: boolean;
  thinkingDuration?: number;

  // Estimates
  estimatedTimeRemaining?: number;
  estimatedTokensRemaining?: number;

  // Metrics
  tokensGenerated?: number;
  elapsedTime?: number;

  // Actions
  onCancel?: () => void;
  onRetry?: () => void;

  // Display options
  compact?: boolean;
  showMetrics?: boolean;
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  }

  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) {
    return `${seconds}s`;
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  if (minutes < 60) {
    return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
  }

  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
}

function formatTokens(count: number): string {
  if (count < 1000) {
    return count.toString();
  }
  if (count < 1000000) {
    return `${(count / 1000).toFixed(1)}k`;
  }
  return `${(count / 1000000).toFixed(1)}M`;
}

const AgentIcon: React.FC<{ status: AgentStatus }> = ({ status }) => {
  const colorClass =
    status === 'completed'
      ? 'text-editor-success'
      : status === 'failed'
      ? 'text-editor-error'
      : status === 'thinking'
      ? 'text-purple-400'
      : status === 'running'
      ? 'text-blue-400'
      : 'text-editor-muted';

  return (
    <div
      className={`p-1.5 rounded-lg bg-editor-surface border border-editor-border ${colorClass}`}
    >
      <Bot size={16} />
    </div>
  );
};

export const AgentProgressCard: React.FC<AgentProgressCardProps> = memo(
  ({
    name,
    status,
    percentComplete,
    currentStep,
    totalSteps,
    stepName,
    message,
    isThinking = false,
    estimatedTimeRemaining,
    tokensGenerated,
    elapsedTime,
    onCancel,
    onRetry,
    compact = false,
    showMetrics = true,
  }) => {
    const isActive = status === 'running' || status === 'thinking';
    const showThinking = isThinking || status === 'thinking';
    const canCancel = isActive && onCancel;
    const canRetry = (status === 'failed' || status === 'cancelled') && onRetry;

    // Compact variant for inline display
    if (compact) {
      return (
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-editor-surface border border-editor-border">
          <span className="text-sm text-editor-text font-medium truncate max-w-32">
            {name}
          </span>
          {showThinking ? (
            <ThinkingIndicator active variant="dots" size="sm" color="accent" />
          ) : (
            <span className="text-xs text-editor-muted">
              {Math.round(percentComplete)}%
            </span>
          )}
          <StatusBadge status={status} size="sm" showIcon={false} />
        </div>
      );
    }

    return (
      <div className="rounded-lg border border-editor-border bg-editor-surface/50 overflow-hidden">
        {/* Header: Agent name + status badge */}
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3 min-w-0">
            <AgentIcon status={status} />
            <span className="font-medium text-editor-text truncate">{name}</span>
          </div>
          <StatusBadge status={status} />
        </div>

        {/* Progress section */}
        <div className="px-4 pb-3">
          {showThinking ? (
            <div className="flex items-center gap-2 py-2">
              <ThinkingIndicator
                active
                variant="wave"
                size="md"
                color="accent"
                label="Processing..."
                showLabel
              />
            </div>
          ) : (
            <ProgressBar
              value={percentComplete}
              showPercentage
              animated
              variant={
                status === 'completed'
                  ? 'success'
                  : status === 'failed'
                  ? 'error'
                  : 'default'
              }
              steps={totalSteps}
              currentStep={currentStep}
              size="md"
            />
          )}
        </div>

        {/* Step info */}
        {stepName && currentStep && totalSteps && (
          <div className="px-4 pb-2 text-sm text-editor-muted">
            Step {currentStep}/{totalSteps}: {stepName}
          </div>
        )}

        {/* Message */}
        {message && (
          <div className="px-4 pb-3 text-sm text-editor-muted truncate">
            {message}
          </div>
        )}

        {/* Metrics row */}
        {showMetrics && (elapsedTime || estimatedTimeRemaining || tokensGenerated) && (
          <div className="px-4 pb-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-editor-muted">
            {elapsedTime !== undefined && elapsedTime > 0 && (
              <span className="flex items-center gap-1">
                <span className="opacity-60">Elapsed:</span>
                <span>{formatDuration(elapsedTime)}</span>
              </span>
            )}
            {estimatedTimeRemaining !== undefined && estimatedTimeRemaining > 0 && (
              <span className="flex items-center gap-1">
                <span className="opacity-60">~</span>
                <span>{formatDuration(estimatedTimeRemaining)} remaining</span>
              </span>
            )}
            {tokensGenerated !== undefined && tokensGenerated > 0 && (
              <span className="flex items-center gap-1">
                <span className="opacity-60">Tokens:</span>
                <span>{formatTokens(tokensGenerated)}</span>
              </span>
            )}
          </div>
        )}

        {/* Actions */}
        {(canCancel || canRetry) && (
          <div className="flex gap-2 px-4 py-3 bg-editor-bg/30 border-t border-editor-border">
            {canCancel && (
              <button
                onClick={onCancel}
                className="flex items-center justify-center gap-2 flex-1 py-2 px-4 bg-editor-error/10 text-editor-error rounded-lg hover:bg-editor-error/20 transition-colors text-sm font-medium"
              >
                <X size={14} />
                Cancel
              </button>
            )}
            {canRetry && (
              <button
                onClick={onRetry}
                className="flex items-center justify-center gap-2 flex-1 py-2 px-4 bg-editor-accent/10 text-editor-accent rounded-lg hover:bg-editor-accent/20 transition-colors text-sm font-medium"
              >
                <RotateCcw size={14} />
                Retry
              </button>
            )}
          </div>
        )}
      </div>
    );
  }
);

AgentProgressCard.displayName = 'AgentProgressCard';

export default AgentProgressCard;
