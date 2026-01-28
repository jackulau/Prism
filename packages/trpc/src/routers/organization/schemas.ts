import { z } from 'zod';

// User profile schema
export const userProfileSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  createdAt: z.date(),
  updatedAt: z.date(),
  githubConnected: z.boolean(),
  githubUsername: z.string().nullable(),
});

// User settings schema
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
  createdAt: z.date().nullable(),
});

export const setProviderKeyInput = z.object({
  provider: providerTypeSchema,
  apiKey: z.string().min(1),
});

export const providerIdInput = z.object({
  provider: providerTypeSchema,
});

export const validateProviderKeyInput = z.object({
  provider: providerTypeSchema,
  apiKey: z.string().min(1),
});

export const validateProviderKeyOutput = z.object({
  valid: z.boolean(),
  error: z.string().nullable(),
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

// Delete account
export const deleteAccountInput = z.object({
  password: z.string().min(8),
});

// Success response
export const successSchema = z.object({
  success: z.literal(true),
});

// Type exports
export type UserProfile = z.infer<typeof userProfileSchema>;
export type UserSettings = z.infer<typeof userSettingsSchema>;
export type UpdateSettingsInput = z.infer<typeof updateSettingsInput>;
export type ProviderKey = z.infer<typeof providerKeySchema>;
export type ProviderType = z.infer<typeof providerTypeSchema>;
export type GitHubConnection = z.infer<typeof githubConnectionSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordInput>;
export type ChangeEmailInput = z.infer<typeof changeEmailInput>;
export type DeleteAccountInput = z.infer<typeof deleteAccountInput>;
