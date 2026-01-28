/**
 * React Query hooks for the Prism API.
 * Provides type-safe data fetching with automatic caching, loading states,
 * and error handling.
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from '@tanstack/react-query';
import { api, type ApiResult, isSuccess } from '../services/trpc-client';
import type {
  AuthResponse,
  UserDTO,
  ConversationDTO,
  ConversationListResponse,
  MessageListResponse,
  KeyStatusResponse,
  KeyValidationResponse,
  DirectoryResponse,
  BrowseDirectoriesResponse,
  WorkspaceListResponse,
  CloneRepoResponse,
  FileContentResponse,
  IntegrationStatusResponse,
  BuildInfo,
  GitHubStatusResponse,
  GitHubReposResponse,
  SuccessResponse,
} from '../types/api.generated';

/** File list response type with simplified children */
type FileListResponse = { files: Array<{ name: string; path: string; type: string; children?: unknown[] }> };

// ============================================================================
// Query Keys
// ============================================================================

/** Centralized query keys for cache management */
export const queryKeys = {
  // Auth
  auth: {
    all: ['auth'] as const,
    me: () => [...queryKeys.auth.all, 'me'] as const,
  },

  // Conversations
  conversations: {
    all: ['conversations'] as const,
    list: (params?: { limit?: number; offset?: number }) =>
      [...queryKeys.conversations.all, 'list', params] as const,
    search: (query: string, limit?: number) =>
      [...queryKeys.conversations.all, 'search', query, limit] as const,
    detail: (id: string) => [...queryKeys.conversations.all, 'detail', id] as const,
    messages: (id: string) => [...queryKeys.conversations.all, 'messages', id] as const,
  },

  // Providers
  providers: {
    all: ['providers'] as const,
    keys: () => [...queryKeys.providers.all, 'keys'] as const,
    keyStatus: (provider: string) =>
      [...queryKeys.providers.all, 'keyStatus', provider] as const,
  },

  // Workspace
  workspace: {
    all: ['workspace'] as const,
    directory: () => [...queryKeys.workspace.all, 'directory'] as const,
    browse: (path?: string) => [...queryKeys.workspace.all, 'browse', path] as const,
    recent: () => [...queryKeys.workspace.all, 'recent'] as const,
  },

  // Files
  files: {
    all: ['files'] as const,
    list: () => [...queryKeys.files.all, 'list'] as const,
    content: (path: string) => [...queryKeys.files.all, 'content', path] as const,
  },

  // Builds
  builds: {
    all: ['builds'] as const,
    detail: (id: string) => [...queryKeys.builds.all, 'detail', id] as const,
  },

  // GitHub
  github: {
    all: ['github'] as const,
    status: () => [...queryKeys.github.all, 'status'] as const,
    repos: () => [...queryKeys.github.all, 'repos'] as const,
  },

  // Integrations
  integrations: {
    all: ['integrations'] as const,
    status: () => [...queryKeys.integrations.all, 'status'] as const,
  },
} as const;

// ============================================================================
// Utility Types
// ============================================================================

/** Custom error class for API errors */
export class ApiError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

/** Transform ApiResult to throwable format for React Query */
function unwrapOrThrow<T>(result: ApiResult<T>): T {
  if (!isSuccess(result)) {
    throw new ApiError(result.error);
  }
  return result.data;
}

// ============================================================================
// Auth Hooks
// ============================================================================

/** Get current user */
export function useCurrentUser(
  options?: Omit<UseQueryOptions<UserDTO, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.auth.me(),
    queryFn: async () => unwrapOrThrow(await api.auth.me()),
    ...options,
  });
}

/** Login mutation */
export function useLogin(
  options?: UseMutationOptions<AuthResponse, ApiError, { email: string; password: string }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ email, password }) =>
      unwrapOrThrow(await api.auth.login(email, password)),
    onSuccess: (data) => {
      api.setToken(data.access_token);
      queryClient.setQueryData(queryKeys.auth.me(), data.user);
    },
    ...options,
  });
}

/** Register mutation */
export function useRegister(
  options?: UseMutationOptions<AuthResponse, ApiError, { email: string; password: string }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ email, password }) =>
      unwrapOrThrow(await api.auth.register(email, password)),
    onSuccess: (data) => {
      api.setToken(data.access_token);
      queryClient.setQueryData(queryKeys.auth.me(), data.user);
    },
    ...options,
  });
}

/** Logout mutation */
export function useLogout(options?: UseMutationOptions<SuccessResponse, ApiError, void>) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.auth.logout()),
    onSuccess: () => {
      api.setToken(null);
      queryClient.clear();
    },
    ...options,
  });
}

/** Guest login mutation */
export function useGuestLogin(options?: UseMutationOptions<AuthResponse, ApiError, void>) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.auth.guest()),
    onSuccess: (data) => {
      api.setToken(data.access_token);
      queryClient.setQueryData(queryKeys.auth.me(), data.user);
    },
    ...options,
  });
}

// ============================================================================
// Conversation Hooks
// ============================================================================

/** List conversations */
export function useConversations(
  params?: { limit?: number; offset?: number },
  options?: Omit<UseQueryOptions<ConversationListResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.conversations.list(params),
    queryFn: async () => unwrapOrThrow(await api.conversations.list(params)),
    ...options,
  });
}

/** Search conversations */
export function useConversationSearch(
  query: string,
  limit?: number,
  options?: Omit<UseQueryOptions<ConversationListResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.conversations.search(query, limit),
    queryFn: async () => unwrapOrThrow(await api.conversations.search(query, limit)),
    enabled: query.length > 0,
    ...options,
  });
}

/** Get single conversation */
export function useConversation(
  id: string,
  options?: Omit<UseQueryOptions<ConversationDTO, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.conversations.detail(id),
    queryFn: async () => unwrapOrThrow(await api.conversations.get(id)),
    enabled: !!id,
    ...options,
  });
}

/** Get conversation messages */
export function useMessages(
  conversationId: string,
  options?: Omit<UseQueryOptions<MessageListResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.conversations.messages(conversationId),
    queryFn: async () => unwrapOrThrow(await api.conversations.messages(conversationId)),
    enabled: !!conversationId,
    ...options,
  });
}

/** Create conversation mutation */
export function useCreateConversation(
  options?: UseMutationOptions<
    ConversationDTO,
    ApiError,
    { provider: string; model: string; systemPrompt?: string }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input) => unwrapOrThrow(await api.conversations.create(input)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.conversations.all });
    },
    ...options,
  });
}

/** Update conversation mutation */
export function useUpdateConversation(
  options?: UseMutationOptions<ConversationDTO, ApiError, { id: string; title: string }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, title }) =>
      unwrapOrThrow(await api.conversations.update(id, title)),
    onSuccess: (data, { id }) => {
      queryClient.setQueryData(queryKeys.conversations.detail(id), data);
      queryClient.invalidateQueries({ queryKey: queryKeys.conversations.list() });
    },
    ...options,
  });
}

/** Delete conversation mutation */
export function useDeleteConversation(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id) => unwrapOrThrow(await api.conversations.delete(id)),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: queryKeys.conversations.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.conversations.all });
    },
    ...options,
  });
}

// ============================================================================
// Provider Hooks
// ============================================================================

/** List provider keys */
export function useProviderKeys(
  options?: Omit<UseQueryOptions<{ providers: string[] }, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.providers.keys(),
    queryFn: async () => unwrapOrThrow(await api.providers.listKeys()),
    ...options,
  });
}

/** Get provider key status */
export function useProviderKeyStatus(
  provider: string,
  options?: Omit<UseQueryOptions<KeyStatusResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.providers.keyStatus(provider),
    queryFn: async () => unwrapOrThrow(await api.providers.keyStatus(provider)),
    enabled: !!provider,
    ...options,
  });
}

/** Set provider key mutation */
export function useSetProviderKey(
  options?: UseMutationOptions<SuccessResponse, ApiError, { provider: string; apiKey: string }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ provider, apiKey }) =>
      unwrapOrThrow(await api.providers.setKey(provider, apiKey)),
    onSuccess: (_, { provider }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.providers.keys() });
      queryClient.invalidateQueries({ queryKey: queryKeys.providers.keyStatus(provider) });
    },
    ...options,
  });
}

/** Delete provider key mutation */
export function useDeleteProviderKey(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (provider) => unwrapOrThrow(await api.providers.deleteKey(provider)),
    onSuccess: (_, provider) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.providers.keys() });
      queryClient.invalidateQueries({ queryKey: queryKeys.providers.keyStatus(provider) });
    },
    ...options,
  });
}

/** Validate provider key mutation */
export function useValidateProviderKey(
  options?: UseMutationOptions<
    KeyValidationResponse,
    ApiError,
    { provider: string; apiKey: string }
  >
) {
  return useMutation({
    mutationFn: async ({ provider, apiKey }) =>
      unwrapOrThrow(await api.providers.validateKey(provider, apiKey)),
    ...options,
  });
}

// ============================================================================
// Workspace Hooks
// ============================================================================

/** Get workspace directory */
export function useWorkspaceDirectory(
  options?: Omit<UseQueryOptions<DirectoryResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.workspace.directory(),
    queryFn: async () => unwrapOrThrow(await api.workspace.getDirectory()),
    ...options,
  });
}

/** Browse directories */
export function useBrowseDirectories(
  path?: string,
  options?: Omit<UseQueryOptions<BrowseDirectoriesResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.workspace.browse(path),
    queryFn: async () => unwrapOrThrow(await api.workspace.browse(path)),
    ...options,
  });
}

/** List recent workspaces */
export function useRecentWorkspaces(
  options?: Omit<UseQueryOptions<WorkspaceListResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.workspace.recent(),
    queryFn: async () => unwrapOrThrow(await api.workspace.listRecent()),
    ...options,
  });
}

/** Set workspace directory mutation */
export function useSetWorkspaceDirectory(
  options?: UseMutationOptions<
    SuccessResponse & { path: string },
    ApiError,
    string
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (directory) =>
      unwrapOrThrow(await api.workspace.setDirectory(directory)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspace.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
    },
    ...options,
  });
}

/** Set current workspace mutation */
export function useSetCurrentWorkspace(
  options?: UseMutationOptions<{ success: boolean; path: string }, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id) => unwrapOrThrow(await api.workspace.setCurrent(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspace.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
    },
    ...options,
  });
}

/** Remove workspace mutation */
export function useRemoveWorkspace(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id) => unwrapOrThrow(await api.workspace.remove(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspace.recent() });
    },
    ...options,
  });
}

// ============================================================================
// File Hooks
// ============================================================================

/** List files */
export function useFiles(
  options?: Omit<UseQueryOptions<FileListResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.files.list(),
    queryFn: async () => unwrapOrThrow(await api.files.list()),
    ...options,
  });
}

/** Get file content */
export function useFileContent(
  path: string,
  options?: Omit<UseQueryOptions<FileContentResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.files.content(path),
    queryFn: async () => unwrapOrThrow(await api.files.get(path)),
    enabled: !!path,
    ...options,
  });
}

/** Write file mutation */
export function useWriteFile(
  options?: UseMutationOptions<
    SuccessResponse & { path: string },
    ApiError,
    { path: string; content: string }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ path, content }) =>
      unwrapOrThrow(await api.files.write(path, content)),
    onSuccess: (_, { path }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.list() });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.content(path) });
    },
    ...options,
  });
}

/** Delete file mutation */
export function useDeleteFile(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (path) => unwrapOrThrow(await api.files.delete(path)),
    onSuccess: (_, path) => {
      queryClient.removeQueries({ queryKey: queryKeys.files.content(path) });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.list() });
    },
    ...options,
  });
}

/** Rename file mutation */
export function useRenameFile(
  options?: UseMutationOptions<
    SuccessResponse,
    ApiError,
    { sourcePath: string; destPath: string }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ sourcePath, destPath }) =>
      unwrapOrThrow(await api.files.rename(sourcePath, destPath)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
    },
    ...options,
  });
}

/** Create directory mutation */
export function useCreateDirectory(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (path) => unwrapOrThrow(await api.files.createDirectory(path)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.list() });
    },
    ...options,
  });
}

// ============================================================================
// Build Hooks
// ============================================================================

/** Get build status */
export function useBuild(
  id: string,
  options?: Omit<UseQueryOptions<BuildInfo, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.builds.detail(id),
    queryFn: async () => unwrapOrThrow(await api.builds.get(id)),
    enabled: !!id,
    ...options,
  });
}

/** Stop build mutation */
export function useStopBuild(
  options?: UseMutationOptions<SuccessResponse, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id) => unwrapOrThrow(await api.builds.stop(id)),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.builds.detail(id) });
    },
    ...options,
  });
}

// ============================================================================
// GitHub Hooks
// ============================================================================

/** Get GitHub status */
export function useGitHubStatus(
  options?: Omit<UseQueryOptions<GitHubStatusResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.github.status(),
    queryFn: async () => unwrapOrThrow(await api.github.status()),
    ...options,
  });
}

/** Get GitHub repos */
export function useGitHubRepos(
  options?: Omit<UseQueryOptions<GitHubReposResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.github.repos(),
    queryFn: async () => unwrapOrThrow(await api.github.repos()),
    ...options,
  });
}

/** Clone GitHub repo mutation */
export function useCloneGitHubRepo(
  options?: UseMutationOptions<
    CloneRepoResponse,
    ApiError,
    { repoUrl: string; branch?: string }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ repoUrl, branch }) =>
      unwrapOrThrow(await api.github.clone(repoUrl, branch)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspace.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
    },
    ...options,
  });
}

/** Disconnect GitHub mutation */
export function useDisconnectGitHub(
  options?: UseMutationOptions<SuccessResponse, ApiError, void>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.github.disconnect()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.github.all });
    },
    ...options,
  });
}

// ============================================================================
// Integration Hooks
// ============================================================================

/** Get integration status */
export function useIntegrationStatus(
  options?: Omit<UseQueryOptions<IntegrationStatusResponse, ApiError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.integrations.status(),
    queryFn: async () => unwrapOrThrow(await api.integrations.status()),
    ...options,
  });
}

/** Set Discord integration mutation */
export function useSetDiscordIntegration(
  options?: UseMutationOptions<
    SuccessResponse,
    ApiError,
    { webhookUrl: string; enabled?: boolean }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ webhookUrl, enabled }) =>
      unwrapOrThrow(await api.integrations.setDiscord(webhookUrl, enabled)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}

/** Delete Discord integration mutation */
export function useDeleteDiscordIntegration(
  options?: UseMutationOptions<SuccessResponse, ApiError, void>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.integrations.deleteDiscord()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}

/** Set Slack integration mutation */
export function useSetSlackIntegration(
  options?: UseMutationOptions<
    SuccessResponse,
    ApiError,
    { webhookUrl: string; channelId?: string; enabled?: boolean }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ webhookUrl, channelId, enabled }) =>
      unwrapOrThrow(await api.integrations.setSlack(webhookUrl, channelId, enabled)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}

/** Delete Slack integration mutation */
export function useDeleteSlackIntegration(
  options?: UseMutationOptions<SuccessResponse, ApiError, void>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.integrations.deleteSlack()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}

/** Set PostHog integration mutation */
export function useSetPostHogIntegration(
  options?: UseMutationOptions<SuccessResponse, ApiError, boolean>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (enabled) =>
      unwrapOrThrow(await api.integrations.setPostHog(enabled)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}

/** Delete PostHog integration mutation */
export function useDeletePostHogIntegration(
  options?: UseMutationOptions<SuccessResponse, ApiError, void>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => unwrapOrThrow(await api.integrations.deletePostHog()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.integrations.status() });
    },
    ...options,
  });
}
