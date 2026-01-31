import { useState, useEffect, useCallback, memo } from 'react'
import {
  StopCircle,
  Copy,
  Check,
  Clock,
  Coins,
  RotateCcw,
} from 'lucide-react'
import { AgentStatusBadge, type AgentStatus } from './AgentStatusBadge'
import { StreamingOutput } from './StreamingOutput'

interface TokenUsage {
  inputTokens: number
  outputTokens: number
}

interface AgentExecutionPanelProps {
  status: AgentStatus
  output: string
  isStreaming?: boolean
  startTime?: number
  endTime?: number
  tokenUsage?: TokenUsage
  onStop?: () => void
  onRestart?: () => void
  onClearOutput?: () => void
  title?: string
  className?: string
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds}s`
}

function formatTokens(count: number): string {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`
  if (count >= 1000) return `${(count / 1000).toFixed(1)}k`
  return count.toString()
}

export const AgentExecutionPanel = memo(function AgentExecutionPanel({
  status,
  output,
  isStreaming = false,
  startTime,
  endTime,
  tokenUsage,
  onStop,
  onRestart,
  onClearOutput,
  title = 'Agent Execution',
  className = '',
}: AgentExecutionPanelProps) {
  const [elapsedTime, setElapsedTime] = useState(0)
  const [copied, setCopied] = useState(false)

  // Timer for elapsed time during running state
  useEffect(() => {
    if (status !== 'running' || !startTime) {
      if (startTime && endTime) {
        setElapsedTime(endTime - startTime)
      }
      return
    }

    const updateElapsed = () => {
      setElapsedTime(Date.now() - startTime)
    }

    updateElapsed()
    const interval = setInterval(updateElapsed, 100)

    return () => clearInterval(interval)
  }, [status, startTime, endTime])

  const handleCopyOutput = useCallback(async () => {
    await navigator.clipboard.writeText(output)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [output])

  const canStop = status === 'running' || status === 'pending'
  const canRestart = status === 'completed' || status === 'failed' || status === 'cancelled'

  return (
    <div className={`flex flex-col border border-editor-border rounded-lg bg-editor-surface overflow-hidden ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-bg/50">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-medium text-editor-text">{title}</h3>
          <AgentStatusBadge status={status} size="sm" />
        </div>

        <div className="flex items-center gap-3">
          {/* Duration timer */}
          {startTime && (
            <div className="flex items-center gap-1.5 text-xs text-editor-muted">
              <Clock size={12} />
              <span className="font-mono">{formatDuration(elapsedTime)}</span>
            </div>
          )}

          {/* Token usage */}
          {tokenUsage && (
            <div className="flex items-center gap-1.5 text-xs text-editor-muted">
              <Coins size={12} />
              <span className="font-mono">
                {formatTokens(tokenUsage.inputTokens)} in / {formatTokens(tokenUsage.outputTokens)} out
              </span>
            </div>
          )}

          {/* Action buttons */}
          <div className="flex items-center gap-1 ml-2">
            {/* Copy output button */}
            <button
              onClick={handleCopyOutput}
              disabled={!output}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              title="Copy output"
            >
              {copied ? (
                <Check size={14} className="text-editor-success" />
              ) : (
                <Copy size={14} />
              )}
            </button>

            {/* Restart button */}
            {canRestart && onRestart && (
              <button
                onClick={onRestart}
                className="p-1.5 rounded text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 transition-colors"
                title="Restart execution"
              >
                <RotateCcw size={14} />
              </button>
            )}

            {/* Stop button */}
            {canStop && onStop && (
              <button
                onClick={onStop}
                className="flex items-center gap-1.5 px-2 py-1 rounded bg-red-500/20 text-red-400 hover:bg-red-500/30 border border-red-500/30 transition-colors text-xs font-medium"
                title="Stop execution"
              >
                <StopCircle size={12} />
                Stop
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Streaming output area */}
      <div className="flex-1 min-h-0">
        <StreamingOutput
          content={output}
          isStreaming={isStreaming}
          onClear={onClearOutput}
          maxHeight="500px"
        />
      </div>

      {/* Footer with summary */}
      {(status === 'completed' || status === 'failed' || status === 'cancelled') && (
        <div className="flex items-center justify-between px-4 py-2 border-t border-editor-border bg-editor-bg/30 text-xs text-editor-muted">
          <span>
            {status === 'completed' && 'Execution completed successfully'}
            {status === 'failed' && 'Execution failed'}
            {status === 'cancelled' && 'Execution was cancelled'}
          </span>
          {elapsedTime > 0 && (
            <span className="font-mono">
              Total time: {formatDuration(elapsedTime)}
            </span>
          )}
        </div>
      )}
    </div>
  )
})

export default AgentExecutionPanel
