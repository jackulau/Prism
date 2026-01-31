import React, { useState } from 'react';
import {
  ChevronDown,
  ChevronUp,
  Bot,
  AlertCircle,
  CheckCircle,
  Loader2,
  Clock,
  Cpu,
} from 'lucide-react';

export type AgentResultStatus = 'pending' | 'running' | 'success' | 'failed';

export interface AgentResult {
  id: string;
  agentId: string;
  role?: string;
  status: AgentResultStatus;
  output?: string;
  error?: string;
  promptTokens?: number;
  completionTokens?: number;
  duration?: number; // milliseconds
  startedAt?: Date;
  completedAt?: Date;
}

interface AgentResultCardProps {
  result: AgentResult;
  defaultExpanded?: boolean;
}

const statusConfig: Record<AgentResultStatus, {
  icon: React.ReactNode;
  bg: string;
  text: string;
  label: string;
}> = {
  pending: {
    icon: <Clock className="w-4 h-4" />,
    bg: 'bg-editor-muted/10',
    text: 'text-editor-muted',
    label: 'Pending',
  },
  running: {
    icon: <Loader2 className="w-4 h-4 animate-spin" />,
    bg: 'bg-editor-warning/10',
    text: 'text-editor-warning',
    label: 'Running',
  },
  success: {
    icon: <CheckCircle className="w-4 h-4" />,
    bg: 'bg-editor-success/10',
    text: 'text-editor-success',
    label: 'Success',
  },
  failed: {
    icon: <AlertCircle className="w-4 h-4" />,
    bg: 'bg-editor-error/10',
    text: 'text-editor-error',
    label: 'Failed',
  },
};

function formatDuration(ms?: number): string {
  if (ms === undefined) return '--';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatTokens(count?: number): string {
  if (count === undefined) return '--';
  return count.toLocaleString();
}

export const AgentResultCard: React.FC<AgentResultCardProps> = ({
  result,
  defaultExpanded = false,
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const config = statusConfig[result.status];
  const totalTokens = (result.promptTokens || 0) + (result.completionTokens || 0);

  const hasOutput = result.output && result.output.trim().length > 0;
  const hasError = result.error && result.error.trim().length > 0;
  const isExpandable = hasOutput || hasError;

  const outputPreview = result.output
    ? result.output.slice(0, 150) + (result.output.length > 150 ? '...' : '')
    : null;

  return (
    <div
      className={`bg-editor-surface border rounded-lg transition-all ${
        result.status === 'failed'
          ? 'border-editor-error/30'
          : 'border-editor-border hover:border-editor-accent/30'
      }`}
    >
      {/* Header - always visible */}
      <div
        className={`p-4 ${isExpandable ? 'cursor-pointer' : ''}`}
        onClick={() => isExpandable && setIsExpanded(!isExpanded)}
      >
        <div className="flex items-start justify-between gap-4">
          {/* Left side: Agent info */}
          <div className="flex items-start gap-3 min-w-0">
            {/* Agent icon */}
            <div className={`p-2 rounded-lg ${config.bg}`}>
              <Bot className={`w-5 h-5 ${config.text}`} />
            </div>

            {/* Agent details */}
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <h4 className="font-medium text-editor-text truncate">
                  {result.agentId}
                </h4>
                {result.role && (
                  <span className="text-xs px-2 py-0.5 bg-editor-accent/10 text-editor-accent rounded-full">
                    {result.role}
                  </span>
                )}
              </div>

              {/* Status and metrics row */}
              <div className="flex items-center gap-4 mt-2 text-xs text-editor-muted flex-wrap">
                {/* Status badge */}
                <span className={`flex items-center gap-1 ${config.text}`}>
                  {config.icon}
                  {config.label}
                </span>

                {/* Duration */}
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {formatDuration(result.duration)}
                </span>

                {/* Tokens */}
                {totalTokens > 0 && (
                  <span className="flex items-center gap-1">
                    <Cpu className="w-3 h-3" />
                    {formatTokens(totalTokens)} tokens
                  </span>
                )}
              </div>

              {/* Output preview (collapsed state) */}
              {!isExpanded && outputPreview && (
                <p className="mt-2 text-sm text-editor-muted line-clamp-2 font-mono">
                  {outputPreview}
                </p>
              )}
            </div>
          </div>

          {/* Right side: Expand toggle */}
          {isExpandable && (
            <button
              className="p-1 text-editor-muted hover:text-editor-text transition-colors"
              onClick={(e) => {
                e.stopPropagation();
                setIsExpanded(!isExpanded);
              }}
            >
              {isExpanded ? (
                <ChevronUp className="w-5 h-5" />
              ) : (
                <ChevronDown className="w-5 h-5" />
              )}
            </button>
          )}
        </div>
      </div>

      {/* Expanded content */}
      {isExpanded && (
        <div className="border-t border-editor-border">
          {/* Token breakdown */}
          {(result.promptTokens !== undefined || result.completionTokens !== undefined) && (
            <div className="px-4 py-3 border-b border-editor-border bg-editor-bg/50">
              <div className="flex items-center gap-6 text-xs">
                <div>
                  <span className="text-editor-muted">Prompt: </span>
                  <span className="text-editor-text font-medium">
                    {formatTokens(result.promptTokens)}
                  </span>
                </div>
                <div>
                  <span className="text-editor-muted">Completion: </span>
                  <span className="text-editor-text font-medium">
                    {formatTokens(result.completionTokens)}
                  </span>
                </div>
                <div>
                  <span className="text-editor-muted">Total: </span>
                  <span className="text-editor-text font-medium">
                    {formatTokens(totalTokens)}
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Output */}
          {hasOutput && (
            <div className="p-4">
              <h5 className="text-xs text-editor-muted uppercase tracking-wide mb-2">
                Output
              </h5>
              <pre className="text-sm text-editor-text bg-editor-bg rounded-lg p-3 overflow-x-auto whitespace-pre-wrap font-mono max-h-96 overflow-y-auto">
                {result.output}
              </pre>
            </div>
          )}

          {/* Error */}
          {hasError && (
            <div className="p-4">
              <h5 className="text-xs text-editor-error uppercase tracking-wide mb-2 flex items-center gap-1">
                <AlertCircle className="w-3 h-3" />
                Error
              </h5>
              <pre className="text-sm text-editor-error bg-editor-error/10 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap font-mono border border-editor-error/20">
                {result.error}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
