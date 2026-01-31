import { useState, useRef, useEffect, useCallback } from 'react';
import { Calendar, ChevronDown, ChevronLeft, ChevronRight } from 'lucide-react';
import type { DateRange, DateRangePreset } from '../../types/tasks';
import { DATE_RANGE_OPTIONS } from '../../types/tasks';

interface DateRangePickerProps {
  /** Current date range value */
  value: DateRange;
  /** Callback when date range changes */
  onChange: (range: DateRange) => void;
  /** Custom class name */
  className?: string;
}

export function DateRangePicker({ value, onChange, className = '' }: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [showCustomPicker, setShowCustomPicker] = useState(false);
  const [tempStartDate, setTempStartDate] = useState<Date | null>(null);
  const [tempEndDate, setTempEndDate] = useState<Date | null>(null);
  const [viewMonth, setViewMonth] = useState(new Date());
  const containerRef = useRef<HTMLDivElement>(null);

  // Get display label
  const getDisplayLabel = () => {
    if (value.preset === 'custom' && value.startDate && value.endDate) {
      return `${formatDate(value.startDate)} - ${formatDate(value.endDate)}`;
    }
    const option = DATE_RANGE_OPTIONS.find((opt) => opt.value === value.preset);
    return option?.label || 'All Time';
  };

  // Handle preset selection
  const handlePresetSelect = useCallback(
    (preset: DateRangePreset) => {
      if (preset === 'custom') {
        setShowCustomPicker(true);
        setTempStartDate(value.startDate);
        setTempEndDate(value.endDate);
        return;
      }

      const range = getDateRangeFromPreset(preset);
      onChange(range);
      setIsOpen(false);
    },
    [onChange, value]
  );

  // Handle custom date selection
  const handleDateClick = (date: Date) => {
    if (!tempStartDate || (tempStartDate && tempEndDate)) {
      setTempStartDate(date);
      setTempEndDate(null);
    } else {
      if (date < tempStartDate) {
        setTempEndDate(tempStartDate);
        setTempStartDate(date);
      } else {
        setTempEndDate(date);
      }
    }
  };

  // Apply custom range
  const handleApplyCustom = () => {
    if (tempStartDate && tempEndDate) {
      onChange({
        preset: 'custom',
        startDate: tempStartDate,
        endDate: tempEndDate,
      });
      setShowCustomPicker(false);
      setIsOpen(false);
    }
  };

  // Cancel custom selection
  const handleCancelCustom = () => {
    setShowCustomPicker(false);
    setTempStartDate(null);
    setTempEndDate(null);
  };

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setShowCustomPicker(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const hasSelection = value.preset !== 'all';

  return (
    <div ref={containerRef} className={`relative ${className}`}>
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
        <Calendar className="w-4 h-4" />
        <span>{getDisplayLabel()}</span>
        <ChevronDown
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 bg-editor-surface border border-editor-border rounded-lg shadow-lg overflow-hidden">
          {!showCustomPicker ? (
            <div className="py-1 w-44">
              {DATE_RANGE_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => handlePresetSelect(option.value)}
                  className={`w-full flex items-center px-3 py-2 text-sm transition-colors ${
                    value.preset === option.value
                      ? 'bg-editor-accent/10 text-editor-accent'
                      : 'text-editor-text hover:bg-editor-accent/5'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          ) : (
            <div className="p-3 w-72">
              {/* Calendar header */}
              <div className="flex items-center justify-between mb-3">
                <button
                  type="button"
                  onClick={() => setViewMonth(addMonths(viewMonth, -1))}
                  className="p-1 text-editor-muted hover:text-editor-text rounded"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <span className="text-sm font-medium text-editor-text">
                  {formatMonthYear(viewMonth)}
                </span>
                <button
                  type="button"
                  onClick={() => setViewMonth(addMonths(viewMonth, 1))}
                  className="p-1 text-editor-muted hover:text-editor-text rounded"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>

              {/* Calendar grid */}
              <CalendarGrid
                viewMonth={viewMonth}
                startDate={tempStartDate}
                endDate={tempEndDate}
                onDateClick={handleDateClick}
              />

              {/* Selected range display */}
              <div className="mt-3 text-xs text-editor-muted text-center">
                {tempStartDate && tempEndDate ? (
                  <span>
                    {formatDate(tempStartDate)} - {formatDate(tempEndDate)}
                  </span>
                ) : tempStartDate ? (
                  <span>Select end date</span>
                ) : (
                  <span>Select start date</span>
                )}
              </div>

              {/* Actions */}
              <div className="flex gap-2 mt-3">
                <button
                  type="button"
                  onClick={handleCancelCustom}
                  className="flex-1 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text border border-editor-border rounded"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleApplyCustom}
                  disabled={!tempStartDate || !tempEndDate}
                  className="flex-1 px-3 py-1.5 text-sm bg-editor-accent text-editor-bg rounded disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Apply
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

interface CalendarGridProps {
  viewMonth: Date;
  startDate: Date | null;
  endDate: Date | null;
  onDateClick: (date: Date) => void;
}

function CalendarGrid({ viewMonth, startDate, endDate, onDateClick }: CalendarGridProps) {
  const days = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
  const weeks = getCalendarWeeks(viewMonth);
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <div>
      {/* Day headers */}
      <div className="grid grid-cols-7 gap-1 mb-1">
        {days.map((day) => (
          <div
            key={day}
            className="text-center text-xs text-editor-muted py-1"
          >
            {day}
          </div>
        ))}
      </div>

      {/* Calendar weeks */}
      {weeks.map((week, weekIndex) => (
        <div key={weekIndex} className="grid grid-cols-7 gap-1">
          {week.map((date, dayIndex) => {
            if (!date) {
              return <div key={dayIndex} className="w-8 h-8" />;
            }

            const isToday = isSameDay(date, today);
            const isSelected =
              (startDate && isSameDay(date, startDate)) ||
              (endDate && isSameDay(date, endDate));
            const isInRange =
              startDate &&
              endDate &&
              date > startDate &&
              date < endDate;
            const isFuture = date > today;

            return (
              <button
                key={dayIndex}
                type="button"
                onClick={() => !isFuture && onDateClick(date)}
                disabled={isFuture}
                className={`w-8 h-8 text-sm rounded transition-colors ${
                  isSelected
                    ? 'bg-editor-accent text-editor-bg'
                    : isInRange
                    ? 'bg-editor-accent/20 text-editor-accent'
                    : isToday
                    ? 'border border-editor-accent text-editor-accent'
                    : isFuture
                    ? 'text-editor-muted/50 cursor-not-allowed'
                    : 'text-editor-text hover:bg-editor-accent/10'
                }`}
              >
                {date.getDate()}
              </button>
            );
          })}
        </div>
      ))}
    </div>
  );
}

// Utility functions
function formatDate(date: Date | null): string {
  if (!date) return '';
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

function formatMonthYear(date: Date): string {
  return date.toLocaleDateString('en-US', {
    month: 'long',
    year: 'numeric',
  });
}

function addMonths(date: Date, months: number): Date {
  const result = new Date(date);
  result.setMonth(result.getMonth() + months);
  return result;
}

function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function getCalendarWeeks(month: Date): (Date | null)[][] {
  const weeks: (Date | null)[][] = [];
  const firstDay = new Date(month.getFullYear(), month.getMonth(), 1);
  const lastDay = new Date(month.getFullYear(), month.getMonth() + 1, 0);

  let week: (Date | null)[] = [];

  // Fill in empty days before the first day of the month
  for (let i = 0; i < firstDay.getDay(); i++) {
    week.push(null);
  }

  // Fill in all days of the month
  for (let day = 1; day <= lastDay.getDate(); day++) {
    week.push(new Date(month.getFullYear(), month.getMonth(), day));
    if (week.length === 7) {
      weeks.push(week);
      week = [];
    }
  }

  // Fill in empty days after the last day of the month
  if (week.length > 0) {
    while (week.length < 7) {
      week.push(null);
    }
    weeks.push(week);
  }

  return weeks;
}

function getDateRangeFromPreset(preset: DateRangePreset): DateRange {
  const now = new Date();
  now.setHours(23, 59, 59, 999);

  switch (preset) {
    case 'today': {
      const start = new Date();
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case '7d': {
      const start = new Date();
      start.setDate(start.getDate() - 7);
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case '30d': {
      const start = new Date();
      start.setDate(start.getDate() - 30);
      start.setHours(0, 0, 0, 0);
      return { preset, startDate: start, endDate: now };
    }
    case 'all':
    default:
      return { preset: 'all', startDate: null, endDate: null };
  }
}
