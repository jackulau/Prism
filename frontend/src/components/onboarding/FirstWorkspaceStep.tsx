import { useState } from 'react';
import { Folder, Sparkles, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAppStore } from '../../store';

interface FirstWorkspaceStepProps {
  onNext: () => void;
  onSkip: () => void;
}

export function FirstWorkspaceStep({ onNext, onSkip }: FirstWorkspaceStepProps) {
  const navigate = useNavigate();
  const { createNewConversation, providers, selectedProvider, selectedModel, setSelectedProvider, setSelectedModel } = useAppStore();
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreateWorkspace = async () => {
    setIsCreating(true);
    setError(null);

    try {
      const conversationId = await createNewConversation();
      if (conversationId) {
        // Navigate to the workspace and complete onboarding
        onNext();
        navigate(`/workspace/${conversationId}`);
      } else {
        setError('Failed to create workspace. Please try again.');
      }
    } catch {
      setError('An error occurred. Please try again.');
    } finally {
      setIsCreating(false);
    }
  };

  // Get available models from selected provider
  const currentProvider = providers.find((p) => p.name === selectedProvider);
  const availableModels = currentProvider?.models || [];

  return (
    <div className="flex flex-col items-center min-h-[400px] px-6 py-8 animate-fade-in">
      {/* Header */}
      <div className="w-14 h-14 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-6">
        <Folder className="w-7 h-7" />
      </div>

      <h2 className="text-2xl font-bold text-editor-text mb-2 text-center">
        Create Your First Workspace
      </h2>
      <p className="text-editor-muted text-center max-w-md mb-8">
        A workspace is where you&apos;ll have conversations with AI. Let&apos;s create one now.
      </p>

      {/* Model selection */}
      <div className="w-full max-w-md space-y-4 mb-8">
        {/* Provider selector */}
        {providers.length > 1 && (
          <div>
            <label className="block text-sm font-medium text-editor-text mb-2">
              Select Provider
            </label>
            <select
              value={selectedProvider}
              onChange={(e) => {
                setSelectedProvider(e.target.value);
                const provider = providers.find((p) => p.name === e.target.value);
                if (provider?.models?.[0]) {
                  setSelectedModel(provider.models[0].id);
                }
              }}
              className="w-full px-4 py-3 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:ring-2 focus:ring-primary/50"
            >
              {providers.map((provider) => (
                <option key={provider.name} value={provider.name}>
                  {provider.name}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Model selector */}
        {availableModels.length > 0 && (
          <div>
            <label className="block text-sm font-medium text-editor-text mb-2">
              Select Model
            </label>
            <select
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="w-full px-4 py-3 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:ring-2 focus:ring-primary/50"
            >
              {availableModels.map((model) => (
                <option key={model.id} value={model.id}>
                  {model.name || model.id}
                </option>
              ))}
            </select>
          </div>
        )}

        {providers.length === 0 && (
          <div className="p-4 bg-editor-surface border border-editor-border rounded-lg text-center">
            <p className="text-sm text-editor-muted">
              No providers configured. You can skip this step and configure providers later.
            </p>
          </div>
        )}
      </div>

      {/* Quick start suggestions */}
      <div className="w-full max-w-md mb-8">
        <p className="text-sm text-editor-muted mb-3">Quick start ideas:</p>
        <div className="space-y-2">
          <QuickStartIdea
            text="Help me debug a React component"
            icon={<Sparkles className="w-4 h-4" />}
          />
          <QuickStartIdea
            text="Write a Python function to parse JSON"
            icon={<Sparkles className="w-4 h-4" />}
          />
          <QuickStartIdea
            text="Explain this code to me"
            icon={<Sparkles className="w-4 h-4" />}
          />
        </div>
      </div>

      {error && (
        <p className="mb-4 text-sm text-red-400">{error}</p>
      )}

      {/* Actions */}
      <div className="flex gap-3">
        <button
          onClick={onSkip}
          className="px-6 py-2.5 text-editor-muted hover:text-editor-text transition-colors"
        >
          Skip for now
        </button>
        <button
          onClick={handleCreateWorkspace}
          disabled={isCreating || providers.length === 0}
          className="px-6 py-2.5 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {isCreating ? (
            <>
              <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              Creating...
            </>
          ) : (
            <>
              Create Workspace
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </button>
      </div>
    </div>
  );
}

function QuickStartIdea({ text, icon }: { text: string; icon: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3 px-4 py-3 bg-editor-surface border border-editor-border rounded-lg">
      <span className="text-primary">{icon}</span>
      <span className="text-sm text-editor-text">{text}</span>
    </div>
  );
}
