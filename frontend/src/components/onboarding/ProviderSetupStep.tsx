import { useState, useEffect } from 'react';
import { Key, ExternalLink, CheckCircle, Cpu, ChevronDown, ChevronUp } from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { apiService } from '../../services/api';

interface ProviderSetupStepProps {
  onNext: () => void;
  onSkip: () => void;
}

interface ProviderConfig {
  name: string;
  key: string;
  placeholder: string;
  keyUrl: string;
  description: string;
}

const PROVIDERS: ProviderConfig[] = [
  {
    name: 'OpenAI',
    key: 'openai',
    placeholder: 'sk-...',
    keyUrl: 'https://platform.openai.com/api-keys',
    description: 'GPT-4.1, o3, o4-mini',
  },
  {
    name: 'Anthropic',
    key: 'anthropic',
    placeholder: 'sk-ant-...',
    keyUrl: 'https://console.anthropic.com/settings/keys',
    description: 'Claude Opus/Sonnet 4.5',
  },
  {
    name: 'Google AI',
    key: 'google_ai',
    placeholder: 'AIza...',
    keyUrl: 'https://aistudio.google.com/app/apikey',
    description: 'Gemini 2.5 Flash/Pro',
  },
  {
    name: 'OpenRouter',
    key: 'openrouter',
    placeholder: 'sk-or-...',
    keyUrl: 'https://openrouter.ai/keys',
    description: '200+ models',
  },
];

export function ProviderSetupStep({ onNext, onSkip }: ProviderSetupStepProps) {
  const { accessToken } = useAuthStore();
  const [apiKeys, setApiKeys] = useState<Record<string, string>>({});
  const [savedKeys, setSavedKeys] = useState<Record<string, boolean>>({});
  const [saving, setSaving] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [ollamaConnected, setOllamaConnected] = useState(false);
  const [showMoreProviders, setShowMoreProviders] = useState(false);

  useEffect(() => {
    if (accessToken) {
      apiService.setToken(accessToken);
      checkOllama();
    }
  }, [accessToken]);

  const checkOllama = async () => {
    try {
      const response = await fetch('/api/v1/providers', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        const ollama = data.providers?.find((p: { name: string }) => p.name === 'ollama');
        setOllamaConnected(ollama?.models?.length > 0);
      }
    } catch {
      // Ollama check failed - will show as not connected
    }
  };

  const handleSaveKey = async (providerKey: string) => {
    const value = apiKeys[providerKey];
    if (!value || !accessToken) return;

    setSaving(providerKey);
    setError(null);

    try {
      const response = await apiService.setProviderKey(providerKey, value);
      if (response.error) {
        setError(response.error);
      } else {
        setSavedKeys((prev) => ({ ...prev, [providerKey]: true }));
        setApiKeys((prev) => ({ ...prev, [providerKey]: '' }));
      }
    } catch {
      setError('Failed to save API key');
    } finally {
      setSaving(null);
    }
  };

  const hasAnyProvider = Object.values(savedKeys).some(Boolean) || ollamaConnected;
  const visibleProviders = showMoreProviders ? PROVIDERS : PROVIDERS.slice(0, 2);

  return (
    <div className="flex flex-col items-center min-h-[400px] px-6 py-8 animate-fade-in">
      {/* Header */}
      <div className="w-14 h-14 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-6">
        <Key className="w-7 h-7" />
      </div>

      <h2 className="text-2xl font-bold text-editor-text mb-2 text-center">
        Connect an LLM Provider
      </h2>
      <p className="text-editor-muted text-center max-w-md mb-8">
        Add at least one API key to start using AI features. You can add more providers later in Settings.
      </p>

      {/* Ollama status */}
      {ollamaConnected && (
        <div className="w-full max-w-md mb-6 p-4 bg-green-500/10 border border-green-500/20 rounded-lg flex items-center gap-3">
          <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
          <div>
            <p className="font-medium text-green-400">Ollama Connected</p>
            <p className="text-sm text-editor-muted">Local models are available</p>
          </div>
        </div>
      )}

      {/* Provider inputs */}
      <div className="w-full max-w-md space-y-4">
        {visibleProviders.map((provider) => (
          <ProviderKeyInput
            key={provider.key}
            provider={provider}
            value={apiKeys[provider.key] || ''}
            onChange={(value) => setApiKeys((prev) => ({ ...prev, [provider.key]: value }))}
            onSave={() => handleSaveKey(provider.key)}
            isSaved={savedKeys[provider.key]}
            isSaving={saving === provider.key}
          />
        ))}

        {!showMoreProviders && (
          <button
            onClick={() => setShowMoreProviders(true)}
            className="w-full py-2 text-sm text-editor-muted hover:text-editor-text flex items-center justify-center gap-1 transition-colors"
          >
            Show more providers
            <ChevronDown className="w-4 h-4" />
          </button>
        )}

        {showMoreProviders && (
          <button
            onClick={() => setShowMoreProviders(false)}
            className="w-full py-2 text-sm text-editor-muted hover:text-editor-text flex items-center justify-center gap-1 transition-colors"
          >
            Show fewer
            <ChevronUp className="w-4 h-4" />
          </button>
        )}

        {/* Local models hint */}
        {!ollamaConnected && (
          <div className="p-4 bg-editor-surface border border-editor-border rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Cpu className="w-4 h-4 text-editor-muted" />
              <span className="text-sm font-medium text-editor-text">Or use local models</span>
            </div>
            <p className="text-xs text-editor-muted">
              Install{' '}
              <a
                href="https://ollama.ai"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                Ollama
              </a>{' '}
              to run models locally without an API key.
            </p>
          </div>
        )}
      </div>

      {error && (
        <p className="mt-4 text-sm text-red-400">{error}</p>
      )}

      {/* Actions */}
      <div className="flex gap-3 mt-8">
        <button
          onClick={onSkip}
          className="px-6 py-2.5 text-editor-muted hover:text-editor-text transition-colors"
        >
          Skip for now
        </button>
        <button
          onClick={onNext}
          disabled={!hasAnyProvider}
          className="px-6 py-2.5 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Continue
        </button>
      </div>
    </div>
  );
}

function ProviderKeyInput({
  provider,
  value,
  onChange,
  onSave,
  isSaved,
  isSaving,
}: {
  provider: ProviderConfig;
  value: string;
  onChange: (value: string) => void;
  onSave: () => void;
  isSaved?: boolean;
  isSaving?: boolean;
}) {
  return (
    <div className="p-4 bg-editor-surface border border-editor-border rounded-lg">
      <div className="flex items-center justify-between mb-2">
        <a
          href={provider.keyUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm font-medium text-primary hover:underline flex items-center gap-1"
        >
          {provider.name}
          <ExternalLink className="w-3 h-3" />
        </a>
        <span className="text-xs text-editor-muted">{provider.description}</span>
      </div>
      <div className="flex gap-2">
        <input
          type="password"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={provider.placeholder}
          disabled={isSaved}
          className="flex-1 px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm disabled:opacity-50"
        />
        {isSaved ? (
          <div className="px-3 py-2 text-green-500 flex items-center gap-1 text-sm">
            <CheckCircle className="w-4 h-4" />
            Saved
          </div>
        ) : (
          <button
            onClick={onSave}
            disabled={!value || isSaving}
            className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSaving ? 'Saving...' : 'Save'}
          </button>
        )}
      </div>
    </div>
  );
}
