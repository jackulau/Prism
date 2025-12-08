import { useState, useEffect } from 'react'
import {
  PanelLeftClose,
  PanelLeft,
  FolderTree as FolderTreeIcon,
  MessageSquare,
  Wifi,
  WifiOff
} from 'lucide-react'
import { SandboxPanel } from './components/sandbox/SandboxPanel'
import { Sidebar } from './components/sidebar/Sidebar'
import { FileTree } from './components/FileTree'
import { MetricsDropdown } from './components/MetricsDropdown'
import { EnhancedChatPanel } from './components/chat/EnhancedChatPanel'
import { SettingsPanel } from './components/settings/SettingsPanel'
import { ToastContainer } from './components/Toast'
import { useAppStore } from './store'
import { useSandboxStore, FileNode } from './store/sandboxStore'
import { apiService } from './services/api'
import { wsService } from './services/websocket'
import { applyTheme } from './config/themes'

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

function App() {
  const [activeTab, setActiveTab] = useState<'preview' | 'code' | 'terminal'>('preview')
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  const {
    isFileTreeOpen,
    toggleFileTree,
    isChatPanelOpen,
    toggleChatPanel,
    connectionStatus,
    metrics,
    theme,
    loadProviders,
  } = useAppStore()

  // Apply theme on mount and when it changes
  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  // Load providers on mount
  useEffect(() => {
    loadProviders()
  }, [loadProviders])

  const { setFiles } = useSandboxStore()

  const isConnected = connectionStatus === 'connected'

  // Connect WebSocket on mount
  useEffect(() => {
    wsService.connect()
    return () => wsService.disconnect()
  }, [])

  // Load initial sandbox files when connected
  useEffect(() => {
    if (isConnected) {
      apiService.getSandboxFiles().then(response => {
        if (response.data?.files) {
          setFiles(convertToFileNodes(response.data.files));
        }
      });
    }
  }, [isConnected, setFiles])

  return (
    <div className="h-screen w-screen flex flex-col bg-editor-bg text-editor-text overflow-hidden">
      {/* Top Header Bar */}
      <header className="h-12 bg-editor-bg border-b border-editor-border flex items-center justify-between px-4 flex-shrink-0">
        {/* Left section */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
            title={isSidebarCollapsed ? 'Show sidebar' : 'Hide sidebar'}
          >
            {isSidebarCollapsed ? <PanelLeft size={20} /> : <PanelLeftClose size={20} />}
          </button>

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

          <div className="h-6 w-px bg-editor-border mx-1" />
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
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <Sidebar
          isCollapsed={isSidebarCollapsed}
          onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />

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

      {/* Settings Panel (slide-out) */}
      <SettingsPanel />

      {/* Toast Notifications */}
      <ToastContainer />
    </div>
  )
}

export default App
