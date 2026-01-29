import { X, Plus, FileCode, FileJson, FileText, File as FileIcon } from 'lucide-react';
import type { FileContext } from '../../types/workspace';

interface ContextBarProps {
  files: FileContext[];
  onRemove: (path: string) => void;
  onAdd?: () => void;
  onFileClick?: (file: FileContext) => void;
}

interface ContextChipProps {
  file: FileContext;
  onRemove: () => void;
  onClick?: () => void;
}

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
      return <FileCode className="w-3.5 h-3.5" />;
    case 'json':
      return <FileJson className="w-3.5 h-3.5" />;
    case 'md':
    case 'txt':
      return <FileText className="w-3.5 h-3.5" />;
    default:
      return <FileIcon className="w-3.5 h-3.5" />;
  }
}

function getLanguageColor(language: string): string {
  const colors: Record<string, string> = {
    typescript: 'text-blue-400',
    javascript: 'text-yellow-400',
    python: 'text-green-400',
    go: 'text-cyan-400',
    rust: 'text-orange-400',
    java: 'text-red-400',
    json: 'text-yellow-500',
    markdown: 'text-gray-400',
    default: 'text-editor-muted',
  };
  return colors[language] || colors.default;
}

function ContextChip({ file, onRemove, onClick }: ContextChipProps) {
  const fileName = file.path.split('/').pop() || file.path;

  return (
    <div
      className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-editor-surface border border-editor-border hover:border-editor-accent/50 transition-colors group cursor-pointer"
      onClick={onClick}
      title={file.path}
    >
      <span className={getLanguageColor(file.language)}>
        {getFileIcon(file.path)}
      </span>
      <span className="text-xs text-editor-text max-w-[120px] truncate">
        {fileName}
      </span>
      {file.selection && (
        <span className="text-xs text-editor-muted">
          :{file.selection.startLine}-{file.selection.endLine}
        </span>
      )}
      <button
        onClick={(e) => {
          e.stopPropagation();
          onRemove();
        }}
        className="p-0.5 rounded hover:bg-editor-error/20 text-editor-muted hover:text-editor-error transition-colors opacity-0 group-hover:opacity-100"
        title="Remove from context"
      >
        <X size={12} />
      </button>
    </div>
  );
}

function AddContextButton({ onClick }: { onClick?: () => void }) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-1 px-2 py-1 rounded-md border border-dashed border-editor-border hover:border-editor-accent/50 text-editor-muted hover:text-editor-text transition-colors"
      title="Add file to context"
    >
      <Plus size={14} />
      <span className="text-xs">Add</span>
    </button>
  );
}

export function ContextBar({ files, onRemove, onAdd, onFileClick }: ContextBarProps) {
  if (files.length === 0 && !onAdd) {
    return null;
  }

  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b border-editor-border bg-editor-surface/30 flex-wrap">
      <span className="text-xs text-editor-muted font-medium">Context:</span>
      {files.length === 0 ? (
        <span className="text-xs text-editor-muted italic">No files attached</span>
      ) : (
        files.map((file) => (
          <ContextChip
            key={file.path}
            file={file}
            onRemove={() => onRemove(file.path)}
            onClick={() => onFileClick?.(file)}
          />
        ))
      )}
      {onAdd && <AddContextButton onClick={onAdd} />}
    </div>
  );
}
