import { X } from 'lucide-react';
import { useAppStore } from '../../store';
import { SettingsPage } from '../../pages/Settings';

export function SettingsPanel() {
  const { isSettingsPanelOpen, toggleSettingsPanel } = useAppStore();

  if (!isSettingsPanelOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40 animate-fade-in"
        onClick={toggleSettingsPanel}
      />

      {/* Panel */}
      <div className="fixed right-0 top-0 bottom-0 w-[600px] max-w-[90vw] bg-editor-bg border-l border-editor-border z-50 overflow-hidden flex flex-col animate-slide-in-right">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border bg-editor-surface/30">
          <h2 className="text-lg font-semibold text-editor-text">Settings</h2>
          <button
            onClick={toggleSettingsPanel}
            className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
            title="Close settings"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content - Reuse existing SettingsPage */}
        <div className="flex-1 overflow-y-auto">
          <SettingsPage />
        </div>
      </div>
    </>
  );
}
