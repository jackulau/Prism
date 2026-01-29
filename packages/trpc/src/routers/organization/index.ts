import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { router } from '../../trpc.js';
import { protectedProcedure } from '../../middleware/auth.js';
import { organizationService } from '../../services/organization.js';
import * as schemas from './schemas.js';

export const organizationRouter = router({
  // User Profile
  getProfile: protectedProcedure
    .output(schemas.userProfileSchema)
    .query(async ({ ctx }) => {
      const profile = await organizationService.getProfile(ctx.session.userId);
      if (!profile) {
        throw new TRPCError({
          code: 'NOT_FOUND',
          message: 'User not found',
        });
      }
      return profile;
    }),

  changePassword: protectedProcedure
    .input(schemas.changePasswordInput)
    .output(schemas.successSchema)
    .mutation(async ({ ctx, input }) => {
      try {
        await organizationService.changePassword(
          ctx.session.userId,
          input.currentPassword,
          input.newPassword
        );
        return { success: true as const };
      } catch (err) {
        throw new TRPCError({
          code: 'BAD_REQUEST',
          message: err instanceof Error ? err.message : 'Failed to change password',
        });
      }
    }),

  changeEmail: protectedProcedure
    .input(schemas.changeEmailInput)
    .output(schemas.successSchema)
    .mutation(async ({ ctx, input }) => {
      try {
        await organizationService.changeEmail(
          ctx.session.userId,
          input.newEmail,
          input.password
        );
        return { success: true as const };
      } catch (err) {
        throw new TRPCError({
          code: 'BAD_REQUEST',
          message: err instanceof Error ? err.message : 'Failed to change email',
        });
      }
    }),

  // User Settings
  getSettings: protectedProcedure
    .output(schemas.userSettingsSchema)
    .query(async ({ ctx }) => {
      const settings = await organizationService.getSettings(ctx.session.userId);
      return settings;
    }),

  updateSettings: protectedProcedure
    .input(schemas.updateSettingsInput)
    .output(schemas.userSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      const settings = await organizationService.updateSettings(
        ctx.session.userId,
        input
      );
      return settings;
    }),

  // Provider Keys
  listProviderKeys: protectedProcedure
    .output(z.array(schemas.providerKeySchema))
    .query(async ({ ctx }) => {
      const keys = await organizationService.listProviderKeys(ctx.session.userId);
      return keys;
    }),

  setProviderKey: protectedProcedure
    .input(schemas.setProviderKeyInput)
    .output(schemas.providerKeySchema)
    .mutation(async ({ ctx, input }) => {
      const key = await organizationService.setProviderKey(
        ctx.session.userId,
        input.provider,
        input.apiKey
      );
      return key;
    }),

  deleteProviderKey: protectedProcedure
    .input(schemas.providerIdInput)
    .output(schemas.successSchema)
    .mutation(async ({ ctx, input }) => {
      await organizationService.deleteProviderKey(
        ctx.session.userId,
        input.provider
      );
      return { success: true as const };
    }),

  validateProviderKey: protectedProcedure
    .input(schemas.validateProviderKeyInput)
    .output(schemas.validateProviderKeyOutput)
    .mutation(async ({ input }) => {
      const result = await organizationService.validateProviderKey(
        input.provider,
        input.apiKey
      );
      return result;
    }),

  // GitHub Connection
  getGitHubConnection: protectedProcedure
    .output(schemas.githubConnectionSchema)
    .query(async ({ ctx }) => {
      const connection = await organizationService.getGitHubConnection(
        ctx.session.userId
      );
      return connection;
    }),

  disconnectGitHub: protectedProcedure
    .output(schemas.successSchema)
    .mutation(async ({ ctx }) => {
      await organizationService.disconnectGitHub(ctx.session.userId);
      return { success: true as const };
    }),

  // Delete account
  deleteAccount: protectedProcedure
    .input(schemas.deleteAccountInput)
    .output(schemas.successSchema)
    .mutation(async ({ ctx, input }) => {
      try {
        await organizationService.deleteAccount(ctx.session.userId, input.password);
        return { success: true as const };
      } catch (err) {
        throw new TRPCError({
          code: 'BAD_REQUEST',
          message: err instanceof Error ? err.message : 'Failed to delete account',
        });
      }
    }),
});

export type OrganizationRouter = typeof organizationRouter;
