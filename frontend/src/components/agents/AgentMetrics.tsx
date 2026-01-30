import {
  Hash,
  DollarSign,
  Clock,
  Cpu,
  Zap,
  Timer,
  TrendingUp,
  Activity,
  BarChart3,
} from 'lucide-react';
import type { AgentExecutionMetrics } from './AgentDetail';

interface AgentMetricsProps {
  metrics?: AgentExecutionMetrics;
  model: string;
  provider: string;
  startedAt?: Date;
  completedAt?: Date;
}

interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  unit?: string;
  highlight?: boolean;
  subValue?: string;
}

function MetricCard({
  icon,
  label,
  value,
  unit,
  highlight,
  subValue,
}: MetricCardProps) {
  return (
    <div
      className={`p-4 rounded-lg border transition-all ${
        highlight
          ? 'bg-editor-accent/10 border-editor-accent/30'
          : 'bg-editor-surface border-editor-border'
      }`}
    >
      <div className="flex items-center gap-2 mb-2">
        <span className={highlight ? 'text-editor-accent' : 'text-editor-muted'}>
          {icon}
        </span>
        <span className="text-xs text-editor-muted uppercase tracking-wide">
          {label}
        </span>
      </div>
      <div className="flex items-baseline gap-1">
        <span
          className={`text-2xl font-semibold ${
            highlight ? 'text-editor-accent' : 'text-editor-text'
          }`}
        >
          {value}
        </span>
        {unit && <span className="text-sm text-editor-muted">{unit}</span>}
      </div>
      {subValue && (
        <div className="mt-1 text-xs text-editor-muted">{subValue}</div>
      )}
    </div>
  );
}

interface ProgressBarProps {
  value: number;
  max: number;
  label: string;
  color?: string;
}

function ProgressBar({
  value,
  max,
  label,
  color = 'bg-editor-accent',
}: ProgressBarProps) {
  const percentage = max > 0 ? Math.min((value / max) * 100, 100) : 0;

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs">
        <span className="text-editor-muted">{label}</span>
        <span className="text-editor-text">
          {value.toLocaleString()} / {max.toLocaleString()}
        </span>
      </div>
      <div className="h-2 bg-editor-surface rounded-full overflow-hidden">
        <div
          className={`h-full ${color} transition-all duration-300 rounded-full`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = ((ms % 60000) / 1000).toFixed(1);
  return `${minutes}m ${seconds}s`;
}

function formatCost(cost: number): string {
  if (cost < 0.01) return `$${cost.toFixed(6)}`;
  if (cost < 1) return `$${cost.toFixed(4)}`;
  return `$${cost.toFixed(2)}`;
}

export function AgentMetrics({
  metrics,
  model,
  provider,
  startedAt,
  completedAt,
}: AgentMetricsProps) {
  const duration =
    metrics?.duration ||
    (startedAt && completedAt
      ? new Date(completedAt).getTime() - new Date(startedAt).getTime()
      : null);

  const tokensPerSecond =
    duration && metrics?.totalTokens
      ? (metrics.totalTokens / (duration / 1000)).toFixed(1)
      : null;

  const estimatedMaxTokens = 4096;

  return (
    <div className="space-y-6">
      {/* Token Usage */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <Hash size={16} className="text-editor-accent" />
          Token Usage
        </h3>
        <div className="grid grid-cols-3 gap-4">
          <MetricCard
            icon={<TrendingUp size={16} />}
            label="Prompt Tokens"
            value={metrics?.promptTokens?.toLocaleString() || '--'}
            subValue="Input tokens"
          />
          <MetricCard
            icon={<Activity size={16} />}
            label="Completion Tokens"
            value={metrics?.completionTokens?.toLocaleString() || '--'}
            subValue="Output tokens"
          />
          <MetricCard
            icon={<Hash size={16} />}
            label="Total Tokens"
            value={metrics?.totalTokens?.toLocaleString() || '--'}
            highlight={!!metrics?.totalTokens}
            subValue={
              tokensPerSecond ? `${tokensPerSecond} tokens/sec` : undefined
            }
          />
        </div>

        {/* Token progress bar */}
        {metrics?.totalTokens && (
          <ProgressBar
            value={metrics.totalTokens}
            max={estimatedMaxTokens}
            label="Token Usage (estimated max)"
            color="bg-editor-accent"
          />
        )}
      </div>

      {/* Cost Breakdown */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <DollarSign size={16} className="text-green-400" />
          Cost Breakdown
        </h3>
        <div className="grid grid-cols-3 gap-4">
          <MetricCard
            icon={<TrendingUp size={16} />}
            label="Input Cost"
            value={
              metrics?.inputCost !== undefined
                ? formatCost(metrics.inputCost)
                : '--'
            }
            subValue="Prompt tokens"
          />
          <MetricCard
            icon={<Activity size={16} />}
            label="Output Cost"
            value={
              metrics?.outputCost !== undefined
                ? formatCost(metrics.outputCost)
                : '--'
            }
            subValue="Completion tokens"
          />
          <MetricCard
            icon={<DollarSign size={16} />}
            label="Total Cost"
            value={
              metrics?.totalCost !== undefined
                ? formatCost(metrics.totalCost)
                : '--'
            }
            highlight={metrics?.totalCost !== undefined}
          />
        </div>
      </div>

      {/* Timing Metrics */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <Clock size={16} className="text-blue-400" />
          Timing
        </h3>
        <div className="grid grid-cols-3 gap-4">
          <MetricCard
            icon={<Timer size={16} />}
            label="Duration"
            value={duration ? formatDuration(duration) : '--'}
            highlight={!!duration}
          />
          <MetricCard
            icon={<Zap size={16} />}
            label="Tokens/sec"
            value={tokensPerSecond || '--'}
            unit="t/s"
          />
          <MetricCard
            icon={<BarChart3 size={16} />}
            label="Iterations"
            value={metrics?.iterationCount?.toString() || '--'}
            subValue="Agent loops"
          />
        </div>
      </div>

      {/* Model & Provider Info */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-editor-text flex items-center gap-2">
          <Cpu size={16} className="text-purple-400" />
          Model Information
        </h3>
        <div className="bg-editor-surface rounded-lg p-4 space-y-3 font-mono text-sm">
          <div className="flex justify-between">
            <span className="text-editor-muted">Provider</span>
            <span className="text-editor-text">{provider}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-editor-muted">Model</span>
            <span className="text-editor-text">{model}</span>
          </div>
          <div className="border-t border-editor-border my-2" />
          <div className="flex justify-between">
            <span className="text-editor-muted">Started At</span>
            <span className="text-editor-text">
              {startedAt
                ? new Date(startedAt).toLocaleString()
                : '--'}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-editor-muted">Completed At</span>
            <span className="text-editor-text">
              {completedAt
                ? new Date(completedAt).toLocaleString()
                : '--'}
            </span>
          </div>
        </div>
      </div>

      {/* No metrics placeholder */}
      {!metrics && (
        <div className="bg-editor-surface/50 rounded-lg p-6 text-center">
          <BarChart3 className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">
            No detailed metrics available
          </h3>
          <p className="text-sm text-editor-muted max-w-md mx-auto">
            Detailed metrics will be available once the agent execution completes.
          </p>
        </div>
      )}
    </div>
  );
}

export default AgentMetrics;
