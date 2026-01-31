import { useState } from 'react';
import { ChevronDown, ChevronUp, Edit2, Trash2, Loader2, CheckCircle, XCircle, Clock } from 'lucide-react';
import type { BatchTask, BatchTaskPriority } from '../../types/batch';

interface BatchTaskItemProps {
  task: BatchTask;
  onEdit: (task: BatchTask) => void;
  onDelete: (taskId: string) => void;
  disabled?: boolean;
}

const priorityColors: Record<BatchTaskPriority, string> = {
  low: 'bg-gray-500/20 text-gray-400',
  normal: 'bg-blue-500/20 text-blue-400',
  high: 'bg-orange-500/20 text-orange-400',
};

const statusIcons: Record<BatchTask['status'], React.ReactNode> = {
  pending: <Clock size={16} className="text-editor-muted" />,
  running: <Loader2 size={16} className="text-blue-400 animate-spin" />,
  completed: <CheckCircle size={16} className="text-green-400" />,
  failed: <XCircle size={16} className="text-red-400" />,
};

export function BatchTaskItem({ task, onEdit, onDelete, disabled = false }: BatchTaskItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const truncatedPrompt = task.prompt.length > 100
    ? task.prompt.slice(0, 100) + '...'
    : task.prompt;

  const isRunning = task.status === 'running';
  const canEdit = !isRunning && !disabled;

  return (
    <div className="border border-editor-border rounded-lg bg-editor-surface overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-3 p-3">
        {/* Status Icon */}
        <div className="flex-shrink-0">
          {statusIcons[task.status]}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <p className="text-sm text-editor-text truncate">
            {isExpanded ? task.prompt : truncatedPrompt}
          </p>
          {task.context && !isExpanded && (
            <p className="text-xs text-editor-muted mt-1 truncate">
              Context: {task.context}
            </p>
          )}
        </div>

        {/* Priority Badge */}
        <span className={`flex-shrink-0 px-2 py-0.5 text-xs rounded-full ${priorityColors[task.priority]}`}>
          {task.priority}
        </span>

        {/* Actions */}
        <div className="flex items-center gap-1">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded transition-colors"
            aria-label={isExpanded ? 'Collapse' : 'Expand'}
          >
            {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          </button>
          <button
            onClick={() => onEdit(task)}
            disabled={!canEdit}
            className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            aria-label="Edit task"
          >
            <Edit2 size={16} />
          </button>
          <button
            onClick={() => onDelete(task.id)}
            disabled={!canEdit}
            className="p-1.5 text-editor-muted hover:text-red-400 hover:bg-editor-bg rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            aria-label="Delete task"
          >
            <Trash2 size={16} />
          </button>
        </div>
      </div>

      {/* Expanded Content */}
      {isExpanded && (
        <div className="px-3 pb-3 pt-0 border-t border-editor-border/50">
          <div className="mt-3 space-y-2">
            <div>
              <p className="text-xs font-medium text-editor-muted mb-1">Prompt</p>
              <p className="text-sm text-editor-text whitespace-pre-wrap">{task.prompt}</p>
            </div>
            {task.context && (
              <div>
                <p className="text-xs font-medium text-editor-muted mb-1">Context</p>
                <p className="text-sm text-editor-text whitespace-pre-wrap">{task.context}</p>
              </div>
            )}
            {task.result && (
              <div>
                <p className="text-xs font-medium text-green-400 mb-1">Result</p>
                <p className="text-sm text-editor-text whitespace-pre-wrap">{task.result}</p>
              </div>
            )}
            {task.error && (
              <div>
                <p className="text-xs font-medium text-red-400 mb-1">Error</p>
                <p className="text-sm text-red-300 whitespace-pre-wrap">{task.error}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
