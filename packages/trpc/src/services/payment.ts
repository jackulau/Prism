import Stripe from 'stripe';
import type {
  PlanType,
  Plan,
  Subscription,
  SubscriptionStatus,
  BillingInterval,
  Usage,
  UsageHistoryItem,
  PaymentMethod,
  Invoice,
  InvoiceStatus,
} from '../routers/payment/schemas.js';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY || '', {
  apiVersion: '2023-10-16',
});

// Stripe Price IDs - these should be configured via environment variables in production
const STRIPE_PRICE_IDS: Record<PlanType, { monthly: string; yearly: string }> = {
  free: { monthly: '', yearly: '' },
  pro: {
    monthly: process.env.STRIPE_PRO_MONTHLY_PRICE_ID || '',
    yearly: process.env.STRIPE_PRO_YEARLY_PRICE_ID || '',
  },
  team: {
    monthly: process.env.STRIPE_TEAM_MONTHLY_PRICE_ID || '',
    yearly: process.env.STRIPE_TEAM_YEARLY_PRICE_ID || '',
  },
  enterprise: { monthly: '', yearly: '' },
};

// Plan definitions
const PLANS: Record<PlanType, Plan> = {
  free: {
    id: 'free',
    name: 'Free',
    type: 'free',
    description: 'For personal projects',
    monthlyPrice: 0,
    yearlyPrice: 0,
    features: ['5,000 tokens/month', '1 workspace', 'Community support'],
    limits: { tokensPerMonth: 5000, workspaces: 1, teamMembers: 1 },
  },
  pro: {
    id: 'pro',
    name: 'Pro',
    type: 'pro',
    description: 'For professional developers',
    monthlyPrice: 2000,
    yearlyPrice: 19200,
    features: ['100,000 tokens/month', 'Unlimited workspaces', 'Priority support'],
    limits: { tokensPerMonth: 100000, workspaces: null, teamMembers: 1 },
  },
  team: {
    id: 'team',
    name: 'Team',
    type: 'team',
    description: 'For small teams',
    monthlyPrice: 5000,
    yearlyPrice: 48000,
    features: [
      '500,000 tokens/month',
      'Unlimited workspaces',
      'Team collaboration',
      'Admin dashboard',
    ],
    limits: { tokensPerMonth: 500000, workspaces: null, teamMembers: 5 },
  },
  enterprise: {
    id: 'enterprise',
    name: 'Enterprise',
    type: 'enterprise',
    description: 'For large organizations',
    monthlyPrice: 0,
    yearlyPrice: 0,
    features: ['Unlimited tokens', 'SSO', 'Dedicated support', 'SLA'],
    limits: { tokensPerMonth: null, workspaces: null, teamMembers: null },
  },
};

function mapStripeStatus(status: Stripe.Subscription.Status): SubscriptionStatus {
  const statusMap: Record<Stripe.Subscription.Status, SubscriptionStatus> = {
    active: 'active',
    canceled: 'canceled',
    incomplete: 'incomplete',
    incomplete_expired: 'incomplete',
    past_due: 'past_due',
    paused: 'canceled',
    trialing: 'trialing',
    unpaid: 'past_due',
  };
  return statusMap[status] || 'incomplete';
}

function mapInvoiceStatus(status: Stripe.Invoice.Status | null): InvoiceStatus {
  if (!status) return 'draft';
  const statusMap: Record<string, InvoiceStatus> = {
    draft: 'draft',
    open: 'open',
    paid: 'paid',
    void: 'void',
    uncollectible: 'uncollectible',
  };
  return statusMap[status] || 'draft';
}

function getPlanTypeFromPriceId(priceId: string): PlanType {
  for (const [planType, prices] of Object.entries(STRIPE_PRICE_IDS)) {
    if (prices.monthly === priceId || prices.yearly === priceId) {
      return planType as PlanType;
    }
  }
  return 'free';
}

function getBillingIntervalFromPrice(priceId: string): BillingInterval {
  for (const prices of Object.values(STRIPE_PRICE_IDS)) {
    if (prices.yearly === priceId) {
      return 'yearly';
    }
  }
  return 'monthly';
}

// In-memory store for demo purposes - replace with database in production
const userCustomerMap = new Map<string, string>();
const userSubscriptionMap = new Map<string, string>();
const usageRecords = new Map<string, Map<string, { tokensUsed: number; apiCalls: number }>>();

export const paymentService = {
  listPlans(): Plan[] {
    return Object.values(PLANS);
  },

  getPlan(planType: PlanType): Plan | null {
    return PLANS[planType] || null;
  },

  async getOrCreateCustomer(userId: string, email?: string): Promise<string> {
    let customerId = userCustomerMap.get(userId);
    if (customerId) {
      return customerId;
    }

    const customer = await stripe.customers.create({
      metadata: { userId },
      email,
    });

    userCustomerMap.set(userId, customer.id);
    return customer.id;
  },

  async getSubscription(userId: string): Promise<Subscription | null> {
    const subscriptionId = userSubscriptionMap.get(userId);
    if (!subscriptionId) {
      return null;
    }

    try {
      const stripeSub = await stripe.subscriptions.retrieve(subscriptionId, {
        expand: ['items.data.price'],
      });

      const priceId = stripeSub.items.data[0]?.price?.id || '';
      const planType = getPlanTypeFromPriceId(priceId);
      const plan = PLANS[planType];

      return {
        id: stripeSub.id,
        plan,
        status: mapStripeStatus(stripeSub.status),
        billingInterval: getBillingIntervalFromPrice(priceId),
        currentPeriodStart: new Date(stripeSub.current_period_start * 1000),
        currentPeriodEnd: new Date(stripeSub.current_period_end * 1000),
        cancelAtPeriodEnd: stripeSub.cancel_at_period_end,
        trialEnd: stripeSub.trial_end ? new Date(stripeSub.trial_end * 1000) : null,
      };
    } catch {
      return null;
    }
  },

  async createSubscription(
    userId: string,
    planType: PlanType,
    billingInterval: BillingInterval,
    paymentMethodId?: string,
    email?: string
  ): Promise<Subscription> {
    const customerId = await this.getOrCreateCustomer(userId, email);

    if (paymentMethodId) {
      await stripe.paymentMethods.attach(paymentMethodId, { customer: customerId });
      await stripe.customers.update(customerId, {
        invoice_settings: { default_payment_method: paymentMethodId },
      });
    }

    const priceId = STRIPE_PRICE_IDS[planType][billingInterval];
    if (!priceId) {
      throw new Error(`No price configured for ${planType} ${billingInterval}`);
    }

    const subscription = await stripe.subscriptions.create({
      customer: customerId,
      items: [{ price: priceId }],
      payment_behavior: 'default_incomplete',
      expand: ['latest_invoice.payment_intent'],
    });

    userSubscriptionMap.set(userId, subscription.id);

    const plan = PLANS[planType];
    return {
      id: subscription.id,
      plan,
      status: mapStripeStatus(subscription.status),
      billingInterval,
      currentPeriodStart: new Date(subscription.current_period_start * 1000),
      currentPeriodEnd: new Date(subscription.current_period_end * 1000),
      cancelAtPeriodEnd: subscription.cancel_at_period_end,
      trialEnd: subscription.trial_end ? new Date(subscription.trial_end * 1000) : null,
    };
  },

  async updateSubscription(
    userId: string,
    updates: {
      planType?: PlanType;
      billingInterval?: BillingInterval;
      cancelAtPeriodEnd?: boolean;
    }
  ): Promise<Subscription> {
    const subscriptionId = userSubscriptionMap.get(userId);
    if (!subscriptionId) {
      throw new Error('No subscription found');
    }

    const currentSub = await stripe.subscriptions.retrieve(subscriptionId);
    const updateParams: Stripe.SubscriptionUpdateParams = {};

    if (updates.planType || updates.billingInterval) {
      const currentPriceId = currentSub.items.data[0]?.price?.id || '';
      const currentPlanType = updates.planType || getPlanTypeFromPriceId(currentPriceId);
      const currentInterval = updates.billingInterval || getBillingIntervalFromPrice(currentPriceId);

      const newPriceId = STRIPE_PRICE_IDS[currentPlanType][currentInterval];
      const firstItem = currentSub.items.data[0];
      if (newPriceId && newPriceId !== currentPriceId && firstItem) {
        updateParams.items = [
          {
            id: firstItem.id,
            price: newPriceId,
          },
        ];
      }
    }

    if (updates.cancelAtPeriodEnd !== undefined) {
      updateParams.cancel_at_period_end = updates.cancelAtPeriodEnd;
    }

    const updatedSub = await stripe.subscriptions.update(subscriptionId, updateParams);

    const priceId = updatedSub.items.data[0]?.price?.id || '';
    const planType = getPlanTypeFromPriceId(priceId);
    const plan = PLANS[planType];

    return {
      id: updatedSub.id,
      plan,
      status: mapStripeStatus(updatedSub.status),
      billingInterval: getBillingIntervalFromPrice(priceId),
      currentPeriodStart: new Date(updatedSub.current_period_start * 1000),
      currentPeriodEnd: new Date(updatedSub.current_period_end * 1000),
      cancelAtPeriodEnd: updatedSub.cancel_at_period_end,
      trialEnd: updatedSub.trial_end ? new Date(updatedSub.trial_end * 1000) : null,
    };
  },

  async cancelSubscription(userId: string, immediately: boolean): Promise<void> {
    const subscriptionId = userSubscriptionMap.get(userId);
    if (!subscriptionId) {
      throw new Error('No subscription found');
    }

    if (immediately) {
      await stripe.subscriptions.cancel(subscriptionId);
      userSubscriptionMap.delete(userId);
    } else {
      await stripe.subscriptions.update(subscriptionId, {
        cancel_at_period_end: true,
      });
    }
  },

  async reactivateSubscription(userId: string): Promise<Subscription> {
    const subscriptionId = userSubscriptionMap.get(userId);
    if (!subscriptionId) {
      throw new Error('No subscription found');
    }

    const updatedSub = await stripe.subscriptions.update(subscriptionId, {
      cancel_at_period_end: false,
    });

    const priceId = updatedSub.items.data[0]?.price?.id || '';
    const planType = getPlanTypeFromPriceId(priceId);
    const plan = PLANS[planType];

    return {
      id: updatedSub.id,
      plan,
      status: mapStripeStatus(updatedSub.status),
      billingInterval: getBillingIntervalFromPrice(priceId),
      currentPeriodStart: new Date(updatedSub.current_period_start * 1000),
      currentPeriodEnd: new Date(updatedSub.current_period_end * 1000),
      cancelAtPeriodEnd: updatedSub.cancel_at_period_end,
      trialEnd: updatedSub.trial_end ? new Date(updatedSub.trial_end * 1000) : null,
    };
  },

  async createCheckoutSession(
    userId: string,
    planType: PlanType,
    billingInterval: BillingInterval,
    successUrl: string,
    cancelUrl: string,
    email?: string
  ): Promise<{ url: string }> {
    const customerId = await this.getOrCreateCustomer(userId, email);
    const priceId = STRIPE_PRICE_IDS[planType][billingInterval];

    if (!priceId) {
      throw new Error(`No price configured for ${planType} ${billingInterval}`);
    }

    const session = await stripe.checkout.sessions.create({
      customer: customerId,
      mode: 'subscription',
      line_items: [{ price: priceId, quantity: 1 }],
      success_url: successUrl,
      cancel_url: cancelUrl,
      metadata: { userId, planType, billingInterval },
    });

    return { url: session.url || '' };
  },

  async createPortalSession(userId: string, returnUrl: string): Promise<{ url: string }> {
    const customerId = userCustomerMap.get(userId);
    if (!customerId) {
      throw new Error('No customer found');
    }

    const session = await stripe.billingPortal.sessions.create({
      customer: customerId,
      return_url: returnUrl,
    });

    return { url: session.url };
  },

  async getUsage(userId: string): Promise<Usage> {
    const subscription = await this.getSubscription(userId);
    const plan = subscription?.plan || PLANS.free;

    const now = new Date();
    const periodStart = subscription?.currentPeriodStart || new Date(now.getFullYear(), now.getMonth(), 1);
    const periodEnd = subscription?.currentPeriodEnd || new Date(now.getFullYear(), now.getMonth() + 1, 0);

    // Get usage from records
    const userUsage = usageRecords.get(userId);
    let totalTokens = 0;
    let totalApiCalls = 0;

    if (userUsage) {
      for (const [dateStr, usage] of userUsage.entries()) {
        const date = new Date(dateStr);
        if (date >= periodStart && date <= periodEnd) {
          totalTokens += usage.tokensUsed;
          totalApiCalls += usage.apiCalls;
        }
      }
    }

    const tokensLimit = plan.limits.tokensPerMonth;
    const tokensRemaining = tokensLimit !== null ? Math.max(0, tokensLimit - totalTokens) : null;

    return {
      tokensUsed: totalTokens,
      tokensLimit,
      tokensRemaining,
      apiCallsThisMonth: totalApiCalls,
      periodStart,
      periodEnd,
    };
  },

  async getUsageHistory(
    userId: string,
    startDate?: Date,
    endDate?: Date,
    _granularity?: 'day' | 'week' | 'month'
  ): Promise<UsageHistoryItem[]> {
    const userUsage = usageRecords.get(userId);
    if (!userUsage) {
      return [];
    }

    const start = startDate || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const end = endDate || new Date();

    const history: UsageHistoryItem[] = [];

    for (const [dateStr, usage] of userUsage.entries()) {
      const date = new Date(dateStr);
      if (date >= start && date <= end) {
        history.push({
          date,
          tokensUsed: usage.tokensUsed,
          apiCalls: usage.apiCalls,
        });
      }
    }

    return history.sort((a, b) => a.date.getTime() - b.date.getTime());
  },

  async recordUsage(userId: string, tokensUsed: number, apiCalls: number): Promise<void> {
    const dateStr = new Date().toISOString().split('T')[0] || '';

    let userUsage = usageRecords.get(userId);
    if (!userUsage) {
      userUsage = new Map();
      usageRecords.set(userId, userUsage);
    }

    const existing = userUsage.get(dateStr) || { tokensUsed: 0, apiCalls: 0 };
    userUsage.set(dateStr, {
      tokensUsed: existing.tokensUsed + tokensUsed,
      apiCalls: existing.apiCalls + apiCalls,
    });
  },

  async listPaymentMethods(userId: string): Promise<PaymentMethod[]> {
    const customerId = userCustomerMap.get(userId);
    if (!customerId) {
      return [];
    }

    const customer = await stripe.customers.retrieve(customerId);
    const defaultPaymentMethodId =
      typeof customer !== 'string' && !customer.deleted
        ? customer.invoice_settings?.default_payment_method
        : null;

    const paymentMethods = await stripe.paymentMethods.list({
      customer: customerId,
      type: 'card',
    });

    return paymentMethods.data.map((pm) => ({
      id: pm.id,
      type: 'card' as const,
      isDefault: pm.id === defaultPaymentMethodId,
      card: pm.card
        ? {
            brand: pm.card.brand,
            last4: pm.card.last4,
            expiryMonth: pm.card.exp_month,
            expiryYear: pm.card.exp_year,
          }
        : null,
    }));
  },

  async setDefaultPaymentMethod(userId: string, paymentMethodId: string): Promise<void> {
    const customerId = userCustomerMap.get(userId);
    if (!customerId) {
      throw new Error('No customer found');
    }

    await stripe.customers.update(customerId, {
      invoice_settings: { default_payment_method: paymentMethodId },
    });
  },

  async deletePaymentMethod(userId: string, paymentMethodId: string): Promise<void> {
    const customerId = userCustomerMap.get(userId);
    if (!customerId) {
      throw new Error('No customer found');
    }

    const paymentMethod = await stripe.paymentMethods.retrieve(paymentMethodId);
    if (paymentMethod.customer !== customerId) {
      throw new Error('Payment method does not belong to user');
    }

    await stripe.paymentMethods.detach(paymentMethodId);
  },

  async listInvoices(
    userId: string,
    limit: number,
    startingAfter?: string
  ): Promise<{ invoices: Invoice[]; hasMore: boolean }> {
    const customerId = userCustomerMap.get(userId);
    if (!customerId) {
      return { invoices: [], hasMore: false };
    }

    const params: Stripe.InvoiceListParams = {
      customer: customerId,
      limit,
    };

    if (startingAfter) {
      params.starting_after = startingAfter;
    }

    const invoices = await stripe.invoices.list(params);

    return {
      invoices: invoices.data.map((inv) => ({
        id: inv.id,
        number: inv.number || '',
        status: mapInvoiceStatus(inv.status),
        amount: inv.amount_due,
        currency: inv.currency,
        periodStart: new Date(inv.period_start * 1000),
        periodEnd: new Date(inv.period_end * 1000),
        paidAt: inv.status_transitions?.paid_at
          ? new Date(inv.status_transitions.paid_at * 1000)
          : null,
        invoiceUrl: inv.hosted_invoice_url || null,
        invoicePdf: inv.invoice_pdf || null,
      })),
      hasMore: invoices.has_more,
    };
  },

  async handleWebhook(event: Stripe.Event): Promise<void> {
    switch (event.type) {
      case 'checkout.session.completed': {
        const session = event.data.object as Stripe.Checkout.Session;
        const userId = session.metadata?.userId;
        if (userId && session.subscription) {
          userSubscriptionMap.set(userId, session.subscription as string);
        }
        break;
      }

      case 'customer.subscription.created':
      case 'customer.subscription.updated': {
        const subscription = event.data.object as Stripe.Subscription;
        const customer = await stripe.customers.retrieve(subscription.customer as string);
        if (typeof customer !== 'string' && !customer.deleted && customer.metadata?.userId) {
          userSubscriptionMap.set(customer.metadata.userId, subscription.id);
        }
        break;
      }

      case 'customer.subscription.deleted': {
        const subscription = event.data.object as Stripe.Subscription;
        const customer = await stripe.customers.retrieve(subscription.customer as string);
        if (typeof customer !== 'string' && !customer.deleted && customer.metadata?.userId) {
          userSubscriptionMap.delete(customer.metadata.userId);
        }
        break;
      }

      case 'invoice.paid': {
        // Handle successful payment - could update subscription status or send notifications
        break;
      }

      case 'invoice.payment_failed': {
        // Handle failed payment - could notify user or update subscription status
        break;
      }
    }
  },
};
