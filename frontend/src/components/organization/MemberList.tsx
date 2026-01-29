import { useState } from 'react';
import { Users, UserPlus, Shield, User, Trash2, X, Loader2 } from 'lucide-react';
import { toast } from '../../store/toastStore';

interface Member {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member';
  joinedAt: string;
  avatarUrl?: string;
}

// Mock data - would come from tRPC in real implementation
const MOCK_MEMBERS: Member[] = [
  {
    id: '1',
    name: 'Current User',
    email: 'user@example.com',
    role: 'owner',
    joinedAt: '2024-01-01',
  },
];

export function MemberList() {
  const [isInviting, setIsInviting] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member'>('member');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const members = MOCK_MEMBERS;

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteEmail) return;

    setIsSubmitting(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));
    toast.success(`Invitation sent to ${inviteEmail}`);
    setInviteEmail('');
    setIsInviting(false);
    setIsSubmitting(false);
  };

  const getRoleIcon = (role: Member['role']) => {
    switch (role) {
      case 'owner':
        return <Shield size={14} className="text-editor-warning" />;
      case 'admin':
        return <Shield size={14} className="text-editor-accent" />;
      default:
        return <User size={14} className="text-editor-muted" />;
    }
  };

  const getRoleBadgeColor = (role: Member['role']) => {
    switch (role) {
      case 'owner':
        return 'bg-editor-warning/10 text-editor-warning';
      case 'admin':
        return 'bg-editor-accent/10 text-editor-accent';
      default:
        return 'bg-editor-muted/10 text-editor-muted';
    }
  };

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Team Members</h2>
        <button
          onClick={() => setIsInviting(true)}
          className="flex items-center gap-2 px-3 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors text-sm"
        >
          <UserPlus size={16} />
          Invite Member
        </button>
      </div>

      {/* Invite Modal */}
      {isInviting && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={() => setIsInviting(false)}
          />
          <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50">
            <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
              <h3 className="text-lg font-semibold text-editor-text">Invite Team Member</h3>
              <button
                onClick={() => setIsInviting(false)}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            <form onSubmit={handleInvite} className="p-6 space-y-4">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Email Address
                </label>
                <input
                  type="email"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="colleague@example.com"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                  required
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-editor-text">
                  Role
                </label>
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member')}
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
                >
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setIsInviting(false)}
                  className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || !inviteEmail}
                  className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 size={16} className="animate-spin" />
                      Sending...
                    </>
                  ) : (
                    <>
                      <UserPlus size={16} />
                      Send Invite
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </>
      )}

      {/* Member List */}
      <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
        <div className="divide-y divide-editor-border">
          {members.map((member) => (
            <div
              key={member.id}
              className="flex items-center justify-between p-4 hover:bg-editor-bg/50 transition-colors"
            >
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 bg-editor-accent/10 rounded-full flex items-center justify-center">
                  <span className="text-editor-accent font-medium">
                    {member.name.charAt(0).toUpperCase()}
                  </span>
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-editor-text">{member.name}</span>
                    <span
                      className={`flex items-center gap-1 px-2 py-0.5 text-xs rounded-full ${getRoleBadgeColor(
                        member.role
                      )}`}
                    >
                      {getRoleIcon(member.role)}
                      {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                    </span>
                  </div>
                  <span className="text-sm text-editor-muted">{member.email}</span>
                </div>
              </div>

              {member.role !== 'owner' && (
                <button
                  className="p-2 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
                  title="Remove member"
                >
                  <Trash2 size={16} />
                </button>
              )}
            </div>
          ))}
        </div>

        {members.length === 0 && (
          <div className="p-8 text-center">
            <Users className="w-12 h-12 text-editor-muted mx-auto mb-4" />
            <p className="text-editor-muted">No team members yet</p>
            <p className="text-sm text-editor-muted mt-1">
              Invite colleagues to collaborate
            </p>
          </div>
        )}
      </div>
    </section>
  );
}
