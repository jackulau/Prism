import { router } from './trpc.js';
import { paymentRouter } from './routers/payment/index.js';

export const appRouter = router({
  payment: paymentRouter,
});

export type AppRouter = typeof appRouter;
