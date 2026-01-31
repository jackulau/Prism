import { useState, useEffect } from 'react';
import { X, Loader2, Briefcase } from 'lucide-react';
import type { OrgWorkspace, CreateOrgWorkspaceInput } from '../../types/organization';

interface OrgWorkspaceModalProps {
  isOpen: boolean;
  workspace?: OrgWorkspace | null;
  isSubmitting: boolean;
  onClose: () => void;
  onSubmit: (data: CreateOrgWorkspaceInput) => void;
}

interface FormErrors {
  name?: string;
  githubRepositoryName?: string;
  slackChannelId?: string;
}

export function OrgWorkspaceModal({
  isOpen,
  workspace,
  isSubmitting,
  onClose,
  onSubmit,
}: OrgWorkspaceModalProps) {
  const [name, setName] = useState('');
  const [githubRepositoryName, setGithubRepositoryName] = useState('');
  const [slackChannelId, setSlackChannelId] = useState('');
  const [errors, setErrors] = useState<FormErrors>({});

  const isEditing = !!workspace;

  useEffect(() => {
    if (isOpen) {
      if (workspace) {
        setName(workspace.name);
        setGithubRepositoryName(workspace.githubRepositoryName || '');
        setSlackChannelId(workspace.slackChannelId || '');
      } else {
        setName('');
        setGithubRepositoryName('');
        setSlackChannelId('');
      }
      setErrors({});
    }
  }, [isOpen, workspace]);

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!name.trim()) {
      newErrors.name = 'Name is required';
    } else if (name.length < 2) {
      newErrors.name = 'Name must be at least 2 characters';
    } else if (name.length > 100) {
      newErrors.name = 'Name must be less than 100 characters';
    }

    if (githubRepositoryName && !/^[\w.-]+\/[\w.-]+$/.test(githubRepositoryName)) {
      newErrors.githubRepositoryName = 'Format: owner/repository';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    onSubmit({
      name: name.trim(),
      githubRepositoryName: githubRepositoryName.trim() || undefined,
      slackChannelId: slackChannelId.trim() || undefined,
    });
  };

  if (!isOpen) return null;

  return (
    <>
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={onClose}
      />
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[480px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50">
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <h3 className="text-lg font-semibold text-editor-text">
            {isEditing ? 'Edit Workspace' : 'Create Workspace'}
          </h3>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Workspace Name <span className="text-editor-error">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Workspace"
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent ${
                errors.name ? 'border-editor-error' : 'border-editor-border'
              }`}
              autoFocus
            />
            {errors.name && (
              <p className="text-sm text-editor-error">{errors.name}</p>
            )}
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              GitHub Repository
            </label>
            <input
              type="text"
              value={githubRepositoryName}
              onChange={(e) => setGithubRepositoryName(e.target.value)}
              placeholder="owner/repository"
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent ${
                errors.githubRepositoryName
                  ? 'border-editor-error'
                  : 'border-editor-border'
              }`}
            />
            {errors.githubRepositoryName && (
              <p className="text-sm text-editor-error">
                {errors.githubRepositoryName}
              </p>
            )}
            <p className="text-xs text-editor-muted">
              Optional. Link a GitHub repository to this workspace.
            </p>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Slack Channel ID
            </label>
            <input
              type="text"
              value={slackChannelId}
              onChange={(e) => setSlackChannelId(e.target.value)}
              placeholder="C01234567"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            />
            <p className="text-xs text-editor-muted">
              Optional. Enable Slack notifications for this workspace.
            </p>
          </div>

          <div className="flex items-center justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
            >
              {isSubmitting ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  {isEditing ? 'Saving...' : 'Creating...'}
                </>
              ) : (
                <>
                  <Briefcase size={16} />
                  {isEditing ? 'Save Changes' : 'Create Workspace'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </>
  );
}
