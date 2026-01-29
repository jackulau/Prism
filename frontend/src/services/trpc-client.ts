/**
 * Type-safe API client for the Prism backend.
 * Provides tRPC-like developer experience with full TypeScript autocomplete,
 * Zod runtime validation, and procedure-style API calls.
 */

import { z } from 'zod';
import {
  API_PATHS,
  type AuthResponse,
  type UserDTO,
  type ConversationDTO,
  type ConversationListResponse,
  type MessageListResponse,
  type KeyStatusResponse,
  type KeyValidationResponse,
  type DirectoryResponse,
  type BrowseDirectoriesResponse,
  type WorkspaceListResponse,
  type CloneRepoResponse,
  type FileContentResponse,
  type IntegrationStatusResponse,
  type BuildInfo,
  type GitHubStatusResponse,
  type GitHubReposResponse,
  type SuccessResponse,
} from '../types/api.generated';

import {
  authResponseSchema,
  userDTOSchema,
  conversationDTOSchema,
  conversationListResponseSchema,
  messageListResponseSchema,
  keyStatusResponseSchema,
  keyValidationResponseSchema,
  directoryResponseSchema,
  browseDirectoriesResponseSchema,
  workspaceListResponseSchema,
  cloneRepoResponseSchema,
  fileListResponseSchema,
  fileContentResponseSchema,
  integrationStatusResponseSchema,
  buildInfoSchema,
  gitHubStatusResponseSchema,
  gitHubReposResponseSchema,
  successResponseSchema,
} from '../schemas/api';

// ============================================================================
// Types
// ============================================================================

/** Result type for API calls - either success with data or error */
export type ApiResult<T> =
  | { success: true; data: T }
  | { success: false; error: string };

/** Options for API requests */
export interface RequestOptions {
  /** Skip Zod validation (use for performance in production) */
  skipValidation?: boolean;
  /** Custom headers */
  headers?: Record<string, string>;
  /** Request timeout in ms */
  timeout?: number;
  /** Abort signal */
  signal?: AbortSignal;
}

// ============================================================================
// Client Implementation
// ============================================================================

class TypedApiClient {
  private token: string | null = null;
  private baseUrl: string = '/api/v1';

  /** Set the authentication token */
  setToken(token: string | null): void {
    this.token = token;
  }

  /** Get the current token */
  getToken(): string | null {
    return this.token;
  }

  /** Set the base URL (for testing or different environments) */
  setBaseUrl(url: string): void {
    this.baseUrl = url;
  }

  /** Internal request method with validation */
  private async request<T>(
    endpoint: string,
    schema: z.ZodType<T>,
    options: RequestInit & RequestOptions = {}
  ): Promise<ApiResult<T>> {
    const { skipValidation, headers: customHeaders, timeout, signal, ...fetchOptions } = options;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...customHeaders,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      // Create abort controller for timeout
      const controller = new AbortController();
      const timeoutId = timeout ? setTimeout(() => controller.abort(), timeout) : null;

      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        ...fetchOptions,
        headers,
        signal: signal || controller.signal,
      });

      if (timeoutId) clearTimeout(timeoutId);

      // Handle empty responses safely
      const contentLength = response.headers.get('Content-Length');
      const contentType = response.headers.get('Content-Type');
      const hasJsonContent = contentType?.includes('application/json');
      const hasContent = contentLength !== '0' && contentLength !== null;

      let json: unknown;
      if (hasContent || hasJsonContent) {
        const text = await response.text();
        if (text) {
          try {
            json = JSON.parse(text);
          } catch {
            if (!response.ok) {
              return { success: false, error: text || `HTTP ${response.status}: ${response.statusText}` };
            }
          }
        }
      }

      if (!response.ok) {
        return {
          success: false,
          error: (json as { error?: string })?.error || `HTTP ${response.status}: ${response.statusText}`,
        };
      }

      // Validate response with Zod
      if (!skipValidation && json !== undefined) {
        const result = schema.safeParse(json);
        if (!result.success) {
          console.warn('API response validation failed:', result.error.issues);
          // Still return data but log warning - don't break the app
        }
      }

      return { success: true, data: json as T };
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        return { success: false, error: 'Request timeout' };
      }
      return {
        success: false,
        error: err instanceof Error ? err.message : 'Network error',
      };
    }
  }

  // ==========================================================================
  // Auth Procedures
  // ==========================================================================

  auth = {
    /** Register a new user */
    register: async (
      email: string,
      password: string,
      options?: RequestOptions
    ): Promise<ApiResult<AuthResponse>> => {
      return this.request(
        API_PATHS.AUTH_REGISTER,
        authResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ email, password }),
          ...options,
        }
      );
    },

    /** Login an existing user */
    login: async (
      email: string,
      password: string,
      options?: RequestOptions
    ): Promise<ApiResult<AuthResponse>> => {
      return this.request(
        API_PATHS.AUTH_LOGIN,
        authResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ email, password }),
          ...options,
        }
      );
    },

    /** Logout the current user */
    logout: async (options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.AUTH_LOGOUT,
        successResponseSchema,
        { method: 'POST', ...options }
      );
    },

    /** Refresh the access token */
    refresh: async (
      refreshToken: string,
      options?: RequestOptions
    ): Promise<ApiResult<AuthResponse>> => {
      return this.request(
        API_PATHS.AUTH_REFRESH,
        authResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ refresh_token: refreshToken }),
          ...options,
        }
      );
    },

    /** Get the current user */
    me: async (options?: RequestOptions): Promise<ApiResult<UserDTO>> => {
      return this.request(API_PATHS.AUTH_ME, userDTOSchema, options);
    },

    /** Login as a guest */
    guest: async (options?: RequestOptions): Promise<ApiResult<AuthResponse>> => {
      return this.request(
        API_PATHS.AUTH_GUEST,
        authResponseSchema,
        { method: 'POST', ...options }
      );
    },
  };

  // ==========================================================================
  // Conversation Procedures
  // ==========================================================================

  conversations = {
    /** List all conversations */
    list: async (
      params?: { limit?: number; offset?: number },
      options?: RequestOptions
    ): Promise<ApiResult<ConversationListResponse>> => {
      const limit = params?.limit ?? 50;
      const offset = params?.offset ?? 0;
      return this.request(
        `${API_PATHS.CONVERSATIONS_LIST}?limit=${limit}&offset=${offset}`,
        conversationListResponseSchema,
        options
      );
    },

    /** Search conversations */
    search: async (
      query: string,
      limit?: number,
      options?: RequestOptions
    ): Promise<ApiResult<ConversationListResponse>> => {
      const searchLimit = limit ?? 20;
      return this.request(
        `${API_PATHS.CONVERSATIONS_SEARCH}?q=${encodeURIComponent(query)}&limit=${searchLimit}`,
        conversationListResponseSchema,
        options
      );
    },

    /** Create a new conversation */
    create: async (
      input: { provider: string; model: string; systemPrompt?: string },
      options?: RequestOptions
    ): Promise<ApiResult<ConversationDTO>> => {
      return this.request(
        API_PATHS.CONVERSATIONS_CREATE,
        conversationDTOSchema,
        {
          method: 'POST',
          body: JSON.stringify({
            provider: input.provider,
            model: input.model,
            system_prompt: input.systemPrompt,
          }),
          ...options,
        }
      );
    },

    /** Get a conversation by ID */
    get: async (
      id: string,
      options?: RequestOptions
    ): Promise<ApiResult<ConversationDTO>> => {
      return this.request(
        API_PATHS.CONVERSATION_GET(id),
        conversationDTOSchema,
        options
      );
    },

    /** Update a conversation */
    update: async (
      id: string,
      title: string,
      options?: RequestOptions
    ): Promise<ApiResult<ConversationDTO>> => {
      return this.request(
        API_PATHS.CONVERSATION_UPDATE(id),
        conversationDTOSchema,
        {
          method: 'PATCH',
          body: JSON.stringify({ title }),
          ...options,
        }
      );
    },

    /** Delete a conversation */
    delete: async (
      id: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.CONVERSATION_DELETE(id),
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Get messages for a conversation */
    messages: async (
      conversationId: string,
      options?: RequestOptions
    ): Promise<ApiResult<MessageListResponse>> => {
      return this.request(
        API_PATHS.CONVERSATION_MESSAGES(conversationId),
        messageListResponseSchema,
        options
      );
    },
  };

  // ==========================================================================
  // Provider Procedures
  // ==========================================================================

  providers = {
    /** List configured providers */
    listKeys: async (options?: RequestOptions): Promise<ApiResult<{ providers: string[] }>> => {
      return this.request(
        API_PATHS.PROVIDERS_KEYS,
        z.object({ providers: z.array(z.string()) }),
        options
      );
    },

    /** Set an API key for a provider */
    setKey: async (
      provider: string,
      apiKey: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.PROVIDER_SET_KEY(provider),
        successResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ api_key: apiKey }),
          ...options,
        }
      );
    },

    /** Delete an API key for a provider */
    deleteKey: async (
      provider: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.PROVIDER_DELETE_KEY(provider),
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Validate an API key for a provider */
    validateKey: async (
      provider: string,
      apiKey: string,
      options?: RequestOptions
    ): Promise<ApiResult<KeyValidationResponse>> => {
      return this.request(
        API_PATHS.PROVIDER_VALIDATE_KEY(provider),
        keyValidationResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ api_key: apiKey }),
          ...options,
        }
      );
    },

    /** Get key status for a provider */
    keyStatus: async (
      provider: string,
      options?: RequestOptions
    ): Promise<ApiResult<KeyStatusResponse>> => {
      return this.request(
        API_PATHS.PROVIDER_KEY_STATUS(provider),
        keyStatusResponseSchema,
        options
      );
    },
  };

  // ==========================================================================
  // Workspace Procedures
  // ==========================================================================

  workspace = {
    /** Get the current workspace directory */
    getDirectory: async (options?: RequestOptions): Promise<ApiResult<DirectoryResponse>> => {
      return this.request(
        API_PATHS.WORKSPACE_DIRECTORY,
        directoryResponseSchema,
        options
      );
    },

    /** Set the workspace directory */
    setDirectory: async (
      directory: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse & { path: string }>> => {
      return this.request(
        API_PATHS.WORKSPACE_DIRECTORY,
        z.object({ success: z.boolean(), path: z.string(), message: z.string().optional() }),
        {
          method: 'POST',
          body: JSON.stringify({ directory }),
          ...options,
        }
      );
    },

    /** Browse directories */
    browse: async (
      path?: string,
      options?: RequestOptions
    ): Promise<ApiResult<BrowseDirectoriesResponse>> => {
      const queryPath = path || '/';
      return this.request(
        `${API_PATHS.WORKSPACE_BROWSE}?path=${encodeURIComponent(queryPath)}`,
        browseDirectoriesResponseSchema,
        options
      );
    },

    /** Open native folder picker */
    pickFolder: async (
      options?: RequestOptions
    ): Promise<ApiResult<{ success?: boolean; path?: string; cancelled?: boolean }>> => {
      return this.request(
        API_PATHS.WORKSPACE_PICK_FOLDER,
        z.object({
          success: z.boolean().optional(),
          path: z.string().optional(),
          cancelled: z.boolean().optional(),
        }),
        { method: 'POST', ...options }
      );
    },

    /** List recent workspaces */
    listRecent: async (options?: RequestOptions): Promise<ApiResult<WorkspaceListResponse>> => {
      return this.request(
        API_PATHS.WORKSPACE_RECENT,
        workspaceListResponseSchema,
        options
      );
    },

    /** Remove a workspace */
    remove: async (
      id: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.WORKSPACE_REMOVE(id),
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Set a workspace as current */
    setCurrent: async (
      id: string,
      options?: RequestOptions
    ): Promise<ApiResult<{ success: boolean; path: string }>> => {
      return this.request(
        API_PATHS.WORKSPACE_SET_CURRENT(id),
        z.object({ success: z.boolean(), path: z.string() }),
        { method: 'POST', ...options }
      );
    },
  };

  // ==========================================================================
  // File/Sandbox Procedures
  // ==========================================================================

  files = {
    /** List files in the sandbox */
    list: async (options?: RequestOptions): Promise<ApiResult<{ files: Array<{ name: string; path: string; type: string; children?: unknown[] }> }>> => {
      return this.request(API_PATHS.SANDBOX_FILES, fileListResponseSchema, options);
    },

    /** Get file content */
    get: async (
      path: string,
      options?: RequestOptions
    ): Promise<ApiResult<FileContentResponse>> => {
      return this.request(
        API_PATHS.SANDBOX_FILE(encodeURIComponent(path)),
        fileContentResponseSchema,
        options
      );
    },

    /** Write file content */
    write: async (
      path: string,
      content: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse & { path: string }>> => {
      return this.request(
        API_PATHS.SANDBOX_FILES,
        z.object({ success: z.boolean(), path: z.string(), message: z.string().optional() }),
        {
          method: 'POST',
          body: JSON.stringify({ path, content }),
          ...options,
        }
      );
    },

    /** Delete a file */
    delete: async (
      path: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.SANDBOX_FILE(encodeURIComponent(path)),
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Rename a file */
    rename: async (
      sourcePath: string,
      destPath: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        '/sandbox/files/rename',
        successResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ source_path: sourcePath, dest_path: destPath }),
          ...options,
        }
      );
    },

    /** Create a directory */
    createDirectory: async (
      path: string,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        '/sandbox/files/mkdir',
        successResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ path }),
          ...options,
        }
      );
    },
  };

  // ==========================================================================
  // Build Procedures
  // ==========================================================================

  builds = {
    /** Get build status */
    get: async (id: string, options?: RequestOptions): Promise<ApiResult<BuildInfo>> => {
      return this.request(API_PATHS.BUILD_GET(id), buildInfoSchema, options);
    },

    /** Stop a build */
    stop: async (id: string, options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.BUILD_STOP(id),
        successResponseSchema,
        { method: 'POST', ...options }
      );
    },
  };

  // ==========================================================================
  // GitHub Procedures
  // ==========================================================================

  github = {
    /** Get GitHub connection status */
    status: async (options?: RequestOptions): Promise<ApiResult<GitHubStatusResponse>> => {
      return this.request(API_PATHS.GITHUB_STATUS, gitHubStatusResponseSchema, options);
    },

    /** Get GitHub repositories */
    repos: async (options?: RequestOptions): Promise<ApiResult<GitHubReposResponse>> => {
      return this.request(API_PATHS.GITHUB_REPOS, gitHubReposResponseSchema, options);
    },

    /** Clone a GitHub repository */
    clone: async (
      repoUrl: string,
      branch?: string,
      options?: RequestOptions
    ): Promise<ApiResult<CloneRepoResponse>> => {
      return this.request(
        API_PATHS.GITHUB_CLONE,
        cloneRepoResponseSchema,
        {
          method: 'POST',
          body: JSON.stringify({ repo_url: repoUrl, branch }),
          ...options,
        }
      );
    },

    /** Disconnect GitHub */
    disconnect: async (options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.GITHUB_DISCONNECT,
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Get GitHub authorization URL */
    getAuthUrl: async (options?: RequestOptions): Promise<ApiResult<{ url: string }>> => {
      return this.request(
        API_PATHS.OAUTH_GITHUB_AUTHORIZE,
        z.object({ url: z.string() }),
        options
      );
    },
  };

  // ==========================================================================
  // Integration Procedures
  // ==========================================================================

  integrations = {
    /** Get all integration statuses */
    status: async (options?: RequestOptions): Promise<ApiResult<IntegrationStatusResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_STATUS,
        integrationStatusResponseSchema,
        options
      );
    },

    /** Set Discord integration */
    setDiscord: async (
      webhookUrl: string,
      enabled?: boolean,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_DISCORD,
        successResponseSchema.extend({ enabled: z.boolean().optional(), connected: z.boolean().optional() }),
        {
          method: 'POST',
          body: JSON.stringify({ webhook_url: webhookUrl, enabled: enabled ?? true }),
          ...options,
        }
      );
    },

    /** Delete Discord integration */
    deleteDiscord: async (options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_DISCORD,
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Set Slack integration */
    setSlack: async (
      webhookUrl: string,
      channelId?: string,
      enabled?: boolean,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_SLACK,
        successResponseSchema.extend({ enabled: z.boolean().optional(), connected: z.boolean().optional() }),
        {
          method: 'POST',
          body: JSON.stringify({
            webhook_url: webhookUrl,
            channel_id: channelId,
            enabled: enabled ?? true,
          }),
          ...options,
        }
      );
    },

    /** Delete Slack integration */
    deleteSlack: async (options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_SLACK,
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },

    /** Set PostHog integration */
    setPostHog: async (
      enabled: boolean,
      options?: RequestOptions
    ): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_POSTHOG,
        successResponseSchema.extend({ enabled: z.boolean().optional(), connected: z.boolean().optional() }),
        {
          method: 'POST',
          body: JSON.stringify({ enabled }),
          ...options,
        }
      );
    },

    /** Delete PostHog integration */
    deletePostHog: async (options?: RequestOptions): Promise<ApiResult<SuccessResponse>> => {
      return this.request(
        API_PATHS.INTEGRATIONS_POSTHOG,
        successResponseSchema,
        { method: 'DELETE', ...options }
      );
    },
  };
}

// ============================================================================
// Exports
// ============================================================================

/** Singleton instance of the typed API client */
export const api = new TypedApiClient();

/** Export the class for testing or custom instances */
export { TypedApiClient };

/** Helper to unwrap ApiResult - throws on error */
export function unwrap<T>(result: ApiResult<T>): T {
  if (!result.success) {
    throw new Error(result.error);
  }
  return result.data;
}

/** Helper to check if result is successful */
export function isSuccess<T>(result: ApiResult<T>): result is { success: true; data: T } {
  return result.success;
}

/** Helper to check if result is an error */
export function isError<T>(result: ApiResult<T>): result is { success: false; error: string } {
  return !result.success;
}
