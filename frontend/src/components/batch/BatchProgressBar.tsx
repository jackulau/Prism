import React, { useMemo } from 'react';
import { Loader2, CheckCircle2, XCircle, Clock } from 'lucide-react';

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface BatchTask {
  id: string;
  status: TaskStatus;
  startedAt?: Date;
  completedAt?: Date;
  error?: string;
  output?: string;
}

interface BatchProgressBarProps {
  tasks: BatchTask[];
  showLabels?: boolean;
  animate?: boolean;
}

interface StatusCount {
  pending: number;
  running: number;
  completed: number;
  failed: number;
  cancelled: number;
}

const statusColors: Record<TaskStatus, string> = {
  pending: 'bg-editor-muted',
  running: 'bg-editor-warning',
  completed: 'bg-editor-success',
  failed: 'bg-editor-error',
  cancelled: 'bg-editor-muted/50',
};

const statusLabels: Record<TaskStatus, string> = {
  pending: 'Pending',
  running: 'Running',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

export const BatchProgressBar: React.FC<BatchProgressBarProps> = ({
  tasks,
  showLabels = true,
  animate = true,
}) => {
  const counts = useMemo<StatusCount>(() => {
    const result: StatusCount = {
      pending: 0,
      running: 0,
      completed: 0,
      failed: 0,
      cancelled: 0,
    };
    tasks.forEach((task) => {
      result[task.status]++;
    });
    return result;
  }, [tasks]);

  const total = tasks.length;
  const completedOrFailed = counts.completed + counts.failed + counts.cancelled;
  const percentage = total > 0 ? Math.round((completedOrFailed / total) * 100) : 0;

  const getSegmentWidth = (count: number): string => {
    if (total === 0) return '0%';
    return `${(count / total) * 100}%`;
  };

  const isActive = counts.running > 0;

  return (
    <div className="space-y-3">
      {/* Progress header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {isActive ? (
            <Loader2 className="w-4 h-4 text-editor-warning animate-spin" />
          ) : percentage === 100 && counts.failed === 0 ? (
            <CheckCircle2 className="w-4 h-4 text-editor-success" />
          ) : percentage === 100 ? (
            <XCircle className="w-4 h-4 text-editor-error" />
          ) : (
            <Clock className="w-4 h-4 text-editor-muted" />
          )}
          <span className="text-sm font-medium text-editor-text">
            Batch Progress
          </span>
        </div>
        <span className="text-sm text-editor-muted">
          {completedOrFailed} / {total} ({percentage}%)
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-3 bg-editor-surface rounded-full overflow-hidden flex">
        {/* Completed segment */}
        {counts.completed > 0 && (
          <div
            className={`${statusColors.completed} transition-all duration-300 ease-out`}
            style={{ width: getSegmentWidth(counts.completed) }}
          />
        )}
        {/* Running segment */}
        {counts.running > 0 && (
          <div
            className={`${statusColors.running} transition-all duration-300 ease-out ${
              animate ? 'animate-pulse' : ''
            }`}
            style={{ width: getSegmentWidth(counts.running) }}
          />
        )}
        {/* Failed segment */}
        {counts.failed > 0 && (
          <div
            className={`${statusColors.failed} transition-all duration-300 ease-out`}
            style={{ width: getSegmentWidth(counts.failed) }}
          />
        )}
        {/* Cancelled segment */}
        {counts.cancelled > 0 && (
          <div
            className={`${statusColors.cancelled} transition-all duration-300 ease-out`}
            style={{ width: getSegmentWidth(counts.cancelled) }}
          />
        )}
        {/* Pending segment */}
        {counts.pending > 0 && (
          <div
            className={`${statusColors.pending} transition-all duration-300 ease-out`}
            style={{ width: getSegmentWidth(counts.pending) }}
          />
        )}
      </div>

      {/* Status labels */}
      {showLabels && (
        <div className="flex flex-wrap gap-4 text-xs">
          {(Object.keys(statusColors) as TaskStatus[]).map((status) => {
            const count = counts[status];
            if (count === 0) return null;
            return (
              <div key={status} className="flex items-center gap-1.5">
                <div className={`w-2.5 h-2.5 rounded-full ${statusColors[status]}`} />
                <span className="text-editor-muted">
                  {statusLabels[status]}: {count}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
