import { Brain } from 'lucide-react';
import { useAppStore } from '../../store';

export function ThinkingToggle() {
  const { extendedThinkingEnabled, setExtendedThinkingEnabled } = useAppStore();

  return (
    <button
      onClick={() => setExtendedThinkingEnabled(!extendedThinkingEnabled)}
      className={`flex items-center gap-1.5 px-2 py-1 rounded-md border transition-colors text-xs ${
        extendedThinkingEnabled
          ? 'bg-editor-accent/20 border-editor-accent text-editor-accent'
          : 'bg-editor-surface border-editor-border text-editor-muted hover:text-editor-text hover:border-editor-accent'
      }`}
      title={extendedThinkingEnabled ? 'Disable extended thinking' : 'Enable extended thinking'}
    >
      <Brain size={12} />
      <span>Thinking</span>
      <span
        className={`w-1.5 h-1.5 rounded-full ${
          extendedThinkingEnabled ? 'bg-editor-accent' : 'bg-editor-muted'
        }`}
      />
    </button>
  );
}
