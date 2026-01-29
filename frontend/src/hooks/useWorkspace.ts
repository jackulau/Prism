/**
 * Workspace tRPC hooks
 *
 * These hooks will work once the workspace router is implemented
 * in packages/trpc/src/routers/workspace.ts and added to the main router.
 *
 * Example future usage:
 *
 * ```ts
 * import { trpc } from '../lib/trpc';
 *
 * export const useCurrentWorkspace = () => {
 *   return trpc.workspace.getCurrent.useQuery();
 * };
 * ```
 */

export {};
