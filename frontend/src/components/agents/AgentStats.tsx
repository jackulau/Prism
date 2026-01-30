import {
  Bot,
  Loader2,
  CheckCircle,
  XCircle,
  Zap,
  DollarSign,
} from 'lucide-react';
import type { AgentExecution } from './types';

interface AgentStatsProps {
  agents: AgentExecution[];
  isLoading?: boolean;
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subValue?: string;
  color: string;
  bgColor: string;
  isLive?: boolean;
}

function StatCard({ icon, label, value, subValue, color, bgColor, isLive }: StatCardProps) {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-center gap-3">
        <div className={`p-2 rounded-lg ${bgColor}`}>
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`text-2xl font-bold ${color}`}>{value}</span>
            {isLive && (
              <span className="flex items-center gap-1 text-xs text-green-400">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
                </span>
                Live
              </span>
            )}
          </div>
          <p className="text-sm text-editor-muted">{label}</p>
          {subValue && (
            <p className="text-xs text-editor-muted/70 mt-0.5">{subValue}</p>
          )}
        </div>
      </div>
    </div>
  );
}

function StatCardSkeleton() {
  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-editor-border" />
        <div className="flex-1">
          <div className="h-7 w-16 bg-editor-border rounded mb-1" />
          <div className="h-4 w-24 bg-editor-border rounded" />
        </div>
      </div>
    </div>
  );
}

function formatTokens(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(2)}M`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
  return value.toString();
}

function formatCost(value: number): string {
  if (value < 0.01) return `$${value.toFixed(4)}`;
  if (value < 1) return `$${value.toFixed(3)}`;
  return `$${value.toFixed(2)}`;
}

export function AgentStats({ agents, isLoading = false }: AgentStatsProps) {
  if (isLoading) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Overview</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <StatCardSkeleton key={i} />
          ))}
        </div>
      </section>
    );
  }

  // Calculate statistics
  const total = agents.length;
  const running = agents.filter(a => a.status === 'running').length;
  const completed = agents.filter(a => a.status === 'completed').length;
  const failed = agents.filter(a => a.status === 'failed').length;
  const pending = agents.filter(a => a.status === 'pending').length;

  const totalTokens = agents.reduce((sum, a) => sum + (a.total_tokens || 0), 0);
  const totalCost = agents.reduce((sum, a) => sum + (a.cost || 0), 0);

  const stats: StatCardProps[] = [
    {
      icon: <Bot size={20} className="text-editor-accent" />,
      label: 'Total Executions',
      value: total,
      color: 'text-editor-text',
      bgColor: 'bg-editor-accent/10',
    },
    {
      icon: <Loader2 size={20} className="text-blue-400 animate-spin" />,
      label: 'Running',
      value: running,
      subValue: pending > 0 ? `${pending} pending` : undefined,
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      isLive: running > 0,
    },
    {
      icon: <CheckCircle size={20} className="text-green-400" />,
      label: 'Completed',
      value: completed,
      subValue: total > 0 ? `${((completed / total) * 100).toFixed(0)}% success` : undefined,
      color: 'text-green-400',
      bgColor: 'bg-green-500/10',
    },
    {
      icon: <XCircle size={20} className="text-red-400" />,
      label: 'Failed',
      value: failed,
      subValue: total > 0 ? `${((failed / total) * 100).toFixed(0)}% failure` : undefined,
      color: 'text-red-400',
      bgColor: 'bg-red-500/10',
    },
    {
      icon: <Zap size={20} className="text-yellow-400" />,
      label: 'Total Tokens',
      value: formatTokens(totalTokens),
      color: 'text-yellow-400',
      bgColor: 'bg-yellow-500/10',
    },
    {
      icon: <DollarSign size={20} className="text-emerald-400" />,
      label: 'Total Cost',
      value: formatCost(totalCost),
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
    },
  ];

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold text-editor-text">Overview</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>
    </section>
  );
}

export default AgentStats;
