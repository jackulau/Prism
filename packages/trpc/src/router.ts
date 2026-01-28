import { router } from './trpc.js';
import { workersRouter } from './routers/workers/index.js';

export const appRouter = router({
  workers: workersRouter,
});

export type AppRouter = typeof appRouter;
