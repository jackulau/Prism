import { useState } from 'react';
import {
  FileJson,
  FileSpreadsheet,
  CheckCircle2,
  XCircle,
  Clock,
  Zap,
  Filter,
  ArrowUpDown,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { useBatchStore } from '../../store/batchStore';
import type { BatchTask, BatchTaskStatus } from '../../types/batch';

type SortField = 'taskId' | 'status' | 'duration' | 'tokensUsed';
type SortDirection = 'asc' | 'desc';

export function BatchResults() {
  const { getCompletedTasks, getFailedTasks, reset } = useBatchStore();
  const completedTasks = getCompletedTasks();
  const failedTasks = getFailedTasks();
  const results = [...completedTasks, ...failedTasks];

  const [statusFilter, setStatusFilter] = useState<BatchTaskStatus | 'all'>('all');
  const [sortField, setSortField] = useState<SortField>('taskId');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const filteredResults = results.filter((r: BatchTask) =>
    statusFilter === 'all' ? true : r.status === statusFilter
  );

  const sortedResults = [...filteredResults].sort((a: BatchTask, b: BatchTask) => {
    let comparison = 0;
    switch (sortField) {
      case 'taskId':
        comparison = a.id.localeCompare(b.id);
        break;
      case 'status':
        comparison = a.status.localeCompare(b.status);
        break;
      case 'duration': {
        const aDuration = a.startedAt && a.completedAt
          ? new Date(a.completedAt).getTime() - new Date(a.startedAt).getTime()
          : 0;
        const bDuration = b.startedAt && b.completedAt
          ? new Date(b.completedAt).getTime() - new Date(b.startedAt).getTime()
          : 0;
        comparison = aDuration - bDuration;
        break;
      }
      case 'tokensUsed':
        comparison = (a.tokenUsage?.total || 0) - (b.tokenUsage?.total || 0);
        break;
    }
    return sortDirection === 'asc' ? comparison : -comparison;
  });

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const exportResults = (format: 'json' | 'csv'): string => {
    if (format === 'json') {
      return JSON.stringify(results.map((r: BatchTask) => ({
        id: r.id,
        name: r.name,
        prompt: r.prompt,
        status: r.status,
        output: r.output,
        error: r.error,
        tokenUsage: r.tokenUsage,
        startedAt: r.startedAt,
        completedAt: r.completedAt,
      })), null, 2);
    }

    // CSV format
    const headers = ['ID', 'Name', 'Status', 'Duration (ms)', 'Tokens', 'Error'];
    const rows = results.map((r: BatchTask) => {
      const duration = r.startedAt && r.completedAt
        ? new Date(r.completedAt).getTime() - new Date(r.startedAt).getTime()
        : '';
      return [
        r.id,
        r.name,
        r.status,
        duration,
        r.tokenUsage?.total || '',
        r.error || '',
      ].map((v) => `"${String(v).replace(/"/g, '""')}"`).join(',');
    });
    return [headers.join(','), ...rows].join('\n');
  };

  const handleExport = (format: 'json' | 'csv') => {
    const content = exportResults(format);
    const blob = new Blob([content], {
      type: format === 'json' ? 'application/json' : 'text/csv',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `batch-results-${new Date().toISOString().slice(0, 10)}.${format}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const toggleRow = (taskId: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev);
      if (next.has(taskId)) {
        next.delete(taskId);
      } else {
        next.add(taskId);
      }
      return next;
    });
  };

  const getStatusIcon = (status: BatchTaskStatus) => {
    switch (status) {
      case 'completed':
        return <CheckCircle2 size={14} className="text-editor-success" />;
      case 'failed':
        return <XCircle size={14} className="text-editor-error" />;
      case 'cancelled':
        return <XCircle size={14} className="text-editor-warning" />;
      default:
        return <Clock size={14} className="text-editor-muted" />;
    }
  };

  if (results.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-editor-muted">
        <FileSpreadsheet size={48} className="mb-4 opacity-50" />
        <p className="text-sm">No results yet</p>
        <p className="text-xs mt-1">Run a batch to see results here</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header with stats and actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h3 className="text-sm font-semibold text-editor-text">
            Results ({results.length})
          </h3>
          <div className="flex items-center gap-2 text-xs">
            <span className="flex items-center gap-1 text-editor-success">
              <CheckCircle2 size={12} />
              {completedTasks.length}
            </span>
            <span className="flex items-center gap-1 text-editor-error">
              <XCircle size={12} />
              {failedTasks.length}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={reset}
            className="px-2 py-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
          >
            Clear
          </button>
          <button
            onClick={() => handleExport('json')}
            className="flex items-center gap-1 px-2 py-1 text-xs bg-editor-surface border border-editor-border rounded hover:bg-editor-accent/10 transition-colors"
            title="Export as JSON"
          >
            <FileJson size={12} />
            JSON
          </button>
          <button
            onClick={() => handleExport('csv')}
            className="flex items-center gap-1 px-2 py-1 text-xs bg-editor-surface border border-editor-border rounded hover:bg-editor-accent/10 transition-colors"
            title="Export as CSV"
          >
            <FileSpreadsheet size={12} />
            CSV
          </button>
        </div>
      </div>

      {/* Filter */}
      <div className="flex items-center gap-2">
        <Filter size={14} className="text-editor-muted" />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as BatchTaskStatus | 'all')}
          className="px-2 py-1 text-xs bg-editor-surface border border-editor-border rounded text-editor-text focus:border-editor-accent focus:outline-none"
        >
          <option value="all">All Status</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      {/* Results Table */}
      <div className="border border-editor-border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-editor-surface">
            <tr>
              <th className="w-8 px-2 py-2"></th>
              <th
                className="px-3 py-2 text-left text-xs font-medium text-editor-muted cursor-pointer hover:text-editor-text"
                onClick={() => handleSort('status')}
              >
                <span className="flex items-center gap-1">
                  Status
                  <ArrowUpDown size={10} />
                </span>
              </th>
              <th className="px-3 py-2 text-left text-xs font-medium text-editor-muted">
                Prompt
              </th>
              <th
                className="px-3 py-2 text-right text-xs font-medium text-editor-muted cursor-pointer hover:text-editor-text"
                onClick={() => handleSort('duration')}
              >
                <span className="flex items-center justify-end gap-1">
                  <Clock size={10} />
                  Duration
                  <ArrowUpDown size={10} />
                </span>
              </th>
              <th
                className="px-3 py-2 text-right text-xs font-medium text-editor-muted cursor-pointer hover:text-editor-text"
                onClick={() => handleSort('tokensUsed')}
              >
                <span className="flex items-center justify-end gap-1">
                  <Zap size={10} />
                  Tokens
                  <ArrowUpDown size={10} />
                </span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-editor-border">
            {sortedResults.map((result: BatchTask) => (
              <ResultRow
                key={result.id}
                result={result}
                isExpanded={expandedRows.has(result.id)}
                onToggle={() => toggleRow(result.id)}
                getStatusIcon={getStatusIcon}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

interface ResultRowProps {
  result: BatchTask;
  isExpanded: boolean;
  onToggle: () => void;
  getStatusIcon: (status: BatchTaskStatus) => JSX.Element;
}

function ResultRow({ result, isExpanded, onToggle, getStatusIcon }: ResultRowProps) {
  const duration = result.startedAt && result.completedAt
    ? new Date(result.completedAt).getTime() - new Date(result.startedAt).getTime()
    : null;

  return (
    <>
      <tr
        className="hover:bg-editor-surface/50 cursor-pointer"
        onClick={onToggle}
      >
        <td className="px-2 py-2">
          {isExpanded ? (
            <ChevronDown size={14} className="text-editor-muted" />
          ) : (
            <ChevronRight size={14} className="text-editor-muted" />
          )}
        </td>
        <td className="px-3 py-2">
          <span className="flex items-center gap-1">
            {getStatusIcon(result.status)}
            <span className="text-xs capitalize">{result.status}</span>
          </span>
        </td>
        <td className="px-3 py-2 text-editor-text max-w-[200px] truncate">
          {result.prompt}
        </td>
        <td className="px-3 py-2 text-right text-editor-muted text-xs">
          {duration ? `${(duration / 1000).toFixed(1)}s` : '-'}
        </td>
        <td className="px-3 py-2 text-right text-editor-muted text-xs">
          {result.tokenUsage?.total?.toLocaleString() || '-'}
        </td>
      </tr>
      {isExpanded && (
        <tr className="bg-editor-surface/30">
          <td colSpan={5} className="px-4 py-3">
            <div className="space-y-2">
              <div>
                <span className="text-xs text-editor-muted block mb-1">Full Prompt:</span>
                <p className="text-sm text-editor-text bg-editor-bg p-2 rounded border border-editor-border">
                  {result.prompt}
                </p>
              </div>
              {result.output && (
                <div>
                  <span className="text-xs text-editor-muted block mb-1">Result:</span>
                  <p className="text-sm text-editor-text bg-editor-bg p-2 rounded border border-editor-border whitespace-pre-wrap">
                    {result.output}
                  </p>
                </div>
              )}
              {result.error && (
                <div>
                  <span className="text-xs text-editor-error block mb-1">Error:</span>
                  <p className="text-sm text-editor-error bg-editor-error/10 p-2 rounded border border-editor-error/30">
                    {result.error}
                  </p>
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  );
}
