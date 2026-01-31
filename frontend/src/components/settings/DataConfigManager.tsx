import { useState, useEffect } from 'react';
import {
  Database,
  Plus,
  Trash2,
  Edit2,
  RefreshCw,
  ChevronRight,
  AlertTriangle,
  Lock,
  Search,
  FolderOpen,
  Key,
} from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useDataConfigStore } from '../../store/dataConfigStore';
import { DataConfigModal } from './DataConfigModal';
import { ConfirmDialog } from '../ConfirmDialog';

export function DataConfigManager() {
  const { accessToken } = useAuthStore();
  const {
    configTypes,
    selectedType,
    configKeys,
    selectedKey,
    configData,
    configUpdatedAt,
    typesLoading,
    keysLoading,
    dataLoading,
    saving,
    deleting,
    error,
    fetchConfigTypes,
    selectType,
    selectKey,
    setConfig,
    deleteConfig,
    clearError,
  } = useDataConfigStore();

  const [searchTerm, setSearchTerm] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  // Fetch config types on mount
  useEffect(() => {
    if (accessToken) {
      fetchConfigTypes();
    }
  }, [accessToken, fetchConfigTypes]);

  // Filter keys by search term
  const filteredKeys = configKeys.filter((key) =>
    key.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleCreateNew = () => {
    setModalMode('create');
    setModalOpen(true);
  };

  const handleEdit = () => {
    if (selectedType && selectedKey && configData) {
      setModalMode('edit');
      setModalOpen(true);
    }
  };

  const handleSave = async (
    configType: string,
    configKey: string,
    value: Record<string, unknown>
  ) => {
    const success = await setConfig(configType, configKey, value);
    if (success && modalMode === 'create') {
      // Select the newly created config
      selectType(configType);
      setTimeout(() => selectKey(configKey), 100);
    }
    return success;
  };

  const handleDelete = async () => {
    if (selectedType && selectedKey) {
      await deleteConfig(selectedType, selectedKey);
      setDeleteDialogOpen(false);
    }
  };

  const handleRefresh = () => {
    fetchConfigTypes();
    if (selectedType) {
      selectType(selectedType);
    }
  };

  return (
    <div className="space-y-4">
      {/* Header with description */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <Lock className="w-4 h-4 text-editor-muted" />
          <p className="text-sm text-editor-muted">
            Securely store encrypted configuration data (API keys, credentials, settings)
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            disabled={typesLoading}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded transition-colors"
            title="Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${typesLoading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={handleCreateNew}
            className="flex items-center gap-2 px-3 py-1.5 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            New Config
          </button>
        </div>
      </div>

      {/* Error display */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
          <AlertTriangle className="w-4 h-4 flex-shrink-0" />
          <span>{error}</span>
          <button
            onClick={clearError}
            className="ml-auto text-xs underline hover:no-underline"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Main content area */}
      <div className="flex gap-4 min-h-[300px]">
        {/* Left panel - Types and Keys */}
        <div className="w-64 flex-shrink-0 bg-editor-bg border border-editor-border rounded-lg overflow-hidden">
          {/* Types section */}
          <div className="border-b border-editor-border">
            <div className="px-3 py-2 text-xs font-medium text-editor-muted uppercase tracking-wider flex items-center gap-2">
              <FolderOpen className="w-3 h-3" />
              Config Types
            </div>
            <div className="max-h-32 overflow-y-auto">
              {typesLoading ? (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  Loading...
                </div>
              ) : configTypes.length === 0 ? (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  No configurations yet
                </div>
              ) : (
                configTypes.map((type) => (
                  <button
                    key={type}
                    onClick={() => selectType(type)}
                    className={`w-full px-3 py-2 text-left text-sm flex items-center gap-2 hover:bg-editor-surface transition-colors ${
                      selectedType === type
                        ? 'bg-editor-surface text-editor-accent'
                        : 'text-editor-text'
                    }`}
                  >
                    <Database className="w-4 h-4 flex-shrink-0" />
                    <span className="truncate">{type}</span>
                    {selectedType === type && (
                      <ChevronRight className="w-4 h-4 ml-auto" />
                    )}
                  </button>
                ))
              )}
            </div>
          </div>

          {/* Keys section */}
          <div>
            <div className="px-3 py-2 text-xs font-medium text-editor-muted uppercase tracking-wider flex items-center gap-2">
              <Key className="w-3 h-3" />
              Keys
              {selectedType && (
                <span className="ml-auto text-xs font-normal">
                  ({filteredKeys.length})
                </span>
              )}
            </div>

            {selectedType && (
              <div className="px-2 pb-2">
                <div className="relative">
                  <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-editor-muted" />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Filter keys..."
                    className="w-full pl-7 pr-2 py-1.5 text-xs bg-editor-surface border border-editor-border rounded"
                  />
                </div>
              </div>
            )}

            <div className="max-h-40 overflow-y-auto">
              {!selectedType ? (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  Select a type
                </div>
              ) : keysLoading ? (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  Loading...
                </div>
              ) : filteredKeys.length === 0 ? (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  {searchTerm ? 'No matches' : 'No keys'}
                </div>
              ) : (
                filteredKeys.map((key) => (
                  <button
                    key={key}
                    onClick={() => selectKey(key)}
                    className={`w-full px-3 py-2 text-left text-sm truncate hover:bg-editor-surface transition-colors ${
                      selectedKey === key
                        ? 'bg-editor-surface text-editor-accent'
                        : 'text-editor-text'
                    }`}
                  >
                    {key}
                  </button>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Right panel - Config data */}
        <div className="flex-1 bg-editor-bg border border-editor-border rounded-lg overflow-hidden flex flex-col">
          {/* Toolbar */}
          <div className="flex items-center justify-between px-3 py-2 border-b border-editor-border">
            <div className="text-sm">
              {selectedType && selectedKey ? (
                <span className="font-medium">
                  {selectedType} / {selectedKey}
                </span>
              ) : (
                <span className="text-editor-muted">
                  Select a configuration to view
                </span>
              )}
            </div>
            {selectedType && selectedKey && configData && (
              <div className="flex gap-1">
                <button
                  onClick={handleEdit}
                  className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
                  title="Edit"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setDeleteDialogOpen(true)}
                  disabled={deleting}
                  className="p-1.5 text-editor-muted hover:text-red-400 hover:bg-editor-surface rounded transition-colors"
                  title="Delete"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>

          {/* Content area */}
          <div className="flex-1 overflow-auto p-3">
            {dataLoading ? (
              <div className="flex items-center justify-center h-full text-editor-muted">
                <RefreshCw className="w-5 h-5 animate-spin" />
              </div>
            ) : configData ? (
              <div className="space-y-3">
                <pre className="text-sm font-mono bg-editor-surface p-3 rounded-lg overflow-auto whitespace-pre-wrap">
                  {JSON.stringify(configData, null, 2)}
                </pre>
                {configUpdatedAt && (
                  <p className="text-xs text-editor-muted">
                    Last updated: {new Date(configUpdatedAt).toLocaleString()}
                  </p>
                )}
              </div>
            ) : selectedType && selectedKey ? (
              <div className="flex items-center justify-center h-full text-editor-muted text-sm">
                Configuration not found
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-editor-muted text-sm gap-2">
                <Database className="w-8 h-8 opacity-50" />
                <p>Select a configuration to view its contents</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create/Edit Modal */}
      <DataConfigModal
        isOpen={modalOpen}
        mode={modalMode}
        initialType={modalMode === 'edit' ? selectedType || '' : ''}
        initialKey={modalMode === 'edit' ? selectedKey || '' : ''}
        initialValue={modalMode === 'edit' ? (configData as Record<string, unknown>) || {} : {}}
        onSave={handleSave}
        onClose={() => setModalOpen(false)}
        saving={saving}
      />

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        title="Delete Configuration"
        message={`Are you sure you want to delete "${selectedType}/${selectedKey}"? This action cannot be undone.`}
        variant="danger"
        onConfirm={handleDelete}
        onCancel={() => setDeleteDialogOpen(false)}
      />
    </div>
  );
}
