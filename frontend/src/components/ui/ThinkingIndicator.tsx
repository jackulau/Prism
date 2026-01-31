import React from 'react';
import { Brain, Loader2 } from 'lucide-react';

export interface ThinkingIndicatorProps {
  /** Whether the indicator is active/animating */
  active: boolean;
  /** Visual variant of the indicator */
  variant?: 'dots' | 'pulse' | 'wave' | 'brain' | 'spinner' | 'typewriter';
  /** Size of the indicator */
  size?: 'sm' | 'md' | 'lg';
  /** Color theme */
  color?: 'default' | 'accent' | 'success' | 'muted';
  /** Optional label text */
  label?: string;
  /** Whether to show the label */
  showLabel?: boolean;
  /** Animation duration in ms (affects speed) */
  duration?: number;
  /** Additional CSS classes */
  className?: string;
}

const sizeClasses = {
  sm: {
    dot: 'w-1.5 h-1.5',
    wave: 'h-3',
    waveBar: 'w-0.5',
    pulse: 'w-3 h-3',
    brain: 'w-3 h-3',
    spinner: 'w-3 h-3',
    typewriter: 'w-0.5 h-3',
    gap: 'gap-0.5',
    label: 'text-xs',
  },
  md: {
    dot: 'w-2 h-2',
    wave: 'h-4',
    waveBar: 'w-1',
    pulse: 'w-4 h-4',
    brain: 'w-4 h-4',
    spinner: 'w-4 h-4',
    typewriter: 'w-0.5 h-4',
    gap: 'gap-1',
    label: 'text-sm',
  },
  lg: {
    dot: 'w-3 h-3',
    wave: 'h-6',
    waveBar: 'w-1.5',
    pulse: 'w-6 h-6',
    brain: 'w-6 h-6',
    spinner: 'w-6 h-6',
    typewriter: 'w-1 h-6',
    gap: 'gap-1.5',
    label: 'text-base',
  },
};

const colorClasses = {
  default: 'text-editor-text',
  accent: 'text-editor-accent',
  success: 'text-editor-success',
  muted: 'text-editor-muted',
};

/** Bouncing dots indicator (like typing indicator) */
const DotsVariant: React.FC<{ size: 'sm' | 'md' | 'lg'; duration: number }> = ({
  size,
  duration,
}) => {
  const sizes = sizeClasses[size];
  const baseDelay = duration / 6;

  return (
    <div className={`flex items-center ${sizes.gap}`} role="presentation">
      {[0, 1, 2].map((i) => (
        <div
          key={i}
          className={`${sizes.dot} rounded-full bg-current motion-safe:animate-bounce-dot`}
          style={{
            animationDuration: `${duration}ms`,
            animationDelay: `${i * baseDelay}ms`,
          }}
        />
      ))}
    </div>
  );
};

/** Expanding rings (radar-like pulse) */
const PulseVariant: React.FC<{ size: 'sm' | 'md' | 'lg'; duration: number }> = ({
  size,
  duration,
}) => {
  const sizes = sizeClasses[size];

  return (
    <div className={`relative ${sizes.pulse}`} role="presentation">
      <div
        className="absolute inset-0 rounded-full bg-current motion-safe:animate-ping opacity-75"
        style={{ animationDuration: `${duration}ms` }}
      />
      <div className="absolute inset-1 rounded-full bg-current opacity-100" />
    </div>
  );
};

/** Sound wave / equalizer bars */
const WaveVariant: React.FC<{ size: 'sm' | 'md' | 'lg'; duration: number }> = ({
  size,
  duration,
}) => {
  const sizes = sizeClasses[size];
  const baseDelay = duration / 10;

  return (
    <div
      className={`flex items-end ${sizes.gap} ${sizes.wave}`}
      role="presentation"
    >
      {[0, 1, 2, 3, 4].map((i) => (
        <div
          key={i}
          className={`${sizes.waveBar} bg-current rounded-full motion-safe:animate-wave`}
          style={{
            animationDuration: `${duration}ms`,
            animationDelay: `${i * baseDelay}ms`,
          }}
        />
      ))}
    </div>
  );
};

/** Brain icon with pulse effect */
const BrainVariant: React.FC<{ size: 'sm' | 'md' | 'lg'; duration: number }> = ({
  size,
  duration,
}) => {
  const sizes = sizeClasses[size];
  const iconSizes = { sm: 12, md: 16, lg: 24 };

  return (
    <div className={`relative ${sizes.brain}`} role="presentation">
      <div
        className="absolute inset-0 rounded-full bg-current motion-safe:animate-ping opacity-30"
        style={{ animationDuration: `${duration}ms` }}
      />
      <Brain
        size={iconSizes[size]}
        className="relative motion-safe:animate-pulse"
        style={{ animationDuration: `${duration}ms` }}
      />
    </div>
  );
};

/** Rotating spinner */
const SpinnerVariant: React.FC<{
  size: 'sm' | 'md' | 'lg';
  duration: number;
}> = ({ size, duration }) => {
  const sizes = sizeClasses[size];
  const iconSizes = { sm: 12, md: 16, lg: 24 };

  return (
    <Loader2
      size={iconSizes[size]}
      className={`${sizes.spinner} motion-safe:animate-spin`}
      style={{ animationDuration: `${duration}ms` }}
      role="presentation"
    />
  );
};

/** Blinking cursor (typewriter effect) */
const TypewriterVariant: React.FC<{
  size: 'sm' | 'md' | 'lg';
  duration: number;
}> = ({ size, duration }) => {
  const sizes = sizeClasses[size];

  return (
    <div
      className={`${sizes.typewriter} bg-current motion-safe:animate-typewriter-blink`}
      style={{ animationDuration: `${duration}ms` }}
      role="presentation"
    />
  );
};

/**
 * ThinkingIndicator - Animated indicator for agent processing states
 *
 * Provides multiple visual variants to communicate when an agent is
 * processing, reasoning, or waiting for LLM inference.
 *
 * @example
 * ```tsx
 * <ThinkingIndicator active={isLoading} variant="dots" label="Thinking..." showLabel />
 * <ThinkingIndicator active={true} variant="wave" color="accent" size="lg" />
 * ```
 */
export const ThinkingIndicator: React.FC<ThinkingIndicatorProps> = ({
  active,
  variant = 'dots',
  size = 'md',
  color = 'default',
  label,
  showLabel = false,
  duration = 1400,
  className = '',
}) => {
  if (!active) {
    return null;
  }

  const sizes = sizeClasses[size];
  const colorClass = colorClasses[color];

  const renderVariant = () => {
    switch (variant) {
      case 'dots':
        return <DotsVariant size={size} duration={duration} />;
      case 'pulse':
        return <PulseVariant size={size} duration={duration} />;
      case 'wave':
        return <WaveVariant size={size} duration={duration} />;
      case 'brain':
        return <BrainVariant size={size} duration={duration} />;
      case 'spinner':
        return <SpinnerVariant size={size} duration={duration} />;
      case 'typewriter':
        return <TypewriterVariant size={size} duration={duration} />;
      default:
        return <DotsVariant size={size} duration={duration} />;
    }
  };

  return (
    <div
      className={`inline-flex items-center ${sizes.gap} ${colorClass} ${className}`}
      role="status"
      aria-live="polite"
      aria-busy={active}
      aria-label={label || 'Processing'}
    >
      {renderVariant()}
      {showLabel && label && (
        <span className={`${sizes.label} ${colorClass}`}>{label}</span>
      )}
      {/* Screen reader only text when no visible label */}
      {!showLabel && (
        <span className="sr-only">{label || 'Processing'}</span>
      )}
    </div>
  );
};

export default ThinkingIndicator;
