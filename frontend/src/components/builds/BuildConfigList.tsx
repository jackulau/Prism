import { Plus, Star, MoreVertical, Trash2, Copy, Settings2 } from 'lucide-react';
import { useState } from 'react';
import type { BuildConfig } from '../../services/buildConfig';
import { ConfirmDialog } from '../ConfirmDialog';

interface BuildConfigListProps {
  configs: BuildConfig[];
  selectedId: string | null;
  onSelect: (config: BuildConfig) => void;
  onCreate: () => void;
  onDelete: (id: string) => Promise<void>;
  onSetDefault: (id: string) => Promise<void>;
  onDuplicate?: (config: BuildConfig) => void;
  isLoading?: boolean;
}

export function BuildConfigList({
  configs,
  selectedId,
  onSelect,
  onCreate,
  onDelete,
  onSetDefault,
  onDuplicate,
  isLoading = false,
}: BuildConfigListProps) {
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [menuOpenId, setMenuOpenId] = useState<string | null>(null);

  const sortedConfigs = [...configs].sort((a, b) => {
    // Default config first
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    // Then by name
    return a.name.localeCompare(b.name);
  });

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="h-16 bg-editor-surface/50 animate-pulse rounded-lg"
          />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {/* Create new button */}
      <button
        onClick={onCreate}
        className="w-full flex items-center gap-2 px-4 py-3 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface border border-editor-border border-dashed rounded-lg transition-colors"
      >
        <Plus size={16} />
        Create New Configuration
      </button>

      {/* Config list */}
      {sortedConfigs.length > 0 ? (
        sortedConfigs.map((config) => (
          <div
            key={config.id}
            className={`relative group rounded-lg border transition-colors cursor-pointer ${
              selectedId === config.id
                ? 'border-editor-accent bg-editor-accent/10'
                : 'border-editor-border hover:border-editor-muted bg-editor-surface/50 hover:bg-editor-surface'
            }`}
          >
            <button
              onClick={() => onSelect(config)}
              className="w-full text-left px-4 py-3"
            >
              <div className="flex items-center gap-2">
                <Settings2 size={14} className="text-editor-muted flex-shrink-0" />
                <span className="font-medium text-editor-text truncate">
                  {config.name}
                </span>
                {config.isDefault && (
                  <span className="flex items-center gap-1 px-1.5 py-0.5 text-xs bg-editor-warning/20 text-editor-warning rounded">
                    <Star size={10} />
                    Default
                  </span>
                )}
              </div>
              {config.description && (
                <p className="mt-1 text-xs text-editor-muted truncate">
                  {config.description}
                </p>
              )}
              <div className="mt-2 flex items-center gap-3 text-xs text-editor-muted">
                <span>{config.commands?.length || 0} commands</span>
                <span>{config.envVars?.length || 0} env vars</span>
              </div>
            </button>

            {/* Actions menu */}
            <div className="absolute top-2 right-2">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setMenuOpenId(menuOpenId === config.id ? null : config.id);
                }}
                className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <MoreVertical size={14} />
              </button>

              {menuOpenId === config.id && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setMenuOpenId(null)}
                  />
                  <div className="absolute right-0 top-full mt-1 w-40 bg-editor-bg border border-editor-border rounded-lg shadow-lg z-20 py-1">
                    {!config.isDefault && (
                      <button
                        onClick={async (e) => {
                          e.stopPropagation();
                          await onSetDefault(config.id);
                          setMenuOpenId(null);
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-editor-text hover:bg-editor-surface transition-colors"
                      >
                        <Star size={14} />
                        Set as Default
                      </button>
                    )}
                    {onDuplicate && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onDuplicate(config);
                          setMenuOpenId(null);
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-editor-text hover:bg-editor-surface transition-colors"
                      >
                        <Copy size={14} />
                        Duplicate
                      </button>
                    )}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setDeleteId(config.id);
                        setMenuOpenId(null);
                      }}
                      className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-400 hover:bg-red-500/10 transition-colors"
                    >
                      <Trash2 size={14} />
                      Delete
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        ))
      ) : (
        <div className="text-center py-8 text-editor-muted text-sm">
          No build configurations yet.
        </div>
      )}

      {/* Delete confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        title="Delete Configuration"
        message="Are you sure you want to delete this build configuration? All commands and environment variables will be permanently removed."
        confirmText="Delete"
        variant="danger"
        onConfirm={async () => {
          if (deleteId) {
            await onDelete(deleteId);
            setDeleteId(null);
          }
        }}
        onCancel={() => setDeleteId(null)}
      />
    </div>
  );
}
