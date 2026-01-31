import { Clock, Loader2, CheckCircle, XCircle, ListTodo } from 'lucide-react';
import type { TaskStats } from '../../hooks/useTasks';

interface TaskStatsWidgetProps {
  stats: TaskStats | null;
  isLoading: boolean;
  error?: string | null;
}

interface StatCardProps {
  icon: typeof Clock;
  iconColor: string;
  label: string;
  value: number;
  isLoading: boolean;
}

function StatCard({ icon: Icon, iconColor, label, value, isLoading }: StatCardProps) {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-center gap-3">
        <div className={`p-2 rounded-lg ${iconColor}`}>
          <Icon size={20} />
        </div>
        <div>
          <p className="text-xs text-editor-muted uppercase tracking-wide">{label}</p>
          {isLoading ? (
            <div className="h-6 w-12 bg-editor-bg rounded animate-pulse mt-1" />
          ) : (
            <p className="text-2xl font-semibold text-editor-text">{value}</p>
          )}
        </div>
      </div>
    </div>
  );
}

export function TaskStatsWidget({ stats, isLoading, error }: TaskStatsWidgetProps) {
  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-red-400">
        Failed to load task statistics: {error}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
      <StatCard
        icon={ListTodo}
        iconColor="bg-editor-accent/10 text-editor-accent"
        label="Total"
        value={stats?.total ?? 0}
        isLoading={isLoading}
      />
      <StatCard
        icon={Clock}
        iconColor="bg-yellow-500/10 text-yellow-500"
        label="Pending"
        value={stats?.pending ?? 0}
        isLoading={isLoading}
      />
      <StatCard
        icon={Loader2}
        iconColor="bg-blue-500/10 text-blue-500"
        label="Running"
        value={stats?.running ?? 0}
        isLoading={isLoading}
      />
      <StatCard
        icon={CheckCircle}
        iconColor="bg-green-500/10 text-green-500"
        label="Completed"
        value={stats?.completed ?? 0}
        isLoading={isLoading}
      />
      <StatCard
        icon={XCircle}
        iconColor="bg-red-500/10 text-red-500"
        label="Failed"
        value={stats?.failed ?? 0}
        isLoading={isLoading}
      />
    </div>
  );
}
