import { useEffect, useRef } from 'react'
import { Trash2 } from 'lucide-react'
import { useSandboxStore } from '../../store/sandboxStore'

export function Terminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const { terminalOutput, clearTerminal } = useSandboxStore()

  // Auto-scroll to bottom when new output is added
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight
    }
  }, [terminalOutput])

  return (
    <div className="h-full flex flex-col bg-editor-surface">
      {/* Terminal Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-bg">
        <span className="text-sm text-editor-muted">Build Output</span>
        <button
          onClick={clearTerminal}
          className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
          title="Clear terminal"
        >
          <Trash2 size={14} />
        </button>
      </div>

      {/* Terminal Content */}
      <div
        ref={terminalRef}
        className="flex-1 overflow-auto p-4 font-mono text-sm"
      >
        {terminalOutput.length > 0 ? (
          terminalOutput.map((line, index) => (
            <div
              key={index}
              className={`whitespace-pre-wrap ${
                line.type === 'error' ? 'text-editor-error' :
                line.type === 'warning' ? 'text-editor-warning' :
                line.type === 'success' ? 'text-editor-success' :
                line.type === 'info' ? 'text-editor-accent' :
                'text-editor-text'
              }`}
            >
              {line.timestamp && (
                <span className="text-editor-muted opacity-50 mr-2">
                  [{new Date(line.timestamp).toLocaleTimeString()}]
                </span>
              )}
              {line.content}
            </div>
          ))
        ) : (
          <div className="text-editor-muted opacity-50">
            <p>$ Waiting for build output...</p>
            <p className="mt-2">Run a build to see output here.</p>
          </div>
        )}
      </div>
    </div>
  )
}
