import { useState, useEffect, useCallback } from 'react';
import { X, Save, Loader2, AlertTriangle, Star, Trash2 } from 'lucide-react';
import { BuildCommandEditor } from './BuildCommandEditor';
import { BuildEnvEditor } from './BuildEnvEditor';
import { ConfirmDialog } from '../ConfirmDialog';
import type { BuildConfig } from '../../services/buildConfig';

type TabType = 'commands' | 'environment';

interface BuildConfigEditorProps {
  config: BuildConfig;
  onUpdate: (id: string, data: { name?: string; description?: string }) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onSetDefault: (id: string) => Promise<void>;
  onAddCommand: (configId: string, name: string, command: string, workingDirectory?: string) => Promise<void>;
  onUpdateCommand: (configId: string, cmdId: string, data: Partial<BuildConfig['commands'][0]>) => Promise<void>;
  onDeleteCommand: (configId: string, cmdId: string) => Promise<void>;
  onReorderCommands: (configId: string, order: string[]) => Promise<void>;
  onSetEnvVar: (configId: string, key: string, value: string, isSecret: boolean) => Promise<void>;
  onDeleteEnvVar: (configId: string, key: string) => Promise<void>;
  onClose?: () => void;
  isSaving?: boolean;
}

export function BuildConfigEditor({
  config,
  onUpdate,
  onDelete,
  onSetDefault,
  onAddCommand,
  onUpdateCommand,
  onDeleteCommand,
  onReorderCommands,
  onSetEnvVar,
  onDeleteEnvVar,
  onClose,
  isSaving = false,
}: BuildConfigEditorProps) {
  const [activeTab, setActiveTab] = useState<TabType>('commands');
  const [name, setName] = useState(config.name);
  const [description, setDescription] = useState(config.description || '');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showUnsavedWarning, setShowUnsavedWarning] = useState(false);

  // Track changes
  useEffect(() => {
    const nameChanged = name !== config.name;
    const descChanged = description !== (config.description || '');
    setHasUnsavedChanges(nameChanged || descChanged);
  }, [name, description, config.name, config.description]);

  // Reset form when config changes
  useEffect(() => {
    setName(config.name);
    setDescription(config.description || '');
    setHasUnsavedChanges(false);
  }, [config.id]);

  const handleSave = useCallback(async () => {
    if (!name.trim()) return;

    const updates: { name?: string; description?: string } = {};
    if (name !== config.name) updates.name = name;
    if (description !== (config.description || '')) updates.description = description || undefined;

    if (Object.keys(updates).length > 0) {
      await onUpdate(config.id, updates);
    }
    setHasUnsavedChanges(false);
  }, [name, description, config.id, config.name, config.description, onUpdate]);

  const handleClose = useCallback(() => {
    if (hasUnsavedChanges) {
      setShowUnsavedWarning(true);
    } else {
      onClose?.();
    }
  }, [hasUnsavedChanges, onClose]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 's' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        if (hasUnsavedChanges && !isSaving) {
          handleSave();
        }
      }
      if (e.key === 'Escape') {
        handleClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [hasUnsavedChanges, isSaving, handleSave, handleClose]);

  return (
    <div className="flex flex-col h-full bg-editor-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border">
        <div className="flex items-center gap-3">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="text-lg font-semibold bg-transparent border-none focus:outline-none text-editor-text placeholder-editor-muted"
            placeholder="Configuration name"
          />
          {config.isDefault && (
            <span className="flex items-center gap-1 px-2 py-0.5 text-xs bg-editor-warning/20 text-editor-warning rounded">
              <Star size={10} />
              Default
            </span>
          )}
          {hasUnsavedChanges && (
            <span className="flex items-center gap-1 text-xs text-editor-warning">
              <AlertTriangle size={12} />
              Unsaved changes
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          {!config.isDefault && (
            <button
              onClick={() => onSetDefault(config.id)}
              disabled={isSaving}
              className="flex items-center gap-1 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
              title="Set as default"
            >
              <Star size={14} />
              Set Default
            </button>
          )}

          <button
            onClick={() => setShowDeleteConfirm(true)}
            disabled={isSaving}
            className="p-2 text-editor-muted hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
            title="Delete configuration"
          >
            <Trash2 size={16} />
          </button>

          <button
            onClick={handleSave}
            disabled={!hasUnsavedChanges || isSaving || !name.trim()}
            className="flex items-center gap-2 px-4 py-1.5 bg-editor-accent text-white text-sm rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSaving ? (
              <>
                <Loader2 size={14} className="animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save size={14} />
                Save
              </>
            )}
          </button>

          {onClose && (
            <button
              onClick={handleClose}
              className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
            >
              <X size={16} />
            </button>
          )}
        </div>
      </div>

      {/* Description */}
      <div className="px-4 py-2 border-b border-editor-border">
        <input
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="w-full bg-transparent border-none focus:outline-none text-sm text-editor-muted placeholder-editor-muted"
          placeholder="Add a description..."
        />
      </div>

      {/* Tabs */}
      <div className="flex border-b border-editor-border">
        <button
          onClick={() => setActiveTab('commands')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 ${
            activeTab === 'commands'
              ? 'border-editor-accent text-editor-accent'
              : 'border-transparent text-editor-muted hover:text-editor-text'
          }`}
        >
          Commands ({config.commands?.length || 0})
        </button>
        <button
          onClick={() => setActiveTab('environment')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 ${
            activeTab === 'environment'
              ? 'border-editor-accent text-editor-accent'
              : 'border-transparent text-editor-muted hover:text-editor-text'
          }`}
        >
          Environment ({config.envVars?.length || 0})
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {activeTab === 'commands' ? (
          <BuildCommandEditor
            commands={config.commands || []}
            onAdd={(cmdName, cmd, workDir) => onAddCommand(config.id, cmdName, cmd, workDir)}
            onUpdate={(cmdId, data) => onUpdateCommand(config.id, cmdId, data)}
            onDelete={(cmdId) => onDeleteCommand(config.id, cmdId)}
            onReorder={(order) => onReorderCommands(config.id, order)}
            disabled={isSaving}
          />
        ) : (
          <BuildEnvEditor
            envVars={config.envVars || []}
            onAdd={(key, value, isSecret) => onSetEnvVar(config.id, key, value, isSecret)}
            onDelete={(key) => onDeleteEnvVar(config.id, key)}
            disabled={isSaving}
          />
        )}
      </div>

      {/* Delete confirmation */}
      <ConfirmDialog
        isOpen={showDeleteConfirm}
        title="Delete Configuration"
        message={`Are you sure you want to delete "${config.name}"? All commands and environment variables will be permanently removed.`}
        confirmText="Delete"
        variant="danger"
        onConfirm={async () => {
          await onDelete(config.id);
          setShowDeleteConfirm(false);
        }}
        onCancel={() => setShowDeleteConfirm(false)}
      />

      {/* Unsaved changes warning */}
      <ConfirmDialog
        isOpen={showUnsavedWarning}
        title="Unsaved Changes"
        message="You have unsaved changes. Are you sure you want to close without saving?"
        confirmText="Discard"
        cancelText="Keep Editing"
        variant="warning"
        onConfirm={() => {
          setShowUnsavedWarning(false);
          onClose?.();
        }}
        onCancel={() => setShowUnsavedWarning(false)}
      />
    </div>
  );
}
