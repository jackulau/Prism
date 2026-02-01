import { useEffect, useState } from 'react';
import { Server, Cpu, Zap, Timer, Thermometer, Hash } from 'lucide-react';
import { useBatchStore } from '../../store/batchStore';
import { useAppStore } from '../../store';
import type { BatchExecutionConfig } from '../../types/batch';

export function BatchConfigForm() {
  const { execution } = useBatchStore();
  const { providers, loadProviders } = useAppStore();

  // Local config state
  const [config, setConfig] = useState<BatchExecutionConfig>({
    provider: execution.config.provider || execution.config.defaultProvider || 'ollama',
    model: execution.config.model || execution.config.defaultModel || '',
    maxConcurrent: execution.config.maxConcurrent || execution.config.maxConcurrency || 3,
    timeout: execution.config.timeout || execution.config.timeoutMs || 120000,
    temperature: execution.config.temperature || 0.7,
    maxTokens: execution.config.maxTokens || 4096,
  });

  const isRunning = execution.status === 'running';

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Update local config when execution config changes
  useEffect(() => {
    setConfig({
      provider: execution.config.provider || execution.config.defaultProvider || config.provider,
      model: execution.config.model || execution.config.defaultModel || config.model,
      maxConcurrent: execution.config.maxConcurrent || execution.config.maxConcurrency || config.maxConcurrent,
      timeout: execution.config.timeout || execution.config.timeoutMs || config.timeout,
      temperature: execution.config.temperature ?? config.temperature,
      maxTokens: execution.config.maxTokens ?? config.maxTokens,
    });
  }, [execution.config]);

  const updateConfig = (updates: Partial<BatchExecutionConfig>) => {
    setConfig((prev) => ({ ...prev, ...updates }));
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
      <h3 className="text-sm font-semibold text-editor-text flex items-center gap-2">
        <Server size={16} />
        Batch Configuration
      </h3>

      {/* Provider Selection */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted">Provider</label>
        <select
          value={config.provider || ''}
          onChange={(e) => {
            const provider = providers.find((p) => p.name === e.target.value);
            updateConfig({
              provider: e.target.value,
              model: provider?.models[0]?.id || '',
            });
          }}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {providers.map((p) => (
            <option key={p.name} value={p.name}>
              {p.name}
            </option>
          ))}
        </select>
      </div>

      {/* Model Selection */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted">Model</label>
        <select
          value={config.model || ''}
          onChange={(e) => updateConfig({ model: e.target.value })}
          disabled={isRunning || !currentProvider?.models.length}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {currentProvider?.models.length === 0 && (
            <option value="">No models available</option>
          )}
          {currentProvider?.models.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>
        {currentModel && (
          <div className="flex items-center gap-1 text-xs text-editor-muted">
            {getProviderIcon(config.provider || '')}
            <span>{currentModel.context_window.toLocaleString()} ctx</span>
            {currentModel.supports_tools && <span>| Tools</span>}
            {currentModel.supports_vision && <span>| Vision</span>}
          </div>
        )}
      </div>

      {/* Max Concurrent */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted flex items-center gap-1">
          <Hash size={12} />
          Max Concurrent Tasks
        </label>
        <input
          type="number"
          min={1}
          max={10}
          value={config.maxConcurrent || 3}
          onChange={(e) => updateConfig({ maxConcurrent: parseInt(e.target.value) || 1 })}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <p className="text-xs text-editor-muted">
          Number of tasks to run simultaneously
        </p>
      </div>

      {/* Timeout */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted flex items-center gap-1">
          <Timer size={12} />
          Timeout (seconds)
        </label>
        <input
          type="number"
          min={30}
          max={600}
          value={(config.timeout || 120000) / 1000}
          onChange={(e) => updateConfig({ timeout: (parseInt(e.target.value) || 120) * 1000 })}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        />
      </div>

      {/* Temperature */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted flex items-center gap-1">
          <Thermometer size={12} />
          Temperature: {config.temperature?.toFixed(1) || '0.7'}
        </label>
        <input
          type="range"
          min={0}
          max={2}
          step={0.1}
          value={config.temperature || 0.7}
          onChange={(e) => updateConfig({ temperature: parseFloat(e.target.value) })}
          disabled={isRunning}
          className="w-full accent-editor-accent disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <div className="flex justify-between text-xs text-editor-muted">
          <span>Precise</span>
          <span>Creative</span>
        </div>
      </div>

      {/* Max Tokens */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted">Max Tokens</label>
        <input
          type="number"
          min={256}
          max={32768}
          step={256}
          value={config.maxTokens || 4096}
          onChange={(e) => updateConfig({ maxTokens: parseInt(e.target.value) || 4096 })}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        />
      </div>
    </div>
  );
}
