import { TRPCError } from '@trpc/server';
import { middleware, publicProcedure } from '../trpc.js';

export const isAuthed = middleware(async ({ ctx, next }) => {
  if (!ctx.session) {
    throw new TRPCError({ code: 'UNAUTHORIZED' });
  }
  return next({
    ctx: {
      session: ctx.session,
    },
  });
});

export const protectedProcedure = publicProcedure.use(isAuthed);
