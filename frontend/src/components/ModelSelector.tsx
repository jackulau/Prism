import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Server, Check, Cpu } from 'lucide-react';
import { useAppStore } from '../store';

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

  const currentProvider = providers.find(p => p.name === selectedProvider);
  const currentModel = currentProvider?.models.find(m => m.id === selectedModel);
  const displayName = currentModel?.name || selectedModel || 'Select Model';
  const isOllama = selectedProvider === 'ollama';

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center gap-2 px-3 py-2 rounded-lg bg-editor-surface border border-editor-border hover:border-editor-accent transition-colors text-sm"
      >
        {isOllama ? (
          <Cpu size={14} className="text-editor-accent flex-shrink-0" />
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
            providers.map((provider) => (
              <div key={provider.name}>
                <div className="px-3 py-2 text-xs font-semibold text-editor-muted uppercase bg-editor-surface/50 border-b border-editor-border sticky top-0 flex items-center gap-2">
                  {provider.name === 'ollama' ? (
                    <>
                      <Cpu size={12} className="text-editor-accent" />
                      <span>{provider.name}</span>
                      <span className="text-editor-accent font-normal normal-case">(Local)</span>
                    </>
                  ) : (
                    <>
                      <Server size={12} />
                      <span>{provider.name}</span>
                    </>
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
                  <div className="px-3 py-2 text-sm text-editor-muted italic">
                    {provider.name === 'ollama'
                      ? 'No models - run "ollama pull llama2"'
                      : 'No models available'}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
