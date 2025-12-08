import React, { useState, useRef, useEffect } from 'react';
import {
  BarChart3,
  Zap,
  Clock,
  Hash,
  Timer,
  Bot,
  ListTodo,
  Folder,
  X,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { useAppStore } from '../store';

interface MetricRowProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  highlight?: boolean;
}

function MetricRow({ icon, label, value, highlight = false }: MetricRowProps) {
  return (
    <div className="flex items-center justify-between py-1.5">
      <div className="flex items-center gap-2">
        <span className={highlight ? 'text-editor-accent' : 'text-editor-muted'}>
          {icon}
        </span>
        <span className="text-xs text-editor-muted">{label}</span>
      </div>
      <span className={`text-sm font-mono ${highlight ? 'text-editor-accent' : 'text-editor-text'}`}>
        {value}
      </span>
    </div>
  );
}

export function MetricsDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const [isQueueExpanded, setIsQueueExpanded] = useState(true);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { metrics, messageQueue, removeFromQueue } = useAppStore();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const formatTime = (ms: number | null): string => {
    if (ms === null) return '--';
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  // Count active agents (currently just 1 if generating)
  const activeAgents = metrics.isGenerating ? 1 : 0;

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`p-2 rounded-lg transition-colors relative ${
          isOpen
            ? 'bg-editor-accent/20 text-editor-accent'
            : 'hover:bg-editor-surface text-editor-muted hover:text-editor-text'
        }`}
        title="View metrics"
      >
        <BarChart3 size={20} />
        {/* Badge for queue count */}
        {messageQueue.length > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[16px] h-4 bg-editor-accent text-[10px] text-white rounded-full flex items-center justify-center px-1">
            {messageQueue.length}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-2 w-72 bg-editor-bg border border-editor-border rounded-lg shadow-xl overflow-hidden z-50">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/50">
            <span className="text-sm font-medium text-editor-text">Metrics</span>
            <span className={`flex items-center gap-1.5 text-xs ${
              metrics.isGenerating ? 'text-editor-success' : 'text-editor-muted'
            }`}>
              <span className={`w-2 h-2 rounded-full ${
                metrics.isGenerating ? 'bg-editor-success animate-pulse' : 'bg-editor-muted'
              }`} />
              {metrics.isGenerating ? 'Generating' : 'Idle'}
            </span>
          </div>

          {/* Metrics Grid */}
          <div className="p-3 space-y-1">
            <MetricRow
              icon={<Zap size={14} />}
              label="Tokens/sec"
              value={`${metrics.tokensPerSecond.toFixed(1)} t/s`}
              highlight={metrics.isGenerating}
            />
            <MetricRow
              icon={<Hash size={14} />}
              label="Total Tokens"
              value={metrics.tokenCount.toLocaleString()}
            />
            <MetricRow
              icon={<Timer size={14} />}
              label="Time to First Token"
              value={formatTime(metrics.timeToFirstToken)}
            />
            <MetricRow
              icon={<Clock size={14} />}
              label="Elapsed Time"
              value={formatTime(metrics.elapsedTime)}
            />

            <div className="border-t border-editor-border my-2" />

            <MetricRow
              icon={<Bot size={14} />}
              label="Active Agents"
              value={activeAgents.toString()}
              highlight={activeAgents > 0}
            />

            {/* Queued Messages Section */}
            <div className="mt-2">
              <button
                onClick={() => setIsQueueExpanded(!isQueueExpanded)}
                className="w-full flex items-center justify-between py-1.5 text-left"
              >
                <div className="flex items-center gap-2">
                  <span className={messageQueue.length > 0 ? 'text-editor-accent' : 'text-editor-muted'}>
                    <ListTodo size={14} />
                  </span>
                  <span className="text-xs text-editor-muted">Queued Messages</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`text-sm font-mono ${messageQueue.length > 0 ? 'text-editor-accent' : 'text-editor-text'}`}>
                    {messageQueue.length}
                  </span>
                  {messageQueue.length > 0 && (
                    isQueueExpanded ? <ChevronUp size={12} className="text-editor-muted" /> : <ChevronDown size={12} className="text-editor-muted" />
                  )}
                </div>
              </button>

              {isQueueExpanded && messageQueue.length > 0 && (
                <div className="mt-2 space-y-1.5 max-h-48 overflow-y-auto">
                  {messageQueue.map((msg, index) => (
                    <div
                      key={msg.id}
                      className="p-2 rounded bg-editor-surface text-xs group"
                    >
                      <div className="flex items-start gap-2">
                        <span className="text-editor-muted flex-shrink-0">{index + 1}.</span>
                        <span className="flex-1 text-editor-text line-clamp-2">{msg.content}</span>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            removeFromQueue(msg.id);
                          }}
                          className="p-0.5 rounded hover:bg-editor-error/20 text-editor-muted hover:text-editor-error transition-colors opacity-0 group-hover:opacity-100 flex-shrink-0"
                          title="Remove from queue"
                        >
                          <X size={10} />
                        </button>
                      </div>
                      <div className="flex items-center gap-3 mt-1.5 ml-4">
                        <span className="flex items-center gap-1 text-editor-muted">
                          <Bot size={10} />
                          <span className="text-editor-accent">{msg.model}</span>
                        </span>
                        <span className="flex items-center gap-1 text-editor-muted">
                          <Folder size={10} />
                          <span>{msg.projectFolder}</span>
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
