/**
 * Payment tRPC hooks
 *
 * These hooks will work once the payment router is implemented
 * in packages/trpc/src/routers/payment.ts and added to the main router.
 *
 * Example future usage:
 *
 * ```ts
 * import { trpc } from '../lib/trpc';
 *
 * export const usePlans = () => {
 *   return trpc.payment.listPlans.useQuery();
 * };
 *
 * export const useCreateCheckout = () => {
 *   return trpc.payment.createCheckoutSession.useMutation({
 *     onSuccess: (data) => {
 *       window.location.href = data.url;
 *     },
 *   });
 * };
 * ```
 */

export {};
