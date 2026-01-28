import { useState, useEffect, useRef } from 'react';
import Editor from '@monaco-editor/react';
import { File, Copy, Check, X, Maximize2, Minimize2 } from 'lucide-react';
import type { FileContext, Range } from '../../types/workspace';
import type { editor } from 'monaco-editor';

interface FilePreviewProps {
  file: FileContext | null;
  highlights?: Range[];
  onClose?: () => void;
  onEdit?: (content: string) => void;
  readOnly?: boolean;
  showHeader?: boolean;
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
    svg: 'xml',
    java: 'java',
    c: 'c',
    cpp: 'cpp',
    h: 'c',
    hpp: 'cpp',
    rb: 'ruby',
    php: 'php',
    swift: 'swift',
    kt: 'kotlin',
    dart: 'dart',
    vue: 'vue',
    svelte: 'svelte',
  };
  return languageMap[ext || ''] || 'plaintext';
}

function getFileIcon(path: string) {
  const ext = path.split('.').pop()?.toLowerCase();
  const iconColors: Record<string, string> = {
    ts: 'text-blue-400',
    tsx: 'text-blue-400',
    js: 'text-yellow-400',
    jsx: 'text-yellow-400',
    py: 'text-green-400',
    go: 'text-cyan-400',
    rs: 'text-orange-400',
    json: 'text-yellow-500',
    md: 'text-gray-400',
  };
  return iconColors[ext || ''] || 'text-editor-muted';
}

export function FilePreview({
  file,
  highlights,
  onClose,
  onEdit,
  readOnly = true,
  showHeader = true,
}: FilePreviewProps) {
  const [copied, setCopied] = useState(false);
  const [isMaximized, setIsMaximized] = useState(false);
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const decorationsRef = useRef<string[]>([]);

  // Apply highlights when they change
  useEffect(() => {
    if (!editorRef.current || !highlights?.length) {
      // Clear decorations if no highlights
      if (editorRef.current && decorationsRef.current.length > 0) {
        decorationsRef.current = editorRef.current.deltaDecorations(decorationsRef.current, []);
      }
      return;
    }

    const decorations = highlights.map((range) => ({
      range: {
        startLineNumber: range.startLine,
        startColumn: range.startColumn,
        endLineNumber: range.endLine,
        endColumn: range.endColumn,
      },
      options: {
        isWholeLine: range.startColumn === 0 && range.endColumn === 0,
        className: 'highlighted-line',
        glyphMarginClassName: 'highlighted-glyph',
        inlineClassName: 'highlighted-inline',
      },
    }));

    decorationsRef.current = editorRef.current.deltaDecorations(
      decorationsRef.current,
      decorations
    );
  }, [highlights]);

  // Scroll to first highlight
  useEffect(() => {
    if (editorRef.current && highlights?.length) {
      const firstHighlight = highlights[0];
      editorRef.current.revealLineInCenter(firstHighlight.startLine);
    }
  }, [highlights, file?.path]);

  const handleCopy = async () => {
    if (file?.content) {
      await navigator.clipboard.writeText(file.content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleEditorMount = (editor: editor.IStandaloneCodeEditor) => {
    editorRef.current = editor;
  };

  const handleEditorChange = (value: string | undefined) => {
    if (value !== undefined && onEdit) {
      onEdit(value);
    }
  };

  if (!file) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-muted bg-editor-bg">
        <File size={48} className="mb-4 opacity-50" />
        <p className="text-lg font-medium">No file selected</p>
        <p className="text-sm mt-1">Select a file from the tree or click a file reference</p>
      </div>
    );
  }

  const language = getLanguageFromPath(file.path);
  const fileName = file.path.split('/').pop() || file.path;
  const iconColor = getFileIcon(file.path);

  return (
    <div className={`h-full flex flex-col bg-editor-bg ${isMaximized ? 'fixed inset-0 z-50' : ''}`}>
      {showHeader && (
        <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-surface/30">
          <div className="flex items-center gap-2 min-w-0">
            <File size={14} className={iconColor} />
            <span className="text-sm text-editor-text truncate" title={file.path}>
              {fileName}
            </span>
            {file.selection && (
              <span className="text-xs text-editor-muted">
                Lines {file.selection.startLine}-{file.selection.endLine}
              </span>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={handleCopy}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
              title="Copy content"
            >
              {copied ? <Check size={14} className="text-editor-success" /> : <Copy size={14} />}
            </button>
            <button
              onClick={() => setIsMaximized(!isMaximized)}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
              title={isMaximized ? 'Restore' : 'Maximize'}
            >
              {isMaximized ? <Minimize2 size={14} /> : <Maximize2 size={14} />}
            </button>
            {onClose && (
              <button
                onClick={onClose}
                className="p-1.5 rounded text-editor-muted hover:text-editor-error hover:bg-editor-error/10 transition-colors"
                title="Close preview"
              >
                <X size={14} />
              </button>
            )}
          </div>
        </div>
      )}

      <div className="flex-1 overflow-hidden">
        <Editor
          height="100%"
          language={language}
          value={file.content}
          onChange={handleEditorChange}
          onMount={handleEditorMount}
          theme="vs-dark"
          options={{
            fontSize: 13,
            fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
            minimap: { enabled: true, scale: 0.75 },
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            lineNumbers: 'on',
            renderLineHighlight: 'line',
            tabSize: 2,
            insertSpaces: true,
            automaticLayout: true,
            readOnly,
            padding: { top: 12, bottom: 12 },
            scrollbar: {
              verticalScrollbarSize: 10,
              horizontalScrollbarSize: 10,
            },
            glyphMargin: highlights && highlights.length > 0,
          }}
        />
      </div>

      {/* Custom styles for highlights */}
      <style>{`
        .highlighted-line {
          background-color: rgba(255, 213, 0, 0.1) !important;
        }
        .highlighted-inline {
          background-color: rgba(255, 213, 0, 0.2) !important;
          border-bottom: 2px solid rgba(255, 213, 0, 0.5);
        }
        .highlighted-glyph {
          background-color: rgba(255, 213, 0, 0.4);
          margin-left: 3px;
          width: 4px !important;
          border-radius: 2px;
        }
      `}</style>
    </div>
  );
}
