import { useState, useEffect } from 'react';
import { useAuthStore } from '../../store/authStore';
import { apiService } from '../../services/api';
import { Play, ChevronDown, ChevronUp, AlertCircle, Loader2 } from 'lucide-react';
import { ProviderTestHistory, TestResult } from './ProviderTestHistory';

interface ProviderTestPanelProps {
  provider: string;
  models: Array<{ id: string; name: string }>;
  apiKey?: string;
  collapsed?: boolean;
}

const PRESET_PROMPTS = [
  { label: 'Quick check', value: 'Hello, respond with one word.' },
  { label: 'Math test', value: 'What is 2+2? Just the number.' },
  { label: 'Creative', value: 'Write a haiku about coding.' },
  { label: 'Custom', value: '' },
];

export function ProviderTestPanel({
  provider,
  models,
  apiKey,
  collapsed = true,
}: ProviderTestPanelProps) {
  const { accessToken } = useAuthStore();
  const [isExpanded, setIsExpanded] = useState(!collapsed);
  const [selectedModel, setSelectedModel] = useState(models[0]?.id || '');
  const [selectedPreset, setSelectedPreset] = useState(0);
  const [customPrompt, setCustomPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [testHistory, setTestHistory] = useState<TestResult[]>([]);

  // Set token on apiService when available
  useEffect(() => {
    if (accessToken) {
      apiService.setToken(accessToken);
    }
  }, [accessToken]);

  // Update selected model when models change
  useEffect(() => {
    if (models.length > 0 && !models.find(m => m.id === selectedModel)) {
      setSelectedModel(models[0].id);
    }
  }, [models, selectedModel]);

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
      const response = await apiService.testProviderPrompt(
        provider.toLowerCase().replace(' ', '_'),
        selectedModel,
        prompt,
        apiKey
      );

      if (response.error) {
        setError(response.error);
        return;
      }

      if (response.data) {
        const result: TestResult = {
          id: `${Date.now()}`,
          provider,
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

  const clearHistory = () => {
    setTestHistory([]);
  };

  if (models.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 border border-editor-border rounded-lg overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-3 py-2 flex items-center justify-between bg-editor-bg hover:bg-editor-surface transition-colors text-sm"
      >
        <div className="flex items-center gap-2">
          <Play className="w-4 h-4 text-editor-muted" />
          <span className="font-medium">Test Provider</span>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 text-editor-muted" />
        ) : (
          <ChevronDown className="w-4 h-4 text-editor-muted" />
        )}
      </button>

      {isExpanded && (
        <div className="p-3 bg-editor-bg border-t border-editor-border space-y-3">
          {/* Model selector */}
          <div>
            <label className="block text-xs text-editor-muted mb-1">Model</label>
            <select
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
            >
              {models.map((model) => (
                <option key={model.id} value={model.id}>
                  {model.name || model.id}
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

            {/* Custom prompt input or selected preset display */}
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
            {selectedPreset === PRESET_PROMPTS.length - 1 && (
              <div className="text-xs text-editor-muted mt-1 text-right">
                {customPrompt.length}/1000
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
            <ProviderTestHistory results={testHistory} onClear={clearHistory} />
          )}
        </div>
      )}
    </div>
  );
}

export default ProviderTestPanel;
