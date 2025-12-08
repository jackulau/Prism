import { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, Loader2, StopCircle } from 'lucide-react'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  isStreaming?: boolean
}

export function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'assistant',
      content: 'Hello! I\'m Prism, your AI coding assistant. I can help you build web applications, write code, and preview your work in the sandbox. What would you like to create today?',
      timestamp: new Date(),
    }
  ])
  const [input, setInput] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isGenerating) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: new Date(),
    }

    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsGenerating(true)

    // Simulate assistant response (replace with actual WebSocket integration)
    setTimeout(() => {
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'I received your message. The chat functionality is being connected to the backend. Soon I\'ll be able to help you write code and display it in the sandbox preview!',
        timestamp: new Date(),
      }
      setMessages(prev => [...prev, assistantMessage])
      setIsGenerating(false)
    }, 1000)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleStop = () => {
    setIsGenerating(false)
  }

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Header */}
      <div className="px-4 py-3 border-b border-editor-border">
        <h2 className="text-lg font-semibold text-editor-text">Chat</h2>
        <p className="text-sm text-editor-muted">Ask me to build anything</p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
        {isGenerating && (
          <div className="flex items-center gap-2 text-editor-muted">
            <Loader2 size={16} className="animate-spin" />
            <span className="text-sm">Generating...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t border-editor-border">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <div className="flex-1 relative">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Type a message... (Enter to send, Shift+Enter for new line)"
              className="w-full px-4 py-3 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted resize-none focus:outline-none focus:border-editor-accent transition-colors"
              rows={1}
              disabled={isGenerating}
            />
          </div>
          {isGenerating ? (
            <button
              type="button"
              onClick={handleStop}
              className="px-4 py-2 bg-editor-error text-white rounded-lg hover:bg-editor-error/80 transition-colors"
            >
              <StopCircle size={20} />
            </button>
          ) : (
            <button
              type="submit"
              disabled={!input.trim()}
              className="px-4 py-2 bg-editor-accent text-editor-bg rounded-lg hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Send size={20} />
            </button>
          )}
        </form>
      </div>
    </div>
  )
}

interface MessageBubbleProps {
  message: Message
}

function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex gap-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div
        className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
          isUser ? 'bg-editor-accent/20' : 'bg-editor-surface'
        }`}
      >
        {isUser ? (
          <User size={16} className="text-editor-accent" />
        ) : (
          <Bot size={16} className="text-editor-muted" />
        )}
      </div>
      <div
        className={`max-w-[80%] px-4 py-3 rounded-lg ${
          isUser
            ? 'bg-editor-accent text-editor-bg'
            : 'bg-editor-surface text-editor-text'
        } ${message.isStreaming ? 'streaming-text' : ''}`}
      >
        <p className="text-sm whitespace-pre-wrap">{message.content}</p>
        <span
          className={`text-xs mt-1 block ${
            isUser ? 'text-editor-bg/70' : 'text-editor-muted'
          }`}
        >
          {message.timestamp.toLocaleTimeString()}
        </span>
      </div>
    </div>
  )
}
