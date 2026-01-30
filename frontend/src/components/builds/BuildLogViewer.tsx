import { useEffect, useRef, useState, useCallback } from 'react';
import { Copy, Check, Download, Search, X, Loader2, ArrowDown } from 'lucide-react';
import type { BuildLog } from '../../services/buildHistory';

interface BuildLogViewerProps {
  logs: BuildLog[];
  isLoading: boolean;
  isRunning?: boolean;
}

export function BuildLogViewer({ logs, isLoading, isRunning }: BuildLogViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [copied, setCopied] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [isAtBottom, setIsAtBottom] = useState(true);

  // Auto-scroll to bottom when new logs arrive (if auto-scroll is enabled)
  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // Detect scroll position to show/hide "scroll to bottom" button
  const handleScroll = useCallback(() => {
    if (!containerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    const atBottom = scrollHeight - scrollTop - clientHeight < 50;
    setIsAtBottom(atBottom);
    if (atBottom) {
      setAutoScroll(true);
    } else {
      setAutoScroll(false);
    }
  }, []);

  const scrollToBottom = () => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
      setAutoScroll(true);
    }
  };

  const handleCopy = async () => {
    const text = logs.map((log) => log.content).join('\n');
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDownload = () => {
    const text = logs.map((log) => `[${log.timestamp}] [${log.stream}] ${log.content}`).join('\n');
    const blob = new Blob([text], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `build-logs-${new Date().toISOString().split('T')[0]}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const filteredLogs = searchQuery
    ? logs.filter((log) => log.content.toLowerCase().includes(searchQuery.toLowerCase()))
    : logs;

  const highlightMatch = (text: string): React.ReactNode => {
    if (!searchQuery) return text;
    const parts = text.split(new RegExp(`(${searchQuery})`, 'gi'));
    return parts.map((part, i) =>
      part.toLowerCase() === searchQuery.toLowerCase() ? (
        <mark key={i} className="bg-editor-warning/40 text-inherit rounded px-0.5">
          {part}
        </mark>
      ) : (
        part
      )
    );
  };

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center text-editor-muted">
        <Loader2 className="animate-spin mr-2" size={20} />
        Loading logs...
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-editor-surface">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border bg-editor-bg">
        <div className="flex items-center gap-2">
          <span className="text-sm text-editor-muted">
            {logs.length} {logs.length === 1 ? 'line' : 'lines'}
            {searchQuery && ` (${filteredLogs.length} matches)`}
          </span>
          {isRunning && (
            <span className="flex items-center gap-1.5 text-xs text-editor-accent">
              <span className="w-1.5 h-1.5 rounded-full bg-editor-accent animate-pulse" />
              Live
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          {showSearch ? (
            <div className="flex items-center gap-1 bg-editor-surface rounded px-2 py-1">
              <Search size={14} className="text-editor-muted" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search logs..."
                className="bg-transparent text-sm text-editor-text outline-none w-40"
                autoFocus
              />
              <button
                onClick={() => {
                  setShowSearch(false);
                  setSearchQuery('');
                }}
                className="p-0.5 text-editor-muted hover:text-editor-text"
              >
                <X size={14} />
              </button>
            </div>
          ) : (
            <button
              onClick={() => setShowSearch(true)}
              className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
              title="Search logs"
            >
              <Search size={16} />
            </button>
          )}

          <button
            onClick={handleCopy}
            className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
            title="Copy logs"
          >
            {copied ? <Check size={16} className="text-editor-success" /> : <Copy size={16} />}
          </button>

          <button
            onClick={handleDownload}
            className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
            title="Download logs"
          >
            <Download size={16} />
          </button>
        </div>
      </div>

      {/* Log content */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="flex-1 overflow-auto font-mono text-sm p-4 relative"
      >
        {filteredLogs.length === 0 ? (
          <div className="text-editor-muted text-center py-8">
            {logs.length === 0 ? (
              <>
                <p>No logs available</p>
                {isRunning && <p className="text-xs mt-1">Waiting for output...</p>}
              </>
            ) : (
              <p>No logs match your search</p>
            )}
          </div>
        ) : (
          <div className="space-y-0.5">
            {filteredLogs.map((log, index) => (
              <div
                key={log.id}
                className={`flex items-start gap-3 hover:bg-editor-border/20 px-2 py-0.5 rounded ${
                  log.stream === 'stderr' ? 'text-editor-error' : 'text-editor-text'
                }`}
              >
                <span className="text-editor-muted/50 select-none w-8 text-right flex-shrink-0">
                  {index + 1}
                </span>
                <span className="whitespace-pre-wrap break-all flex-1">
                  {highlightMatch(log.content)}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Scroll to bottom button */}
        {!isAtBottom && logs.length > 0 && (
          <button
            onClick={scrollToBottom}
            className="absolute bottom-4 right-4 p-2 bg-editor-accent text-editor-bg rounded-full shadow-lg hover:bg-editor-accent/80 transition-colors"
            title="Scroll to bottom"
          >
            <ArrowDown size={16} />
          </button>
        )}
      </div>
    </div>
  );
}
