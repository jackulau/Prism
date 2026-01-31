import { useEffect, useRef, useState, useCallback, memo } from 'react'
import {
  Trash2,
  Copy,
  Check,
  ChevronDown,
  WrapText,
  ArrowDownToLine,
} from 'lucide-react'

interface StreamingOutputProps {
  content: string
  isStreaming?: boolean
  onClear?: () => void
  maxHeight?: string
  showLineNumbers?: boolean
}

export const StreamingOutput = memo(function StreamingOutput({
  content,
  isStreaming = false,
  onClear,
  maxHeight = '400px',
  showLineNumbers = false,
}: StreamingOutputProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [autoScroll, setAutoScroll] = useState(true)
  const [wordWrap, setWordWrap] = useState(true)
  const [copied, setCopied] = useState(false)
  const [isAtBottom, setIsAtBottom] = useState(true)

  const lines = content.split('\n')

  const scrollToBottom = useCallback(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [])

  const handleScroll = useCallback(() => {
    if (!containerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current
    const atBottom = scrollHeight - scrollTop - clientHeight < 50
    setIsAtBottom(atBottom)
    if (!atBottom && autoScroll) {
      setAutoScroll(false)
    }
  }, [autoScroll])

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      scrollToBottom()
    }
  }, [content, autoScroll, scrollToBottom])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleAutoScrollToggle = () => {
    setAutoScroll(!autoScroll)
    if (!autoScroll) {
      scrollToBottom()
    }
  }

  const handleScrollToBottom = () => {
    scrollToBottom()
    setAutoScroll(true)
  }

  return (
    <div className="flex flex-col border border-editor-border rounded-lg bg-editor-surface overflow-hidden">
      {/* Header with controls */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-editor-border bg-editor-bg/50">
        <span className="text-xs text-editor-muted font-medium">Output</span>
        <div className="flex items-center gap-1">
          {/* Word wrap toggle */}
          <button
            onClick={() => setWordWrap(!wordWrap)}
            className={`p-1.5 rounded transition-colors ${
              wordWrap
                ? 'text-editor-accent bg-editor-accent/10'
                : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
            }`}
            title={wordWrap ? 'Disable word wrap' : 'Enable word wrap'}
            aria-pressed={wordWrap}
          >
            <WrapText size={14} />
          </button>

          {/* Auto-scroll toggle */}
          <button
            onClick={handleAutoScrollToggle}
            className={`p-1.5 rounded transition-colors ${
              autoScroll
                ? 'text-editor-accent bg-editor-accent/10'
                : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
            }`}
            title={autoScroll ? 'Disable auto-scroll' : 'Enable auto-scroll'}
            aria-pressed={autoScroll}
          >
            <ChevronDown size={14} />
          </button>

          {/* Copy button */}
          <button
            onClick={handleCopy}
            disabled={!content}
            className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Copy output"
          >
            {copied ? (
              <Check size={14} className="text-editor-success" />
            ) : (
              <Copy size={14} />
            )}
          </button>

          {/* Clear button */}
          {onClear && (
            <button
              onClick={onClear}
              disabled={!content}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              title="Clear output"
            >
              <Trash2 size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Output content */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="overflow-auto font-mono text-sm"
        style={{ maxHeight }}
      >
        {content ? (
          <div className="p-3">
            {lines.map((line, index) => (
              <div
                key={index}
                className={`${wordWrap ? 'whitespace-pre-wrap break-words' : 'whitespace-pre'} ${
                  showLineNumbers ? 'flex' : ''
                }`}
              >
                {showLineNumbers && (
                  <span className="text-editor-muted/50 select-none pr-3 text-right w-8 flex-shrink-0">
                    {index + 1}
                  </span>
                )}
                <span className="text-editor-text">
                  {line}
                  {isStreaming && index === lines.length - 1 && (
                    <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
                  )}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-4 text-center text-editor-muted text-sm">
            <p>Waiting for output...</p>
          </div>
        )}
      </div>

      {/* Scroll to bottom indicator */}
      {!isAtBottom && content && (
        <button
          onClick={handleScrollToBottom}
          className="absolute bottom-4 right-4 p-2 rounded-full bg-editor-accent text-editor-bg shadow-lg hover:bg-editor-accent/90 transition-colors"
          title="Scroll to bottom"
        >
          <ArrowDownToLine size={16} />
        </button>
      )}
    </div>
  )
})

export default StreamingOutput
