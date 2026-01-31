import { useState, useRef, useEffect, useCallback } from 'react';
import { ChevronDown, Check, Search } from 'lucide-react';
import type { FilterOption } from '../../types/tasks';

interface TaskFilterDropdownProps<T extends string = string> {
  /** Label displayed on the button */
  label: string;
  /** Available options */
  options: FilterOption<T>[];
  /** Currently selected value(s) */
  value: T | T[];
  /** Callback when value changes */
  onChange: (value: T | T[]) => void;
  /** Enable multi-select mode */
  multiSelect?: boolean;
  /** Show search input for filtering options */
  searchable?: boolean;
  /** Placeholder for search input */
  searchPlaceholder?: string;
  /** Custom class name */
  className?: string;
}

export function TaskFilterDropdown<T extends string = string>({
  label,
  options,
  value,
  onChange,
  multiSelect = false,
  searchable = false,
  searchPlaceholder = 'Search...',
  className = '',
}: TaskFilterDropdownProps<T>) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Get selected values as array
  const selectedValues = Array.isArray(value) ? value : [value];

  // Filter options based on search
  const filteredOptions = searchable
    ? options.filter((opt) =>
        opt.label.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : options;

  // Get display label
  const getDisplayLabel = () => {
    if (multiSelect && selectedValues.length > 1) {
      return `${label} (${selectedValues.length})`;
    }
    const selectedOption = options.find((opt) => selectedValues.includes(opt.value));
    return selectedOption?.label || label;
  };

  // Handle option selection
  const handleSelect = useCallback(
    (optionValue: T) => {
      if (multiSelect) {
        const newValues = selectedValues.includes(optionValue)
          ? selectedValues.filter((v) => v !== optionValue)
          : [...selectedValues, optionValue];
        onChange(newValues as T[]);
      } else {
        onChange(optionValue);
        setIsOpen(false);
      }
    },
    [multiSelect, selectedValues, onChange]
  );

  // Handle keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!isOpen) {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
          e.preventDefault();
          setIsOpen(true);
          setFocusedIndex(0);
        }
        return;
      }

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setFocusedIndex((prev) =>
            prev < filteredOptions.length - 1 ? prev + 1 : prev
          );
          break;
        case 'ArrowUp':
          e.preventDefault();
          setFocusedIndex((prev) => (prev > 0 ? prev - 1 : 0));
          break;
        case 'Enter':
        case ' ':
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < filteredOptions.length) {
            handleSelect(filteredOptions[focusedIndex].value);
          }
          break;
        case 'Escape':
          e.preventDefault();
          setIsOpen(false);
          setFocusedIndex(-1);
          break;
        case 'Tab':
          setIsOpen(false);
          setFocusedIndex(-1);
          break;
      }
    },
    [isOpen, filteredOptions, focusedIndex, handleSelect]
  );

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setFocusedIndex(-1);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Focus search input when opened
  useEffect(() => {
    if (isOpen && searchable && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isOpen, searchable]);

  // Reset search and focus when closed
  useEffect(() => {
    if (!isOpen) {
      setSearchQuery('');
      setFocusedIndex(-1);
    }
  }, [isOpen]);

  const hasSelection = multiSelect
    ? selectedValues.length > 0
    : !selectedValues.includes('all' as T);

  return (
    <div
      ref={containerRef}
      className={`relative ${className}`}
      onKeyDown={handleKeyDown}
    >
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-2 px-3 py-2 bg-editor-bg border rounded-lg text-sm transition-colors ${
          hasSelection
            ? 'border-editor-accent text-editor-accent'
            : 'border-editor-border text-editor-text'
        } hover:border-editor-accent/50 focus:outline-none focus:border-editor-accent`}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <span>{getDisplayLabel()}</span>
        <ChevronDown
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div
          className="absolute z-50 mt-1 w-56 bg-editor-surface border border-editor-border rounded-lg shadow-lg overflow-hidden"
          role="listbox"
          aria-multiselectable={multiSelect}
        >
          {searchable && (
            <div className="p-2 border-b border-editor-border">
              <div className="relative">
                <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
                <input
                  ref={searchInputRef}
                  type="text"
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setFocusedIndex(0);
                  }}
                  placeholder={searchPlaceholder}
                  className="w-full pl-8 pr-3 py-1.5 bg-editor-bg border border-editor-border rounded text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                />
              </div>
            </div>
          )}

          <div className="max-h-60 overflow-y-auto py-1">
            {filteredOptions.length === 0 ? (
              <div className="px-3 py-2 text-sm text-editor-muted">
                No options found
              </div>
            ) : (
              filteredOptions.map((option, index) => {
                const isSelected = selectedValues.includes(option.value);
                const isFocused = index === focusedIndex;

                return (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => handleSelect(option.value)}
                    className={`w-full flex items-center justify-between px-3 py-2 text-sm transition-colors ${
                      isFocused
                        ? 'bg-editor-accent/10 text-editor-accent'
                        : isSelected
                        ? 'text-editor-accent'
                        : 'text-editor-text hover:bg-editor-accent/5'
                    }`}
                    role="option"
                    aria-selected={isSelected}
                  >
                    <span className="flex items-center gap-2">
                      {multiSelect && (
                        <span
                          className={`w-4 h-4 border rounded flex items-center justify-center ${
                            isSelected
                              ? 'bg-editor-accent border-editor-accent'
                              : 'border-editor-border'
                          }`}
                        >
                          {isSelected && (
                            <Check className="w-3 h-3 text-editor-bg" />
                          )}
                        </span>
                      )}
                      {option.label}
                    </span>
                    {option.count !== undefined && (
                      <span className="text-editor-muted text-xs">
                        {option.count}
                      </span>
                    )}
                    {!multiSelect && isSelected && (
                      <Check className="w-4 h-4 text-editor-accent" />
                    )}
                  </button>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}
