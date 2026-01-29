import { useCallback, useState } from 'react';
import { PanelLeftClose, PanelLeftOpen, PanelRightClose, PanelRightOpen } from 'lucide-react';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '../ui/ResizablePanel';
import { FileTree } from '../FileTree';
import { WorkspaceChat } from './WorkspaceChat';
import { FilePreview } from './FilePreview';
import { useWorkspaceStore } from '../../store/workspaceStore';
import { useSandboxStore } from '../../store/sandboxStore';
import { wsService } from '../../services/websocket';
import type { FileContext, Range } from '../../types/workspace';

interface WorkspaceChatLayoutProps {
  className?: string;
}

export function WorkspaceChatLayout({ className = '' }: WorkspaceChatLayoutProps) {
  const {
    isFileTreeCollapsed,
    isPreviewCollapsed,
    fileTreeWidth,
    previewWidth,
    toggleFileTree,
    togglePreview,
    setFileTreeWidth,
    setPreviewWidth,
    addContextFile,
  } = useWorkspaceStore();

  const { getFileContent: getSandboxFileContent } = useSandboxStore();

  const [previewFile, setPreviewFile] = useState<FileContext | null>(null);
  const [previewHighlights, setPreviewHighlights] = useState<Range[]>([]);

  // Handle file click from chat messages
  const handleFileClick = useCallback(async (path: string, line?: number) => {
    // Get file content from sandbox store or request it
    let content = getSandboxFileContent(path);

    if (!content) {
      try {
        content = await wsService.requestFile(path);
      } catch {
        console.error('Failed to load file:', path);
        return;
      }
    }

    const language = getLanguageFromPath(path);

    const fileContext: FileContext = {
      path,
      content: content || '',
      language,
    };

    setPreviewFile(fileContext);

    // Set highlights if line is specified
    if (line) {
      setPreviewHighlights([{
        startLine: line,
        startColumn: 0,
        endLine: line,
        endColumn: 0,
      }]);
    } else {
      setPreviewHighlights([]);
    }

    // Open preview panel if collapsed
    if (isPreviewCollapsed) {
      togglePreview();
    }
  }, [getSandboxFileContent, isPreviewCollapsed, togglePreview]);

  // Handle adding file to context from file tree
  const handleAddContext = useCallback(() => {
    if (previewFile) {
      addContextFile(previewFile);
    }
  }, [previewFile, addContextFile]);

  // Handle file tree width change
  const handleFileTreeResize = useCallback((size: number) => {
    setFileTreeWidth(size);
  }, [setFileTreeWidth]);

  // Handle preview width change
  const handlePreviewResize = useCallback((size: number) => {
    setPreviewWidth(size);
  }, [setPreviewWidth]);

  // Handle closing preview
  const handleClosePreview = useCallback(() => {
    setPreviewFile(null);
    setPreviewHighlights([]);
  }, []);

  return (
    <div className={`h-full flex flex-col bg-editor-bg ${className}`}>
      {/* Toolbar */}
      <div className="flex items-center justify-between px-2 py-1 border-b border-editor-border bg-editor-surface/30">
        <div className="flex items-center gap-1">
          <button
            onClick={toggleFileTree}
            className="p-1.5 rounded hover:bg-editor-border/50 text-editor-muted hover:text-editor-text transition-colors"
            title={isFileTreeCollapsed ? 'Show Explorer' : 'Hide Explorer'}
          >
            {isFileTreeCollapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}
          </button>
        </div>
        <span className="text-xs text-editor-muted font-medium">Workspace Chat</span>
        <div className="flex items-center gap-1">
          <button
            onClick={togglePreview}
            className="p-1.5 rounded hover:bg-editor-border/50 text-editor-muted hover:text-editor-text transition-colors"
            title={isPreviewCollapsed ? 'Show Preview' : 'Hide Preview'}
          >
            {isPreviewCollapsed ? <PanelRightOpen size={16} /> : <PanelRightClose size={16} />}
          </button>
        </div>
      </div>

      {/* Main content with resizable panels */}
      <div className="flex-1 overflow-hidden">
        <ResizablePanelGroup direction="horizontal">
          {/* File Tree Panel */}
          {!isFileTreeCollapsed && (
            <>
              <ResizablePanel
                defaultSize={fileTreeWidth}
                minSize={150}
                maxSize={400}
                onResize={handleFileTreeResize}
                order={1}
              >
                <FileTree />
              </ResizablePanel>
              <ResizableHandle onDoubleClick={toggleFileTree} />
            </>
          )}

          {/* Chat Panel */}
          <ResizablePanel order={2}>
            <WorkspaceChat
              onFileClick={handleFileClick}
              onAddContext={handleAddContext}
            />
          </ResizablePanel>

          {/* Preview Panel */}
          {!isPreviewCollapsed && previewFile && (
            <>
              <ResizableHandle onDoubleClick={togglePreview} />
              <ResizablePanel
                defaultSize={previewWidth}
                minSize={200}
                maxSize={600}
                onResize={handlePreviewResize}
                order={3}
              >
                <FilePreview
                  file={previewFile}
                  highlights={previewHighlights}
                  onClose={handleClosePreview}
                  readOnly
                />
              </ResizablePanel>
            </>
          )}
        </ResizablePanelGroup>
      </div>
    </div>
  );
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
    json: 'json',
    md: 'markdown',
    py: 'python',
    go: 'go',
    rs: 'rust',
    yaml: 'yaml',
    yml: 'yaml',
    sh: 'shell',
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
