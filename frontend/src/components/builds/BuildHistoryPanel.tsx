import { useEffect, useRef, useState } from 'react';
import { RefreshCw, Filter, ChevronDown, X } from 'lucide-react';
import { BuildHistoryList } from './BuildHistoryList';
import { BuildDetailHeader } from './BuildDetailHeader';
import { BuildLogViewer } from './BuildLogViewer';
import { useBuildHistoryStore } from '../../store/buildHistoryStore';
import type { BuildStatus } from '../../services/buildHistory';

const STATUS_OPTIONS: { value: BuildStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All Builds' },
  { value: 'running', label: 'Running' },
  { value: 'pending', label: 'Pending' },
  { value: 'success', label: 'Success' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' },
];

interface BuildHistoryPanelProps {
  pollingInterval?: number;
}

export function BuildHistoryPanel({ pollingInterval = 5000 }: BuildHistoryPanelProps) {
  const {
    builds,
    selectedBuild,
    logs,
    isLoading,
    isLoadingLogs,
    filter,
    total,
    fetchBuilds,
    fetchMoreBuilds,
    setSelectedBuild,
    setFilter,
    deleteBuild,
    cancelBuild,
    refreshBuild,
  } = useBuildHistoryStore();

  const filterDropdownRef = useRef<HTMLDivElement>(null);
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);

  // Initial fetch
  useEffect(() => {
    fetchBuilds();
  }, [fetchBuilds]);

  // Polling for running builds
  useEffect(() => {
    const hasRunningBuilds = builds.some(
      (b) => b.status === 'running' || b.status === 'pending'
    );

    if (!hasRunningBuilds || pollingInterval <= 0) return;

    const intervalId = setInterval(() => {
      // Refresh running builds
      builds
        .filter((b) => b.status === 'running' || b.status === 'pending')
        .forEach((b) => refreshBuild(b.id));
    }, pollingInterval);

    return () => clearInterval(intervalId);
  }, [builds, pollingInterval, refreshBuild]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        filterDropdownRef.current &&
        !filterDropdownRef.current.contains(event.target as Node)
      ) {
        setShowFilterDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleFilterChange = (status: BuildStatus | 'all') => {
    setFilter({ status: status === 'all' ? undefined : status });
    setShowFilterDropdown(false);
  };

  const hasMore = builds.length < total;
  const isRunning = selectedBuild?.status === 'running' || selectedBuild?.status === 'pending';
  const currentFilterLabel = STATUS_OPTIONS.find(
    (opt) => opt.value === (filter.status || 'all')
  )?.label;

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-editor-border">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-semibold text-editor-text">Build History</h2>

          {/* Filter Dropdown */}
          <div ref={filterDropdownRef} className="relative">
            <button
              onClick={() => setShowFilterDropdown(!showFilterDropdown)}
              className="flex items-center gap-1.5 px-2 py-1 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
            >
              <Filter size={14} />
              <span>{currentFilterLabel}</span>
              <ChevronDown size={14} />
            </button>

            {showFilterDropdown && (
              <div className="absolute top-full left-0 mt-1 bg-editor-surface border border-editor-border rounded-lg shadow-lg py-1 z-50 min-w-[140px]">
                {STATUS_OPTIONS.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => handleFilterChange(option.value)}
                    className={`w-full text-left px-3 py-1.5 text-sm hover:bg-editor-border/50 transition-colors ${
                      (filter.status || 'all') === option.value
                        ? 'text-editor-accent'
                        : 'text-editor-text'
                    }`}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            )}
          </div>

          {filter.status && (
            <button
              onClick={() => setFilter({})}
              className="flex items-center gap-1 px-2 py-1 text-xs text-editor-muted hover:text-editor-text bg-editor-surface rounded transition-colors"
            >
              Clear filter
              <X size={12} />
            </button>
          )}
        </div>

        <button
          onClick={() => fetchBuilds()}
          disabled={isLoading}
          className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors disabled:opacity-50"
          title="Refresh"
        >
          <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 flex min-h-0">
        {/* Build List */}
        <div className="w-80 flex-shrink-0 border-r border-editor-border flex flex-col">
          <BuildHistoryList
            builds={builds}
            selectedBuildId={selectedBuild?.id || null}
            isLoading={isLoading}
            hasMore={hasMore}
            onSelectBuild={setSelectedBuild}
            onLoadMore={fetchMoreBuilds}
          />
        </div>

        {/* Build Details + Logs */}
        <div className="flex-1 flex flex-col min-w-0">
          {selectedBuild ? (
            <>
              <BuildDetailHeader
                build={selectedBuild}
                onCancel={() => cancelBuild(selectedBuild.id)}
                onDelete={() => deleteBuild(selectedBuild.id)}
              />
              <BuildLogViewer
                logs={logs}
                isLoading={isLoadingLogs}
                isRunning={isRunning}
              />
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-editor-muted">
              <p>Select a build to view details</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
