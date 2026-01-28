import { z } from 'zod';

// Subscription plans
export const planTypeSchema = z.enum(['free', 'pro', 'team', 'enterprise']);

export const billingIntervalSchema = z.enum(['monthly', 'yearly']);

export const planLimitsSchema = z.object({
  tokensPerMonth: z.number().nullable(),
  workspaces: z.number().nullable(),
  teamMembers: z.number().nullable(),
});

export const planSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: planTypeSchema,
  description: z.string(),
  monthlyPrice: z.number(),
  yearlyPrice: z.number(),
  features: z.array(z.string()),
  limits: planLimitsSchema,
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
export const cardDetailsSchema = z.object({
  brand: z.string(),
  last4: z.string(),
  expiryMonth: z.number(),
  expiryYear: z.number(),
});

export const paymentMethodSchema = z.object({
  id: z.string(),
  type: z.enum(['card', 'bank_account']),
  isDefault: z.boolean(),
  card: cardDetailsSchema.nullable(),
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
  amount: z.number(),
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

export const listInvoicesInput = z.object({
  limit: z.number().min(1).max(100).default(10),
  startingAfter: z.string().optional(),
});

// Type exports
export type PlanType = z.infer<typeof planTypeSchema>;
export type BillingInterval = z.infer<typeof billingIntervalSchema>;
export type Plan = z.infer<typeof planSchema>;
export type PlanLimits = z.infer<typeof planLimitsSchema>;
export type SubscriptionStatus = z.infer<typeof subscriptionStatusSchema>;
export type Subscription = z.infer<typeof subscriptionSchema>;
export type Usage = z.infer<typeof usageSchema>;
export type UsageHistoryItem = z.infer<typeof usageHistoryItemSchema>;
export type PaymentMethod = z.infer<typeof paymentMethodSchema>;
export type InvoiceStatus = z.infer<typeof invoiceStatusSchema>;
export type Invoice = z.infer<typeof invoiceSchema>;
export type CreateSubscriptionInput = z.infer<typeof createSubscriptionInput>;
export type UpdateSubscriptionInput = z.infer<typeof updateSubscriptionInput>;
export type CreateCheckoutInput = z.infer<typeof createCheckoutInput>;
export type PortalInput = z.infer<typeof portalInput>;
export type UsageHistoryInput = z.infer<typeof usageHistoryInput>;
export type ListInvoicesInput = z.infer<typeof listInvoicesInput>;
