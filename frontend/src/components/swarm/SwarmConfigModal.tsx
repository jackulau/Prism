import { useState, useCallback } from 'react';
import {
  X,
  Plus,
  Trash2,
  Loader2,
  CheckCircle,
  Save,
  AlertCircle,
  Users,
  Zap,
  GitBranch,
  MessageSquare,
  Layers,
  Target,
} from 'lucide-react';
import { toast } from '../../store/toastStore';

// ============================================================================
// Types (aligned with swarm-api-types)
// ============================================================================

export type SwarmStrategy =
  | 'parallel'
  | 'pipeline'
  | 'debate'
  | 'consensus'
  | 'map_reduce'
  | 'specialist';

export type AgentRole =
  | 'general'
  | 'planner'
  | 'coder'
  | 'reviewer'
  | 'researcher'
  | 'writer'
  | 'analyst'
  | 'debugger'
  | 'tester'
  | 'synthesizer';

export interface AgentRoleConfig {
  id: string;
  role: AgentRole;
  count: number;
  provider?: string;
  model?: string;
}

export interface SwarmConfigFormData {
  name: string;
  strategy: SwarmStrategy;
  agents: AgentRoleConfig[];
  maxAgents: number;
  timeout: number;
  useSynthesizer: boolean;
}

// ============================================================================
// Constants
// ============================================================================

const STRATEGY_INFO: Record<SwarmStrategy, { label: string; description: string; icon: React.ReactNode; minAgents: number }> = {
  parallel: {
    label: 'Parallel',
    description: 'All agents work independently on the task simultaneously',
    icon: <Zap size={16} />,
    minAgents: 1,
  },
  pipeline: {
    label: 'Pipeline',
    description: 'Agents execute in sequence, passing output to the next',
    icon: <GitBranch size={16} />,
    minAgents: 2,
  },
  debate: {
    label: 'Debate',
    description: 'Agents critique and refine each other\'s work',
    icon: <MessageSquare size={16} />,
    minAgents: 2,
  },
  consensus: {
    label: 'Consensus',
    description: 'Agents collaborate to reach agreement on a solution',
    icon: <Users size={16} />,
    minAgents: 2,
  },
  map_reduce: {
    label: 'Map-Reduce',
    description: 'Task split, parallel processing, then synthesis',
    icon: <Layers size={16} />,
    minAgents: 2,
  },
  specialist: {
    label: 'Specialist',
    description: 'Route tasks to specialized agents based on expertise',
    icon: <Target size={16} />,
    minAgents: 1,
  },
};

const ROLE_INFO: Record<AgentRole, { label: string; description: string }> = {
  general: { label: 'General', description: 'General-purpose assistant' },
  planner: { label: 'Planner', description: 'Task planning and breakdown' },
  coder: { label: 'Coder', description: 'Software development' },
  reviewer: { label: 'Reviewer', description: 'Code review and quality checks' },
  researcher: { label: 'Researcher', description: 'Research and information gathering' },
  writer: { label: 'Writer', description: 'Technical writing and documentation' },
  analyst: { label: 'Analyst', description: 'Data analysis and insights' },
  debugger: { label: 'Debugger', description: 'Debugging and issue resolution' },
  tester: { label: 'Tester', description: 'Testing and quality assurance' },
  synthesizer: { label: 'Synthesizer', description: 'Combining multiple outputs' },
};

const ALL_STRATEGIES = Object.keys(STRATEGY_INFO) as SwarmStrategy[];
const ALL_ROLES = Object.keys(ROLE_INFO) as AgentRole[];

const DEFAULT_CONFIG: SwarmConfigFormData = {
  name: '',
  strategy: 'parallel',
  agents: [{ id: crypto.randomUUID(), role: 'coder', count: 1 }],
  maxAgents: 5,
  timeout: 300,
  useSynthesizer: true,
};

// ============================================================================
// Component Props
// ============================================================================

interface SwarmConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (config: SwarmConfigFormData) => void;
  initialConfig?: Partial<SwarmConfigFormData>;
  isSaving?: boolean;
}

// ============================================================================
// Main Component
// ============================================================================

export function SwarmConfigModal({
  isOpen,
  onClose,
  onSave,
  initialConfig,
  isSaving = false,
}: SwarmConfigModalProps) {
  const [config, setConfig] = useState<SwarmConfigFormData>(() => ({
    ...DEFAULT_CONFIG,
    ...initialConfig,
    agents: initialConfig?.agents?.length ? initialConfig.agents : DEFAULT_CONFIG.agents,
  }));
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Validation
  const validate = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};

    if (!config.name.trim()) {
      newErrors.name = 'Swarm name is required';
    }

    if (config.agents.length === 0) {
      newErrors.agents = 'At least one agent is required';
    }

    const totalAgents = config.agents.reduce((sum, a) => sum + a.count, 0);
    if (totalAgents > config.maxAgents) {
      newErrors.agents = `Total agents (${totalAgents}) exceeds maximum (${config.maxAgents})`;
    }

    const strategyInfo = STRATEGY_INFO[config.strategy];
    if (totalAgents < strategyInfo.minAgents) {
      newErrors.strategy = `${strategyInfo.label} strategy requires at least ${strategyInfo.minAgents} agent(s)`;
    }

    if (config.timeout < 10 || config.timeout > 3600) {
      newErrors.timeout = 'Timeout must be between 10 and 3600 seconds';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [config]);

  // Handlers
  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (validate()) {
        onSave(config);
      } else {
        toast.error('Please fix the errors before saving');
      }
    },
    [config, validate, onSave]
  );

  const handleAddAgent = useCallback(() => {
    const usedRoles = new Set(config.agents.map((a) => a.role));
    const availableRole = ALL_ROLES.find((r) => !usedRoles.has(r)) || 'general';

    setConfig((prev) => ({
      ...prev,
      agents: [
        ...prev.agents,
        { id: crypto.randomUUID(), role: availableRole, count: 1 },
      ],
    }));
  }, [config.agents]);

  const handleRemoveAgent = useCallback((id: string) => {
    setConfig((prev) => ({
      ...prev,
      agents: prev.agents.filter((a) => a.id !== id),
    }));
  }, []);

  const handleAgentChange = useCallback(
    (id: string, field: keyof AgentRoleConfig, value: string | number) => {
      setConfig((prev) => ({
        ...prev,
        agents: prev.agents.map((a) =>
          a.id === id ? { ...a, [field]: value } : a
        ),
      }));
    },
    []
  );

  const handleStrategyChange = useCallback((strategy: SwarmStrategy) => {
    setConfig((prev) => ({ ...prev, strategy }));
    setErrors((prev) => ({ ...prev, strategy: '' }));
  }, []);

  if (!isOpen) return null;

  const totalAgents = config.agents.reduce((sum, a) => sum + a.count, 0);

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40 transition-opacity duration-200"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[640px] max-w-[95vw] max-h-[90vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border shrink-0">
          <div className="flex items-center gap-3">
            <Users size={24} className="text-editor-accent" />
            <h2 className="text-lg font-semibold text-editor-text">
              Configure Swarm
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto">
          <div className="p-6 space-y-6">
            {/* Swarm Name */}
            <div className="space-y-2">
              <label className="block text-sm font-medium text-editor-text">
                Swarm Name
              </label>
              <input
                type="text"
                value={config.name}
                onChange={(e) =>
                  setConfig((prev) => ({ ...prev, name: e.target.value }))
                }
                placeholder="My Swarm"
                className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent transition-colors ${
                  errors.name ? 'border-editor-error' : 'border-editor-border'
                }`}
              />
              {errors.name && (
                <p className="flex items-center gap-1 text-sm text-editor-error">
                  <AlertCircle size={14} />
                  {errors.name}
                </p>
              )}
            </div>

            {/* Strategy Selection */}
            <div className="space-y-3">
              <label className="block text-sm font-medium text-editor-text">
                Orchestration Strategy
              </label>
              <div className="grid grid-cols-2 gap-2">
                {ALL_STRATEGIES.map((strategy) => {
                  const info = STRATEGY_INFO[strategy];
                  const isSelected = config.strategy === strategy;
                  return (
                    <button
                      key={strategy}
                      type="button"
                      onClick={() => handleStrategyChange(strategy)}
                      className={`flex items-start gap-3 p-3 rounded-lg border transition-all text-left ${
                        isSelected
                          ? 'border-editor-accent bg-editor-accent/10'
                          : 'border-editor-border bg-editor-surface hover:border-editor-muted'
                      }`}
                    >
                      <span
                        className={`mt-0.5 ${
                          isSelected ? 'text-editor-accent' : 'text-editor-muted'
                        }`}
                      >
                        {info.icon}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p
                          className={`text-sm font-medium ${
                            isSelected ? 'text-editor-accent' : 'text-editor-text'
                          }`}
                        >
                          {info.label}
                        </p>
                        <p className="text-xs text-editor-muted line-clamp-2">
                          {info.description}
                        </p>
                      </div>
                    </button>
                  );
                })}
              </div>
              {errors.strategy && (
                <p className="flex items-center gap-1 text-sm text-editor-error">
                  <AlertCircle size={14} />
                  {errors.strategy}
                </p>
              )}
            </div>

            {/* Agent Configuration */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-editor-text">
                  Agent Roles
                </label>
                <span className="text-xs text-editor-muted">
                  {totalAgents} / {config.maxAgents} agents
                </span>
              </div>
              <div className="space-y-2">
                {config.agents.map((agent) => (
                  <div
                    key={agent.id}
                    className="flex items-center gap-3 p-3 bg-editor-surface rounded-lg border border-editor-border"
                  >
                    {/* Role Select */}
                    <select
                      value={agent.role}
                      onChange={(e) =>
                        handleAgentChange(agent.id, 'role', e.target.value as AgentRole)
                      }
                      className="flex-1 px-3 py-1.5 bg-editor-bg border border-editor-border rounded text-sm text-editor-text focus:outline-none focus:border-editor-accent"
                    >
                      {ALL_ROLES.map((role) => (
                        <option key={role} value={role}>
                          {ROLE_INFO[role].label}
                        </option>
                      ))}
                    </select>

                    {/* Count Input */}
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-editor-muted">Count:</span>
                      <input
                        type="number"
                        min={1}
                        max={config.maxAgents}
                        value={agent.count}
                        onChange={(e) =>
                          handleAgentChange(
                            agent.id,
                            'count',
                            Math.max(1, parseInt(e.target.value) || 1)
                          )
                        }
                        className="w-16 px-2 py-1.5 bg-editor-bg border border-editor-border rounded text-sm text-editor-text text-center focus:outline-none focus:border-editor-accent"
                      />
                    </div>

                    {/* Remove Button */}
                    <button
                      type="button"
                      onClick={() => handleRemoveAgent(agent.id)}
                      disabled={config.agents.length <= 1}
                      className="p-1.5 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))}
              </div>
              <button
                type="button"
                onClick={handleAddAgent}
                disabled={totalAgents >= config.maxAgents}
                className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <Plus size={16} />
                Add Agent Role
              </button>
              {errors.agents && (
                <p className="flex items-center gap-1 text-sm text-editor-error">
                  <AlertCircle size={14} />
                  {errors.agents}
                </p>
              )}
            </div>

            {/* Settings */}
            <div className="space-y-4">
              <label className="block text-sm font-medium text-editor-text">
                Settings
              </label>
              <div className="grid grid-cols-2 gap-4">
                {/* Max Agents */}
                <div className="space-y-2">
                  <label className="block text-xs text-editor-muted">
                    Max Agents
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={20}
                    value={config.maxAgents}
                    onChange={(e) =>
                      setConfig((prev) => ({
                        ...prev,
                        maxAgents: Math.max(1, parseInt(e.target.value) || 5),
                      }))
                    }
                    className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
                  />
                </div>

                {/* Timeout */}
                <div className="space-y-2">
                  <label className="block text-xs text-editor-muted">
                    Timeout (seconds)
                  </label>
                  <input
                    type="number"
                    min={10}
                    max={3600}
                    value={config.timeout}
                    onChange={(e) =>
                      setConfig((prev) => ({
                        ...prev,
                        timeout: Math.max(10, parseInt(e.target.value) || 300),
                      }))
                    }
                    className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent ${
                      errors.timeout ? 'border-editor-error' : 'border-editor-border'
                    }`}
                  />
                  {errors.timeout && (
                    <p className="flex items-center gap-1 text-xs text-editor-error">
                      <AlertCircle size={12} />
                      {errors.timeout}
                    </p>
                  )}
                </div>
              </div>

              {/* Synthesizer Toggle */}
              <div className="flex items-center gap-3 py-2">
                <input
                  type="checkbox"
                  id="useSynthesizer"
                  checked={config.useSynthesizer}
                  onChange={(e) =>
                    setConfig((prev) => ({
                      ...prev,
                      useSynthesizer: e.target.checked,
                    }))
                  }
                  className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
                />
                <label
                  htmlFor="useSynthesizer"
                  className="text-sm text-editor-text"
                >
                  Use synthesizer to combine agent outputs
                </label>
              </div>
            </div>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-editor-border bg-editor-surface/50 shrink-0">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => {
              if (validate()) {
                toast.info('Template saving not yet implemented');
              }
            }}
            className="flex items-center gap-2 px-4 py-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <Save size={16} />
            Save as Template
          </button>
          <button
            onClick={handleSubmit}
            disabled={isSaving}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 transition-colors"
          >
            {isSaving ? (
              <>
                <Loader2 size={16} className="animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <CheckCircle size={16} />
                Create Swarm
              </>
            )}
          </button>
        </div>
      </div>
    </>
  );
}

export default SwarmConfigModal;
