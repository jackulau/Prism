import { useState, useEffect } from 'react';
import { Plus, Key, AlertTriangle, Loader2, ChevronDown, ChevronUp } from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useAPIKeysStore } from '../../store/apiKeysStore';
import { APIKeyCard } from './APIKeyCard';
import { ProviderKeyCard } from './ProviderKeyCard';
import { CreateAPIKeyModal } from './CreateAPIKeyModal';
import { NewKeyDisplay } from './NewKeyDisplay';

export function APIKeysSettings() {
  const { accessToken } = useAuthStore();
  const {
    keys,
    providerKeys,
    newKeyValue,
    isLoading,
    isCreating,
    isDeleting,
    isRotating,
    isRenaming,
    error,
    setToken,
    fetchKeys,
    fetchProviderKeys,
    createKey,
    deleteKey,
    rotateKey,
    renameKey,
    clearNewKey,
    clearError,
  } = useAPIKeysStore();

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showProviderKeys, setShowProviderKeys] = useState(true);

  useEffect(() => {
    if (accessToken) {
      setToken(accessToken);
      fetchKeys();
      fetchProviderKeys();
    }
  }, [accessToken, setToken, fetchKeys, fetchProviderKeys]);

  const handleCreate = async (name: string, expiresInDays?: number, scopes?: string[]) => {
    clearError();
    const success = await createKey(name, expiresInDays, scopes);
    if (success) {
      setShowCreateModal(false);
    }
    return success;
  };

  const handleDelete = (id: string) => {
    clearError();
    deleteKey(id);
  };

  const handleRotate = (id: string) => {
    clearError();
    rotateKey(id);
  };

  const handleRename = (id: string, name: string) => {
    clearError();
    renameKey(id, name);
  };

  return (
    <div className="space-y-6">
      {/* Error Display */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
          <AlertTriangle className="w-4 h-4 flex-shrink-0" />
          {error}
          <button
            onClick={clearError}
            className="ml-auto text-xs hover:text-red-300"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Your API Keys Section */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-editor-muted">Your API Keys</h3>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-3 py-1.5 text-sm bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors flex items-center gap-1.5"
          >
            <Plus className="w-4 h-4" />
            Create New Key
          </button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="w-6 h-6 animate-spin text-editor-muted" />
          </div>
        ) : keys.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 px-4 bg-editor-bg rounded-lg border border-editor-border text-center">
            <Key className="w-10 h-10 text-editor-muted mb-3" />
            <p className="text-sm text-editor-muted">No API keys yet</p>
            <p className="text-xs text-editor-muted mt-1">
              Create an API key to access the Prism API programmatically
            </p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="mt-4 px-4 py-2 text-sm bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors flex items-center gap-1.5"
            >
              <Plus className="w-4 h-4" />
              Create Your First Key
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {keys.map((key) => (
              <div key={key.id} className="group">
                <APIKeyCard
                  apiKey={key}
                  onRotate={handleRotate}
                  onDelete={handleDelete}
                  onRename={handleRename}
                  isDeleting={isDeleting === key.id}
                  isRotating={isRotating === key.id}
                  isRenaming={isRenaming === key.id}
                />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Provider Keys Section */}
      {providerKeys.length > 0 && (
        <div>
          <button
            onClick={() => setShowProviderKeys(!showProviderKeys)}
            className="flex items-center justify-between w-full mb-3 group"
          >
            <h3 className="text-sm font-medium text-editor-muted">Provider Keys</h3>
            <span className="text-editor-muted group-hover:text-editor-text transition-colors">
              {showProviderKeys ? (
                <ChevronUp className="w-4 h-4" />
              ) : (
                <ChevronDown className="w-4 h-4" />
              )}
            </span>
          </button>

          {showProviderKeys && (
            <div className="space-y-3">
              {providerKeys.map((pk) => (
                <ProviderKeyCard key={pk.provider} providerKey={pk} />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Create Modal */}
      <CreateAPIKeyModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreate={handleCreate}
        isCreating={isCreating}
      />

      {/* New Key Display Modal */}
      <NewKeyDisplay
        isOpen={!!newKeyValue}
        keyValue={newKeyValue || ''}
        onClose={clearNewKey}
      />
    </div>
  );
}
