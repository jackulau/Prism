import {
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  Timer,
  Zap,
  Activity,
} from 'lucide-react';
import { useBatchStore } from '../../store/batchStore';

export function BatchProgressTracker() {
  const { tasks, execution, isRunning } = useBatchStore();

  const totalTasks = tasks.length;
  const completedTasks = tasks.filter((t) => t.status === 'completed').length;
  const failedTasks = tasks.filter((t) => t.status === 'failed').length;
  const runningTasks = tasks.filter((t) => t.status === 'running').length;
  const pendingTasks = tasks.filter((t) => t.status === 'pending').length;
  const cancelledTasks = tasks.filter((t) => t.status === 'cancelled').length;

  const progressPercent = totalTasks > 0
    ? Math.round(((completedTasks + failedTasks + cancelledTasks) / totalTasks) * 100)
    : 0;

  const totalTokens = tasks.reduce((sum, t) => sum + (t.tokensUsed || 0), 0);
  const totalDuration = tasks.reduce((sum, t) => sum + (t.duration || 0), 0);
  const avgDuration = completedTasks > 0 ? totalDuration / completedTasks : 0;

  const elapsedTime = execution?.startedAt
    ? Date.now() - execution.startedAt.getTime()
    : 0;

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    const mins = Math.floor(ms / 60000);
    const secs = Math.floor((ms % 60000) / 1000);
    return `${mins}m ${secs}s`;
  };

  if (!execution && totalTasks === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-editor-muted">
        <Activity size={48} className="mb-4 opacity-50" />
        <p className="text-sm">No batch running</p>
        <p className="text-xs mt-1">Add tasks and start a batch to track progress</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-editor-text flex items-center gap-2">
          <Activity size={16} />
          Progress
        </h3>
        {isRunning && (
          <span className="flex items-center gap-1 text-xs text-editor-accent">
            <Loader2 size={12} className="animate-spin" />
            Running
          </span>
        )}
        {execution?.status === 'completed' && (
          <span className="flex items-center gap-1 text-xs text-editor-success">
            <CheckCircle2 size={12} />
            Completed
          </span>
        )}
        {execution?.status === 'cancelled' && (
          <span className="flex items-center gap-1 text-xs text-editor-warning">
            <XCircle size={12} />
            Cancelled
          </span>
        )}
      </div>

      {/* Main Progress Bar */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-editor-muted">Overall Progress</span>
          <span className="text-editor-text font-medium">{progressPercent}%</span>
        </div>
        <div className="h-3 bg-editor-surface rounded-full overflow-hidden border border-editor-border">
          <div className="h-full flex">
            {completedTasks > 0 && (
              <div
                className="bg-editor-success transition-all duration-500"
                style={{ width: `${(completedTasks / totalTasks) * 100}%` }}
              />
            )}
            {failedTasks > 0 && (
              <div
                className="bg-editor-error transition-all duration-500"
                style={{ width: `${(failedTasks / totalTasks) * 100}%` }}
              />
            )}
            {cancelledTasks > 0 && (
              <div
                className="bg-editor-warning transition-all duration-500"
                style={{ width: `${(cancelledTasks / totalTasks) * 100}%` }}
              />
            )}
            {runningTasks > 0 && (
              <div
                className="bg-editor-accent animate-pulse transition-all duration-500"
                style={{ width: `${(runningTasks / totalTasks) * 100}%` }}
              />
            )}
          </div>
        </div>
        <div className="flex items-center justify-between text-xs text-editor-muted">
          <span>
            {completedTasks + failedTasks + cancelledTasks} / {totalTasks} tasks
          </span>
          {isRunning && <span>{runningTasks} running</span>}
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-3">
        <StatCard
          icon={<CheckCircle2 size={14} className="text-editor-success" />}
          label="Completed"
          value={completedTasks}
          subtext={`${totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0}%`}
        />
        <StatCard
          icon={<XCircle size={14} className="text-editor-error" />}
          label="Failed"
          value={failedTasks}
          subtext={`${totalTasks > 0 ? Math.round((failedTasks / totalTasks) * 100) : 0}%`}
        />
        <StatCard
          icon={<Clock size={14} className="text-editor-muted" />}
          label="Pending"
          value={pendingTasks}
        />
        <StatCard
          icon={<Loader2 size={14} className="text-editor-accent" />}
          label="Running"
          value={runningTasks}
        />
      </div>

      {/* Timing & Token Stats */}
      <div className="space-y-3 pt-4 border-t border-editor-border">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-2 text-xs text-editor-muted">
            <Timer size={12} />
            Elapsed Time
          </span>
          <span className="text-sm text-editor-text font-medium">
            {formatDuration(elapsedTime)}
          </span>
        </div>

        {avgDuration > 0 && (
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-xs text-editor-muted">
              <Clock size={12} />
              Avg. Duration
            </span>
            <span className="text-sm text-editor-text">
              {formatDuration(avgDuration)}
            </span>
          </div>
        )}

        {totalTokens > 0 && (
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-xs text-editor-muted">
              <Zap size={12} />
              Total Tokens
            </span>
            <span className="text-sm text-editor-text">
              {totalTokens.toLocaleString()}
            </span>
          </div>
        )}

        {pendingTasks > 0 && avgDuration > 0 && (
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-xs text-editor-muted">
              <Timer size={12} />
              Est. Remaining
            </span>
            <span className="text-sm text-editor-text">
              {formatDuration(pendingTasks * avgDuration)}
            </span>
          </div>
        )}
      </div>

      {/* Task Status List */}
      {tasks.length > 0 && (
        <div className="space-y-2 pt-4 border-t border-editor-border">
          <span className="text-xs text-editor-muted">Task Status</span>
          <div className="max-h-[200px] overflow-y-auto space-y-1">
            {tasks.map((task, index) => (
              <div
                key={task.id}
                className="flex items-center gap-2 p-2 rounded bg-editor-surface/50 text-xs"
              >
                <span className="text-editor-muted w-6">{index + 1}.</span>
                <TaskStatusIndicator status={task.status} progress={task.progress} />
                <span className="flex-1 truncate text-editor-text">
                  {task.prompt.slice(0, 40)}...
                </span>
                {task.duration && (
                  <span className="text-editor-muted">
                    {formatDuration(task.duration)}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: number;
  subtext?: string;
}

function StatCard({ icon, label, value, subtext }: StatCardProps) {
  return (
    <div className="p-3 rounded-lg bg-editor-surface border border-editor-border">
      <div className="flex items-center gap-2 mb-1">
        {icon}
        <span className="text-xs text-editor-muted">{label}</span>
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-xl font-semibold text-editor-text">{value}</span>
        {subtext && <span className="text-xs text-editor-muted">{subtext}</span>}
      </div>
    </div>
  );
}

interface TaskStatusIndicatorProps {
  status: string;
  progress?: number;
}

function TaskStatusIndicator({ status, progress }: TaskStatusIndicatorProps) {
  switch (status) {
    case 'completed':
      return <CheckCircle2 size={12} className="text-editor-success" />;
    case 'failed':
      return <XCircle size={12} className="text-editor-error" />;
    case 'cancelled':
      return <XCircle size={12} className="text-editor-warning" />;
    case 'running':
      return (
        <div className="relative w-4 h-4">
          <Loader2 size={12} className="text-editor-accent animate-spin" />
          {progress !== undefined && (
            <span className="absolute -bottom-3 left-1/2 -translate-x-1/2 text-[8px] text-editor-accent">
              {progress}%
            </span>
          )}
        </div>
      );
    default:
      return <Clock size={12} className="text-editor-muted" />;
  }
}
