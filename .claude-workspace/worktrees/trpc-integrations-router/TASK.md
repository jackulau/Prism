---
id: trpc-integrations-router
name: Integrations Router Implementation
wave: 2
priority: 2
dependencies:
- trpc-core-setup
estimated_hours: 3
tags:
- backend
- api
- trpc
- integrations
---

## Objective

Implement the tRPC integrations router for managing external service connections (Discord, Slack, PostHog, MCP servers).

## Context

The existing Go backend has integration management at `/api/v1/integrations/`. This includes:
- Discord webhook configuration
- Slack webhook/bot configuration
- PostHog analytics toggle
- MCP (Model Context Protocol) server connections

Existing Go endpoints:
- `GET /integrations/status` - Get all integration statuses
- `POST /integrations/discord` - Configure Discord
- `DELETE /integrations/discord` - Disconnect Discord
- `POST /integrations/slack` - Configure Slack
- `DELETE /integrations/slack` - Disconnect Slack
- `POST /integrations/posthog` - Enable PostHog
- `DELETE /integrations/posthog` - Disable PostHog
- `GET/POST /mcp/servers` - List/configure MCP servers
- `DELETE /mcp/servers/:id` - Remove MCP connection

## Implementation

### 1. Define Zod Schemas

**File: `packages/trpc/src/routers/integrations/schemas.ts`**
```typescript
import { z } from 'zod';

// Integration types
export const integrationTypeSchema = z.enum([
  'discord',
  'slack',
  'posthog',
  'mcp',
]);

// Base integration settings
export const baseIntegrationSchema = z.object({
  enabled: z.boolean(),
  connectedAt: z.date().nullable(),
});

// Discord integration
export const discordSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('discord'),
  webhookUrl: z.string().url().optional(),
});

export const discordConfigInput = z.object({
  webhookUrl: z.string().url(),
  enabled: z.boolean().default(true),
});

// Slack integration
export const slackSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('slack'),
  webhookUrl: z.string().url().optional(),
  channelId: z.string().optional(),
  botToken: z.string().optional(), // Not exposed in response
});

export const slackConfigInput = z.object({
  webhookUrl: z.string().url(),
  channelId: z.string().optional(),
  botToken: z.string().optional(),
  enabled: z.boolean().default(true),
});

// PostHog integration
export const posthogSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('posthog'),
});

export const posthogConfigInput = z.object({
  enabled: z.boolean(),
});

// MCP Server integration
export const mcpServerSchema = z.object({
  id: z.string(),
  name: z.string(),
  url: z.string().url(),
  apiKey: z.string().optional(), // Not exposed in list
  enabled: z.boolean(),
  lastConnectedAt: z.date().nullable(),
  toolCount: z.number().optional(),
});

export const mcpServerInput = z.object({
  name: z.string().min(1),
  url: z.string().url(),
  apiKey: z.string().optional(),
});

export const mcpServerIdInput = z.object({
  id: z.string(),
});

// Stdio MCP Server
export const stdioMcpServerSchema = z.object({
  id: z.string(),
  name: z.string(),
  command: z.string(),
  args: z.array(z.string()),
  env: z.record(z.string()).optional(),
  enabled: z.boolean(),
});

export const stdioMcpServerInput = z.object({
  name: z.string().min(1),
  command: z.string().min(1),
  args: z.array(z.string()).default([]),
  env: z.record(z.string()).optional(),
});

// Combined status
export const integrationStatusSchema = z.object({
  discord: discordSettingsSchema.nullable(),
  slack: slackSettingsSchema.nullable(),
  posthog: posthogSettingsSchema.nullable(),
  mcpServers: z.array(mcpServerSchema),
  stdioServers: z.array(stdioMcpServerSchema),
});

export type IntegrationType = z.infer<typeof integrationTypeSchema>;
export type IntegrationStatus = z.infer<typeof integrationStatusSchema>;
export type MCPServer = z.infer<typeof mcpServerSchema>;
```

### 2. Implement Integrations Router

**File: `packages/trpc/src/routers/integrations/index.ts`**
```typescript
import { router, protectedProcedure } from '../../trpc';
import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import * as schemas from './schemas';

export const integrationsRouter = router({
  // Get all integration statuses
  getStatus: protectedProcedure
    .output(schemas.integrationStatusSchema)
    .query(async ({ ctx }) => {
      const status = await integrationService.getAllStatus(ctx.session.userId);
      return status;
    }),

  // Discord
  configureDiscord: protectedProcedure
    .input(schemas.discordConfigInput)
    .output(schemas.discordSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      const settings = await integrationService.setDiscord(
        ctx.session.userId,
        input.webhookUrl,
        input.enabled
      );
      return settings;
    }),

  disconnectDiscord: protectedProcedure
    .mutation(async ({ ctx }) => {
      await integrationService.deleteDiscord(ctx.session.userId);
      return { success: true };
    }),

  // Slack
  configureSlack: protectedProcedure
    .input(schemas.slackConfigInput)
    .output(schemas.slackSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      const settings = await integrationService.setSlack(
        ctx.session.userId,
        input.webhookUrl,
        input.channelId,
        input.botToken,
        input.enabled
      );
      return settings;
    }),

  disconnectSlack: protectedProcedure
    .mutation(async ({ ctx }) => {
      await integrationService.deleteSlack(ctx.session.userId);
      return { success: true };
    }),

  // PostHog
  configurePostHog: protectedProcedure
    .input(schemas.posthogConfigInput)
    .output(schemas.posthogSettingsSchema)
    .mutation(async ({ ctx, input }) => {
      const settings = await integrationService.setPostHog(
        ctx.session.userId,
        input.enabled
      );
      return settings;
    }),

  disconnectPostHog: protectedProcedure
    .mutation(async ({ ctx }) => {
      await integrationService.deletePostHog(ctx.session.userId);
      return { success: true };
    }),

  // MCP HTTP Servers
  listMcpServers: protectedProcedure
    .output(z.array(schemas.mcpServerSchema))
    .query(async ({ ctx }) => {
      const servers = await integrationService.listMcpServers(ctx.session.userId);
      return servers;
    }),

  addMcpServer: protectedProcedure
    .input(schemas.mcpServerInput)
    .output(schemas.mcpServerSchema)
    .mutation(async ({ ctx, input }) => {
      const server = await integrationService.addMcpServer(
        ctx.session.userId,
        input.name,
        input.url,
        input.apiKey
      );
      return server;
    }),

  removeMcpServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .mutation(async ({ ctx, input }) => {
      await integrationService.removeMcpServer(ctx.session.userId, input.id);
      return { success: true };
    }),

  // MCP Stdio Servers
  listStdioServers: protectedProcedure
    .output(z.array(schemas.stdioMcpServerSchema))
    .query(async ({ ctx }) => {
      const servers = await integrationService.listStdioServers(ctx.session.userId);
      return servers;
    }),

  addStdioServer: protectedProcedure
    .input(schemas.stdioMcpServerInput)
    .output(schemas.stdioMcpServerSchema)
    .mutation(async ({ ctx, input }) => {
      const server = await integrationService.addStdioServer(
        ctx.session.userId,
        input.name,
        input.command,
        input.args,
        input.env
      );
      return server;
    }),

  removeStdioServer: protectedProcedure
    .input(schemas.mcpServerIdInput)
    .mutation(async ({ ctx, input }) => {
      await integrationService.removeStdioServer(ctx.session.userId, input.id);
      return { success: true };
    }),
});
```

### 3. Create Integration Service

**File: `packages/trpc/src/services/integration.ts`**
```typescript
import type { IntegrationStatus, MCPServer, StdioMcpServer } from '../routers/integrations/schemas';

export const integrationService = {
  async getAllStatus(userId: string): Promise<IntegrationStatus> {
    // Aggregate all integration settings
  },

  // Discord
  async setDiscord(userId: string, webhookUrl: string, enabled: boolean) {
    // Encrypt webhook URL before storage
  },
  async deleteDiscord(userId: string) {},

  // Slack
  async setSlack(
    userId: string,
    webhookUrl: string,
    channelId?: string,
    botToken?: string,
    enabled: boolean = true
  ) {
    // Encrypt sensitive fields
  },
  async deleteSlack(userId: string) {},

  // PostHog
  async setPostHog(userId: string, enabled: boolean) {},
  async deletePostHog(userId: string) {},

  // MCP HTTP Servers
  async listMcpServers(userId: string): Promise<MCPServer[]> {},
  async addMcpServer(userId: string, name: string, url: string, apiKey?: string): Promise<MCPServer> {},
  async removeMcpServer(userId: string, id: string) {},

  // MCP Stdio Servers
  async listStdioServers(userId: string): Promise<StdioMcpServer[]> {},
  async addStdioServer(
    userId: string,
    name: string,
    command: string,
    args: string[],
    env?: Record<string, string>
  ): Promise<StdioMcpServer> {},
  async removeStdioServer(userId: string, id: string) {},
};
```

## Acceptance Criteria

- [ ] All integration Zod schemas defined
- [ ] Discord webhook configuration working
- [ ] Slack webhook/bot configuration working
- [ ] PostHog analytics toggle working
- [ ] MCP HTTP server management working
- [ ] MCP Stdio server management working
- [ ] Sensitive data (webhooks, tokens) encrypted at rest
- [ ] Integration status aggregation working
- [ ] User isolation enforced

## Files to Create/Modify

- `packages/trpc/src/routers/integrations/schemas.ts` - Zod schemas
- `packages/trpc/src/routers/integrations/index.ts` - Router implementation
- `packages/trpc/src/services/integration.ts` - Integration service
- `packages/trpc/src/router.ts` - Add integrations router (modify)

## Integration Points

- **Provides**: Integration management via tRPC
- **Consumes**: trpc-core-setup (protectedProcedure, router, context)
- **Conflicts**: Avoid modifying Go integration handlers

## Notes

- Existing Go tables: discord_settings, slack_settings, posthog_settings, mcp_connections, mcp_stdio_servers
- Webhook URLs and tokens are encrypted in Go using AES-256-GCM
- Need to use same encryption key/scheme for compatibility
