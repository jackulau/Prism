import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { History, X, Loader2, RotateCcw, LayoutList, GitBranch, Folder } from 'lucide-react';
import { HistoryFilters } from './HistoryFilters';
import { HistoryEntryCard } from './HistoryEntryCard';
import { HistoryTimeline } from './HistoryTimeline';
import type { FileHistoryEntry, HistoryFilters as HistoryFiltersType, FileHistoryStats } from '../../types';

interface FileHistoryPanelProps {
  filePath?: string;
  onSelectEntry: (entry: FileHistoryEntry) => void;
  onRestoreEntry: (entry: FileHistoryEntry) => void;
  isOpen: boolean;
  onClose: () => void;
  // Data props (to be connected to store)
  entries: FileHistoryEntry[];
  stats?: FileHistoryStats | null;
  isLoading: boolean;
  error?: string | null;
  onLoadHistory: (filePath?: string) => void;
}

type ViewMode = 'list' | 'timeline';
type TimelineGroupBy = 'day' | 'hour' | 'file';

export function FileHistoryPanel({
  filePath,
  onSelectEntry,
  onRestoreEntry,
  isOpen,
  onClose,
  entries,
  stats,
  isLoading,
  error,
  onLoadHistory,
}: FileHistoryPanelProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [timelineGroupBy, setTimelineGroupBy] = useState<TimelineGroupBy>('day');
  const [filters, setFilters] = useState<HistoryFiltersType>({});
  const [selectedEntryId, setSelectedEntryId] = useState<string | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  // Load history when panel opens or filePath changes
  useEffect(() => {
    if (isOpen) {
      onLoadHistory(filePath);
    }
  }, [isOpen, filePath, onLoadHistory]);

  // Filter entries based on current filters
  const filteredEntries = useMemo(() => {
    return entries.filter((entry) => {
      // File path filter
      if (filters.filePath && entry.file_path !== filters.filePath) {
        return false;
      }

      // Operation type filter
      if (filters.operations?.length && !filters.operations.includes(entry.operation as 'create' | 'update' | 'delete')) {
        return false;
      }

      // Search query filter
      if (filters.searchQuery) {
        const query = filters.searchQuery.toLowerCase();
        const matchesPath = entry.file_path.toLowerCase().includes(query);
        const matchesDescription = entry.description?.toLowerCase().includes(query);
        const matchesAgent = entry.agent_name?.toLowerCase().includes(query);
        if (!matchesPath && !matchesDescription && !matchesAgent) {
          return false;
        }
      }

      // Date range filter
      if (filters.dateRange) {
        const entryDate = new Date(entry.created_at);
        if (entryDate < filters.dateRange.start || entryDate > filters.dateRange.end) {
          return false;
        }
      }

      return true;
    });
  }, [entries, filters]);

  // Get unique file paths for filter dropdown
  const availableFiles = useMemo(() => {
    const paths = new Set(entries.map((e) => e.file_path));
    return Array.from(paths).sort();
  }, [entries]);

  // Get selected entry
  const selectedEntry = useMemo(() => {
    return filteredEntries.find((e) => e.id === selectedEntryId) || null;
  }, [filteredEntries, selectedEntryId]);

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Escape to close
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
        return;
      }

      // Focus filter input
      if (e.key === 'f' && !e.metaKey && !e.ctrlKey) {
        const searchInput = panelRef.current?.querySelector('input[type="text"]') as HTMLInputElement;
        if (searchInput && document.activeElement !== searchInput) {
          e.preventDefault();
          searchInput.focus();
        }
        return;
      }

      // Navigate entries
      const currentIndex = filteredEntries.findIndex((e) => e.id === selectedEntryId);

      if (e.key === 'ArrowDown' || e.key === 'j') {
        e.preventDefault();
        const nextIndex = currentIndex < filteredEntries.length - 1 ? currentIndex + 1 : 0;
        setSelectedEntryId(filteredEntries[nextIndex]?.id || null);
      } else if (e.key === 'ArrowUp' || e.key === 'k') {
        e.preventDefault();
        const prevIndex = currentIndex > 0 ? currentIndex - 1 : filteredEntries.length - 1;
        setSelectedEntryId(filteredEntries[prevIndex]?.id || null);
      } else if (e.key === 'Enter' && selectedEntry) {
        e.preventDefault();
        onSelectEntry(selectedEntry);
      } else if (e.key === 'r' && selectedEntry && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        onRestoreEntry(selectedEntry);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, filteredEntries, selectedEntryId, selectedEntry, onClose, onSelectEntry, onRestoreEntry]);

  const handleSelectEntry = useCallback((entry: FileHistoryEntry) => {
    setSelectedEntryId(entry.id);
    onSelectEntry(entry);
  }, [onSelectEntry]);

  const handlePreviewEntry = useCallback((entry: FileHistoryEntry) => {
    setSelectedEntryId(entry.id);
    onSelectEntry(entry);
  }, [onSelectEntry]);

  const handleRestoreEntry = useCallback((entry: FileHistoryEntry) => {
    onRestoreEntry(entry);
  }, [onRestoreEntry]);

  if (!isOpen) {
    return null;
  }

  return (
    <div
      ref={panelRef}
      className="w-80 bg-editor-bg border-l border-editor-border flex flex-col h-full"
    >
      {/* Header */}
      <div className="p-3 border-b border-editor-border flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2">
          <History size={16} className="text-editor-accent" />
          <span className="text-sm font-medium text-editor-text">File History</span>
          {stats && (
            <span className="text-xs text-editor-muted">
              ({stats.total_entries} {stats.total_entries === 1 ? 'entry' : 'entries'})
            </span>
          )}
        </div>
        <button
          onClick={onClose}
          className="p-1 rounded hover:bg-editor-border/50 text-editor-muted hover:text-editor-text transition-colors"
          title="Close (Esc)"
        >
          <X size={16} />
        </button>
      </div>

      {/* View mode toggle */}
      <div className="p-2 border-b border-editor-border flex items-center gap-2 shrink-0">
        <div className="flex items-center gap-1 bg-editor-surface rounded p-0.5">
          <button
            onClick={() => setViewMode('list')}
            className={`p-1.5 rounded transition-colors ${
              viewMode === 'list'
                ? 'bg-editor-bg text-editor-text'
                : 'text-editor-muted hover:text-editor-text'
            }`}
            title="List view"
          >
            <LayoutList size={14} />
          </button>
          <button
            onClick={() => setViewMode('timeline')}
            className={`p-1.5 rounded transition-colors ${
              viewMode === 'timeline'
                ? 'bg-editor-bg text-editor-text'
                : 'text-editor-muted hover:text-editor-text'
            }`}
            title="Timeline view"
          >
            <GitBranch size={14} />
          </button>
        </div>

        {viewMode === 'timeline' && (
          <select
            value={timelineGroupBy}
            onChange={(e) => setTimelineGroupBy(e.target.value as TimelineGroupBy)}
            className="text-xs bg-editor-surface border border-editor-border rounded px-2 py-1 text-editor-text focus:outline-none focus:border-editor-accent"
          >
            <option value="day">By day</option>
            <option value="hour">By hour</option>
            <option value="file">By file</option>
          </select>
        )}

        <button
          onClick={() => onLoadHistory(filePath)}
          className="ml-auto p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-border/50 transition-colors"
          title="Refresh history"
          disabled={isLoading}
        >
          <RotateCcw size={14} className={isLoading ? 'animate-spin' : ''} />
        </button>
      </div>

      {/* Filters */}
      <HistoryFilters
        onFilterChange={setFilters}
        availableFiles={availableFiles}
        currentFilters={filters}
      />

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 size={24} className="animate-spin text-editor-accent" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center h-full gap-3 p-4">
            <p className="text-sm text-editor-error text-center">{error}</p>
            <button
              onClick={() => onLoadHistory(filePath)}
              className="px-3 py-1.5 text-sm bg-editor-accent text-white rounded hover:bg-editor-accent/80 transition-colors"
            >
              Retry
            </button>
          </div>
        ) : filteredEntries.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-2 text-editor-muted p-4">
            <Folder size={32} className="opacity-50" />
            <p className="text-sm">No history entries</p>
            {filters.filePath || filters.operations?.length || filters.searchQuery ? (
              <button
                onClick={() => setFilters({})}
                className="text-xs text-editor-accent hover:underline"
              >
                Clear filters
              </button>
            ) : null}
          </div>
        ) : viewMode === 'list' ? (
          <div className="h-full overflow-y-auto">
            {filteredEntries.map((entry) => (
              <HistoryEntryCard
                key={entry.id}
                entry={entry}
                isSelected={entry.id === selectedEntryId}
                onClick={() => handleSelectEntry(entry)}
                onPreview={() => handlePreviewEntry(entry)}
                onRestore={() => handleRestoreEntry(entry)}
              />
            ))}
          </div>
        ) : (
          <HistoryTimeline
            entries={filteredEntries}
            selectedId={selectedEntryId || undefined}
            onSelect={handleSelectEntry}
            groupBy={timelineGroupBy}
          />
        )}
      </div>

      {/* Footer with keyboard shortcuts hint */}
      <div className="p-2 border-t border-editor-border text-xs text-editor-muted shrink-0">
        <span className="opacity-70">
          <kbd className="px-1 bg-editor-surface rounded">↑↓</kbd> navigate
          {' • '}
          <kbd className="px-1 bg-editor-surface rounded">Enter</kbd> preview
          {' • '}
          <kbd className="px-1 bg-editor-surface rounded">r</kbd> restore
          {' • '}
          <kbd className="px-1 bg-editor-surface rounded">f</kbd> filter
        </span>
      </div>
    </div>
  );
}
