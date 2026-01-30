import { useState } from 'react';
import {
  Plus,
  Trash2,
  GripVertical,
  CheckCircle2,
  XCircle,
  Loader2,
  Clock,
  AlertCircle,
} from 'lucide-react';
import { useBatchStore } from '../../store/batchStore';
import type { BatchTask, BatchTaskStatus } from '../../types/batch';

export function BatchTaskList() {
  const { tasks, addTask, removeTask, clearTasks, isRunning } = useBatchStore();
  const [newPrompt, setNewPrompt] = useState('');
  const [newSystemPrompt, setNewSystemPrompt] = useState('');
  const [showSystemPrompt, setShowSystemPrompt] = useState(false);

  const handleAddTask = () => {
    if (!newPrompt.trim()) return;
    addTask(newPrompt.trim(), newSystemPrompt.trim() || undefined);
    setNewPrompt('');
    setNewSystemPrompt('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleAddTask();
    }
  };

  const getStatusIcon = (status: BatchTaskStatus) => {
    switch (status) {
      case 'completed':
        return <CheckCircle2 size={14} className="text-editor-success" />;
      case 'failed':
        return <XCircle size={14} className="text-editor-error" />;
      case 'running':
        return <Loader2 size={14} className="text-editor-accent animate-spin" />;
      case 'cancelled':
        return <AlertCircle size={14} className="text-editor-warning" />;
      default:
        return <Clock size={14} className="text-editor-muted" />;
    }
  };

  const pendingCount = tasks.filter((t) => t.status === 'pending').length;
  const runningCount = tasks.filter((t) => t.status === 'running').length;
  const completedCount = tasks.filter((t) => t.status === 'completed').length;
  const failedCount = tasks.filter((t) => t.status === 'failed').length;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-editor-text">
          Tasks ({tasks.length})
        </h3>
        {tasks.length > 0 && !isRunning && (
          <button
            onClick={clearTasks}
            className="text-xs text-editor-muted hover:text-editor-error transition-colors"
          >
            Clear All
          </button>
        )}
      </div>

      {/* Stats */}
      {tasks.length > 0 && (
        <div className="flex items-center gap-3 text-xs">
          {pendingCount > 0 && (
            <span className="flex items-center gap-1 text-editor-muted">
              <Clock size={12} />
              {pendingCount} pending
            </span>
          )}
          {runningCount > 0 && (
            <span className="flex items-center gap-1 text-editor-accent">
              <Loader2 size={12} className="animate-spin" />
              {runningCount} running
            </span>
          )}
          {completedCount > 0 && (
            <span className="flex items-center gap-1 text-editor-success">
              <CheckCircle2 size={12} />
              {completedCount} done
            </span>
          )}
          {failedCount > 0 && (
            <span className="flex items-center gap-1 text-editor-error">
              <XCircle size={12} />
              {failedCount} failed
            </span>
          )}
        </div>
      )}

      {/* Add Task Form */}
      {!isRunning && (
        <div className="space-y-2">
          <div className="flex gap-2">
            <textarea
              value={newPrompt}
              onChange={(e) => setNewPrompt(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Enter task prompt..."
              rows={2}
              className="flex-1 px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none resize-none"
            />
            <button
              onClick={handleAddTask}
              disabled={!newPrompt.trim()}
              className="px-3 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors self-end"
            >
              <Plus size={18} />
            </button>
          </div>

          <button
            onClick={() => setShowSystemPrompt(!showSystemPrompt)}
            className="text-xs text-editor-muted hover:text-editor-text transition-colors"
          >
            {showSystemPrompt ? '- Hide' : '+ Add'} system prompt
          </button>

          {showSystemPrompt && (
            <textarea
              value={newSystemPrompt}
              onChange={(e) => setNewSystemPrompt(e.target.value)}
              placeholder="Optional system prompt for this task..."
              rows={2}
              className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none resize-none"
            />
          )}
        </div>
      )}

      {/* Task List */}
      <div className="space-y-2 max-h-[400px] overflow-y-auto">
        {tasks.length === 0 ? (
          <div className="py-8 text-center text-editor-muted">
            <p className="text-sm">No tasks added yet</p>
            <p className="text-xs mt-1">Add prompts above to create a batch</p>
          </div>
        ) : (
          tasks.map((task, index) => (
            <TaskItem
              key={task.id}
              task={task}
              index={index}
              onRemove={() => removeTask(task.id)}
              getStatusIcon={getStatusIcon}
              disabled={isRunning}
            />
          ))
        )}
      </div>
    </div>
  );
}

interface TaskItemProps {
  task: BatchTask;
  index: number;
  onRemove: () => void;
  getStatusIcon: (status: BatchTaskStatus) => JSX.Element;
  disabled: boolean;
}

function TaskItem({ task, index, onRemove, getStatusIcon, disabled }: TaskItemProps) {
  return (
    <div
      className={`flex items-start gap-2 p-3 rounded-lg border transition-colors ${
        task.status === 'running'
          ? 'bg-editor-accent/10 border-editor-accent/30'
          : task.status === 'completed'
          ? 'bg-editor-success/10 border-editor-success/30'
          : task.status === 'failed'
          ? 'bg-editor-error/10 border-editor-error/30'
          : 'bg-editor-surface border-editor-border'
      }`}
    >
      <div className="flex items-center gap-2 text-editor-muted">
        {!disabled && (
          <GripVertical size={14} className="cursor-grab opacity-50 hover:opacity-100" />
        )}
        <span className="text-xs w-6">{index + 1}.</span>
      </div>

      <div className="flex-1 min-w-0">
        <p className="text-sm text-editor-text line-clamp-2">{task.prompt}</p>
        {task.systemPrompt && (
          <p className="text-xs text-editor-muted mt-1 line-clamp-1">
            System: {task.systemPrompt}
          </p>
        )}
        {task.status === 'running' && task.progress !== undefined && (
          <div className="mt-2">
            <div className="h-1 bg-editor-border rounded-full overflow-hidden">
              <div
                className="h-full bg-editor-accent transition-all duration-300"
                style={{ width: `${task.progress}%` }}
              />
            </div>
            <span className="text-xs text-editor-muted">{task.progress}%</span>
          </div>
        )}
        {task.error && (
          <p className="text-xs text-editor-error mt-1">{task.error}</p>
        )}
      </div>

      <div className="flex items-center gap-2">
        {getStatusIcon(task.status)}
        {!disabled && (
          <button
            onClick={onRemove}
            className="p-1 text-editor-muted hover:text-editor-error transition-colors"
          >
            <Trash2 size={14} />
          </button>
        )}
      </div>
    </div>
  );
}
