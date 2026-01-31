import { useState } from 'react';
import {
  Users,
  Play,
  Pencil,
  Copy,
  Trash2,
  Plus,
  Clock,
  ChevronRight,
} from 'lucide-react';
import { ConfirmDialog } from '../ConfirmDialog';
import { toast } from '../../store/toastStore';
import {
  useSwarmStore,
  type Swarm,
  type SwarmStrategy,
  type SwarmExecutionStatus,
  type SwarmAgent,
} from '../../store/swarmStore';

const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
};

const getTimestamp = (): string => {
  return new Date().toISOString();
};

interface SwarmListProps {
  onEdit?: (swarm: Swarm) => void;
  onRun?: (swarm: Swarm) => void;
  onCreate?: () => void;
}

const strategyLabels: Record<SwarmStrategy, string> = {
  sequential: 'Sequential',
  parallel: 'Parallel',
  hierarchical: 'Hierarchical',
  collaborative: 'Collaborative',
  competitive: 'Competitive',
};

const strategyColors: Record<SwarmStrategy, string> = {
  sequential: 'bg-blue-500/10 text-blue-400',
  parallel: 'bg-purple-500/10 text-purple-400',
  hierarchical: 'bg-amber-500/10 text-amber-400',
  collaborative: 'bg-green-500/10 text-green-400',
  competitive: 'bg-red-500/10 text-red-400',
};

export function SwarmList({ onEdit, onRun, onCreate }: SwarmListProps) {
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const {
    swarms,
    swarmHistory,
    deleteSwarm,
    createSwarm,
    isExecuting,
  } = useSwarmStore();

  const getSwarmStatus = (swarmId: string): SwarmExecutionStatus | 'idle' => {
    const recentExecution = swarmHistory.find((exec) => exec.swarmId === swarmId);
    if (!recentExecution) return 'idle';
    return recentExecution.status;
  };

  const getLastRunTime = (swarmId: string): string | null => {
    const recentExecution = swarmHistory.find((exec) => exec.swarmId === swarmId);
    if (!recentExecution) return null;
    return recentExecution.completedAt || recentExecution.startedAt;
  };

  const getStatusColor = (status: SwarmExecutionStatus | 'idle') => {
    switch (status) {
      case 'running':
        return 'bg-editor-warning/10 text-editor-warning';
      case 'completed':
        return 'bg-editor-success/10 text-editor-success';
      case 'failed':
      case 'cancelled':
        return 'bg-editor-error/10 text-editor-error';
      case 'pending':
        return 'bg-blue-500/10 text-blue-400';
      default:
        return 'bg-editor-muted/10 text-editor-muted';
    }
  };

  const handleDuplicate = (swarm: Swarm) => {
    const timestamp = getTimestamp();
    const duplicatedAgents: SwarmAgent[] = swarm.agents.map((agent) => ({
      id: generateId(),
      name: agent.name,
      role: agent.role,
      description: agent.description,
      model: agent.model,
      provider: agent.provider,
      status: 'idle' as const,
      systemPrompt: agent.systemPrompt,
      tools: agent.tools,
      maxTokens: agent.maxTokens,
      temperature: agent.temperature,
      createdAt: timestamp,
      updatedAt: timestamp,
    }));

    const duplicated = createSwarm({
      name: `${swarm.name} (Copy)`,
      description: swarm.description,
      strategy: swarm.strategy,
      agents: duplicatedAgents,
      maxConcurrency: swarm.maxConcurrency,
      timeout: swarm.timeout,
      retryAttempts: swarm.retryAttempts,
      isActive: false,
    });
    toast.success(`Swarm "${duplicated.name}" created`);
  };

  const handleDelete = () => {
    if (deleteId) {
      const swarm = swarms.find((s) => s.id === deleteId);
      deleteSwarm(deleteId);
      toast.success(`Swarm "${swarm?.name}" deleted`);
      setDeleteId(null);
    }
  };

  const handleRun = (swarm: Swarm) => {
    if (onRun) {
      onRun(swarm);
    } else {
      toast.info(`Running swarm "${swarm.name}"...`);
    }
  };

  if (swarms.length === 0) {
    return (
      <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
        <Users className="w-12 h-12 text-editor-muted mx-auto mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">
          No swarms configured
        </h3>
        <p className="text-editor-muted mb-6">
          Create your first swarm to orchestrate multiple AI agents working together.
        </p>
        {onCreate && (
          <button
            onClick={onCreate}
            className="inline-flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            Create your first swarm
          </button>
        )}
      </div>
    );
  }

  return (
    <>
      <div className="space-y-4">
        {swarms.map((swarm) => {
          const status = getSwarmStatus(swarm.id);
          const lastRun = getLastRunTime(swarm.id);
          const isRunning = status === 'running';
          const rolesSummary = [...new Set(swarm.agents.map((a) => a.role))].join(', ');

          return (
            <div
              key={swarm.id}
              className="bg-editor-surface border border-editor-border rounded-lg p-4 hover:border-editor-accent/30 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-4">
                  <div className="p-2 bg-editor-accent/10 rounded-lg">
                    <Users className="w-5 h-5 text-editor-accent" />
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-medium text-editor-text">{swarm.name}</h3>
                      <span
                        className={`px-2 py-0.5 text-xs rounded-full ${strategyColors[swarm.strategy]}`}
                      >
                        {strategyLabels[swarm.strategy]}
                      </span>
                      <span
                        className={`px-2 py-0.5 text-xs rounded-full ${getStatusColor(status)}`}
                      >
                        {status}
                      </span>
                    </div>
                    {swarm.description && (
                      <p className="text-sm text-editor-muted">{swarm.description}</p>
                    )}
                    <div className="flex items-center gap-4 text-xs text-editor-muted">
                      <span className="flex items-center gap-1">
                        <Users size={12} />
                        {swarm.agents.length} agent{swarm.agents.length !== 1 ? 's' : ''}
                      </span>
                      {rolesSummary && (
                        <span className="flex items-center gap-1">
                          <ChevronRight size={12} />
                          {rolesSummary}
                        </span>
                      )}
                      {lastRun && (
                        <span className="flex items-center gap-1">
                          <Clock size={12} />
                          Last run: {new Date(lastRun).toLocaleString()}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-1">
                  <button
                    onClick={() => handleRun(swarm)}
                    disabled={isRunning || isExecuting}
                    className={`p-2 rounded-lg transition-colors ${
                      isRunning || isExecuting
                        ? 'text-editor-muted cursor-not-allowed'
                        : 'text-editor-muted hover:text-editor-success hover:bg-editor-success/10'
                    }`}
                    title={isRunning ? 'Swarm is running' : 'Run swarm'}
                  >
                    <Play size={18} />
                  </button>
                  <button
                    onClick={() => onEdit?.(swarm)}
                    className="p-2 text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 rounded-lg transition-colors"
                    title="Edit swarm"
                  >
                    <Pencil size={18} />
                  </button>
                  <button
                    onClick={() => handleDuplicate(swarm)}
                    className="p-2 text-editor-muted hover:text-editor-accent hover:bg-editor-accent/10 rounded-lg transition-colors"
                    title="Duplicate swarm"
                  >
                    <Copy size={18} />
                  </button>
                  <button
                    onClick={() => setDeleteId(swarm.id)}
                    disabled={isRunning}
                    className={`p-2 rounded-lg transition-colors ${
                      isRunning
                        ? 'text-editor-muted cursor-not-allowed'
                        : 'text-editor-muted hover:text-editor-error hover:bg-editor-error/10'
                    }`}
                    title={isRunning ? 'Cannot delete running swarm' : 'Delete swarm'}
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <ConfirmDialog
        isOpen={deleteId !== null}
        title="Delete Swarm"
        message="Are you sure you want to delete this swarm? This action cannot be undone and will also remove all execution history for this swarm."
        confirmText="Delete Swarm"
        variant="danger"
        onConfirm={handleDelete}
        onCancel={() => setDeleteId(null)}
      />
    </>
  );
}
