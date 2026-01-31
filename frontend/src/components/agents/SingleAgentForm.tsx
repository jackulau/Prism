import { useState, useEffect, useRef } from 'react';
import { Play, ChevronDown, Server, Check, Cpu, Zap, RotateCcw, Settings } from 'lucide-react';
import { useAgentStore } from '../../store/agentStore';
import { useAppStore } from '../../store';

interface SingleAgentFormProps {
  onSubmit: () => void;
  disabled?: boolean;
}

export function SingleAgentForm({ onSubmit, disabled }: SingleAgentFormProps) {
  const { config, setConfig, resetConfig } = useAgentStore();
  const { providers, loadProviders } = useAppStore();
  const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsModelDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleModelSelect = (providerName: string, modelId: string) => {
    setConfig({ provider: providerName, model: modelId });
    setIsModelDropdownOpen(false);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!config.model) return;
    onSubmit();
  };

  const getProviderIcon = (name: string) => {
    if (name === 'ollama') return <Cpu size={12} className="text-editor-accent" />;
    if (name === 'groq') return <Zap size={12} className="text-yellow-500" />;
    return <Server size={12} />;
  };

  const currentProvider = providers.find(p => p.name === config.provider);
  const currentModel = currentProvider?.models.find(m => m.id === config.model);
  const displayModelName = currentModel?.name || config.model || 'Select Model';

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Agent Name */}
      <div>
        <label htmlFor="agent-name" className="block text-sm font-medium text-editor-text mb-1.5">
          Agent Name
        </label>
        <input
          id="agent-name"
          type="text"
          value={config.name}
          onChange={(e) => setConfig({ name: e.target.value })}
          placeholder="My Agent"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent transition-colors"
          disabled={disabled}
        />
      </div>

      {/* Model Selection */}
      <div>
        <label className="block text-sm font-medium text-editor-text mb-1.5">
          Model
        </label>
        <div ref={dropdownRef} className="relative">
          <button
            type="button"
            onClick={() => setIsModelDropdownOpen(!isModelDropdownOpen)}
            className="w-full flex items-center gap-2 px-3 py-2 rounded-lg bg-editor-surface border border-editor-border hover:border-editor-accent transition-colors text-sm"
            disabled={disabled}
          >
            {config.provider === 'ollama' ? (
              <Cpu size={14} className="text-editor-accent flex-shrink-0" />
            ) : config.provider === 'groq' ? (
              <Zap size={14} className="text-yellow-500 flex-shrink-0" />
            ) : (
              <Server size={14} className="text-editor-muted flex-shrink-0" />
            )}
            <span className="text-editor-text truncate flex-1 text-left">{displayModelName}</span>
            <ChevronDown size={14} className={`text-editor-muted transition-transform flex-shrink-0 ${isModelDropdownOpen ? 'rotate-180' : ''}`} />
          </button>

          {isModelDropdownOpen && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 max-h-60 overflow-y-auto">
              {providers.length === 0 ? (
                <div className="px-3 py-4 text-center text-sm text-editor-muted">
                  Loading providers...
                </div>
              ) : (
                providers.map((provider) => (
                  <div key={provider.name}>
                    <div className="px-3 py-2 text-xs font-semibold text-editor-muted uppercase bg-editor-surface/50 border-b border-editor-border sticky top-0 flex items-center gap-2">
                      {getProviderIcon(provider.name)}
                      <span>{provider.name}</span>
                    </div>
                    {provider.models.length > 0 ? (
                      provider.models.map((model) => (
                        <button
                          key={`${provider.name}-${model.id}`}
                          type="button"
                          onClick={() => handleModelSelect(provider.name, model.id)}
                          className="w-full flex items-center justify-between px-3 py-2 hover:bg-editor-surface text-left transition-colors"
                        >
                          <div className="flex-1 min-w-0">
                            <div className="text-sm text-editor-text truncate">{model.name}</div>
                            <div className="text-xs text-editor-muted">
                              {model.context_window.toLocaleString()} ctx
                              {model.supports_tools && ' | Tools'}
                            </div>
                          </div>
                          {config.provider === provider.name && config.model === model.id && (
                            <Check size={14} className="text-editor-accent flex-shrink-0 ml-2" />
                          )}
                        </button>
                      ))
                    ) : (
                      <div className="px-3 py-2 text-sm text-editor-muted italic">
                        No models available
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>

      {/* System Prompt */}
      <div>
        <label htmlFor="system-prompt" className="block text-sm font-medium text-editor-text mb-1.5">
          System Prompt
        </label>
        <textarea
          id="system-prompt"
          value={config.systemPrompt}
          onChange={(e) => setConfig({ systemPrompt: e.target.value })}
          placeholder="You are a helpful AI assistant..."
          rows={4}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent transition-colors resize-none"
          disabled={disabled}
        />
      </div>

      {/* Advanced Settings Toggle */}
      <button
        type="button"
        onClick={() => setShowAdvanced(!showAdvanced)}
        className="flex items-center gap-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
      >
        <Settings size={14} />
        <span>Advanced Settings</span>
        <ChevronDown size={14} className={`transition-transform ${showAdvanced ? 'rotate-180' : ''}`} />
      </button>

      {/* Advanced Settings */}
      {showAdvanced && (
        <div className="space-y-4 p-4 bg-editor-surface/50 rounded-lg border border-editor-border">
          {/* Temperature */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label htmlFor="temperature" className="text-sm font-medium text-editor-text">
                Temperature
              </label>
              <span className="text-sm text-editor-muted">{config.temperature.toFixed(1)}</span>
            </div>
            <input
              id="temperature"
              type="range"
              min="0"
              max="2"
              step="0.1"
              value={config.temperature}
              onChange={(e) => setConfig({ temperature: parseFloat(e.target.value) })}
              className="w-full accent-editor-accent"
              disabled={disabled}
            />
            <div className="flex justify-between text-xs text-editor-muted mt-1">
              <span>Precise</span>
              <span>Creative</span>
            </div>
          </div>

          {/* Max Tokens */}
          <div>
            <label htmlFor="max-tokens" className="block text-sm font-medium text-editor-text mb-1.5">
              Max Tokens
            </label>
            <input
              id="max-tokens"
              type="number"
              min="256"
              max="128000"
              step="256"
              value={config.maxTokens}
              onChange={(e) => setConfig({ maxTokens: parseInt(e.target.value) || 4096 })}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent transition-colors"
              disabled={disabled}
            />
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-3 pt-2">
        <button
          type="submit"
          disabled={disabled || !config.model}
          className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          <Play size={18} />
          Run Agent
        </button>
        <button
          type="button"
          onClick={resetConfig}
          disabled={disabled}
          className="p-2.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          title="Reset to defaults"
        >
          <RotateCcw size={18} />
        </button>
      </div>
    </form>
  );
}
