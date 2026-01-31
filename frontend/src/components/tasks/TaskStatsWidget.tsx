/**
 * Task Statistics Widget Component.
 * Displays task queue metrics in a grid of stat cards.
 */

import { useState } from 'react';
import {
  ListTodo,
  Clock,
  Play,
  CheckCircle,
  XCircle,
  Percent,
  RefreshCw,
  AlertCircle,
} from 'lucide-react';
import { useTaskStats } from '../../hooks/useTaskStats';

interface TaskStatsWidgetProps {
  /** Enable auto-polling for live updates */
  enablePolling?: boolean;
  /** Polling interval in milliseconds (default: 5000) */
  pollingInterval?: number;
  /** Optional className for the container */
  className?: string;
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subtitle?: string;
  colorClass: string;
  isLive?: boolean;
}

function StatCard({ icon, label, value, subtitle, colorClass, isLive }: StatCardProps) {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-center gap-3 mb-2">
        <div className={`p-2 rounded-lg ${colorClass.replace('text-', 'bg-')}/10`}>
          {icon}
        </div>
        <span className="text-sm text-editor-muted">{label}</span>
        {isLive && (
          <span className="ml-auto flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-editor-success animate-pulse" />
            <span className="text-xs text-editor-muted">Live</span>
          </span>
        )}
      </div>
      <div className="text-2xl font-bold text-editor-text">{value}</div>
      {subtitle && <div className="text-xs text-editor-muted mt-1">{subtitle}</div>}
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
      {[1, 2, 3, 4, 5].map((i) => (
        <div
          key={i}
          className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="w-9 h-9 bg-editor-border rounded-lg" />
            <div className="h-4 bg-editor-border rounded w-16" />
          </div>
          <div className="h-8 bg-editor-border rounded w-12 mb-1" />
          <div className="h-3 bg-editor-border rounded w-20" />
        </div>
      ))}
    </div>
  );
}

interface ErrorStateProps {
  error: string;
  onRetry: () => void;
}

function ErrorState({ error, onRetry }: ErrorStateProps) {
  return (
    <div className="bg-editor-surface border border-editor-error/30 rounded-lg p-6">
      <div className="flex flex-col items-center text-center gap-4">
        <div className="p-3 rounded-full bg-editor-error/10">
          <AlertCircle className="w-8 h-8 text-editor-error" />
        </div>
        <div>
          <h3 className="text-lg font-medium text-editor-text mb-1">
            Failed to load task statistics
          </h3>
          <p className="text-sm text-editor-muted">{error}</p>
        </div>
        <button
          onClick={onRetry}
          className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
        >
          <RefreshCw className="w-4 h-4" />
          Retry
        </button>
      </div>
    </div>
  );
}

export function TaskStatsWidget({
  enablePolling = false,
  pollingInterval = 5000,
  className = '',
}: TaskStatsWidgetProps) {
  const [pollingEnabled, setPollingEnabled] = useState(enablePolling);

  const {
    stats,
    isLoading,
    error,
    refetch,
    isPolling,
    startPolling,
    stopPolling,
  } = useTaskStats({
    polling: pollingEnabled,
    pollingInterval,
  });

  const togglePolling = () => {
    if (isPolling) {
      stopPolling();
      setPollingEnabled(false);
    } else {
      startPolling();
      setPollingEnabled(true);
    }
  };

  if (isLoading && !stats) {
    return (
      <section className={`space-y-4 ${className}`}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-editor-text">Task Queue</h2>
        </div>
        <LoadingSkeleton />
      </section>
    );
  }

  if (error && !stats) {
    return (
      <section className={`space-y-4 ${className}`}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-editor-text">Task Queue</h2>
        </div>
        <ErrorState error={error} onRetry={refetch} />
      </section>
    );
  }

  // Calculate success rate
  const totalFinished = (stats?.completed ?? 0) + (stats?.failed ?? 0);
  const successRate = totalFinished > 0
    ? Math.round((stats?.completed ?? 0) / totalFinished * 100)
    : 0;

  return (
    <section className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Task Queue</h2>
        <div className="flex items-center gap-2">
          <button
            onClick={togglePolling}
            className={`flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg transition-colors ${
              isPolling
                ? 'bg-editor-success/10 text-editor-success'
                : 'bg-editor-surface text-editor-muted hover:text-editor-text'
            }`}
            title={isPolling ? 'Stop live updates' : 'Enable live updates'}
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isPolling ? 'animate-spin' : ''}`} />
            {isPolling ? 'Live' : 'Auto-refresh'}
          </button>
          {!isPolling && (
            <button
              onClick={refetch}
              className="flex items-center gap-1.5 text-sm text-editor-muted hover:text-editor-text px-3 py-1.5 rounded-lg bg-editor-surface transition-colors"
              title="Refresh stats"
            >
              <RefreshCw className="w-3.5 h-3.5" />
              Refresh
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
        <StatCard
          icon={<ListTodo className="w-5 h-5 text-editor-text" />}
          label="Total Tasks"
          value={stats?.total ?? 0}
          colorClass="text-editor-text"
        />
        <StatCard
          icon={<Clock className="w-5 h-5 text-editor-warning" />}
          label="Pending"
          value={stats?.pending ?? 0}
          subtitle="Queue depth"
          colorClass="text-editor-warning"
        />
        <StatCard
          icon={<Play className="w-5 h-5 text-editor-accent" />}
          label="Running"
          value={stats?.running ?? 0}
          colorClass="text-editor-accent"
          isLive={isPolling && (stats?.running ?? 0) > 0}
        />
        <StatCard
          icon={<CheckCircle className="w-5 h-5 text-editor-success" />}
          label="Completed"
          value={stats?.completed ?? 0}
          colorClass="text-editor-success"
        />
        <StatCard
          icon={<XCircle className="w-5 h-5 text-editor-error" />}
          label="Failed"
          value={stats?.failed ?? 0}
          colorClass="text-editor-error"
        />
      </div>

      {/* Success Rate Bar */}
      <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
        <div className="flex items-center gap-3 mb-3">
          <div className="p-2 rounded-lg bg-editor-success/10">
            <Percent className="w-5 h-5 text-editor-success" />
          </div>
          <span className="text-sm text-editor-muted">Success Rate</span>
          <span className="ml-auto text-lg font-bold text-editor-text">{successRate}%</span>
        </div>
        <div className="h-2 bg-editor-bg rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${
              successRate >= 90
                ? 'bg-editor-success'
                : successRate >= 70
                ? 'bg-editor-warning'
                : 'bg-editor-error'
            }`}
            style={{ width: `${successRate}%` }}
          />
        </div>
        <div className="flex justify-between text-xs text-editor-muted mt-2">
          <span>{stats?.completed ?? 0} completed</span>
          <span>{stats?.failed ?? 0} failed</span>
        </div>
      </div>
    </section>
  );
}
