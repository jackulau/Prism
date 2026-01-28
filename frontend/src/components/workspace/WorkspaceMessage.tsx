import { useState } from 'react';
import { Bot, User, Copy, Check, RefreshCw, FileCode } from 'lucide-react';
import { MarkdownRenderer } from '../markdown/MarkdownRenderer';
import { InlineCodeDiff } from './CodeDiff';
import type { Message } from '../../types';
import type { MessagePart, EditSuggestion } from '../../types/workspace';

interface WorkspaceMessageProps {
  message: Message;
  onFileClick?: (path: string, line?: number) => void;
  onApplyEdit?: (edit: EditSuggestion) => void;
  onRejectEdit?: (editId: string) => void;
  onRegenerate?: () => void;
}

interface FileReferenceProps {
  path: string;
  line?: number;
  onClick: () => void;
}

function FileReference({ path, line, onClick }: FileReferenceProps) {
  const fileName = path.split('/').pop() || path;

  return (
    <button
      onClick={onClick}
      className="inline-flex items-center gap-1 px-1.5 py-0.5 mx-0.5 rounded bg-editor-surface border border-editor-border hover:border-editor-accent/50 text-editor-accent transition-colors text-sm"
      title={`Open ${path}${line ? ` at line ${line}` : ''}`}
    >
      <FileCode size={12} />
      <span>{fileName}</span>
      {line && <span className="text-editor-muted">:{line}</span>}
    </button>
  );
}

// Parse message content to extract file references and code edits
function parseMessageWithRefs(content: string): MessagePart[] {
  const parts: MessagePart[] = [];

  // Pattern for file references: `path/to/file.ext:123` or just `path/to/file.ext`
  // They're typically wrapped in backticks in the content
  const fileRefPattern = /`([^`]+\.[a-zA-Z]+)(?::(\d+))?`/g;

  let lastIndex = 0;
  let match;

  while ((match = fileRefPattern.exec(content)) !== null) {
    // Add text before this match
    if (match.index > lastIndex) {
      parts.push({
        type: 'text',
        content: content.slice(lastIndex, match.index),
      });
    }

    // Add file reference
    const path = match[1];
    const line = match[2] ? parseInt(match[2], 10) : undefined;

    // Simple heuristic: only treat as file ref if it looks like a path
    if (path.includes('/') || path.includes('.')) {
      parts.push({
        type: 'file-ref',
        path,
        line,
      });
    } else {
      // Not a file reference, just code
      parts.push({
        type: 'text',
        content: match[0],
      });
    }

    lastIndex = match.index + match[0].length;
  }

  // Add remaining text
  if (lastIndex < content.length) {
    parts.push({
      type: 'text',
      content: content.slice(lastIndex),
    });
  }

  // If no parts were created, return the whole content as text
  if (parts.length === 0) {
    parts.push({ type: 'text', content });
  }

  return parts;
}

export function WorkspaceMessage({
  message,
  onFileClick,
  onApplyEdit,
  onRejectEdit,
  onRegenerate,
}: WorkspaceMessageProps) {
  const [copied, setCopied] = useState(false);
  const isUser = message.role === 'user';
  const parts = parseMessageWithRefs(message.content);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(message.content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={`py-4 px-4 ${isUser ? 'bg-editor-surface/30' : ''} group relative`}>
      <div className="max-w-3xl mx-auto">
        {/* Header row */}
        <div className="flex items-center gap-3 mb-2">
          <div
            className={`w-8 h-8 rounded-lg flex items-center justify-center border ${
              isUser
                ? 'bg-blue-500/20 text-blue-400 border-blue-500/30'
                : 'bg-purple-500/20 text-purple-400 border-purple-500/30'
            }`}
          >
            {isUser ? <User size={18} /> : <Bot size={18} />}
          </div>
          <div className="flex items-center gap-2 flex-1">
            <span className="font-medium text-editor-text">{isUser ? 'You' : 'Assistant'}</span>
            <span className="text-xs text-editor-muted">
              {message.timestamp.toLocaleTimeString()}
            </span>
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
              <button
                onClick={handleCopy}
                className="p-1.5 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
                title="Copy message"
              >
                {copied ? <Check size={14} className="text-editor-success" /> : <Copy size={14} />}
              </button>
              {!isUser && onRegenerate && (
                <button
                  onClick={onRegenerate}
                  className="p-1.5 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
                  title="Regenerate response"
                >
                  <RefreshCw size={14} />
                </button>
              )}
            </div>
          )}
        </div>

        {/* Message content with file references */}
        <div className="pl-11">
          <div className="text-editor-text leading-relaxed">
            {parts.map((part, i) => {
              if (part.type === 'text' && part.content) {
                return (
                  <MarkdownRenderer
                    key={i}
                    content={part.content}
                    isStreaming={message.isStreaming}
                  />
                );
              }
              if (part.type === 'file-ref' && part.path) {
                return (
                  <FileReference
                    key={i}
                    path={part.path}
                    line={part.line}
                    onClick={() => onFileClick?.(part.path!, part.line)}
                  />
                );
              }
              if (part.type === 'code-edit' && part.original && part.modified && part.language) {
                const editId = `edit-${i}`;
                return (
                  <InlineCodeDiff
                    key={i}
                    original={part.original}
                    modified={part.modified}
                    language={part.language}
                    onAccept={() =>
                      onApplyEdit?.({
                        id: editId,
                        path: part.path || '',
                        original: part.original!,
                        modified: part.modified!,
                        startLine: 0,
                        endLine: 0,
                        description: '',
                        status: 'pending',
                      })
                    }
                    onReject={() => onRejectEdit?.(editId)}
                  />
                );
              }
              return null;
            })}
          </div>
        </div>

        {/* Metrics for completed assistant messages */}
        {!isUser && message.metrics && !message.isStreaming && (
          <div className="pl-11 mt-3 flex items-center gap-4 text-xs text-editor-muted">
            <span>
              {message.metrics.totalTokens} tokens
            </span>
            {message.metrics.tokensPerSecond != null && (
              <span>
                {message.metrics.tokensPerSecond.toFixed(1)} t/s
              </span>
            )}
            {message.metrics.timeToFirstToken != null && (
              <span>
                TTFT: {message.metrics.timeToFirstToken.toFixed(0)}ms
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
