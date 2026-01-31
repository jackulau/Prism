import { Loader2, Inbox, ChevronLeft, ChevronRight } from 'lucide-react';
import { TaskCard } from './TaskCard';
import type { Task, PaginationState } from '../../hooks/useTasks';

interface TaskListProps {
  tasks: Task[];
  isLoading: boolean;
  error?: string | null;
  pagination: PaginationState;
  onPageChange: (page: number) => void;
  onCancel?: (taskId: string) => void;
  onRetry?: (taskId: string) => void;
  isActionPending?: boolean;
  onRetryLoad?: () => void;
}

export function TaskList({
  tasks,
  isLoading,
  error,
  pagination,
  onPageChange,
  onCancel,
  onRetry,
  isActionPending,
  onRetryLoad,
}: TaskListProps) {
  const totalPages = Math.ceil(pagination.total / pagination.limit);
  const hasNextPage = pagination.page < totalPages;
  const hasPrevPage = pagination.page > 1;

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <Loader2 size={32} className="animate-spin text-editor-accent mb-4" />
        <p className="text-editor-muted">Loading tasks...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-6 text-center max-w-md">
          <p className="text-red-400 mb-4">{error}</p>
          {onRetryLoad && (
            <button
              onClick={onRetryLoad}
              className="px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              Try Again
            </button>
          )}
        </div>
      </div>
    );
  }

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <div className="p-4 rounded-full bg-editor-surface mb-4">
          <Inbox size={32} className="text-editor-muted" />
        </div>
        <h3 className="text-lg font-medium text-editor-text mb-1">No tasks found</h3>
        <p className="text-editor-muted text-center max-w-md">
          There are no tasks matching your current filters. Tasks will appear here when agents execute prompts.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Task cards */}
      <div className="space-y-3">
        {tasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            onCancel={onCancel}
            onRetry={onRetry}
            isActionPending={isActionPending}
          />
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-4 border-t border-editor-border">
          <p className="text-sm text-editor-muted">
            Showing {(pagination.page - 1) * pagination.limit + 1} to{' '}
            {Math.min(pagination.page * pagination.limit, pagination.total)} of{' '}
            {pagination.total} tasks
          </p>

          <div className="flex items-center gap-2">
            <button
              onClick={() => onPageChange(pagination.page - 1)}
              disabled={!hasPrevPage}
              className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Previous page"
            >
              <ChevronLeft size={18} />
            </button>

            <span className="text-sm text-editor-text px-2">
              Page {pagination.page} of {totalPages}
            </span>

            <button
              onClick={() => onPageChange(pagination.page + 1)}
              disabled={!hasNextPage}
              className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Next page"
            >
              <ChevronRight size={18} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
