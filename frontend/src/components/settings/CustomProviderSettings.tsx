import { useState, useEffect } from 'react';
import { useAuthStore } from '../../store/authStore';
import { Server, Plus, X, CheckCircle, XCircle, RefreshCw, Trash2, ExternalLink, Play, Clock, Zap, Loader2, MessageSquare, AlertCircle } from 'lucide-react';
import { apiService } from '../../services/api';

interface CustomProvider {
  id: string;
  name: string;
  base_url: string;
  has_api_key: boolean;
  models: string[];
  supports_tools: boolean;
  supports_vision: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface TestResult {
  success: boolean;
  accessible: boolean;
  models_available: boolean;
  auth_required: boolean;
  model_count: number;
  message: string;
}

export function CustomProviderSettings() {
  const { accessToken } = useAuthStore();
  const [providers, setProviders] = useState<CustomProvider[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAdding, setIsAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // New provider form state
  const [newName, setNewName] = useState('');
  const [newBaseUrl, setNewBaseUrl] = useState('');
  const [newApiKey, setNewApiKey] = useState('');
  const [newSupportsTools, setNewSupportsTools] = useState(false);
  const [newSupportsVision, setNewSupportsVision] = useState(false);
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  const [testing, setTesting] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchProviders();
  }, [accessToken]);

  const fetchProviders = async () => {
    if (!accessToken) return;
    try {
      const response = await fetch('/api/v1/providers/custom', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        setProviders(data.providers || []);
      }
    } catch {
      setError('Failed to fetch custom providers');
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    if (!newBaseUrl || !accessToken) return;

    setTesting(true);
    setTestResult(null);
    setError(null);

    try {
      const response = await fetch('/api/v1/providers/custom/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          base_url: newBaseUrl,
          api_key: newApiKey,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setTestResult(data);
      } else {
        const data = await response.json();
        setError(data.error || 'Failed to test endpoint');
      }
    } catch {
      setError('Failed to test endpoint');
    } finally {
      setTesting(false);
    }
  };

  const handleAdd = async () => {
    if (!newName || !newBaseUrl || !accessToken) return;

    setSaving(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/providers/custom', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          name: newName,
          base_url: newBaseUrl,
          api_key: newApiKey,
          models: [],
          supports_tools: newSupportsTools,
          supports_vision: newSupportsVision,
        }),
      });

      if (response.ok) {
        // Reset form
        setNewName('');
        setNewBaseUrl('');
        setNewApiKey('');
        setNewSupportsTools(false);
        setNewSupportsVision(false);
        setTestResult(null);
        setIsAdding(false);
        await fetchProviders();
      } else {
        const data = await response.json();
        setError(data.error || 'Failed to add provider');
      }
    } catch {
      setError('Failed to add provider');
    } finally {
      setSaving(false);
    }
  };

  const handleRemove = async (id: string) => {
    if (!accessToken) return;

    try {
      const response = await fetch(`/api/v1/providers/custom/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      if (response.ok) {
        await fetchProviders();
      } else {
        const data = await response.json();
        setError(data.error || 'Failed to remove provider');
      }
    } catch {
      setError('Failed to remove provider');
    }
  };

  const handleFetchModels = async (id: string) => {
    if (!accessToken) return;

    try {
      const response = await fetch(`/api/v1/providers/custom/${id}/fetch-models`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      if (response.ok) {
        await fetchProviders();
      }
    } catch {
      // Silent failure - user can try again
    }
  };

  if (loading) {
    return (
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Server className="w-5 h-5" />
          <h2 className="text-xl font-semibold">Custom Providers</h2>
        </div>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
          <div className="flex items-center justify-center py-4">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center gap-2">
        <Server className="w-5 h-5" />
        <h2 className="text-xl font-semibold">Custom Providers</h2>
      </div>
      <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
        <p className="text-editor-muted text-sm mb-4">
          Add custom OpenAI-compatible endpoints like vLLM, llama.cpp server, LocalAI, or other compatible services.
        </p>

        {error && (
          <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-lg mb-4">
            <p className="text-sm text-red-400">{error}</p>
          </div>
        )}

        {/* Existing providers list */}
        <div className="space-y-3">
          {providers.map((provider) => (
            <CustomProviderCard
              key={provider.id}
              provider={provider}
              onRemove={() => handleRemove(provider.id)}
              onFetchModels={() => handleFetchModels(provider.id)}
            />
          ))}
        </div>

        {/* Add new provider form */}
        {isAdding ? (
          <div className="mt-4 p-4 bg-editor-bg rounded-lg border border-editor-border">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Add Custom Provider</h3>
              <button
                onClick={() => {
                  setIsAdding(false);
                  setNewName('');
                  setNewBaseUrl('');
                  setNewApiKey('');
                  setTestResult(null);
                  setError(null);
                }}
                className="p-1 hover:bg-editor-surface rounded"
              >
                <X className="w-4 h-4 text-editor-muted" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <input
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="My Local LLM"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Base URL</label>
                <input
                  type="url"
                  value={newBaseUrl}
                  onChange={(e) => setNewBaseUrl(e.target.value)}
                  placeholder="http://localhost:8080"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
                />
                <p className="text-xs text-editor-muted mt-1">
                  The base URL of your OpenAI-compatible API (without /v1)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">API Key (optional)</label>
                <input
                  type="password"
                  value={newApiKey}
                  onChange={(e) => setNewApiKey(e.target.value)}
                  placeholder="Leave empty if not required"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
                />
              </div>

              <div className="flex gap-4">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={newSupportsTools}
                    onChange={(e) => setNewSupportsTools(e.target.checked)}
                    className="rounded border-editor-border"
                  />
                  Supports Tool Calling
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={newSupportsVision}
                    onChange={(e) => setNewSupportsVision(e.target.checked)}
                    className="rounded border-editor-border"
                  />
                  Supports Vision
                </label>
              </div>

              {/* Test result */}
              {testResult && (
                <div className={`p-3 rounded-lg ${
                  testResult.success
                    ? 'bg-green-500/20 border border-green-500/50'
                    : 'bg-yellow-500/20 border border-yellow-500/50'
                }`}>
                  <div className="flex items-center gap-2 mb-1">
                    {testResult.success ? (
                      <CheckCircle className="w-4 h-4 text-green-400" />
                    ) : (
                      <XCircle className="w-4 h-4 text-yellow-400" />
                    )}
                    <span className="text-sm font-medium">
                      {testResult.success ? 'Connection successful!' : 'Connection issue'}
                    </span>
                  </div>
                  <p className="text-xs text-editor-muted">{testResult.message}</p>
                  {testResult.model_count > 0 && (
                    <p className="text-xs text-editor-muted mt-1">
                      Found {testResult.model_count} models
                    </p>
                  )}
                </div>
              )}

              <div className="flex justify-end gap-2">
                <button
                  onClick={handleTest}
                  disabled={!newBaseUrl || testing}
                  className="px-4 py-2 text-sm border border-editor-border rounded-lg hover:bg-editor-surface disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {testing ? (
                    <RefreshCw className="w-4 h-4 animate-spin" />
                  ) : (
                    <ExternalLink className="w-4 h-4" />
                  )}
                  Test Connection
                </button>
                <button
                  onClick={handleAdd}
                  disabled={!newName || !newBaseUrl || saving}
                  className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? 'Adding...' : 'Add Provider'}
                </button>
              </div>
            </div>
          </div>
        ) : (
          <button
            onClick={() => setIsAdding(true)}
            className="mt-4 w-full px-4 py-3 border border-dashed border-editor-border rounded-lg text-sm text-editor-muted hover:border-editor-accent hover:text-editor-text transition-colors flex items-center justify-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Add Custom Provider
          </button>
        )}
      </div>
    </section>
  );
}

function CustomProviderCard({
  provider,
  onRemove,
  onFetchModels,
}: {
  provider: CustomProvider;
  onRemove: () => void;
  onFetchModels: () => void;
}) {
  const [fetching, setFetching] = useState(false);
  const [showTestPanel, setShowTestPanel] = useState(false);

  const handleFetchModels = async () => {
    setFetching(true);
    await onFetchModels();
    setFetching(false);
  };

  return (
    <div className="p-3 bg-editor-bg rounded-lg">
      <div className="flex items-center justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="font-medium truncate">{provider.name}</p>
            {provider.has_api_key && (
              <span className="text-xs px-1.5 py-0.5 rounded bg-green-500/20 text-green-400">
                API Key
              </span>
            )}
            {provider.supports_tools && (
              <span className="text-xs px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
                Tools
              </span>
            )}
            {provider.supports_vision && (
              <span className="text-xs px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-400">
                Vision
              </span>
            )}
          </div>
          <p className="text-sm text-editor-muted truncate">{provider.base_url}</p>
          {provider.models && provider.models.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {provider.models.slice(0, 3).map((model) => (
                <span
                  key={model}
                  className="text-xs px-1.5 py-0.5 rounded bg-editor-surface border border-editor-border"
                >
                  {model}
                </span>
              ))}
              {provider.models.length > 3 && (
                <span className="text-xs px-1.5 py-0.5 text-editor-muted">
                  +{provider.models.length - 3} more
                </span>
              )}
            </div>
          )}
        </div>
        <div className="flex items-center gap-2 ml-4">
          {provider.models && provider.models.length > 0 && (
            <button
              onClick={() => setShowTestPanel(!showTestPanel)}
              className="p-2 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text transition-colors"
              title="Test provider with a prompt"
            >
              <Play className="w-4 h-4" />
            </button>
          )}
          <button
            onClick={handleFetchModels}
            disabled={fetching}
            className="p-2 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text transition-colors"
            title="Fetch models from endpoint"
          >
            <RefreshCw className={`w-4 h-4 ${fetching ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={onRemove}
            className="p-2 hover:bg-editor-surface rounded text-red-400 hover:text-red-300 transition-colors"
            title="Remove provider"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Test panel for custom provider */}
      {showTestPanel && provider.models && provider.models.length > 0 && (
        <CustomProviderTestPanel
          providerId={provider.id}
          providerName={provider.name}
          models={provider.models}
        />
      )}
    </div>
  );
}

interface CustomTestResult {
  id: string;
  model: string;
  prompt: string;
  response: string;
  latencyMs: number;
  tokensUsed: { input: number; output: number };
  timestamp: Date;
}

const PRESET_PROMPTS = [
  { label: 'Quick check', value: 'Hello, respond with one word.' },
  { label: 'Math test', value: 'What is 2+2? Just the number.' },
  { label: 'Creative', value: 'Write a haiku about coding.' },
  { label: 'Custom', value: '' },
];

function CustomProviderTestPanel({
  providerId,
  providerName,
  models,
}: {
  providerId: string;
  providerName: string;
  models: string[];
}) {
  const { accessToken } = useAuthStore();
  const [selectedModel, setSelectedModel] = useState(models[0] || '');
  const [selectedPreset, setSelectedPreset] = useState(0);
  const [customPrompt, setCustomPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [testHistory, setTestHistory] = useState<CustomTestResult[]>([]);

  useEffect(() => {
    if (accessToken) {
      apiService.setToken(accessToken);
    }
  }, [accessToken]);

  const getPrompt = () => {
    if (selectedPreset === PRESET_PROMPTS.length - 1) {
      return customPrompt;
    }
    return PRESET_PROMPTS[selectedPreset].value;
  };

  const handleTest = async () => {
    const prompt = getPrompt();
    if (!prompt || !selectedModel) return;

    setLoading(true);
    setError(null);

    try {
      // Custom providers use the provider name prefixed with "custom_"
      const response = await apiService.testProviderPrompt(
        `custom_${providerId}`,
        selectedModel,
        prompt
      );

      if (response.error) {
        setError(response.error);
        return;
      }

      if (response.data) {
        const result: CustomTestResult = {
          id: `${Date.now()}`,
          model: selectedModel,
          prompt,
          response: response.data.response,
          latencyMs: response.data.latency_ms,
          tokensUsed: response.data.tokens_used,
          timestamp: new Date(),
        };

        setTestHistory(prev => [result, ...prev.slice(0, 4)]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send test prompt');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-3 pt-3 border-t border-editor-border space-y-3">
      <div className="flex items-center gap-2 text-xs text-editor-muted">
        <Play className="w-3 h-3" />
        <span>Test {providerName}</span>
      </div>

      {/* Model selector */}
      <div>
        <label className="block text-xs text-editor-muted mb-1">Model</label>
        <select
          value={selectedModel}
          onChange={(e) => setSelectedModel(e.target.value)}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
        >
          {models.map((model) => (
            <option key={model} value={model}>
              {model}
            </option>
          ))}
        </select>
      </div>

      {/* Preset prompt selector */}
      <div>
        <label className="block text-xs text-editor-muted mb-1">Test Prompt</label>
        <div className="flex flex-wrap gap-2 mb-2">
          {PRESET_PROMPTS.map((preset, index) => (
            <button
              key={index}
              onClick={() => setSelectedPreset(index)}
              className={`px-2 py-1 text-xs rounded-full border transition-colors ${
                selectedPreset === index
                  ? 'bg-editor-accent/20 border-editor-accent text-editor-accent'
                  : 'border-editor-border text-editor-muted hover:border-editor-accent hover:text-editor-text'
              }`}
            >
              {preset.label}
            </button>
          ))}
        </div>

        {selectedPreset === PRESET_PROMPTS.length - 1 ? (
          <textarea
            value={customPrompt}
            onChange={(e) => setCustomPrompt(e.target.value)}
            placeholder="Enter your test prompt..."
            maxLength={1000}
            rows={3}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm resize-none"
          />
        ) : (
          <div className="px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-muted italic">
            {PRESET_PROMPTS[selectedPreset].value}
          </div>
        )}
      </div>

      {/* Error display */}
      {error && (
        <div className="flex items-start gap-2 p-2 bg-red-500/20 border border-red-500/50 rounded-lg">
          <AlertCircle className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
          <p className="text-xs text-red-400">{error}</p>
        </div>
      )}

      {/* Send button */}
      <button
        onClick={handleTest}
        disabled={loading || !getPrompt() || !selectedModel}
        className="w-full px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" />
            Sending...
          </>
        ) : (
          <>
            <Play className="w-4 h-4" />
            Send Test
          </>
        )}
      </button>

      {/* Test history */}
      {testHistory.length > 0 && (
        <div className="mt-3 border-t border-editor-border pt-3">
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-xs font-medium text-editor-muted flex items-center gap-1">
              <MessageSquare className="w-3 h-3" />
              Recent Tests ({testHistory.length})
            </h4>
            <button
              onClick={() => setTestHistory([])}
              className="text-xs text-editor-muted hover:text-red-400 transition-colors"
            >
              Clear
            </button>
          </div>
          <div className="space-y-2">
            {testHistory.map((result) => (
              <div key={result.id} className="p-2 bg-editor-surface rounded-lg border border-editor-border">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-medium text-editor-accent truncate">
                    {result.model}
                  </span>
                  <div className="flex items-center gap-3 text-xs text-editor-muted">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {result.latencyMs}ms
                    </span>
                    <span className="flex items-center gap-1">
                      <Zap className="w-3 h-3" />
                      {result.tokensUsed.input}/{result.tokensUsed.output}
                    </span>
                  </div>
                </div>
                <div className="mb-2">
                  <span className="text-[10px] uppercase tracking-wider text-editor-muted">Prompt</span>
                  <p className="text-xs text-editor-muted truncate">{result.prompt}</p>
                </div>
                <div>
                  <span className="text-[10px] uppercase tracking-wider text-editor-muted">Response</span>
                  <p className="text-xs text-editor-text whitespace-pre-wrap break-words max-h-24 overflow-y-auto">
                    {result.response}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default CustomProviderSettings;
