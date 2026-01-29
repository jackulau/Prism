---
id: trpc-payment-router
name: Payment Router Implementation
wave: 2
priority: 2
dependencies:
- trpc-core-setup
estimated_hours: 4
tags:
- backend
- api
- trpc
- payment
- billing
---

## Objective

Implement the tRPC payment router for billing operations, subscription management, and usage tracking.

## Context

The existing Go backend does not have payment/billing implemented yet. This router will establish the foundation for:
- Subscription plan management
- Usage tracking (tokens, API calls)
- Billing history
- Payment method management
- Stripe integration

This is a greenfield implementation as no payment infrastructure exists in the current codebase.

## Implementation

### 1. Define Zod Schemas

**File: `packages/trpc/src/routers/payment/schemas.ts`**
```typescript
import { z } from 'zod';

// Subscription plans
export const planTypeSchema = z.enum([
  'free',
  'pro',
  'team',
  'enterprise',
]);

export const billingIntervalSchema = z.enum([
  'monthly',
  'yearly',
]);

export const planSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: planTypeSchema,
  description: z.string(),
  monthlyPrice: z.number(), // in cents
  yearlyPrice: z.number(), // in cents
  features: z.array(z.string()),
  limits: z.object({
    tokensPerMonth: z.number().nullable(), // null = unlimited
    workspaces: z.number().nullable(),
    teamMembers: z.number().nullable(),
  }),
});

// Subscription status
export const subscriptionStatusSchema = z.enum([
  'active',
  'canceled',
  'past_due',
  'trialing',
  'incomplete',
]);

export const subscriptionSchema = z.object({
  id: z.string(),
  plan: planSchema,
  status: subscriptionStatusSchema,
  billingInterval: billingIntervalSchema,
  currentPeriodStart: z.date(),
  currentPeriodEnd: z.date(),
  cancelAtPeriodEnd: z.boolean(),
  trialEnd: z.date().nullable(),
});

// Usage tracking
export const usageSchema = z.object({
  tokensUsed: z.number(),
  tokensLimit: z.number().nullable(),
  tokensRemaining: z.number().nullable(),
  apiCallsThisMonth: z.number(),
  periodStart: z.date(),
  periodEnd: z.date(),
});

export const usageHistoryItemSchema = z.object({
  date: z.date(),
  tokensUsed: z.number(),
  apiCalls: z.number(),
});

// Payment methods
export const paymentMethodSchema = z.object({
  id: z.string(),
  type: z.enum(['card', 'bank_account']),
  isDefault: z.boolean(),
  card: z.object({
    brand: z.string(),
    last4: z.string(),
    expiryMonth: z.number(),
    expiryYear: z.number(),
  }).nullable(),
});

// Invoices
export const invoiceStatusSchema = z.enum([
  'draft',
  'open',
  'paid',
  'void',
  'uncollectible',
]);

export const invoiceSchema = z.object({
  id: z.string(),
  number: z.string(),
  status: invoiceStatusSchema,
  amount: z.number(), // in cents
  currency: z.string(),
  periodStart: z.date(),
  periodEnd: z.date(),
  paidAt: z.date().nullable(),
  invoiceUrl: z.string().nullable(),
  invoicePdf: z.string().nullable(),
});

// Input schemas
export const createSubscriptionInput = z.object({
  planType: planTypeSchema,
  billingInterval: billingIntervalSchema,
  paymentMethodId: z.string().optional(),
});

export const updateSubscriptionInput = z.object({
  planType: planTypeSchema.optional(),
  billingInterval: billingIntervalSchema.optional(),
  cancelAtPeriodEnd: z.boolean().optional(),
});

export const createCheckoutInput = z.object({
  planType: planTypeSchema,
  billingInterval: billingIntervalSchema,
  successUrl: z.string().url(),
  cancelUrl: z.string().url(),
});

export const portalInput = z.object({
  returnUrl: z.string().url(),
});

export const usageHistoryInput = z.object({
  startDate: z.date().optional(),
  endDate: z.date().optional(),
  granularity: z.enum(['day', 'week', 'month']).default('day'),
});

export type PlanType = z.infer<typeof planTypeSchema>;
export type Subscription = z.infer<typeof subscriptionSchema>;
export type Usage = z.infer<typeof usageSchema>;
export type Invoice = z.infer<typeof invoiceSchema>;
```

### 2. Implement Payment Router

**File: `packages/trpc/src/routers/payment/index.ts`**
```typescript
import { router, protectedProcedure, publicProcedure } from '../../trpc';
import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import * as schemas from './schemas';

export const paymentRouter = router({
  // Plans
  listPlans: publicProcedure
    .output(z.array(schemas.planSchema))
    .query(async () => {
      const plans = await paymentService.listPlans();
      return plans;
    }),

  getPlan: publicProcedure
    .input(z.object({ planType: schemas.planTypeSchema }))
    .output(schemas.planSchema.nullable())
    .query(async ({ input }) => {
      const plan = await paymentService.getPlan(input.planType);
      return plan;
    }),

  // Subscription
  getSubscription: protectedProcedure
    .output(schemas.subscriptionSchema.nullable())
    .query(async ({ ctx }) => {
      const subscription = await paymentService.getSubscription(ctx.session.userId);
      return subscription;
    }),

  createSubscription: protectedProcedure
    .input(schemas.createSubscriptionInput)
    .output(schemas.subscriptionSchema)
    .mutation(async ({ ctx, input }) => {
      const subscription = await paymentService.createSubscription(
        ctx.session.userId,
        input.planType,
        input.billingInterval,
        input.paymentMethodId
      );
      return subscription;
    }),

  updateSubscription: protectedProcedure
    .input(schemas.updateSubscriptionInput)
    .output(schemas.subscriptionSchema)
    .mutation(async ({ ctx, input }) => {
      const subscription = await paymentService.updateSubscription(
        ctx.session.userId,
        input
      );
      return subscription;
    }),

  cancelSubscription: protectedProcedure
    .input(z.object({ immediately: z.boolean().default(false) }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.cancelSubscription(ctx.session.userId, input.immediately);
      return { success: true };
    }),

  reactivateSubscription: protectedProcedure
    .mutation(async ({ ctx }) => {
      const subscription = await paymentService.reactivateSubscription(ctx.session.userId);
      return subscription;
    }),

  // Checkout (Stripe Checkout Session)
  createCheckoutSession: protectedProcedure
    .input(schemas.createCheckoutInput)
    .output(z.object({ url: z.string() }))
    .mutation(async ({ ctx, input }) => {
      const session = await paymentService.createCheckoutSession(
        ctx.session.userId,
        input.planType,
        input.billingInterval,
        input.successUrl,
        input.cancelUrl
      );
      return { url: session.url };
    }),

  // Customer Portal
  createPortalSession: protectedProcedure
    .input(schemas.portalInput)
    .output(z.object({ url: z.string() }))
    .mutation(async ({ ctx, input }) => {
      const session = await paymentService.createPortalSession(
        ctx.session.userId,
        input.returnUrl
      );
      return { url: session.url };
    }),

  // Usage
  getUsage: protectedProcedure
    .output(schemas.usageSchema)
    .query(async ({ ctx }) => {
      const usage = await paymentService.getUsage(ctx.session.userId);
      return usage;
    }),

  getUsageHistory: protectedProcedure
    .input(schemas.usageHistoryInput)
    .output(z.array(schemas.usageHistoryItemSchema))
    .query(async ({ ctx, input }) => {
      const history = await paymentService.getUsageHistory(
        ctx.session.userId,
        input.startDate,
        input.endDate,
        input.granularity
      );
      return history;
    }),

  // Payment Methods
  listPaymentMethods: protectedProcedure
    .output(z.array(schemas.paymentMethodSchema))
    .query(async ({ ctx }) => {
      const methods = await paymentService.listPaymentMethods(ctx.session.userId);
      return methods;
    }),

  setDefaultPaymentMethod: protectedProcedure
    .input(z.object({ paymentMethodId: z.string() }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.setDefaultPaymentMethod(
        ctx.session.userId,
        input.paymentMethodId
      );
      return { success: true };
    }),

  deletePaymentMethod: protectedProcedure
    .input(z.object({ paymentMethodId: z.string() }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.deletePaymentMethod(
        ctx.session.userId,
        input.paymentMethodId
      );
      return { success: true };
    }),

  // Invoices
  listInvoices: protectedProcedure
    .input(z.object({
      limit: z.number().min(1).max(100).default(10),
      startingAfter: z.string().optional(),
    }))
    .output(z.object({
      invoices: z.array(schemas.invoiceSchema),
      hasMore: z.boolean(),
    }))
    .query(async ({ ctx, input }) => {
      const result = await paymentService.listInvoices(
        ctx.session.userId,
        input.limit,
        input.startingAfter
      );
      return result;
    }),
});
```

### 3. Create Payment Service

**File: `packages/trpc/src/services/payment.ts`**
```typescript
import Stripe from 'stripe';
import type { PlanType, Subscription, Usage, Invoice, PaymentMethod } from '../routers/payment/schemas';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2023-10-16',
});

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
    monthlyPrice: 2000, // $20/month
    yearlyPrice: 19200, // $192/year ($16/month)
    features: ['100,000 tokens/month', 'Unlimited workspaces', 'Priority support'],
    limits: { tokensPerMonth: 100000, workspaces: null, teamMembers: 1 },
  },
  team: {
    id: 'team',
    name: 'Team',
    type: 'team',
    description: 'For small teams',
    monthlyPrice: 5000, // $50/month
    yearlyPrice: 48000, // $480/year ($40/month)
    features: ['500,000 tokens/month', 'Unlimited workspaces', 'Team collaboration', 'Admin dashboard'],
    limits: { tokensPerMonth: 500000, workspaces: null, teamMembers: 5 },
  },
  enterprise: {
    id: 'enterprise',
    name: 'Enterprise',
    type: 'enterprise',
    description: 'For large organizations',
    monthlyPrice: 0, // Custom pricing
    yearlyPrice: 0,
    features: ['Unlimited tokens', 'SSO', 'Dedicated support', 'SLA'],
    limits: { tokensPerMonth: null, workspaces: null, teamMembers: null },
  },
};

export const paymentService = {
  listPlans() {
    return Object.values(PLANS);
  },

  getPlan(planType: PlanType) {
    return PLANS[planType] || null;
  },

  async getSubscription(userId: string): Promise<Subscription | null> {
    // Get Stripe customer ID from database
    // Retrieve subscription from Stripe
  },

  async createSubscription(
    userId: string,
    planType: PlanType,
    billingInterval: 'monthly' | 'yearly',
    paymentMethodId?: string
  ): Promise<Subscription> {
    // Create or get Stripe customer
    // Create Stripe subscription
    // Store subscription ID in database
  },

  async updateSubscription(
    userId: string,
    updates: { planType?: PlanType; billingInterval?: string; cancelAtPeriodEnd?: boolean }
  ): Promise<Subscription> {
    // Update Stripe subscription
  },

  async cancelSubscription(userId: string, immediately: boolean) {
    // Cancel Stripe subscription
  },

  async reactivateSubscription(userId: string): Promise<Subscription> {
    // Remove cancel_at_period_end
  },

  async createCheckoutSession(
    userId: string,
    planType: PlanType,
    billingInterval: 'monthly' | 'yearly',
    successUrl: string,
    cancelUrl: string
  ) {
    // Create Stripe Checkout Session
    return stripe.checkout.sessions.create({
      mode: 'subscription',
      // ...
    });
  },

  async createPortalSession(userId: string, returnUrl: string) {
    // Create Stripe Customer Portal Session
    return stripe.billingPortal.sessions.create({
      customer: customerId,
      return_url: returnUrl,
    });
  },

  async getUsage(userId: string): Promise<Usage> {
    // Aggregate usage from tool_executions table
    // Compare against plan limits
  },

  async getUsageHistory(
    userId: string,
    startDate?: Date,
    endDate?: Date,
    granularity?: 'day' | 'week' | 'month'
  ) {
    // Query aggregated usage by time period
  },

  async listPaymentMethods(userId: string): Promise<PaymentMethod[]> {
    // List Stripe payment methods
  },

  async setDefaultPaymentMethod(userId: string, paymentMethodId: string) {
    // Update default payment method in Stripe
  },

  async deletePaymentMethod(userId: string, paymentMethodId: string) {
    // Detach payment method from Stripe customer
  },

  async listInvoices(userId: string, limit: number, startingAfter?: string) {
    // List Stripe invoices
  },

  // Webhook handlers
  async handleWebhook(event: Stripe.Event) {
    switch (event.type) {
      case 'customer.subscription.created':
      case 'customer.subscription.updated':
      case 'customer.subscription.deleted':
      case 'invoice.paid':
      case 'invoice.payment_failed':
        // Handle subscription events
    }
  },
};
```

### 4. Database Schema Additions

New tables needed:
```sql
CREATE TABLE subscriptions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  stripe_customer_id TEXT NOT NULL,
  stripe_subscription_id TEXT,
  plan_type TEXT NOT NULL,
  status TEXT NOT NULL,
  current_period_start DATETIME,
  current_period_end DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE usage_records (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  date DATE NOT NULL,
  tokens_used INTEGER DEFAULT 0,
  api_calls INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, date)
);
```

## Acceptance Criteria

- [ ] Plan listing working (public endpoint)
- [ ] Subscription CRUD working
- [ ] Stripe Checkout integration working
- [ ] Customer Portal integration working
- [ ] Usage tracking working
- [ ] Usage history with aggregation
- [ ] Payment method management working
- [ ] Invoice listing working
- [ ] Webhook handling for subscription events
- [ ] Database migrations for new tables

## Files to Create/Modify

- `packages/trpc/src/routers/payment/schemas.ts` - Zod schemas
- `packages/trpc/src/routers/payment/index.ts` - Router implementation
- `packages/trpc/src/services/payment.ts` - Payment service with Stripe
- `packages/trpc/src/router.ts` - Add payment router (modify)
- Database migration for subscriptions and usage tables

## Integration Points

- **Provides**: Payment/billing operations via tRPC
- **Consumes**: trpc-core-setup (protectedProcedure, publicProcedure, router, context)
- **Conflicts**: None (greenfield implementation)

## Notes

- Stripe is the assumed payment provider
- Need environment variables: STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET
- Consider implementing metered billing for usage-based pricing
- Existing tool_executions table can be used for usage aggregation
