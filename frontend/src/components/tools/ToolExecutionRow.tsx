import { useState, memo } from 'react'
import {
  ChevronRight,
  ChevronDown,
  Wrench,
  Loader2,
  CheckCircle,
  XCircle,
  Clock,
  Copy,
  Check,
  Server,
} from 'lucide-react'

export interface ToolExecution {
  id: string
  message_id: string
  tool_name: string
  parameters: string
  result?: string
  status: string
  execution_time_ms?: number
  container_id?: string
  created_at: string
}

interface ToolExecutionRowProps {
  execution: ToolExecution
  defaultExpanded?: boolean
}

const statusConfig: Record<string, { color: string; icon: typeof Clock; spin: boolean; label: string }> = {
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
}

function JsonDisplay({ data, label }: { data: string; label: string }) {
  const [copied, setCopied] = useState(false)

  let formattedData: string
  try {
    const parsed = JSON.parse(data)
    formattedData = JSON.stringify(parsed, null, 2)
  } catch {
    formattedData = data
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(formattedData)
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
        {formattedData}
      </pre>
    </div>
  )
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`
  }
  const seconds = ms / 1000
  if (seconds < 60) {
    return `${seconds.toFixed(1)}s`
  }
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds.toFixed(0)}s`
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

export const ToolExecutionRow = memo(function ToolExecutionRow({
  execution,
  defaultExpanded = false,
}: ToolExecutionRowProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded || execution.status === 'running')

  const config = statusConfig[execution.status] || statusConfig.pending
  const StatusIcon = config.icon

  const hasParams = execution.parameters && execution.parameters !== '{}'
  const hasResult = execution.result && execution.result !== ''

  return (
    <div className="rounded-lg border border-editor-border bg-editor-surface/50 overflow-hidden transition-all hover:border-editor-border/80">
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

        {/* Tool name and timestamp */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-editor-text truncate">
              {execution.tool_name}
            </span>
          </div>
          <div className="flex items-center gap-2 text-xs text-editor-muted">
            <span>{formatTimestamp(execution.created_at)}</span>
            {execution.execution_time_ms !== undefined && execution.execution_time_ms > 0 && (
              <>
                <span>-</span>
                <span>{formatDuration(execution.execution_time_ms)}</span>
              </>
            )}
          </div>
        </div>

        {/* Container badge (if present) */}
        {execution.container_id && (
          <span className="flex items-center gap-1 px-1.5 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded border border-purple-500/30">
            <Server size={10} />
            Container
          </span>
        )}

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
            <JsonDisplay data={execution.parameters} label="Parameters" />
          )}

          {/* Result section */}
          {hasResult && (
            <JsonDisplay
              data={execution.result || ''}
              label={execution.status === 'failed' ? 'Error' : 'Result'}
            />
          )}

          {/* No content message */}
          {!hasParams && !hasResult && (
            <div className="px-3 py-4 text-center text-sm text-editor-muted">
              {execution.status === 'running' ? 'Executing...' : 'No data available'}
            </div>
          )}

          {/* Container info */}
          {execution.container_id && (
            <div className="px-3 py-2 bg-editor-bg/30 border-t border-editor-border">
              <span className="text-xs text-editor-muted">
                Container ID: <code className="font-mono">{execution.container_id}</code>
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  )
})

export default ToolExecutionRow
