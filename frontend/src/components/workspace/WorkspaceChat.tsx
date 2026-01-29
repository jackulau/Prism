import { useState, useRef, useEffect } from 'react';
import { Send, StopCircle, Paperclip, X } from 'lucide-react';
import { useAppStore } from '../../store';
import { useWorkspaceStore } from '../../store/workspaceStore';
import { wsService } from '../../services/websocket';
import { ContextBar } from './ContextBar';
import { WorkspaceMessage } from './WorkspaceMessage';
import { ModeSwitcher } from '../chat/ModeSwitcher';
import { ThinkingToggle } from '../chat/ThinkingToggle';
import { toast } from '../../store/toastStore';
import type { FileContext, EditSuggestion } from '../../types/workspace';

interface WorkspaceChatProps {
  onFileClick?: (path: string, line?: number) => void;
  onAddContext?: () => void;
}

export function WorkspaceChat({ onFileClick, onAddContext }: WorkspaceChatProps) {
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const {
    messages,
    metrics,
    connectionStatus,
    currentConversationId,
    addToQueue,
    endGeneration,
    createNewConversation,
    chatMode,
    extendedThinkingEnabled,
  } = useAppStore();

  const {
    contextFiles,
    removeContextFile,
    clearContextFiles,
    addEditSuggestion,
    acceptEdit,
    rejectEdit,
  } = useWorkspaceStore();

  const isGenerating = metrics.isGenerating;
  const isConnected = connectionStatus === 'connected';

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [input]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMessage = input.trim();
    setInput('');

    // If no conversation, create one first
    if (!currentConversationId) {
      const newConvId = await createNewConversation();
      if (!newConvId) {
        toast.error('No model selected - please select a model first');
        return;
      }
      addToQueue(userMessage);
      return;
    }

    // If not connected or generating, queue the message
    if (!isConnected || isGenerating) {
      addToQueue(userMessage);
      if (!isConnected) {
        toast.info('Message queued - will send when connected');
      }
      return;
    }

    // Build file context from context files
    const fileContextData = contextFiles.length > 0 ? contextFiles[0] : null;

    // Send message via WebSocket with context
    wsService.sendChatMessage(
      currentConversationId,
      userMessage,
      undefined,
      {
        mode: chatMode,
        extendedThinking: extendedThinkingEnabled,
        fileContext: fileContextData ? {
          path: fileContextData.path,
          content: fileContextData.content,
          language: fileContextData.language,
        } : undefined,
      }
    );
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
    endGeneration();
  };

  const handleFileClick = (path: string, line?: number) => {
    onFileClick?.(path, line);
  };

  const handleApplyEdit = (edit: EditSuggestion) => {
    addEditSuggestion(edit);
    acceptEdit(edit.id);
    toast.success('Changes applied');
  };

  const handleRejectEdit = (editId: string) => {
    rejectEdit(editId);
    toast.info('Changes rejected');
  };

  const handleContextFileClick = (file: FileContext) => {
    onFileClick?.(file.path, file.selection?.startLine);
  };

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Context Bar */}
      <ContextBar
        files={contextFiles}
        onRemove={removeContextFile}
        onAdd={onAddContext}
        onFileClick={handleContextFileClick}
      />

      {/* Messages */}
      <div className="flex-1 overflow-y-auto">
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center px-4">
            <div className="w-16 h-16 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-editor-accent/20 overflow-hidden">
              <img src="/logo.png" alt="Prism Logo" className="w-full h-full object-contain" />
            </div>
            <h1 className="text-2xl font-bold text-editor-text mb-2">Workspace Chat</h1>
            <p className="text-editor-muted max-w-md mb-4">
              Ask questions about your code with file context. Add files from the tree to include them in your questions.
            </p>
            {contextFiles.length === 0 && (
              <p className="text-sm text-editor-muted/70">
                Click a file in the explorer to add it as context
              </p>
            )}
          </div>
        ) : (
          <div className="divide-y divide-editor-border/50">
            {messages.map((message) => (
              <WorkspaceMessage
                key={message.id}
                message={message}
                onFileClick={handleFileClick}
                onApplyEdit={handleApplyEdit}
                onRejectEdit={handleRejectEdit}
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

      {/* Input */}
      <div className="p-4 border-t border-editor-border">
        <form onSubmit={handleSubmit} className="relative">
          <div className="relative bg-editor-surface rounded-xl border border-editor-border focus-within:border-editor-accent transition-colors">
            {/* Textarea row */}
            <div className="flex items-end">
              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={
                  contextFiles.length > 0
                    ? `Ask about ${contextFiles.map(f => f.path.split('/').pop()).join(', ')}...`
                    : 'Ask about your code...'
                }
                rows={1}
                className="flex-1 bg-transparent text-editor-text placeholder-editor-muted py-3 px-4 resize-none focus:outline-none max-h-48"
              />
            </div>

            {/* Context bar with controls and send button */}
            <div className="flex items-center gap-2 px-3 py-2 border-t border-editor-border/50">
              {/* Context count */}
              {contextFiles.length > 0 && (
                <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-editor-bg border border-editor-border text-xs">
                  <Paperclip size={12} className="text-editor-accent" />
                  <span className="text-editor-text">{contextFiles.length} file{contextFiles.length !== 1 ? 's' : ''}</span>
                  <button
                    type="button"
                    onClick={clearContextFiles}
                    className="text-editor-muted hover:text-editor-error transition-colors"
                    title="Clear all context files"
                  >
                    <X size={12} />
                  </button>
                </div>
              )}

              <div className="flex-1" />

              {/* Mode switcher */}
              <ModeSwitcher />

              {/* Thinking toggle */}
              <ThinkingToggle />

              {/* Send button */}
              <button
                type="submit"
                disabled={!input.trim()}
                className="p-2 rounded-lg bg-editor-accent text-white hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                title={!isConnected ? 'Queue message (not connected)' : isGenerating ? 'Queue message' : 'Send message'}
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
            </div>
            <span className="text-xs text-editor-muted">Shift + Enter for new line</span>
          </div>
        </form>
      </div>
    </div>
  );
}
