import { X, Clock, Bot, Folder } from 'lucide-react';
import { useAppStore } from '../../store';

export function MessageQueue() {
  const { messageQueue, removeFromQueue } = useAppStore();

  if (messageQueue.length === 0) return null;

  return (
    <div className="border-t border-editor-border px-4 py-2 bg-editor-surface/30">
      <div className="flex items-center gap-2 mb-2">
        <Clock size={14} className="text-editor-muted" />
        <span className="text-xs text-editor-muted">
          {messageQueue.length} message{messageQueue.length !== 1 ? 's' : ''} queued
        </span>
      </div>
      <div className="space-y-1.5 max-h-40 overflow-y-auto">
        {messageQueue.map((msg, index) => (
          <div
            key={msg.id}
            className="px-2 py-2 rounded bg-editor-surface text-sm group"
          >
            <div className="flex items-center gap-2">
              <span className="text-editor-muted text-xs w-4 flex-shrink-0">{index + 1}</span>
              <span className="flex-1 truncate text-editor-text">{msg.content}</span>
              <button
                onClick={() => removeFromQueue(msg.id)}
                className="p-1 rounded hover:bg-editor-error/20 text-editor-muted hover:text-editor-error transition-colors opacity-0 group-hover:opacity-100"
                title="Remove from queue"
              >
                <X size={12} />
              </button>
            </div>
            <div className="flex items-center gap-3 mt-1.5 ml-6">
              <span className="flex items-center gap-1 text-xs text-editor-muted">
                <Bot size={10} />
                <span className="text-editor-accent">{msg.model}</span>
              </span>
              <span className="flex items-center gap-1 text-xs text-editor-muted">
                <Folder size={10} />
                <span>{msg.projectFolder}</span>
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
