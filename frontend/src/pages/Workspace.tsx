import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import {
  FolderTree as FolderTreeIcon,
  MessageSquare,
  Wifi,
  WifiOff,
  History,
  ChevronDown
} from 'lucide-react';
import { SandboxPanel } from '../components/sandbox/SandboxPanel';
import { FileTree } from '../components/FileTree';
import { MetricsDropdown } from '../components/MetricsDropdown';
import { EnhancedChatPanel } from '../components/chat/EnhancedChatPanel';
import { BuildHistoryPanel } from '../components/builds/BuildHistoryPanel';
import { useAppStore } from '../store';
import { useSandboxStore, FileNode } from '../store/sandboxStore';
import { apiService } from '../services/api';
import { wsService } from '../services/websocket';

// Helper to convert API sandbox files to FileNode format
function convertToFileNodes(files: Array<{
  name: string;
  path: string;
  is_directory: boolean;
  children?: Array<{
    name: string;
    path: string;
    is_directory: boolean;
  }>;
}>): FileNode[] {
  return files.map(f => ({
    name: f.name,
    path: f.path,
    isDirectory: f.is_directory,
    children: f.children ? convertToFileNodes(f.children) : undefined,
  }));
}

export default function Workspace() {
  const { id } = useParams<{ id?: string }>();
  const [activeTab, setActiveTab] = useState<'preview' | 'code' | 'terminal'>('preview');
  const [isBuildHistoryOpen, setIsBuildHistoryOpen] = useState(false);
  const [buildHistoryHeight, setBuildHistoryHeight] = useState(300);

  const {
    isFileTreeOpen,
    toggleFileTree,
    isChatPanelOpen,
    toggleChatPanel,
    connectionStatus,
    metrics,
  } = useAppStore();

  const { setFiles } = useSandboxStore();

  const isConnected = connectionStatus === 'connected';

  // Load sandbox files when connected
  useEffect(() => {
    if (isConnected) {
      apiService.getSandboxFiles().then(response => {
        if (response.data?.files) {
          setFiles(convertToFileNodes(response.data.files));
        }
      });
    }
  }, [isConnected, setFiles]);

  return (
    <>
      {/* Header Bar */}
      <header className="h-12 bg-editor-bg border-b border-editor-border flex items-center justify-between px-4 flex-shrink-0">
        {/* Left section */}
        <div className="flex items-center gap-2">
          <button
            onClick={toggleFileTree}
            className={`p-2 rounded-lg transition-colors ${
              isFileTreeOpen
                ? 'bg-editor-accent/20 text-editor-accent'
                : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
            }`}
            title={isFileTreeOpen ? 'Hide file tree' : 'Show file tree'}
          >
            <FolderTreeIcon size={20} />
          </button>

          <button
            onClick={toggleChatPanel}
            className={`p-2 rounded-lg transition-colors ${
              isChatPanelOpen
                ? 'bg-editor-accent/20 text-editor-accent'
                : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
            }`}
            title={isChatPanelOpen ? 'Hide chat' : 'Show chat'}
          >
            <MessageSquare size={20} />
          </button>

          <button
            onClick={() => setIsBuildHistoryOpen(!isBuildHistoryOpen)}
            className={`p-2 rounded-lg transition-colors ${
              isBuildHistoryOpen
                ? 'bg-editor-accent/20 text-editor-accent'
                : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
            }`}
            title={isBuildHistoryOpen ? 'Hide build history' : 'Show build history'}
          >
            <History size={20} />
          </button>

          <div className="h-6 w-px bg-editor-border mx-1" />

          {id && (
            <span className="text-sm text-editor-muted">
              Workspace: <span className="text-editor-text font-mono">{id}</span>
            </span>
          )}
        </div>

        {/* Center section - Live metrics when generating */}
        {metrics.isGenerating && (
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-editor-success animate-pulse" />
              <span className="text-editor-success">Live</span>
            </div>
            <div className="flex items-center gap-1 text-editor-text">
              <span className="font-mono">{metrics.tokensPerSecond.toFixed(1)}</span>
              <span className="text-editor-muted">t/s</span>
            </div>
            <div className="flex items-center gap-1 text-editor-text">
              <span className="font-mono">{metrics.tokenCount}</span>
              <span className="text-editor-muted">tokens</span>
            </div>
            {metrics.timeToFirstToken !== null && (
              <div className="flex items-center gap-1 text-editor-text">
                <span className="font-mono">{metrics.timeToFirstToken.toFixed(0)}</span>
                <span className="text-editor-muted">ms TTFT</span>
              </div>
            )}
          </div>
        )}

        {/* Right section */}
        <div className="flex items-center gap-2">
          {/* Connection status */}
          <div
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs ${
              isConnected
                ? 'bg-editor-success/20 text-editor-success'
                : connectionStatus === 'connecting'
                ? 'bg-editor-warning/20 text-editor-warning'
                : 'bg-editor-error/20 text-editor-error'
            }`}
          >
            {isConnected ? <Wifi size={14} /> : <WifiOff size={14} />}
            <span className="capitalize">{connectionStatus}</span>
          </div>

          {/* Reconnect button when connection fails */}
          {connectionStatus === 'error' && (
            <button
              onClick={() => wsService.manualReconnect()}
              className="px-3 py-1 text-xs bg-editor-accent text-white rounded-full hover:bg-editor-accent/80 transition-colors"
            >
              Reconnect
            </button>
          )}

          <div className="h-6 w-px bg-editor-border mx-1" />

          <MetricsDropdown />
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Upper Section */}
        <div className="flex-1 flex overflow-hidden min-h-0">
          {/* File Tree Panel */}
          {isFileTreeOpen && (
            <div className="w-64 flex-shrink-0 border-r border-editor-border">
              <FileTree />
            </div>
          )}

          {/* Main Content Area */}
          <div className="flex-1 flex overflow-hidden">
            {/* Chat Panel */}
            {isChatPanelOpen && (
              <div className="w-1/2 border-r border-editor-border flex flex-col min-w-0">
                <EnhancedChatPanel />
              </div>
            )}

            {/* Sandbox Preview Panel */}
            <div className={`flex flex-col min-w-0 ${isChatPanelOpen ? 'w-1/2' : 'flex-1'}`}>
              <SandboxPanel
                activeTab={activeTab}
                onTabChange={setActiveTab}
              />
            </div>
          </div>
        </div>

        {/* Build History Panel (collapsible bottom panel) */}
        {isBuildHistoryOpen && (
          <div
            className="flex-shrink-0 border-t border-editor-border"
            style={{ height: buildHistoryHeight }}
          >
            {/* Resize Handle */}
            <div
              className="h-1 bg-editor-border hover:bg-editor-accent cursor-row-resize"
              onMouseDown={(e) => {
                e.preventDefault();
                const startY = e.clientY;
                const startHeight = buildHistoryHeight;
                const handleMouseMove = (moveEvent: MouseEvent) => {
                  const delta = startY - moveEvent.clientY;
                  const newHeight = Math.min(Math.max(startHeight + delta, 150), 600);
                  setBuildHistoryHeight(newHeight);
                };
                const handleMouseUp = () => {
                  document.removeEventListener('mousemove', handleMouseMove);
                  document.removeEventListener('mouseup', handleMouseUp);
                };
                document.addEventListener('mousemove', handleMouseMove);
                document.addEventListener('mouseup', handleMouseUp);
              }}
            />
            {/* Panel Header with collapse button */}
            <div className="h-8 flex items-center justify-between px-3 bg-editor-bg border-b border-editor-border">
              <span className="text-xs font-medium text-editor-muted uppercase tracking-wide">
                Build History
              </span>
              <button
                onClick={() => setIsBuildHistoryOpen(false)}
                className="p-1 text-editor-muted hover:text-editor-text rounded transition-colors"
                title="Close panel"
              >
                <ChevronDown size={14} />
              </button>
            </div>
            <div className="h-[calc(100%-36px)]">
              <BuildHistoryPanel />
            </div>
          </div>
        )}
      </div>
    </>
  );
}
