import { create } from 'zustand';

interface Organization {
  id: string;
  name: string;
  slug: string;
  createdAt: string;
}

interface OrganizationMember {
  id: string;
  userId: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member';
  joinedAt: string;
}

interface OrganizationState {
  currentOrg: Organization | null;
  members: OrganizationMember[];
  isLoading: boolean;
  error: string | null;

  setCurrentOrg: (org: Organization | null) => void;
  setMembers: (members: OrganizationMember[]) => void;
  addMember: (member: OrganizationMember) => void;
  removeMember: (memberId: string) => void;
  updateMemberRole: (memberId: string, role: OrganizationMember['role']) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  currentOrg: null,
  members: [],
  isLoading: false,
  error: null,
};

export const useOrganizationStore = create<OrganizationState>((set) => ({
  ...initialState,

  setCurrentOrg: (org) => set({ currentOrg: org }),

  setMembers: (members) => set({ members }),

  addMember: (member) =>
    set((state) => ({
      members: [...state.members, member],
    })),

  removeMember: (memberId) =>
    set((state) => ({
      members: state.members.filter((m) => m.id !== memberId),
    })),

  updateMemberRole: (memberId, role) =>
    set((state) => ({
      members: state.members.map((m) =>
        m.id === memberId ? { ...m, role } : m
      ),
    })),

  setLoading: (isLoading) => set({ isLoading }),

  setError: (error) => set({ error }),

  reset: () => set(initialState),
}));
