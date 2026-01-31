import React, { useEffect, useState, useMemo } from 'react';
import {
  Play,
  CheckCircle2,
  XCircle,
  Clock,
  Loader2,
  GitBranch,
  ArrowRight,
  MessageSquare,
  Users,
  Zap,
  Timer,
} from 'lucide-react';
import {
  useSwarmStore,
  type SwarmExecution,
  type SwarmAgentStatus,
  type SwarmStrategy,
  type SwarmAgent,
  type AgentExecutionResult,
} from '../../store/swarmStore';

// ============================================================================
// Types
// ============================================================================

type ExecutionPhase = 'initializing' | 'running' | 'synthesizing' | 'complete' | 'failed';

interface SwarmProgressChartProps {
  executionId?: string;
  className?: string;
}

interface AgentCounts {
  running: number;
  completed: number;
  failed: number;
  pending: number;
  total: number;
}

// ============================================================================
// Helper Functions
// ============================================================================

function getPhaseFromExecution(execution: SwarmExecution | null): ExecutionPhase {
  if (!execution) return 'initializing';

  switch (execution.status) {
    case 'pending':
      return 'initializing';
    case 'running':
      // Check if we're synthesizing (high progress, most agents done)
      if (execution.progress >= 90) return 'synthesizing';
      return 'running';
    case 'completed':
      return 'complete';
    case 'failed':
    case 'cancelled':
      return 'failed';
    default:
      return 'running';
  }
}

function getAgentCounts(agentResults: AgentExecutionResult[]): AgentCounts {
  const counts: AgentCounts = {
    running: 0,
    completed: 0,
    failed: 0,
    pending: 0,
    total: agentResults.length,
  };

  for (const result of agentResults) {
    switch (result.status) {
      case 'running':
        counts.running++;
        break;
      case 'completed':
        counts.completed++;
        break;
      case 'failed':
        counts.failed++;
        break;
      case 'idle':
      case 'waiting':
      case 'paused':
      default:
        counts.pending++;
        break;
    }
  }

  return counts;
}

function formatElapsedTime(startedAt: string, completedAt?: string): string {
  const start = new Date(startedAt).getTime();
  const end = completedAt ? new Date(completedAt).getTime() : Date.now();
  const elapsed = Math.max(0, end - start);

  const seconds = Math.floor(elapsed / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

// ============================================================================
// Sub-Components
// ============================================================================

interface PhaseIndicatorProps {
  currentPhase: ExecutionPhase;
}

const PhaseIndicator: React.FC<PhaseIndicatorProps> = ({ currentPhase }) => {
  const phases: { id: ExecutionPhase; label: string; icon: React.ReactNode }[] = [
    { id: 'initializing', label: 'Initializing', icon: <Play className="w-3 h-3" /> },
    { id: 'running', label: 'Running', icon: <Loader2 className="w-3 h-3" /> },
    { id: 'synthesizing', label: 'Synthesizing', icon: <Zap className="w-3 h-3" /> },
    { id: 'complete', label: 'Complete', icon: <CheckCircle2 className="w-3 h-3" /> },
  ];

  const getPhaseIndex = (phase: ExecutionPhase): number => {
    if (phase === 'failed') return -1;
    return phases.findIndex((p) => p.id === phase);
  };

  const currentIndex = getPhaseIndex(currentPhase);
  const isFailed = currentPhase === 'failed';

  return (
    <div className="flex items-center gap-1">
      {phases.map((phase, index) => {
        const isActive = index === currentIndex;
        const isCompleted = index < currentIndex && !isFailed;
        const isFuturePhase = index > currentIndex || isFailed;

        return (
          <React.Fragment key={phase.id}>
            <div
              className={`flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium transition-all ${
                isActive
                  ? 'bg-editor-accent text-editor-bg'
                  : isCompleted
                  ? 'bg-editor-success/20 text-editor-success'
                  : isFuturePhase
                  ? 'bg-editor-surface text-editor-muted'
                  : 'bg-editor-surface text-editor-muted'
              }`}
            >
              <span className={isActive ? 'animate-spin' : ''}>
                {isCompleted ? <CheckCircle2 className="w-3 h-3" /> : phase.icon}
              </span>
              <span className="hidden sm:inline">{phase.label}</span>
            </div>
            {index < phases.length - 1 && (
              <ArrowRight
                className={`w-3 h-3 ${
                  index < currentIndex ? 'text-editor-success' : 'text-editor-muted/50'
                }`}
              />
            )}
          </React.Fragment>
        );
      })}
      {isFailed && (
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium bg-editor-error/20 text-editor-error">
          <XCircle className="w-3 h-3" />
          <span className="hidden sm:inline">Failed</span>
        </div>
      )}
    </div>
  );
};

interface ProgressBarProps {
  progress: number;
  counts: AgentCounts;
}

const ProgressBar: React.FC<ProgressBarProps> = ({ progress, counts }) => {
  const completedPercent = counts.total > 0 ? (counts.completed / counts.total) * 100 : 0;
  const runningPercent = counts.total > 0 ? (counts.running / counts.total) * 100 : 0;
  const failedPercent = counts.total > 0 ? (counts.failed / counts.total) * 100 : 0;

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-xs">
        <span className="text-editor-muted">Progress</span>
        <span className="text-editor-text font-medium">{Math.round(progress)}%</span>
      </div>
      <div className="h-3 bg-editor-surface rounded-full overflow-hidden flex">
        <div
          className="h-full bg-editor-success transition-all duration-500 ease-out"
          style={{ width: `${completedPercent}%` }}
        />
        <div
          className="h-full bg-editor-accent animate-pulse transition-all duration-500 ease-out"
          style={{ width: `${runningPercent}%` }}
        />
        <div
          className="h-full bg-editor-error transition-all duration-500 ease-out"
          style={{ width: `${failedPercent}%` }}
        />
      </div>
    </div>
  );
};

interface AgentStatusBreakdownProps {
  counts: AgentCounts;
}

const AgentStatusBreakdown: React.FC<AgentStatusBreakdownProps> = ({ counts }) => {
  const statuses = [
    { label: 'Running', count: counts.running, color: 'text-editor-accent', bgColor: 'bg-editor-accent/20' },
    { label: 'Completed', count: counts.completed, color: 'text-editor-success', bgColor: 'bg-editor-success/20' },
    { label: 'Failed', count: counts.failed, color: 'text-editor-error', bgColor: 'bg-editor-error/20' },
    { label: 'Pending', count: counts.pending, color: 'text-editor-muted', bgColor: 'bg-editor-muted/20' },
  ];

  return (
    <div className="flex items-center gap-3 flex-wrap">
      <div className="flex items-center gap-1.5 text-xs text-editor-muted">
        <Users className="w-3.5 h-3.5" />
        <span>Agents:</span>
      </div>
      {statuses.map(
        (status) =>
          status.count > 0 && (
            <div
              key={status.label}
              className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs ${status.bgColor} ${status.color}`}
            >
              <span className="font-semibold">{status.count}</span>
              <span>{status.label}</span>
            </div>
          )
      )}
    </div>
  );
};

interface ElapsedTimeDisplayProps {
  startedAt: string;
  completedAt?: string;
  isRunning: boolean;
}

const ElapsedTimeDisplay: React.FC<ElapsedTimeDisplayProps> = ({
  startedAt,
  completedAt,
  isRunning,
}) => {
  const [elapsed, setElapsed] = useState(() => formatElapsedTime(startedAt, completedAt));

  useEffect(() => {
    if (!isRunning || completedAt) {
      setElapsed(formatElapsedTime(startedAt, completedAt));
      return;
    }

    const interval = setInterval(() => {
      setElapsed(formatElapsedTime(startedAt));
    }, 1000);

    return () => clearInterval(interval);
  }, [startedAt, completedAt, isRunning]);

  return (
    <div className="flex items-center gap-1.5 text-xs">
      <Timer className={`w-3.5 h-3.5 ${isRunning ? 'text-editor-accent' : 'text-editor-muted'}`} />
      <span className="text-editor-muted">Elapsed:</span>
      <span className={`font-mono ${isRunning ? 'text-editor-text' : 'text-editor-muted'}`}>
        {elapsed}
      </span>
    </div>
  );
};

// ============================================================================
// Strategy-Specific Visualizations
// ============================================================================

interface StrategyVisualizationProps {
  agents: SwarmAgent[];
  agentResults: AgentExecutionResult[];
  strategy: SwarmStrategy;
}

const getStatusColor = (status: SwarmAgentStatus): string => {
  switch (status) {
    case 'running':
      return 'bg-editor-accent';
    case 'completed':
      return 'bg-editor-success';
    case 'failed':
      return 'bg-editor-error';
    case 'paused':
      return 'bg-editor-warning';
    default:
      return 'bg-editor-muted/50';
  }
};

const getStatusBorderColor = (status: SwarmAgentStatus): string => {
  switch (status) {
    case 'running':
      return 'border-editor-accent';
    case 'completed':
      return 'border-editor-success';
    case 'failed':
      return 'border-editor-error';
    case 'paused':
      return 'border-editor-warning';
    default:
      return 'border-editor-muted/50';
  }
};

const ParallelVisualization: React.FC<{ agentResults: AgentExecutionResult[]; agents: SwarmAgent[] }> = ({
  agentResults,
  agents,
}) => {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-1.5 text-xs text-editor-muted">
        <GitBranch className="w-3.5 h-3.5" />
        <span>Parallel Execution</span>
      </div>
      <div className="flex gap-1 flex-wrap">
        {agentResults.map((result, index) => {
          const agent = agents.find((a) => a.id === result.agentId);
          return (
            <div
              key={result.agentId}
              className={`w-8 h-8 rounded-md flex items-center justify-center text-xs font-medium transition-all duration-300 ${getStatusColor(result.status)} ${
                result.status === 'running' ? 'animate-pulse' : ''
              }`}
              title={`${agent?.name || `Agent ${index + 1}`}: ${result.status}`}
            >
              {result.status === 'running' ? (
                <Loader2 className="w-4 h-4 text-editor-bg animate-spin" />
              ) : result.status === 'completed' ? (
                <CheckCircle2 className="w-4 h-4 text-editor-bg" />
              ) : result.status === 'failed' ? (
                <XCircle className="w-4 h-4 text-editor-bg" />
              ) : (
                <span className="text-editor-bg">{index + 1}</span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

const PipelineVisualization: React.FC<{ agentResults: AgentExecutionResult[]; agents: SwarmAgent[] }> = ({
  agentResults,
  agents,
}) => {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-1.5 text-xs text-editor-muted">
        <ArrowRight className="w-3.5 h-3.5" />
        <span>Pipeline Execution</span>
      </div>
      <div className="flex items-center gap-0.5 overflow-x-auto pb-2">
        {agentResults.map((result, index) => {
          const agent = agents.find((a) => a.id === result.agentId);
          return (
            <React.Fragment key={result.agentId}>
              <div
                className={`flex-shrink-0 px-3 py-2 rounded-md border-2 transition-all duration-300 ${getStatusBorderColor(result.status)} ${
                  result.status === 'running' ? 'bg-editor-accent/10' : 'bg-editor-surface'
                }`}
                title={`${agent?.name || `Agent ${index + 1}`}: ${result.status}`}
              >
                <div className="flex items-center gap-2">
                  {result.status === 'running' ? (
                    <Loader2 className="w-3.5 h-3.5 text-editor-accent animate-spin" />
                  ) : result.status === 'completed' ? (
                    <CheckCircle2 className="w-3.5 h-3.5 text-editor-success" />
                  ) : result.status === 'failed' ? (
                    <XCircle className="w-3.5 h-3.5 text-editor-error" />
                  ) : (
                    <Clock className="w-3.5 h-3.5 text-editor-muted" />
                  )}
                  <span className="text-xs font-medium text-editor-text truncate max-w-[60px]">
                    {agent?.name || `Step ${index + 1}`}
                  </span>
                </div>
              </div>
              {index < agentResults.length - 1 && (
                <ArrowRight
                  className={`w-4 h-4 flex-shrink-0 ${
                    result.status === 'completed' ? 'text-editor-success' : 'text-editor-muted/50'
                  }`}
                />
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
};

const DebateVisualization: React.FC<{ agentResults: AgentExecutionResult[]; agents: SwarmAgent[] }> = ({
  agentResults,
  agents,
}) => {
  // Split agents into two groups for debate visualization
  const halfLength = Math.ceil(agentResults.length / 2);
  const leftAgents = agentResults.slice(0, halfLength);
  const rightAgents = agentResults.slice(halfLength);

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-1.5 text-xs text-editor-muted">
        <MessageSquare className="w-3.5 h-3.5" />
        <span>Debate Mode</span>
      </div>
      <div className="flex gap-4">
        <div className="flex-1 space-y-1">
          {leftAgents.map((result, index) => {
            const agent = agents.find((a) => a.id === result.agentId);
            return (
              <div
                key={result.agentId}
                className={`flex items-center gap-2 px-2 py-1.5 rounded-md border transition-all ${getStatusBorderColor(result.status)} ${
                  result.status === 'running' ? 'bg-editor-accent/10' : 'bg-editor-surface'
                }`}
              >
                {result.status === 'running' ? (
                  <Loader2 className="w-3 h-3 text-editor-accent animate-spin" />
                ) : result.status === 'completed' ? (
                  <CheckCircle2 className="w-3 h-3 text-editor-success" />
                ) : result.status === 'failed' ? (
                  <XCircle className="w-3 h-3 text-editor-error" />
                ) : (
                  <Clock className="w-3 h-3 text-editor-muted" />
                )}
                <span className="text-xs text-editor-text truncate">
                  {agent?.name || `Agent ${index + 1}`}
                </span>
              </div>
            );
          })}
        </div>
        <div className="flex items-center">
          <MessageSquare className="w-4 h-4 text-editor-muted rotate-180" />
          <MessageSquare className="w-4 h-4 text-editor-muted" />
        </div>
        <div className="flex-1 space-y-1">
          {rightAgents.map((result, index) => {
            const agent = agents.find((a) => a.id === result.agentId);
            return (
              <div
                key={result.agentId}
                className={`flex items-center gap-2 px-2 py-1.5 rounded-md border transition-all ${getStatusBorderColor(result.status)} ${
                  result.status === 'running' ? 'bg-editor-accent/10' : 'bg-editor-surface'
                }`}
              >
                {result.status === 'running' ? (
                  <Loader2 className="w-3 h-3 text-editor-accent animate-spin" />
                ) : result.status === 'completed' ? (
                  <CheckCircle2 className="w-3 h-3 text-editor-success" />
                ) : result.status === 'failed' ? (
                  <XCircle className="w-3 h-3 text-editor-error" />
                ) : (
                  <Clock className="w-3 h-3 text-editor-muted" />
                )}
                <span className="text-xs text-editor-text truncate">
                  {agent?.name || `Agent ${halfLength + index + 1}`}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

const StrategyVisualization: React.FC<StrategyVisualizationProps> = ({
  agents,
  agentResults,
  strategy,
}) => {
  switch (strategy) {
    case 'sequential':
      return <PipelineVisualization agentResults={agentResults} agents={agents} />;
    case 'parallel':
      return <ParallelVisualization agentResults={agentResults} agents={agents} />;
    case 'competitive':
      return <DebateVisualization agentResults={agentResults} agents={agents} />;
    case 'hierarchical':
    case 'collaborative':
    default:
      return <ParallelVisualization agentResults={agentResults} agents={agents} />;
  }
};

// ============================================================================
// Main Component
// ============================================================================

export const SwarmProgressChart: React.FC<SwarmProgressChartProps> = ({
  executionId,
  className = '',
}) => {
  const { swarmHistory, swarms } = useSwarmStore();

  // Get the execution to display
  const execution = useMemo(() => {
    if (executionId) {
      return swarmHistory.find((e) => e.id === executionId) || null;
    }
    // Default to most recent execution
    return swarmHistory[0] || null;
  }, [executionId, swarmHistory]);

  // Get the swarm configuration
  const swarm = useMemo(() => {
    if (!execution) return null;
    return swarms.find((s) => s.id === execution.swarmId) || null;
  }, [execution, swarms]);

  if (!execution) {
    return (
      <div className={`p-4 bg-editor-bg rounded-lg border border-editor-border ${className}`}>
        <div className="flex items-center justify-center text-editor-muted text-sm">
          No swarm execution to display
        </div>
      </div>
    );
  }

  const phase = getPhaseFromExecution(execution);
  const counts = getAgentCounts(execution.agentResults);
  const isRunning = execution.status === 'running';

  return (
    <div
      className={`p-4 bg-editor-bg rounded-lg border border-editor-border space-y-4 ${className}`}
    >
      {/* Header with swarm name and phase */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          <Users className="w-4 h-4 text-editor-accent" />
          <span className="text-sm font-medium text-editor-text">
            {swarm?.name || 'Swarm Execution'}
          </span>
          {isRunning && (
            <span className="flex items-center gap-1 text-xs text-editor-accent">
              <span className="w-1.5 h-1.5 rounded-full bg-editor-accent animate-pulse" />
              Active
            </span>
          )}
        </div>
        <PhaseIndicator currentPhase={phase} />
      </div>

      {/* Progress bar */}
      <ProgressBar progress={execution.progress} counts={counts} />

      {/* Agent status and elapsed time */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <AgentStatusBreakdown counts={counts} />
        <ElapsedTimeDisplay
          startedAt={execution.startedAt}
          completedAt={execution.completedAt}
          isRunning={isRunning}
        />
      </div>

      {/* Strategy-specific visualization */}
      {swarm && execution.agentResults.length > 0 && (
        <div className="pt-2 border-t border-editor-border">
          <StrategyVisualization
            agents={swarm.agents}
            agentResults={execution.agentResults}
            strategy={swarm.strategy}
          />
        </div>
      )}

      {/* Error display */}
      {execution.error && (
        <div className="p-3 rounded-md bg-editor-error/10 border border-editor-error/30">
          <div className="flex items-start gap-2">
            <XCircle className="w-4 h-4 text-editor-error mt-0.5 flex-shrink-0" />
            <div className="text-xs text-editor-error">{execution.error}</div>
          </div>
        </div>
      )}

      {/* Final output preview */}
      {execution.output && phase === 'complete' && (
        <div className="p-3 rounded-md bg-editor-success/10 border border-editor-success/30">
          <div className="flex items-start gap-2">
            <CheckCircle2 className="w-4 h-4 text-editor-success mt-0.5 flex-shrink-0" />
            <div className="text-xs text-editor-text line-clamp-3">{execution.output}</div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SwarmProgressChart;
