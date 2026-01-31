import { useState, useRef, useEffect } from 'react'
import { ChevronDown, User, Wrench, X, Filter } from 'lucide-react'
import type { AgentInfo } from '../../store/sandboxStore'

export interface HistoryFilters {
  agentId?: string | null
  agentName?: string | null
  toolName?: string | null
}

export interface HistoryFiltersProps {
  filters: HistoryFilters
  agentList: AgentInfo[]
  toolList: string[]
  onChange: (filters: HistoryFilters) => void
}

interface DropdownProps {
  label: string
  icon: React.ReactNode
  value: string | null | undefined
  options: Array<{ value: string; label: string }>
  onSelect: (value: string | null) => void
  placeholder?: string
}

function Dropdown({ label, icon, value, options, onSelect, placeholder = 'All' }: DropdownProps) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const selectedOption = options.find(o => o.value === value)
  const displayValue = selectedOption?.label || placeholder

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-1.5 px-2 py-1 text-xs rounded-md border transition-colors ${
          value
            ? 'bg-editor-accent/10 border-editor-accent/50 text-editor-accent'
            : 'bg-editor-surface border-editor-border text-editor-muted hover:border-editor-accent/50'
        }`}
      >
        {icon}
        <span className="max-w-[80px] truncate">{displayValue}</span>
        <ChevronDown size={12} className={`transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-48 bg-editor-bg border border-editor-border rounded-md shadow-lg z-50 max-h-48 overflow-y-auto">
          <div className="p-1 border-b border-editor-border">
            <div className="px-2 py-1 text-xs text-editor-muted">{label}</div>
          </div>
          <button
            onClick={() => {
              onSelect(null)
              setIsOpen(false)
            }}
            className={`w-full px-2 py-1.5 text-left text-xs hover:bg-editor-surface transition-colors ${
              !value ? 'text-editor-accent bg-editor-accent/10' : 'text-editor-text'
            }`}
          >
            {placeholder}
          </button>
          {options.map((option) => (
            <button
              key={option.value}
              onClick={() => {
                onSelect(option.value)
                setIsOpen(false)
              }}
              className={`w-full px-2 py-1.5 text-left text-xs hover:bg-editor-surface transition-colors truncate ${
                value === option.value ? 'text-editor-accent bg-editor-accent/10' : 'text-editor-text'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function HistoryFiltersBar({
  filters,
  agentList,
  toolList,
  onChange,
}: HistoryFiltersProps) {
  const hasActiveFilters = filters.agentId || filters.toolName

  const handleClearAll = () => {
    onChange({
      agentId: null,
      agentName: null,
      toolName: null,
    })
  }

  const agentOptions = agentList.map(agent => ({
    value: agent.id,
    label: agent.name,
  }))

  const toolOptions = toolList.map(tool => ({
    value: tool,
    label: tool,
  }))

  return (
    <div className="flex items-center gap-2 px-3 py-2 border-b border-editor-border bg-editor-surface/30">
      <Filter size={12} className="text-editor-muted" />

      <Dropdown
        label="Filter by Agent"
        icon={<User size={12} />}
        value={filters.agentId}
        options={agentOptions}
        onSelect={(value) => {
          const agent = agentList.find(a => a.id === value)
          onChange({
            ...filters,
            agentId: value,
            agentName: agent?.name || null,
          })
        }}
        placeholder="All agents"
      />

      <Dropdown
        label="Filter by Tool"
        icon={<Wrench size={12} />}
        value={filters.toolName}
        options={toolOptions}
        onSelect={(value) => onChange({ ...filters, toolName: value })}
        placeholder="All tools"
      />

      {hasActiveFilters && (
        <button
          onClick={handleClearAll}
          className="flex items-center gap-1 px-1.5 py-0.5 text-xs text-editor-muted hover:text-editor-error transition-colors"
          title="Clear all filters"
        >
          <X size={12} />
          <span>Clear</span>
        </button>
      )}
    </div>
  )
}
