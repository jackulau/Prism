import { useMemo, useState } from 'react';
import { ChevronDown, ChevronRight, Plus, Pencil, Trash2 } from 'lucide-react';
import type { FileHistoryEntry, FileHistoryOperation } from '../../types';

interface HistoryTimelineProps {
  entries: FileHistoryEntry[];
  selectedId?: string;
  onSelect: (entry: FileHistoryEntry) => void;
  groupBy: 'day' | 'hour' | 'file';
}

interface TimelineGroup {
  key: string;
  label: string;
  entries: FileHistoryEntry[];
}

const OPERATION_COLORS: Record<FileHistoryOperation, string> = {
  create: 'bg-green-500',
  update: 'bg-blue-500',
  delete: 'bg-red-500',
};

const OPERATION_ICONS: Record<FileHistoryOperation, typeof Plus> = {
  create: Plus,
  update: Pencil,
  delete: Trash2,
};

function formatGroupLabel(date: Date, groupBy: 'day' | 'hour' | 'file'): string {
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

  if (groupBy === 'hour') {
    if (diffDays === 0) {
      return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
    }
    return date.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
  }

  if (diffDays === 0) {
    return 'Today';
  } else if (diffDays === 1) {
    return 'Yesterday';
  } else if (diffDays < 7) {
    return date.toLocaleDateString('en-US', { weekday: 'long' });
  } else {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined });
  }
}

function getGroupKey(entry: FileHistoryEntry, groupBy: 'day' | 'hour' | 'file'): string {
  if (groupBy === 'file') {
    return entry.file_path;
  }

  const date = new Date(entry.created_at);

  if (groupBy === 'hour') {
    return `${date.toDateString()}-${date.getHours()}`;
  }

  return date.toDateString();
}

function groupEntries(entries: FileHistoryEntry[], groupBy: 'day' | 'hour' | 'file'): TimelineGroup[] {
  const groups = new Map<string, TimelineGroup>();

  entries.forEach((entry) => {
    const key = getGroupKey(entry, groupBy);

    if (!groups.has(key)) {
      const date = new Date(entry.created_at);
      const label = groupBy === 'file'
        ? entry.file_path.split('/').pop() || entry.file_path
        : formatGroupLabel(date, groupBy);

      groups.set(key, { key, label, entries: [] });
    }

    groups.get(key)!.entries.push(entry);
  });

  // Sort groups by most recent first
  const sortedGroups = Array.from(groups.values()).sort((a, b) => {
    const aDate = new Date(a.entries[0].created_at);
    const bDate = new Date(b.entries[0].created_at);
    return bDate.getTime() - aDate.getTime();
  });

  return sortedGroups;
}

function TimelineMarker({
  entry,
  isSelected,
  onClick,
}: {
  entry: FileHistoryEntry;
  isSelected: boolean;
  onClick: () => void;
}) {
  const operation = entry.operation as FileHistoryOperation;
  const color = OPERATION_COLORS[operation] || OPERATION_COLORS.update;
  const Icon = OPERATION_ICONS[operation] || OPERATION_ICONS.update;
  const fileName = entry.file_path.split('/').pop() || entry.file_path;
  const time = new Date(entry.created_at).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  });

  return (
    <button
      onClick={onClick}
      className={`group relative flex items-start gap-3 w-full py-2 px-2 rounded transition-colors ${
        isSelected
          ? 'bg-editor-accent/10'
          : 'hover:bg-editor-border/20'
      }`}
    >
      {/* Timeline connector line */}
      <div className="relative flex flex-col items-center">
        {/* Dot */}
        <div
          className={`w-3 h-3 rounded-full border-2 border-editor-bg ${color} ${
            isSelected ? 'ring-2 ring-editor-accent ring-offset-1 ring-offset-editor-bg' : ''
          }`}
        />
        {/* Line */}
        <div className="w-0.5 flex-1 bg-editor-border/50 -mt-0.5" />
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0 text-left -mt-0.5">
        <div className="flex items-center gap-2">
          <Icon size={12} className={color.replace('bg-', 'text-').replace('-500', '-400')} />
          <span className="text-xs text-editor-text truncate">{fileName}</span>
          <span className="text-xs text-editor-muted ml-auto">{time}</span>
        </div>
        {entry.description && (
          <p className="text-xs text-editor-muted mt-0.5 truncate">{entry.description}</p>
        )}
      </div>
    </button>
  );
}

function TimelineGroupSection({
  group,
  selectedId,
  onSelect,
  defaultExpanded = true,
}: {
  group: TimelineGroup;
  selectedId?: string;
  onSelect: (entry: FileHistoryEntry) => void;
  defaultExpanded?: boolean;
}) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  const operationCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    group.entries.forEach((entry) => {
      counts[entry.operation] = (counts[entry.operation] || 0) + 1;
    });
    return counts;
  }, [group.entries]);

  return (
    <div className="mb-2">
      {/* Group header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-2 w-full px-2 py-1.5 text-xs font-medium text-editor-muted hover:text-editor-text transition-colors"
      >
        {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        <span>{group.label}</span>
        <span className="ml-auto text-editor-muted/70">
          {group.entries.length} {group.entries.length === 1 ? 'change' : 'changes'}
        </span>
        {/* Operation indicators */}
        <div className="flex gap-1">
          {Object.entries(operationCounts).map(([op, count]) => (
            <span
              key={op}
              className={`w-2 h-2 rounded-full ${OPERATION_COLORS[op as FileHistoryOperation]}`}
              title={`${count} ${op}`}
            />
          ))}
        </div>
      </button>

      {/* Group entries */}
      {isExpanded && (
        <div className="ml-2 border-l border-editor-border/30">
          {group.entries.map((entry, index) => (
            <div key={entry.id} className="relative">
              <TimelineMarker
                entry={entry}
                isSelected={entry.id === selectedId}
                onClick={() => onSelect(entry)}
              />
              {/* Hide the connector line for the last item */}
              {index === group.entries.length - 1 && (
                <div className="absolute left-[17px] bottom-0 w-0.5 h-2 bg-editor-bg" />
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function HistoryTimeline({
  entries,
  selectedId,
  onSelect,
  groupBy,
}: HistoryTimelineProps) {
  const groups = useMemo(() => groupEntries(entries, groupBy), [entries, groupBy]);

  if (entries.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-editor-muted text-sm p-4">
        No history entries
      </div>
    );
  }

  return (
    <div className="overflow-y-auto">
      {groups.map((group, index) => (
        <TimelineGroupSection
          key={group.key}
          group={group}
          selectedId={selectedId}
          onSelect={onSelect}
          defaultExpanded={index < 3} // Auto-expand first 3 groups
        />
      ))}
    </div>
  );
}
