import { useState, useRef, useEffect, useMemo } from 'react'
import { Send, Bot, User, StopCircle, Paperclip, Hash, Zap, Clock, RotateCcw, X } from 'lucide-react'
import { Highlight, themes, type RenderProps } from 'prism-react-renderer'
import { useAppStore } from '../../store'
import { wsService } from '../../services/websocket'
import { MessageQueue } from './MessageQueue'
import type { Message } from '../../types'

// Code block component with syntax highlighting
function CodeBlock({ code, language, isStreaming }: { code: string; language: string; isStreaming?: boolean }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="rounded-lg overflow-hidden border border-editor-border bg-editor-surface my-3">
      <div className="flex items-center justify-between px-4 py-2 bg-editor-bg/50 border-b border-editor-border">
        <span className="text-xs text-editor-muted">{language}</span>
        <button
          onClick={handleCopy}
          className="text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <Highlight theme={themes.nightOwl} code={code.trim()} language={language}>
        {({ className, style, tokens, getLineProps, getTokenProps }: RenderProps) => (
          <pre className={`${className} p-4 text-sm font-mono overflow-x-auto`} style={{ ...style, background: 'transparent', margin: 0 }}>
            {tokens.map((line, i) => {
              const lineProps = getLineProps({ line, key: i })
              const isLastLine = i === tokens.length - 1
              return (
                <div key={i} {...lineProps} className="table-row">
                  <span className="table-cell pr-4 text-editor-muted/50 select-none text-right w-8">{i + 1}</span>
                  <span className="table-cell">
                    {line.map((token, key) => (
                      <span key={key} {...getTokenProps({ token, key })} />
                    ))}
                    {isStreaming && isLastLine && (
                      <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
                    )}
                  </span>
                </div>
              )
            })}
          </pre>
        )}
      </Highlight>
    </div>
  )
}

// Parse content for code blocks and formatting
function parseContent(content: string, isStreaming: boolean) {
  const elements: React.ReactNode[] = []
  let currentIndex = 0
  const codeBlockRegex = /```(\w*)\n?([\s\S]*?)```/g
  let match

  while ((match = codeBlockRegex.exec(content)) !== null) {
    if (match.index > currentIndex) {
      const textBefore = content.slice(currentIndex, match.index)
      elements.push(
        <span key={`text-${currentIndex}`} className="whitespace-pre-wrap">
          {parseInlineFormatting(textBefore)}
        </span>
      )
    }

    const language = match[1] || 'text'
    const code = match[2]
    const isLastBlock = codeBlockRegex.lastIndex >= content.length - 3

    elements.push(
      <CodeBlock
        key={`code-${match.index}`}
        code={code}
        language={language}
        isStreaming={isStreaming && isLastBlock}
      />
    )
    currentIndex = codeBlockRegex.lastIndex
  }

  if (currentIndex < content.length) {
    const remainingText = content.slice(currentIndex)
    const unclosedCodeBlock = remainingText.match(/```(\w*)\n?([\s\S]*)$/)

    if (unclosedCodeBlock && isStreaming) {
      elements.push(
        <CodeBlock
          key="code-partial"
          code={unclosedCodeBlock[2]}
          language={unclosedCodeBlock[1] || 'text'}
          isStreaming={true}
        />
      )
    } else {
      elements.push(
        <span key={`text-${currentIndex}`} className="whitespace-pre-wrap">
          {parseInlineFormatting(remainingText)}
          {isStreaming && !unclosedCodeBlock && (
            <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
          )}
        </span>
      )
    }
  }

  return elements
}

// Parse inline formatting
function parseInlineFormatting(text: string) {
  const elements: React.ReactNode[] = []
  let remaining = text
  let keyIndex = 0

  while (remaining.length > 0) {
    const boldMatch = remaining.match(/^\*\*(.+?)\*\*/)
    if (boldMatch) {
      elements.push(<strong key={`bold-${keyIndex++}`} className="font-semibold">{boldMatch[1]}</strong>)
      remaining = remaining.slice(boldMatch[0].length)
      continue
    }

    const codeMatch = remaining.match(/^`([^`]+)`/)
    if (codeMatch) {
      elements.push(
        <code key={`code-${keyIndex++}`} className="px-1.5 py-0.5 rounded bg-editor-surface text-editor-accent text-sm font-mono border border-editor-border">
          {codeMatch[1]}
        </code>
      )
      remaining = remaining.slice(codeMatch[0].length)
      continue
    }

    elements.push(remaining[0])
    remaining = remaining.slice(1)
  }

  return elements
}

// Message bubble component
function MessageBubble({ message, onRollback, isGenerating, onStopAndRollback }: {
  message: Message;
  onRollback?: (id: string) => void;
  isGenerating?: boolean;
  onStopAndRollback?: (id: string) => void;
}) {
  const isUser = message.role === 'user'
  const content = useMemo(() => parseContent(message.content, message.isStreaming || false), [message.content, message.isStreaming])

  const handleRollback = () => {
    if (isGenerating && onStopAndRollback) {
      // Stop generation first, then rollback
      onStopAndRollback(message.id)
    } else if (onRollback) {
      onRollback(message.id)
    }
  }

  return (
    <div className={`py-4 px-4 ${isUser ? 'bg-editor-surface/30' : ''} group`}>
      <div className="max-w-3xl mx-auto">
        <div className="flex items-center gap-3 mb-2">
          <div className={`w-8 h-8 rounded-lg flex items-center justify-center border ${
            isUser
              ? 'bg-blue-500/20 text-blue-400 border-blue-500/30'
              : 'bg-purple-500/20 text-purple-400 border-purple-500/30'
          }`}>
            {isUser ? <User size={18} /> : <Bot size={18} />}
          </div>
          <div className="flex items-center gap-2 flex-1">
            <span className="font-medium text-editor-text">{isUser ? 'You' : 'Assistant'}</span>
            <span className="text-xs text-editor-muted">{message.timestamp.toLocaleTimeString()}</span>
            {message.isStreaming && (
              <span className="flex items-center gap-1 text-xs text-editor-accent">
                <span className="w-1.5 h-1.5 rounded-full bg-editor-accent animate-pulse" />
                Generating...
              </span>
            )}
          </div>
          {/* Rollback button for user messages - now works during generation too */}
          {isUser && (onRollback || onStopAndRollback) && (
            <button
              onClick={handleRollback}
              className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-editor-warning/20 text-editor-muted hover:text-editor-warning transition-all"
              title={isGenerating ? "Stop and rollback" : "Rollback to this message"}
            >
              <RotateCcw size={14} />
            </button>
          )}
        </div>

        <div className="pl-11 text-editor-text leading-relaxed">{content}</div>

        {/* Metrics for completed assistant messages */}
        {!isUser && message.metrics && !message.isStreaming && (
          <div className="pl-11 mt-3 flex items-center gap-4 text-xs text-editor-muted">
            <span className="flex items-center gap-1">
              <Hash size={12} />
              {message.metrics.totalTokens} tokens
            </span>
            {message.metrics.tokensPerSecond != null && (
              <span className="flex items-center gap-1">
                <Zap size={12} />
                {message.metrics.tokensPerSecond.toFixed(1)} t/s
              </span>
            )}
            {message.metrics.timeToFirstToken != null && (
              <span className="flex items-center gap-1">
                <Clock size={12} />
                TTFT: {message.metrics.timeToFirstToken.toFixed(0)}ms
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export function EnhancedChatPanel() {
  const [input, setInput] = useState('')
  const [attachments, setAttachments] = useState<File[]>([])
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const {
    messages,
    metrics,
    connectionStatus,
    currentConversationId,
    addToQueue,
    rollbackToMessage,
    messageQueue,
    endGeneration,
  } = useAppStore()

  const isGenerating = metrics.isGenerating

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`
    }
  }, [input])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || !currentConversationId) return

    const userMessage = input.trim()
    setInput('')

    // If generating, queue the message instead of sending immediately
    if (isGenerating) {
      addToQueue(userMessage)
      return
    }

    // Send message via WebSocket
    wsService.sendChatMessage(currentConversationId, userMessage)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleStop = () => {
    if (currentConversationId) {
      wsService.stopGeneration(currentConversationId)
    }
    endGeneration()
  }

  const handleStopAndRollback = (messageId: string) => {
    // Stop generation first
    handleStop()
    // Then rollback to the message
    rollbackToMessage(messageId)
  }

  const isConnected = connectionStatus === 'connected'

  const handleAttachClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    setAttachments((prev) => [...prev, ...files])
    e.target.value = '' // Reset for same file selection
  }

  const removeAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index))
  }

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto">
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center px-4">
            <div className="w-16 h-16 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-editor-accent/20 overflow-hidden">
              <img src="/logo.png" alt="Prism Logo" className="w-full h-full object-contain" />
            </div>
            <h1 className="text-2xl font-bold text-editor-text mb-2">Welcome to Prism</h1>
            <p className="text-editor-muted max-w-md mb-8">
              Your AI-powered code assistant. Ask questions, generate code, or get help with your projects.
            </p>
            <div className="grid grid-cols-2 gap-3 max-w-md">
              {[
                { icon: '💡', text: 'Explain TypeScript generics' },
                { icon: '🔧', text: 'Help fix a React bug' },
                { icon: '📝', text: 'Write a file tree component' },
                { icon: '🚀', text: 'Optimize my code' },
              ].map((suggestion, i) => (
                <button
                  key={i}
                  onClick={() => setInput(suggestion.text)}
                  className="flex items-center gap-2 px-4 py-3 rounded-lg border border-editor-border bg-editor-surface/50 hover:bg-editor-surface text-editor-text text-sm transition-colors hover:border-editor-accent/50"
                >
                  <span>{suggestion.icon}</span>
                  <span>{suggestion.text}</span>
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="divide-y divide-editor-border/50">
            {messages.map(message => (
              <MessageBubble
                key={message.id}
                message={message}
                onRollback={rollbackToMessage}
                isGenerating={isGenerating}
                onStopAndRollback={handleStopAndRollback}
              />
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Generation status bar */}
      {isGenerating && (
        <div className="px-4 py-2 bg-editor-surface/50 border-t border-editor-border flex items-center justify-between">
          <div className="flex items-center gap-3 text-sm">
            <span className="flex items-center gap-2 text-editor-accent">
              <span className="w-2 h-2 rounded-full bg-editor-accent animate-pulse" />
              Generating...
            </span>
            <span className="text-editor-muted">{metrics.tokenCount} tokens</span>
            <span className="text-editor-muted">{metrics.tokensPerSecond.toFixed(1)} t/s</span>
            {metrics.timeToFirstToken !== null && (
              <span className="text-editor-muted">TTFT: {metrics.timeToFirstToken.toFixed(0)}ms</span>
            )}
          </div>
          <button
            onClick={handleStop}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-editor-error/20 text-editor-error hover:bg-editor-error/30 transition-colors text-sm"
          >
            <StopCircle size={14} />
            Stop
          </button>
        </div>
      )}

      {/* Message Queue */}
      <MessageQueue />

      {/* Input */}
      <div className="p-4 border-t border-editor-border">
        <form onSubmit={handleSubmit} className="max-w-3xl mx-auto">
          {/* Attachment Preview */}
          {attachments.length > 0 && (
            <div className="flex gap-2 mb-2 flex-wrap">
              {attachments.map((file, i) => (
                <div
                  key={`${file.name}-${i}`}
                  className="flex items-center gap-1 bg-editor-surface px-2 py-1 rounded-lg text-sm border border-editor-border"
                >
                  <Paperclip size={12} className="text-editor-muted" />
                  <span className="truncate max-w-[150px] text-editor-text">{file.name}</span>
                  <button
                    type="button"
                    onClick={() => removeAttachment(i)}
                    className="p-0.5 hover:bg-editor-error/20 rounded text-editor-muted hover:text-editor-error transition-colors"
                  >
                    <X size={12} />
                  </button>
                </div>
              ))}
            </div>
          )}

          <div className="relative flex items-end gap-2 bg-editor-surface rounded-xl border border-editor-border focus-within:border-editor-accent transition-colors">
            {/* Hidden file input */}
            <input
              ref={fileInputRef}
              type="file"
              multiple
              className="hidden"
              onChange={handleFileSelect}
              accept="image/*,.pdf,.txt,.md,.js,.ts,.jsx,.tsx,.json,.html,.css,.py,.go,.rs"
            />

            <button
              type="button"
              onClick={handleAttachClick}
              className={`p-3 transition-colors relative ${
                attachments.length > 0
                  ? 'text-editor-accent'
                  : 'text-editor-muted hover:text-editor-text'
              }`}
              title="Attach file"
            >
              <Paperclip size={20} />
              {attachments.length > 0 && (
                <span className="absolute -top-1 -right-1 bg-editor-accent text-white text-xs w-4 h-4 rounded-full flex items-center justify-center">
                  {attachments.length}
                </span>
              )}
            </button>

            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={!isConnected ? "Connecting..." : !currentConversationId ? "Select or create a conversation..." : isGenerating ? "Type to queue message..." : "Send a message..."}
              disabled={!isConnected || !currentConversationId}
              rows={1}
              className="flex-1 bg-transparent text-editor-text placeholder-editor-muted py-3 pr-2 resize-none focus:outline-none max-h-48 disabled:opacity-50"
            />

            <button
              type="submit"
              disabled={!input.trim() || !isConnected || !currentConversationId}
              className="m-2 p-2.5 rounded-lg bg-editor-accent text-white hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              title={isGenerating ? "Queue message" : "Send message"}
            >
              <Send size={20} />
            </button>
          </div>

          <div className="flex items-center justify-between mt-2 px-1">
            <div className="flex items-center gap-2 text-xs text-editor-muted">
              <span className={`flex items-center gap-1 ${isConnected ? 'text-editor-success' : 'text-editor-error'}`}>
                <span className={`w-1.5 h-1.5 rounded-full ${isConnected ? 'bg-editor-success' : 'bg-editor-error'}`} />
                {connectionStatus}
              </span>
              <span className="text-editor-border">|</span>
              <span>{messages.length} messages</span>
              {messageQueue.length > 0 && (
                <>
                  <span className="text-editor-border">|</span>
                  <span className="text-editor-accent">{messageQueue.length} queued</span>
                </>
              )}
            </div>
            <span className="text-xs text-editor-muted">Shift + Enter for new line</span>
          </div>
        </form>
      </div>
    </div>
  )
}
