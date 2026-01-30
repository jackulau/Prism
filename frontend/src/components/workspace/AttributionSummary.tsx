import { User, Bot, Wrench, BarChart3, TrendingUp } from 'lucide-react'
import type { AttributionSummary as AttributionSummaryType } from '../../store/sandboxStore'

export interface AttributionSummaryProps {
  summary: AttributionSummaryType
  onAgentClick?: (agentId: string) => void
  onToolClick?: (toolName: string) => void
}

interface BarChartProps {
  data: Record<string, number>
  onItemClick?: (key: string) => void
  colorFn?: (key: string) => string
}

function HorizontalBarChart({ data, onItemClick, colorFn }: BarChartProps) {
  const entries = Object.entries(data).sort((a, b) => b[1] - a[1]).slice(0, 5)
  const maxValue = Math.max(...entries.map(([, v]) => v), 1)

  return (
    <div className="space-y-1.5">
      {entries.map(([key, value]) => {
        const percentage = (value / maxValue) * 100
        const color = colorFn?.(key) || 'bg-editor-accent'

        return (
          <button
            key={key}
            onClick={() => onItemClick?.(key)}
            className="w-full group"
          >
            <div className="flex items-center justify-between text-xs mb-0.5">
              <span className="text-editor-text truncate max-w-[120px] group-hover:text-editor-accent transition-colors">
                {key}
              </span>
              <span className="text-editor-muted">{value}</span>
            </div>
            <div className="h-1.5 bg-editor-surface rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${color} group-hover:opacity-80`}
                style={{ width: `${percentage}%` }}
              />
            </div>
          </button>
        )
      })}
    </div>
  )
}

interface SparklineProps {
  data: Record<string, number>
}

function Sparkline({ data }: SparklineProps) {
  const entries = Object.entries(data).slice(-14) // Last 14 days
  if (entries.length === 0) return null

  const values = entries.map(([, v]) => v)
  const maxValue = Math.max(...values, 1)
  const minValue = Math.min(...values)
  const range = maxValue - minValue || 1

  const points = values
    .map((value, index) => {
      const x = (index / (values.length - 1)) * 100
      const y = 100 - ((value - minValue) / range) * 100
      return `${x},${y}`
    })
    .join(' ')

  return (
    <div className="h-8 w-full">
      <svg
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        className="w-full h-full"
      >
        <polyline
          points={points}
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className="text-editor-accent"
        />
      </svg>
    </div>
  )
}

function getAgentColor(agentName: string): string {
  // Simple hash to generate consistent colors per agent
  const hash = agentName.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  const colors = [
    'bg-blue-500',
    'bg-purple-500',
    'bg-green-500',
    'bg-cyan-500',
    'bg-pink-500',
    'bg-orange-500',
  ]
  return colors[hash % colors.length]
}

function getToolColor(_toolName: string): string {
  return 'bg-editor-muted'
}

export function AttributionSummary({
  summary,
  onAgentClick,
  onToolClick,
}: AttributionSummaryProps) {
  const hasAgentData = Object.keys(summary.by_agent).length > 0
  const hasToolData = Object.keys(summary.by_tool).length > 0
  const hasTimeline = Object.keys(summary.timeline_by_day).length > 0

  return (
    <div className="p-3 space-y-4">
      {/* Total Changes */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-editor-text">
          <BarChart3 size={16} className="text-editor-accent" />
          <span>Total Changes</span>
        </div>
        <span className="text-lg font-semibold text-editor-text">
          {summary.total_changes.toLocaleString()}
        </span>
      </div>

      {/* Activity Timeline */}
      {hasTimeline && (
        <div>
          <div className="flex items-center gap-2 mb-2 text-xs text-editor-muted">
            <TrendingUp size={12} />
            <span>Activity (14 days)</span>
          </div>
          <Sparkline data={summary.timeline_by_day} />
        </div>
      )}

      {/* Most Active */}
      {(summary.most_active_agent || summary.most_used_tool) && (
        <div className="flex gap-4">
          {summary.most_active_agent && (
            <div className="flex-1">
              <div className="text-xs text-editor-muted mb-1">Most Active</div>
              <div className="flex items-center gap-1.5">
                <User size={14} className="text-blue-400" />
                <span className="text-sm text-editor-text truncate">
                  {summary.most_active_agent}
                </span>
              </div>
            </div>
          )}
          {summary.most_used_tool && (
            <div className="flex-1">
              <div className="text-xs text-editor-muted mb-1">Top Tool</div>
              <div className="flex items-center gap-1.5">
                <Wrench size={14} className="text-editor-muted" />
                <span className="text-sm text-editor-text truncate">
                  {summary.most_used_tool}
                </span>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Changes by Agent */}
      {hasAgentData && (
        <div>
          <div className="flex items-center gap-2 mb-2 text-xs text-editor-muted">
            <Bot size={12} />
            <span>By Agent</span>
          </div>
          <HorizontalBarChart
            data={summary.by_agent}
            onItemClick={onAgentClick}
            colorFn={getAgentColor}
          />
        </div>
      )}

      {/* Changes by Tool */}
      {hasToolData && (
        <div>
          <div className="flex items-center gap-2 mb-2 text-xs text-editor-muted">
            <Wrench size={12} />
            <span>By Tool</span>
          </div>
          <HorizontalBarChart
            data={summary.by_tool}
            onItemClick={onToolClick}
            colorFn={getToolColor}
          />
        </div>
      )}

      {/* Empty state */}
      {!hasAgentData && !hasToolData && summary.total_changes === 0 && (
        <div className="text-center py-4 text-sm text-editor-muted">
          No attribution data available
        </div>
      )}
    </div>
  )
}
