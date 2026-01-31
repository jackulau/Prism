import React from 'react';

export interface CircularProgressProps {
  /** Current progress value (0-100 by default, or 0-max if max is specified) */
  value?: number;
  /** Maximum value (default 100) */
  max?: number;
  /** Show spinning animation for unknown progress */
  indeterminate?: boolean;

  /** Size in pixels (default 40) */
  size?: number;
  /** Stroke width in pixels (default 4) */
  strokeWidth?: number;
  /** Color variant */
  variant?: 'default' | 'success' | 'warning' | 'error' | 'accent';
  /** Show percentage text in center */
  showPercentage?: boolean;

  /** Custom content to display in center (overrides showPercentage) */
  children?: React.ReactNode;

  /** Custom aria-label for accessibility */
  ariaLabel?: string;
  /** Additional CSS classes */
  className?: string;
}

const variantColors = {
  default: 'stroke-editor-accent',
  success: 'stroke-editor-success',
  warning: 'stroke-editor-warning',
  error: 'stroke-editor-error',
  accent: 'stroke-editor-accent',
};

const trackColors = {
  default: 'stroke-editor-surface',
  success: 'stroke-editor-success/20',
  warning: 'stroke-editor-warning/20',
  error: 'stroke-editor-error/20',
  accent: 'stroke-editor-accent/20',
};

export const CircularProgress: React.FC<CircularProgressProps> = ({
  value = 0,
  max = 100,
  indeterminate = false,
  size = 40,
  strokeWidth = 4,
  variant = 'default',
  showPercentage = false,
  children,
  ariaLabel,
  className = '',
}) => {
  const percentage = max > 0 ? Math.min(Math.max((value / max) * 100, 0), 100) : 0;

  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (percentage / 100) * circumference;

  const center = size / 2;

  return (
    <div
      className={`relative inline-flex items-center justify-center ${className}`}
      style={{ width: size, height: size }}
      role="progressbar"
      aria-valuenow={indeterminate ? undefined : percentage}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={ariaLabel || `Progress: ${Math.round(percentage)}%`}
    >
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${size} ${size}`}
        className={indeterminate ? 'animate-spin' : ''}
      >
        {/* Background track */}
        <circle
          cx={center}
          cy={center}
          r={radius}
          fill="none"
          strokeWidth={strokeWidth}
          className={trackColors[variant]}
        />
        {/* Progress arc */}
        <circle
          cx={center}
          cy={center}
          r={radius}
          fill="none"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          className={`${variantColors[variant]} transition-all duration-300 ease-out`}
          style={{
            strokeDasharray: circumference,
            strokeDashoffset: indeterminate ? circumference * 0.75 : strokeDashoffset,
            transform: 'rotate(-90deg)',
            transformOrigin: 'center',
          }}
        />
      </svg>
      {/* Center content */}
      {(children || showPercentage) && !indeterminate && (
        <div className="absolute inset-0 flex items-center justify-center">
          {children || (
            <span
              className="text-editor-text font-medium"
              style={{ fontSize: size * 0.25 }}
            >
              {Math.round(percentage)}%
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default CircularProgress;
