import { useNavigate } from 'react-router-dom';
import { Zap, Clock, ArrowUpRight } from 'lucide-react';
import { trpc } from '../../lib/trpc';

export function UsageSummary() {
  const navigate = useNavigate();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: usage, isLoading } = (trpc as any).payment.getUsage.useQuery();

  if (isLoading) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Usage This Period</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
            >
              <div className="h-4 bg-editor-border rounded w-1/3 mb-4" />
              <div className="h-2 bg-editor-border rounded w-full mb-2" />
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
    if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
    return value.toString();
  };

  const formatHours = (hours: number) => {
    if (hours >= 1) return `${hours.toFixed(1)}h`;
    return `${Math.round(hours * 60)}m`;
  };

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Usage This Period</h2>
        <button
          onClick={() => navigate('/usage')}
          className="flex items-center gap-1 text-sm text-editor-accent hover:text-editor-accent/80 transition-colors"
        >
          View Details
          <ArrowUpRight size={14} />
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Token Usage */}
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <div className="flex items-center gap-2 mb-3">
            <div className="p-1.5 bg-editor-accent/10 rounded-lg">
              <Zap size={16} className="text-editor-accent" />
            </div>
            <span className="font-medium text-editor-text">Token Usage</span>
          </div>
          <div className="space-y-2">
            <div className="h-2 bg-editor-bg rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${
                  tokenPercent >= 90
                    ? 'bg-editor-error'
                    : tokenPercent >= 70
                    ? 'bg-editor-warning'
                    : 'bg-editor-accent'
                }`}
                style={{ width: `${tokenPercent}%` }}
              />
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-editor-muted">
                {formatTokens(tokenUsed)} / {formatTokens(tokenLimit)} tokens
              </span>
              <span
                className={`font-medium ${
                  tokenPercent >= 90
                    ? 'text-editor-error'
                    : tokenPercent >= 70
                    ? 'text-editor-warning'
                    : 'text-editor-text'
                }`}
              >
                {tokenPercent.toFixed(0)}%
              </span>
            </div>
          </div>
        </div>

        {/* Sandbox Hours */}
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <div className="flex items-center gap-2 mb-3">
            <div className="p-1.5 bg-editor-success/10 rounded-lg">
              <Clock size={16} className="text-editor-success" />
            </div>
            <span className="font-medium text-editor-text">Sandbox Hours</span>
          </div>
          <div className="space-y-2">
            <div className="h-2 bg-editor-bg rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${
                  sandboxPercent >= 90
                    ? 'bg-editor-error'
                    : sandboxPercent >= 70
                    ? 'bg-editor-warning'
                    : 'bg-editor-success'
                }`}
                style={{ width: `${sandboxPercent}%` }}
              />
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-editor-muted">
                {formatHours(sandboxUsed)} / {formatHours(sandboxLimit)} hours
              </span>
              <span
                className={`font-medium ${
                  sandboxPercent >= 90
                    ? 'text-editor-error'
                    : sandboxPercent >= 70
                    ? 'text-editor-warning'
                    : 'text-editor-text'
                }`}
              >
                {sandboxPercent.toFixed(0)}%
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
