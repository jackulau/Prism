import { useState, useRef, useEffect } from 'react'
import { MessageSquare, ExternalLink } from 'lucide-react'

export interface ConversationLinkProps {
  messageId: string
  conversationId: string
  preview?: string
  onClick?: () => void
}

export function ConversationLink({
  messageId: _messageId,
  conversationId: _conversationId,
  preview,
  onClick,
}: ConversationLinkProps) {
  // Note: messageId and conversationId are part of the interface for future navigation use
  void _messageId
  void _conversationId
  const [showTooltip, setShowTooltip] = useState(false)
  const [tooltipPosition, setTooltipPosition] = useState({ top: 0, left: 0 })
  const buttonRef = useRef<HTMLButtonElement>(null)
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const handleMouseEnter = () => {
    timeoutRef.current = setTimeout(() => {
      if (buttonRef.current) {
        const rect = buttonRef.current.getBoundingClientRect()
        setTooltipPosition({
          top: rect.bottom + 4,
          left: rect.left,
        })
        setShowTooltip(true)
      }
    }, 400)
  }

  const handleMouseLeave = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    setShowTooltip(false)
  }

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onClick?.()
  }

  return (
    <>
      <button
        ref={buttonRef}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className="p-1 rounded hover:bg-editor-surface text-editor-muted hover:text-editor-accent transition-colors"
        title="View in conversation"
      >
        <MessageSquare size={14} />
      </button>

      {showTooltip && preview && (
        <div
          className="fixed z-50 max-w-xs p-2 rounded-md bg-editor-bg border border-editor-border shadow-lg animate-fade-in"
          style={{ top: tooltipPosition.top, left: tooltipPosition.left }}
          onMouseEnter={() => {
            if (timeoutRef.current) clearTimeout(timeoutRef.current)
          }}
          onMouseLeave={handleMouseLeave}
        >
          <div className="flex items-start gap-2">
            <MessageSquare size={12} className="text-editor-muted mt-0.5 flex-shrink-0" />
            <div className="text-xs text-editor-text line-clamp-3">
              {preview}
            </div>
          </div>
          <div className="flex items-center gap-1 mt-1.5 text-xs text-editor-accent">
            <ExternalLink size={10} />
            <span>Click to view</span>
          </div>
        </div>
      )}
    </>
  )
}
