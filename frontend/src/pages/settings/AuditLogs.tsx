import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  Filter,
  Calendar,
  Download,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  X,
  AlertTriangle,
} from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useAuditStore } from '../../store/auditStore';
import { auditService } from '../../services/audit';
import { AuditLogEntry } from '../../components/audit/AuditLogEntry';
import type { AuditActionType, AuditLogFilter } from '../../types/audit';

const actionTypeOptions: { value: AuditActionType; label: string }[] = [
  { value: 'user.login', label: 'User Login' },
  { value: 'user.logout', label: 'User Logout' },
  { value: 'user.create', label: 'User Created' },
  { value: 'user.update', label: 'User Updated' },
  { value: 'user.delete', label: 'User Deleted' },
  { value: 'workspace.create', label: 'Workspace Created' },
  { value: 'workspace.update', label: 'Workspace Updated' },
  { value: 'workspace.delete', label: 'Workspace Deleted' },
  { value: 'conversation.create', label: 'Conversation Created' },
  { value: 'conversation.delete', label: 'Conversation Deleted' },
  { value: 'api_key.create', label: 'API Key Created' },
  { value: 'api_key.revoke', label: 'API Key Revoked' },
  { value: 'settings.update', label: 'Settings Updated' },
  { value: 'export.request', label: 'Export Requested' },
  { value: 'export.download', label: 'Export Downloaded' },
  { value: 'member.invite', label: 'Member Invited' },
  { value: 'member.remove', label: 'Member Removed' },
  { value: 'role.change', label: 'Role Changed' },
];

const resourceTypeOptions = [
  { value: 'user', label: 'User' },
  { value: 'workspace', label: 'Workspace' },
  { value: 'conversation', label: 'Conversation' },
  { value: 'api_key', label: 'API Key' },
  { value: 'settings', label: 'Settings' },
  { value: 'export', label: 'Export' },
  { value: 'member', label: 'Member' },
];

export default function AuditLogs() {
  const { accessToken } = useAuthStore();
  const {
    auditLogs,
    auditLogsTotal,
    auditLogsPage,
    auditLogsPageSize,
    auditLogsHasMore,
    auditFilter,
    auditLogsLoading,
    auditLogsError,
    setAuditLogs,
    setAuditFilter,
    setAuditLogsPage,
    setAuditLogsLoading,
    setAuditLogsError,
  } = useAuditStore();

  const [showFilters, setShowFilters] = useState(false);
  const [searchQuery, setSearchQuery] = useState(auditFilter.search || '');
  const [selectedActions, setSelectedActions] = useState<AuditActionType[]>(auditFilter.actions || []);
  const [selectedResourceType, setSelectedResourceType] = useState(auditFilter.resourceType || '');
  const [startDate, setStartDate] = useState(
    auditFilter.startDate ? new Date(auditFilter.startDate).toISOString().split('T')[0] : ''
  );
  const [endDate, setEndDate] = useState(
    auditFilter.endDate ? new Date(auditFilter.endDate).toISOString().split('T')[0] : ''
  );
  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    if (accessToken) {
      auditService.setToken(accessToken);
    }
  }, [accessToken]);

  const fetchLogs = useCallback(async (filter: AuditLogFilter, page: number) => {
    setAuditLogsLoading(true);
    setAuditLogsError(null);

    const response = await auditService.getAuditLogs(filter, page, auditLogsPageSize);
    if (response.error) {
      setAuditLogsError(response.error);
    } else if (response.data) {
      setAuditLogs(response.data.entries, response.data.total, response.data.hasMore);
    }
    setAuditLogsLoading(false);
  }, [auditLogsPageSize, setAuditLogs, setAuditLogsLoading, setAuditLogsError]);

  useEffect(() => {
    fetchLogs(auditFilter, auditLogsPage);
  }, [auditFilter, auditLogsPage, fetchLogs]);

  const applyFilters = () => {
    const newFilter: AuditLogFilter = {
      search: searchQuery || undefined,
      actions: selectedActions.length > 0 ? selectedActions : undefined,
      resourceType: selectedResourceType || undefined,
      startDate: startDate ? new Date(startDate) : undefined,
      endDate: endDate ? new Date(endDate) : undefined,
    };
    setAuditFilter(newFilter);
    setShowFilters(false);
  };

  const clearFilters = () => {
    setSearchQuery('');
    setSelectedActions([]);
    setSelectedResourceType('');
    setStartDate('');
    setEndDate('');
    setAuditFilter({});
    setShowFilters(false);
  };

  const handleExport = async () => {
    setIsExporting(true);
    const response = await auditService.exportAuditLogs(auditFilter, 'csv');
    if (response.data?.downloadUrl) {
      window.open(response.data.downloadUrl, '_blank');
    }
    setIsExporting(false);
  };

  const handleResourceClick = (resourceType: string, resourceId: string) => {
    setSelectedResourceType(resourceType);
    setAuditFilter({ ...auditFilter, resourceType, resourceId });
  };

  const totalPages = Math.ceil(auditLogsTotal / auditLogsPageSize);
  const hasActiveFilters =
    auditFilter.search ||
    auditFilter.actions?.length ||
    auditFilter.resourceType ||
    auditFilter.startDate ||
    auditFilter.endDate;

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-editor-text">Audit Logs</h1>
            <p className="text-editor-muted mt-1">
              View and search activity logs across your organization
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => fetchLogs(auditFilter, auditLogsPage)}
              disabled={auditLogsLoading}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg transition-colors disabled:opacity-50"
            >
              <RefreshCw size={16} className={auditLogsLoading ? 'animate-spin' : ''} />
              Refresh
            </button>
            <button
              onClick={handleExport}
              disabled={isExporting || auditLogs.length === 0}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors disabled:opacity-50"
            >
              <Download size={16} />
              Export
            </button>
          </div>
        </div>

        {/* Search and Filters Bar */}
        <div className="flex items-center gap-3">
          {/* Search Input */}
          <div className="flex-1 relative">
            <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  applyFilters();
                }
              }}
              placeholder="Search by user, action, or resource..."
              className="w-full pl-10 pr-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent"
            />
          </div>

          {/* Filter Toggle */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-2 px-3 py-2 text-sm rounded-lg border transition-colors ${
              hasActiveFilters
                ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                : 'bg-editor-surface border-editor-border text-editor-text hover:bg-editor-bg'
            }`}
          >
            <Filter size={16} />
            Filters
            {hasActiveFilters && (
              <span className="px-1.5 py-0.5 text-xs bg-editor-accent text-white rounded-full">
                {[
                  auditFilter.actions?.length || 0,
                  auditFilter.resourceType ? 1 : 0,
                  auditFilter.startDate || auditFilter.endDate ? 1 : 0,
                ].reduce((a, b) => a + b, 0)}
              </span>
            )}
          </button>
        </div>

        {/* Expanded Filters */}
        {showFilters && (
          <div className="bg-editor-surface border border-editor-border rounded-lg p-4 space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Date Range */}
              <div className="space-y-2">
                <label className="flex items-center gap-1.5 text-sm font-medium text-editor-text">
                  <Calendar size={14} />
                  Date Range
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="flex-1 px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text"
                  />
                  <span className="text-editor-muted">to</span>
                  <input
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    className="flex-1 px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text"
                  />
                </div>
              </div>

              {/* Action Types */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-editor-text">Action Types</label>
                <select
                  multiple
                  value={selectedActions}
                  onChange={(e) =>
                    setSelectedActions(
                      Array.from(e.target.selectedOptions, (opt) => opt.value as AuditActionType)
                    )
                  }
                  className="w-full px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text h-24"
                >
                  {actionTypeOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Resource Type */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-editor-text">Resource Type</label>
                <select
                  value={selectedResourceType}
                  onChange={(e) => setSelectedResourceType(e.target.value)}
                  className="w-full px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text"
                >
                  <option value="">All Resources</option>
                  {resourceTypeOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Filter Actions */}
            <div className="flex items-center justify-end gap-2 pt-2 border-t border-editor-border">
              <button
                onClick={clearFilters}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text"
              >
                <X size={14} />
                Clear All
              </button>
              <button
                onClick={applyFilters}
                className="px-4 py-1.5 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90"
              >
                Apply Filters
              </button>
            </div>
          </div>
        )}

        {/* Error Message */}
        {auditLogsError && (
          <div className="flex items-center gap-2 p-4 bg-editor-error/10 border border-editor-error/20 rounded-lg">
            <AlertTriangle size={18} className="text-editor-error" />
            <span className="text-editor-error">{auditLogsError}</span>
          </div>
        )}

        {/* Results Summary */}
        <div className="flex items-center justify-between text-sm text-editor-muted">
          <span>
            Showing {auditLogs.length} of {auditLogsTotal.toLocaleString()} entries
          </span>
          {hasActiveFilters && (
            <button onClick={clearFilters} className="text-editor-accent hover:underline">
              Clear filters
            </button>
          )}
        </div>

        {/* Audit Log List */}
        <div className="space-y-3">
          {auditLogsLoading && auditLogs.length === 0 ? (
            Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="h-20 bg-editor-surface border border-editor-border rounded-lg animate-pulse"
              />
            ))
          ) : auditLogs.length === 0 ? (
            <div className="text-center py-12 bg-editor-surface border border-editor-border rounded-lg">
              <Search size={48} className="mx-auto mb-4 text-editor-muted opacity-50" />
              <p className="text-editor-muted">No audit logs found</p>
              {hasActiveFilters && (
                <p className="text-sm text-editor-muted mt-1">
                  Try adjusting your filters or search query
                </p>
              )}
            </div>
          ) : (
            auditLogs.map((entry) => (
              <AuditLogEntry
                key={entry.id}
                entry={entry}
                onResourceClick={handleResourceClick}
              />
            ))
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-4 border-t border-editor-border">
            <span className="text-sm text-editor-muted">
              Page {auditLogsPage} of {totalPages}
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setAuditLogsPage(auditLogsPage - 1)}
                disabled={auditLogsPage <= 1 || auditLogsLoading}
                className="flex items-center gap-1 px-3 py-1.5 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={16} />
                Previous
              </button>
              <button
                onClick={() => setAuditLogsPage(auditLogsPage + 1)}
                disabled={!auditLogsHasMore || auditLogsLoading}
                className="flex items-center gap-1 px-3 py-1.5 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
