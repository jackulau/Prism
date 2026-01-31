import { useState, useEffect } from 'react';
import { Settings2, Plus, X, Loader2 } from 'lucide-react';
import { BuildConfigList } from './BuildConfigList';
import { BuildConfigEditor } from './BuildConfigEditor';
import { useBuildConfigStore } from '../../store/buildConfigStore';
import { useAuthStore } from '../../store/authStore';
import type { BuildConfig, CreateBuildConfigInput } from '../../services/buildConfig';

interface BuildConfigPanelProps {
  workspaceId?: string;
  onClose?: () => void;
}

export function BuildConfigPanel({ workspaceId, onClose }: BuildConfigPanelProps) {
  const { accessToken } = useAuthStore();
  const {
    configs,
    selectedConfig,
    isLoading,
    isSaving,
    error,
    fetchConfigs,
    createConfig,
    updateConfig,
    deleteConfig,
    setDefault,
    selectConfig,
    addCommand,
    updateCommand,
    deleteCommand,
    reorderCommands,
    setEnvVar,
    deleteEnvVar,
    setToken,
    clearError,
  } = useBuildConfigStore();

  const [isCreating, setIsCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');

  // Initialize token and fetch configs
  useEffect(() => {
    if (accessToken) {
      setToken(accessToken);
      fetchConfigs(workspaceId);
    }
  }, [accessToken, workspaceId, setToken, fetchConfigs]);

  const handleCreate = async () => {
    if (!newName.trim()) return;

    const input: CreateBuildConfigInput = {
      name: newName.trim(),
      description: newDescription.trim() || undefined,
      workspaceId,
    };

    const created = await createConfig(input);
    if (created) {
      setIsCreating(false);
      setNewName('');
      setNewDescription('');
    }
  };

  const handleDuplicate = async (config: BuildConfig) => {
    const input: CreateBuildConfigInput = {
      name: `${config.name} (Copy)`,
      description: config.description,
      workspaceId: config.workspaceId,
      orgWorkspaceId: config.orgWorkspaceId,
    };

    const created = await createConfig(input);
    if (created) {
      // Copy commands and env vars
      for (const cmd of config.commands || []) {
        await addCommand(created.id, {
          name: cmd.name,
          command: cmd.command,
          workingDirectory: cmd.workingDirectory,
          runOrder: cmd.runOrder,
          isEnabled: cmd.isEnabled,
        });
      }
      // Note: Can't copy env var values since secrets are masked
      // User will need to re-enter secret values
    }
  };

  return (
    <div className="flex flex-col h-full bg-editor-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border">
        <div className="flex items-center gap-2">
          <Settings2 size={20} className="text-editor-accent" />
          <h2 className="text-lg font-semibold text-editor-text">Build Configurations</h2>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={16} />
          </button>
        )}
      </div>

      {/* Error display */}
      {error && (
        <div className="mx-4 mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-sm text-red-400 flex items-center justify-between">
          <span>{error}</span>
          <button onClick={clearError} className="p-1 hover:bg-red-500/20 rounded">
            <X size={14} />
          </button>
        </div>
      )}

      {/* Main content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left sidebar - config list */}
        <div className="w-72 border-r border-editor-border flex flex-col">
          <div className="p-3 flex-1 overflow-y-auto">
            {isCreating ? (
              <div className="p-4 bg-editor-surface rounded-lg space-y-3 border border-editor-accent/30">
                <h3 className="text-sm font-medium text-editor-text">New Configuration</h3>
                <input
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="Configuration name"
                  autoFocus
                  className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                />
                <input
                  type="text"
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  placeholder="Description (optional)"
                  className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                />
                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => {
                      setIsCreating(false);
                      setNewName('');
                      setNewDescription('');
                    }}
                    className="px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleCreate}
                    disabled={!newName.trim() || isSaving}
                    className="flex items-center gap-1 px-4 py-1.5 bg-editor-accent text-white text-sm rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {isSaving ? (
                      <Loader2 size={14} className="animate-spin" />
                    ) : (
                      <Plus size={14} />
                    )}
                    Create
                  </button>
                </div>
              </div>
            ) : (
              <BuildConfigList
                configs={configs}
                selectedId={selectedConfig?.id || null}
                onSelect={selectConfig}
                onCreate={() => setIsCreating(true)}
                onDelete={deleteConfig}
                onSetDefault={setDefault}
                onDuplicate={handleDuplicate}
                isLoading={isLoading}
              />
            )}
          </div>
        </div>

        {/* Right side - editor */}
        <div className="flex-1 overflow-hidden">
          {selectedConfig ? (
            <BuildConfigEditor
              config={selectedConfig}
              onUpdate={updateConfig}
              onDelete={deleteConfig}
              onSetDefault={setDefault}
              onAddCommand={async (configId, name, command, workDir) => {
                await addCommand(configId, {
                  name,
                  command,
                  workingDirectory: workDir,
                });
              }}
              onUpdateCommand={updateCommand}
              onDeleteCommand={deleteCommand}
              onReorderCommands={reorderCommands}
              onSetEnvVar={async (configId, key, value, isSecret) => {
                await setEnvVar(configId, { key, value, isSecret });
              }}
              onDeleteEnvVar={deleteEnvVar}
              onClose={() => selectConfig(null)}
              isSaving={isSaving}
            />
          ) : (
            <div className="flex items-center justify-center h-full text-editor-muted">
              <div className="text-center">
                <Settings2 size={48} className="mx-auto mb-4 opacity-50" />
                <p className="text-lg">Select a configuration to edit</p>
                <p className="text-sm mt-1">or create a new one from the sidebar</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
