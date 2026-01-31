import { create } from 'zustand';
import type { Team, TeamMember, Role } from '../types/team';

interface TeamState {
  // Current team context
  currentTeam: Team | null;
  showAllTeams: boolean;

  // Team list for the organization
  teams: Team[];
  isLoadingTeams: boolean;
  teamsError: string | null;

  // Team members cache (keyed by team id)
  membersByTeam: Record<string, TeamMember[]>;
  isLoadingMembers: boolean;
  membersError: string | null;

  // Available roles
  roles: Role[];
  isLoadingRoles: boolean;

  // Actions
  setCurrentTeam: (team: Team | null) => void;
  setShowAllTeams: (show: boolean) => void;
  setTeams: (teams: Team[]) => void;
  addTeam: (team: Team) => void;
  updateTeam: (teamId: string, updates: Partial<Team>) => void;
  removeTeam: (teamId: string) => void;
  setTeamsLoading: (loading: boolean) => void;
  setTeamsError: (error: string | null) => void;

  // Member actions
  setTeamMembers: (teamId: string, members: TeamMember[]) => void;
  addTeamMember: (teamId: string, member: TeamMember) => void;
  updateTeamMember: (teamId: string, memberId: string, updates: Partial<TeamMember>) => void;
  removeTeamMember: (teamId: string, memberId: string) => void;
  setMembersLoading: (loading: boolean) => void;
  setMembersError: (error: string | null) => void;

  // Role actions
  setRoles: (roles: Role[]) => void;
  addRole: (role: Role) => void;
  updateRole: (roleId: string, updates: Partial<Role>) => void;
  removeRole: (roleId: string) => void;
  setRolesLoading: (loading: boolean) => void;

  // Reset
  reset: () => void;
}

const initialState = {
  currentTeam: null,
  showAllTeams: false,
  teams: [],
  isLoadingTeams: false,
  teamsError: null,
  membersByTeam: {},
  isLoadingMembers: false,
  membersError: null,
  roles: [],
  isLoadingRoles: false,
};

export const useTeamStore = create<TeamState>((set) => ({
  ...initialState,

  // Team context
  setCurrentTeam: (team) => set({ currentTeam: team, showAllTeams: false }),
  setShowAllTeams: (show) => set({ showAllTeams: show, currentTeam: show ? null : null }),

  // Teams list
  setTeams: (teams) => set({ teams }),
  addTeam: (team) => set((state) => ({ teams: [...state.teams, team] })),
  updateTeam: (teamId, updates) =>
    set((state) => ({
      teams: state.teams.map((t) => (t.id === teamId ? { ...t, ...updates } : t)),
      currentTeam:
        state.currentTeam?.id === teamId
          ? { ...state.currentTeam, ...updates }
          : state.currentTeam,
    })),
  removeTeam: (teamId) =>
    set((state) => ({
      teams: state.teams.filter((t) => t.id !== teamId),
      currentTeam: state.currentTeam?.id === teamId ? null : state.currentTeam,
    })),
  setTeamsLoading: (isLoadingTeams) => set({ isLoadingTeams }),
  setTeamsError: (teamsError) => set({ teamsError }),

  // Members
  setTeamMembers: (teamId, members) =>
    set((state) => ({
      membersByTeam: { ...state.membersByTeam, [teamId]: members },
    })),
  addTeamMember: (teamId, member) =>
    set((state) => ({
      membersByTeam: {
        ...state.membersByTeam,
        [teamId]: [...(state.membersByTeam[teamId] || []), member],
      },
      teams: state.teams.map((t) =>
        t.id === teamId ? { ...t, memberCount: t.memberCount + 1 } : t
      ),
    })),
  updateTeamMember: (teamId, memberId, updates) =>
    set((state) => ({
      membersByTeam: {
        ...state.membersByTeam,
        [teamId]: (state.membersByTeam[teamId] || []).map((m) =>
          m.id === memberId ? { ...m, ...updates } : m
        ),
      },
    })),
  removeTeamMember: (teamId, memberId) =>
    set((state) => ({
      membersByTeam: {
        ...state.membersByTeam,
        [teamId]: (state.membersByTeam[teamId] || []).filter((m) => m.id !== memberId),
      },
      teams: state.teams.map((t) =>
        t.id === teamId ? { ...t, memberCount: Math.max(0, t.memberCount - 1) } : t
      ),
    })),
  setMembersLoading: (isLoadingMembers) => set({ isLoadingMembers }),
  setMembersError: (membersError) => set({ membersError }),

  // Roles
  setRoles: (roles) => set({ roles }),
  addRole: (role) => set((state) => ({ roles: [...state.roles, role] })),
  updateRole: (roleId, updates) =>
    set((state) => ({
      roles: state.roles.map((r) => (r.id === roleId ? { ...r, ...updates } : r)),
    })),
  removeRole: (roleId) =>
    set((state) => ({
      roles: state.roles.filter((r) => r.id !== roleId),
    })),
  setRolesLoading: (isLoadingRoles) => set({ isLoadingRoles }),

  // Reset
  reset: () => set(initialState),
}));
