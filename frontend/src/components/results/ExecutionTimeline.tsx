import React from 'react';
import { Play, CheckCircle, XCircle, Clock, Loader2 } from 'lucide-react';

export type TimelineItemStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface TimelineItem {
  id: string;
  label: string;
  status: TimelineItemStatus;
  startedAt?: Date;
  completedAt?: Date;
  duration?: number; // milliseconds
  parallel?: boolean; // if true, runs in parallel with next items
}

interface ExecutionTimelineProps {
  items: TimelineItem[];
  executionStartedAt?: Date;
  executionCompletedAt?: Date;
  className?: string;
}

const statusColors: Record<TimelineItemStatus, {
  bg: string;
  border: string;
  text: string;
  bar: string;
}> = {
  pending: {
    bg: 'bg-editor-muted/20',
    border: 'border-editor-muted/30',
    text: 'text-editor-muted',
    bar: 'bg-editor-muted/40',
  },
  running: {
    bg: 'bg-editor-warning/20',
    border: 'border-editor-warning/30',
    text: 'text-editor-warning',
    bar: 'bg-editor-warning',
  },
  completed: {
    bg: 'bg-editor-success/20',
    border: 'border-editor-success/30',
    text: 'text-editor-success',
    bar: 'bg-editor-success',
  },
  failed: {
    bg: 'bg-editor-error/20',
    border: 'border-editor-error/30',
    text: 'text-editor-error',
    bar: 'bg-editor-error',
  },
};

const StatusIcon: React.FC<{ status: TimelineItemStatus }> = ({ status }) => {
  switch (status) {
    case 'pending':
      return <Clock className="w-3 h-3" />;
    case 'running':
      return <Loader2 className="w-3 h-3 animate-spin" />;
    case 'completed':
      return <CheckCircle className="w-3 h-3" />;
    case 'failed':
      return <XCircle className="w-3 h-3" />;
  }
};

function formatDuration(ms?: number): string {
  if (ms === undefined || ms === 0) return '--';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function calculateProgress(item: TimelineItem, executionStartedAt?: Date): {
  left: number;
  width: number;
} {
  if (!executionStartedAt || !item.startedAt) {
    return { left: 0, width: 0 };
  }

  const totalDuration = Date.now() - executionStartedAt.getTime();
  if (totalDuration === 0) return { left: 0, width: 100 };

  const startOffset = item.startedAt.getTime() - executionStartedAt.getTime();
  const itemDuration = item.duration || (item.completedAt
    ? item.completedAt.getTime() - item.startedAt.getTime()
    : Date.now() - item.startedAt.getTime());

  const left = Math.max(0, Math.min(100, (startOffset / totalDuration) * 100));
  const width = Math.max(5, Math.min(100 - left, (itemDuration / totalDuration) * 100));

  return { left, width };
}

export const ExecutionTimeline: React.FC<ExecutionTimelineProps> = ({
  items,
  executionStartedAt,
  executionCompletedAt,
  className = '',
}) => {
  if (items.length === 0) {
    return (
      <div className={`bg-editor-surface border border-editor-border rounded-lg p-6 text-center ${className}`}>
        <Clock className="w-8 h-8 text-editor-muted mx-auto mb-2" />
        <p className="text-editor-muted">No timeline data available</p>
      </div>
    );
  }

  const totalDuration = executionStartedAt
    ? (executionCompletedAt || new Date()).getTime() - executionStartedAt.getTime()
    : 0;

  return (
    <div className={`bg-editor-surface border border-editor-border rounded-lg p-4 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <Play className="w-4 h-4 text-editor-accent" />
          Execution Timeline
        </h3>
        <span className="text-xs text-editor-muted">
          Total: {formatDuration(totalDuration)}
        </span>
      </div>

      {/* Timeline bars */}
      <div className="space-y-3">
        {items.map((item, index) => {
          const colors = statusColors[item.status];
          const { left, width } = calculateProgress(item, executionStartedAt);

          return (
            <div key={item.id} className="group">
              {/* Label row */}
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className={`${colors.text}`}>
                    <StatusIcon status={item.status} />
                  </span>
                  <span className="text-sm text-editor-text truncate max-w-48">
                    {item.label}
                  </span>
                  {item.parallel && index < items.length - 1 && (
                    <span className="text-xs px-1.5 py-0.5 bg-editor-accent/10 text-editor-accent rounded">
                      parallel
                    </span>
                  )}
                </div>
                <span className="text-xs text-editor-muted">
                  {formatDuration(item.duration)}
                </span>
              </div>

              {/* Bar container */}
              <div className="relative h-6 bg-editor-bg rounded overflow-hidden">
                {/* Background grid lines */}
                <div className="absolute inset-0 flex">
                  {[...Array(10)].map((_, i) => (
                    <div
                      key={i}
                      className="flex-1 border-r border-editor-border/30 last:border-r-0"
                    />
                  ))}
                </div>

                {/* Progress bar */}
                <div
                  className={`absolute top-1 bottom-1 rounded transition-all duration-300 ${colors.bar}`}
                  style={{
                    left: `${left}%`,
                    width: `${width}%`,
                    minWidth: item.status !== 'pending' ? '4px' : '0',
                  }}
                >
                  {/* Animated pulse for running items */}
                  {item.status === 'running' && (
                    <div className="absolute inset-0 bg-white/20 animate-pulse rounded" />
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Time scale */}
      <div className="mt-4 pt-2 border-t border-editor-border">
        <div className="flex justify-between text-xs text-editor-muted">
          <span>0s</span>
          <span>{formatDuration(totalDuration / 4)}</span>
          <span>{formatDuration(totalDuration / 2)}</span>
          <span>{formatDuration((totalDuration / 4) * 3)}</span>
          <span>{formatDuration(totalDuration)}</span>
        </div>
      </div>

      {/* Legend */}
      <div className="mt-3 flex items-center gap-4 text-xs">
        {(['completed', 'running', 'failed', 'pending'] as const).map((status) => {
          const colors = statusColors[status];
          const hasItemWithStatus = items.some(i => i.status === status);
          if (!hasItemWithStatus) return null;

          return (
            <div key={status} className="flex items-center gap-1.5">
              <div className={`w-3 h-3 rounded ${colors.bar}`} />
              <span className="text-editor-muted capitalize">{status}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
