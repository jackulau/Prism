import { router } from './trpc.js';
import { paymentRouter } from './routers/payment/index.js';
import { workspaceRouter } from './routers/workspace/index.js';
import { workersRouter } from './routers/workers/index.js';
import { integrationsRouter } from './routers/integrations/index.js';
import { organizationRouter } from './routers/organization/index.js';

export const appRouter = router({
  payment: paymentRouter,
  workspace: workspaceRouter,
  workers: workersRouter,
  integrations: integrationsRouter,
  organization: organizationRouter,
});

export type AppRouter = typeof appRouter;
