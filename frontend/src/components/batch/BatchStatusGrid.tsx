import React, { useMemo, useCallback, useState, useRef, useEffect } from 'react';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  XOctagon,
  Grid3X3,
  List,
} from 'lucide-react';
import type { BatchTask, TaskStatus } from './BatchProgressBar';
import { TaskResultCard } from './TaskResultCard';

interface BatchStatusGridProps {
  tasks: BatchTask[];
  onRetry?: (taskId: string) => void;
  retryingTasks?: Set<string>;
  viewMode?: 'grid' | 'list';
  onViewModeChange?: (mode: 'grid' | 'list') => void;
  virtualize?: boolean;
  virtualRowHeight?: number;
}

const statusIcons: Record<TaskStatus, React.ReactNode> = {
  pending: <Clock className="w-4 h-4" />,
  running: <Loader2 className="w-4 h-4 animate-spin" />,
  completed: <CheckCircle2 className="w-4 h-4" />,
  failed: <XCircle className="w-4 h-4" />,
  cancelled: <XOctagon className="w-4 h-4" />,
};

const statusColors: Record<TaskStatus, string> = {
  pending: 'text-editor-muted bg-editor-muted/10 border-editor-muted/20',
  running: 'text-editor-warning bg-editor-warning/10 border-editor-warning/30',
  completed: 'text-editor-success bg-editor-success/10 border-editor-success/20',
  failed: 'text-editor-error bg-editor-error/10 border-editor-error/30',
  cancelled: 'text-editor-muted bg-editor-muted/5 border-editor-muted/10',
};

interface GridCellProps {
  task: BatchTask;
  onClick: (taskId: string) => void;
}

const GridCell: React.FC<GridCellProps> = ({ task, onClick }) => {
  const colorClass = statusColors[task.status];

  return (
    <button
      onClick={() => onClick(task.id)}
      className={`p-3 rounded-lg border transition-all hover:scale-[1.02] ${colorClass}`}
      title={`${task.id} - ${task.status}`}
    >
      <div className="flex items-center gap-2">
        {statusIcons[task.status]}
        <span className="text-xs font-medium truncate">{task.id}</span>
      </div>
    </button>
  );
};

export const BatchStatusGrid: React.FC<BatchStatusGridProps> = ({
  tasks,
  onRetry,
  retryingTasks = new Set(),
  viewMode = 'grid',
  onViewModeChange,
  virtualize = false,
  virtualRowHeight = 100,
}) => {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [visibleRange, setVisibleRange] = useState({ start: 0, end: 50 });

  const selectedTask = useMemo(
    () => tasks.find((t) => t.id === selectedTaskId) || null,
    [tasks, selectedTaskId]
  );

  const handleTaskClick = useCallback((taskId: string) => {
    setSelectedTaskId((prev) => (prev === taskId ? null : taskId));
  }, []);

  const handleRetry = useCallback(
    (taskId: string) => {
      if (onRetry) {
        onRetry(taskId);
      }
    },
    [onRetry]
  );

  // Handle scroll for virtualization
  useEffect(() => {
    if (!virtualize || !containerRef.current) return;

    const handleScroll = () => {
      if (!containerRef.current) return;
      const { scrollTop, clientHeight } = containerRef.current;
      const start = Math.floor(scrollTop / virtualRowHeight);
      const visible = Math.ceil(clientHeight / virtualRowHeight);
      const buffer = 5;
      setVisibleRange({
        start: Math.max(0, start - buffer),
        end: Math.min(tasks.length, start + visible + buffer * 2),
      });
    };

    const container = containerRef.current;
    container.addEventListener('scroll', handleScroll);
    handleScroll(); // Initial calculation

    return () => container.removeEventListener('scroll', handleScroll);
  }, [virtualize, virtualRowHeight, tasks.length]);

  const visibleTasks = useMemo(() => {
    if (!virtualize) return tasks;
    return tasks.slice(visibleRange.start, visibleRange.end);
  }, [tasks, virtualize, visibleRange]);

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-editor-muted">
        <Grid3X3 className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-sm">No tasks in batch</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* View mode toggle */}
      {onViewModeChange && (
        <div className="flex items-center justify-end">
          <div className="flex items-center bg-editor-surface border border-editor-border rounded-lg p-0.5">
            <button
              onClick={() => onViewModeChange('grid')}
              className={`p-1.5 rounded transition-colors ${
                viewMode === 'grid'
                  ? 'bg-editor-accent text-white'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
              title="Grid view"
            >
              <Grid3X3 className="w-4 h-4" />
            </button>
            <button
              onClick={() => onViewModeChange('list')}
              className={`p-1.5 rounded transition-colors ${
                viewMode === 'list'
                  ? 'bg-editor-accent text-white'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
              title="List view"
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Grid view */}
      {viewMode === 'grid' && (
        <div
          ref={containerRef}
          className={`${virtualize ? 'overflow-y-auto max-h-96' : ''}`}
          style={virtualize ? { height: '400px' } : undefined}
        >
          {virtualize && (
            <div style={{ height: visibleRange.start * virtualRowHeight }} />
          )}
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-2">
            {visibleTasks.map((task) => (
              <GridCell key={task.id} task={task} onClick={handleTaskClick} />
            ))}
          </div>
          {virtualize && (
            <div
              style={{
                height: (tasks.length - visibleRange.end) * virtualRowHeight,
              }}
            />
          )}
        </div>
      )}

      {/* List view */}
      {viewMode === 'list' && (
        <div
          ref={containerRef}
          className={`space-y-2 ${virtualize ? 'overflow-y-auto max-h-96' : ''}`}
          style={virtualize ? { height: '400px' } : undefined}
        >
          {virtualize && (
            <div style={{ height: visibleRange.start * virtualRowHeight }} />
          )}
          {visibleTasks.map((task) => (
            <TaskResultCard
              key={task.id}
              task={task}
              onRetry={onRetry ? handleRetry : undefined}
              isRetrying={retryingTasks.has(task.id)}
              defaultExpanded={false}
            />
          ))}
          {virtualize && (
            <div
              style={{
                height: (tasks.length - visibleRange.end) * virtualRowHeight,
              }}
            />
          )}
        </div>
      )}

      {/* Selected task detail modal (for grid view) */}
      {viewMode === 'grid' && selectedTask && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div
            className="bg-editor-bg border border-editor-border rounded-lg max-w-2xl w-full max-h-[80vh] overflow-y-auto"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="sticky top-0 bg-editor-bg border-b border-editor-border p-3 flex items-center justify-between">
              <span className="font-medium text-editor-text">Task Details</span>
              <button
                onClick={() => setSelectedTaskId(null)}
                className="p-1 text-editor-muted hover:text-editor-text rounded"
              >
                <XCircle className="w-5 h-5" />
              </button>
            </div>
            <div className="p-4">
              <TaskResultCard
                task={selectedTask}
                onRetry={onRetry ? handleRetry : undefined}
                isRetrying={retryingTasks.has(selectedTask.id)}
                defaultExpanded={true}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
