import React, { useRef, useEffect } from 'react';
import {
  Send,
  Square,
  Paperclip,
  Mic,
} from 'lucide-react';
import { useAppStore } from '../store';
import { ChatMessage } from './ChatMessage';
import { wsService } from '../services/websocket';

export const ChatInterface: React.FC = () => {
  const {
    messages,
    inputValue,
    setInputValue,
    streamingMessageId,
    connectionStatus,
    currentConversationId,
    metrics,
  } = useAppStore();

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, messages[messages.length - 1]?.content]);

  // Auto-resize textarea
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
      inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, 200)}px`;
    }
  }, [inputValue]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim() || streamingMessageId) return;

    const conversationId = currentConversationId || `conv-${Date.now()}`;
    wsService.sendChatMessage(conversationId, inputValue.trim());
    setInputValue('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const handleStop = () => {
    if (currentConversationId) {
      wsService.stopGeneration(currentConversationId);
    }
  };

  const isConnected = connectionStatus === 'connected';
  const isGenerating = streamingMessageId !== null;

  return (
    <div className="flex flex-col h-full bg-editor-bg">
      {/* Messages area */}
      <div className="flex-1 overflow-y-auto">
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center px-4">
            <div className="w-16 h-16 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-editor-accent/20 overflow-hidden">
              <img src="/logo.png" alt="Prism Logo" className="w-full h-full object-contain" />
            </div>
            <h1 className="text-2xl font-bold text-editor-text mb-2">
              Welcome to Prism
            </h1>
            <p className="text-editor-muted max-w-md mb-8">
              Your AI-powered code assistant. Ask questions, generate code, or get help with your projects.
            </p>
            <div className="grid grid-cols-2 gap-3 max-w-lg">
              {[
                { icon: '💡', text: 'Explain this code' },
                { icon: '🔧', text: 'Fix this bug' },
                { icon: '📝', text: 'Write a function' },
                { icon: '🚀', text: 'Optimize performance' },
              ].map((suggestion, i) => (
                <button
                  key={i}
                  onClick={() => setInputValue(suggestion.text)}
                  className="flex items-center gap-2 px-4 py-3 rounded-lg border border-editor-border bg-editor-surface/50 hover:bg-editor-surface text-editor-text text-sm transition-smooth hover:border-editor-accent/50"
                >
                  <span>{suggestion.icon}</span>
                  <span>{suggestion.text}</span>
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="divide-y divide-editor-border/50">
            {messages.map((message) => (
              <ChatMessage key={message.id} message={message} />
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
              Generating response...
            </span>
            <span className="text-editor-muted">
              {metrics.tokenCount} tokens
            </span>
            <span className="text-editor-muted">
              {metrics.tokensPerSecond.toFixed(1)} t/s
            </span>
          </div>
          <button
            onClick={handleStop}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-editor-error/20 text-editor-error hover:bg-editor-error/30 transition-smooth text-sm"
          >
            <Square className="w-3.5 h-3.5" />
            Stop
          </button>
        </div>
      )}

      {/* Input area */}
      <div className="border-t border-editor-border p-4">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
          <div className="relative flex items-end gap-2 bg-editor-surface rounded-xl border border-editor-border focus-within:border-editor-accent transition-smooth">
            {/* Attachment button */}
            <button
              type="button"
              className="p-3 text-editor-muted hover:text-editor-text transition-smooth"
              title="Attach file"
            >
              <Paperclip className="w-5 h-5" />
            </button>

            {/* Input */}
            <textarea
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={isConnected ? "Send a message..." : "Connecting..."}
              disabled={!isConnected}
              rows={1}
              className="flex-1 bg-transparent text-editor-text placeholder-editor-muted py-3 pr-2 resize-none focus:outline-none max-h-48 disabled:opacity-50"
            />

            {/* Voice input button */}
            <button
              type="button"
              className="p-3 text-editor-muted hover:text-editor-text transition-smooth"
              title="Voice input"
            >
              <Mic className="w-5 h-5" />
            </button>

            {/* Send button */}
            <button
              type="submit"
              disabled={!inputValue.trim() || !isConnected || isGenerating}
              className="m-2 p-2.5 rounded-lg bg-editor-accent text-white hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-smooth"
              title="Send message"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>

          {/* Input footer */}
          <div className="flex items-center justify-between mt-2 px-1">
            <div className="flex items-center gap-2 text-xs text-editor-muted">
              <span
                className={`flex items-center gap-1 ${
                  isConnected ? 'text-editor-success' : 'text-editor-error'
                }`}
              >
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    isConnected ? 'bg-editor-success' : 'bg-editor-error'
                  }`}
                />
                {connectionStatus}
              </span>
              <span className="text-editor-border">|</span>
              <span>{messages.length} messages</span>
            </div>
            <div className="flex items-center gap-2 text-xs text-editor-muted">
              <span>Shift + Enter for new line</span>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};
