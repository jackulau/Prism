import { Activity, CheckCircle, Zap, DollarSign, Clock, AlertCircle } from 'lucide-react';
import { MetricCard, MetricCardSkeleton, type MetricCardVariant } from './MetricCard';
import type { AggregatedResults, ExecutionMetrics } from '../../types/results';

export interface ResultsMetricsSummaryProps {
  /** Aggregated results data */
  data?: AggregatedResults | null;
  /** Loading state */
  loading?: boolean;
  /** Error state */
  error?: Error | string | null;
  /** Additional CSS classes */
  className?: string;
}

function formatNumber(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(2)}M`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
  return value.toLocaleString();
}

function formatCurrency(value: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  }).format(value);
}

function formatDuration(ms?: number): string {
  if (ms === undefined || ms === null) return '--';
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

function getSuccessRateVariant(rate: number): MetricCardVariant {
  if (rate >= 90) return 'success';
  if (rate >= 70) return 'warning';
  return 'error';
}

function getExecutionCounts(data: AggregatedResults): {
  total: number;
  completed: number;
  failed: number;
  pending: number;
} {
  if (data.type === 'batch' && data.batch) {
    return {
      total: data.batch.totalCount,
      completed: data.batch.completedCount,
      failed: data.batch.failedCount,
      pending: data.batch.pendingCount,
    };
  }
  if (data.type === 'swarm' && data.swarm) {
    const agents = data.swarm.agentResults || [];
    return {
      total: agents.length,
      completed: agents.filter((a) => a.status === 'completed').length,
      failed: agents.filter((a) => a.status === 'failed').length,
      pending: agents.filter((a) => a.status === 'pending' || a.status === 'running').length,
    };
  }
  if (data.type === 'single' && data.execution) {
    const isComplete = data.execution.status === 'completed';
    const isFailed = data.execution.status === 'failed';
    return {
      total: 1,
      completed: isComplete ? 1 : 0,
      failed: isFailed ? 1 : 0,
      pending: !isComplete && !isFailed ? 1 : 0,
    };
  }
  return { total: 0, completed: 0, failed: 0, pending: 0 };
}

function getMetrics(data: AggregatedResults): ExecutionMetrics {
  return data.totalMetrics;
}

function getAverageDuration(data: AggregatedResults): number | undefined {
  if (data.type === 'batch' && data.batch) {
    const executions = data.batch.executions || [];
    const durations = executions
      .filter((e) => e.metrics?.durationMs !== undefined)
      .map((e) => e.metrics.durationMs!);
    if (durations.length === 0) return undefined;
    return durations.reduce((sum, d) => sum + d, 0) / durations.length;
  }
  if (data.type === 'swarm' && data.swarm) {
    const agents = data.swarm.agentResults || [];
    const durations = agents
      .filter((a) => a.metrics?.durationMs !== undefined)
      .map((a) => a.metrics.durationMs!);
    if (durations.length === 0) return undefined;
    return durations.reduce((sum, d) => sum + d, 0) / durations.length;
  }
  if (data.type === 'single' && data.execution) {
    return data.execution.metrics?.durationMs;
  }
  return undefined;
}

function LoadingState({ className = '' }: { className?: string }) {
  return (
    <div className={`grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 ${className}`}>
      {[1, 2, 3, 4, 5].map((i) => (
        <MetricCardSkeleton key={i} />
      ))}
    </div>
  );
}

function ErrorState({ error, className = '' }: { error: Error | string; className?: string }) {
  const message = typeof error === 'string' ? error : error.message;
  return (
    <div
      className={`bg-editor-error/10 border border-editor-error/30 rounded-lg p-6 ${className}`}
    >
      <div className="flex items-center gap-3">
        <div className="p-2 bg-editor-error/20 rounded-lg">
          <AlertCircle size={20} className="text-editor-error" />
        </div>
        <div>
          <h3 className="font-medium text-editor-error">Failed to load metrics</h3>
          <p className="text-sm text-editor-muted mt-1">{message}</p>
        </div>
      </div>
    </div>
  );
}

function EmptyState({ className = '' }: { className?: string }) {
  return (
    <div className={`bg-editor-surface border border-editor-border rounded-lg p-6 ${className}`}>
      <div className="flex items-center gap-3">
        <div className="p-2 bg-editor-muted/10 rounded-lg">
          <Activity size={20} className="text-editor-muted" />
        </div>
        <div>
          <h3 className="font-medium text-editor-text">No results yet</h3>
          <p className="text-sm text-editor-muted mt-1">
            Run an execution to see metrics here.
          </p>
        </div>
      </div>
    </div>
  );
}

export function ResultsMetricsSummary({
  data,
  loading = false,
  error = null,
  className = '',
}: ResultsMetricsSummaryProps) {
  if (loading) {
    return <LoadingState className={className} />;
  }

  if (error) {
    return <ErrorState error={error} className={className} />;
  }

  if (!data) {
    return <EmptyState className={className} />;
  }

  const counts = getExecutionCounts(data);
  const metrics = getMetrics(data);
  const avgDuration = getAverageDuration(data);
  const successRate = counts.total > 0 ? (counts.completed / counts.total) * 100 : 0;

  const typeLabel =
    data.type === 'swarm' ? 'Swarm' : data.type === 'batch' ? 'Batch' : 'Single';

  return (
    <div className={`grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 ${className}`}>
      {/* Total Executions */}
      <MetricCard
        icon={Activity}
        title="Executions"
        value={formatNumber(counts.total)}
        secondaryValue={`${typeLabel}: ${counts.completed} completed, ${counts.failed} failed`}
        variant="default"
      />

      {/* Success Rate */}
      <MetricCard
        icon={CheckCircle}
        title="Success Rate"
        value={`${successRate.toFixed(1)}%`}
        secondaryValue={`${counts.completed} of ${counts.total} succeeded`}
        variant={getSuccessRateVariant(successRate)}
      />

      {/* Total Tokens */}
      <MetricCard
        icon={Zap}
        title="Total Tokens"
        value={formatNumber(metrics.totalTokens)}
        secondaryValue={`${formatNumber(metrics.promptTokens)} in / ${formatNumber(metrics.completionTokens)} out`}
        variant="default"
      />

      {/* Total Cost */}
      <MetricCard
        icon={DollarSign}
        title="Total Cost"
        value={formatCurrency(metrics.totalCost, metrics.currency)}
        secondaryValue={`Input: ${formatCurrency(metrics.inputCost, metrics.currency)} / Output: ${formatCurrency(metrics.outputCost, metrics.currency)}`}
        variant="default"
      />

      {/* Average Duration */}
      <MetricCard
        icon={Clock}
        title="Avg Duration"
        value={formatDuration(avgDuration)}
        secondaryValue={
          metrics.tokensPerSecond
            ? `${metrics.tokensPerSecond.toFixed(1)} tokens/sec`
            : undefined
        }
        variant="default"
      />
    </div>
  );
}

export default ResultsMetricsSummary;
