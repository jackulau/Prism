import React from 'react';

export interface ProgressBarProps {
  /** Current progress value (0-100 by default, or 0-max if max is specified) */
  value: number;
  /** Maximum value (default 100) */
  max?: number;
  /** Show pulsing animation for unknown progress */
  indeterminate?: boolean;

  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Color variant */
  variant?: 'default' | 'success' | 'warning' | 'error' | 'accent';
  /** Show percentage text */
  showPercentage?: boolean;
  /** Enable smooth transition animation on value changes */
  animated?: boolean;
  /** Show striped pattern for active progress */
  striped?: boolean;

  /** Total number of steps for step progress */
  steps?: number;
  /** Current step (1-indexed) */
  currentStep?: number;
  /** Show step indicators */
  showSteps?: boolean;
  /** Labels for each step */
  stepLabels?: string[];

  /** Primary label above the progress bar */
  label?: string;
  /** Secondary label (e.g., "3 of 10 complete") */
  sublabel?: string;

  /** Custom aria-label for accessibility */
  ariaLabel?: string;
  /** Additional CSS classes */
  className?: string;
}

const sizeClasses = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-3',
};

const variantClasses = {
  default: 'bg-editor-accent',
  success: 'bg-editor-success',
  warning: 'bg-editor-warning',
  error: 'bg-editor-error',
  accent: 'bg-editor-accent',
};

const variantBgClasses = {
  default: 'bg-editor-surface',
  success: 'bg-editor-success/20',
  warning: 'bg-editor-warning/20',
  error: 'bg-editor-error/20',
  accent: 'bg-editor-accent/20',
};

const stepVariantClasses = {
  default: {
    active: 'bg-editor-accent border-editor-accent',
    inactive: 'bg-editor-surface border-editor-border',
    line: 'bg-editor-accent',
    lineInactive: 'bg-editor-border',
  },
  success: {
    active: 'bg-editor-success border-editor-success',
    inactive: 'bg-editor-surface border-editor-border',
    line: 'bg-editor-success',
    lineInactive: 'bg-editor-border',
  },
  warning: {
    active: 'bg-editor-warning border-editor-warning',
    inactive: 'bg-editor-surface border-editor-border',
    line: 'bg-editor-warning',
    lineInactive: 'bg-editor-border',
  },
  error: {
    active: 'bg-editor-error border-editor-error',
    inactive: 'bg-editor-surface border-editor-border',
    line: 'bg-editor-error',
    lineInactive: 'bg-editor-border',
  },
  accent: {
    active: 'bg-editor-accent border-editor-accent',
    inactive: 'bg-editor-surface border-editor-border',
    line: 'bg-editor-accent',
    lineInactive: 'bg-editor-border',
  },
};

export const ProgressBar: React.FC<ProgressBarProps> = ({
  value,
  max = 100,
  indeterminate = false,
  size = 'md',
  variant = 'default',
  showPercentage = false,
  animated = true,
  striped = false,
  steps,
  currentStep,
  showSteps = false,
  stepLabels,
  label,
  sublabel,
  ariaLabel,
  className = '',
}) => {
  const percentage = max > 0 ? Math.min(Math.max((value / max) * 100, 0), 100) : 0;
  const isStepProgress = steps !== undefined && steps > 0;

  // Step Progress Variant
  if (isStepProgress && showSteps) {
    const stepCount = steps;
    const current = currentStep ?? Math.ceil((percentage / 100) * stepCount);
    const stepStyles = stepVariantClasses[variant];

    return (
      <div className={`w-full ${className}`}>
        {(label || sublabel) && (
          <div className="flex justify-between items-center mb-2">
            {label && <span className="text-sm text-editor-text font-medium">{label}</span>}
            {sublabel && <span className="text-xs text-editor-muted">{sublabel}</span>}
          </div>
        )}
        <div
          className="flex items-center justify-between w-full"
          role="progressbar"
          aria-valuenow={current}
          aria-valuemin={1}
          aria-valuemax={stepCount}
          aria-label={ariaLabel || label || `Step ${current} of ${stepCount}`}
        >
          {Array.from({ length: stepCount }).map((_, index) => {
            const stepNum = index + 1;
            const isActive = stepNum <= current;
            const isLast = stepNum === stepCount;

            return (
              <React.Fragment key={stepNum}>
                <div className="flex flex-col items-center">
                  <div
                    className={`
                      w-6 h-6 rounded-full border-2 flex items-center justify-center
                      text-xs font-medium transition-all duration-300
                      ${isActive ? stepStyles.active : stepStyles.inactive}
                      ${isActive ? 'text-white' : 'text-editor-muted'}
                    `}
                  >
                    {stepNum}
                  </div>
                  {stepLabels?.[index] && (
                    <span className={`text-xs mt-1 ${isActive ? 'text-editor-text' : 'text-editor-muted'}`}>
                      {stepLabels[index]}
                    </span>
                  )}
                </div>
                {!isLast && (
                  <div
                    className={`
                      flex-1 h-0.5 mx-2 transition-all duration-300
                      ${stepNum < current ? stepStyles.line : stepStyles.lineInactive}
                    `}
                  />
                )}
              </React.Fragment>
            );
          })}
        </div>
      </div>
    );
  }

  // Standard Progress Bar
  return (
    <div className={`w-full ${className}`}>
      {(label || sublabel || showPercentage) && (
        <div className="flex justify-between items-center mb-1">
          <div className="flex flex-col">
            {label && <span className="text-sm text-editor-text font-medium">{label}</span>}
            {sublabel && <span className="text-xs text-editor-muted">{sublabel}</span>}
          </div>
          {showPercentage && !indeterminate && (
            <span className="text-sm text-editor-text font-medium">
              {Math.round(percentage)}%
            </span>
          )}
        </div>
      )}
      <div
        className={`
          w-full rounded-full overflow-hidden
          ${variantBgClasses[variant]}
          ${sizeClasses[size]}
        `}
        role="progressbar"
        aria-valuenow={indeterminate ? undefined : percentage}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={ariaLabel || label || `Progress: ${Math.round(percentage)}%`}
      >
        <div
          className={`
            h-full rounded-full
            ${variantClasses[variant]}
            ${animated && !indeterminate ? 'transition-all duration-300 ease-out' : ''}
            ${indeterminate ? 'animate-progress-shimmer' : ''}
            ${striped && !indeterminate ? 'animate-progress-stripe' : ''}
          `}
          style={{
            width: indeterminate ? '100%' : `${percentage}%`,
            ...(striped && !indeterminate
              ? {
                  backgroundImage:
                    'linear-gradient(45deg, rgba(255,255,255,0.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.15) 75%, transparent 75%, transparent)',
                  backgroundSize: '1rem 1rem',
                }
              : {}),
          }}
        />
      </div>
    </div>
  );
};

export default ProgressBar;
