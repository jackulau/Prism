import React, { useState, useCallback } from 'react';
import {
  ChevronDown,
  ChevronUp,
  Copy,
  Check,
  RefreshCw,
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  XOctagon,
} from 'lucide-react';
import type { BatchTask, TaskStatus } from './BatchProgressBar';

interface TaskResultCardProps {
  task: BatchTask;
  onRetry?: (taskId: string) => void;
  isRetrying?: boolean;
  defaultExpanded?: boolean;
}

const statusConfig: Record<
  TaskStatus,
  { icon: React.ReactNode; color: string; bgColor: string; label: string }
> = {
  pending: {
    icon: <Clock className="w-3.5 h-3.5" />,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
    label: 'Pending',
  },
  running: {
    icon: <Loader2 className="w-3.5 h-3.5 animate-spin" />,
    color: 'text-editor-warning',
    bgColor: 'bg-editor-warning/10',
    label: 'Running',
  },
  completed: {
    icon: <CheckCircle2 className="w-3.5 h-3.5" />,
    color: 'text-editor-success',
    bgColor: 'bg-editor-success/10',
    label: 'Completed',
  },
  failed: {
    icon: <XCircle className="w-3.5 h-3.5" />,
    color: 'text-editor-error',
    bgColor: 'bg-editor-error/10',
    label: 'Failed',
  },
  cancelled: {
    icon: <XOctagon className="w-3.5 h-3.5" />,
    color: 'text-editor-muted',
    bgColor: 'bg-editor-muted/10',
    label: 'Cancelled',
  },
};

export const TaskResultCard: React.FC<TaskResultCardProps> = ({
  task,
  onRetry,
  isRetrying = false,
  defaultExpanded = false,
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [copied, setCopied] = useState(false);

  const config = statusConfig[task.status];
  const hasContent = task.output || task.error;
  const canExpand = hasContent || task.status === 'running';

  const formatDuration = (start?: Date, end?: Date): string => {
    if (!start) return '--';
    const endTime = end || new Date();
    const ms = endTime.getTime() - start.getTime();
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const mins = Math.floor(ms / 60000);
    const secs = Math.round((ms % 60000) / 1000);
    return `${mins}m ${secs}s`;
  };

  const handleCopy = useCallback(async () => {
    const content = task.output || task.error || '';
    if (!content) return;

    try {
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }, [task.output, task.error]);

  const handleRetry = useCallback(() => {
    if (onRetry && !isRetrying) {
      onRetry(task.id);
    }
  }, [onRetry, isRetrying, task.id]);

  return (
    <div
      className={`border rounded-lg transition-all ${
        task.status === 'failed'
          ? 'border-editor-error/30 bg-editor-error/5'
          : task.status === 'running'
          ? 'border-editor-warning/30 bg-editor-warning/5'
          : 'border-editor-border bg-editor-surface'
      }`}
    >
      {/* Header */}
      <div
        className={`flex items-center justify-between p-3 ${
          canExpand ? 'cursor-pointer hover:bg-editor-bg/50' : ''
        }`}
        onClick={() => canExpand && setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          {/* Status indicator */}
          <div className={`p-1.5 rounded ${config.bgColor}`}>
            <span className={config.color}>{config.icon}</span>
          </div>

          {/* Task info */}
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-editor-text truncate">
                {task.id}
              </span>
              <span
                className={`px-2 py-0.5 text-xs rounded-full ${config.bgColor} ${config.color}`}
              >
                {config.label}
              </span>
            </div>
            <div className="text-xs text-editor-muted mt-0.5">
              Duration: {formatDuration(task.startedAt, task.completedAt)}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {(task.status === 'failed' || task.status === 'cancelled') && onRetry && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleRetry();
              }}
              disabled={isRetrying}
              className="p-1.5 text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 rounded transition-colors disabled:opacity-50"
              title="Retry task"
            >
              {isRetrying ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4" />
              )}
            </button>
          )}
          {canExpand && (
            <div className="text-editor-muted">
              {isExpanded ? (
                <ChevronUp className="w-4 h-4" />
              ) : (
                <ChevronDown className="w-4 h-4" />
              )}
            </div>
          )}
        </div>
      </div>

      {/* Expanded content */}
      {isExpanded && canExpand && (
        <div className="border-t border-editor-border">
          {/* Output/Error content */}
          {hasContent && (
            <div className="p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-editor-muted uppercase">
                  {task.error ? 'Error' : 'Output'}
                </span>
                <button
                  onClick={handleCopy}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded transition-colors"
                >
                  {copied ? (
                    <>
                      <Check className="w-3 h-3" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="w-3 h-3" />
                      Copy
                    </>
                  )}
                </button>
              </div>
              <pre
                className={`text-xs font-mono p-3 rounded-lg overflow-x-auto max-h-64 overflow-y-auto ${
                  task.error
                    ? 'bg-editor-error/10 text-editor-error'
                    : 'bg-editor-bg text-editor-text'
                }`}
              >
                {task.error || task.output || 'No output'}
              </pre>
            </div>
          )}

          {/* Running indicator */}
          {task.status === 'running' && !hasContent && (
            <div className="p-4 flex items-center justify-center gap-2 text-editor-warning">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span className="text-sm">Task is running...</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
