import { FileCode, FileJson, FileText, File, X } from 'lucide-react';
import { useAppStore } from '../../store';

// Get appropriate icon based on file extension
function getFileIcon(filename: string) {
  const ext = filename.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'js':
    case 'jsx':
    case 'ts':
    case 'tsx':
    case 'py':
    case 'go':
    case 'rs':
    case 'java':
    case 'cpp':
    case 'c':
    case 'h':
      return <FileCode size={12} className="text-editor-accent" />;
    case 'json':
      return <FileJson size={12} className="text-editor-accent" />;
    case 'md':
    case 'txt':
    case 'html':
    case 'css':
      return <FileText size={12} className="text-editor-accent" />;
    default:
      return <File size={12} className="text-editor-accent" />;
  }
}

export function FileContextBadge() {
  const { selectedFile, fileContextEnabled, setFileContextEnabled } = useAppStore();

  // Don't render if no file is selected or context is disabled
  if (!selectedFile || !fileContextEnabled) {
    return null;
  }

  const filename = selectedFile.name;

  return (
    <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-editor-surface border border-editor-border text-xs">
      {getFileIcon(filename)}
      <span className="text-editor-text max-w-[120px] truncate" title={selectedFile.path}>
        {filename}
      </span>
      <button
        onClick={() => setFileContextEnabled(false)}
        className="text-editor-muted hover:text-editor-text transition-colors ml-0.5"
        title="Remove file context"
      >
        <X size={12} />
      </button>
    </div>
  );
}
