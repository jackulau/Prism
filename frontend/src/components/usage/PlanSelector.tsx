import { Check, Zap, Loader2 } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { toast } from '../../store/toastStore';

interface Plan {
  id: string;
  name: string;
  price: number;
  interval: 'month' | 'year';
  tokenLimit: number;
  sandboxHours: number;
  features: string[];
  popular?: boolean;
}

const PLANS: Plan[] = [
  {
    id: 'free',
    name: 'Free',
    price: 0,
    interval: 'month',
    tokenLimit: 500000,
    sandboxHours: 2,
    features: [
      '$5 worth of tokens',
      '2 hours sandbox time',
      'Basic support',
      'Community access',
    ],
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 20,
    interval: 'month',
    tokenLimit: 2000000,
    sandboxHours: 24,
    features: [
      '$20 worth of tokens',
      '24 hours sandbox time',
      'Priority support',
      'Advanced integrations',
      'Custom workers',
    ],
    popular: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: -1, // Custom pricing
    interval: 'month',
    tokenLimit: -1, // Unlimited
    sandboxHours: -1, // Unlimited
    features: [
      'Unlimited tokens',
      'Unlimited sandbox time',
      'Dedicated support',
      'Custom integrations',
      'SLA guarantee',
      'SSO & SAML',
    ],
  },
];

export function PlanSelector() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: currentPlan } = (trpc as any).payment.getCurrentPlan.useQuery();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const createCheckout = (trpc as any).payment.createCheckoutSession.useMutation({
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

  const handleSelectPlan = (planId: string) => {
    if (planId === 'enterprise') {
      // Open contact form or email
      window.location.href = 'mailto:enterprise@prism.ai?subject=Enterprise Plan Inquiry';
      return;
    }

    if (planId === 'free') {
      toast.info('You are already on the Free plan');
      return;
    }

    createCheckout.mutate({ planId });
  };

  const formatTokens = (value: number) => {
    if (value === -1) return 'Unlimited';
    if (value >= 1000000) return `${(value / 1000000).toFixed(0)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(0)}K`;
    return value.toString();
  };

  const formatHours = (hours: number) => {
    if (hours === -1) return 'Unlimited';
    return `${hours}h`;
  };

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold text-editor-text">Plans</h2>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {PLANS.map((plan) => {
          const isCurrent = currentPlan?.planId === plan.id;
          const isPopular = plan.popular;

          return (
            <div
              key={plan.id}
              className={`relative bg-editor-surface border rounded-xl p-6 ${
                isPopular
                  ? 'border-editor-accent shadow-lg shadow-editor-accent/10'
                  : 'border-editor-border'
              }`}
            >
              {isPopular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="px-3 py-1 bg-editor-accent text-white text-xs font-medium rounded-full">
                    Most Popular
                  </span>
                </div>
              )}

              <div className="mb-6">
                <h3 className="text-xl font-bold text-editor-text mb-2">{plan.name}</h3>
                <div className="flex items-baseline gap-1">
                  {plan.price === -1 ? (
                    <span className="text-2xl font-bold text-editor-text">Custom</span>
                  ) : (
                    <>
                      <span className="text-3xl font-bold text-editor-text">
                        ${plan.price}
                      </span>
                      <span className="text-editor-muted">/{plan.interval}</span>
                    </>
                  )}
                </div>
              </div>

              {/* Limits */}
              <div className="space-y-3 mb-6">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-editor-muted">Tokens</span>
                  <span className="font-medium text-editor-text">
                    {formatTokens(plan.tokenLimit)}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-editor-muted">Sandbox</span>
                  <span className="font-medium text-editor-text">
                    {formatHours(plan.sandboxHours)}
                  </span>
                </div>
              </div>

              {/* Features */}
              <ul className="space-y-2 mb-6">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-sm">
                    <Check size={16} className="text-editor-success flex-shrink-0 mt-0.5" />
                    <span className="text-editor-muted">{feature}</span>
                  </li>
                ))}
              </ul>

              {/* CTA Button */}
              <button
                onClick={() => handleSelectPlan(plan.id)}
                disabled={isCurrent || createCheckout.isPending}
                className={`w-full py-2.5 rounded-lg font-medium transition-colors flex items-center justify-center gap-2 ${
                  isCurrent
                    ? 'bg-editor-success/10 text-editor-success cursor-default'
                    : isPopular
                    ? 'bg-editor-accent text-white hover:bg-editor-accent/90'
                    : 'bg-editor-bg border border-editor-border text-editor-text hover:border-editor-accent/50'
                } disabled:opacity-50`}
              >
                {createCheckout.isPending ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Processing...
                  </>
                ) : isCurrent ? (
                  <>
                    <Check size={16} />
                    Current Plan
                  </>
                ) : plan.id === 'enterprise' ? (
                  'Contact Sales'
                ) : (
                  <>
                    <Zap size={16} />
                    {plan.price === 0 ? 'Get Started' : 'Upgrade'}
                  </>
                )}
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}
