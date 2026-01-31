import { useState } from 'react';
import {
  RefreshCw,
  Download,
  ChevronDown,
  Calendar,
  FileJson,
  FileSpreadsheet,
} from 'lucide-react';

interface TimeRangeOption {
  label: string;
  value: string;
  days: number;
}

const TIME_RANGE_OPTIONS: TimeRangeOption[] = [
  { label: 'Last 24 hours', value: '24h', days: 1 },
  { label: 'Last 7 days', value: '7d', days: 7 },
  { label: 'Last 30 days', value: '30d', days: 30 },
  { label: 'Last 90 days', value: '90d', days: 90 },
];

export interface ResultsHeaderProps {
  onRefresh?: () => void;
  onExport?: (format: 'csv' | 'json') => void;
  onTimeRangeChange?: (start: Date, end: Date) => void;
  isLoading?: boolean;
  selectedTimeRange?: string;
}

export function ResultsHeader({
  onRefresh,
  onExport,
  onTimeRangeChange,
  isLoading = false,
  selectedTimeRange = '7d',
}: ResultsHeaderProps) {
  const [showExportMenu, setShowExportMenu] = useState(false);
  const [showTimeRangeMenu, setShowTimeRangeMenu] = useState(false);

  const handleExport = (format: 'csv' | 'json') => {
    onExport?.(format);
    setShowExportMenu(false);
  };

  const handleTimeRangeSelect = (option: TimeRangeOption) => {
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - option.days);
    onTimeRangeChange?.(start, end);
    setShowTimeRangeMenu(false);
  };

  const currentTimeRange = TIME_RANGE_OPTIONS.find(
    (opt) => opt.value === selectedTimeRange
  ) || TIME_RANGE_OPTIONS[1];

  return (
    <div className="flex items-center justify-between">
      <div className="space-y-1">
        <h1 className="text-2xl font-bold text-editor-text">Agent Results</h1>
        <p className="text-editor-muted">
          View and analyze execution results from batch and swarm operations
        </p>
      </div>

      <div className="flex items-center gap-3">
        {/* Time Range Selector */}
        <div className="relative">
          <button
            onClick={() => setShowTimeRangeMenu(!showTimeRangeMenu)}
            className="flex items-center gap-2 px-3 py-2 text-sm text-editor-text bg-editor-bg border border-editor-border rounded-lg hover:bg-editor-hover transition-colors"
          >
            <Calendar size={16} />
            <span>{currentTimeRange.label}</span>
            <ChevronDown size={14} />
          </button>

          {showTimeRangeMenu && (
            <>
              <div
                className="fixed inset-0 z-10"
                onClick={() => setShowTimeRangeMenu(false)}
              />
              <div className="absolute right-0 mt-2 w-48 bg-editor-bg border border-editor-border rounded-lg shadow-lg overflow-hidden z-20">
                {TIME_RANGE_OPTIONS.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => handleTimeRangeSelect(option)}
                    className={`w-full px-4 py-2 text-sm text-left hover:bg-editor-hover transition-colors ${
                      option.value === selectedTimeRange
                        ? 'text-editor-accent bg-editor-accent/10'
                        : 'text-editor-text'
                    }`}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Refresh Button */}
        <button
          onClick={onRefresh}
          disabled={isLoading}
          className="flex items-center gap-2 px-3 py-2 text-sm text-editor-text bg-editor-bg border border-editor-border rounded-lg hover:bg-editor-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
          <span>Refresh</span>
        </button>

        {/* Export Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowExportMenu(!showExportMenu)}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Download size={16} />
            <span>Export</span>
            <ChevronDown size={14} />
          </button>

          {showExportMenu && (
            <>
              <div
                className="fixed inset-0 z-10"
                onClick={() => setShowExportMenu(false)}
              />
              <div className="absolute right-0 mt-2 w-48 bg-editor-bg border border-editor-border rounded-lg shadow-lg overflow-hidden z-20">
                <button
                  onClick={() => handleExport('csv')}
                  className="w-full flex items-center gap-3 px-4 py-2 text-sm text-editor-text hover:bg-editor-hover transition-colors"
                >
                  <FileSpreadsheet size={16} />
                  <span>Export as CSV</span>
                </button>
                <button
                  onClick={() => handleExport('json')}
                  className="w-full flex items-center gap-3 px-4 py-2 text-sm text-editor-text hover:bg-editor-hover transition-colors"
                >
                  <FileJson size={16} />
                  <span>Export as JSON</span>
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
