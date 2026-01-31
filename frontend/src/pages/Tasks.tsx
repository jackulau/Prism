import { useCallback } from 'react';
import { RefreshCw } from 'lucide-react';
import { TaskStatsWidget, TaskFilters, TaskList } from '../components/tasks';
import {
  useTasks,
  useTaskStats,
  useTaskFilters,
  useCancelTask,
  useRetryTask,
} from '../hooks/useTasks';
import { toast } from '../store/toastStore';

export default function Tasks() {
  const { filters, setFilters, clearFilters, hasActiveFilters } = useTaskFilters();
  const { stats, isLoading: statsLoading, error: statsError, refetch: refetchStats } = useTaskStats();
  const {
    tasks,
    isLoading: tasksLoading,
    error: tasksError,
    pagination,
    setPage,
    refetch: refetchTasks,
  } = useTasks(filters);

  const { cancelTask, isLoading: cancelLoading } = useCancelTask();
  const { retryTask, isLoading: retryLoading } = useRetryTask();

  const isActionPending = cancelLoading || retryLoading;

  const handleRefresh = useCallback(() => {
    refetchStats();
    refetchTasks();
  }, [refetchStats, refetchTasks]);

  const handleCancel = useCallback(
    async (taskId: string) => {
      const success = await cancelTask(taskId);
      if (success) {
        toast.success('Task cancelled');
        handleRefresh();
      } else {
        toast.error('Failed to cancel task');
      }
    },
    [cancelTask, handleRefresh]
  );

  const handleRetry = useCallback(
    async (taskId: string) => {
      const success = await retryTask(taskId);
      if (success) {
        toast.success('Task queued for retry');
        handleRefresh();
      } else {
        toast.error('Failed to retry task');
      }
    },
    [retryTask, handleRefresh]
  );

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Task Queue</h1>
            <p className="text-editor-muted">
              Monitor and manage your agent task executions
            </p>
          </div>
          <button
            onClick={handleRefresh}
            disabled={tasksLoading || statsLoading}
            className="flex items-center gap-2 px-4 py-2 text-editor-muted hover:text-editor-text hover:bg-editor-hover rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw
              size={18}
              className={tasksLoading || statsLoading ? 'animate-spin' : ''}
            />
            Refresh
          </button>
        </div>

        {/* Stats Widget */}
        <TaskStatsWidget stats={stats} isLoading={statsLoading} error={statsError} />

        {/* Filters */}
        <TaskFilters
          filters={filters}
          onChange={setFilters}
          onClear={clearFilters}
          hasActiveFilters={hasActiveFilters}
        />

        {/* Task List */}
        <TaskList
          tasks={tasks}
          isLoading={tasksLoading}
          error={tasksError}
          pagination={pagination}
          onPageChange={setPage}
          onCancel={handleCancel}
          onRetry={handleRetry}
          isActionPending={isActionPending}
          onRetryLoad={refetchTasks}
        />
      </div>
    </div>
  );
}
