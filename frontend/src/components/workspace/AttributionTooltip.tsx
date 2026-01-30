import { useState, useRef, useEffect } from 'react'
import { Clock, User, Bot, Workflow, Wrench, ExternalLink } from 'lucide-react'
import type { FileHistoryEntry } from '../../store/sandboxStore'

export interface AttributionTooltipProps {
  entry: FileHistoryEntry
  children: React.ReactNode
  onNavigateToMessage?: (messageId: string, conversationId: string) => void
}

function formatTimestamp(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getAgentIcon(type?: string) {
  switch (type) {
    case 'assistant':
      return <User size={14} className="text-blue-400" />
    case 'autonomous':
      return <Bot size={14} className="text-purple-400" />
    case 'workflow':
      return <Workflow size={14} className="text-green-400" />
    default:
      return null
  }
}

export function AttributionTooltip({
  entry,
  children,
  onNavigateToMessage,
}: AttributionTooltipProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [position, setPosition] = useState({ top: 0, left: 0 })
  const triggerRef = useRef<HTMLDivElement>(null)
  const tooltipRef = useRef<HTMLDivElement>(null)
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  const hasAttribution = entry.agent_name || entry.tool_name

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const handleMouseEnter = () => {
    if (!hasAttribution) return

    timeoutRef.current = setTimeout(() => {
      if (triggerRef.current) {
        const rect = triggerRef.current.getBoundingClientRect()
        setPosition({
          top: rect.bottom + 8,
          left: rect.left,
        })
        setIsVisible(true)
      }
    }, 300)
  }

  const handleMouseLeave = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    setIsVisible(false)
  }

  const handleConversationClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (entry.message_id && entry.conversation_id && onNavigateToMessage) {
      onNavigateToMessage(entry.message_id, entry.conversation_id)
    }
  }

  return (
    <div
      ref={triggerRef}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      className="inline-block"
    >
      {children}

      {isVisible && hasAttribution && (
        <div
          ref={tooltipRef}
          className="fixed z-50 w-64 p-3 rounded-lg bg-editor-bg border border-editor-border shadow-xl animate-fade-in"
          style={{ top: position.top, left: position.left }}
          onMouseEnter={() => {
            if (timeoutRef.current) clearTimeout(timeoutRef.current)
          }}
          onMouseLeave={handleMouseLeave}
        >
          {/* Agent Info */}
          {entry.agent_name && (
            <div className="flex items-center gap-2 mb-2">
              {getAgentIcon(entry.agent_type)}
              <div>
                <div className="text-sm font-medium text-editor-text">
                  {entry.agent_name}
                </div>
                <div className="text-xs text-editor-muted capitalize">
                  {entry.agent_type || 'Agent'}
                </div>
              </div>
            </div>
          )}

          {/* Tool Info */}
          {entry.tool_name && (
            <div className="flex items-center gap-2 mb-2">
              <Wrench size={14} className="text-editor-muted" />
              <div className="text-sm text-editor-text">
                {entry.tool_name}
              </div>
            </div>
          )}

          {/* Description */}
          {entry.description && (
            <p className="text-xs text-editor-muted mb-2 line-clamp-2">
              {entry.description}
            </p>
          )}

          {/* Timestamp */}
          <div className="flex items-center gap-2 text-xs text-editor-muted">
            <Clock size={12} />
            <span>{formatTimestamp(entry.created_at)}</span>
          </div>

          {/* Conversation Link */}
          {entry.message_id && entry.conversation_id && onNavigateToMessage && (
            <button
              onClick={handleConversationClick}
              className="mt-2 w-full flex items-center justify-center gap-1.5 px-2 py-1.5 rounded-md bg-editor-surface border border-editor-border hover:border-editor-accent/50 text-xs text-editor-muted hover:text-editor-text transition-colors"
            >
              <ExternalLink size={12} />
              View conversation
            </button>
          )}
        </div>
      )}
    </div>
  )
}
