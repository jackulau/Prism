import { useState, useEffect } from 'react';
import { DiffEditor } from '@monaco-editor/react';
import { X, RotateCcw, AlertTriangle, Check, FileCode, Clock, Archive } from 'lucide-react';
import { useSandboxStore, type FileHistoryEntry } from '../../store/sandboxStore';
import { wsService } from '../../services/websocket';

interface RestorePreviewDialogProps {
  entry: FileHistoryEntry;
  currentContent: string;
  isOpen: boolean;
  onClose: () => void;
}

function getLanguageFromPath(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase();
  const languageMap: Record<string, string> = {
    js: 'javascript',
    jsx: 'javascript',
    ts: 'typescript',
    tsx: 'typescript',
    html: 'html',
    css: 'css',
    scss: 'scss',
    less: 'less',
    json: 'json',
    md: 'markdown',
    py: 'python',
    go: 'go',
    rs: 'rust',
    yaml: 'yaml',
    yml: 'yaml',
    sh: 'shell',
    bash: 'shell',
    sql: 'sql',
    xml: 'xml',
    java: 'java',
    c: 'c',
    cpp: 'cpp',
    rb: 'ruby',
    php: 'php',
    swift: 'swift',
    kt: 'kotlin',
  };
  return languageMap[ext || ''] || 'plaintext';
}

function calculateChangeSummary(original: string, modified: string) {
  const originalLines = original.split('\n');
  const modifiedLines = modified.split('\n');

  let additions = 0;
  let deletions = 0;

  // Simple line-by-line comparison
  const maxLines = Math.max(originalLines.length, modifiedLines.length);
  for (let i = 0; i < maxLines; i++) {
    const origLine = originalLines[i] || '';
    const modLine = modifiedLines[i] || '';
    if (origLine !== modLine) {
      if (!origLine && modLine) additions++;
      else if (origLine && !modLine) deletions++;
      else {
        additions++;
        deletions++;
      }
    }
  }

  const sizeChange = modified.length - original.length;

  return { additions, deletions, sizeChange };
}

export function RestorePreviewDialog({
  entry,
  currentContent,
  isOpen,
  onClose,
}: RestorePreviewDialogProps) {
  const [historicalContent, setHistoricalContent] = useState<string | null>(null);
  const [createBackup, setCreateBackup] = useState(true);
  const [isLoadingContent, setIsLoadingContent] = useState(false);

  const {
    isRestoring,
    restoreError,
    restoreConflicts,
    confirmRestore,
    cancelRestore,
    historyContent,
    isLoadingHistory,
  } = useSandboxStore();

  // Load historical content when dialog opens
  useEffect(() => {
    if (isOpen && entry) {
      setIsLoadingContent(true);
      wsService.requestHistoryContent(entry.id);
    }
  }, [isOpen, entry]);

  // Update local state when history content is loaded
  useEffect(() => {
    if (historyContent !== null && !isLoadingHistory) {
      setHistoricalContent(historyContent);
      setIsLoadingContent(false);
    }
  }, [historyContent, isLoadingHistory]);

  if (!isOpen) return null;

  const language = getLanguageFromPath(entry.file_path);
  const fileName = entry.file_path.split('/').pop() || entry.file_path;
  const changeSummary = historicalContent
    ? calculateChangeSummary(currentContent, historicalContent)
    : null;

  const handleConfirm = () => {
    confirmRestore(createBackup);
  };

  const handleClose = () => {
    cancelRestore();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-[90vw] h-[85vh] bg-editor-surface border border-editor-border rounded-lg flex flex-col overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-bg">
          <div className="flex items-center gap-3">
            <RotateCcw className="w-5 h-5 text-blue-400" />
            <div>
              <h2 className="text-lg font-semibold text-editor-text">Restore Preview</h2>
              <div className="flex items-center gap-2 text-sm text-editor-muted">
                <FileCode className="w-4 h-4" />
                <span>{fileName}</span>
                <Clock className="w-4 h-4 ml-2" />
                <span>{entry.created_at}</span>
              </div>
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
                <p className="text-sm font-medium text-yellow-500">Newer version detected</p>
                <p className="text-xs text-editor-muted mt-1">
                  The current file has been modified after this version was saved.
                  Restoring will overwrite those changes.
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

        {/* Change Summary */}
        {changeSummary && (
          <div className="px-4 py-2 bg-editor-bg/50 border-b border-editor-border flex items-center gap-6 text-sm">
            <span className="text-green-400">+{changeSummary.additions} additions</span>
            <span className="text-red-400">-{changeSummary.deletions} deletions</span>
            <span className="text-editor-muted">
              Size: {changeSummary.sizeChange >= 0 ? '+' : ''}{changeSummary.sizeChange} bytes
            </span>
          </div>
        )}

        {/* Diff Editor */}
        <div className="flex-1 overflow-hidden">
          {isLoadingContent || !historicalContent ? (
            <div className="flex items-center justify-center h-full text-editor-muted">
              <div className="flex items-center gap-2">
                <div className="w-5 h-5 border-2 border-editor-muted border-t-transparent rounded-full animate-spin" />
                <span>Loading content...</span>
              </div>
            </div>
          ) : (
            <DiffEditor
              height="100%"
              language={language}
              original={currentContent}
              modified={historicalContent}
              theme="vs-dark"
              options={{
                fontSize: 13,
                fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
                readOnly: true,
                renderSideBySide: true,
                automaticLayout: true,
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                lineNumbers: 'on',
                glyphMargin: true,
                scrollbar: {
                  verticalScrollbarSize: 10,
                  horizontalScrollbarSize: 10,
                },
              }}
            />
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-editor-border bg-editor-bg">
          <label className="flex items-center gap-2 text-sm text-editor-text cursor-pointer">
            <input
              type="checkbox"
              checked={createBackup}
              onChange={(e) => setCreateBackup(e.target.checked)}
              className="w-4 h-4 rounded border-editor-border bg-editor-bg text-blue-500 focus:ring-blue-500 focus:ring-offset-0"
            />
            <Archive className="w-4 h-4 text-editor-muted" />
            <span>Create backup before restore</span>
          </label>

          <div className="flex items-center gap-3">
            <button
              onClick={handleClose}
              className="px-4 py-2 rounded-lg text-sm text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleConfirm}
              disabled={isRestoring || !historicalContent}
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm text-white bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isRestoring ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Restoring...</span>
                </>
              ) : (
                <>
                  <Check className="w-4 h-4" />
                  <span>Restore</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
