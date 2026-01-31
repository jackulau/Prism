import React from 'react';
import {
  Cpu,
  DollarSign,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  Zap,
} from 'lucide-react';

export interface ExecutionMetricsData {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  maxTokens?: number; // For progress bar
  inputCost?: number;
  outputCost?: number;
  totalCost?: number;
  successCount: number;
  failureCount: number;
  totalTasks: number;
  averageDuration?: number; // milliseconds
  fastestDuration?: number;
  slowestDuration?: number;
}

interface ExecutionMetricsProps {
  metrics: ExecutionMetricsData;
  className?: string;
}

interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subValue?: string;
  highlight?: boolean;
  trend?: 'up' | 'down' | 'neutral';
}

const MetricCard: React.FC<MetricCardProps> = ({
  icon,
  label,
  value,
  subValue,
  highlight,
  trend,
}) => (
  <div
    className={`p-4 rounded-lg border transition-all ${
      highlight
        ? 'bg-editor-accent/10 border-editor-accent/30'
        : 'bg-editor-bg border-editor-border'
    }`}
  >
    <div className="flex items-center gap-2 mb-2">
      <span className={highlight ? 'text-editor-accent' : 'text-editor-muted'}>
        {icon}
      </span>
      <span className="text-xs text-editor-muted uppercase tracking-wide">{label}</span>
      {trend && trend !== 'neutral' && (
        <TrendingUp
          className={`w-3 h-3 ${
            trend === 'up' ? 'text-editor-success' : 'text-editor-error rotate-180'
          }`}
        />
      )}
    </div>
    <div className="flex items-baseline gap-1">
      <span
        className={`text-2xl font-semibold ${
          highlight ? 'text-editor-accent' : 'text-editor-text'
        }`}
      >
        {value}
      </span>
    </div>
    {subValue && (
      <div className="mt-1 text-xs text-editor-muted">{subValue}</div>
    )}
  </div>
);

interface ProgressBarProps {
  value: number;
  max: number;
  label: string;
  showPercentage?: boolean;
  color?: string;
}

const ProgressBar: React.FC<ProgressBarProps> = ({
  value,
  max,
  label,
  showPercentage = true,
  color = 'bg-editor-accent',
}) => {
  const percentage = max > 0 ? Math.min((value / max) * 100, 100) : 0;
  const isWarning = percentage > 80;
  const barColor = isWarning ? 'bg-editor-warning' : color;

  return (
    <div className="space-y-1.5">
      <div className="flex justify-between text-xs">
        <span className="text-editor-muted">{label}</span>
        <span className="text-editor-text">
          {value.toLocaleString()} / {max.toLocaleString()}
          {showPercentage && (
            <span className="text-editor-muted ml-1">({percentage.toFixed(0)}%)</span>
          )}
        </span>
      </div>
      <div className="h-2 bg-editor-bg rounded-full overflow-hidden">
        <div
          className={`h-full ${barColor} transition-all duration-300 rounded-full`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
};

function formatDuration(ms?: number): string {
  if (ms === undefined) return '--';
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

function formatCost(cost?: number): string {
  if (cost === undefined || cost === 0) return '--';
  if (cost < 0.01) return `$${cost.toFixed(4)}`;
  return `$${cost.toFixed(2)}`;
}

export const ExecutionMetrics: React.FC<ExecutionMetricsProps> = ({
  metrics,
  className = '',
}) => {
  const successRate = metrics.totalTasks > 0
    ? ((metrics.successCount / metrics.totalTasks) * 100).toFixed(0)
    : '0';

  return (
    <div className={`bg-editor-surface border border-editor-border rounded-lg p-4 ${className}`}>
      {/* Header */}
      <h3 className="text-sm font-medium text-editor-text flex items-center gap-2 mb-4">
        <Zap className="w-4 h-4 text-editor-accent" />
        Execution Metrics
      </h3>

      {/* Primary metrics grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
        <MetricCard
          icon={<Cpu className="w-4 h-4" />}
          label="Total Tokens"
          value={metrics.totalTokens.toLocaleString()}
          subValue={`${metrics.promptTokens.toLocaleString()} prompt + ${metrics.completionTokens.toLocaleString()} completion`}
          highlight
        />
        <MetricCard
          icon={<DollarSign className="w-4 h-4" />}
          label="Total Cost"
          value={formatCost(metrics.totalCost)}
          subValue={metrics.inputCost !== undefined && metrics.outputCost !== undefined
            ? `${formatCost(metrics.inputCost)} input + ${formatCost(metrics.outputCost)} output`
            : undefined}
        />
        <MetricCard
          icon={<CheckCircle className="w-4 h-4" />}
          label="Success Rate"
          value={`${successRate}%`}
          subValue={`${metrics.successCount} of ${metrics.totalTasks} tasks`}
          trend={Number(successRate) >= 90 ? 'up' : Number(successRate) < 50 ? 'down' : 'neutral'}
        />
        <MetricCard
          icon={<Clock className="w-4 h-4" />}
          label="Avg Duration"
          value={formatDuration(metrics.averageDuration)}
          subValue={metrics.fastestDuration !== undefined && metrics.slowestDuration !== undefined
            ? `${formatDuration(metrics.fastestDuration)} - ${formatDuration(metrics.slowestDuration)}`
            : undefined}
        />
      </div>

      {/* Token usage progress */}
      {metrics.maxTokens && (
        <div className="mb-4">
          <ProgressBar
            value={metrics.totalTokens}
            max={metrics.maxTokens}
            label="Token Limit Usage"
          />
        </div>
      )}

      {/* Success/Failure breakdown */}
      <div className="bg-editor-bg rounded-lg p-3">
        <h4 className="text-xs text-editor-muted uppercase tracking-wide mb-3">
          Task Breakdown
        </h4>
        <div className="flex items-center gap-4">
          {/* Success bar segment */}
          <div className="flex-1">
            <div className="flex items-center justify-between text-xs mb-1">
              <span className="flex items-center gap-1 text-editor-success">
                <CheckCircle className="w-3 h-3" />
                Success
              </span>
              <span className="text-editor-text">{metrics.successCount}</span>
            </div>
            <div className="h-3 bg-editor-surface rounded-full overflow-hidden">
              <div
                className="h-full bg-editor-success transition-all duration-300 rounded-full"
                style={{
                  width: `${metrics.totalTasks > 0 ? (metrics.successCount / metrics.totalTasks) * 100 : 0}%`,
                }}
              />
            </div>
          </div>

          {/* Failure bar segment */}
          <div className="flex-1">
            <div className="flex items-center justify-between text-xs mb-1">
              <span className="flex items-center gap-1 text-editor-error">
                <XCircle className="w-3 h-3" />
                Failed
              </span>
              <span className="text-editor-text">{metrics.failureCount}</span>
            </div>
            <div className="h-3 bg-editor-surface rounded-full overflow-hidden">
              <div
                className="h-full bg-editor-error transition-all duration-300 rounded-full"
                style={{
                  width: `${metrics.totalTasks > 0 ? (metrics.failureCount / metrics.totalTasks) * 100 : 0}%`,
                }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Token breakdown details */}
      <div className="mt-4 grid grid-cols-2 gap-4 text-xs">
        <div className="bg-editor-bg rounded-lg p-3 space-y-2">
          <h4 className="text-editor-muted uppercase tracking-wide">Input Tokens</h4>
          <div className="flex justify-between">
            <span className="text-editor-muted">Prompt</span>
            <span className="text-editor-text font-mono">{metrics.promptTokens.toLocaleString()}</span>
          </div>
          {metrics.inputCost !== undefined && (
            <div className="flex justify-between">
              <span className="text-editor-muted">Cost</span>
              <span className="text-editor-text font-mono">{formatCost(metrics.inputCost)}</span>
            </div>
          )}
        </div>
        <div className="bg-editor-bg rounded-lg p-3 space-y-2">
          <h4 className="text-editor-muted uppercase tracking-wide">Output Tokens</h4>
          <div className="flex justify-between">
            <span className="text-editor-muted">Completion</span>
            <span className="text-editor-text font-mono">{metrics.completionTokens.toLocaleString()}</span>
          </div>
          {metrics.outputCost !== undefined && (
            <div className="flex justify-between">
              <span className="text-editor-muted">Cost</span>
              <span className="text-editor-text font-mono">{formatCost(metrics.outputCost)}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
