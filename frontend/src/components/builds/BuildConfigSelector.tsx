import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Settings2, Star, Check } from 'lucide-react';
import type { BuildConfig } from '../../services/buildConfig';

interface BuildConfigSelectorProps {
  configs: BuildConfig[];
  selectedId: string | null;
  onSelect: (config: BuildConfig) => void;
  onManage?: () => void;
  disabled?: boolean;
  compact?: boolean;
}

export function BuildConfigSelector({
  configs,
  selectedId,
  onSelect,
  onManage,
  disabled = false,
  compact = false,
}: BuildConfigSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedConfig = configs.find((c) => c.id === selectedId);
  const defaultConfig = configs.find((c) => c.isDefault);
  const displayConfig = selectedConfig || defaultConfig;

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const sortedConfigs = [...configs].sort((a, b) => {
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    return a.name.localeCompare(b.name);
  });

  if (configs.length === 0) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted">
        <Settings2 size={14} />
        <span>No build configs</span>
        {onManage && (
          <button
            onClick={onManage}
            className="text-editor-accent hover:underline"
          >
            Create one
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`flex items-center gap-2 rounded-lg border transition-colors ${
          compact
            ? 'px-2 py-1 text-xs'
            : 'px-3 py-2 text-sm'
        } ${
          isOpen
            ? 'border-editor-accent bg-editor-accent/10'
            : 'border-editor-border bg-editor-surface hover:border-editor-muted'
        } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
      >
        <Settings2 size={compact ? 12 : 14} className="text-editor-muted" />
        <span className="text-editor-text truncate max-w-[150px]">
          {displayConfig?.name || 'Select config'}
        </span>
        {displayConfig?.isDefault && (
          <Star size={compact ? 10 : 12} className="text-editor-warning" />
        )}
        <ChevronDown
          size={compact ? 12 : 14}
          className={`text-editor-muted transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-editor-bg border border-editor-border rounded-lg shadow-lg z-50 py-1">
          <div className="px-3 py-2 text-xs font-medium text-editor-muted uppercase tracking-wider border-b border-editor-border">
            Build Configuration
          </div>

          <div className="max-h-64 overflow-y-auto">
            {sortedConfigs.map((config) => (
              <button
                key={config.id}
                onClick={() => {
                  onSelect(config);
                  setIsOpen(false);
                }}
                className={`w-full flex items-center gap-2 px-3 py-2 text-sm text-left transition-colors ${
                  selectedId === config.id
                    ? 'bg-editor-accent/10 text-editor-accent'
                    : 'text-editor-text hover:bg-editor-surface'
                }`}
              >
                <div className="w-4">
                  {selectedId === config.id && <Check size={14} />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="truncate">{config.name}</span>
                    {config.isDefault && (
                      <span className="flex items-center gap-0.5 px-1 py-0.5 text-[10px] bg-editor-warning/20 text-editor-warning rounded">
                        <Star size={8} />
                        Default
                      </span>
                    )}
                  </div>
                  {config.description && (
                    <p className="text-xs text-editor-muted truncate mt-0.5">
                      {config.description}
                    </p>
                  )}
                </div>
              </button>
            ))}
          </div>

          {onManage && (
            <>
              <div className="border-t border-editor-border" />
              <button
                onClick={() => {
                  onManage();
                  setIsOpen(false);
                }}
                className="w-full flex items-center gap-2 px-3 py-2 text-sm text-editor-accent hover:bg-editor-surface transition-colors"
              >
                <Settings2 size={14} />
                Manage Configurations
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
