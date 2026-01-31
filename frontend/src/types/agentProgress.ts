// Agent Progress Types for tracking agent execution state

/**
 * Status of an individual agent's execution
 */
export type AgentStatus = 'pending' | 'running' | 'thinking' | 'completed' | 'failed' | 'cancelled';

/**
 * Status of a swarm's overall execution
 */
export type SwarmStatus = 'running' | 'synthesizing' | 'completed' | 'failed';

/**
 * An event that occurred during agent progress
 */
export interface ProgressEvent {
  type: string;
  timestamp: number;
  data: Record<string, unknown>;
}

/**
 * Historical metrics for a completed agent run
 */
export interface ProgressMetrics {
  agentId: string;
  duration: number;
  tokensGenerated: number;
  stepsCompleted: number;
  completedAt: number;
}

/**
 * Progress state for an individual agent
 */
export interface AgentProgress {
  agentId: string;
  name: string;
  status: AgentStatus;
  currentStep: number;
  totalSteps: number;
  percentComplete: number;
  stepName: string;
  message: string;
  startedAt: number;
  estimatedTimeRemaining: number | null;
  estimatedTokensRemaining: number | null;
  isThinking: boolean;
  thinkingStartedAt: number | null;
  tokensGenerated: number;
  events: ProgressEvent[];
}

/**
 * Progress state for a swarm of agents
 */
export interface SwarmProgress {
  swarmId: string;
  agents: Map<string, AgentProgress>;
  overallPercent: number;
  completedAgents: number;
  totalAgents: number;
  status: SwarmStatus;
}
