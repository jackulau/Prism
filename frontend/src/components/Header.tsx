import React from 'react';
import {
  PanelLeftClose,
  PanelLeft,
  FolderTree,
  BarChart3,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { useAppStore } from '../store';

export const Header: React.FC = () => {
  const {
    isSidebarOpen,
    toggleSidebar,
    isFileTreeOpen,
    toggleFileTree,
    isMetricsPanelOpen,
    toggleMetricsPanel,
    connectionStatus,
    metrics,
  } = useAppStore();

  const isConnected = connectionStatus === 'connected';

  return (
    <header className="h-12 bg-editor-bg border-b border-editor-border flex items-center justify-between px-4">
      {/* Left section */}
      <div className="flex items-center gap-2">
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-smooth"
          title={isSidebarOpen ? 'Hide sidebar' : 'Show sidebar'}
        >
          {isSidebarOpen ? (
            <PanelLeftClose className="w-5 h-5" />
          ) : (
            <PanelLeft className="w-5 h-5" />
          )}
        </button>

        <button
          onClick={toggleFileTree}
          className={`p-2 rounded-lg transition-smooth ${
            isFileTreeOpen
              ? 'bg-editor-accent/20 text-editor-accent'
              : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
          }`}
          title={isFileTreeOpen ? 'Hide file tree' : 'Show file tree'}
        >
          <FolderTree className="w-5 h-5" />
        </button>

        <div className="h-6 w-px bg-editor-border mx-1" />

        <span className="text-sm text-editor-muted">
          Prism AI Assistant
        </span>
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
              : 'bg-editor-error/20 text-editor-error'
          }`}
        >
          {isConnected ? (
            <Wifi className="w-3.5 h-3.5" />
          ) : (
            <WifiOff className="w-3.5 h-3.5" />
          )}
          <span className="capitalize">{connectionStatus}</span>
        </div>

        <div className="h-6 w-px bg-editor-border mx-1" />

        <button
          onClick={toggleMetricsPanel}
          className={`p-2 rounded-lg transition-smooth ${
            isMetricsPanelOpen
              ? 'bg-editor-accent/20 text-editor-accent'
              : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
          }`}
          title={isMetricsPanelOpen ? 'Hide metrics' : 'Show metrics'}
        >
          <BarChart3 className="w-5 h-5" />
        </button>
      </div>
    </header>
  );
};
