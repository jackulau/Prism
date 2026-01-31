import type {
  Team,
  TeamMember,
  Role,
  CreateTeamRequest,
  UpdateTeamRequest,
  AddTeamMemberRequest,
  UpdateTeamMemberRequest,
  CreateRoleRequest,
  UpdateRoleRequest,
} from '../types/team';
import { useTeamStore } from '../store/teamStore';

const API_BASE_URL = '/api/v1';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

class TeamService {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });

      const contentType = response.headers.get('Content-Type');
      const hasJsonContent = contentType?.includes('application/json');

      let data: T | undefined;
      if (hasJsonContent) {
        const text = await response.text();
        if (text) {
          try {
            data = JSON.parse(text);
          } catch {
            if (!response.ok) {
              return { error: text || 'An error occurred' };
            }
          }
        }
      }

      if (!response.ok) {
        return { error: (data as { error?: string })?.error || 'An error occurred' };
      }

      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  // Team CRUD
  async listTeams(organizationId: string): Promise<ApiResponse<{ teams: Team[] }>> {
    return this.request(`/organizations/${organizationId}/teams`);
  }

  async getTeam(organizationId: string, teamId: string): Promise<ApiResponse<Team>> {
    return this.request(`/organizations/${organizationId}/teams/${teamId}`);
  }

  async createTeam(
    organizationId: string,
    data: CreateTeamRequest
  ): Promise<ApiResponse<Team>> {
    return this.request(`/organizations/${organizationId}/teams`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateTeam(
    organizationId: string,
    teamId: string,
    data: UpdateTeamRequest
  ): Promise<ApiResponse<Team>> {
    return this.request(`/organizations/${organizationId}/teams/${teamId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteTeam(organizationId: string, teamId: string): Promise<ApiResponse<void>> {
    return this.request(`/organizations/${organizationId}/teams/${teamId}`, {
      method: 'DELETE',
    });
  }

  // Team Members
  async listTeamMembers(
    organizationId: string,
    teamId: string
  ): Promise<ApiResponse<{ members: TeamMember[] }>> {
    return this.request(`/organizations/${organizationId}/teams/${teamId}/members`);
  }

  async addTeamMember(
    organizationId: string,
    teamId: string,
    data: AddTeamMemberRequest
  ): Promise<ApiResponse<TeamMember>> {
    return this.request(`/organizations/${organizationId}/teams/${teamId}/members`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateTeamMember(
    organizationId: string,
    teamId: string,
    memberId: string,
    data: UpdateTeamMemberRequest
  ): Promise<ApiResponse<TeamMember>> {
    return this.request(
      `/organizations/${organizationId}/teams/${teamId}/members/${memberId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      }
    );
  }

  async removeTeamMember(
    organizationId: string,
    teamId: string,
    memberId: string
  ): Promise<ApiResponse<void>> {
    return this.request(
      `/organizations/${organizationId}/teams/${teamId}/members/${memberId}`,
      {
        method: 'DELETE',
      }
    );
  }

  // Roles
  async listRoles(organizationId: string): Promise<ApiResponse<{ roles: Role[] }>> {
    return this.request(`/organizations/${organizationId}/roles`);
  }

  async getRole(organizationId: string, roleId: string): Promise<ApiResponse<Role>> {
    return this.request(`/organizations/${organizationId}/roles/${roleId}`);
  }

  async createRole(
    organizationId: string,
    data: CreateRoleRequest
  ): Promise<ApiResponse<Role>> {
    return this.request(`/organizations/${organizationId}/roles`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateRole(
    organizationId: string,
    roleId: string,
    data: UpdateRoleRequest
  ): Promise<ApiResponse<Role>> {
    return this.request(`/organizations/${organizationId}/roles/${roleId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteRole(organizationId: string, roleId: string): Promise<ApiResponse<void>> {
    return this.request(`/organizations/${organizationId}/roles/${roleId}`, {
      method: 'DELETE',
    });
  }

  // Convenience methods that update the store
  async loadTeams(organizationId: string): Promise<void> {
    const store = useTeamStore.getState();
    store.setTeamsLoading(true);
    store.setTeamsError(null);

    const response = await this.listTeams(organizationId);

    if (response.data?.teams) {
      store.setTeams(response.data.teams);
    } else if (response.error) {
      store.setTeamsError(response.error);
    }

    store.setTeamsLoading(false);
  }

  async loadTeamMembers(organizationId: string, teamId: string): Promise<void> {
    const store = useTeamStore.getState();
    store.setMembersLoading(true);
    store.setMembersError(null);

    const response = await this.listTeamMembers(organizationId, teamId);

    if (response.data?.members) {
      store.setTeamMembers(teamId, response.data.members);
    } else if (response.error) {
      store.setMembersError(response.error);
    }

    store.setMembersLoading(false);
  }

  async loadRoles(organizationId: string): Promise<void> {
    const store = useTeamStore.getState();
    store.setRolesLoading(true);

    const response = await this.listRoles(organizationId);

    if (response.data?.roles) {
      store.setRoles(response.data.roles);
    }

    store.setRolesLoading(false);
  }
}

export const teamService = new TeamService();
