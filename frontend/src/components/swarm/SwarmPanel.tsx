import { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Play,
  Pause,
  Square,
  Settings,
  ChevronDown,
  ChevronUp,
  Users,
  MessageSquare,
  Clock,
  Zap,
  Filter,
  X,
} from 'lucide-react';
import { SwarmAgentCard } from './SwarmAgentCard';
import { SwarmProgressChart } from './SwarmProgressChart';
import {
  useSwarmStore,
  type Swarm,
  type SwarmExecution,
  type SwarmStrategy,
  type SwarmAgent,
  type AgentExecutionResult,
} from '../../store/swarmStore';

// ============================================================================
// Types
// ============================================================================

interface SwarmPanelProps {
  swarmId?: string;
  executionId?: string;
  onEdit?: (swarm: Swarm) => void;
  onClose?: () => void;
  className?: string;
}

interface InterAgentMessage {
  id: string;
  fromAgentId: string;
  toAgentId?: string;
  content: string;
  timestamp: string;
  type: 'message' | 'handoff' | 'result' | 'error';
}

// ============================================================================
// Strategy Configuration
// ============================================================================

const strategyLabels: Record<SwarmStrategy, string> = {
  sequential: 'Sequential',
  parallel: 'Parallel',
  hierarchical: 'Hierarchical',
  collaborative: 'Collaborative',
  competitive: 'Competitive',
};

const strategyColors: Record<SwarmStrategy, string> = {
  sequential: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  parallel: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  hierarchical: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  collaborative: 'bg-green-500/20 text-green-400 border-green-500/30',
  competitive: 'bg-red-500/20 text-red-400 border-red-500/30',
};

// ============================================================================
// Helper Functions
// ============================================================================

function formatDuration(startedAt: string, completedAt?: string): string {
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

function calculateTotalTokens(agentResults: AgentExecutionResult[]): number {
  return agentResults.reduce((total, result) => total + (result.tokensUsed || 0), 0);
}

// ============================================================================
// Sub-Components
// ============================================================================

interface HeaderProps {
  swarm: Swarm;
  execution: SwarmExecution | null;
  isExecuting: boolean;
  onRun: () => void;
  onPause: () => void;
  onStop: () => void;
  onEdit: () => void;
  onClose?: () => void;
}

const Header: React.FC<HeaderProps> = ({
  swarm,
  execution,
  isExecuting,
  onRun,
  onPause,
  onStop,
  onEdit,
  onClose,
}) => {
  const isRunning = execution?.status === 'running';
  const isPaused = false; // TODO: Add pause state when backend supports it

  return (
    <div className="flex items-center justify-between p-4 border-b border-editor-border bg-editor-bg">
      <div className="flex items-center gap-3">
        <div className="p-2 bg-editor-accent/10 rounded-lg">
          <Users className="w-5 h-5 text-editor-accent" />
        </div>
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-semibold text-editor-text">{swarm.name}</h2>
            <span
              className={`px-2 py-0.5 text-xs font-medium rounded-full border ${strategyColors[swarm.strategy]}`}
            >
              {strategyLabels[swarm.strategy]}
            </span>
            {isRunning && (
              <span className="flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full bg-editor-accent/20 text-editor-accent">
                <span className="w-1.5 h-1.5 rounded-full bg-editor-accent animate-pulse" />
                Running
              </span>
            )}
          </div>
          {swarm.description && (
            <p className="text-sm text-editor-muted mt-0.5">{swarm.description}</p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Execution controls */}
        {isRunning ? (
          <>
            <button
              onClick={onPause}
              disabled={isPaused}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-editor-warning bg-editor-warning/10 border border-editor-warning/30 rounded-lg hover:bg-editor-warning/20 transition-colors disabled:opacity-50"
            >
              <Pause size={14} />
              Pause
            </button>
            <button
              onClick={onStop}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-editor-error bg-editor-error/10 border border-editor-error/30 rounded-lg hover:bg-editor-error/20 transition-colors"
            >
              <Square size={14} />
              Stop
            </button>
          </>
        ) : (
          <button
            onClick={onRun}
            disabled={isExecuting}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-editor-success bg-editor-success/10 border border-editor-success/30 rounded-lg hover:bg-editor-success/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Play size={14} />
            Run
          </button>
        )}

        {/* Settings */}
        <button
          onClick={onEdit}
          className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          title="Edit swarm configuration"
        >
          <Settings size={18} />
        </button>

        {/* Close button */}
        {onClose && (
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
            title="Close panel"
          >
            <X size={18} />
          </button>
        )}
      </div>
    </div>
  );
};

interface AgentGridProps {
  agents: SwarmAgent[];
  agentResults: AgentExecutionResult[];
  strategy: SwarmStrategy;
  onViewAgentDetails: (agentId: string) => void;
  onStopAgent: (agentId: string) => void;
}

const AgentGrid: React.FC<AgentGridProps> = ({
  agents,
  agentResults,
  strategy,
  onViewAgentDetails,
  onStopAgent,
}) => {
  // Merge agent data with execution results
  const agentsWithStatus = useMemo(() => {
    return agents.map((agent) => {
      const result = agentResults.find((r) => r.agentId === agent.id);
      return {
        ...agent,
        status: result?.status || agent.status,
      };
    });
  }, [agents, agentResults]);

  // Get grid layout class based on strategy
  const getGridClass = () => {
    switch (strategy) {
      case 'sequential':
        // Horizontal flow for pipeline
        return 'flex flex-wrap gap-3';
      case 'competitive':
        // Two columns for debate
        return 'grid grid-cols-2 gap-3';
      case 'parallel':
      case 'hierarchical':
      case 'collaborative':
      default:
        // Responsive grid for others
        return 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3';
    }
  };

  return (
    <div className={getGridClass()}>
      {agentsWithStatus.map((agent) => (
        <SwarmAgentCard
          key={agent.id}
          agent={agent}
          onViewDetails={onViewAgentDetails}
          onStop={onStopAgent}
          compact={strategy === 'sequential' || agents.length > 6}
        />
      ))}
    </div>
  );
};

interface ResultsSectionProps {
  execution: SwarmExecution;
  agents: SwarmAgent[];
}

const ResultsSection: React.FC<ResultsSectionProps> = ({ execution, agents }) => {
  const [expandedAgents, setExpandedAgents] = useState<Set<string>>(new Set());

  const toggleAgent = (agentId: string) => {
    setExpandedAgents((prev) => {
      const next = new Set(prev);
      if (next.has(agentId)) {
        next.delete(agentId);
      } else {
        next.add(agentId);
      }
      return next;
    });
  };

  const totalTokens = calculateTotalTokens(execution.agentResults);
  const duration = formatDuration(execution.startedAt, execution.completedAt);

  return (
    <div className="space-y-4">
      {/* Final Output */}
      {execution.output && (
        <div className="p-4 bg-editor-surface rounded-lg border border-editor-border">
          <h4 className="text-sm font-medium text-editor-text mb-2 flex items-center gap-2">
            <Zap className="w-4 h-4 text-editor-accent" />
            Final Result
          </h4>
          <div className="text-sm text-editor-text whitespace-pre-wrap bg-editor-bg p-3 rounded-md">
            {execution.output}
          </div>
        </div>
      )}

      {/* Execution Summary */}
      <div className="flex items-center gap-4 text-xs text-editor-muted">
        <span className="flex items-center gap-1">
          <Clock className="w-3.5 h-3.5" />
          Duration: {duration}
        </span>
        {totalTokens > 0 && (
          <span className="flex items-center gap-1">
            <Zap className="w-3.5 h-3.5" />
            Tokens: {totalTokens.toLocaleString()}
          </span>
        )}
        <span className="flex items-center gap-1">
          <Users className="w-3.5 h-3.5" />
          Agents: {execution.agentResults.length}
        </span>
      </div>

      {/* Individual Agent Outputs */}
      <div className="space-y-2">
        <h4 className="text-sm font-medium text-editor-muted">Agent Outputs</h4>
        {execution.agentResults.map((result) => {
          const agent = agents.find((a) => a.id === result.agentId);
          const isExpanded = expandedAgents.has(result.agentId);

          return (
            <div
              key={result.agentId}
              className="bg-editor-surface rounded-lg border border-editor-border overflow-hidden"
            >
              <button
                onClick={() => toggleAgent(result.agentId)}
                className="w-full flex items-center justify-between p-3 hover:bg-editor-bg/50 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-editor-text">
                    {agent?.name || result.agentId}
                  </span>
                  <span
                    className={`px-1.5 py-0.5 text-xs rounded ${
                      result.status === 'completed'
                        ? 'bg-editor-success/20 text-editor-success'
                        : result.status === 'failed'
                        ? 'bg-editor-error/20 text-editor-error'
                        : 'bg-editor-muted/20 text-editor-muted'
                    }`}
                  >
                    {result.status}
                  </span>
                  {result.tokensUsed && (
                    <span className="text-xs text-editor-muted">
                      {result.tokensUsed.toLocaleString()} tokens
                    </span>
                  )}
                </div>
                {isExpanded ? (
                  <ChevronUp className="w-4 h-4 text-editor-muted" />
                ) : (
                  <ChevronDown className="w-4 h-4 text-editor-muted" />
                )}
              </button>
              {isExpanded && (
                <div className="px-3 pb-3">
                  {result.output ? (
                    <div className="text-sm text-editor-text whitespace-pre-wrap bg-editor-bg p-3 rounded-md">
                      {result.output}
                    </div>
                  ) : result.error ? (
                    <div className="text-sm text-editor-error bg-editor-error/10 p-3 rounded-md">
                      {result.error}
                    </div>
                  ) : (
                    <div className="text-sm text-editor-muted italic">No output available</div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

interface MessageLogProps {
  messages: InterAgentMessage[];
  agents: SwarmAgent[];
}

const MessageLog: React.FC<MessageLogProps> = ({ messages, agents }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [filterAgent, setFilterAgent] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<string | null>(null);

  const filteredMessages = useMemo(() => {
    return messages.filter((msg) => {
      if (filterAgent && msg.fromAgentId !== filterAgent && msg.toAgentId !== filterAgent) {
        return false;
      }
      if (filterType && msg.type !== filterType) {
        return false;
      }
      return true;
    });
  }, [messages, filterAgent, filterType]);

  const getAgentName = (agentId: string) => {
    return agents.find((a) => a.id === agentId)?.name || agentId;
  };

  const getTypeColor = (type: InterAgentMessage['type']) => {
    switch (type) {
      case 'message':
        return 'text-blue-400';
      case 'handoff':
        return 'text-purple-400';
      case 'result':
        return 'text-green-400';
      case 'error':
        return 'text-red-400';
      default:
        return 'text-editor-muted';
    }
  };

  return (
    <div className="bg-editor-surface rounded-lg border border-editor-border overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between p-3 hover:bg-editor-bg/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <MessageSquare className="w-4 h-4 text-editor-muted" />
          <span className="text-sm font-medium text-editor-text">Message Log</span>
          <span className="text-xs text-editor-muted">({messages.length} messages)</span>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 text-editor-muted" />
        ) : (
          <ChevronDown className="w-4 h-4 text-editor-muted" />
        )}
      </button>

      {isExpanded && (
        <div className="border-t border-editor-border">
          {/* Filters */}
          <div className="flex items-center gap-2 p-2 border-b border-editor-border bg-editor-bg/30">
            <Filter className="w-3.5 h-3.5 text-editor-muted" />
            <select
              value={filterAgent || ''}
              onChange={(e) => setFilterAgent(e.target.value || null)}
              className="text-xs bg-editor-surface border border-editor-border rounded px-2 py-1 text-editor-text"
            >
              <option value="">All agents</option>
              {agents.map((agent) => (
                <option key={agent.id} value={agent.id}>
                  {agent.name}
                </option>
              ))}
            </select>
            <select
              value={filterType || ''}
              onChange={(e) => setFilterType(e.target.value || null)}
              className="text-xs bg-editor-surface border border-editor-border rounded px-2 py-1 text-editor-text"
            >
              <option value="">All types</option>
              <option value="message">Messages</option>
              <option value="handoff">Handoffs</option>
              <option value="result">Results</option>
              <option value="error">Errors</option>
            </select>
          </div>

          {/* Messages */}
          <div className="max-h-64 overflow-y-auto p-2 space-y-1">
            {filteredMessages.length === 0 ? (
              <div className="text-xs text-editor-muted text-center py-4">No messages to display</div>
            ) : (
              filteredMessages.map((msg) => (
                <div key={msg.id} className="flex items-start gap-2 text-xs p-2 bg-editor-bg/50 rounded">
                  <span className={`font-medium ${getTypeColor(msg.type)}`}>[{msg.type}]</span>
                  <span className="text-editor-accent">{getAgentName(msg.fromAgentId)}</span>
                  {msg.toAgentId && (
                    <>
                      <span className="text-editor-muted">→</span>
                      <span className="text-editor-accent">{getAgentName(msg.toAgentId)}</span>
                    </>
                  )}
                  <span className="text-editor-text flex-1 truncate">{msg.content}</span>
                  <span className="text-editor-muted">
                    {new Date(msg.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Main Component
// ============================================================================

export function SwarmPanel({
  swarmId,
  executionId,
  onEdit,
  onClose,
  className = '',
}: SwarmPanelProps) {
  const {
    swarms,
    activeSwarm,
    swarmHistory,
    isExecuting,
    setActiveSwarm,
    startExecution,
    cancelExecution,
  } = useSwarmStore();

  // Inter-agent messages will be populated via WebSocket events
  // TODO: Wire up swarm.message events from WebSocket when backend implements them
  const [interAgentMessages, _setInterAgentMessages] = useState<InterAgentMessage[]>([]);
  const [elapsedTime, setElapsedTime] = useState<string>('0s');

  // Get the swarm to display
  const swarm = useMemo(() => {
    if (swarmId) {
      return swarms.find((s) => s.id === swarmId) || null;
    }
    return activeSwarm;
  }, [swarmId, swarms, activeSwarm]);

  // Get current execution
  const currentExecution = useMemo(() => {
    if (executionId) {
      return swarmHistory.find((e) => e.id === executionId) || null;
    }
    if (swarm) {
      return swarmHistory.find((e) => e.swarmId === swarm.id) || null;
    }
    return null;
  }, [executionId, swarm, swarmHistory]);

  // Update elapsed time for running executions
  useEffect(() => {
    if (!currentExecution || currentExecution.status !== 'running') {
      if (currentExecution?.startedAt) {
        setElapsedTime(formatDuration(currentExecution.startedAt, currentExecution.completedAt));
      }
      return;
    }

    const interval = setInterval(() => {
      setElapsedTime(formatDuration(currentExecution.startedAt));
    }, 1000);

    return () => clearInterval(interval);
  }, [currentExecution]);

  // Set active swarm when swarmId changes
  useEffect(() => {
    if (swarmId && swarmId !== activeSwarm?.id) {
      setActiveSwarm(swarmId);
    }
  }, [swarmId, activeSwarm?.id, setActiveSwarm]);

  // Event handlers
  const handleRun = useCallback(() => {
    if (!swarm) return;
    // TODO: Show input modal for execution input
    startExecution(swarm.id, 'User initiated execution');
  }, [swarm, startExecution]);

  const handlePause = useCallback(() => {
    // TODO: Implement pause when backend supports it
    console.log('Pause not yet implemented');
  }, []);

  const handleStop = useCallback(() => {
    if (currentExecution) {
      cancelExecution(currentExecution.id);
    }
  }, [currentExecution, cancelExecution]);

  const handleEdit = useCallback(() => {
    if (swarm && onEdit) {
      onEdit(swarm);
    }
  }, [swarm, onEdit]);

  const handleViewAgentDetails = useCallback((agentId: string) => {
    // TODO: Implement agent detail modal
    console.log('View agent details:', agentId);
  }, []);

  const handleStopAgent = useCallback((agentId: string) => {
    // TODO: Implement individual agent stop
    console.log('Stop agent:', agentId);
  }, []);

  // Show empty state if no swarm selected
  if (!swarm) {
    return (
      <div className={`flex items-center justify-center h-full bg-editor-bg ${className}`}>
        <div className="text-center p-8">
          <Users className="w-12 h-12 text-editor-muted mx-auto mb-4" />
          <h3 className="text-lg font-medium text-editor-text mb-2">No Swarm Selected</h3>
          <p className="text-editor-muted">Select a swarm from the list to view details and monitor execution.</p>
        </div>
      </div>
    );
  }

  const isRunning = currentExecution?.status === 'running';
  const isComplete = currentExecution?.status === 'completed';
  const isFailed = currentExecution?.status === 'failed';

  return (
    <div className={`flex flex-col h-full bg-editor-bg ${className}`}>
      {/* Header */}
      <Header
        swarm={swarm}
        execution={currentExecution}
        isExecuting={isExecuting}
        onRun={handleRun}
        onPause={handlePause}
        onStop={handleStop}
        onEdit={handleEdit}
        onClose={onClose}
      />

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        {/* Progress Section */}
        {currentExecution && (
          <section>
            <h3 className="text-sm font-medium text-editor-muted mb-3 flex items-center gap-2">
              <Clock className="w-4 h-4" />
              Progress
              {isRunning && <span className="text-xs text-editor-accent">{elapsedTime}</span>}
            </h3>
            <SwarmProgressChart executionId={currentExecution.id} />
          </section>
        )}

        {/* Agent Grid Section */}
        <section>
          <h3 className="text-sm font-medium text-editor-muted mb-3 flex items-center gap-2">
            <Users className="w-4 h-4" />
            Agents ({swarm.agents.length})
          </h3>
          <AgentGrid
            agents={swarm.agents}
            agentResults={currentExecution?.agentResults || []}
            strategy={swarm.strategy}
            onViewAgentDetails={handleViewAgentDetails}
            onStopAgent={handleStopAgent}
          />
        </section>

        {/* Results Section (when complete) */}
        {(isComplete || isFailed) && currentExecution && (
          <section>
            <h3 className="text-sm font-medium text-editor-muted mb-3 flex items-center gap-2">
              <Zap className="w-4 h-4" />
              Results
            </h3>
            <ResultsSection execution={currentExecution} agents={swarm.agents} />
          </section>
        )}

        {/* Message Log */}
        {interAgentMessages.length > 0 && (
          <section>
            <MessageLog messages={interAgentMessages} agents={swarm.agents} />
          </section>
        )}
      </div>
    </div>
  );
}

export default SwarmPanel;
