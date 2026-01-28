import { useState } from 'react';
import { DiffEditor } from '@monaco-editor/react';
import { Check, X, ChevronDown, ChevronUp, FileCode } from 'lucide-react';
import type { EditSuggestion } from '../../types/workspace';

interface CodeDiffProps {
  edit: EditSuggestion;
  onAccept: () => void;
  onReject: () => void;
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

export function CodeDiff({ edit, onAccept, onReject }: CodeDiffProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const language = getLanguageFromPath(edit.path);
  const fileName = edit.path.split('/').pop() || edit.path;

  // Calculate line count for height estimation
  const originalLines = edit.original.split('\n').length;
  const modifiedLines = edit.modified.split('\n').length;
  const maxLines = Math.max(originalLines, modifiedLines);
  const editorHeight = Math.min(Math.max(maxLines * 20 + 40, 120), 400);

  return (
    <div className="rounded-lg border border-editor-border overflow-hidden bg-editor-surface/50 my-2">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-editor-surface border-b border-editor-border">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1 rounded hover:bg-editor-border/50 text-editor-muted transition-colors"
          >
            {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          </button>
          <FileCode size={14} className="text-blue-400" />
          <span className="text-sm font-medium text-editor-text">{fileName}</span>
          <span className="text-xs text-editor-muted">
            Lines {edit.startLine}-{edit.endLine}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onReject}
            className="flex items-center gap-1 px-2 py-1 rounded text-sm text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
            title="Reject changes"
          >
            <X size={14} />
            <span>Reject</span>
          </button>
          <button
            onClick={onAccept}
            className="flex items-center gap-1 px-2 py-1 rounded text-sm text-white bg-editor-success hover:bg-editor-success/80 transition-colors"
            title="Accept changes"
          >
            <Check size={14} />
            <span>Accept</span>
          </button>
        </div>
      </div>

      {/* Description */}
      {edit.description && (
        <div className="px-3 py-2 bg-editor-bg/50 border-b border-editor-border">
          <p className="text-xs text-editor-muted">{edit.description}</p>
        </div>
      )}

      {/* Diff Editor */}
      {isExpanded && (
        <div style={{ height: editorHeight }}>
          <DiffEditor
            height="100%"
            language={language}
            original={edit.original}
            modified={edit.modified}
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
              glyphMargin: false,
              folding: false,
              lineDecorationsWidth: 0,
              lineNumbersMinChars: 3,
              scrollbar: {
                verticalScrollbarSize: 8,
                horizontalScrollbarSize: 8,
              },
            }}
          />
        </div>
      )}
    </div>
  );
}

interface InlineCodeDiffProps {
  original: string;
  modified: string;
  language: string;
  onAccept: () => void;
  onReject: () => void;
}

export function InlineCodeDiff({ original, modified, language, onAccept, onReject }: InlineCodeDiffProps) {
  const originalLines = original.split('\n').length;
  const modifiedLines = modified.split('\n').length;
  const maxLines = Math.max(originalLines, modifiedLines);
  const editorHeight = Math.min(Math.max(maxLines * 20 + 20, 100), 300);

  return (
    <div className="rounded-lg border border-editor-border overflow-hidden my-2">
      <div className="flex items-center justify-between px-3 py-2 bg-editor-surface/50 border-b border-editor-border">
        <span className="text-sm font-medium text-editor-text">Suggested Changes</span>
        <div className="flex items-center gap-2">
          <button
            onClick={onReject}
            className="p-1.5 rounded text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
            title="Reject"
          >
            <X size={16} />
          </button>
          <button
            onClick={onAccept}
            className="p-1.5 rounded text-white bg-editor-success hover:bg-editor-success/80 transition-colors"
            title="Accept"
          >
            <Check size={16} />
          </button>
        </div>
      </div>
      <div style={{ height: editorHeight }}>
        <DiffEditor
          height="100%"
          language={language}
          original={original}
          modified={modified}
          theme="vs-dark"
          options={{
            fontSize: 13,
            fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
            readOnly: true,
            renderSideBySide: false,
            automaticLayout: true,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            lineNumbers: 'off',
            glyphMargin: false,
            folding: false,
            scrollbar: {
              verticalScrollbarSize: 6,
              horizontalScrollbarSize: 6,
            },
          }}
        />
      </div>
    </div>
  );
}
