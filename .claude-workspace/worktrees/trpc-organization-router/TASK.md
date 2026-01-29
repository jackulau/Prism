---
id: trpc-organization-router
name: Organization Router Implementation
wave: 2
priority: 2
dependencies:
- trpc-core-setup
estimated_hours: 3
tags:
- backend
- api
- trpc
- organization
- settings
---

## Objective

Implement the tRPC organization router for managing organization/user settings, including user profile, provider keys, and application preferences.

## Context

The existing Go backend has user-scoped settings (no multi-tenant organizations yet). This router will handle:
- User profile management
- Provider API key management (OpenAI, Anthropic, Google, Ollama)
- User preferences (theme, default model, etc.)
- GitHub connection management

Existing Go endpoints:
- `GET /auth/me` - Get current user info
- `POST/DELETE /oauth/github/*` - GitHub connection
- Provider key management via separate handlers

## Implementation

### 1. Define Zod Schemas

**File: `packages/trpc/src/routers/organization/schemas.ts`**
```typescript
import { z } from 'zod';

// User profile
export const userProfileSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  createdAt: z.date(),
  updatedAt: z.date(),
  githubConnected: z.boolean(),
  githubUsername: z.string().nullable(),
});

// User settings
export const userSettingsSchema = z.object({
  theme: z.enum(['light', 'dark', 'system']).default('system'),
  defaultProvider: z.string().nullable(),
  defaultModel: z.string().nullable(),
  editorFontSize: z.number().min(10).max(24).default(14),
  editorTabSize: z.number().min(2).max(8).default(2),
});

export const updateSettingsInput = z.object({
  theme: z.enum(['light', 'dark', 'system']).optional(),
  defaultProvider: z.string().nullable().optional(),
  defaultModel: z.string().nullable().optional(),
  editorFontSize: z.number().min(10).max(24).optional(),
  editorTabSize: z.number().min(2).max(8).optional(),
});

// Provider key management
export const providerTypeSchema = z.enum([
  'openai',
  'anthropic',
  'google',
  'ollama',
  'openrouter',
]);

export const providerKeySchema = z.object({
  provider: providerTypeSchema,
  isConfigured: z.boolean(),
  lastUsedAt: z.date().nullable(),
  // API key is NEVER returned - only isConfigured status
});

export const setProviderKeyInput = z.object({
  provider: providerTypeSchema,
  apiKey: z.string().min(1),
});

export const providerIdInput = z.object({
  provider: providerTypeSchema,
});

// GitHub connection
export const githubConnectionSchema = z.object({
  connected: z.boolean(),
  username: z.string().nullable(),
  connectedAt: z.date().nullable(),
});

// Password change
export const changePasswordInput = z.object({
  currentPassword: z.string().min(8),
  newPassword: z.string().min(8),
});

// Email change
export const changeEmailInput = z.object({
  newEmail: z.string().email(),
  password: z.string().min(8),
});

export type UserProfile = z.infer<typeof userProfileSchema>;
export type UserSettings = z.infer<typeof userSettingsSchema>;
export type ProviderKey = z.infer<typeof providerKeySchema>;
export type ProviderType = z.infer<typeof providerTypeSchema>;
```

### 2. Implement Organization Router

**File: `packages/trpc/src/routers/organization/index.ts`**
```typescript
import { router, protectedProcedure } from '../../trpc';
import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import * as schemas from './schemas';

export const organizationRouter = router({
  // User Profile
  getProfile: protectedProcedure
    .output(schemas.userProfileSchema)
    .query(async ({ ctx }) => {
      const profile = await organizationService.getProfile(ctx.session.userId);
      if (!profile) {
        throw new TRPCError({ code: 'NOT_FOUND', message: 'User not found' });
      }
      return profile;
    }),

  changePassword: protectedProcedure
    .input(schemas.changePasswordInput)
    .mutation(async ({ ctx, input }) => {
      await organizationService.changePassword(
        ctx.session.userId,
        input.currentPassword,
        input.newPassword
      );
      return { success: true };
    }),

  changeEmail: protectedProcedure
    .input(schemas.changeEmailInput)
    .mutation(async ({ ctx, input }) => {
      await organizationService.changeEmail(
        ctx.session.userId,
        input.newEmail,
        input.password
      );
      return { success: true };
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
    .mutation(async ({ ctx, input }) => {
      await organizationService.deleteProviderKey(
        ctx.session.userId,
        input.provider
      );
      return { success: true };
    }),

  validateProviderKey: protectedProcedure
    .input(schemas.providerIdInput)
    .output(z.object({
      valid: z.boolean(),
      error: z.string().nullable(),
    }))
    .mutation(async ({ ctx, input }) => {
      const result = await organizationService.validateProviderKey(
        ctx.session.userId,
        input.provider
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
    .mutation(async ({ ctx }) => {
      await organizationService.disconnectGitHub(ctx.session.userId);
      return { success: true };
    }),

  // Delete account
  deleteAccount: protectedProcedure
    .input(z.object({ password: z.string().min(8) }))
    .mutation(async ({ ctx, input }) => {
      await organizationService.deleteAccount(ctx.session.userId, input.password);
      return { success: true };
    }),
});
```

### 3. Create Organization Service

**File: `packages/trpc/src/services/organization.ts`**
```typescript
import type { UserProfile, UserSettings, ProviderKey, ProviderType } from '../routers/organization/schemas';

export const organizationService = {
  // Profile
  async getProfile(userId: string): Promise<UserProfile | null> {},
  async changePassword(userId: string, currentPassword: string, newPassword: string) {
    // Verify current password
    // Hash new password with Argon2id (match Go implementation)
    // Update password
  },
  async changeEmail(userId: string, newEmail: string, password: string) {
    // Verify password
    // Check email not taken
    // Update email
  },

  // Settings
  async getSettings(userId: string): Promise<UserSettings> {
    // Return settings with defaults
  },
  async updateSettings(userId: string, updates: Partial<UserSettings>): Promise<UserSettings> {},

  // Provider Keys
  async listProviderKeys(userId: string): Promise<ProviderKey[]> {
    // Return list with isConfigured status only
    // Never expose actual API keys
  },
  async setProviderKey(userId: string, provider: ProviderType, apiKey: string): Promise<ProviderKey> {
    // Encrypt API key before storage
    // Use same encryption as Go backend
  },
  async deleteProviderKey(userId: string, provider: ProviderType) {},
  async validateProviderKey(userId: string, provider: ProviderType) {
    // Make test API call to validate key
  },

  // GitHub
  async getGitHubConnection(userId: string) {},
  async disconnectGitHub(userId: string) {},

  // Account
  async deleteAccount(userId: string, password: string) {
    // Verify password
    // Delete all user data
    // Cascade delete: workspaces, settings, sessions, etc.
  },
};
```

## Acceptance Criteria

- [ ] User profile retrieval working
- [ ] Password change with verification
- [ ] Email change with verification
- [ ] User settings CRUD working
- [ ] Provider key management working (encrypt/decrypt)
- [ ] API keys never exposed in responses
- [ ] Provider key validation working
- [ ] GitHub connection status working
- [ ] Account deletion with cascade

## Files to Create/Modify

- `packages/trpc/src/routers/organization/schemas.ts` - Zod schemas
- `packages/trpc/src/routers/organization/index.ts` - Router implementation
- `packages/trpc/src/services/organization.ts` - Organization service
- `packages/trpc/src/router.ts` - Add organization router (modify)

## Integration Points

- **Provides**: Organization/user settings via tRPC
- **Consumes**: trpc-core-setup (protectedProcedure, router, context)
- **Conflicts**: Avoid modifying Go auth/user handlers

## Notes

- Existing Go tables: users, user_settings, provider_keys, sessions
- Password hashing uses Argon2id (memory=64MB, time=1, threads=4)
- API keys encrypted with AES-256-GCM
- GitHub OAuth flow handled separately (existing /oauth/github endpoints)
