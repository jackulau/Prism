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

// Workers router exports
export { workersRouter, type WorkersRouter } from './routers/workers/index.js';
export * from './routers/workers/schemas.js';
export { workerService } from './services/worker.js';

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
