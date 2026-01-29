import { z } from 'zod';
import { router } from '../../trpc.js';
import { protectedProcedure } from '../../middleware/auth.js';
import { integrationService } from '../../services/integration.js';
import * as schemas from './schemas.js';

export const integrationsRouter = router({
  // Get all integration statuses
  getStatus: protectedProcedure
    .output(schemas.integrationStatusSchema)
    .query(async ({ ctx }) => {
      return integrationService.getStatus(ctx.token);
    }),

  // Discord
  configureDiscord: protectedProcedure
    .input(schemas.discordConfigInput)
    .output(schemas.discordSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      return integrationService.setDiscord(
        ctx.token,
        input.webhookUrl,
        input.enabled
      );
    }),

  disconnectDiscord: protectedProcedure
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx }) => {
      await integrationService.deleteDiscord(ctx.token);
      return { success: true };
    }),

  // Slack
  configureSlack: protectedProcedure
    .input(schemas.slackConfigInput)
    .output(schemas.slackSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      return integrationService.setSlack(
        ctx.token,
        input.webhookUrl,
        input.channelId,
        input.enabled
      );
    }),

  disconnectSlack: protectedProcedure
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx }) => {
      await integrationService.deleteSlack(ctx.token);
      return { success: true };
    }),

  // PostHog
  configurePostHog: protectedProcedure
    .input(schemas.posthogConfigInput)
    .output(schemas.posthogSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      return integrationService.setPostHog(ctx.token, input.enabled);
    }),

  disconnectPostHog: protectedProcedure
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx }) => {
      await integrationService.deletePostHog(ctx.token);
      return { success: true };
    }),

  // MCP HTTP Servers
  listMcpServers: protectedProcedure
    .output(z.array(schemas.mcpServerSchema))
    .query(async ({ ctx }) => {
      return integrationService.listMcpServers(ctx.token);
    }),

  getMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(schemas.mcpServerSchema)
    .query(async ({ ctx, input }) => {
      return integrationService.getMcpServer(ctx.token, input.id);
    }),

  addMcpServer: protectedProcedure
    .input(schemas.mcpServerInput)
    .output(schemas.mcpServerSchema)
    .mutation(async ({ ctx, input }) => {
      return integrationService.addMcpServer(
        ctx.token,
        input.name,
        input.url,
        input.apiKey
      );
    }),

  updateMcpServer: protectedProcedure
    .input(schemas.mcpServerUpdateInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.updateMcpServer(ctx.token, input.id, {
        name: input.name,
        url: input.url,
        apiKey: input.apiKey,
      });
      return { success: true };
    }),

  removeMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.removeMcpServer(ctx.token, input.id);
      return { success: true };
    }),

  testMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean(), error: z.string().optional() }))
    .mutation(async ({ ctx, input }) => {
      return integrationService.testMcpServer(ctx.token, input.id);
    }),

  refreshMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.refreshMcpServer(ctx.token, input.id);
      return { success: true };
    }),

  enableMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.enableMcpServer(ctx.token, input.id);
      return { success: true };
    }),

  disableMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.disableMcpServer(ctx.token, input.id);
      return { success: true };
    }),

  listMcpTools: protectedProcedure
    .output(z.array(schemas.mcpToolSchema))
    .query(async ({ ctx }) => {
      return integrationService.listMcpTools(ctx.token);
    }),

  // MCP Stdio Servers
  listStdioServers: protectedProcedure
    .output(z.array(schemas.stdioMcpServerSchema))
    .query(async ({ ctx }) => {
      return integrationService.listStdioServers(ctx.token);
    }),

  getStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(schemas.stdioMcpServerSchema)
    .query(async ({ ctx, input }) => {
      return integrationService.getStdioServer(ctx.token, input.id);
    }),

  addStdioServer: protectedProcedure
    .input(schemas.stdioMcpServerInput)
    .output(schemas.stdioMcpServerSchema)
    .mutation(async ({ ctx, input }) => {
      return integrationService.addStdioServer(
        ctx.token,
        input.name,
        input.command,
        input.args,
        input.env
      );
    }),

  updateStdioServer: protectedProcedure
    .input(schemas.stdioMcpServerUpdateInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.updateStdioServer(ctx.token, input.id, {
        name: input.name,
        command: input.command,
        args: input.args,
        env: input.env,
        enabled: input.enabled,
      });
      return { success: true };
    }),

  removeStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.removeStdioServer(ctx.token, input.id);
      return { success: true };
    }),

  startStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.startStdioServer(ctx.token, input.id);
      return { success: true };
    }),

  stopStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.stopStdioServer(ctx.token, input.id);
      return { success: true };
    }),

  restartStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .output(z.object({ success: z.boolean() }))
    .mutation(async ({ ctx, input }) => {
      await integrationService.restartStdioServer(ctx.token, input.id);
      return { success: true };
    }),
});

export type IntegrationsRouter = typeof integrationsRouter;
