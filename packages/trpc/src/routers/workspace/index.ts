import { z } from 'zod';
import { router } from '../../trpc.js';
import { protectedProcedure } from '../../middleware/auth.js';
import { notFound } from '../../utils/errors.js';
import { workspaceService } from '../../services/workspace.js';
import {
  workspaceSchema,
  setDirectoryInput,
  setDirectoryOutput,
  browseDirectoryInput,
  browseDirectoryOutput,
  listRecentInput,
  workspaceIdInput,
  pickFolderOutput,
  successOutput,
} from './schemas.js';

export const workspaceRouter = router({
  getCurrent: protectedProcedure
    .output(z.object({ path: z.string() }).nullable())
    .query(async ({ ctx }) => {
      return workspaceService.getCurrent(ctx.session.userId, ctx.session.token);
    }),

  setDirectory: protectedProcedure
    .input(setDirectoryInput)
    .output(setDirectoryOutput)
    .mutation(async ({ ctx, input }) => {
      return workspaceService.setDirectory(
        ctx.session.userId,
        input.path,
        ctx.session.token
      );
    }),

  browse: protectedProcedure
    .input(browseDirectoryInput)
    .output(browseDirectoryOutput)
    .query(async ({ ctx, input }) => {
      return workspaceService.browse(
        input.path,
        input.showHidden,
        ctx.session.token
      );
    }),

  pickFolder: protectedProcedure.output(pickFolderOutput).mutation(async ({ ctx }) => {
    return workspaceService.pickFolder(ctx.session.token);
  }),

  listRecent: protectedProcedure
    .input(listRecentInput)
    .output(z.array(workspaceSchema))
    .query(async ({ ctx, input }) => {
      return workspaceService.listRecent(
        ctx.session.userId,
        input.limit,
        ctx.session.token
      );
    }),

  setCurrent: protectedProcedure
    .input(workspaceIdInput)
    .output(z.object({ success: z.boolean(), path: z.string() }))
    .mutation(async ({ ctx, input }) => {
      const result = await workspaceService.setCurrent(
        ctx.session.userId,
        input.id,
        ctx.session.token
      );
      if (!result.success) {
        throw notFound('Workspace');
      }
      return result;
    }),

  remove: protectedProcedure
    .input(workspaceIdInput)
    .output(successOutput)
    .mutation(async ({ ctx, input }) => {
      await workspaceService.delete(
        ctx.session.userId,
        input.id,
        ctx.session.token
      );
      return { success: true };
    }),

  getById: protectedProcedure
    .input(workspaceIdInput)
    .output(workspaceSchema.nullable())
    .query(async ({ ctx, input }) => {
      return workspaceService.getById(
        input.id,
        ctx.session.userId,
        ctx.session.token
      );
    }),
});
