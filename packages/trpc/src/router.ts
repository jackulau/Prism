import { router } from './trpc.js';
import { workspaceRouter } from './routers/workspace/index.js';

export const appRouter = router({
  workspace: workspaceRouter,
});

export type AppRouter = typeof appRouter;
