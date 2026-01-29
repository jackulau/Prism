import { z } from 'zod';
import { TRPCError } from '@trpc/server';
import { router } from '../../trpc.js';
import { protectedProcedure } from '../../middleware/auth.js';
import { workerService } from '../../services/worker.js';
import * as schemas from './schemas.js';

export const workersRouter = router({
  /**
   * Run a single agent task
   */
  runTask: protectedProcedure
    .input(schemas.runTaskInput)
    .output(schemas.executionSchema)
    .mutation(async ({ ctx, input }) => {
      const execution = await workerService.runTask(
        ctx.session.userId,
        input.task,
        input.config
      );
      return execution;
    }),

  /**
   * Run multiple agent tasks in parallel
   */
  runParallel: protectedProcedure
    .input(schemas.runParallelInput)
    .output(schemas.executionSchema)
    .mutation(async ({ ctx, input }) => {
      const execution = await workerService.runParallel(
        ctx.session.userId,
        input.tasks,
        input.config
      );
      return execution;
    }),

  /**
   * Run multiple agent tasks sequentially
   */
  runSequential: protectedProcedure
    .input(schemas.runSequentialInput)
    .output(schemas.executionSchema)
    .mutation(async ({ ctx, input }) => {
      const execution = await workerService.runSequential(
        ctx.session.userId,
        input.tasks,
        input.config
      );
      return execution;
    }),

  /**
   * Run a multi-agent swarm
   */
  runSwarm: protectedProcedure
    .input(schemas.runSwarmInput)
    .output(schemas.swarmSchema)
    .mutation(async ({ ctx, input }) => {
      const swarm = await workerService.runSwarm(
        ctx.session.userId,
        input.prompt,
        input.name,
        input.strategy,
        input.roles,
        input.config,
        input.timeout
      );
      return swarm;
    }),

  /**
   * Get execution status by ID
   */
  getExecution: protectedProcedure
    .input(schemas.executionIdInput)
    .output(schemas.executionSchema.nullable())
    .query(async ({ input }) => {
      const execution = await workerService.getExecution(input.executionId);
      return execution;
    }),

  /**
   * List all executions for the current user
   */
  listExecutions: protectedProcedure
    .output(z.array(schemas.executionSchema))
    .query(async ({ ctx }) => {
      const executions = await workerService.listExecutions(ctx.session.userId);
      return executions;
    }),

  /**
   * Cancel a running execution
   */
  cancelExecution: protectedProcedure
    .input(schemas.executionIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ input }) => {
      try {
        await workerService.cancelExecution(input.executionId);
        return { success: true };
      } catch (error) {
        throw new TRPCError({
          code: 'BAD_REQUEST',
          message:
            error instanceof Error ? error.message : 'Failed to cancel execution',
        });
      }
    }),

  /**
   * Get swarm status by ID
   */
  getSwarm: protectedProcedure
    .input(schemas.swarmIdInput)
    .output(schemas.swarmSchema.nullable())
    .query(async ({ input }) => {
      const swarm = await workerService.getSwarm(input.swarmId);
      return swarm;
    }),

  /**
   * List all swarms for the current user
   */
  listSwarms: protectedProcedure
    .output(z.array(schemas.swarmSchema))
    .query(async ({ ctx }) => {
      const swarms = await workerService.listSwarms(ctx.session.userId);
      return swarms;
    }),

  /**
   * Cancel a running swarm
   */
  cancelSwarm: protectedProcedure
    .input(schemas.swarmIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ input }) => {
      try {
        await workerService.cancelSwarm(input.swarmId);
        return { success: true };
      } catch (error) {
        throw new TRPCError({
          code: 'BAD_REQUEST',
          message:
            error instanceof Error ? error.message : 'Failed to cancel swarm',
        });
      }
    }),

  /**
   * Get manager statistics
   */
  getStats: protectedProcedure
    .output(schemas.managerStatsSchema)
    .query(async ({ ctx }) => {
      const stats = await workerService.getStats(ctx.session.userId);
      return stats;
    }),
});

export type WorkersRouter = typeof workersRouter;
