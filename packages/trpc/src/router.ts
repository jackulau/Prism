import { router } from './trpc.js';
import { organizationRouter } from './routers/organization/index.js';

export const appRouter = router({
  organization: organizationRouter,
});

export type AppRouter = typeof appRouter;
