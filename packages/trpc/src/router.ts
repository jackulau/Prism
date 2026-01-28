import { router } from './trpc.js';
import { integrationsRouter } from './routers/integrations/index.js';

export const appRouter = router({
  integrations: integrationsRouter,
});

export type AppRouter = typeof appRouter;
