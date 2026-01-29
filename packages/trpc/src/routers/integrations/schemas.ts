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
  connected: z.boolean(),
});

// Discord integration
export const discordSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('discord'),
});

export const discordConfigInput = z.object({
  webhookUrl: z.string().url(),
  enabled: z.boolean().default(true),
});

// Slack integration
export const slackSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('slack'),
  channelId: z.string().optional(),
});

export const slackConfigInput = z.object({
  webhookUrl: z.string().url(),
  channelId: z.string().optional(),
  enabled: z.boolean().default(true),
});

// PostHog integration
export const posthogSettingsSchema = baseIntegrationSchema.extend({
  type: z.literal('posthog'),
});

export const posthogConfigInput = z.object({
  enabled: z.boolean(),
});

// MCP HTTP Server integration
export const mcpServerSchema = z.object({
  id: z.string(),
  name: z.string(),
  url: z.string().url(),
  enabled: z.boolean(),
  hasApiKey: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
  lastSync: z.string().nullable(),
  lastError: z.string().nullable(),
  manifest: z
    .object({
      name: z.string(),
      version: z.string(),
      description: z.string().optional(),
      toolCount: z.number(),
    })
    .nullable(),
});

export const mcpServerInput = z.object({
  name: z.string().min(1),
  url: z.string().url(),
  apiKey: z.string().optional(),
});

export const mcpServerIdInput = z.object({
  id: z.string(),
});

export const mcpServerUpdateInput = z.object({
  id: z.string(),
  name: z.string().optional(),
  url: z.string().url().optional(),
  apiKey: z.string().optional(),
});

// MCP Stdio Server integration
export const stdioMcpServerSchema = z.object({
  id: z.string(),
  name: z.string(),
  command: z.string(),
  args: z.array(z.string()),
  env: z.array(z.string()).optional(),
  enabled: z.boolean(),
  running: z.boolean(),
  toolCount: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  lastError: z.string().nullable(),
});

export const stdioMcpServerInput = z.object({
  name: z.string().min(1),
  command: z.string().min(1),
  args: z.array(z.string()).default([]),
  env: z.array(z.string()).optional(),
});

export const stdioMcpServerUpdateInput = z.object({
  id: z.string(),
  name: z.string().optional(),
  command: z.string().optional(),
  args: z.array(z.string()).optional(),
  env: z.array(z.string()).optional(),
  enabled: z.boolean().optional(),
});

// Combined status response
export const integrationStatusSchema = z.object({
  discord: discordSettingsSchema,
  slack: slackSettingsSchema,
  posthog: posthogSettingsSchema,
});

// MCP tool schema
export const mcpToolSchema = z.object({
  serverId: z.string(),
  serverName: z.string(),
  name: z.string(),
  description: z.string().optional(),
  parameters: z.record(z.unknown()).optional(),
});

// Type exports
export type IntegrationType = z.infer<typeof integrationTypeSchema>;
export type IntegrationStatus = z.infer<typeof integrationStatusSchema>;
export type DiscordSettings = z.infer<typeof discordSettingsSchema>;
export type SlackSettings = z.infer<typeof slackSettingsSchema>;
export type PostHogSettings = z.infer<typeof posthogSettingsSchema>;
export type MCPServer = z.infer<typeof mcpServerSchema>;
export type StdioMCPServer = z.infer<typeof stdioMcpServerSchema>;
export type MCPTool = z.infer<typeof mcpToolSchema>;
