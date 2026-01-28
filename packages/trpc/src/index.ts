import express from 'express';
import cors from 'cors';
import { createExpressMiddleware } from '@trpc/server/adapters/express';
import { appRouter } from './router.js';
import { createContext, setJWTSecret } from './context.js';

export { appRouter, type AppRouter } from './router.js';
export { createContext, setJWTSecret, type Context, type Session } from './context.js';
export { router, publicProcedure, middleware, mergeRouters } from './trpc.js';
export { protectedProcedure, isAuthed } from './middleware/auth.js';
export { createRateLimitMiddleware, rateLimit } from './middleware/rateLimit.js';
export {
  notFound,
  unauthorized,
  forbidden,
  badRequest,
  internalError,
  conflict,
} from './utils/errors.js';

// Payment router exports
export { paymentRouter, type PaymentRouter } from './routers/payment/index.js';
export { paymentService } from './services/payment.js';
export type {
  PlanType,
  BillingInterval,
  Plan,
  PlanLimits,
  SubscriptionStatus,
  Subscription,
  Usage,
  UsageHistoryItem,
  PaymentMethod,
  InvoiceStatus,
  Invoice,
  CreateSubscriptionInput,
  UpdateSubscriptionInput,
  CreateCheckoutInput,
  PortalInput,
  UsageHistoryInput,
  ListInvoicesInput,
} from './routers/payment/schemas.js';

export interface TRPCMiddlewareOptions {
  onError?: (opts: { error: Error; path: string | undefined }) => void;
}

export function createTRPCMiddleware(options: TRPCMiddlewareOptions = {}) {
  return createExpressMiddleware({
    router: appRouter,
    createContext,
    onError({ error, path }) {
      console.error(`tRPC Error on ${path}:`, error);
      options.onError?.({ error, path });
    },
  });
}

export interface ServerOptions {
  port?: number;
  corsOrigins?: string | string[];
  jwtSecret?: string;
}

export function startServer(options: ServerOptions = {}) {
  const { port = 3001, corsOrigins = '*', jwtSecret } = options;

  if (jwtSecret) {
    setJWTSecret(jwtSecret);
  }

  const app = express();

  app.use(
    cors({
      origin: corsOrigins,
      credentials: true,
    })
  );

  app.use('/trpc', createTRPCMiddleware());

  app.get('/health', (_req, res) => {
    res.json({ status: 'ok' });
  });

  app.listen(port, () => {
    console.log(`tRPC server running on port ${port}`);
  });

  return app;
}
