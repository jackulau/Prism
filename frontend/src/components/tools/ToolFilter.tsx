import { X, Cpu, Wrench, Zap, Box } from 'lucide-react';
import type { ToolType } from '../../types/tools';

interface ToolFilterProps {
  searchQuery: string;
  selectedProvider: string | null;
  selectedType: ToolType;
  providers: string[];
  onSearchChange: (query: string) => void;
  onProviderChange: (provider: string | null) => void;
  onTypeChange: (type: ToolType) => void;
  onReset: () => void;
}

export function ToolFilter({
  searchQuery,
  selectedProvider,
  selectedType,
  providers,
  onSearchChange,
  onProviderChange,
  onTypeChange,
  onReset,
}: ToolFilterProps) {
  const hasFilters = searchQuery || selectedProvider || selectedType !== 'all';

  const typeOptions: { value: ToolType; label: string; icon: typeof Box }[] = [
    { value: 'all', label: 'All', icon: Box },
    { value: 'model', label: 'Models', icon: Cpu },
    { value: 'builtin', label: 'Builtin', icon: Zap },
    { value: 'custom', label: 'Custom', icon: Wrench },
  ];

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="relative">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search tools..."
          className="w-full px-4 py-2.5 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent transition-colors"
        />
        {searchQuery && (
          <button
            onClick={() => onSearchChange('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-editor-muted hover:text-editor-text transition-colors"
          >
            <X size={16} />
          </button>
        )}
      </div>

      {/* Type filter */}
      <div className="flex flex-wrap gap-2">
        {typeOptions.map(({ value, label, icon: Icon }) => (
          <button
            key={value}
            onClick={() => onTypeChange(value)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm transition-colors ${
              selectedType === value
                ? 'bg-editor-accent text-white'
                : 'bg-editor-surface border border-editor-border text-editor-muted hover:text-editor-text hover:border-editor-accent/30'
            }`}
          >
            <Icon size={14} />
            {label}
          </button>
        ))}
      </div>

      {/* Provider filter */}
      {providers.length > 0 && (
        <div>
          <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-2">
            Provider
          </label>
          <select
            value={selectedProvider || ''}
            onChange={(e) => onProviderChange(e.target.value || null)}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent transition-colors"
          >
            <option value="">All Providers</option>
            {providers.map((provider) => (
              <option key={provider} value={provider}>
                {provider}
              </option>
            ))}
          </select>
        </div>
      )}

      {/* Clear filters */}
      {hasFilters && (
        <button
          onClick={onReset}
          className="flex items-center gap-2 px-3 py-2 w-full justify-center text-sm text-editor-muted hover:text-editor-text bg-editor-bg border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors"
        >
          <X size={14} />
          Clear Filters
        </button>
      )}
    </div>
  );
}
