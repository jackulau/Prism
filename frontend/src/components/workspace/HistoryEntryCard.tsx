import { Clock } from 'lucide-react'
import type { FileHistoryEntry } from '../../store/sandboxStore'
import { AttributionBadge } from './AttributionBadge'
import { AttributionTooltip } from './AttributionTooltip'
import { ConversationLink } from './ConversationLink'

export interface HistoryEntryCardProps {
  entry: FileHistoryEntry
  isSelected?: boolean
  onClick?: () => void
  onNavigateToMessage?: (messageId: string, conversationId: string) => void
}

function getOperationStyle(operation: string): string {
  switch (operation) {
    case 'create':
      return 'bg-green-500/20 text-green-400'
    case 'update':
      return 'bg-blue-500/20 text-blue-400'
    case 'delete':
      return 'bg-red-500/20 text-red-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
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

export function HistoryEntryCard({
  entry,
  isSelected = false,
  onClick,
  onNavigateToMessage,
}: HistoryEntryCardProps) {
  const hasAttribution = entry.agent_name || entry.tool_name

  return (
    <AttributionTooltip
      entry={entry}
      onNavigateToMessage={onNavigateToMessage}
    >
      <button
        onClick={onClick}
        className={`w-full p-3 text-left border-b border-editor-border/50 hover:bg-editor-border/30 transition-colors ${
          isSelected ? 'bg-editor-accent/10' : ''
        }`}
      >
        {/* Header: Operation badge and size */}
        <div className="flex items-center justify-between gap-2 mb-2">
          <div className="flex items-center gap-2">
            <span className={`text-xs px-1.5 py-0.5 rounded ${getOperationStyle(entry.operation)}`}>
              {entry.operation}
            </span>
            <span className="text-xs text-editor-muted">
              {formatSize(entry.size)}
            </span>
          </div>
          {entry.message_id && entry.conversation_id && (
            <ConversationLink
              messageId={entry.message_id}
              conversationId={entry.conversation_id}
              onClick={() => onNavigateToMessage?.(entry.message_id!, entry.conversation_id!)}
            />
          )}
        </div>

        {/* Attribution badge */}
        {hasAttribution && (
          <div className="mb-2">
            <AttributionBadge
              agentName={entry.agent_name}
              agentType={entry.agent_type}
              toolName={entry.tool_name}
              size="sm"
            />
          </div>
        )}

        {/* Description if available */}
        {entry.description && (
          <p className="text-xs text-editor-muted mb-2 line-clamp-2">
            {entry.description}
          </p>
        )}

        {/* Timestamp */}
        <div className="flex items-center gap-1.5 text-xs text-editor-muted">
          <Clock size={10} />
          <span>{formatTimestamp(entry.created_at)}</span>
        </div>
      </button>
    </AttributionTooltip>
  )
}
