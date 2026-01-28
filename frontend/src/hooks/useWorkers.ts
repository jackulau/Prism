/**
 * Workers tRPC hooks
 *
 * These hooks will work once the workers router is implemented
 * in packages/trpc/src/routers/workers.ts and added to the main router.
 *
 * Example future usage:
 *
 * ```ts
 * import { trpc } from '../lib/trpc';
 *
 * export const useRunTask = () => {
 *   const utils = trpc.useUtils();
 *   return trpc.workers.runTask.useMutation({
 *     onSuccess: () => {
 *       utils.workers.listExecutions.invalidate();
 *     },
 *   });
 * };
 * ```
 */

export {};
