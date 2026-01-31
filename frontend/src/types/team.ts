// Team types for team management and role-based access control

export type Permission =
  | 'agents:create'
  | 'agents:read'
  | 'agents:update'
  | 'agents:delete'
  | 'workflows:create'
  | 'workflows:read'
  | 'workflows:update'
  | 'workflows:delete'
  | 'integrations:manage'
  | 'settings:manage'
  | 'team:manage'
  | 'team:invite'
  | 'billing:view'
  | 'billing:manage';

export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: Permission[];
  isBuiltIn: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface TeamMember {
  id: string;
  userId: string;
  teamId: string;
  name: string;
  email: string;
  avatarUrl?: string;
  role: Role;
  joinedAt: string;
}

export interface Team {
  id: string;
  organizationId: string;
  name: string;
  slug: string;
  description?: string;
  memberCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface TeamWithMembers extends Team {
  members: TeamMember[];
}

// Predefined roles that exist in every organization
export const PREDEFINED_ROLES: Omit<Role, 'id' | 'createdAt' | 'updatedAt'>[] = [
  {
    name: 'Admin',
    description: 'Full access to all team resources and settings',
    permissions: [
      'agents:create',
      'agents:read',
      'agents:update',
      'agents:delete',
      'workflows:create',
      'workflows:read',
      'workflows:update',
      'workflows:delete',
      'integrations:manage',
      'settings:manage',
      'team:manage',
      'team:invite',
      'billing:view',
    ],
    isBuiltIn: true,
  },
  {
    name: 'Member',
    description: 'Can view and use team resources',
    permissions: [
      'agents:read',
      'workflows:read',
      'workflows:create',
      'workflows:update',
    ],
    isBuiltIn: true,
  },
  {
    name: 'Viewer',
    description: 'Read-only access to team resources',
    permissions: ['agents:read', 'workflows:read'],
    isBuiltIn: true,
  },
];

// Permission display metadata
export const PERMISSION_LABELS: Record<Permission, { label: string; category: string }> = {
  'agents:create': { label: 'Create Agents', category: 'Agents' },
  'agents:read': { label: 'View Agents', category: 'Agents' },
  'agents:update': { label: 'Edit Agents', category: 'Agents' },
  'agents:delete': { label: 'Delete Agents', category: 'Agents' },
  'workflows:create': { label: 'Create Workflows', category: 'Workflows' },
  'workflows:read': { label: 'View Workflows', category: 'Workflows' },
  'workflows:update': { label: 'Edit Workflows', category: 'Workflows' },
  'workflows:delete': { label: 'Delete Workflows', category: 'Workflows' },
  'integrations:manage': { label: 'Manage Integrations', category: 'Integrations' },
  'settings:manage': { label: 'Manage Settings', category: 'Settings' },
  'team:manage': { label: 'Manage Team', category: 'Team' },
  'team:invite': { label: 'Invite Members', category: 'Team' },
  'billing:view': { label: 'View Billing', category: 'Billing' },
  'billing:manage': { label: 'Manage Billing', category: 'Billing' },
};

// Group permissions by category for UI display
export function groupPermissionsByCategory(): Record<string, Permission[]> {
  const groups: Record<string, Permission[]> = {};
  for (const [permission, meta] of Object.entries(PERMISSION_LABELS)) {
    if (!groups[meta.category]) {
      groups[meta.category] = [];
    }
    groups[meta.category].push(permission as Permission);
  }
  return groups;
}

// API request/response types
export interface CreateTeamRequest {
  name: string;
  description?: string;
  initialMembers?: { email: string; roleId: string }[];
}

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
}

export interface AddTeamMemberRequest {
  email: string;
  roleId: string;
}

export interface UpdateTeamMemberRequest {
  roleId: string;
}

export interface CreateRoleRequest {
  name: string;
  description: string;
  permissions: Permission[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  permissions?: Permission[];
}
