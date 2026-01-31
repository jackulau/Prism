import { useState, useEffect } from 'react';
import {
  ArrowLeft,
  Users,
  UserPlus,
  Trash2,
  X,
  Loader2,
  Edit2,
  Check,
} from 'lucide-react';
import { useOrganizationStore } from '../../store/organizationStore';
import { useTeamStore } from '../../store/teamStore';
import { teamService } from '../../services/team';
import { toast } from '../../store/toastStore';
import { ConfirmDialog } from '../ConfirmDialog';
import type { Team, TeamMember, Role } from '../../types/team';

interface TeamDetailProps {
  team: Team;
  onBack: () => void;
}

export function TeamDetail({ team, onBack }: TeamDetailProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState(team.name);
  const [editDescription, setEditDescription] = useState(team.description || '');
  const [isSaving, setIsSaving] = useState(false);

  const [isAddingMember, setIsAddingMember] = useState(false);
  const [newMemberEmail, setNewMemberEmail] = useState('');
  const [selectedRoleId, setSelectedRoleId] = useState('');
  const [isAddingSubmitting, setIsAddingSubmitting] = useState(false);

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [removeMemberId, setRemoveMemberId] = useState<string | null>(null);

  const { currentOrg } = useOrganizationStore();
  const {
    membersByTeam,
    roles,
    isLoadingMembers,
    isLoadingRoles,
    updateTeam,
    removeTeam,
    addTeamMember,
    removeTeamMember,
    updateTeamMember,
  } = useTeamStore();

  const members = membersByTeam[team.id] || [];

  // Load members and roles on mount
  useEffect(() => {
    if (currentOrg?.id) {
      teamService.loadTeamMembers(currentOrg.id, team.id);
      teamService.loadRoles(currentOrg.id);
    }
  }, [currentOrg?.id, team.id]);

  // Set default role when roles load
  useEffect(() => {
    if (roles.length > 0 && !selectedRoleId) {
      const memberRole = roles.find((r) => r.name === 'Member');
      setSelectedRoleId(memberRole?.id || roles[0].id);
    }
  }, [roles, selectedRoleId]);

  const handleSaveDetails = async () => {
    if (!currentOrg?.id || !editName.trim()) return;

    setIsSaving(true);
    const response = await teamService.updateTeam(currentOrg.id, team.id, {
      name: editName.trim(),
      description: editDescription.trim() || undefined,
    });

    if (response.data) {
      updateTeam(team.id, response.data);
      toast.success('Team updated');
      setIsEditing(false);
    } else if (response.error) {
      toast.error(response.error);
    }

    setIsSaving(false);
  };

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentOrg?.id || !newMemberEmail.trim() || !selectedRoleId) return;

    setIsAddingSubmitting(true);
    const response = await teamService.addTeamMember(currentOrg.id, team.id, {
      email: newMemberEmail.trim(),
      roleId: selectedRoleId,
    });

    if (response.data) {
      addTeamMember(team.id, response.data);
      toast.success(`${newMemberEmail} added to team`);
      setNewMemberEmail('');
      setIsAddingMember(false);
    } else if (response.error) {
      toast.error(response.error);
    }

    setIsAddingSubmitting(false);
  };

  const handleRemoveMember = async () => {
    if (!currentOrg?.id || !removeMemberId) return;

    const response = await teamService.removeTeamMember(
      currentOrg.id,
      team.id,
      removeMemberId
    );

    if (!response.error) {
      removeTeamMember(team.id, removeMemberId);
      toast.success('Member removed from team');
    } else {
      toast.error(response.error);
    }

    setRemoveMemberId(null);
  };

  const handleRoleChange = async (memberId: string, roleId: string) => {
    if (!currentOrg?.id) return;

    const response = await teamService.updateTeamMember(
      currentOrg.id,
      team.id,
      memberId,
      { roleId }
    );

    if (response.data) {
      updateTeamMember(team.id, memberId, { role: response.data.role });
      toast.success('Role updated');
    } else if (response.error) {
      toast.error(response.error);
    }
  };

  const handleDeleteTeam = async () => {
    if (!currentOrg?.id) return;

    setIsDeleting(true);
    const response = await teamService.deleteTeam(currentOrg.id, team.id);

    if (!response.error) {
      removeTeam(team.id);
      toast.success('Team deleted');
      onBack();
    } else {
      toast.error(response.error);
    }

    setIsDeleting(false);
    setDeleteConfirmOpen(false);
  };

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Back Button */}
        <button
          onClick={onBack}
          className="flex items-center gap-2 text-editor-muted hover:text-editor-text transition-colors"
        >
          <ArrowLeft size={18} />
          <span>Back to Teams</span>
        </button>

        {/* Team Header */}
        <div className="bg-editor-surface border border-editor-border rounded-lg p-6">
          {isEditing ? (
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Team Name
                </label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Description
                </label>
                <textarea
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent resize-none"
                />
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={handleSaveDetails}
                  disabled={isSaving || !editName.trim()}
                  className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
                >
                  {isSaving ? (
                    <Loader2 size={16} className="animate-spin" />
                  ) : (
                    <Check size={16} />
                  )}
                  Save
                </button>
                <button
                  onClick={() => {
                    setIsEditing(false);
                    setEditName(team.name);
                    setEditDescription(team.description || '');
                  }}
                  className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-4">
                <div className="w-16 h-16 rounded-xl bg-editor-accent/10 flex items-center justify-center">
                  <span className="text-2xl font-bold text-editor-accent">
                    {team.name.charAt(0).toUpperCase()}
                  </span>
                </div>
                <div>
                  <h1 className="text-2xl font-bold text-editor-text">
                    {team.name}
                  </h1>
                  {team.description && (
                    <p className="text-editor-muted mt-1">{team.description}</p>
                  )}
                  <div className="flex items-center gap-4 mt-2 text-sm text-editor-muted">
                    <span>
                      {team.memberCount} member{team.memberCount !== 1 ? 's' : ''}
                    </span>
                    <span>Created {new Date(team.createdAt).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
                  title="Edit team"
                >
                  <Edit2 size={18} />
                </button>
                <button
                  onClick={() => setDeleteConfirmOpen(true)}
                  className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                  title="Delete team"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Members Section */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-editor-text">Members</h2>
            <button
              onClick={() => setIsAddingMember(true)}
              className="flex items-center gap-2 px-3 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors text-sm"
            >
              <UserPlus size={16} />
              Add Member
            </button>
          </div>

          {/* Add Member Modal */}
          {isAddingMember && (
            <>
              <div
                className="fixed inset-0 bg-black/50 z-40"
                onClick={() => setIsAddingMember(false)}
              />
              <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50">
                <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
                  <h3 className="text-lg font-semibold text-editor-text">
                    Add Team Member
                  </h3>
                  <button
                    onClick={() => setIsAddingMember(false)}
                    className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
                  >
                    <X size={20} />
                  </button>
                </div>

                <form onSubmit={handleAddMember} className="p-6 space-y-4">
                  <div className="space-y-2">
                    <label className="block text-sm font-medium text-editor-text">
                      Email Address
                    </label>
                    <input
                      type="email"
                      value={newMemberEmail}
                      onChange={(e) => setNewMemberEmail(e.target.value)}
                      placeholder="colleague@example.com"
                      className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                      required
                      autoFocus
                    />
                  </div>

                  <div className="space-y-2">
                    <label className="block text-sm font-medium text-editor-text">
                      Role
                    </label>
                    {isLoadingRoles ? (
                      <div className="flex items-center gap-2 text-editor-muted">
                        <Loader2 size={16} className="animate-spin" />
                        Loading roles...
                      </div>
                    ) : (
                      <select
                        value={selectedRoleId}
                        onChange={(e) => setSelectedRoleId(e.target.value)}
                        className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
                      >
                        {roles.map((role) => (
                          <option key={role.id} value={role.id}>
                            {role.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>

                  <div className="flex items-center justify-end gap-3 pt-4">
                    <button
                      type="button"
                      onClick={() => setIsAddingMember(false)}
                      className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isAddingSubmitting || !newMemberEmail.trim() || !selectedRoleId}
                      className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
                    >
                      {isAddingSubmitting ? (
                        <>
                          <Loader2 size={16} className="animate-spin" />
                          Adding...
                        </>
                      ) : (
                        <>
                          <UserPlus size={16} />
                          Add Member
                        </>
                      )}
                    </button>
                  </div>
                </form>
              </div>
            </>
          )}

          {/* Members List */}
          <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
            {isLoadingMembers ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 size={20} className="animate-spin text-editor-muted" />
                <span className="ml-3 text-editor-muted">Loading members...</span>
              </div>
            ) : members.length === 0 ? (
              <div className="py-12 text-center">
                <Users size={32} className="mx-auto mb-3 text-editor-muted opacity-50" />
                <p className="text-editor-muted">No members yet</p>
                <button
                  onClick={() => setIsAddingMember(true)}
                  className="mt-4 text-sm text-editor-accent hover:underline"
                >
                  Add the first member
                </button>
              </div>
            ) : (
              <div className="divide-y divide-editor-border">
                {members.map((member) => (
                  <MemberRow
                    key={member.id}
                    member={member}
                    roles={roles}
                    onRoleChange={(roleId) => handleRoleChange(member.id, roleId)}
                    onRemove={() => setRemoveMemberId(member.id)}
                  />
                ))}
              </div>
            )}
          </div>
        </section>

        {/* Delete Team Confirmation */}
        <ConfirmDialog
          isOpen={deleteConfirmOpen}
          title="Delete Team"
          message={`Are you sure you want to delete "${team.name}"? This will remove all team members and cannot be undone.`}
          confirmText={isDeleting ? 'Deleting...' : 'Delete Team'}
          variant="danger"
          onConfirm={handleDeleteTeam}
          onCancel={() => setDeleteConfirmOpen(false)}
        />

        {/* Remove Member Confirmation */}
        <ConfirmDialog
          isOpen={removeMemberId !== null}
          title="Remove Member"
          message="Are you sure you want to remove this member from the team?"
          confirmText="Remove"
          variant="danger"
          onConfirm={handleRemoveMember}
          onCancel={() => setRemoveMemberId(null)}
        />
      </div>
    </div>
  );
}

interface MemberRowProps {
  member: TeamMember;
  roles: Role[];
  onRoleChange: (roleId: string) => void;
  onRemove: () => void;
}

function MemberRow({ member, roles, onRoleChange, onRemove }: MemberRowProps) {
  return (
    <div className="flex items-center justify-between p-4 hover:bg-editor-bg/50 transition-colors">
      <div className="flex items-center gap-4">
        <div className="w-10 h-10 bg-editor-accent/10 rounded-full flex items-center justify-center">
          {member.avatarUrl ? (
            <img
              src={member.avatarUrl}
              alt={member.name}
              className="w-10 h-10 rounded-full"
            />
          ) : (
            <span className="text-editor-accent font-medium">
              {member.name.charAt(0).toUpperCase()}
            </span>
          )}
        </div>
        <div>
          <div className="font-medium text-editor-text">{member.name}</div>
          <div className="text-sm text-editor-muted">{member.email}</div>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <select
          value={member.role.id}
          onChange={(e) => onRoleChange(e.target.value)}
          className="px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
        >
          {roles.map((role) => (
            <option key={role.id} value={role.id}>
              {role.name}
            </option>
          ))}
        </select>

        <button
          onClick={onRemove}
          className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
          title="Remove member"
        >
          <Trash2 size={16} />
        </button>
      </div>
    </div>
  );
}
