import { useState, memo } from 'react'
import {
  ChevronRight,
  ChevronDown,
  Wrench,
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  Copy,
  Check,
} from 'lucide-react'
import type { ToolCall } from '../../types'

interface ToolCallCardProps {
  toolCall: ToolCall
  onApprove?: (id: string) => void
  onReject?: (id: string) => void
  isMCP?: boolean
  serverName?: string
}

const statusConfig = {
  pending: {
    color: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: Clock,
    spin: false,
    label: 'Pending',
  },
  running: {
    color: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: Loader2,
    spin: true,
    label: 'Running',
  },
  completed: {
    color: 'bg-green-500/20 text-green-400 border-green-500/30',
    icon: CheckCircle,
    spin: false,
    label: 'Completed',
  },
  failed: {
    color: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: XCircle,
    spin: false,
    label: 'Failed',
  },
  rejected: {
    color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
    icon: XCircle,
    spin: false,
    label: 'Rejected',
  },
}

function JsonDisplay({ data, label }: { data: unknown; label: string }) {
  const [copied, setCopied] = useState(false)

  const jsonString = JSON.stringify(data, null, 2)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(jsonString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="border-t border-editor-border">
      <div className="flex items-center justify-between px-3 py-2 bg-editor-bg/30">
        <span className="text-xs text-editor-muted font-medium">{label}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          {copied ? (
            <>
              <Check size={10} className="text-editor-success" />
              <span className="text-editor-success">Copied</span>
            </>
          ) : (
            <>
              <Copy size={10} />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>
      <pre className="px-3 py-2 text-xs font-mono text-editor-text overflow-x-auto max-h-48 bg-editor-bg/20">
        {jsonString}
      </pre>
    </div>
  )
}

export const ToolCallCard = memo(function ToolCallCard({
  toolCall,
  onApprove,
  onReject,
  isMCP,
  serverName,
}: ToolCallCardProps) {
  const [isExpanded, setIsExpanded] = useState(toolCall.status === 'pending')

  const config = statusConfig[toolCall.status]
  const StatusIcon = config.icon

  // Format tool name for display
  const displayName = toolCall.name.replace(/^mcp_[^_]+_/, '') // Remove MCP prefix if present
  const hasParams = toolCall.parameters && Object.keys(toolCall.parameters).length > 0
  const hasResult = toolCall.result !== undefined && toolCall.result !== null

  return (
    <div className="rounded-lg border border-editor-border bg-editor-surface/50 overflow-hidden my-3 transition-all hover:border-editor-border/80">
      {/* Header */}
      <div
        className="flex items-center gap-2 px-3 py-2.5 cursor-pointer hover:bg-editor-surface/70 transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        {/* Expand/Collapse icon */}
        <span className="text-editor-muted">
          {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </span>

        {/* Tool icon */}
        <div className="p-1 rounded bg-orange-500/20 text-orange-400">
          <Wrench size={12} />
        </div>

        {/* Tool name and server */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-editor-text truncate">{displayName}</span>
            {isMCP && serverName && (
              <span className="px-1.5 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded border border-purple-500/30">
                MCP
              </span>
            )}
          </div>
          {serverName && (
            <span className="text-xs text-editor-muted truncate">{serverName}</span>
          )}
        </div>

        {/* Status badge */}
        <span
          className={`flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium border ${config.color}`}
        >
          <StatusIcon size={12} className={config.spin ? 'animate-spin' : ''} />
          {config.label}
        </span>
      </div>

      {/* Expandable content */}
      {isExpanded && (
        <div className="border-t border-editor-border">
          {/* Parameters section */}
          {hasParams && (
            <JsonDisplay data={toolCall.parameters} label="Parameters" />
          )}

          {/* Result section */}
          {hasResult && (
            <JsonDisplay data={toolCall.result} label="Result" />
          )}

          {/* No content message */}
          {!hasParams && !hasResult && toolCall.status === 'pending' && (
            <div className="px-3 py-4 text-center text-sm text-editor-muted">
              Waiting for approval...
            </div>
          )}

          {/* Approval buttons for pending tools */}
          {toolCall.status === 'pending' && onApprove && onReject && (
            <div className="flex gap-2 p-3 bg-editor-bg/30 border-t border-editor-border">
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onApprove(toolCall.id)
                }}
                className="flex-1 py-2 px-4 bg-editor-success/20 text-editor-success rounded-lg hover:bg-editor-success/30 transition-colors font-medium text-sm flex items-center justify-center gap-2"
              >
                <CheckCircle size={14} />
                Approve
              </button>
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onReject(toolCall.id)
                }}
                className="flex-1 py-2 px-4 bg-editor-error/20 text-editor-error rounded-lg hover:bg-editor-error/30 transition-colors font-medium text-sm flex items-center justify-center gap-2"
              >
                <XCircle size={14} />
                Reject
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
})

export default ToolCallCard
