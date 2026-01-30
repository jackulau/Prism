import React, { useMemo } from 'react';
import {
  CheckCircle2,
  XCircle,
  Clock,
  RefreshCw,
  Loader2,
  Activity,
  XOctagon,
  TrendingUp,
} from 'lucide-react';
import type { BatchTask } from './BatchProgressBar';

interface BatchSummaryProps {
  tasks: BatchTask[];
  onRetryFailed?: () => void;
  isRetrying?: boolean;
  batchStartedAt?: Date;
  batchCompletedAt?: Date;
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color: string;
  bgColor: string;
  subValue?: string;
}

const StatCard: React.FC<StatCardProps> = ({
  icon,
  label,
  value,
  color,
  bgColor,
  subValue,
}) => (
  <div className={`p-4 rounded-lg border ${bgColor} border-current/10`}>
    <div className="flex items-center gap-2 mb-2">
      <span className={color}>{icon}</span>
      <span className="text-xs text-editor-muted uppercase tracking-wide">
        {label}
      </span>
    </div>
    <div className={`text-2xl font-bold ${color}`}>{value}</div>
    {subValue && (
      <div className="text-xs text-editor-muted mt-1">{subValue}</div>
    )}
  </div>
);

export const BatchSummary: React.FC<BatchSummaryProps> = ({
  tasks,
  onRetryFailed,
  isRetrying = false,
  batchStartedAt,
  batchCompletedAt,
}) => {
  const stats = useMemo(() => {
    const total = tasks.length;
    const completed = tasks.filter((t) => t.status === 'completed').length;
    const failed = tasks.filter((t) => t.status === 'failed').length;
    const cancelled = tasks.filter((t) => t.status === 'cancelled').length;
    const running = tasks.filter((t) => t.status === 'running').length;
    const pending = tasks.filter((t) => t.status === 'pending').length;

    const finishedTasks = tasks.filter(
      (t) => t.startedAt && t.completedAt
    );

    let totalDurationMs = 0;
    if (batchStartedAt) {
      const endTime = batchCompletedAt || new Date();
      totalDurationMs = endTime.getTime() - batchStartedAt.getTime();
    } else {
      finishedTasks.forEach((t) => {
        if (t.startedAt && t.completedAt) {
          totalDurationMs += t.completedAt.getTime() - t.startedAt.getTime();
        }
      });
    }

    const successRate = total > 0 ? Math.round((completed / total) * 100) : 0;
    const isComplete = running === 0 && pending === 0;

    return {
      total,
      completed,
      failed,
      cancelled,
      running,
      pending,
      totalDurationMs,
      successRate,
      isComplete,
    };
  }, [tasks, batchStartedAt, batchCompletedAt]);

  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const mins = Math.floor(ms / 60000);
    const secs = Math.round((ms % 60000) / 1000);
    return `${mins}m ${secs}s`;
  };

  const hasFailedTasks = stats.failed > 0 || stats.cancelled > 0;

  return (
    <div className="space-y-4">
      {/* Stats grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatCard
          icon={<Activity className="w-4 h-4" />}
          label="Total Tasks"
          value={stats.total}
          color="text-editor-text"
          bgColor="bg-editor-surface"
        />
        <StatCard
          icon={<CheckCircle2 className="w-4 h-4" />}
          label="Completed"
          value={stats.completed}
          color="text-editor-success"
          bgColor="bg-editor-success/5"
          subValue={`${stats.successRate}% success rate`}
        />
        <StatCard
          icon={<XCircle className="w-4 h-4" />}
          label="Failed"
          value={stats.failed}
          color="text-editor-error"
          bgColor="bg-editor-error/5"
        />
        <StatCard
          icon={<Clock className="w-4 h-4" />}
          label="Duration"
          value={formatDuration(stats.totalDurationMs)}
          color="text-editor-accent"
          bgColor="bg-editor-accent/5"
          subValue={stats.isComplete ? 'Completed' : 'In progress'}
        />
      </div>

      {/* Additional stats row */}
      <div className="flex flex-wrap items-center gap-4 text-sm">
        {stats.running > 0 && (
          <div className="flex items-center gap-2 text-editor-warning">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span>{stats.running} running</span>
          </div>
        )}
        {stats.pending > 0 && (
          <div className="flex items-center gap-2 text-editor-muted">
            <Clock className="w-4 h-4" />
            <span>{stats.pending} pending</span>
          </div>
        )}
        {stats.cancelled > 0 && (
          <div className="flex items-center gap-2 text-editor-muted">
            <XOctagon className="w-4 h-4" />
            <span>{stats.cancelled} cancelled</span>
          </div>
        )}
      </div>

      {/* Success rate bar */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-editor-muted flex items-center gap-1">
            <TrendingUp className="w-3 h-3" />
            Success Rate
          </span>
          <span
            className={
              stats.successRate >= 90
                ? 'text-editor-success'
                : stats.successRate >= 70
                ? 'text-editor-warning'
                : 'text-editor-error'
            }
          >
            {stats.successRate}%
          </span>
        </div>
        <div className="h-2 bg-editor-surface rounded-full overflow-hidden">
          <div
            className={`h-full transition-all duration-500 rounded-full ${
              stats.successRate >= 90
                ? 'bg-editor-success'
                : stats.successRate >= 70
                ? 'bg-editor-warning'
                : 'bg-editor-error'
            }`}
            style={{ width: `${stats.successRate}%` }}
          />
        </div>
      </div>

      {/* Retry failed button */}
      {hasFailedTasks && onRetryFailed && stats.isComplete && (
        <div className="pt-2 border-t border-editor-border">
          <button
            onClick={onRetryFailed}
            disabled={isRetrying}
            className="flex items-center justify-center gap-2 w-full px-4 py-2.5 bg-editor-error/10 text-editor-error border border-editor-error/30 rounded-lg hover:bg-editor-error/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isRetrying ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Retrying {stats.failed + stats.cancelled} tasks...
              </>
            ) : (
              <>
                <RefreshCw className="w-4 h-4" />
                Retry {stats.failed + stats.cancelled} Failed Tasks
              </>
            )}
          </button>
        </div>
      )}
    </div>
  );
};
