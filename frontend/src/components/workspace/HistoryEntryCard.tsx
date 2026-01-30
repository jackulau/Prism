import { FileCode, FileJson, FileText, File as FileIcon, Eye, RotateCcw, Plus, Pencil, Trash2 } from 'lucide-react';
import type { FileHistoryEntry, FileHistoryOperation } from '../../types';

interface HistoryEntryCardProps {
  entry: FileHistoryEntry;
  isSelected: boolean;
  onClick: () => void;
  onRestore: () => void;
  onPreview: () => void;
}

const OPERATION_STYLES: Record<FileHistoryOperation, { bg: string; text: string; icon: typeof Plus }> = {
  create: { bg: 'bg-green-500/20', text: 'text-green-400', icon: Plus },
  update: { bg: 'bg-blue-500/20', text: 'text-blue-400', icon: Pencil },
  delete: { bg: 'bg-red-500/20', text: 'text-red-400', icon: Trash2 },
};

function getFileIcon(path: string) {
  const ext = path.split('.').pop()?.toLowerCase();

  switch (ext) {
    case 'ts':
    case 'tsx':
    case 'js':
    case 'jsx':
    case 'py':
    case 'go':
    case 'rs':
    case 'java':
    case 'c':
    case 'cpp':
    case 'h':
      return <FileCode className="w-4 h-4" />;
    case 'json':
      return <FileJson className="w-4 h-4" />;
    case 'md':
    case 'txt':
      return <FileText className="w-4 h-4" />;
    default:
      return <FileIcon className="w-4 h-4" />;
  }
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) {
    return 'just now';
  } else if (diffMinutes < 60) {
    return `${diffMinutes}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else if (diffDays < 7) {
    return `${diffDays}d ago`;
  } else {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function HistoryEntryCard({
  entry,
  isSelected,
  onClick,
  onRestore,
  onPreview,
}: HistoryEntryCardProps) {
  const operation = entry.operation as FileHistoryOperation;
  const opStyle = OPERATION_STYLES[operation] || OPERATION_STYLES.update;
  const OperationIcon = opStyle.icon;
  const fileName = entry.file_path.split('/').pop() || entry.file_path;
  const dirPath = entry.file_path.split('/').slice(0, -1).join('/');

  return (
    <div
      className={`group relative p-3 border-b border-editor-border/50 hover:bg-editor-border/20 transition-colors cursor-pointer ${
        isSelected ? 'bg-editor-accent/10 border-l-2 border-l-editor-accent' : ''
      }`}
      onClick={onClick}
    >
      {/* Main content */}
      <div className="flex items-start gap-3">
        {/* Operation indicator */}
        <div className={`p-1.5 rounded ${opStyle.bg}`}>
          <OperationIcon size={14} className={opStyle.text} />
        </div>

        {/* Entry details */}
        <div className="flex-1 min-w-0">
          {/* File name and operation badge */}
          <div className="flex items-center gap-2 mb-1">
            <span className="text-editor-muted">{getFileIcon(entry.file_path)}</span>
            <span className="text-sm text-editor-text font-medium truncate" title={entry.file_path}>
              {fileName}
            </span>
            <span className={`text-xs px-1.5 py-0.5 rounded ${opStyle.bg} ${opStyle.text}`}>
              {operation}
            </span>
          </div>

          {/* Path and metadata */}
          <div className="flex items-center gap-2 text-xs text-editor-muted">
            {dirPath && (
              <>
                <span className="truncate max-w-[150px]" title={dirPath}>
                  {dirPath}
                </span>
                <span>•</span>
              </>
            )}
            <span>{formatBytes(entry.size)}</span>
            <span>•</span>
            <span title={new Date(entry.created_at).toLocaleString()}>
              {formatRelativeTime(entry.created_at)}
            </span>
          </div>

          {/* Attribution info */}
          {(entry.agent_name || entry.tool_name) && (
            <div className="mt-1.5 flex items-center gap-2 text-xs text-editor-muted">
              {entry.agent_name && (
                <span className="px-1.5 py-0.5 rounded bg-editor-surface border border-editor-border">
                  {entry.agent_name}
                </span>
              )}
              {entry.tool_name && (
                <span className="text-editor-accent">{entry.tool_name}</span>
              )}
            </div>
          )}

          {/* Description if available */}
          {entry.description && (
            <p className="mt-1.5 text-xs text-editor-muted line-clamp-2">
              {entry.description}
            </p>
          )}
        </div>
      </div>

      {/* Quick action buttons - show on hover */}
      <div className="absolute right-2 top-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onPreview();
          }}
          className="p-1.5 rounded bg-editor-bg border border-editor-border text-editor-muted hover:text-editor-text hover:border-editor-text/30 transition-colors"
          title="Preview this version"
        >
          <Eye size={14} />
        </button>
        {operation !== 'delete' && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onRestore();
            }}
            className="p-1.5 rounded bg-editor-bg border border-editor-border text-editor-muted hover:text-editor-accent hover:border-editor-accent/30 transition-colors"
            title="Restore this version"
          >
            <RotateCcw size={14} />
          </button>
        )}
      </div>
    </div>
  );
}
