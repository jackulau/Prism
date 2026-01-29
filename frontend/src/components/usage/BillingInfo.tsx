import { CreditCard, Calendar, ArrowUpRight, Loader2 } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { toast } from '../../store/toastStore';

export function BillingInfo() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: subscription, isLoading } = (trpc as any).payment.getSubscription.useQuery();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: usage } = (trpc as any).payment.getUsage.useQuery();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const managePortal = (trpc as any).payment.createPortalSession.useMutation({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onSuccess: (data: any) => {
      if (data.url) {
        window.location.href = data.url;
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return (
      <section className="space-y-4">
        <div className="bg-editor-surface border border-editor-border rounded-lg p-6 animate-pulse">
          <div className="h-6 bg-editor-border rounded w-1/3 mb-4" />
          <div className="h-4 bg-editor-border rounded w-1/2 mb-2" />
          <div className="h-4 bg-editor-border rounded w-1/4" />
        </div>
      </section>
    );
  }

  const billingPeriodStart = usage?.billingPeriod?.start
    ? new Date(usage.billingPeriod.start)
    : null;
  const billingPeriodEnd = usage?.billingPeriod?.end
    ? new Date(usage.billingPeriod.end)
    : null;

  const formatDate = (date: Date | null) => {
    if (!date) return 'N/A';
    return date.toLocaleDateString(undefined, {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getPlanName = () => {
    if (!subscription || subscription.status === 'inactive') return 'Free';
    return subscription.planName || 'Pro';
  };

  const isActive = subscription?.status === 'active';

  return (
    <section className="space-y-4">
      <div className="bg-editor-surface border border-editor-border rounded-lg p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-4">
            {/* Plan Info */}
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h3 className="text-lg font-semibold text-editor-text">
                  {getPlanName()} Plan
                </h3>
                {isActive && (
                  <span className="px-2 py-0.5 bg-editor-success/10 text-editor-success text-xs rounded-full">
                    Active
                  </span>
                )}
              </div>
              {subscription?.cancelAtPeriodEnd && (
                <p className="text-sm text-editor-warning">
                  Cancels at end of billing period
                </p>
              )}
            </div>

            {/* Billing Period */}
            <div className="flex items-center gap-2 text-sm text-editor-muted">
              <Calendar size={16} />
              <span>
                Billing period: {formatDate(billingPeriodStart)} -{' '}
                {formatDate(billingPeriodEnd)}
              </span>
            </div>

            {/* Payment Method */}
            {subscription?.paymentMethod && (
              <div className="flex items-center gap-2 text-sm text-editor-muted">
                <CreditCard size={16} />
                <span>
                  {subscription.paymentMethod.brand?.toUpperCase()} ending in{' '}
                  {subscription.paymentMethod.last4}
                </span>
              </div>
            )}
          </div>

          {/* Manage Button */}
          {isActive && (
            <button
              onClick={() => managePortal.mutate()}
              disabled={managePortal.isPending}
              className="flex items-center gap-2 px-4 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text hover:border-editor-accent/50 transition-colors disabled:opacity-50"
            >
              {managePortal.isPending ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Loading...
                </>
              ) : (
                <>
                  Manage Billing
                  <ArrowUpRight size={16} />
                </>
              )}
            </button>
          )}
        </div>

        {/* Next Invoice */}
        {subscription?.nextInvoice && (
          <div className="mt-6 pt-6 border-t border-editor-border">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-editor-muted">Next invoice</p>
                <p className="text-lg font-semibold text-editor-text">
                  ${(subscription.nextInvoice.amount / 100).toFixed(2)}
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-editor-muted">Due date</p>
                <p className="text-editor-text">
                  {formatDate(new Date(subscription.nextInvoice.dueDate))}
                </p>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
