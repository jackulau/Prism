import { useState, useEffect } from 'react';
import {
  Shield,
  Plus,
  Edit2,
  Trash2,
  X,
  Loader2,
  Check,
  AlertTriangle,
  Lock,
} from 'lucide-react';
import { useOrganizationStore } from '../../store/organizationStore';
import { useTeamStore } from '../../store/teamStore';
import { teamService } from '../../services/team';
import { toast } from '../../store/toastStore';
import { ConfirmDialog } from '../ConfirmDialog';
import {
  type Role,
  type Permission,
  PERMISSION_LABELS,
  groupPermissionsByCategory,
} from '../../types/team';

export function RoleManager() {
  const [isCreating, setIsCreating] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [deleteRoleId, setDeleteRoleId] = useState<string | null>(null);

  const { currentOrg } = useOrganizationStore();
  const { roles, isLoadingRoles, addRole, updateRole, removeRole } = useTeamStore();

  // Load roles on mount
  useEffect(() => {
    if (currentOrg?.id) {
      teamService.loadRoles(currentOrg.id);
    }
  }, [currentOrg?.id]);

  const handleDeleteRole = async () => {
    if (!currentOrg?.id || !deleteRoleId) return;

    const response = await teamService.deleteRole(currentOrg.id, deleteRoleId);

    if (!response.error) {
      removeRole(deleteRoleId);
      toast.success('Role deleted');
    } else {
      toast.error(response.error);
    }

    setDeleteRoleId(null);
  };

  const builtInRoles = roles.filter((r) => r.isBuiltIn);
  const customRoles = roles.filter((r) => !r.isBuiltIn);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-editor-text">Roles & Permissions</h2>
          <p className="text-sm text-editor-muted">
            Define what team members can do in your organization
          </p>
        </div>
        <button
          onClick={() => setIsCreating(true)}
          className="flex items-center gap-2 px-3 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors text-sm"
        >
          <Plus size={16} />
          Create Role
        </button>
      </div>

      {/* Loading State */}
      {isLoadingRoles ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 size={24} className="animate-spin text-editor-muted" />
        </div>
      ) : (
        <>
          {/* Built-in Roles */}
          <section className="space-y-3">
            <h3 className="text-sm font-medium text-editor-muted uppercase tracking-wider">
              Built-in Roles
            </h3>
            <div className="space-y-3">
              {builtInRoles.map((role) => (
                <RoleCard
                  key={role.id}
                  role={role}
                  onEdit={() => setEditingRole(role)}
                />
              ))}
            </div>
          </section>

          {/* Custom Roles */}
          {customRoles.length > 0 && (
            <section className="space-y-3">
              <h3 className="text-sm font-medium text-editor-muted uppercase tracking-wider">
                Custom Roles
              </h3>
              <div className="space-y-3">
                {customRoles.map((role) => (
                  <RoleCard
                    key={role.id}
                    role={role}
                    onEdit={() => setEditingRole(role)}
                    onDelete={() => setDeleteRoleId(role.id)}
                  />
                ))}
              </div>
            </section>
          )}
        </>
      )}

      {/* Create/Edit Role Modal */}
      {(isCreating || editingRole) && (
        <RoleEditor
          role={editingRole}
          onClose={() => {
            setIsCreating(false);
            setEditingRole(null);
          }}
          onSave={async (data) => {
            if (!currentOrg?.id) return;

            if (editingRole) {
              const response = await teamService.updateRole(
                currentOrg.id,
                editingRole.id,
                data
              );
              if (response.data) {
                updateRole(editingRole.id, response.data);
                toast.success('Role updated');
              } else if (response.error) {
                toast.error(response.error);
                return;
              }
            } else {
              const response = await teamService.createRole(currentOrg.id, {
                name: data.name!,
                description: data.description!,
                permissions: data.permissions!,
              });
              if (response.data) {
                addRole(response.data);
                toast.success('Role created');
              } else if (response.error) {
                toast.error(response.error);
                return;
              }
            }

            setIsCreating(false);
            setEditingRole(null);
          }}
        />
      )}

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteRoleId !== null}
        title="Delete Role"
        message="Are you sure you want to delete this role? Members with this role will lose their permissions."
        confirmText="Delete"
        variant="danger"
        onConfirm={handleDeleteRole}
        onCancel={() => setDeleteRoleId(null)}
      />
    </div>
  );
}

interface RoleCardProps {
  role: Role;
  onEdit: () => void;
  onDelete?: () => void;
}

function RoleCard({ role, onEdit, onDelete }: RoleCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
      <div className="flex items-center justify-between p-4">
        <div
          className="flex items-center gap-3 flex-1 cursor-pointer"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          <div className="w-10 h-10 rounded-lg bg-editor-accent/10 flex items-center justify-center">
            <Shield size={20} className="text-editor-accent" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-editor-text">{role.name}</span>
              {role.isBuiltIn && (
                <span className="flex items-center gap-1 px-2 py-0.5 bg-editor-muted/10 rounded text-xs text-editor-muted">
                  <Lock size={10} />
                  Built-in
                </span>
              )}
            </div>
            <p className="text-sm text-editor-muted">{role.description}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-xs text-editor-muted px-2 py-1 bg-editor-bg rounded">
            {role.permissions.length} permissions
          </span>
          {!role.isBuiltIn && (
            <>
              <button
                onClick={onEdit}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
                title="Edit role"
              >
                <Edit2 size={16} />
              </button>
              {onDelete && (
                <button
                  onClick={onDelete}
                  className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                  title="Delete role"
                >
                  <Trash2 size={16} />
                </button>
              )}
            </>
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="px-4 pb-4 pt-2 border-t border-editor-border">
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {role.permissions.map((permission) => (
              <div
                key={permission}
                className="flex items-center gap-2 px-2 py-1.5 bg-editor-bg rounded text-sm"
              >
                <Check size={12} className="text-editor-success flex-shrink-0" />
                <span className="text-editor-text truncate">
                  {PERMISSION_LABELS[permission]?.label || permission}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface RoleEditorProps {
  role: Role | null;
  onClose: () => void;
  onSave: (data: {
    name?: string;
    description?: string;
    permissions?: Permission[];
  }) => Promise<void>;
}

function RoleEditor({ role, onClose, onSave }: RoleEditorProps) {
  const [name, setName] = useState(role?.name || '');
  const [description, setDescription] = useState(role?.description || '');
  const [permissions, setPermissions] = useState<Set<Permission>>(
    new Set(role?.permissions || [])
  );
  const [isSaving, setIsSaving] = useState(false);

  const permissionGroups = groupPermissionsByCategory();

  const togglePermission = (permission: Permission) => {
    const newPermissions = new Set(permissions);
    if (newPermissions.has(permission)) {
      newPermissions.delete(permission);
    } else {
      newPermissions.add(permission);
    }
    setPermissions(newPermissions);
  };

  const toggleCategory = (category: string) => {
    const categoryPermissions = permissionGroups[category];
    const allSelected = categoryPermissions.every((p) => permissions.has(p));

    const newPermissions = new Set(permissions);
    categoryPermissions.forEach((p) => {
      if (allSelected) {
        newPermissions.delete(p);
      } else {
        newPermissions.add(p);
      }
    });
    setPermissions(newPermissions);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsSaving(true);
    await onSave({
      name: name.trim(),
      description: description.trim(),
      permissions: Array.from(permissions),
    });
    setIsSaving(false);
  };

  const isBuiltIn = role?.isBuiltIn || false;

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] max-w-[90vw] max-h-[85vh] overflow-hidden bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 flex flex-col">
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border flex-shrink-0">
          <h3 className="text-lg font-semibold text-editor-text">
            {role ? 'Edit Role' : 'Create Role'}
          </h3>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto">
          <div className="p-6 space-y-6">
            {/* Basic Info */}
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Role Name
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., Developer, QA Engineer"
                  disabled={isBuiltIn}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent disabled:opacity-50"
                  required
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Description
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="What can users with this role do?"
                  rows={2}
                  disabled={isBuiltIn}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none disabled:opacity-50"
                />
              </div>
            </div>

            {/* Permissions */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-editor-text">
                  Permissions
                </label>
                <span className="text-sm text-editor-muted">
                  {permissions.size} selected
                </span>
              </div>

              {isBuiltIn && (
                <div className="flex items-center gap-2 p-3 bg-editor-warning/10 border border-editor-warning/20 rounded-lg text-sm">
                  <AlertTriangle size={16} className="text-editor-warning flex-shrink-0" />
                  <span className="text-editor-warning">
                    Built-in roles cannot be modified
                  </span>
                </div>
              )}

              <div className="space-y-4">
                {Object.entries(permissionGroups).map(([category, perms]) => {
                  const allSelected = perms.every((p) => permissions.has(p));
                  const someSelected = perms.some((p) => permissions.has(p));

                  return (
                    <div
                      key={category}
                      className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden"
                    >
                      <button
                        type="button"
                        onClick={() => !isBuiltIn && toggleCategory(category)}
                        disabled={isBuiltIn}
                        className="w-full flex items-center justify-between px-4 py-3 hover:bg-editor-bg/50 disabled:hover:bg-transparent transition-colors"
                      >
                        <span className="font-medium text-editor-text">
                          {category}
                        </span>
                        <div
                          className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                            allSelected
                              ? 'bg-editor-accent border-editor-accent'
                              : someSelected
                              ? 'bg-editor-accent/30 border-editor-accent'
                              : 'border-editor-muted'
                          } ${isBuiltIn ? 'opacity-50' : ''}`}
                        >
                          {(allSelected || someSelected) && (
                            <Check size={12} className="text-white" />
                          )}
                        </div>
                      </button>

                      <div className="px-4 pb-3 pt-1 space-y-2">
                        {perms.map((permission) => (
                          <label
                            key={permission}
                            className={`flex items-center gap-3 px-3 py-2 rounded cursor-pointer transition-colors ${
                              isBuiltIn
                                ? 'opacity-50 cursor-not-allowed'
                                : 'hover:bg-editor-bg/50'
                            }`}
                          >
                            <input
                              type="checkbox"
                              checked={permissions.has(permission)}
                              onChange={() => togglePermission(permission)}
                              disabled={isBuiltIn}
                              className="sr-only"
                            />
                            <div
                              className={`w-4 h-4 rounded border-2 flex items-center justify-center transition-colors ${
                                permissions.has(permission)
                                  ? 'bg-editor-accent border-editor-accent'
                                  : 'border-editor-muted'
                              }`}
                            >
                              {permissions.has(permission) && (
                                <Check size={10} className="text-white" />
                              )}
                            </div>
                            <span className="text-sm text-editor-text">
                              {PERMISSION_LABELS[permission]?.label || permission}
                            </span>
                          </label>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-editor-border flex-shrink-0">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
            >
              Cancel
            </button>
            {!isBuiltIn && (
              <button
                type="submit"
                disabled={isSaving || !name.trim()}
                className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
              >
                {isSaving ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Check size={16} />
                    {role ? 'Save Changes' : 'Create Role'}
                  </>
                )}
              </button>
            )}
          </div>
        </form>
      </div>
    </>
  );
}
