/**
 * Organization tRPC hooks
 *
 * These hooks will work once the organization router is implemented
 * in packages/trpc/src/routers/organization.ts and added to the main router.
 *
 * Example future usage:
 *
 * ```ts
 * import { trpc } from '../lib/trpc';
 *
 * export const useProfile = () => {
 *   return trpc.organization.getProfile.useQuery();
 * };
 *
 * export const useUpdateSettings = () => {
 *   const utils = trpc.useUtils();
 *   return trpc.organization.updateSettings.useMutation({
 *     onSuccess: () => {
 *       utils.organization.getSettings.invalidate();
 *     },
 *   });
 * };
 * ```
 */

export {};
