import { useState, useEffect } from 'react';
import { ChevronDown, Check, Server, Cpu, Zap } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';
import { useAppStore } from '../../../store';
import type { AgentStepConfig as AgentConfig } from '../../../types/workflow';
import { StateVariableInput } from './StateVariablePicker';

interface AgentStepConfigProps {
  nodeId: string;
}

export function AgentStepConfig({ nodeId }: AgentStepConfigProps) {
  const { getSelectedNode, updateNodeConfig } = useWorkflowStore();
  const { providers, loadProviders } = useAppStore();
  const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false);

  const node = getSelectedNode();
  const nodeData = node?.data as { config?: { agentConfig?: AgentConfig } } | undefined;
  const config = nodeData?.config?.agentConfig || node?.config?.agentConfig;

  // Load providers on mount
  useEffect(() => {
    if (providers.length === 0) {
      loadProviders();
    }
  }, [providers.length, loadProviders]);

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<AgentConfig>) => {
    updateNodeConfig(nodeId, {
      agentConfig: { ...config, ...updates },
    });
  };

  const currentProvider = providers.find((p) => p.name === config.provider);
  const currentModel = currentProvider?.models.find((m) => m.id === config.model);

  const getProviderIcon = (name: string) => {
    if (name === 'ollama') return <Cpu size={14} className="text-editor-accent" />;
    if (name === 'groq') return <Zap size={14} className="text-yellow-500" />;
    return <Server size={14} className="text-editor-muted" />;
  };

  return (
    <div className="space-y-4">
      {/* Provider Selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Provider
        </label>
        <select
          value={config.provider}
          onChange={(e) => {
            const newProvider = providers.find((p) => p.name === e.target.value);
            updateConfig({
              provider: e.target.value,
              model: newProvider?.models[0]?.id || '',
            });
          }}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
        >
          {providers.map((provider) => (
            <option key={provider.name} value={provider.name}>
              {provider.name}
            </option>
          ))}
        </select>
      </div>

      {/* Model Selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Model
        </label>
        <div className="relative">
          <button
            type="button"
            onClick={() => setIsModelDropdownOpen(!isModelDropdownOpen)}
            className="w-full flex items-center gap-2 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text hover:border-editor-accent transition-colors text-left"
          >
            {getProviderIcon(config.provider)}
            <span className="flex-1 truncate">
              {currentModel?.name || config.model || 'Select model...'}
            </span>
            <ChevronDown
              size={14}
              className={`text-editor-muted transition-transform ${isModelDropdownOpen ? 'rotate-180' : ''}`}
            />
          </button>

          {isModelDropdownOpen && currentProvider && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 max-h-64 overflow-y-auto">
              {currentProvider.models.length > 0 ? (
                currentProvider.models.map((model) => (
                  <button
                    key={model.id}
                    onClick={() => {
                      updateConfig({ model: model.id });
                      setIsModelDropdownOpen(false);
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 hover:bg-editor-surface text-left transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="text-sm text-editor-text truncate">{model.name}</div>
                      <div className="text-xs text-editor-muted">
                        {model.context_window.toLocaleString()} ctx
                        {model.supports_tools && ' | Tools'}
                        {model.supports_vision && ' | Vision'}
                      </div>
                    </div>
                    {config.model === model.id && (
                      <Check size={14} className="text-editor-accent flex-shrink-0 ml-2" />
                    )}
                  </button>
                ))
              ) : (
                <div className="px-3 py-4 text-sm text-editor-muted text-center">
                  No models available for this provider
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* System Prompt */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          System Prompt
          <span className="text-editor-muted font-normal ml-1">(optional)</span>
        </label>
        <textarea
          value={config.systemPrompt || ''}
          onChange={(e) => updateConfig({ systemPrompt: e.target.value })}
          placeholder="You are a helpful assistant..."
          rows={3}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none font-mono text-sm"
        />
      </div>

      {/* User Prompt */}
      <StateVariableInput
        label="User Prompt"
        value={config.prompt}
        onChange={(value) => updateConfig({ prompt: value })}
        nodeId={nodeId}
        placeholder="Enter the prompt for the agent..."
        rows={4}
      />

      {/* Advanced Settings */}
      <div className="space-y-4 pt-2">
        <h4 className="text-sm font-medium text-editor-text">Advanced Settings</h4>

        <div className="grid grid-cols-2 gap-4">
          {/* Temperature */}
          <div className="space-y-2">
            <label className="block text-xs text-editor-muted">
              Temperature ({(config.temperature || 0.7).toFixed(1)})
            </label>
            <input
              type="range"
              value={config.temperature || 0.7}
              onChange={(e) => updateConfig({ temperature: parseFloat(e.target.value) })}
              min={0}
              max={2}
              step={0.1}
              className="w-full accent-editor-accent"
            />
          </div>

          {/* Max Tokens */}
          <div className="space-y-2">
            <label className="block text-xs text-editor-muted">
              Max Tokens
            </label>
            <input
              type="number"
              value={config.maxTokens || 4096}
              onChange={(e) => updateConfig({ maxTokens: parseInt(e.target.value) || 4096 })}
              min={1}
              max={100000}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent text-sm"
            />
          </div>
        </div>
      </div>

      {/* Tools */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Available Tools
          <span className="text-editor-muted font-normal ml-1">(optional)</span>
        </label>
        <input
          type="text"
          value={config.tools?.join(', ') || ''}
          onChange={(e) =>
            updateConfig({
              tools: e.target.value
                .split(',')
                .map((t) => t.trim())
                .filter(Boolean),
            })
          }
          placeholder="tool1, tool2, tool3..."
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent text-sm"
        />
        <p className="text-xs text-editor-muted">
          Comma-separated list of tool names the agent can use
        </p>
      </div>

      {/* Output Key */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Output Key
          <span className="text-editor-muted font-normal ml-1">(optional)</span>
        </label>
        <input
          type="text"
          value={config.outputKey || ''}
          onChange={(e) => updateConfig({ outputKey: e.target.value })}
          placeholder="agentResponse"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
        <p className="text-xs text-editor-muted">
          Store the agent's response in workflow state with this key
        </p>
      </div>
    </div>
  );
}
