import React, { useMemo } from 'react';
import { User, Bot, Wrench, AlertCircle, Clock, Zap, Hash } from 'lucide-react';
import { CodeBlock } from './CodeBlock';
import type { Message } from '../types';

interface ChatMessageProps {
  message: Message;
}

// Simple markdown parser for code blocks and basic formatting
const parseContent = (content: string, isStreaming: boolean): React.ReactNode[] => {
  const elements: React.ReactNode[] = [];
  let currentIndex = 0;

  // Match code blocks: ```language\ncode\n```
  const codeBlockRegex = /```(\w*)\n?([\s\S]*?)```/g;
  let match;

  while ((match = codeBlockRegex.exec(content)) !== null) {
    // Add text before code block
    if (match.index > currentIndex) {
      const textBefore = content.slice(currentIndex, match.index);
      elements.push(
        <span key={`text-${currentIndex}`} className="whitespace-pre-wrap">
          {parseInlineFormatting(textBefore)}
        </span>
      );
    }

    // Add code block
    const language = match[1] || 'text';
    const code = match[2];
    const isLastBlock = codeBlockRegex.lastIndex >= content.length - 3;

    elements.push(
      <CodeBlock
        key={`code-${match.index}`}
        code={code}
        language={language}
        isStreaming={isStreaming && isLastBlock}
      />
    );

    currentIndex = codeBlockRegex.lastIndex;
  }

  // Add remaining text
  if (currentIndex < content.length) {
    const remainingText = content.slice(currentIndex);
    // Check if we're in the middle of a code block
    const unclosedCodeBlock = remainingText.match(/```(\w*)\n?([\s\S]*)$/);

    if (unclosedCodeBlock && isStreaming) {
      elements.push(
        <CodeBlock
          key={`code-partial`}
          code={unclosedCodeBlock[2]}
          language={unclosedCodeBlock[1] || 'text'}
          isStreaming={true}
        />
      );
    } else {
      elements.push(
        <span key={`text-${currentIndex}`} className="whitespace-pre-wrap">
          {parseInlineFormatting(remainingText)}
          {isStreaming && (
            <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
          )}
        </span>
      );
    }
  }

  return elements;
};

// Parse inline formatting: **bold**, *italic*, `code`, [link](url)
const parseInlineFormatting = (text: string): React.ReactNode[] => {
  const elements: React.ReactNode[] = [];
  let remaining = text;
  let keyIndex = 0;

  while (remaining.length > 0) {
    // Bold: **text**
    const boldMatch = remaining.match(/^\*\*(.+?)\*\*/);
    if (boldMatch) {
      elements.push(
        <strong key={`bold-${keyIndex++}`} className="font-semibold text-editor-text">
          {boldMatch[1]}
        </strong>
      );
      remaining = remaining.slice(boldMatch[0].length);
      continue;
    }

    // Italic: *text*
    const italicMatch = remaining.match(/^\*(.+?)\*/);
    if (italicMatch) {
      elements.push(
        <em key={`italic-${keyIndex++}`} className="italic">
          {italicMatch[1]}
        </em>
      );
      remaining = remaining.slice(italicMatch[0].length);
      continue;
    }

    // Inline code: `code`
    const codeMatch = remaining.match(/^`([^`]+)`/);
    if (codeMatch) {
      elements.push(
        <code
          key={`code-${keyIndex++}`}
          className="px-1.5 py-0.5 rounded bg-editor-surface text-editor-accent text-sm font-mono border border-editor-border"
        >
          {codeMatch[1]}
        </code>
      );
      remaining = remaining.slice(codeMatch[0].length);
      continue;
    }

    // Link: [text](url)
    const linkMatch = remaining.match(/^\[([^\]]+)\]\(([^)]+)\)/);
    if (linkMatch) {
      elements.push(
        <a
          key={`link-${keyIndex++}`}
          href={linkMatch[2]}
          target="_blank"
          rel="noopener noreferrer"
          className="text-editor-accent hover:underline"
        >
          {linkMatch[1]}
        </a>
      );
      remaining = remaining.slice(linkMatch[0].length);
      continue;
    }

    // Regular character
    elements.push(remaining[0]);
    remaining = remaining.slice(1);
  }

  return elements;
};

const getRoleIcon = (role: Message['role']) => {
  switch (role) {
    case 'user':
      return <User className="w-5 h-5" />;
    case 'assistant':
      return <Bot className="w-5 h-5" />;
    case 'tool':
      return <Wrench className="w-5 h-5" />;
    default:
      return <AlertCircle className="w-5 h-5" />;
  }
};

const getRoleColor = (role: Message['role']) => {
  switch (role) {
    case 'user':
      return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
    case 'assistant':
      return 'bg-purple-500/20 text-purple-400 border-purple-500/30';
    case 'tool':
      return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
    default:
      return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
  }
};

const getRoleName = (role: Message['role']) => {
  switch (role) {
    case 'user':
      return 'You';
    case 'assistant':
      return 'Assistant';
    case 'tool':
      return 'Tool';
    default:
      return 'System';
  }
};

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
  const content = useMemo(
    () => parseContent(message.content, message.isStreaming || false),
    [message.content, message.isStreaming]
  );

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div
      className={`group py-4 px-4 transition-smooth ${
        message.role === 'user' ? 'bg-editor-surface/30' : ''
      }`}
    >
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-3 mb-2">
          <div
            className={`w-8 h-8 rounded-lg flex items-center justify-center border ${getRoleColor(
              message.role
            )}`}
          >
            {getRoleIcon(message.role)}
          </div>
          <div className="flex items-center gap-2">
            <span className="font-medium text-editor-text">
              {getRoleName(message.role)}
            </span>
            <span className="text-xs text-editor-muted">
              {formatTime(message.timestamp)}
            </span>
            {message.isStreaming && (
              <span className="flex items-center gap-1 text-xs text-editor-accent">
                <span className="w-1.5 h-1.5 rounded-full bg-editor-accent animate-pulse" />
                Generating...
              </span>
            )}
          </div>
        </div>

        {/* Content */}
        <div className="pl-11 text-editor-text leading-relaxed">{content}</div>

        {/* Metrics (for completed assistant messages) */}
        {message.role === 'assistant' && message.metrics && !message.isStreaming && (
          <div className="pl-11 mt-3 flex items-center gap-4 text-xs text-editor-muted">
            <span className="flex items-center gap-1">
              <Hash className="w-3 h-3" />
              {message.metrics.totalTokens} tokens
            </span>
            {message.metrics.tokensPerSecond && (
              <span className="flex items-center gap-1">
                <Zap className="w-3 h-3" />
                {message.metrics.tokensPerSecond.toFixed(1)} t/s
              </span>
            )}
            {message.metrics.timeToFirstToken && (
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                TTFT: {message.metrics.timeToFirstToken.toFixed(0)}ms
              </span>
            )}
          </div>
        )}

        {/* Tool calls */}
        {message.toolCalls && message.toolCalls.length > 0 && (
          <div className="pl-11 mt-3 space-y-2">
            {message.toolCalls.map((tool) => (
              <div
                key={tool.id}
                className="rounded-lg border border-editor-border bg-editor-surface/50 p-3"
              >
                <div className="flex items-center gap-2 mb-2">
                  <Wrench className="w-4 h-4 text-orange-400" />
                  <span className="font-medium text-sm text-editor-text">
                    {tool.name}
                  </span>
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      tool.status === 'completed'
                        ? 'bg-editor-success/20 text-editor-success'
                        : tool.status === 'failed'
                        ? 'bg-editor-error/20 text-editor-error'
                        : tool.status === 'running'
                        ? 'bg-editor-warning/20 text-editor-warning'
                        : 'bg-editor-muted/20 text-editor-muted'
                    }`}
                  >
                    {tool.status}
                  </span>
                </div>
                {tool.result !== undefined && tool.result !== null && (
                  <pre className="text-xs text-editor-muted bg-editor-bg rounded p-2 overflow-x-auto">
                    {typeof tool.result === 'string'
                      ? tool.result
                      : JSON.stringify(tool.result, null, 2)}
                  </pre>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
