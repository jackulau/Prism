import { useState, useCallback } from 'react';
import { Plus, Users, ChevronRight } from 'lucide-react';
import { SwarmList, SwarmPanel } from '../components/swarm';
import { SwarmConfigModal, type SwarmConfigFormData } from '../components/swarm/SwarmConfigModal';
import { useSwarmStore, type Swarm } from '../store/swarmStore';
import { toast } from '../store/toastStore';

export default function SwarmOrchestration() {
  const [selectedSwarmId, setSelectedSwarmId] = useState<string | null>(null);
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false);
  const [editingSwarm, setEditingSwarm] = useState<Swarm | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const { swarms, createSwarm, updateSwarm } = useSwarmStore();

  const selectedSwarm = selectedSwarmId
    ? swarms.find((s) => s.id === selectedSwarmId)
    : null;

  const handleOpenCreateModal = useCallback(() => {
    setEditingSwarm(null);
    setIsConfigModalOpen(true);
  }, []);

  const handleEditSwarm = useCallback((swarm: Swarm) => {
    setEditingSwarm(swarm);
    setIsConfigModalOpen(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setIsConfigModalOpen(false);
    setEditingSwarm(null);
  }, []);

  const handleSaveSwarm = useCallback(
    async (config: SwarmConfigFormData) => {
      setIsSaving(true);
      try {
        const timestamp = new Date().toISOString();

        if (editingSwarm) {
          // Update existing swarm
          const updatedAgents = config.agents.map((agentConfig) => ({
            id: crypto.randomUUID(),
            name: `${agentConfig.role.charAt(0).toUpperCase() + agentConfig.role.slice(1)} Agent`,
            role: agentConfig.role,
            model: agentConfig.model || 'claude-3-5-sonnet',
            provider: agentConfig.provider || 'anthropic',
            status: 'idle' as const,
            createdAt: timestamp,
            updatedAt: timestamp,
          }));

          updateSwarm(editingSwarm.id, {
            name: config.name,
            strategy: mapConfigStrategyToStoreStrategy(config.strategy),
            agents: updatedAgents,
            maxConcurrency: config.maxAgents,
            timeout: config.timeout,
            updatedAt: timestamp,
          });
          toast.success(`Swarm "${config.name}" updated`);
        } else {
          // Create new swarm
          const agents = config.agents.flatMap((agentConfig) =>
            Array.from({ length: agentConfig.count }, () => ({
              id: crypto.randomUUID(),
              name: `${agentConfig.role.charAt(0).toUpperCase() + agentConfig.role.slice(1)} Agent`,
              role: agentConfig.role,
              model: agentConfig.model || 'claude-3-5-sonnet',
              provider: agentConfig.provider || 'anthropic',
              status: 'idle' as const,
              createdAt: timestamp,
              updatedAt: timestamp,
            }))
          );

          const newSwarm = createSwarm({
            name: config.name,
            strategy: mapConfigStrategyToStoreStrategy(config.strategy),
            agents,
            maxConcurrency: config.maxAgents,
            timeout: config.timeout,
            isActive: true,
          });
          toast.success(`Swarm "${newSwarm.name}" created`);
          setSelectedSwarmId(newSwarm.id);
        }

        handleCloseModal();
      } catch (error) {
        toast.error('Failed to save swarm configuration');
        console.error('Error saving swarm:', error);
      } finally {
        setIsSaving(false);
      }
    },
    [editingSwarm, createSwarm, updateSwarm, handleCloseModal]
  );

  const handleRunSwarm = useCallback((swarm: Swarm) => {
    setSelectedSwarmId(swarm.id);
    toast.info(`Starting swarm "${swarm.name}"...`);
  }, []);

  // Map config modal strategy to store strategy (they have slightly different sets)
  function mapConfigStrategyToStoreStrategy(
    strategy: SwarmConfigFormData['strategy']
  ): Swarm['strategy'] {
    const strategyMap: Record<SwarmConfigFormData['strategy'], Swarm['strategy']> = {
      parallel: 'parallel',
      pipeline: 'sequential',
      debate: 'competitive',
      consensus: 'collaborative',
      map_reduce: 'parallel',
      specialist: 'hierarchical',
    };
    return strategyMap[strategy];
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Swarm Orchestration</h1>
            <p className="text-editor-muted">
              Configure and run multi-agent swarms
            </p>
          </div>
          <button
            onClick={handleOpenCreateModal}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={18} />
            New Swarm
          </button>
        </div>

        {/* Main Content */}
        {swarms.length === 0 ? (
          // Empty state - no swarms
          <div className="bg-editor-surface border border-editor-border rounded-lg p-12 text-center">
            <Users className="w-16 h-16 text-editor-muted mx-auto mb-6" />
            <h2 className="text-xl font-semibold text-editor-text mb-3">
              Welcome to Swarm Orchestration
            </h2>
            <p className="text-editor-muted max-w-md mx-auto mb-8">
              Create your first swarm to orchestrate multiple AI agents working together.
              Configure different strategies like parallel, sequential, or collaborative execution.
            </p>
            <button
              onClick={handleOpenCreateModal}
              className="inline-flex items-center gap-2 px-6 py-3 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              <Plus size={20} />
              Create Your First Swarm
            </button>
          </div>
        ) : (
          // Two-column layout
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Left column - Swarm List */}
            <div className="lg:col-span-1 space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-medium text-editor-muted uppercase tracking-wide">
                  Your Swarms
                </h2>
                <span className="text-xs text-editor-muted">
                  {swarms.length} swarm{swarms.length !== 1 ? 's' : ''}
                </span>
              </div>
              <SwarmList
                onEdit={handleEditSwarm}
                onRun={handleRunSwarm}
                onCreate={handleOpenCreateModal}
              />
            </div>

            {/* Right column - Swarm Panel */}
            <div className="lg:col-span-2">
              {selectedSwarm ? (
                <SwarmPanel
                  swarmId={selectedSwarm.id}
                  onEdit={handleEditSwarm}
                  onClose={() => setSelectedSwarmId(null)}
                />
              ) : (
                // No swarm selected
                <div className="bg-editor-surface border border-editor-border rounded-lg p-12 text-center h-full flex flex-col items-center justify-center min-h-[400px]">
                  <ChevronRight className="w-12 h-12 text-editor-muted mb-4" />
                  <h3 className="text-lg font-medium text-editor-text mb-2">
                    Select a Swarm
                  </h3>
                  <p className="text-editor-muted max-w-sm">
                    Choose a swarm from the list to view its details, monitor execution progress,
                    and manage its agents.
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Config Modal */}
        <SwarmConfigModal
          isOpen={isConfigModalOpen}
          onClose={handleCloseModal}
          onSave={handleSaveSwarm}
          initialConfig={
            editingSwarm
              ? {
                  name: editingSwarm.name,
                  strategy: mapStoreStrategyToConfigStrategy(editingSwarm.strategy),
                  maxAgents: editingSwarm.maxConcurrency,
                  timeout: editingSwarm.timeout || 300,
                }
              : undefined
          }
          isSaving={isSaving}
        />
      </div>
    </div>
  );
}

// Map store strategy back to config modal strategy
function mapStoreStrategyToConfigStrategy(
  strategy: Swarm['strategy']
): SwarmConfigFormData['strategy'] {
  const strategyMap: Record<Swarm['strategy'], SwarmConfigFormData['strategy']> = {
    parallel: 'parallel',
    sequential: 'pipeline',
    competitive: 'debate',
    collaborative: 'consensus',
    hierarchical: 'specialist',
  };
  return strategyMap[strategy];
}
