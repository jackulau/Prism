import { useNavigate } from 'react-router-dom';
import { CreditCard, Zap, ArrowRight } from 'lucide-react';
import { trpc } from '../../lib/trpc';

export function SubscriptionInfo() {
  const navigate = useNavigate();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: subscription, isLoading } = (trpc as any).payment.getSubscription.useQuery();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: usage } = (trpc as any).payment.getUsage.useQuery();

  if (isLoading) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Subscription</h2>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-6 animate-pulse">
          <div className="h-6 bg-editor-border rounded w-1/3 mb-4" />
          <div className="h-4 bg-editor-border rounded w-1/2" />
        </div>
      </section>
    );
  }

  const isActive = subscription?.status === 'active';
  const planName = subscription?.planName || 'Free';

  const tokenUsed = usage?.tokenUsage?.used ?? 0;
  const tokenLimit = usage?.tokenUsage?.limit ?? 500000;
  const tokenPercent = tokenLimit > 0 ? Math.min((tokenUsed / tokenLimit) * 100, 100) : 0;

  const formatTokens = (value: number) => {
    if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(0)}K`;
    return value.toString();
  };

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold text-editor-text">Subscription</h2>

      <div className="bg-editor-surface border border-editor-border rounded-lg p-6">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-editor-accent/10 rounded-lg">
              <CreditCard className="w-6 h-6 text-editor-accent" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-lg font-medium text-editor-text">{planName} Plan</h3>
                {isActive && (
                  <span className="px-2 py-0.5 bg-editor-success/10 text-editor-success text-xs rounded-full">
                    Active
                  </span>
                )}
              </div>
              <p className="text-sm text-editor-muted">
                {isActive
                  ? `Next billing: ${
                      subscription?.nextInvoice?.dueDate
                        ? new Date(subscription.nextInvoice.dueDate).toLocaleDateString()
                        : 'N/A'
                    }`
                  : 'No active subscription'}
              </p>
            </div>
          </div>

          <button
            onClick={() => navigate('/usage')}
            className="flex items-center gap-2 px-3 py-2 text-sm text-editor-accent hover:bg-editor-accent/10 rounded-lg transition-colors"
          >
            Manage
            <ArrowRight size={16} />
          </button>
        </div>

        {/* Usage Mini Chart */}
        <div className="mt-6 pt-6 border-t border-editor-border">
          <div className="flex items-center gap-2 mb-3">
            <Zap size={16} className="text-editor-accent" />
            <span className="text-sm font-medium text-editor-text">Token Usage</span>
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
            <div className="flex items-center justify-between text-xs text-editor-muted">
              <span>{formatTokens(tokenUsed)} used</span>
              <span>{formatTokens(tokenLimit)} limit</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
