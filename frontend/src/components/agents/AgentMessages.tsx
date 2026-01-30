import { memo } from 'react';
import { User, Bot, Settings, Wrench, Hash } from 'lucide-react';
import type { Message } from '../../types';

interface AgentMessagesProps {
  messages: Message[];
}

const roleStyles = {
  user: {
    bg: 'bg-editor-surface border-editor-accent/30',
    icon: User,
    iconColor: 'text-editor-accent',
    label: 'User',
  },
  assistant: {
    bg: 'bg-editor-bg border-editor-border',
    icon: Bot,
    iconColor: 'text-editor-accent',
    label: 'Assistant',
  },
  system: {
    bg: 'bg-editor-muted/10 border-editor-muted/30',
    icon: Settings,
    iconColor: 'text-editor-muted',
    label: 'System',
  },
  tool: {
    bg: 'bg-yellow-500/10 border-yellow-500/30',
    icon: Wrench,
    iconColor: 'text-yellow-400',
    label: 'Tool',
  },
};

function formatTimestamp(date: Date): string {
  return new Date(date).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

const MessageCard = memo(function MessageCard({ message }: { message: Message }) {
  const role = message.role as keyof typeof roleStyles;
  const style = roleStyles[role] || roleStyles.user;
  const Icon = style.icon;

  const totalTokens =
    (message.input_tokens || 0) + (message.output_tokens || 0);

  return (
    <div
      className={`rounded-lg border p-4 ${style.bg} transition-colors hover:border-opacity-50`}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className={`p-1.5 rounded-lg bg-editor-surface/50 ${style.iconColor}`}>
            <Icon size={14} />
          </div>
          <span className="text-sm font-medium text-editor-text">
            {style.label}
          </span>
          {message.model && (
            <span className="text-xs text-editor-muted bg-editor-surface px-2 py-0.5 rounded">
              {message.model}
            </span>
          )}
        </div>
        <div className="flex items-center gap-3 text-xs text-editor-muted">
          {totalTokens > 0 && (
            <span className="flex items-center gap-1">
              <Hash size={12} />
              {totalTokens.toLocaleString()} tokens
            </span>
          )}
          <span>{formatTimestamp(message.timestamp)}</span>
        </div>
      </div>

      {/* Content */}
      <div className="text-sm text-editor-text whitespace-pre-wrap break-words">
        {message.content || (
          <span className="text-editor-muted italic">No content</span>
        )}
      </div>

      {/* Thinking content (if available) */}
      {message.thinking_content && (
        <div className="mt-3 pt-3 border-t border-editor-border">
          <div className="text-xs text-editor-muted mb-2 font-medium">
            Thinking Process
          </div>
          <div className="text-sm text-editor-muted italic whitespace-pre-wrap break-words bg-editor-surface/30 rounded-lg p-3">
            {message.thinking_content}
          </div>
        </div>
      )}

      {/* Tool calls preview (if any) */}
      {message.toolCalls && message.toolCalls.length > 0 && (
        <div className="mt-3 pt-3 border-t border-editor-border">
          <div className="text-xs text-editor-muted mb-2">
            Tool Calls ({message.toolCalls.length})
          </div>
          <div className="space-y-1">
            {message.toolCalls.map((toolCall) => (
              <div
                key={toolCall.id}
                className="flex items-center gap-2 text-xs bg-editor-surface/50 rounded px-2 py-1"
              >
                <Wrench size={12} className="text-yellow-400" />
                <span className="font-mono text-editor-text">
                  {toolCall.name}
                </span>
                <span
                  className={`px-1.5 py-0.5 rounded text-[10px] ${
                    toolCall.status === 'completed'
                      ? 'bg-green-500/20 text-green-400'
                      : toolCall.status === 'failed'
                      ? 'bg-red-500/20 text-red-400'
                      : toolCall.status === 'running'
                      ? 'bg-blue-500/20 text-blue-400'
                      : 'bg-editor-muted/20 text-editor-muted'
                  }`}
                >
                  {toolCall.status}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Token breakdown (if available) */}
      {(message.input_tokens || message.output_tokens) && (
        <div className="mt-3 pt-3 border-t border-editor-border flex items-center gap-4 text-xs text-editor-muted">
          {message.input_tokens && (
            <span>Input: {message.input_tokens.toLocaleString()}</span>
          )}
          {message.output_tokens && (
            <span>Output: {message.output_tokens.toLocaleString()}</span>
          )}
        </div>
      )}
    </div>
  );
});

export function AgentMessages({ messages }: AgentMessagesProps) {
  if (!messages || messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Bot className="w-12 h-12 text-editor-muted mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">
          No messages yet
        </h3>
        <p className="text-sm text-editor-muted max-w-md">
          Messages will appear here as the agent processes your request.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {messages.map((message, index) => (
        <MessageCard key={message.id || index} message={message} />
      ))}
    </div>
  );
}

export default AgentMessages;
