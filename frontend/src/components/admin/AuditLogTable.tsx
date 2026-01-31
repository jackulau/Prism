import { useEffect, useState } from 'react';
import { useAuditStore, type AuditLog, type AuditFilters } from '../../store/auditStore';
import {
  getEventIcon,
  getEventLabel,
  getCategoryLabel,
  getEventColor,
  formatRelativeTime,
  parseUserAgent,
  getEventCategories,
  getEventTypesByCategory,
} from '../../utils/auditHelpers';
import {
  Loader2,
  CheckCircle,
  XCircle,
  Filter,
  X,
  Download,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

interface AuditLogTableProps {
  showFilters?: boolean;
}

export function AuditLogTable({ showFilters = true }: AuditLogTableProps) {
  const {
    allLogs,
    allLogsLoading,
    allLogsError,
    allLogsTotal,
    allLogsOffset,
    filters,
    setFilters,
    clearFilters,
    fetchAllLogs,
  } = useAuditStore();

  const [isFiltersOpen, setIsFiltersOpen] = useState(false);
  const [localFilters, setLocalFilters] = useState<AuditFilters>({});
  const pageSize = 20;
  const currentPage = Math.floor(allLogsOffset / pageSize);
  const totalPages = Math.ceil(allLogsTotal / pageSize);

  useEffect(() => {
    fetchAllLogs(true);
  }, [fetchAllLogs, filters]);

  const handleFilterChange = (key: keyof AuditFilters, value: string | boolean | undefined) => {
    const newFilters = { ...localFilters, [key]: value };
    if (!value && value !== false) {
      delete newFilters[key];
    }
    setLocalFilters(newFilters);
  };

  const applyFilters = () => {
    setFilters(localFilters);
    setIsFiltersOpen(false);
  };

  const handleClearFilters = () => {
    setLocalFilters({});
    clearFilters();
    setIsFiltersOpen(false);
  };

  const handlePreviousPage = () => {
    if (currentPage > 0) {
      useAuditStore.setState({ allLogsOffset: (currentPage - 1) * pageSize });
      fetchAllLogs(true);
    }
  };

  const handleNextPage = () => {
    if (currentPage < totalPages - 1) {
      fetchAllLogs(false);
    }
  };

  const exportToCSV = () => {
    const headers = ['Time', 'Event Type', 'Category', 'Action', 'IP Address', 'Success', 'User ID'];
    const rows = allLogs.map((log) => [
      new Date(log.created_at).toISOString(),
      log.event_type,
      log.event_category,
      log.action,
      log.ip_address || '',
      log.success ? 'Yes' : 'No',
      log.user_id || '',
    ]);

    const csv = [headers, ...rows].map((row) => row.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const categories = getEventCategories();
  const eventTypes = localFilters.category
    ? getEventTypesByCategory(localFilters.category)
    : [];

  const hasActiveFilters = Object.keys(filters).length > 0;

  if (allLogsError) {
    return (
      <div className="text-red-400 text-sm p-4 bg-red-500/10 rounded-lg">
        Failed to load audit logs: {allLogsError}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      {showFilters && (
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsFiltersOpen(!isFiltersOpen)}
              className={`px-3 py-1.5 text-sm rounded-lg flex items-center gap-2 transition-colors ${
                hasActiveFilters
                  ? 'bg-editor-accent/20 text-editor-accent border border-editor-accent/30'
                  : 'bg-editor-surface border border-editor-border hover:bg-editor-bg'
              }`}
            >
              <Filter className="w-4 h-4" />
              Filters
              {hasActiveFilters && (
                <span className="px-1.5 py-0.5 text-xs bg-editor-accent text-white rounded-full">
                  {Object.keys(filters).length}
                </span>
              )}
            </button>

            {hasActiveFilters && (
              <button
                onClick={handleClearFilters}
                className="px-2 py-1.5 text-sm text-editor-muted hover:text-editor-text transition-colors"
              >
                Clear
              </button>
            )}
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm text-editor-muted">
              {allLogsTotal} total logs
            </span>
            <button
              onClick={exportToCSV}
              className="px-3 py-1.5 text-sm bg-editor-surface border border-editor-border rounded-lg flex items-center gap-2 hover:bg-editor-bg transition-colors"
            >
              <Download className="w-4 h-4" />
              Export
            </button>
          </div>
        </div>
      )}

      {/* Filters Panel */}
      {isFiltersOpen && (
        <div className="p-4 bg-editor-surface rounded-lg border border-editor-border space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">Filter Logs</h3>
            <button
              onClick={() => setIsFiltersOpen(false)}
              className="p-1 hover:bg-editor-bg rounded"
            >
              <X className="w-4 h-4 text-editor-muted" />
            </button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <label className="block text-xs text-editor-muted mb-1">Category</label>
              <select
                value={localFilters.category || ''}
                onChange={(e) => {
                  handleFilterChange('category', e.target.value || undefined);
                  handleFilterChange('event_type', undefined);
                }}
                className="w-full px-2 py-1.5 text-sm bg-editor-bg border border-editor-border rounded"
              >
                <option value="">All categories</option>
                {categories.map((cat) => (
                  <option key={cat.value} value={cat.value}>
                    {cat.label}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs text-editor-muted mb-1">Event Type</label>
              <select
                value={localFilters.event_type || ''}
                onChange={(e) => handleFilterChange('event_type', e.target.value || undefined)}
                className="w-full px-2 py-1.5 text-sm bg-editor-bg border border-editor-border rounded"
                disabled={!localFilters.category}
              >
                <option value="">All types</option>
                {eventTypes.map((et) => (
                  <option key={et.value} value={et.value}>
                    {et.label}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs text-editor-muted mb-1">Start Date</label>
              <input
                type="date"
                value={localFilters.start_date || ''}
                onChange={(e) => handleFilterChange('start_date', e.target.value || undefined)}
                className="w-full px-2 py-1.5 text-sm bg-editor-bg border border-editor-border rounded"
              />
            </div>

            <div>
              <label className="block text-xs text-editor-muted mb-1">End Date</label>
              <input
                type="date"
                value={localFilters.end_date || ''}
                onChange={(e) => handleFilterChange('end_date', e.target.value || undefined)}
                className="w-full px-2 py-1.5 text-sm bg-editor-bg border border-editor-border rounded"
              />
            </div>

            <div>
              <label className="block text-xs text-editor-muted mb-1">Status</label>
              <select
                value={localFilters.success === undefined ? '' : localFilters.success.toString()}
                onChange={(e) => {
                  const val = e.target.value;
                  handleFilterChange('success', val === '' ? undefined : val === 'true');
                }}
                className="w-full px-2 py-1.5 text-sm bg-editor-bg border border-editor-border rounded"
              >
                <option value="">All</option>
                <option value="true">Success</option>
                <option value="false">Failed</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <button
              onClick={handleClearFilters}
              className="px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text"
            >
              Clear
            </button>
            <button
              onClick={applyFilters}
              className="px-4 py-1.5 text-sm bg-primary text-white rounded hover:bg-primary/90"
            >
              Apply Filters
            </button>
          </div>
        </div>
      )}

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-editor-border">
              <th className="text-left py-2 px-3 font-medium text-editor-muted">Time</th>
              <th className="text-left py-2 px-3 font-medium text-editor-muted">Event</th>
              <th className="text-left py-2 px-3 font-medium text-editor-muted">Category</th>
              <th className="text-left py-2 px-3 font-medium text-editor-muted">IP Address</th>
              <th className="text-left py-2 px-3 font-medium text-editor-muted">Status</th>
            </tr>
          </thead>
          <tbody>
            {allLogs.map((log) => (
              <LogRow key={log.id} log={log} />
            ))}
            {allLogs.length === 0 && !allLogsLoading && (
              <tr>
                <td colSpan={5} className="py-8 text-center text-editor-muted">
                  No audit logs found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {allLogsLoading && (
        <div className="flex items-center justify-center py-4">
          <Loader2 className="w-5 h-5 animate-spin text-editor-muted" />
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-4 border-t border-editor-border">
          <span className="text-sm text-editor-muted">
            Page {currentPage + 1} of {totalPages}
          </span>
          <div className="flex items-center gap-2">
            <button
              onClick={handlePreviousPage}
              disabled={currentPage === 0}
              className="p-1.5 rounded hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              onClick={handleNextPage}
              disabled={currentPage >= totalPages - 1}
              className="p-1.5 rounded hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function LogRow({ log }: { log: AuditLog }) {
  const Icon = getEventIcon(log.event_type);
  const colorClass = getEventColor(log.success, log.event_type);
  const { browser, os } = parseUserAgent(log.user_agent);

  return (
    <tr className="border-b border-editor-border/50 hover:bg-editor-surface/50 transition-colors">
      <td className="py-2.5 px-3">
        <div className="text-editor-text">{formatRelativeTime(log.created_at)}</div>
        <div className="text-xs text-editor-muted">
          {new Date(log.created_at).toLocaleString()}
        </div>
      </td>
      <td className="py-2.5 px-3">
        <div className="flex items-center gap-2">
          <Icon className={`w-4 h-4 ${colorClass}`} />
          <span className={colorClass}>{getEventLabel(log.event_type)}</span>
        </div>
        {browser !== 'Unknown' && (
          <div className="text-xs text-editor-muted mt-0.5">
            {browser} on {os}
          </div>
        )}
      </td>
      <td className="py-2.5 px-3">
        <span className="px-2 py-0.5 text-xs bg-editor-surface rounded-full">
          {getCategoryLabel(log.event_category)}
        </span>
      </td>
      <td className="py-2.5 px-3">
        <span className="font-mono text-xs text-editor-muted">
          {log.ip_address || '-'}
        </span>
      </td>
      <td className="py-2.5 px-3">
        {log.success ? (
          <div className="flex items-center gap-1 text-green-400">
            <CheckCircle className="w-3.5 h-3.5" />
            <span className="text-xs">Success</span>
          </div>
        ) : (
          <div className="flex items-center gap-1 text-red-400">
            <XCircle className="w-3.5 h-3.5" />
            <span className="text-xs">Failed</span>
          </div>
        )}
      </td>
    </tr>
  );
}

export default AuditLogTable;
