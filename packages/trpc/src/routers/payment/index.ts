import { z } from 'zod';
import { router, publicProcedure } from '../../trpc.js';
import { protectedProcedure } from '../../middleware/auth.js';
import { paymentService } from '../../services/payment.js';
import * as schemas from './schemas.js';

export const paymentRouter = router({
  // Plans - public endpoints
  listPlans: publicProcedure
    .output(z.array(schemas.planSchema))
    .query(async () => {
      return paymentService.listPlans();
    }),

  getPlan: publicProcedure
    .input(z.object({ planType: schemas.planTypeSchema }))
    .output(schemas.planSchema.nullable())
    .query(async ({ input }) => {
      return paymentService.getPlan(input.planType);
    }),

  // Subscription management - protected endpoints
  getSubscription: protectedProcedure
    .output(schemas.subscriptionSchema.nullable())
    .query(async ({ ctx }) => {
      return paymentService.getSubscription(ctx.session.userId);
    }),

  createSubscription: protectedProcedure
    .input(schemas.createSubscriptionInput)
    .output(schemas.subscriptionSchema)
    .mutation(async ({ ctx, input }) => {
      return paymentService.createSubscription(
        ctx.session.userId,
        input.planType,
        input.billingInterval,
        input.paymentMethodId,
        ctx.session.email
      );
    }),

  updateSubscription: protectedProcedure
    .input(schemas.updateSubscriptionInput)
    .output(schemas.subscriptionSchema)
    .mutation(async ({ ctx, input }) => {
      return paymentService.updateSubscription(ctx.session.userId, {
        planType: input.planType,
        billingInterval: input.billingInterval,
        cancelAtPeriodEnd: input.cancelAtPeriodEnd,
      });
    }),

  cancelSubscription: protectedProcedure
    .input(z.object({ immediately: z.boolean().default(false) }))
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.cancelSubscription(ctx.session.userId, input.immediately);
      return { success: true };
    }),

  reactivateSubscription: protectedProcedure
    .output(schemas.subscriptionSchema)
    .mutation(async ({ ctx }) => {
      return paymentService.reactivateSubscription(ctx.session.userId);
    }),

  // Checkout and Portal
  createCheckoutSession: protectedProcedure
    .input(schemas.createCheckoutInput)
    .output(z.object({ url: z.string() }))
    .mutation(async ({ ctx, input }) => {
      return paymentService.createCheckoutSession(
        ctx.session.userId,
        input.planType,
        input.billingInterval,
        input.successUrl,
        input.cancelUrl,
        ctx.session.email
      );
    }),

  createPortalSession: protectedProcedure
    .input(schemas.portalInput)
    .output(z.object({ url: z.string() }))
    .mutation(async ({ ctx, input }) => {
      return paymentService.createPortalSession(ctx.session.userId, input.returnUrl);
    }),

  // Usage tracking
  getUsage: protectedProcedure
    .output(schemas.usageSchema)
    .query(async ({ ctx }) => {
      return paymentService.getUsage(ctx.session.userId);
    }),

  getUsageHistory: protectedProcedure
    .input(schemas.usageHistoryInput)
    .output(z.array(schemas.usageHistoryItemSchema))
    .query(async ({ ctx, input }) => {
      return paymentService.getUsageHistory(
        ctx.session.userId,
        input.startDate,
        input.endDate,
        input.granularity
      );
    }),

  // Payment methods
  listPaymentMethods: protectedProcedure
    .output(z.array(schemas.paymentMethodSchema))
    .query(async ({ ctx }) => {
      return paymentService.listPaymentMethods(ctx.session.userId);
    }),

  setDefaultPaymentMethod: protectedProcedure
    .input(z.object({ paymentMethodId: z.string() }))
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.setDefaultPaymentMethod(ctx.session.userId, input.paymentMethodId);
      return { success: true };
    }),

  deletePaymentMethod: protectedProcedure
    .input(z.object({ paymentMethodId: z.string() }))
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await paymentService.deletePaymentMethod(ctx.session.userId, input.paymentMethodId);
      return { success: true };
    }),

  // Invoices
  listInvoices: protectedProcedure
    .input(schemas.listInvoicesInput)
    .output(
      z.object({
        invoices: z.array(schemas.invoiceSchema),
        hasMore: z.boolean(),
      })
    )
    .query(async ({ ctx, input }) => {
      return paymentService.listInvoices(ctx.session.userId, input.limit, input.startingAfter);
    }),
});

export type PaymentRouter = typeof paymentRouter;
