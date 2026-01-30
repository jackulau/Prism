import { useState } from 'react';
import { Plus, Trash2, Lock, Unlock, Upload, AlertCircle } from 'lucide-react';
import { EnvVarInput } from './EnvVarInput';
import { ConfirmDialog } from '../ConfirmDialog';
import type { BuildEnvVar } from '../../services/buildConfig';

interface BuildEnvEditorProps {
  envVars: BuildEnvVar[];
  onAdd: (key: string, value: string, isSecret: boolean) => Promise<void>;
  onDelete: (key: string) => Promise<void>;
  disabled?: boolean;
}

// Validate env var key: alphanumeric + underscore only
function isValidEnvKey(key: string): boolean {
  return /^[A-Za-z_][A-Za-z0-9_]*$/.test(key);
}

export function BuildEnvEditor({
  envVars,
  onAdd,
  onDelete,
  disabled = false,
}: BuildEnvEditorProps) {
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [newIsSecret, setNewIsSecret] = useState(false);
  const [isAdding, setIsAdding] = useState(false);
  const [keyError, setKeyError] = useState<string | null>(null);
  const [deleteKey, setDeleteKey] = useState<string | null>(null);
  const [isImporting, setIsImporting] = useState(false);

  const handleAdd = async () => {
    if (!newKey || !newValue) return;

    // Validate key format
    if (!isValidEnvKey(newKey)) {
      setKeyError('Key must start with a letter or underscore and contain only alphanumeric characters and underscores');
      return;
    }

    // Check for duplicate
    if (envVars.some((v) => v.key === newKey)) {
      setKeyError('This key already exists');
      return;
    }

    setIsAdding(true);
    setKeyError(null);
    try {
      await onAdd(newKey, newValue, newIsSecret);
      setNewKey('');
      setNewValue('');
      setNewIsSecret(false);
    } finally {
      setIsAdding(false);
    }
  };

  const handleKeyChange = (key: string) => {
    setNewKey(key.toUpperCase());
    setKeyError(null);
  };

  const handleImportFile = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setIsImporting(true);
    try {
      const text = await file.text();
      const lines = text.split('\n');

      for (const line of lines) {
        const trimmed = line.trim();
        // Skip comments and empty lines
        if (!trimmed || trimmed.startsWith('#')) continue;

        // Parse KEY=VALUE format
        const eqIndex = trimmed.indexOf('=');
        if (eqIndex === -1) continue;

        const key = trimmed.slice(0, eqIndex).trim();
        let value = trimmed.slice(eqIndex + 1).trim();

        // Remove quotes if present
        if ((value.startsWith('"') && value.endsWith('"')) ||
            (value.startsWith("'") && value.endsWith("'"))) {
          value = value.slice(1, -1);
        }

        if (key && value && isValidEnvKey(key) && !envVars.some((v) => v.key === key)) {
          await onAdd(key, value, false);
        }
      }
    } finally {
      setIsImporting(false);
      // Reset file input
      event.target.value = '';
    }
  };

  const sortedEnvVars = [...envVars].sort((a, b) => a.key.localeCompare(b.key));

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-editor-text">Environment Variables</h3>
        <label className="flex items-center gap-2 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg cursor-pointer transition-colors">
          <Upload size={14} />
          {isImporting ? 'Importing...' : 'Import .env'}
          <input
            type="file"
            accept=".env,.env.local,.env.development,.env.production,text/plain"
            onChange={handleImportFile}
            disabled={disabled || isImporting}
            className="hidden"
          />
        </label>
      </div>

      {/* Env vars table */}
      {sortedEnvVars.length > 0 ? (
        <div className="border border-editor-border rounded-lg overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-editor-surface border-b border-editor-border">
                <th className="px-3 py-2 text-left text-xs font-medium text-editor-muted uppercase tracking-wider w-48">
                  Key
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-editor-muted uppercase tracking-wider">
                  Value
                </th>
                <th className="px-3 py-2 text-center text-xs font-medium text-editor-muted uppercase tracking-wider w-20">
                  Secret
                </th>
                <th className="px-3 py-2 w-12" />
              </tr>
            </thead>
            <tbody className="divide-y divide-editor-border">
              {sortedEnvVars.map((envVar) => (
                <tr key={envVar.key} className="hover:bg-editor-surface/50">
                  <td className="px-3 py-2">
                    <code className="text-sm font-mono text-editor-accent">{envVar.key}</code>
                  </td>
                  <td className="px-3 py-2">
                    <EnvVarInput
                      value={envVar.value}
                      isSecret={envVar.isSecret}
                      readOnly
                    />
                  </td>
                  <td className="px-3 py-2 text-center">
                    <span title={envVar.isSecret ? 'Secret' : 'Not secret'}>
                      {envVar.isSecret ? (
                        <Lock size={14} className="inline text-editor-warning" />
                      ) : (
                        <Unlock size={14} className="inline text-editor-muted" />
                      )}
                    </span>
                  </td>
                  <td className="px-3 py-2">
                    <button
                      onClick={() => setDeleteKey(envVar.key)}
                      disabled={disabled}
                      className="p-1 text-editor-muted hover:text-red-400 transition-colors disabled:opacity-50"
                      title="Delete variable"
                    >
                      <Trash2 size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-center py-8 text-editor-muted text-sm border border-editor-border border-dashed rounded-lg">
          No environment variables defined.
        </div>
      )}

      {/* Add new var form */}
      <div className="p-4 bg-editor-surface rounded-lg space-y-3">
        <h4 className="text-sm font-medium text-editor-text">Add Variable</h4>
        <div className="flex gap-3">
          <div className="w-48">
            <input
              type="text"
              value={newKey}
              onChange={(e) => handleKeyChange(e.target.value)}
              placeholder="KEY_NAME"
              disabled={disabled || isAdding}
              className={`w-full px-3 py-2 bg-editor-bg border rounded-lg text-sm font-mono text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent ${
                keyError ? 'border-red-500' : 'border-editor-border'
              }`}
            />
          </div>
          <div className="flex-1">
            <EnvVarInput
              value={newValue}
              isSecret={newIsSecret}
              onChange={setNewValue}
              placeholder="Value"
            />
          </div>
          <label className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted cursor-pointer select-none">
            <input
              type="checkbox"
              checked={newIsSecret}
              onChange={(e) => setNewIsSecret(e.target.checked)}
              disabled={disabled || isAdding}
              className="rounded border-editor-border bg-editor-bg text-editor-accent focus:ring-editor-accent"
            />
            Secret
          </label>
          <button
            onClick={handleAdd}
            disabled={disabled || isAdding || !newKey || !newValue}
            className="flex items-center gap-1 px-4 py-2 bg-editor-accent text-white text-sm rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Plus size={14} />
            Add
          </button>
        </div>
        {keyError && (
          <div className="flex items-center gap-1 text-xs text-red-400">
            <AlertCircle size={12} />
            {keyError}
          </div>
        )}
      </div>

      {/* Delete confirmation */}
      <ConfirmDialog
        isOpen={!!deleteKey}
        title="Delete Environment Variable"
        message={`Are you sure you want to delete the variable "${deleteKey}"? This action cannot be undone.`}
        confirmText="Delete"
        variant="danger"
        onConfirm={async () => {
          if (deleteKey) {
            await onDelete(deleteKey);
            setDeleteKey(null);
          }
        }}
        onCancel={() => setDeleteKey(null)}
      />
    </div>
  );
}
