import { useState, useEffect, useCallback } from 'react'
import { RefreshCw, Wrench, Filter, ChevronDown, X } from 'lucide-react'
import { ToolExecutionRow, type ToolExecution } from './ToolExecutionRow'

interface ToolExecutionLogsProps {
  initialToolName?: string
}

const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'pending', label: 'Pending' },
]

export function ToolExecutionLogs({ initialToolName = '' }: ToolExecutionLogsProps) {
  const [executions, setExecutions] = useState<ToolExecution[]>([])
  const [toolNames, setToolNames] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Filter state
  const [toolNameFilter, setToolNameFilter] = useState(initialToolName)
  const [statusFilter, setStatusFilter] = useState('')
  const [showFilters, setShowFilters] = useState(false)

  // Pagination state
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const limit = 25

  const fetchExecutions = useCallback(async (refresh = false) => {
    if (refresh) {
      setIsRefreshing(true)
    } else {
      setIsLoading(true)
    }
    setError(null)

    try {
      const params = new URLSearchParams()
      if (toolNameFilter) params.append('tool_name', toolNameFilter)
      if (statusFilter) params.append('status', statusFilter)
      params.append('limit', limit.toString())
      params.append('offset', offset.toString())

      const response = await fetch(`/api/v1/tools/executions?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('access_token')}`,
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch executions')
      }

      const data = await response.json()
      setExecutions(data.executions || [])
      setTotal(data.total || 0)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load executions')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [toolNameFilter, statusFilter, offset])

  const fetchToolNames = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/tools/executions/tool-names', {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('access_token')}`,
        },
      })

      if (response.ok) {
        const data = await response.json()
        setToolNames(data.tool_names || [])
      }
    } catch {
      // Silently fail for tool names
    }
  }, [])

  useEffect(() => {
    fetchExecutions()
  }, [fetchExecutions])

  useEffect(() => {
    fetchToolNames()
  }, [fetchToolNames])

  const handleRefresh = () => {
    fetchExecutions(true)
  }

  const handleFilterChange = () => {
    setOffset(0)
    fetchExecutions()
  }

  const clearFilters = () => {
    setToolNameFilter('')
    setStatusFilter('')
    setOffset(0)
  }

  const hasActiveFilters = toolNameFilter !== '' || statusFilter !== ''
  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit) + 1

  if (isLoading && !isRefreshing) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="bg-editor-surface border border-editor-border rounded-lg p-4 animate-pulse"
          >
            <div className="flex items-center gap-4">
              <div className="w-8 h-8 bg-editor-border rounded-lg" />
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-editor-border rounded w-1/4" />
                <div className="h-3 bg-editor-border rounded w-1/2" />
              </div>
              <div className="w-20 h-6 bg-editor-border rounded-full" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-editor-error/10 border border-editor-error/20 rounded-lg p-4 text-center">
        <p className="text-editor-error">{error}</p>
        <button
          onClick={() => fetchExecutions()}
          className="mt-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/80 transition-colors"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header with controls */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg border transition-colors ${
              hasActiveFilters
                ? 'border-editor-accent bg-editor-accent/10 text-editor-accent'
                : 'border-editor-border hover:bg-editor-surface text-editor-muted hover:text-editor-text'
            }`}
          >
            <Filter size={14} />
            <span className="text-sm">Filters</span>
            {hasActiveFilters && (
              <span className="px-1.5 py-0.5 text-xs bg-editor-accent text-white rounded-full">
                {(toolNameFilter ? 1 : 0) + (statusFilter ? 1 : 0)}
              </span>
            )}
            <ChevronDown size={14} className={showFilters ? 'rotate-180' : ''} />
          </button>

          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="flex items-center gap-1 px-2 py-1.5 text-sm text-editor-muted hover:text-editor-text transition-colors"
            >
              <X size={14} />
              Clear
            </button>
          )}
        </div>

        <div className="flex items-center gap-2">
          <span className="text-sm text-editor-muted">
            {total} execution{total !== 1 ? 's' : ''}
          </span>
          <button
            onClick={handleRefresh}
            disabled={isRefreshing}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-editor-border hover:bg-editor-surface transition-colors disabled:opacity-50"
          >
            <RefreshCw size={14} className={isRefreshing ? 'animate-spin' : ''} />
            <span className="text-sm">Refresh</span>
          </button>
        </div>
      </div>

      {/* Filter controls */}
      {showFilters && (
        <div className="flex flex-wrap items-center gap-4 p-4 bg-editor-surface border border-editor-border rounded-lg">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-editor-muted">Tool Name</label>
            <select
              value={toolNameFilter}
              onChange={(e) => {
                setToolNameFilter(e.target.value)
                handleFilterChange()
              }}
              className="px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
            >
              <option value="">All Tools</option>
              {toolNames.map((name) => (
                <option key={name} value={name}>
                  {name}
                </option>
              ))}
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-editor-muted">Status</label>
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value)
                handleFilterChange()
              }}
              className="px-3 py-1.5 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
            >
              {STATUS_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      )}

      {/* Execution list */}
      {executions.length === 0 ? (
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <Wrench className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">No tool executions</h3>
          <p className="text-editor-muted">
            {hasActiveFilters
              ? 'No executions match your filters. Try adjusting or clearing the filters.'
              : 'Tool executions will appear here when tools are used.'}
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {executions.map((execution) => (
            <ToolExecutionRow
              key={execution.id}
              execution={execution}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-4 border-t border-editor-border">
          <span className="text-sm text-editor-muted">
            Page {currentPage} of {totalPages}
          </span>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setOffset(Math.max(0, offset - limit))}
              disabled={offset === 0}
              className="px-3 py-1.5 text-sm border border-editor-border rounded-lg hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              onClick={() => setOffset(offset + limit)}
              disabled={offset + limit >= total}
              className="px-3 py-1.5 text-sm border border-editor-border rounded-lg hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default ToolExecutionLogs
