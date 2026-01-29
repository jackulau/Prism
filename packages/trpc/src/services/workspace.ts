import type {
  Workspace,
  BrowseDirectoryOutput,
  PickFolderOutput,
} from '../routers/workspace/schemas.js';

const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';

interface GoWorkspaceResponse {
  id: string;
  path: string;
  name: string;
  is_current: boolean;
  last_accessed_at: string | null;
  created_at?: string;
}

interface GoListWorkspacesResponse {
  workspaces: GoWorkspaceResponse[];
}

interface GoBrowseResponse {
  current_path: string;
  parent_path: string;
  directories: Array<{ name: string; path: string }>;
}

interface GoPickFolderResponse {
  success?: boolean;
  path?: string;
  cancelled?: boolean;
}

interface GoDirectoryResponse {
  path: string;
}

interface GoSetDirectoryResponse {
  success: boolean;
  path: string;
}

function mapWorkspace(ws: GoWorkspaceResponse, userId: string): Workspace {
  return {
    id: ws.id,
    userId: userId,
    path: ws.path,
    name: ws.name,
    isCurrent: ws.is_current,
    lastAccessedAt: ws.last_accessed_at,
    createdAt: ws.created_at || new Date().toISOString(),
  };
}

async function fetchWithAuth<T>(
  path: string,
  token: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${GO_BACKEND_URL}${path}`, {
    ...options,
    headers: {
      ...options.headers,
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || `HTTP ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export const workspaceService = {
  async getCurrent(
    _userId: string,
    token: string
  ): Promise<{ path: string } | null> {
    try {
      const result = await fetchWithAuth<GoDirectoryResponse>(
        '/api/v1/workspace/directory',
        token
      );
      return { path: result.path };
    } catch {
      return null;
    }
  },

  async setDirectory(
    _userId: string,
    path: string,
    token: string
  ): Promise<{ success: boolean; path: string }> {
    const result = await fetchWithAuth<GoSetDirectoryResponse>(
      '/api/v1/workspace/directory',
      token,
      {
        method: 'POST',
        body: JSON.stringify({ directory: path }),
      }
    );
    return { success: result.success, path: result.path };
  },

  async browse(
    path: string | undefined,
    _showHidden: boolean,
    token: string
  ): Promise<BrowseDirectoryOutput> {
    const queryPath = path || '/';
    const result = await fetchWithAuth<GoBrowseResponse>(
      `/api/v1/workspace/browse?path=${encodeURIComponent(queryPath)}`,
      token
    );
    return {
      currentPath: result.current_path,
      parentPath: result.parent_path || null,
      directories: result.directories,
    };
  },

  async pickFolder(token: string): Promise<PickFolderOutput> {
    try {
      const result = await fetchWithAuth<GoPickFolderResponse>(
        '/api/v1/workspace/pick-folder',
        token,
        { method: 'POST' }
      );
      return {
        path: result.path || null,
        cancelled: result.cancelled,
      };
    } catch {
      return { path: null, cancelled: true };
    }
  },

  async listRecent(
    userId: string,
    limit: number,
    token: string
  ): Promise<Workspace[]> {
    const result = await fetchWithAuth<GoListWorkspacesResponse>(
      `/api/v1/workspace/recent?limit=${limit}`,
      token
    );
    return (result.workspaces || []).map((ws) => mapWorkspace(ws, userId));
  },

  async setCurrent(
    _userId: string,
    workspaceId: string,
    token: string
  ): Promise<{ success: boolean; path: string }> {
    const result = await fetchWithAuth<GoSetDirectoryResponse>(
      `/api/v1/workspace/${workspaceId}/current`,
      token,
      { method: 'POST' }
    );
    return { success: result.success, path: result.path };
  },

  async delete(_userId: string, workspaceId: string, token: string): Promise<void> {
    await fetchWithAuth<{ success: boolean }>(
      `/api/v1/workspace/${workspaceId}`,
      token,
      { method: 'DELETE' }
    );
  },

  async getById(
    workspaceId: string,
    userId: string,
    token: string
  ): Promise<Workspace | null> {
    try {
      const result = await fetchWithAuth<GoWorkspaceResponse>(
        `/api/v1/workspace/${workspaceId}`,
        token
      );
      return mapWorkspace(result, userId);
    } catch {
      return null;
    }
  },
};
