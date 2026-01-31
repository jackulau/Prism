import { useState } from 'react';
import { X, Loader2 } from 'lucide-react';

interface CreateAPIKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (name: string, expiresInDays?: number, scopes?: string[]) => Promise<boolean>;
  isCreating: boolean;
}

const EXPIRATION_OPTIONS = [
  { label: 'Never', value: undefined },
  { label: '30 days', value: 30 },
  { label: '90 days', value: 90 },
  { label: '1 year', value: 365 },
];

const SCOPE_OPTIONS = [
  { id: 'read', label: 'Read', description: 'Access conversations and files' },
  { id: 'write', label: 'Write', description: 'Create and modify data' },
  { id: 'execute', label: 'Execute', description: 'Run code and commands' },
  { id: 'admin', label: 'Admin', description: 'Full administrative access' },
];

export function CreateAPIKeyModal({
  isOpen,
  onClose,
  onCreate,
  isCreating,
}: CreateAPIKeyModalProps) {
  const [name, setName] = useState('');
  const [expiration, setExpiration] = useState<number | undefined>(undefined);
  const [selectedScopes, setSelectedScopes] = useState<string[]>(['read', 'write']);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    const success = await onCreate(
      name.trim(),
      expiration,
      selectedScopes.length > 0 ? selectedScopes : undefined
    );

    if (success) {
      // Reset form
      setName('');
      setExpiration(undefined);
      setSelectedScopes(['read', 'write']);
      onClose();
    }
  };

  const handleScopeToggle = (scopeId: string) => {
    setSelectedScopes((prev) =>
      prev.includes(scopeId)
        ? prev.filter((s) => s !== scopeId)
        : [...prev, scopeId]
    );
  };

  const handleClose = () => {
    if (!isCreating) {
      setName('');
      setExpiration(undefined);
      setSelectedScopes(['read', 'write']);
      setError(null);
      onClose();
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[480px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50">
        <div className="flex items-center justify-between p-4 border-b border-editor-border">
          <h3 className="text-lg font-semibold">Create API Key</h3>
          <button
            onClick={handleClose}
            disabled={isCreating}
            className="p-1 hover:bg-editor-surface rounded transition-colors disabled:opacity-50"
          >
            <X className="w-5 h-5 text-editor-muted" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Name <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Production Integration"
              disabled={isCreating}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm focus:outline-none focus:border-editor-accent disabled:opacity-50"
              autoFocus
            />
          </div>

          {/* Expiration */}
          <div>
            <label className="block text-sm font-medium mb-2">Expiration</label>
            <select
              value={expiration ?? ''}
              onChange={(e) => setExpiration(e.target.value ? Number(e.target.value) : undefined)}
              disabled={isCreating}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm focus:outline-none focus:border-editor-accent disabled:opacity-50"
            >
              {EXPIRATION_OPTIONS.map((option) => (
                <option key={option.label} value={option.value ?? ''}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          {/* Scopes */}
          <div>
            <label className="block text-sm font-medium mb-2">Permissions</label>
            <div className="space-y-2">
              {SCOPE_OPTIONS.map((scope) => (
                <label
                  key={scope.id}
                  className={`flex items-start gap-3 p-3 rounded-lg cursor-pointer transition-colors ${
                    selectedScopes.includes(scope.id)
                      ? 'bg-editor-accent/10 border border-editor-accent/30'
                      : 'bg-editor-surface border border-editor-border hover:border-editor-muted'
                  } ${isCreating ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  <input
                    type="checkbox"
                    checked={selectedScopes.includes(scope.id)}
                    onChange={() => handleScopeToggle(scope.id)}
                    disabled={isCreating}
                    className="mt-0.5 rounded border-editor-border text-editor-accent focus:ring-editor-accent"
                  />
                  <div>
                    <span className="font-medium text-sm">{scope.label}</span>
                    <p className="text-xs text-editor-muted mt-0.5">{scope.description}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Error */}
          {error && (
            <div className="text-sm text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg p-3">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={handleClose}
              disabled={isCreating}
              className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isCreating || !name.trim()}
              className="px-4 py-2 text-sm bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2"
            >
              {isCreating && <Loader2 className="w-4 h-4 animate-spin" />}
              Create Key
            </button>
          </div>
        </form>
      </div>
    </>
  );
}
