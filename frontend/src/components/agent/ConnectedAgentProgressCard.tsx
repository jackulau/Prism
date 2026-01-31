import React from 'react';
import { AgentProgressCard, AgentProgressCardProps } from './AgentProgressCard';
import type { AgentStatus } from './StatusBadge';

/**
 * Agent progress state interface for store integration.
 * When an agentProgressStore is implemented, it should provide
 * data conforming to this interface.
 */
export interface AgentProgressState {
  agentId: string;
  name: string;
  status: AgentStatus;
  percentComplete: number;
  currentStep?: number;
  totalSteps?: number;
  stepName?: string;
  message?: string;
  isThinking?: boolean;
  thinkingDuration?: number;
  estimatedTimeRemaining?: number;
  estimatedTokensRemaining?: number;
  tokensGenerated?: number;
  elapsedTime?: number;
}

export interface ConnectedAgentProgressCardProps {
  agentId: string;
  compact?: boolean;
  showMetrics?: boolean;
  onCancel?: () => void;
  onRetry?: () => void;
}

/**
 * Hook placeholder for when an agentProgressStore is implemented.
 *
 * Example usage when store is available:
 * ```ts
 * import { useAgentProgressStore } from '../../store/agentProgressStore';
 *
 * export const useAgentProgress = (agentId: string): AgentProgressState | null => {
 *   return useAgentProgressStore(state => state.getAgentProgress(agentId));
 * };
 * ```
 *
 * For now, this returns null to indicate no progress data is available.
 */
const useAgentProgress = (_agentId: string): AgentProgressState | null => {
  // Placeholder - returns null until agentProgressStore is implemented
  // In production, this would connect to the actual store
  return null;
};

/**
 * ConnectedAgentProgressCard - Store-connected version of AgentProgressCard
 *
 * This component reads agent progress data from the application store
 * and passes it to the AgentProgressCard component.
 *
 * When the agentProgressStore is implemented, update the useAgentProgress
 * hook above to connect to the actual store.
 *
 * @example
 * ```tsx
 * <ConnectedAgentProgressCard
 *   agentId="agent-123"
 *   onCancel={() => cancelAgent('agent-123')}
 *   onRetry={() => retryAgent('agent-123')}
 * />
 * ```
 */
export const ConnectedAgentProgressCard: React.FC<ConnectedAgentProgressCardProps> = ({
  agentId,
  compact,
  showMetrics,
  onCancel,
  onRetry,
}) => {
  const progress = useAgentProgress(agentId);

  if (!progress) {
    return null;
  }

  return (
    <AgentProgressCard
      agentId={progress.agentId}
      name={progress.name}
      status={progress.status}
      percentComplete={progress.percentComplete}
      currentStep={progress.currentStep}
      totalSteps={progress.totalSteps}
      stepName={progress.stepName}
      message={progress.message}
      isThinking={progress.isThinking}
      thinkingDuration={progress.thinkingDuration}
      estimatedTimeRemaining={progress.estimatedTimeRemaining}
      estimatedTokensRemaining={progress.estimatedTokensRemaining}
      tokensGenerated={progress.tokensGenerated}
      elapsedTime={progress.elapsedTime}
      compact={compact}
      showMetrics={showMetrics}
      onCancel={onCancel}
      onRetry={onRetry}
    />
  );
};

/**
 * Demo/Preview component for testing the AgentProgressCard with sample data.
 * Useful for Storybook or development previews.
 */
export const AgentProgressCardPreview: React.FC<{
  variant?: 'running' | 'thinking' | 'completed' | 'failed' | 'pending';
  compact?: boolean;
}> = ({ variant = 'running', compact = false }) => {
  const sampleData: Record<string, Omit<AgentProgressCardProps, 'agentId'>> = {
    pending: {
      name: 'Code Assistant',
      status: 'pending',
      percentComplete: 0,
      message: 'Waiting to start...',
    },
    running: {
      name: 'Code Assistant',
      status: 'running',
      percentComplete: 45,
      currentStep: 2,
      totalSteps: 5,
      stepName: 'Analyzing codebase',
      message: 'Scanning source files for patterns...',
      tokensGenerated: 1523,
      elapsedTime: 12500,
      estimatedTimeRemaining: 15000,
    },
    thinking: {
      name: 'Code Assistant',
      status: 'thinking',
      percentComplete: 60,
      isThinking: true,
      thinkingDuration: 5000,
      currentStep: 3,
      totalSteps: 5,
      stepName: 'Reasoning',
      tokensGenerated: 2841,
      elapsedTime: 25000,
    },
    completed: {
      name: 'Code Assistant',
      status: 'completed',
      percentComplete: 100,
      currentStep: 5,
      totalSteps: 5,
      stepName: 'Complete',
      message: 'Successfully generated implementation plan',
      tokensGenerated: 4892,
      elapsedTime: 45000,
    },
    failed: {
      name: 'Code Assistant',
      status: 'failed',
      percentComplete: 67,
      currentStep: 4,
      totalSteps: 5,
      stepName: 'Error',
      message: 'Rate limit exceeded. Please try again later.',
      tokensGenerated: 3156,
      elapsedTime: 30000,
    },
  };

  const data = sampleData[variant];

  return (
    <AgentProgressCard
      agentId="preview"
      compact={compact}
      onCancel={variant === 'running' || variant === 'thinking' ? () => {} : undefined}
      onRetry={variant === 'failed' ? () => {} : undefined}
      {...data}
    />
  );
};

export default ConnectedAgentProgressCard;
