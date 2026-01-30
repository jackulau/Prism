import { useEffect } from 'react';
import { Server, Cpu, Zap, Timer, Thermometer, Hash } from 'lucide-react';
import { useBatchStore } from '../../store/batchStore';
import { useAppStore } from '../../store';

export function BatchConfigForm() {
  const { config, setConfig, isRunning } = useBatchStore();
  const { providers, loadProviders } = useAppStore();

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

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
          value={config.provider}
          onChange={(e) => {
            const provider = providers.find((p) => p.name === e.target.value);
            setConfig({
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
          value={config.model}
          onChange={(e) => setConfig({ model: e.target.value })}
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
            {getProviderIcon(config.provider)}
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
          value={config.maxConcurrent}
          onChange={(e) => setConfig({ maxConcurrent: parseInt(e.target.value) || 1 })}
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
          value={config.timeout / 1000}
          onChange={(e) => setConfig({ timeout: (parseInt(e.target.value) || 120) * 1000 })}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        />
      </div>

      {/* Temperature */}
      <div className="space-y-2">
        <label className="text-xs text-editor-muted flex items-center gap-1">
          <Thermometer size={12} />
          Temperature: {config.temperature}
        </label>
        <input
          type="range"
          min={0}
          max={2}
          step={0.1}
          value={config.temperature}
          onChange={(e) => setConfig({ temperature: parseFloat(e.target.value) })}
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
          value={config.maxTokens}
          onChange={(e) => setConfig({ maxTokens: parseInt(e.target.value) || 4096 })}
          disabled={isRunning}
          className="w-full px-3 py-2 rounded-lg bg-editor-surface border border-editor-border text-editor-text text-sm focus:border-editor-accent focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        />
      </div>
    </div>
  );
}
