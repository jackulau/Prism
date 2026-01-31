import { useState } from 'react';
import { X, RotateCcw, Check, FileCode, Clock, Archive, CheckSquare, Square, AlertTriangle } from 'lucide-react';
import { useSandboxStore, type FileHistoryEntry, type RestoreResult } from '../../store/sandboxStore';

interface BatchRestoreDialogProps {
  entries: FileHistoryEntry[];
  isOpen: boolean;
  onClose: () => void;
}

export function BatchRestoreDialog({
  entries,
  isOpen,
  onClose,
}: BatchRestoreDialogProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(
    new Set(entries.map((e) => e.id))
  );
  const [createBackup, setCreateBackup] = useState(true);

  const {
    isRestoring,
    restoreError,
    restoreConflicts,
    batchRestoreResults,
    confirmBatchRestore,
    setShowBatchRestoreDialog,
    setBatchRestoreEntries,
  } = useSandboxStore();

  if (!isOpen) return null;

  const toggleSelection = (id: string) => {
    const newSelected = new Set(selectedIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedIds(newSelected);
  };

  const toggleAll = () => {
    if (selectedIds.size === entries.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(entries.map((e) => e.id)));
    }
  };

  const handleConfirm = () => {
    const idsToRestore = Array.from(selectedIds);
    if (idsToRestore.length > 0) {
      confirmBatchRestore(idsToRestore, createBackup);
    }
  };

  const handleClose = () => {
    setShowBatchRestoreDialog(false);
    setBatchRestoreEntries([]);
    onClose();
  };

  const getResultForEntry = (id: string): RestoreResult | undefined => {
    return batchRestoreResults.find((r) => r.history_id === id);
  };

  const hasConflict = (filePath: string): boolean => {
    return restoreConflicts.some((c) => c.file_path === filePath);
  };

  const allSelected = selectedIds.size === entries.length;
  const someSelected = selectedIds.size > 0 && selectedIds.size < entries.length;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-[600px] max-h-[80vh] bg-editor-surface border border-editor-border rounded-lg flex flex-col overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-bg">
          <div className="flex items-center gap-3">
            <RotateCcw className="w-5 h-5 text-blue-400" />
            <div>
              <h2 className="text-lg font-semibold text-editor-text">Batch Restore</h2>
              <p className="text-sm text-editor-muted">
                {entries.length} files selected for restore
              </p>
            </div>
          </div>
          <button
            onClick={handleClose}
            className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Conflict Warning */}
        {restoreConflicts.length > 0 && (
          <div className="px-4 py-3 bg-yellow-500/10 border-b border-yellow-500/30">
            <div className="flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-yellow-500">
                  {restoreConflicts.length} file(s) have newer versions
                </p>
                <p className="text-xs text-editor-muted mt-1">
                  Some files have been modified after their selected versions were saved.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Error Display */}
        {restoreError && (
          <div className="px-4 py-3 bg-red-500/10 border-b border-red-500/30">
            <div className="flex items-start gap-3">
              <X className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-red-500">Restore failed</p>
                <p className="text-xs text-editor-muted mt-1">{restoreError}</p>
              </div>
            </div>
          </div>
        )}

        {/* Select All */}
        <div className="px-4 py-2 border-b border-editor-border bg-editor-bg/50">
          <button
            onClick={toggleAll}
            className="flex items-center gap-2 text-sm text-editor-text hover:text-blue-400 transition-colors"
          >
            {allSelected ? (
              <CheckSquare className="w-4 h-4 text-blue-400" />
            ) : someSelected ? (
              <Square className="w-4 h-4 text-blue-400" style={{ opacity: 0.5 }} />
            ) : (
              <Square className="w-4 h-4" />
            )}
            <span>{allSelected ? 'Deselect All' : 'Select All'}</span>
            <span className="text-editor-muted">
              ({selectedIds.size} of {entries.length})
            </span>
          </button>
        </div>

        {/* File List */}
        <div className="flex-1 overflow-y-auto">
          {entries.map((entry) => {
            const isSelected = selectedIds.has(entry.id);
            const result = getResultForEntry(entry.id);
            const conflict = hasConflict(entry.file_path);

            return (
              <div
                key={entry.id}
                className={`flex items-center gap-3 px-4 py-3 border-b border-editor-border/50 hover:bg-editor-bg/50 transition-colors ${
                  result ? (result.success ? 'bg-green-500/5' : 'bg-red-500/5') : ''
                }`}
              >
                <button
                  onClick={() => toggleSelection(entry.id)}
                  disabled={isRestoring || !!result}
                  className="flex-shrink-0 text-editor-text hover:text-blue-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isSelected ? (
                    <CheckSquare className="w-5 h-5 text-blue-400" />
                  ) : (
                    <Square className="w-5 h-5" />
                  )}
                </button>

                <FileCode className="w-4 h-4 text-editor-muted flex-shrink-0" />

                <div className="flex-1 min-w-0">
                  <p className="text-sm text-editor-text truncate">
                    {entry.file_path}
                  </p>
                  <div className="flex items-center gap-3 text-xs text-editor-muted mt-0.5">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {entry.created_at}
                    </span>
                    <span>{entry.operation}</span>
                    <span>{(entry.size / 1024).toFixed(1)} KB</span>
                  </div>
                </div>

                {/* Status indicators */}
                <div className="flex-shrink-0" title={result?.error || (conflict ? "Newer version exists" : undefined)}>
                  {result ? (
                    result.success ? (
                      <Check className="w-5 h-5 text-green-500" />
                    ) : (
                      <X className="w-5 h-5 text-red-500" />
                    )
                  ) : conflict ? (
                    <AlertTriangle className="w-5 h-5 text-yellow-500" />
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-editor-border bg-editor-bg">
          <label className="flex items-center gap-2 text-sm text-editor-text cursor-pointer">
            <input
              type="checkbox"
              checked={createBackup}
              onChange={(e) => setCreateBackup(e.target.checked)}
              disabled={isRestoring}
              className="w-4 h-4 rounded border-editor-border bg-editor-bg text-blue-500 focus:ring-blue-500 focus:ring-offset-0"
            />
            <Archive className="w-4 h-4 text-editor-muted" />
            <span>Create backups</span>
          </label>

          <div className="flex items-center gap-3">
            <button
              onClick={handleClose}
              className="px-4 py-2 rounded-lg text-sm text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
            >
              {batchRestoreResults.length > 0 ? 'Close' : 'Cancel'}
            </button>
            {batchRestoreResults.length === 0 && (
              <button
                onClick={handleConfirm}
                disabled={isRestoring || selectedIds.size === 0}
                className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm text-white bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isRestoring ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    <span>Restoring...</span>
                  </>
                ) : (
                  <>
                    <RotateCcw className="w-4 h-4" />
                    <span>Restore {selectedIds.size} file{selectedIds.size !== 1 ? 's' : ''}</span>
                  </>
                )}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
