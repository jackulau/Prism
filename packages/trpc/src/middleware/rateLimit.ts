import { TRPCError } from '@trpc/server';
import { middleware } from '../trpc.js';

interface RateLimitEntry {
  count: number;
  resetAt: number;
}

const rateLimitStore = new Map<string, RateLimitEntry>();

interface RateLimitOptions {
  windowMs?: number;
  maxRequests?: number;
}

export function createRateLimitMiddleware(options: RateLimitOptions = {}) {
  const { windowMs = 60000, maxRequests = 60 } = options;

  return middleware(async ({ ctx, next }) => {
    const identifier = ctx.session?.userId ?? 'anonymous';
    const now = Date.now();

    let entry = rateLimitStore.get(identifier);

    if (!entry || now > entry.resetAt) {
      entry = { count: 0, resetAt: now + windowMs };
      rateLimitStore.set(identifier, entry);
    }

    entry.count++;

    if (entry.count > maxRequests) {
      throw new TRPCError({
        code: 'TOO_MANY_REQUESTS',
        message: 'Rate limit exceeded. Please try again later.',
      });
    }

    return next();
  });
}

export const rateLimit = createRateLimitMiddleware();

setInterval(() => {
  const now = Date.now();
  for (const [key, entry] of rateLimitStore.entries()) {
    if (now > entry.resetAt) {
      rateLimitStore.delete(key);
    }
  }
}, 60000);
