import { useState } from 'react';
import { Building, Pencil, Check, X } from 'lucide-react';
import { trpc } from '../../lib/trpc';
import { toast } from '../../store/toastStore';

export function OrgNameEditor() {
  const [isEditing, setIsEditing] = useState(false);
  const [newName, setNewName] = useState('');

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: profile, isLoading } = (trpc as any).organization.getProfile.useQuery();

  // Note: The organization router currently has user profile management
  // This would need to be extended for actual organization management
  // For now, we'll display user profile info

  const handleStartEdit = () => {
    setNewName(profile?.name || '');
    setIsEditing(true);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setNewName('');
  };

  const handleSave = () => {
    // This would call an organization update mutation when available
    toast.info('Organization rename feature coming soon');
    setIsEditing(false);
  };

  if (isLoading) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Organization Details</h2>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-6 animate-pulse">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-editor-border rounded-lg" />
            <div className="flex-1">
              <div className="h-5 bg-editor-border rounded w-1/3 mb-2" />
              <div className="h-4 bg-editor-border rounded w-1/4" />
            </div>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold text-editor-text">Organization Details</h2>

      <div className="bg-editor-surface border border-editor-border rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-editor-accent/10 rounded-lg">
              <Building className="w-6 h-6 text-editor-accent" />
            </div>

            {isEditing ? (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="Organization name"
                  className="px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
                  autoFocus
                />
                <button
                  onClick={handleSave}
                  className="p-2 text-editor-success hover:bg-editor-success/10 rounded-lg transition-colors"
                  title="Save"
                >
                  <Check size={18} />
                </button>
                <button
                  onClick={handleCancel}
                  className="p-2 text-editor-muted hover:bg-editor-surface rounded-lg transition-colors"
                  title="Cancel"
                >
                  <X size={18} />
                </button>
              </div>
            ) : (
              <div>
                <h3 className="text-lg font-medium text-editor-text">
                  {profile?.name || 'My Organization'}
                </h3>
                <p className="text-sm text-editor-muted">{profile?.email}</p>
              </div>
            )}
          </div>

          {!isEditing && (
            <button
              onClick={handleStartEdit}
              className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
            >
              <Pencil size={16} />
              Edit
            </button>
          )}
        </div>
      </div>
    </section>
  );
}
