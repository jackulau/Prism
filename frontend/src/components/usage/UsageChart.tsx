import { Zap, Clock } from 'lucide-react';
import { trpc } from '../../lib/trpc';

export function UsageChart() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: usage, isLoading } = (trpc as any).payment.getUsage.useQuery();

  if (isLoading) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Current Usage</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
            >
              <div className="h-4 bg-editor-border rounded w-1/3 mb-4" />
              <div className="h-3 bg-editor-border rounded w-full mb-2" />
              <div className="h-3 bg-editor-border rounded w-1/2" />
            </div>
          ))}
        </div>
      </section>
    );
  }

  const tokenUsed = usage?.tokenUsage?.used ?? 0;
  const tokenLimit = usage?.tokenUsage?.limit ?? 500;
  const tokenPercent = tokenLimit > 0 ? Math.min((tokenUsed / tokenLimit) * 100, 100) : 0;

  const sandboxUsed = usage?.sandboxUsage?.used ?? 0;
  const sandboxLimit = usage?.sandboxUsage?.limit ?? 2;
  const sandboxPercent = sandboxLimit > 0 ? Math.min((sandboxUsed / sandboxLimit) * 100, 100) : 0;

  const formatTokens = (value: number) => {
    if (value >= 1000000) return `${(value / 1000000).toFixed(2)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
    return value.toString();
  };

  const formatHours = (hours: number) => {
    if (hours >= 1) return `${hours.toFixed(1)} hours`;
    return `${Math.round(hours * 60)} minutes`;
  };

  const getProgressColor = (percent: number) => {
    if (percent >= 90) return 'bg-editor-error';
    if (percent >= 70) return 'bg-editor-warning';
    return 'bg-editor-accent';
  };

  const metrics = [
    {
      icon: Zap,
      label: 'Token Usage',
      used: tokenUsed,
      limit: tokenLimit,
      percent: tokenPercent,
      format: formatTokens,
      color: 'text-editor-accent',
      bgColor: 'bg-editor-accent/10',
    },
    {
      icon: Clock,
      label: 'Sandbox Hours',
      used: sandboxUsed,
      limit: sandboxLimit,
      percent: sandboxPercent,
      format: formatHours,
      color: 'text-editor-success',
      bgColor: 'bg-editor-success/10',
    },
  ];

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold text-editor-text">Current Usage</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {metrics.map((metric) => {
          const Icon = metric.icon;
          return (
            <div
              key={metric.label}
              className="bg-editor-surface border border-editor-border rounded-lg p-6"
            >
              <div className="flex items-center gap-3 mb-4">
                <div className={`p-2 rounded-lg ${metric.bgColor}`}>
                  <Icon size={20} className={metric.color} />
                </div>
                <span className="font-medium text-editor-text">{metric.label}</span>
              </div>

              {/* Progress Bar */}
              <div className="space-y-2">
                <div className="h-3 bg-editor-bg rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${getProgressColor(metric.percent)}`}
                    style={{ width: `${metric.percent}%` }}
                  />
                </div>

                <div className="flex items-center justify-between text-sm">
                  <span className="text-editor-muted">
                    {metric.format(metric.used)} / {metric.format(metric.limit)}
                  </span>
                  <span
                    className={`font-medium ${
                      metric.percent >= 90
                        ? 'text-editor-error'
                        : metric.percent >= 70
                        ? 'text-editor-warning'
                        : 'text-editor-text'
                    }`}
                  >
                    {metric.percent.toFixed(0)}%
                  </span>
                </div>
              </div>

              {/* Warning if high usage */}
              {metric.percent >= 80 && (
                <div
                  className={`mt-4 p-3 rounded-lg ${
                    metric.percent >= 90
                      ? 'bg-editor-error/10 border border-editor-error/20'
                      : 'bg-editor-warning/10 border border-editor-warning/20'
                  }`}
                >
                  <p
                    className={`text-sm ${
                      metric.percent >= 90 ? 'text-editor-error' : 'text-editor-warning'
                    }`}
                  >
                    {metric.percent >= 90
                      ? 'Usage limit almost reached. Consider upgrading your plan.'
                      : 'Approaching usage limit.'}
                  </p>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </section>
  );
}
