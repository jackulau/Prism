import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Server, Check, Cpu, Zap, ExternalLink } from 'lucide-react';
import { useAppStore } from '../store';

// Provider metadata for tooltips and help info
const providerInfo: Record<string, {
  description: string;
  keyUrl: string;
  features: string[];
  icon?: 'local' | 'fast' | 'default';
}> = {
  openai: {
    description: 'OpenAI GPT models',
    keyUrl: 'https://platform.openai.com/api-keys',
    features: ['Tools', 'Vision', 'Reasoning'],
  },
  anthropic: {
    description: 'Claude AI models',
    keyUrl: 'https://console.anthropic.com/settings/keys',
    features: ['Tools', 'Vision', '200K context'],
  },
  google: {
    description: 'Google Gemini models',
    keyUrl: 'https://aistudio.google.com/app/apikey',
    features: ['Tools', 'Vision', '1M context'],
  },
  openrouter: {
    description: '200+ models via single API',
    keyUrl: 'https://openrouter.ai/keys',
    features: ['Multi-provider', 'Pay-per-use'],
  },
  groq: {
    description: 'Ultra-fast inference',
    keyUrl: 'https://console.groq.com/keys',
    features: ['Fast', 'Free tier', 'Llama/Mixtral'],
    icon: 'fast',
  },
  deepseek: {
    description: 'Cost-effective reasoning',
    keyUrl: 'https://platform.deepseek.com/api_keys',
    features: ['Cheap', 'Coding', 'Reasoning'],
  },
  ollama: {
    description: 'Local models - no API key',
    keyUrl: 'https://ollama.ai',
    features: ['Private', 'Free', 'Offline'],
    icon: 'local',
  },
};

export function ModelSelector() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const {
    providers,
    selectedProvider,
    selectedModel,
    setSelectedProvider,
    setSelectedModel,
    loadProviders,
  } = useAppStore();

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (providerName: string, modelId: string) => {
    setSelectedProvider(providerName);
    setSelectedModel(modelId);
    setIsOpen(false);
  };

  const getProviderIcon = (name: string) => {
    const info = providerInfo[name];
    if (info?.icon === 'local') return <Cpu size={12} className="text-editor-accent" />;
    if (info?.icon === 'fast') return <Zap size={12} className="text-yellow-500" />;
    return <Server size={12} />;
  };

  const currentProvider = providers.find(p => p.name === selectedProvider);
  const currentModel = currentProvider?.models.find(m => m.id === selectedModel);
  const displayName = currentModel?.name || selectedModel || 'Select Model';
  const isOllama = selectedProvider === 'ollama';
  const isGroq = selectedProvider === 'groq';

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center gap-2 px-3 py-2 rounded-lg bg-editor-surface border border-editor-border hover:border-editor-accent transition-colors text-sm"
      >
        {isOllama ? (
          <Cpu size={14} className="text-editor-accent flex-shrink-0" />
        ) : isGroq ? (
          <Zap size={14} className="text-yellow-500 flex-shrink-0" />
        ) : (
          <Server size={14} className="text-editor-muted flex-shrink-0" />
        )}
        <span className="text-editor-text truncate flex-1 text-left">{displayName}</span>
        <ChevronDown size={14} className={`text-editor-muted transition-transform flex-shrink-0 ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 max-h-80 overflow-y-auto">
          {providers.length === 0 ? (
            <div className="px-3 py-4 text-center text-sm text-editor-muted">
              Loading providers...
            </div>
          ) : (
            providers.map((provider) => {
              const info = providerInfo[provider.name];
              return (
              <div key={provider.name}>
                <div className="px-3 py-2 text-xs font-semibold text-editor-muted uppercase bg-editor-surface/50 border-b border-editor-border sticky top-0 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {getProviderIcon(provider.name)}
                    <span>{provider.name}</span>
                    {info?.icon === 'local' && (
                      <span className="text-editor-accent font-normal normal-case">(Local)</span>
                    )}
                    {info?.icon === 'fast' && (
                      <span className="text-yellow-500 font-normal normal-case">(Fast)</span>
                    )}
                  </div>
                  {info?.keyUrl && provider.name !== 'ollama' && provider.models.length === 0 && (
                    <a
                      href={info.keyUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-1 text-editor-accent font-normal normal-case hover:underline"
                      onClick={(e) => e.stopPropagation()}
                    >
                      Get API key
                      <ExternalLink size={10} />
                    </a>
                  )}
                </div>
                {provider.models.length > 0 ? (
                  provider.models.map((model) => (
                    <button
                      key={`${provider.name}-${model.id}`}
                      onClick={() => handleSelect(provider.name, model.id)}
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
                      {selectedProvider === provider.name && selectedModel === model.id && (
                        <Check size={14} className="text-editor-accent flex-shrink-0 ml-2" />
                      )}
                    </button>
                  ))
                ) : (
                  <div className="px-3 py-2 text-sm text-editor-muted">
                    {provider.name === 'ollama' ? (
                      <span className="italic">No models - run "ollama pull llama3.2"</span>
                    ) : info ? (
                      <div className="space-y-1">
                        <span className="italic block">No API key configured</span>
                        <span className="text-xs block">{info.description}</span>
                        {info.features.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {info.features.map((f) => (
                              <span
                                key={f}
                                className="px-1.5 py-0.5 text-xs bg-editor-surface rounded border border-editor-border"
                              >
                                {f}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    ) : (
                      <span className="italic">No models available</span>
                    )}
                  </div>
                )}
              </div>
            );
            })
          )}
        </div>
      )}
    </div>
  );
}
