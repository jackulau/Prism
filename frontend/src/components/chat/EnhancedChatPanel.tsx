import { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, StopCircle, Paperclip, Hash, Zap, Clock, RotateCcw, X, Trash2, Plus, HelpCircle, Cpu, Download, Plug, MessageSquare, Copy, Check, Pencil, RefreshCw } from 'lucide-react'
import { useAppStore } from '../../store'
import { wsService } from '../../services/websocket'
import { MessageQueue } from './MessageQueue'
import { ModelSelector } from '../ModelSelector'
import { ModeSwitcher } from './ModeSwitcher'
import { ThinkingToggle } from './ThinkingToggle'
import { MarkdownRenderer } from '../markdown/MarkdownRenderer'
import { ToolCallCard } from './ToolCallCard'
import { toast } from '../../store/toastStore'
import type { Message, ToolCall } from '../../types'

// Command definition
interface Command {
  name: string
  description: string
  icon: React.ReactNode
  hasArgs?: boolean
}

// Available commands
const COMMANDS: Command[] = [
  { name: '/clear', description: 'Clear message context', icon: <Trash2 size={14} /> },
  { name: '/new', description: 'Start new conversation', icon: <Plus size={14} /> },
  { name: '/help', description: 'Show available commands', icon: <HelpCircle size={14} /> },
  { name: '/model', description: 'Switch model', icon: <Cpu size={14} />, hasArgs: true },
  { name: '/export', description: 'Export conversation', icon: <Download size={14} /> },
  { name: '/mcp', description: 'MCP tools and connections', icon: <Plug size={14} />, hasArgs: true },
  { name: '/conversations', description: 'List and manage conversations', icon: <MessageSquare size={14} />, hasArgs: true },
]


// Message bubble component
function MessageBubble({
  message,
  onRollback,
  isGenerating,
  onStopAndRollback,
  onCopy,
  onEdit,
  onRegenerate,
  onToolApprove,
  onToolReject,
}: {
  message: Message;
  onRollback?: (id: string) => void;
  isGenerating?: boolean;
  onStopAndRollback?: (id: string) => void;
  onCopy?: (content: string) => void;
  onEdit?: (id: string, content: string) => void;
  onRegenerate?: (id: string) => void;
  onToolApprove?: (executionId: string) => void;
  onToolReject?: (executionId: string) => void;
}) {
  const [copied, setCopied] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState(message.content)
  const isUser = message.role === 'user'

  const handleRollback = () => {
    if (isGenerating && onStopAndRollback) {
      onStopAndRollback(message.id)
    } else if (onRollback) {
      onRollback(message.id)
    }
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(message.content)
    setCopied(true)
    onCopy?.(message.content)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleSaveEdit = () => {
    if (editContent.trim() !== message.content) {
      onEdit?.(message.id, editContent.trim())
    }
    setIsEditing(false)
  }

  const handleCancelEdit = () => {
    setEditContent(message.content)
    setIsEditing(false)
  }

  // Filter out tool call text markers from content for cleaner display
  // The tool calls are shown separately via ToolCallCard
  const cleanContent = message.toolCalls && message.toolCalls.length > 0
    ? message.content.replace(/\n\n\*\*Using tool:\*\* `[^`]+`\n/g, '')
    : message.content

  return (
    <div className={`py-4 px-4 ${isUser ? 'bg-editor-surface/30' : ''} group relative`}>
      <div className="max-w-3xl mx-auto">
        {/* Header row */}
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

          {/* Action buttons - shown on hover */}
          {!message.isStreaming && (
            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              {/* Copy button */}
              <button
                onClick={handleCopy}
                className="p-1.5 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
                title="Copy message"
              >
                {copied ? <Check size={14} className="text-editor-success" /> : <Copy size={14} />}
              </button>

              {/* Edit button for user messages */}
              {isUser && onEdit && (
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-1.5 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
                  title="Edit message"
                >
                  <Pencil size={14} />
                </button>
              )}

              {/* Regenerate button for assistant messages */}
              {!isUser && onRegenerate && (
                <button
                  onClick={() => onRegenerate(message.id)}
                  className="p-1.5 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
                  title="Regenerate response"
                >
                  <RefreshCw size={14} />
                </button>
              )}

              {/* Rollback button for user messages */}
              {isUser && (onRollback || onStopAndRollback) && (
                <button
                  onClick={handleRollback}
                  className="p-1.5 rounded-lg hover:bg-editor-warning/20 text-editor-muted hover:text-editor-warning transition-colors"
                  title={isGenerating ? "Stop and rollback" : "Rollback to this message"}
                >
                  <RotateCcw size={14} />
                </button>
              )}
            </div>
          )}
        </div>

        {/* Message content */}
        <div className="pl-11">
          {isEditing ? (
            <div className="space-y-2">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className="w-full p-3 bg-editor-surface border border-editor-border rounded-lg text-editor-text resize-none focus:outline-none focus:border-editor-accent"
                rows={Math.min(10, editContent.split('\n').length + 1)}
                autoFocus
              />
              <div className="flex gap-2">
                <button
                  onClick={handleSaveEdit}
                  className="px-3 py-1.5 bg-editor-accent text-white rounded-lg text-sm hover:bg-editor-accent/80 transition-colors"
                >
                  Save & Regenerate
                </button>
                <button
                  onClick={handleCancelEdit}
                  className="px-3 py-1.5 bg-editor-surface text-editor-text rounded-lg text-sm hover:bg-editor-border transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div className="text-editor-text leading-relaxed">
              <MarkdownRenderer content={cleanContent} isStreaming={message.isStreaming} />
            </div>
          )}

          {/* Tool calls display */}
          {message.toolCalls && message.toolCalls.length > 0 && (
            <div className="mt-2 space-y-2">
              {message.toolCalls.map((toolCall: ToolCall) => (
                <ToolCallCard
                  key={toolCall.id}
                  toolCall={toolCall}
                  isMCP={toolCall.isMCP}
                  serverName={toolCall.serverName}
                  onApprove={toolCall.status === 'pending' ? onToolApprove : undefined}
                  onReject={toolCall.status === 'pending' ? onToolReject : undefined}
                />
              ))}
            </div>
          )}
        </div>

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
  const [showCommands, setShowCommands] = useState(false)
  const [filteredCommands, setFilteredCommands] = useState<Command[]>([])
  const [selectedCommandIndex, setSelectedCommandIndex] = useState(0)
  const [inputHistory, setInputHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [tempInput, setTempInput] = useState('') // Store current input when navigating history
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
    clearMessages,
    createNewConversation,
    providers,
    setSelectedProvider,
    setSelectedModel,
    addMessage,
    // Chat options
    chatMode,
    extendedThinkingEnabled,
    selectedFile,
    fileContextEnabled,
  } = useAppStore()

  const isGenerating = metrics.isGenerating
  const isConnected = connectionStatus === 'connected'

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

  // Filter commands when input starts with /
  useEffect(() => {
    if (input.startsWith('/')) {
      const search = input.toLowerCase()
      const matches = COMMANDS.filter(cmd => cmd.name.toLowerCase().startsWith(search))
      setFilteredCommands(matches)
      setShowCommands(matches.length > 0)
      setSelectedCommandIndex(0)
    } else {
      setShowCommands(false)
      setFilteredCommands([])
    }
  }, [input])

  // Execute a command
  const executeCommand = async (command: string) => {
    const parts = command.trim().split(/\s+/)
    const cmd = parts[0].toLowerCase()
    const args = parts.slice(1).join(' ')

    switch (cmd) {
      case '/clear':
        clearMessages()
        toast.success('Context cleared')
        break

      case '/new':
        await createNewConversation()
        toast.success('New conversation started')
        break

      case '/help':
        // Show help as a system message
        const helpText = COMMANDS.map(c => `**${c.name}** - ${c.description}`).join('\n')
        addMessage({
          id: `help-${Date.now()}`,
          role: 'assistant',
          content: `## Available Commands\n\n${helpText}\n\n*Type a command and press Enter to execute*`,
          timestamp: new Date(),
        })
        break

      case '/model':
        if (args) {
          // Find matching model
          for (const provider of providers) {
            const model = provider.models.find(m =>
              m.id.toLowerCase().includes(args.toLowerCase()) ||
              m.name.toLowerCase().includes(args.toLowerCase())
            )
            if (model) {
              setSelectedProvider(provider.name)
              setSelectedModel(model.id)
              toast.success(`Switched to ${provider.name} / ${model.name}`)
              return
            }
          }
          toast.error(`Model "${args}" not found`)
        } else {
          // Show available models
          const modelList = providers.flatMap(p =>
            p.models.map(m => `- **${p.name}**: ${m.name}`)
          ).join('\n')
          addMessage({
            id: `models-${Date.now()}`,
            role: 'assistant',
            content: `## Available Models\n\n${modelList || 'No models available'}\n\n*Usage: /model <name>*`,
            timestamp: new Date(),
          })
        }
        break

      case '/export':
        if (messages.length === 0) {
          toast.error('No messages to export')
          return
        }
        const exportContent = messages.map(m =>
          `## ${m.role === 'user' ? 'You' : 'Assistant'}\n${m.content}`
        ).join('\n\n---\n\n')
        const blob = new Blob([exportContent], { type: 'text/markdown' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `conversation-${new Date().toISOString().split('T')[0]}.md`
        a.click()
        URL.revokeObjectURL(url)
        toast.success('Conversation exported')
        break

      default:
        toast.error(`Unknown command: ${cmd}`)
    }
  }

  // Helper function to convert file to base64
  const fileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.readAsDataURL(file)
      reader.onload = () => {
        const result = reader.result as string
        // Remove the data URL prefix (e.g., "data:image/png;base64,")
        const base64 = result.split(',')[1]
        resolve(base64)
      }
      reader.onerror = (error) => reject(error)
    })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() && attachments.length === 0) return

    const userMessage = input.trim()

    // Add to input history (avoid duplicates of last entry)
    if (userMessage && (inputHistory.length === 0 || inputHistory[inputHistory.length - 1] !== userMessage)) {
      setInputHistory(prev => [...prev, userMessage])
    }
    setHistoryIndex(-1)
    setTempInput('')

    setInput('')
    setShowCommands(false)

    // Handle commands
    if (userMessage.startsWith('/')) {
      setAttachments([]) // Clear attachments if command
      await executeCommand(userMessage)
      return
    }

    // Process attachments to base64
    let attachmentData: Array<{ name: string; type: string; data: string }> = []
    if (attachments.length > 0) {
      try {
        attachmentData = await Promise.all(
          attachments.map(async (file) => ({
            name: file.name,
            type: file.type,
            data: await fileToBase64(file),
          }))
        )
      } catch {
        toast.error('Failed to process attachments')
        return
      }
    }
    setAttachments([]) // Clear attachments after processing

    // If no conversation, create one first
    if (!currentConversationId) {
      const newConvId = await createNewConversation()
      if (!newConvId) {
        toast.error('No model selected - please select a model first')
        return
      }
      // Queue the message to be sent after conversation is created
      addToQueue(userMessage)
      return
    }

    // If not connected or generating, queue the message
    if (!isConnected || isGenerating) {
      addToQueue(userMessage)
      if (!isConnected) {
        toast.info('Message queued - will send when connected')
      }
      return
    }

    // Build file context if enabled and file is selected
    let fileContext = null
    if (fileContextEnabled && selectedFile) {
      // Get file content from sandbox store if available
      const sandboxStore = (await import('../../store/sandboxStore')).useSandboxStore.getState()
      const content = sandboxStore.fileContents[selectedFile.path] || ''
      if (content) {
        fileContext = {
          path: selectedFile.path,
          content,
          language: selectedFile.language,
        }
      }
    }

    // Send message via WebSocket with attachments and options
    wsService.sendChatMessage(
      currentConversationId,
      userMessage,
      attachmentData.length > 0 ? attachmentData : undefined,
      {
        mode: chatMode,
        extendedThinking: extendedThinkingEnabled,
        fileContext,
      }
    )
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Handle command autocomplete navigation
    if (showCommands && filteredCommands.length > 0) {
      if (e.key === 'Tab') {
        e.preventDefault()
        const cmd = filteredCommands[selectedCommandIndex]
        setInput(cmd.name + (cmd.hasArgs ? ' ' : ''))
        setShowCommands(false)
        return
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedCommandIndex(i => (i + 1) % filteredCommands.length)
        return
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedCommandIndex(i => i === 0 ? filteredCommands.length - 1 : i - 1)
        return
      }
      if (e.key === 'Escape') {
        e.preventDefault()
        setShowCommands(false)
        return
      }
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        const cmd = filteredCommands[selectedCommandIndex]
        if (cmd.hasArgs) {
          // Commands with args: just autocomplete (let user type args)
          setInput(cmd.name + ' ')
          setShowCommands(false)
        } else {
          // Commands without args: execute immediately
          setInput('')
          setShowCommands(false)
          executeCommand(cmd.name)
        }
        return
      }
    }

    // Handle input history navigation (when command dropdown is not visible)
    if (e.key === 'ArrowUp' && inputHistory.length > 0) {
      // Only navigate history if cursor is at start or input is empty
      const textarea = textareaRef.current
      if (textarea && (textarea.selectionStart === 0 || input === '')) {
        e.preventDefault()
        if (historyIndex === -1) {
          // Save current input before navigating
          setTempInput(input)
          setHistoryIndex(inputHistory.length - 1)
          setInput(inputHistory[inputHistory.length - 1])
        } else if (historyIndex > 0) {
          setHistoryIndex(historyIndex - 1)
          setInput(inputHistory[historyIndex - 1])
        }
        return
      }
    }

    if (e.key === 'ArrowDown' && historyIndex !== -1) {
      e.preventDefault()
      if (historyIndex < inputHistory.length - 1) {
        setHistoryIndex(historyIndex + 1)
        setInput(inputHistory[historyIndex + 1])
      } else {
        // Return to current input
        setHistoryIndex(-1)
        setInput(tempInput)
      }
      return
    }

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

  const handleCopy = (_content: string) => {
    toast.success('Copied to clipboard')
  }

  const handleEdit = (messageId: string, newContent: string) => {
    // Find the message and rollback to it with edited content
    const messageIndex = messages.findIndex(m => m.id === messageId)
    if (messageIndex === -1) return

    // Rollback to before this message
    rollbackToMessage(messageId)

    // Send the edited message as a new message
    if (currentConversationId && isConnected) {
      wsService.sendChatMessage(currentConversationId, newContent, undefined, {
        mode: chatMode,
        extendedThinking: extendedThinkingEnabled,
      })
    }
  }

  const handleRegenerate = (messageId: string) => {
    // Find the user message before this assistant message
    const messageIndex = messages.findIndex(m => m.id === messageId)
    if (messageIndex <= 0) return

    // Find the last user message before this
    let userMessageIndex = messageIndex - 1
    while (userMessageIndex >= 0 && messages[userMessageIndex].role !== 'user') {
      userMessageIndex--
    }

    if (userMessageIndex < 0) return

    const userMessage = messages[userMessageIndex]

    // Rollback to the user message
    rollbackToMessage(userMessage.id)

    // Resend the user message
    if (currentConversationId && isConnected) {
      wsService.sendChatMessage(currentConversationId, userMessage.content, undefined, {
        mode: chatMode,
        extendedThinking: extendedThinkingEnabled,
      })
    }
  }

  const handleToolApprove = (executionId: string) => {
    if (!currentConversationId) return
    wsService.send({
      type: 'tool.confirm' as never,
      conversation_id: currentConversationId,
      execution_id: executionId,
      approved: true,
    })
    toast.success('Tool approved')
  }

  const handleToolReject = (executionId: string) => {
    if (!currentConversationId) return
    wsService.send({
      type: 'tool.confirm' as never,
      conversation_id: currentConversationId,
      execution_id: executionId,
      approved: false,
    })
    toast.info('Tool rejected')
  }

  const handleAttachClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB per file
    const MAX_TOTAL_SIZE = 25 * 1024 * 1024 // 25MB total

    const files = Array.from(e.target.files || [])
    const validFiles: File[] = []
    let totalSize = attachments.reduce((sum, f) => sum + f.size, 0)

    for (const file of files) {
      if (file.size > MAX_FILE_SIZE) {
        console.error(`File "${file.name}" exceeds 10MB limit`)
        continue
      }
      if (totalSize + file.size > MAX_TOTAL_SIZE) {
        console.error('Total attachment size would exceed 25MB limit')
        break
      }
      totalSize += file.size
      validFiles.push(file)
    }

    if (validFiles.length > 0) {
      setAttachments((prev) => [...prev, ...validFiles])
    }
    e.target.value = '' // Reset for same file selection
  }

  const removeAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index))
  }

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Chat Header with Model Selector */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/30">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium text-editor-text">Model</span>
          <ModelSelector />
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              clearMessages()
              toast.success('Context cleared')
            }}
            className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors"
            title="Clear context (/clear)"
          >
            <Trash2 size={16} />
          </button>
        </div>
      </div>

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
                onCopy={handleCopy}
                onEdit={handleEdit}
                onRegenerate={handleRegenerate}
                onToolApprove={handleToolApprove}
                onToolReject={handleToolReject}
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
        <form onSubmit={handleSubmit} className="max-w-3xl mx-auto relative">
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

          {/* Command Autocomplete Dropdown */}
          {showCommands && filteredCommands.length > 0 && (
            <div className="absolute bottom-full left-0 right-0 mb-2 bg-editor-surface border border-editor-border rounded-lg shadow-lg overflow-hidden z-10">
              <div className="px-3 py-2 border-b border-editor-border bg-editor-bg/50">
                <span className="text-xs text-editor-muted">Commands</span>
                <span className="text-xs text-editor-muted/50 ml-2">Enter or Tab to select, ↑↓ to navigate</span>
              </div>
              {filteredCommands.map((cmd, i) => (
                <button
                  key={cmd.name}
                  type="button"
                  className={`w-full px-4 py-2.5 text-left flex items-center gap-3 transition-colors ${
                    i === selectedCommandIndex
                      ? 'bg-editor-accent/20 text-editor-accent'
                      : 'hover:bg-editor-surface text-editor-text'
                  }`}
                  onClick={() => {
                    if (cmd.hasArgs) {
                      setInput(cmd.name + ' ')
                      setShowCommands(false)
                      textareaRef.current?.focus()
                    } else {
                      setInput('')
                      setShowCommands(false)
                      executeCommand(cmd.name)
                    }
                  }}
                >
                  <span className={`${i === selectedCommandIndex ? 'text-editor-accent' : 'text-editor-muted'}`}>
                    {cmd.icon}
                  </span>
                  <span className="font-mono text-sm">{cmd.name}</span>
                  <span className="text-editor-muted text-sm">{cmd.description}</span>
                </button>
              ))}
            </div>
          )}

          <div className="relative bg-editor-surface rounded-xl border border-editor-border focus-within:border-editor-accent transition-colors">
            {/* Hidden file input */}
            <input
              ref={fileInputRef}
              type="file"
              multiple
              className="hidden"
              onChange={handleFileSelect}
              accept="image/*,.pdf,.txt,.md,.js,.ts,.jsx,.tsx,.json,.html,.css,.py,.go,.rs"
            />

            {/* Textarea row */}
            <div className="flex items-end">
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
                placeholder={
                  !isConnected
                    ? "Type message (will queue until connected)..."
                    : isGenerating
                      ? "Type to queue message..."
                      : "Send a message... (type / for commands)"
                }
                rows={1}
                className="flex-1 bg-transparent text-editor-text placeholder-editor-muted py-3 pr-3 resize-none focus:outline-none max-h-48"
              />
            </div>

            {/* Context bar with controls and send button */}
            <div className="flex items-center gap-2 px-3 py-2 border-t border-editor-border/50">
              {/* File context badge */}
              {fileContextEnabled && selectedFile && (
                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-editor-bg border border-editor-border text-xs">
                  <span className="text-editor-accent">📄</span>
                  <span className="text-editor-text max-w-[100px] truncate" title={selectedFile.path}>
                    {selectedFile.name}
                  </span>
                  <button
                    type="button"
                    onClick={() => useAppStore.getState().setFileContextEnabled(false)}
                    className="text-editor-muted hover:text-editor-text transition-colors"
                    title="Remove file context"
                  >
                    <X size={12} />
                  </button>
                </div>
              )}

              <div className="flex-1" />

              {/* Mode switcher - inline */}
              <ModeSwitcher />

              {/* Thinking toggle - inline */}
              <ThinkingToggle />

              {/* Send button */}
              <button
                type="submit"
                disabled={!input.trim()}
                className="p-2 rounded-lg bg-editor-accent text-white hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                title={!isConnected ? "Queue message (not connected)" : isGenerating ? "Queue message" : "Send message"}
              >
                <Send size={18} />
              </button>
            </div>
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
