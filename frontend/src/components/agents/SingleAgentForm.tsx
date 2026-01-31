import { useState, useEffect } from 'react';
import { X, Bot, Play, Loader2, AlertCircle } from 'lucide-react';
import { useAppStore } from '../../store';
import { AgentConfigSection } from './AgentConfigSection';
import { validateAgentConfig, validateTaskInput } from '../../schemas/agentConfig';
import type {
  SingleAgentFormProps,
  SingleAgentFormState,
  AgentConfig,
} from '../../types/agent';

const DEFAULT_STATE: SingleAgentFormState = {
  provider: '',
  model: '',
  prompt: '',
  temperature: 0.7,
  maxTokens: 4096,
  systemPrompt: '',
  enabledTools: [],
  showAdvanced: false,
};

/**
 * Modal form for configuring and launching a single agent.
 * Provides provider/model selection, task prompt input, and advanced configuration options.
 */
export function SingleAgentForm({
  onSubmit,
  onClose,
  initialValues,
  isSubmitting = false,
  availableTools = [],
}: SingleAgentFormProps) {
  const { providers, selectedProvider, selectedModel, loadProviders } = useAppStore();

  const [formState, setFormState] = useState<SingleAgentFormState>(() => ({
    ...DEFAULT_STATE,
    provider: initialValues?.provider ?? selectedProvider ?? '',
    model: initialValues?.model ?? selectedModel ?? '',
    ...initialValues,
  }));

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Load providers on mount
  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  // Sync with store's selected provider/model if form values are empty
  useEffect(() => {
    if (!formState.provider && selectedProvider) {
      setFormState((prev) => ({ ...prev, provider: selectedProvider }));
    }
    if (!formState.model && selectedModel) {
      setFormState((prev) => ({ ...prev, model: selectedModel }));
    }
  }, [selectedProvider, selectedModel, formState.provider, formState.model]);

  const currentProvider = providers.find((p) => p.name === formState.provider);
  const availableModels = currentProvider?.models ?? [];

  const updateField = <K extends keyof SingleAgentFormState>(
    field: K,
    value: SingleAgentFormState[K]
  ) => {
    setFormState((prev) => ({ ...prev, [field]: value }));
    setTouched((prev) => ({ ...prev, [field]: true }));

    // Clear error when field is updated
    if (errors[field]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[field];
        return next;
      });
    }
  };

  const handleProviderChange = (provider: string) => {
    updateField('provider', provider);
    // Reset model when provider changes
    const newProvider = providers.find((p) => p.name === provider);
    if (newProvider && newProvider.models.length > 0) {
      updateField('model', newProvider.models[0].id);
    } else {
      updateField('model', '');
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formState.provider) {
      newErrors.provider = 'Provider is required';
    }

    if (!formState.model) {
      newErrors.model = 'Model is required';
    }

    const promptValidation = validateTaskInput(formState.prompt);
    if (!promptValidation.valid) {
      newErrors.prompt = promptValidation.message || 'Invalid prompt';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Mark all fields as touched
    setTouched({
      provider: true,
      model: true,
      prompt: true,
    });

    if (!validateForm()) {
      return;
    }

    const config: AgentConfig = {
      provider: formState.provider,
      model: formState.model,
      prompt: formState.prompt.trim(),
      executionConfig: {
        temperature: formState.temperature,
        maxTokens: formState.maxTokens,
        systemPrompt: formState.systemPrompt || undefined,
        enabledTools: formState.enabledTools,
      },
    };

    // Final validation with schema
    const validation = validateAgentConfig(config);
    if (!validation.success) {
      const schemaErrors: Record<string, string> = {};
      validation.errors?.errors.forEach((err) => {
        const path = err.path.join('.');
        schemaErrors[path] = err.message;
      });
      setErrors(schemaErrors);
      return;
    }

    await onSubmit(config);
  };

  const getFieldError = (field: string): string | undefined => {
    return touched[field] ? errors[field] : undefined;
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal */}
      <div
        className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[640px] max-w-[90vw] max-h-[90vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="agent-form-title"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Bot className="w-5 h-5 text-editor-accent" />
            </div>
            <h2 id="agent-form-title" className="text-lg font-semibold text-editor-text">
              Run Agent
            </h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
            aria-label="Close"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="p-6 space-y-5 overflow-y-auto max-h-[calc(90vh-140px)]"
        >
          {/* Provider Selection */}
          <div className="space-y-2">
            <label
              htmlFor="provider-select"
              className="block text-sm font-medium text-editor-text"
            >
              Provider
            </label>
            <select
              id="provider-select"
              value={formState.provider}
              onChange={(e) => handleProviderChange(e.target.value)}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent ${
                getFieldError('provider')
                  ? 'border-red-500'
                  : 'border-editor-border'
              }`}
            >
              <option value="">Select a provider...</option>
              {providers.map((provider) => (
                <option key={provider.name} value={provider.name}>
                  {provider.name}
                  {provider.models.length === 0 && ' (no models)'}
                </option>
              ))}
            </select>
            {getFieldError('provider') && (
              <p className="flex items-center gap-1 text-xs text-red-500">
                <AlertCircle size={12} />
                {getFieldError('provider')}
              </p>
            )}
          </div>

          {/* Model Selection */}
          <div className="space-y-2">
            <label
              htmlFor="model-select"
              className="block text-sm font-medium text-editor-text"
            >
              Model
            </label>
            <select
              id="model-select"
              value={formState.model}
              onChange={(e) => updateField('model', e.target.value)}
              disabled={!formState.provider || availableModels.length === 0}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent disabled:opacity-50 disabled:cursor-not-allowed ${
                getFieldError('model') ? 'border-red-500' : 'border-editor-border'
              }`}
            >
              <option value="">Select a model...</option>
              {availableModels.map((model) => (
                <option key={model.id} value={model.id}>
                  {model.name}
                  {model.supports_tools && ' | Tools'}
                  {model.supports_vision && ' | Vision'}
                </option>
              ))}
            </select>
            {getFieldError('model') && (
              <p className="flex items-center gap-1 text-xs text-red-500">
                <AlertCircle size={12} />
                {getFieldError('model')}
              </p>
            )}
            {formState.provider && availableModels.length === 0 && (
              <p className="text-xs text-editor-muted">
                No models available. Check your API key configuration.
              </p>
            )}
          </div>

          {/* Task Prompt */}
          <div className="space-y-2">
            <label
              htmlFor="prompt-input"
              className="block text-sm font-medium text-editor-text"
            >
              Task Prompt
            </label>
            <textarea
              id="prompt-input"
              value={formState.prompt}
              onChange={(e) => updateField('prompt', e.target.value)}
              placeholder="Describe the task you want the agent to perform..."
              rows={5}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none ${
                getFieldError('prompt') ? 'border-red-500' : 'border-editor-border'
              }`}
            />
            {getFieldError('prompt') && (
              <p className="flex items-center gap-1 text-xs text-red-500">
                <AlertCircle size={12} />
                {getFieldError('prompt')}
              </p>
            )}
          </div>

          {/* Advanced Configuration */}
          <AgentConfigSection
            temperature={formState.temperature}
            onTemperatureChange={(value) => updateField('temperature', value)}
            maxTokens={formState.maxTokens}
            onMaxTokensChange={(value) => updateField('maxTokens', value)}
            systemPrompt={formState.systemPrompt}
            onSystemPromptChange={(value) => updateField('systemPrompt', value)}
            enabledTools={formState.enabledTools}
            onToolsChange={(tools) => updateField('enabledTools', tools)}
            availableTools={availableTools}
            defaultCollapsed={true}
          />
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-editor-border bg-editor-surface/50">
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={isSubmitting || !formState.provider || !formState.model}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSubmitting ? (
              <>
                <Loader2 size={18} className="animate-spin" />
                Running...
              </>
            ) : (
              <>
                <Play size={18} />
                Run Agent
              </>
            )}
          </button>
        </div>
      </div>
    </>
  );
}
