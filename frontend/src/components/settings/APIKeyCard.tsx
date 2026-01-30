import { useState } from 'react';
import { Key, RefreshCw, Trash2, Edit2, Check, X, AlertTriangle, Loader2 } from 'lucide-react';
import { type APIKey } from '../../store/apiKeysStore';

interface APIKeyCardProps {
  apiKey: APIKey;
  onRotate: (id: string) => void;
  onDelete: (id: string) => void;
  onRename: (id: string, name: string) => void;
  isDeleting: boolean;
  isRotating: boolean;
  isRenaming: boolean;
}

function formatRelativeTime(dateString: string | null): string {
  if (!dateString) return 'Never';
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  if (diffDays < 30) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  return date.toLocaleDateString();
}

function getExpirationStatus(expiresAt: string | null): { text: string; isExpiringSoon: boolean; isExpired: boolean } {
  if (!expiresAt) return { text: 'Never', isExpiringSoon: false, isExpired: false };

  const expireDate = new Date(expiresAt);
  const now = new Date();
  const diffMs = expireDate.getTime() - now.getTime();
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffDays < 0) return { text: 'Expired', isExpiringSoon: false, isExpired: true };
  if (diffDays === 0) return { text: 'Expires today', isExpiringSoon: true, isExpired: false };
  if (diffDays <= 7) return { text: `Expires in ${diffDays} day${diffDays === 1 ? '' : 's'}`, isExpiringSoon: true, isExpired: false };
  if (diffDays <= 30) return { text: `Expires in ${diffDays} days`, isExpiringSoon: false, isExpired: false };
  return { text: expireDate.toLocaleDateString(), isExpiringSoon: false, isExpired: false };
}

export function APIKeyCard({
  apiKey,
  onRotate,
  onDelete,
  onRename,
  isDeleting,
  isRotating,
  isRenaming,
}: APIKeyCardProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState(apiKey.name);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showRotateConfirm, setShowRotateConfirm] = useState(false);

  const expiration = getExpirationStatus(apiKey.expires_at);

  const handleSaveName = () => {
    if (editName.trim() && editName !== apiKey.name) {
      onRename(apiKey.id, editName.trim());
    }
    setIsEditing(false);
  };

  const handleCancelEdit = () => {
    setEditName(apiKey.name);
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveName();
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  return (
    <div className={`p-4 bg-editor-bg rounded-lg border ${expiration.isExpired ? 'border-red-500/50' : expiration.isExpiringSoon ? 'border-yellow-500/50' : 'border-editor-border'}`}>
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <Key className="w-5 h-5 text-editor-accent mt-0.5 flex-shrink-0" />
          <div className="flex-1 min-w-0">
            {/* Name - editable */}
            {isEditing ? (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  onKeyDown={handleKeyDown}
                  autoFocus
                  className="px-2 py-1 text-sm font-medium bg-editor-surface border border-editor-border rounded focus:outline-none focus:border-editor-accent"
                />
                <button
                  onClick={handleSaveName}
                  disabled={isRenaming}
                  className="p-1 text-green-500 hover:bg-green-500/20 rounded"
                >
                  {isRenaming ? <Loader2 className="w-4 h-4 animate-spin" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={handleCancelEdit}
                  className="p-1 text-editor-muted hover:bg-editor-surface rounded"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <span className="font-medium truncate">{apiKey.name}</span>
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-1 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Rename"
                >
                  <Edit2 className="w-3 h-3" />
                </button>
              </div>
            )}

            {/* Key prefix */}
            <div className="text-sm text-editor-muted font-mono mt-1">
              {apiKey.prefix}...
            </div>

            {/* Scopes */}
            {apiKey.scopes.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {apiKey.scopes.map((scope) => (
                  <span
                    key={scope}
                    className="px-2 py-0.5 text-xs rounded-full bg-editor-surface border border-editor-border text-editor-text"
                  >
                    {scope}
                  </span>
                ))}
              </div>
            )}

            {/* Metadata */}
            <div className="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-xs text-editor-muted">
              <span>Created: {new Date(apiKey.created_at).toLocaleDateString()}</span>
              <span>Last used: {formatRelativeTime(apiKey.last_used_at)}</span>
              <span className={expiration.isExpired ? 'text-red-400' : expiration.isExpiringSoon ? 'text-yellow-400' : ''}>
                {expiration.isExpired && <AlertTriangle className="w-3 h-3 inline mr-1" />}
                Expires: {expiration.text}
              </span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 ml-4 flex-shrink-0">
          {showRotateConfirm ? (
            <div className="flex items-center gap-2 bg-yellow-500/10 border border-yellow-500/30 rounded-lg px-3 py-1.5">
              <span className="text-xs text-yellow-400">Rotate key?</span>
              <button
                onClick={() => {
                  onRotate(apiKey.id);
                  setShowRotateConfirm(false);
                }}
                disabled={isRotating}
                className="px-2 py-0.5 text-xs bg-yellow-500 text-white rounded hover:bg-yellow-600 disabled:opacity-50"
              >
                {isRotating ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Yes'}
              </button>
              <button
                onClick={() => setShowRotateConfirm(false)}
                className="px-2 py-0.5 text-xs text-editor-muted hover:text-editor-text"
              >
                No
              </button>
            </div>
          ) : showDeleteConfirm ? (
            <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-1.5">
              <span className="text-xs text-red-400">Delete key?</span>
              <button
                onClick={() => {
                  onDelete(apiKey.id);
                  setShowDeleteConfirm(false);
                }}
                disabled={isDeleting}
                className="px-2 py-0.5 text-xs bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
              >
                {isDeleting ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Yes'}
              </button>
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="px-2 py-0.5 text-xs text-editor-muted hover:text-editor-text"
              >
                No
              </button>
            </div>
          ) : (
            <>
              <button
                onClick={() => setShowRotateConfirm(true)}
                disabled={isRotating || isDeleting}
                className="px-3 py-1.5 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-border transition-colors disabled:opacity-50 flex items-center gap-1.5"
                title="Rotate (generate new key)"
              >
                <RefreshCw className={`w-4 h-4 ${isRotating ? 'animate-spin' : ''}`} />
                Rotate
              </button>
              <button
                onClick={() => setShowDeleteConfirm(true)}
                disabled={isDeleting || isRotating}
                className="px-3 py-1.5 text-sm text-red-400 border border-red-400/20 rounded-lg hover:bg-red-400/10 transition-colors disabled:opacity-50 flex items-center gap-1.5"
                title="Delete key"
              >
                <Trash2 className="w-4 h-4" />
                Delete
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
