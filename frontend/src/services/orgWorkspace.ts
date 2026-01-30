import type {
  OrgWorkspace,
  CreateOrgWorkspaceInput,
  UpdateOrgWorkspaceInput,
  OrgWorkspaceListParams,
  OrgWorkspaceListResponse,
} from '../types/organization';

const API_BASE_URL = '/api/v1';

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('access_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Request failed with status ${response.status}`);
  }
  return response.json();
}

export const orgWorkspaceService = {
  async list(
    orgId: string,
    params?: OrgWorkspaceListParams
  ): Promise<OrgWorkspaceListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const queryString = searchParams.toString();
    const url = `${API_BASE_URL}/organizations/${orgId}/workspaces${
      queryString ? `?${queryString}` : ''
    }`;

    const response = await fetch(url, {
      method: 'GET',
      headers: getAuthHeaders(),
    });

    return handleResponse<OrgWorkspaceListResponse>(response);
  },

  async create(orgId: string, data: CreateOrgWorkspaceInput): Promise<OrgWorkspace> {
    const response = await fetch(
      `${API_BASE_URL}/organizations/${orgId}/workspaces`,
      {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(data),
      }
    );

    return handleResponse<OrgWorkspace>(response);
  },

  async get(orgId: string, workspaceId: string): Promise<OrgWorkspace> {
    const response = await fetch(
      `${API_BASE_URL}/organizations/${orgId}/workspaces/${workspaceId}`,
      {
        method: 'GET',
        headers: getAuthHeaders(),
      }
    );

    return handleResponse<OrgWorkspace>(response);
  },

  async update(
    orgId: string,
    workspaceId: string,
    data: UpdateOrgWorkspaceInput
  ): Promise<OrgWorkspace> {
    const response = await fetch(
      `${API_BASE_URL}/organizations/${orgId}/workspaces/${workspaceId}`,
      {
        method: 'PATCH',
        headers: getAuthHeaders(),
        body: JSON.stringify(data),
      }
    );

    return handleResponse<OrgWorkspace>(response);
  },

  async delete(orgId: string, workspaceId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/organizations/${orgId}/workspaces/${workspaceId}`,
      {
        method: 'DELETE',
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to delete workspace');
    }
  },
};
