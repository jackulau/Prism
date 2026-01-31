import { useState } from 'react';
import {
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  X,
  RotateCcw,
} from 'lucide-react';
import type { Task, TaskStatus } from '../../hooks/useTasks';

interface TaskCardProps {
  task: Task;
  onCancel?: (taskId: string) => void;
  onRetry?: (taskId: string) => void;
  isActionPending?: boolean;
}

const statusConfig: Record<TaskStatus, { icon: typeof Clock; color: string; label: string }> = {
  pending: {
    icon: Clock,
    color: 'text-yellow-500',
    label: 'Pending',
  },
  running: {
    icon: Loader2,
    color: 'text-blue-500',
    label: 'Running',
  },
  completed: {
    icon: CheckCircle,
    color: 'text-green-500',
    label: 'Completed',
  },
  failed: {
    icon: XCircle,
    color: 'text-red-500',
    label: 'Failed',
  },
  cancelled: {
    icon: AlertCircle,
    color: 'text-gray-500',
    label: 'Cancelled',
  },
};

function formatTime(timestamp: number): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}

function formatDuration(task: Task): string {
  if (!task.started_at || !task.completed_at) return '-';

  const durationMs = task.completed_at - task.started_at;
  const seconds = Math.floor(durationMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

function priorityLabel(priority: number): string {
  if (priority >= 3) return 'High';
  if (priority >= 2) return 'Medium';
  return 'Low';
}

export function TaskCard({ task, onCancel, onRetry, isActionPending }: TaskCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const config = statusConfig[task.status];
  const StatusIcon = config.icon;

  const canCancel = task.status === 'pending' || task.status === 'running';
  const canRetry = task.status === 'failed' || task.status === 'cancelled';

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 hover:border-editor-accent/30 transition-colors">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          {/* Status badge and time */}
          <div className="flex items-center gap-2 mb-2">
            <span
              className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${config.color} bg-current/10`}
            >
              <StatusIcon
                size={12}
                className={task.status === 'running' ? 'animate-spin' : ''}
              />
              {config.label}
            </span>
            <span className="text-xs text-editor-muted">
              {formatTime(task.created_at)}
            </span>
          </div>

          {/* Prompt */}
          <p
            className={`text-editor-text ${isExpanded ? '' : 'line-clamp-2'}`}
            title={task.prompt}
          >
            {task.prompt}
          </p>

          {/* Error message for failed tasks */}
          {task.error && (
            <div className="mt-2 p-2 bg-red-500/10 border border-red-500/20 rounded text-sm text-red-400">
              {task.error}
            </div>
          )}

          {/* Result preview for completed tasks */}
          {task.status === 'completed' && task.result && isExpanded && (
            <div className="mt-2 p-2 bg-green-500/10 border border-green-500/20 rounded">
              <p className="text-xs font-medium text-green-400 mb-1">Result:</p>
              <pre className="text-xs text-editor-muted overflow-x-auto">
                {JSON.stringify(task.result, null, 2)}
              </pre>
            </div>
          )}

          {/* Metadata row */}
          <div className="flex items-center gap-4 mt-2 text-xs text-editor-muted">
            <span>Priority: {priorityLabel(task.priority)}</span>
            {(task.status === 'completed' || task.status === 'failed') && (
              <span>Duration: {formatDuration(task)}</span>
            )}
            <span className="font-mono text-editor-muted/70">
              {task.id.slice(0, 8)}...
            </span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1">
          {/* Expand/collapse button */}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors"
            title={isExpanded ? 'Collapse' : 'Expand'}
          >
            {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          </button>

          {/* Cancel button */}
          {canCancel && onCancel && (
            <button
              onClick={() => onCancel(task.id)}
              disabled={isActionPending}
              className="p-1.5 rounded text-editor-muted hover:text-red-400 hover:bg-red-400/10 transition-colors disabled:opacity-50"
              title="Cancel task"
            >
              <X size={16} />
            </button>
          )}

          {/* Retry button */}
          {canRetry && onRetry && (
            <button
              onClick={() => onRetry(task.id)}
              disabled={isActionPending}
              className="p-1.5 rounded text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 transition-colors disabled:opacity-50"
              title="Retry task"
            >
              <RotateCcw size={16} />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
