import type { AnyRouter } from '@trpc/server';
import { router } from './trpc.js';

export const appRouter: AnyRouter = router({});

export type AppRouter = typeof appRouter;
