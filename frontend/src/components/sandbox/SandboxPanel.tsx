import { Globe, Code, Terminal as TerminalIcon, RefreshCw, ExternalLink, Play, Square } from 'lucide-react'
import { BrowserPreview } from './BrowserPreview'
import { CodeViewer } from './CodeViewer'
import { Terminal } from './Terminal'
import { useSandboxStore } from '../../store/sandboxStore'

interface SandboxPanelProps {
  activeTab: 'preview' | 'code' | 'terminal'
  onTabChange: (tab: 'preview' | 'code' | 'terminal') => void
}

export function SandboxPanel({ activeTab, onTabChange }: SandboxPanelProps) {
  const {
    previewUrl,
    isRunning,
    buildStatus,
    refreshPreview,
    startBuild,
    stopBuild
  } = useSandboxStore()

  const tabs = [
    { id: 'preview' as const, label: 'Preview', icon: Globe },
    { id: 'code' as const, label: 'Code', icon: Code },
    { id: 'terminal' as const, label: 'Terminal', icon: TerminalIcon },
  ]

  return (
    <div className="h-full flex flex-col bg-editor-surface">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-bg">
        <div className="flex items-center gap-2">
          {/* Tabs */}
          <div className="flex items-center bg-editor-surface rounded-lg p-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => onTabChange(tab.id)}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'bg-editor-accent text-editor-bg'
                    : 'text-editor-muted hover:text-editor-text hover:bg-editor-border/50'
                }`}
              >
                <tab.icon size={14} />
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {/* Build Status */}
          {buildStatus && (
            <span className={`text-xs px-2 py-1 rounded ${
              buildStatus === 'success' ? 'bg-editor-success/20 text-editor-success' :
              buildStatus === 'error' ? 'bg-editor-error/20 text-editor-error' :
              buildStatus === 'building' ? 'bg-editor-warning/20 text-editor-warning' :
              'bg-editor-muted/20 text-editor-muted'
            }`}>
              {buildStatus === 'building' ? 'Building...' : buildStatus}
            </span>
          )}

          {/* Build Controls */}
          {isRunning ? (
            <button
              onClick={stopBuild}
              className="p-2 rounded-md text-editor-error hover:bg-editor-error/10 transition-colors"
              title="Stop build"
            >
              <Square size={16} />
            </button>
          ) : (
            <button
              onClick={() => startBuild()}
              className="p-2 rounded-md text-editor-success hover:bg-editor-success/10 transition-colors"
              title="Run build"
            >
              <Play size={16} />
            </button>
          )}

          {activeTab === 'preview' && (
            <>
              <button
                onClick={refreshPreview}
                className="p-2 rounded-md text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
                title="Refresh preview"
              >
                <RefreshCw size={16} />
              </button>
              {previewUrl && (
                <a
                  href={previewUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-2 rounded-md text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
                  title="Open in new tab"
                >
                  <ExternalLink size={16} />
                </a>
              )}
            </>
          )}
        </div>
      </div>

      {/* URL Bar (only for preview) */}
      {activeTab === 'preview' && (
        <div className="px-4 py-2 border-b border-editor-border bg-editor-bg/50">
          <div className="flex items-center gap-2 bg-editor-surface rounded-lg px-3 py-1.5">
            <Globe size={14} className="text-editor-muted" />
            <input
              type="text"
              value={previewUrl || 'No preview available'}
              readOnly
              className="flex-1 bg-transparent text-sm text-editor-text outline-none"
            />
          </div>
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'preview' && <BrowserPreview />}
        {activeTab === 'code' && <CodeViewer />}
        {activeTab === 'terminal' && <Terminal />}
      </div>
    </div>
  )
}
