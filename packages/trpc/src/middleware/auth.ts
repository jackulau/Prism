import { TRPCError } from '@trpc/server';
import { middleware, publicProcedure } from '../trpc.js';

export const isAuthed = middleware(async ({ ctx, next }) => {
  if (!ctx.session || !ctx.token) {
    throw new TRPCError({ code: 'UNAUTHORIZED' });
  }
  return next({
    ctx: {
      session: ctx.session,
      token: ctx.token,
    },
  });
});

export const protectedProcedure = publicProcedure.use(isAuthed);
