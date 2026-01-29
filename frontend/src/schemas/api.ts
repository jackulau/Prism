/**
 * Zod validation schemas for API responses.
 * These schemas provide runtime validation for API responses to catch
 * type mismatches between backend and frontend.
 */

import { z } from 'zod';

// ============================================================================
// Base Schemas
// ============================================================================

/** Schema for error responses */
export const errorResponseSchema = z.object({
  error: z.string(),
  code: z.string().optional(),
  details: z.string().optional(),
});

/** Schema for success responses */
export const successResponseSchema = z.object({
  success: z.boolean(),
  message: z.string().optional(),
});

/** Factory for API response wrapper */
export function apiResponseSchema<T extends z.ZodType>(dataSchema: T) {
  return z.object({
    data: dataSchema.optional(),
    error: z.string().optional(),
  });
}

/** Factory for paginated response wrapper */
export function paginatedResponseSchema<T extends z.ZodType>(itemSchema: T) {
  return z.object({
    items: z.array(itemSchema),
    total: z.number(),
    limit: z.number(),
    offset: z.number(),
    has_more: z.boolean(),
  });
}

// ============================================================================
// Auth Schemas
// ============================================================================

export const userDTOSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  created_at: z.string(),
});

export const authResponseSchema = z.object({
  access_token: z.string(),
  refresh_token: z.string(),
  expires_at: z.string(),
  user: userDTOSchema,
});

export const registerRequestSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

export const loginRequestSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
});

export const refreshRequestSchema = z.object({
  refresh_token: z.string(),
});

// ============================================================================
// Conversation Schemas
// ============================================================================

export const conversationDTOSchema = z.object({
  id: z.string(),
  title: z.string(),
  provider: z.string(),
  model: z.string(),
  system_prompt: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const messageDTOSchema = z.object({
  id: z.string(),
  role: z.string(),
  content: z.string(),
  tool_calls: z.array(z.record(z.unknown())).optional(),
  tool_call_id: z.string().optional(),
  created_at: z.string(),
});

export const createConversationRequestSchema = z.object({
  provider: z.string().min(1, 'Provider is required'),
  model: z.string().min(1, 'Model is required'),
  system_prompt: z.string().optional(),
});

export const updateConversationRequestSchema = z.object({
  title: z.string().min(1, 'Title is required'),
});

export const conversationListResponseSchema = z.object({
  conversations: z.array(conversationDTOSchema),
});

export const messageListResponseSchema = z.object({
  messages: z.array(messageDTOSchema),
});

// ============================================================================
// Provider Schemas
// ============================================================================

export const setKeyRequestSchema = z.object({
  api_key: z.string().min(1, 'API key is required'),
});

export const validateKeyRequestSchema = z.object({
  api_key: z.string().min(1, 'API key is required'),
});

export const keyStatusResponseSchema = z.object({
  has_key: z.boolean(),
  provider: z.string(),
});

export const keyValidationResponseSchema = z.object({
  valid: z.boolean(),
  message: z.string().optional(),
});

export const providerListResponseSchema = z.object({
  providers: z.array(z.string()),
});

// ============================================================================
// Workspace Schemas
// ============================================================================

export const directoryResponseSchema = z.object({
  path: z.string(),
});

export const setDirectoryRequestSchema = z.object({
  directory: z.string().min(1, 'Directory is required'),
});

export const browseDirectoryEntrySchema = z.object({
  name: z.string(),
  path: z.string(),
});

export const browseDirectoriesResponseSchema = z.object({
  current_path: z.string(),
  parent_path: z.string(),
  directories: z.array(browseDirectoryEntrySchema),
});

export const workspaceInfoSchema = z.object({
  id: z.string(),
  path: z.string(),
  name: z.string(),
  is_current: z.boolean(),
  last_accessed_at: z.string().optional(),
});

export const workspaceListResponseSchema = z.object({
  workspaces: z.array(workspaceInfoSchema),
});

export const cloneRepoRequestSchema = z.object({
  repo_url: z.string().url('Invalid repository URL'),
  branch: z.string().optional(),
});

export const cloneRepoResponseSchema = z.object({
  success: z.boolean(),
  path: z.string(),
  message: z.string(),
});

// ============================================================================
// File Schemas
// ============================================================================

/** Base file node schema (non-recursive for simpler validation) */
const baseFileNodeSchema = z.object({
  name: z.string(),
  path: z.string(),
  type: z.enum(['file', 'directory']),
});

/** File node schema with optional children (uses any for deeply nested to avoid recursive type issues) */
export const fileNodeSchema = baseFileNodeSchema.extend({
  children: z.array(z.any()).optional(),
});

export const fileListResponseSchema = z.object({
  files: z.array(fileNodeSchema),
});

export const fileContentResponseSchema = z.object({
  path: z.string(),
  content: z.string(),
});

export const writeFileRequestSchema = z.object({
  path: z.string().min(1, 'File path is required'),
  content: z.string(),
});

export const renameFileRequestSchema = z.object({
  source_path: z.string().min(1, 'Source path is required'),
  dest_path: z.string().min(1, 'Destination path is required'),
});

export const createDirectoryRequestSchema = z.object({
  path: z.string().min(1, 'Directory path is required'),
});

// ============================================================================
// Integration Schemas
// ============================================================================

export const integrationStatusSchema = z.object({
  enabled: z.boolean(),
  connected: z.boolean(),
  channel_id: z.string().optional(),
});

export const integrationStatusResponseSchema = z.object({
  discord: integrationStatusSchema,
  slack: integrationStatusSchema,
  posthog: integrationStatusSchema,
});

export const setIntegrationRequestSchema = z.object({
  webhook_url: z.string().optional(),
  channel_id: z.string().optional(),
  enabled: z.boolean(),
});

// ============================================================================
// Build Schemas
// ============================================================================

export const buildStatusSchema = z.enum(['pending', 'running', 'success', 'failed']);

export const buildInfoSchema = z.object({
  id: z.string(),
  status: buildStatusSchema,
  command: z.string(),
  start_time: z.string(),
  end_time: z.string().optional(),
  duration_ms: z.number().optional(),
  error: z.string().optional(),
  preview_url: z.string().optional(),
});

// ============================================================================
// GitHub Schemas
// ============================================================================

export const gitHubStatusResponseSchema = z.object({
  connected: z.boolean(),
  username: z.string().optional(),
  avatar_url: z.string().optional(),
});

export const gitHubRepoInfoSchema = z.object({
  id: z.number(),
  name: z.string(),
  full_name: z.string(),
  description: z.string().optional(),
  private: z.boolean(),
  html_url: z.string(),
  clone_url: z.string(),
  default_branch: z.string(),
});

export const gitHubReposResponseSchema = z.object({
  repos: z.array(gitHubRepoInfoSchema),
});

// ============================================================================
// WebSocket Schemas
// ============================================================================

export const wsMessageTypeSchema = z.enum([
  'chat.message',
  'chat.chunk',
  'chat.complete',
  'chat.stop',
  'tool.started',
  'tool.completed',
  'tool.confirm',
  'agent.check_in',
  'agent.run',
  'agent.run_parallel',
  'agent.stop',
  'swarm.run',
  'swarm.stop',
  'build.start',
  'build.output',
  'build.complete',
  'build.stop',
  'preview.ready',
  'preview.content',
  'files.updated',
  'file.request',
  'file.history_request',
  'error',
]);

export const fileContextSchema = z.object({
  path: z.string(),
  content: z.string(),
  language: z.string().optional(),
});

export const chatModeSchema = z.enum(['plan', 'ask-before-edits', 'edit-automatically']);

export const chatMessagePayloadSchema = z.object({
  content: z.string(),
  conversation_id: z.string().optional(),
  model: z.string().optional(),
  provider: z.string().optional(),
  system_prompt: z.string().optional(),
  mode: chatModeSchema.optional(),
  enable_thinking: z.boolean().optional(),
  file_context: z.array(fileContextSchema).optional(),
  mcp_tools: z.array(z.string()).optional(),
});

export const messageMetricsSchema = z.object({
  input_tokens: z.number().optional(),
  output_tokens: z.number().optional(),
  thinking_tokens: z.number().optional(),
  total_tokens: z.number().optional(),
  latency_ms: z.number().optional(),
});

export const chatChunkPayloadSchema = z.object({
  content: z.string(),
  conversation_id: z.string(),
  thinking: z.boolean().optional(),
});

export const chatCompletePayloadSchema = z.object({
  conversation_id: z.string(),
  message_id: z.string(),
  metrics: messageMetricsSchema.optional(),
});

export const toolStartedPayloadSchema = z.object({
  tool_call_id: z.string(),
  tool_name: z.string(),
  parameters: z.record(z.unknown()),
});

export const toolCompletedPayloadSchema = z.object({
  tool_call_id: z.string(),
  tool_name: z.string(),
  result: z.string(),
  success: z.boolean(),
});

export const toolConfirmPayloadSchema = z.object({
  tool_call_id: z.string(),
  approved: z.boolean(),
});

export const errorPayloadSchema = z.object({
  message: z.string(),
  code: z.string().optional(),
});

// ============================================================================
// Type Exports
// ============================================================================

export type ErrorResponse = z.infer<typeof errorResponseSchema>;
export type SuccessResponse = z.infer<typeof successResponseSchema>;
export type UserDTO = z.infer<typeof userDTOSchema>;
export type AuthResponse = z.infer<typeof authResponseSchema>;
export type RegisterRequest = z.infer<typeof registerRequestSchema>;
export type LoginRequest = z.infer<typeof loginRequestSchema>;
export type RefreshRequest = z.infer<typeof refreshRequestSchema>;
export type ConversationDTO = z.infer<typeof conversationDTOSchema>;
export type MessageDTO = z.infer<typeof messageDTOSchema>;
export type CreateConversationRequest = z.infer<typeof createConversationRequestSchema>;
export type UpdateConversationRequest = z.infer<typeof updateConversationRequestSchema>;
export type ConversationListResponse = z.infer<typeof conversationListResponseSchema>;
export type MessageListResponse = z.infer<typeof messageListResponseSchema>;
export type SetKeyRequest = z.infer<typeof setKeyRequestSchema>;
export type ValidateKeyRequest = z.infer<typeof validateKeyRequestSchema>;
export type KeyStatusResponse = z.infer<typeof keyStatusResponseSchema>;
export type KeyValidationResponse = z.infer<typeof keyValidationResponseSchema>;
export type ProviderListResponse = z.infer<typeof providerListResponseSchema>;
export type DirectoryResponse = z.infer<typeof directoryResponseSchema>;
export type SetDirectoryRequest = z.infer<typeof setDirectoryRequestSchema>;
export type BrowseDirectoryEntry = z.infer<typeof browseDirectoryEntrySchema>;
export type BrowseDirectoriesResponse = z.infer<typeof browseDirectoriesResponseSchema>;
export type WorkspaceInfo = z.infer<typeof workspaceInfoSchema>;
export type WorkspaceListResponse = z.infer<typeof workspaceListResponseSchema>;
export type CloneRepoRequest = z.infer<typeof cloneRepoRequestSchema>;
export type CloneRepoResponse = z.infer<typeof cloneRepoResponseSchema>;
export type FileNode = z.infer<typeof fileNodeSchema>;
export type FileListResponse = z.infer<typeof fileListResponseSchema>;
export type FileContentResponse = z.infer<typeof fileContentResponseSchema>;
export type WriteFileRequest = z.infer<typeof writeFileRequestSchema>;
export type RenameFileRequest = z.infer<typeof renameFileRequestSchema>;
export type CreateDirectoryRequest = z.infer<typeof createDirectoryRequestSchema>;
export type IntegrationStatus = z.infer<typeof integrationStatusSchema>;
export type IntegrationStatusResponse = z.infer<typeof integrationStatusResponseSchema>;
export type SetIntegrationRequest = z.infer<typeof setIntegrationRequestSchema>;
export type BuildStatus = z.infer<typeof buildStatusSchema>;
export type BuildInfo = z.infer<typeof buildInfoSchema>;
export type GitHubStatusResponse = z.infer<typeof gitHubStatusResponseSchema>;
export type GitHubRepoInfo = z.infer<typeof gitHubRepoInfoSchema>;
export type GitHubReposResponse = z.infer<typeof gitHubReposResponseSchema>;
export type WSMessageType = z.infer<typeof wsMessageTypeSchema>;
export type FileContext = z.infer<typeof fileContextSchema>;
export type ChatMode = z.infer<typeof chatModeSchema>;
export type ChatMessagePayload = z.infer<typeof chatMessagePayloadSchema>;
export type MessageMetrics = z.infer<typeof messageMetricsSchema>;
export type ChatChunkPayload = z.infer<typeof chatChunkPayloadSchema>;
export type ChatCompletePayload = z.infer<typeof chatCompletePayloadSchema>;
export type ToolStartedPayload = z.infer<typeof toolStartedPayloadSchema>;
export type ToolCompletedPayload = z.infer<typeof toolCompletedPayloadSchema>;
export type ToolConfirmPayload = z.infer<typeof toolConfirmPayloadSchema>;
export type ErrorPayload = z.infer<typeof errorPayloadSchema>;
