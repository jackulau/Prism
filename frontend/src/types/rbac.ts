// Role types for RBAC
export type Role = 'user' | 'admin';

// Permission identifiers matching backend
export type Permission =
  | 'conversations:view'
  | 'conversations:create'
  | 'own_data:manage'
  | 'users:manage'
  | 'organization:manage'
  | 'conversations:view_all'
  | 'providers:manage'
  | 'settings:manage';

// Permission description for UI
export interface PermissionInfo {
  id: Permission;
  name: string;
  description: string;
}

// All available permissions with descriptions
export const PERMISSION_INFO: Record<Permission, PermissionInfo> = {
  'conversations:view': {
    id: 'conversations:view',
    name: 'View Conversations',
    description: 'View your own conversations',
  },
  'conversations:create': {
    id: 'conversations:create',
    name: 'Create Conversations',
    description: 'Create new conversations',
  },
  'own_data:manage': {
    id: 'own_data:manage',
    name: 'Manage Own Data',
    description: 'Manage your own data and settings',
  },
  'users:manage': {
    id: 'users:manage',
    name: 'Manage Users',
    description: 'View and manage all users',
  },
  'organization:manage': {
    id: 'organization:manage',
    name: 'Manage Organization',
    description: 'Manage organization settings',
  },
  'conversations:view_all': {
    id: 'conversations:view_all',
    name: 'View All Conversations',
    description: 'View all conversations in the organization',
  },
  'providers:manage': {
    id: 'providers:manage',
    name: 'Manage Providers',
    description: 'Manage LLM providers and API keys',
  },
  'settings:manage': {
    id: 'settings:manage',
    name: 'Manage Settings',
    description: 'Manage system-wide settings',
  },
};

// Role to permissions mapping
export const ROLE_PERMISSIONS: Record<Role, Permission[]> = {
  user: [
    'conversations:view',
    'conversations:create',
    'own_data:manage',
  ],
  admin: [
    'conversations:view',
    'conversations:create',
    'own_data:manage',
    'users:manage',
    'organization:manage',
    'conversations:view_all',
    'providers:manage',
    'settings:manage',
  ],
};

// Valid roles for validation
export const VALID_ROLES: Role[] = ['user', 'admin'];

// Check if a role string is valid
export function isValidRole(role: string): role is Role {
  return VALID_ROLES.includes(role as Role);
}

// Check if a role has a specific permission
export function hasPermission(role: Role | undefined, permission: Permission): boolean {
  if (!role || !isValidRole(role)) {
    return false;
  }
  return ROLE_PERMISSIONS[role].includes(permission);
}

// Get all permissions for a role
export function getPermissions(role: Role): Permission[] {
  return ROLE_PERMISSIONS[role] || [];
}

// Get permission info for display
export function getPermissionInfo(permission: Permission): PermissionInfo {
  return PERMISSION_INFO[permission];
}

// Role display names
export const ROLE_DISPLAY_NAMES: Record<Role, string> = {
  user: 'User',
  admin: 'Administrator',
};

// Get display name for a role
export function getRoleDisplayName(role: Role): string {
  return ROLE_DISPLAY_NAMES[role] || role;
}
