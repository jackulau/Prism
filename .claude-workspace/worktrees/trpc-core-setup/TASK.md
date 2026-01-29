---
id: trpc-core-setup
name: tRPC Core Setup with Router Configuration
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- api
- trpc
- core
---

## Objective

Set up the core tRPC infrastructure with router configuration, context creation with auth, and protected procedure middleware.

## Context

This is the foundational task that establishes the tRPC layer. The existing codebase is a Go/Fiber backend with JWT authentication. We need to add a TypeScript/Node.js tRPC server that can either:
1. Run as a separate service alongside the Go backend
2. Replace portions of the Go REST API

This task creates the base tRPC setup that all routers will depend on.

## Implementation

### 1. Create tRPC Server Package Structure

Create directory: `packages/trpc/` with the following structure:
```
packages/trpc/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts           # Main exports
│   ├── router.ts          # Root router combining all sub-routers
│   ├── context.ts         # Context creation with auth
│   ├── trpc.ts            # tRPC initialization and procedures
│   ├── middleware/
│   │   ├── auth.ts        # Protected procedure middleware
│   │   └── rateLimit.ts   # Rate limiting middleware
│   └── utils/
│       └── errors.ts      # TRPCError helpers
```

### 2. Initialize tRPC with Context

**File: `packages/trpc/src/trpc.ts`**
```typescript
import { initTRPC, TRPCError } from '@trpc/server';
import superjson from 'superjson';
import type { Context } from './context';

const t = initTRPC.context<Context>().create({
  transformer: superjson,
  errorFormatter({ shape, error }) {
    return {
      ...shape,
      data: {
        ...shape.data,
        zodError: error.cause instanceof ZodError ? error.cause.flatten() : null,
      },
    };
  },
});

export const router = t.router;
export const publicProcedure = t.procedure;
export const middleware = t.middleware;
```

### 3. Create Auth Context

**File: `packages/trpc/src/context.ts`**
```typescript
import type { CreateExpressContextOptions } from '@trpc/server/adapters/express';

export interface Session {
  userId: string;
  email: string;
}

export interface Context {
  session: Session | null;
}

export const createContext = async (opts: CreateExpressContextOptions): Promise<Context> => {
  const authHeader = opts.req.headers.authorization;

  if (authHeader?.startsWith('Bearer ')) {
    const token = authHeader.slice(7);
    // Validate JWT token - integrate with existing Go JWT validation
    // or use jose library for standalone validation
    const session = await validateToken(token);
    return { session };
  }

  return { session: null };
};
```

### 4. Create Protected Procedure Middleware

**File: `packages/trpc/src/middleware/auth.ts`**
```typescript
import { TRPCError } from '@trpc/server';
import { middleware } from '../trpc';

export const isAuthed = middleware(async ({ ctx, next }) => {
  if (!ctx.session) {
    throw new TRPCError({ code: 'UNAUTHORIZED' });
  }
  return next({
    ctx: {
      session: ctx.session, // Session is now non-null
    },
  });
});

export const protectedProcedure = publicProcedure.use(isAuthed);
```

### 5. Create Error Utilities

**File: `packages/trpc/src/utils/errors.ts`**
```typescript
import { TRPCError } from '@trpc/server';

export const notFound = (resource: string) =>
  new TRPCError({
    code: 'NOT_FOUND',
    message: `${resource} not found`,
  });

export const unauthorized = (message = 'Unauthorized') =>
  new TRPCError({
    code: 'UNAUTHORIZED',
    message,
  });

export const forbidden = (message = 'Forbidden') =>
  new TRPCError({
    code: 'FORBIDDEN',
    message,
  });

export const badRequest = (message: string) =>
  new TRPCError({
    code: 'BAD_REQUEST',
    message,
  });
```

### 6. Create Root Router Shell

**File: `packages/trpc/src/router.ts`**
```typescript
import { router } from './trpc';

// Routers will be imported here as they are implemented
// import { workspaceRouter } from './routers/workspace';
// import { workersRouter } from './routers/workers';

export const appRouter = router({
  // workspace: workspaceRouter,
  // workers: workersRouter,
  // integrations: integrationsRouter,
  // organization: organizationRouter,
  // payment: paymentRouter,
});

export type AppRouter = typeof appRouter;
```

### 7. Set Up Express Adapter

**File: `packages/trpc/src/index.ts`**
```typescript
import express from 'express';
import cors from 'cors';
import { createExpressMiddleware } from '@trpc/server/adapters/express';
import { appRouter } from './router';
import { createContext } from './context';

export { appRouter, type AppRouter } from './router';
export { createContext, type Context } from './context';
export { publicProcedure, protectedProcedure } from './trpc';

export const createTRPCMiddleware = () =>
  createExpressMiddleware({
    router: appRouter,
    createContext,
    onError({ error, path }) {
      console.error(`tRPC Error on ${path}:`, error);
    },
  });

// Standalone server (optional)
export const startServer = (port = 3001) => {
  const app = express();
  app.use(cors());
  app.use('/trpc', createTRPCMiddleware());
  app.listen(port, () => {
    console.log(`tRPC server running on port ${port}`);
  });
};
```

## Acceptance Criteria

- [ ] tRPC package initialized with proper TypeScript configuration
- [ ] Context creation extracts and validates JWT from Authorization header
- [ ] Protected procedure middleware correctly rejects unauthenticated requests
- [ ] Public procedure allows unauthenticated access
- [ ] TRPCError utilities provide consistent error handling
- [ ] Root router can combine sub-routers
- [ ] Express adapter configured for `/trpc` endpoint
- [ ] SuperJSON transformer enabled for Date/Map/Set serialization

## Files to Create/Modify

- `packages/trpc/package.json` - Package configuration
- `packages/trpc/tsconfig.json` - TypeScript config
- `packages/trpc/src/trpc.ts` - tRPC initialization
- `packages/trpc/src/context.ts` - Auth context creation
- `packages/trpc/src/router.ts` - Root router
- `packages/trpc/src/middleware/auth.ts` - Protected procedure
- `packages/trpc/src/middleware/rateLimit.ts` - Rate limiting
- `packages/trpc/src/utils/errors.ts` - Error helpers
- `packages/trpc/src/index.ts` - Main exports

## Integration Points

- **Provides**: Base tRPC setup, publicProcedure, protectedProcedure, Context types
- **Consumes**: JWT validation (may need to call Go backend or implement standalone)
- **Conflicts**: None (new package)

## Notes

- Consider whether to validate JWT tokens by calling the Go backend or implementing standalone JWT validation
- The Go backend uses HS256 signing with a shared secret from config
- Existing auth pattern: `Authorization: Bearer <token>` header
