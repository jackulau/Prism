import { useState, useEffect, useCallback } from 'react';
import { X, AlertTriangle } from 'lucide-react';

interface DataConfigModalProps {
  isOpen: boolean;
  mode: 'create' | 'edit';
  initialType?: string;
  initialKey?: string;
  initialValue?: Record<string, unknown>;
  onSave: (configType: string, configKey: string, value: Record<string, unknown>) => Promise<boolean>;
  onClose: () => void;
  saving?: boolean;
}

export function DataConfigModal({
  isOpen,
  mode,
  initialType = '',
  initialKey = '',
  initialValue = {},
  onSave,
  onClose,
  saving = false,
}: DataConfigModalProps) {
  const [configType, setConfigType] = useState(initialType);
  const [configKey, setConfigKey] = useState(initialKey);
  const [jsonValue, setJsonValue] = useState('');
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [touched, setTouched] = useState(false);

  // Reset form when modal opens
  useEffect(() => {
    if (isOpen) {
      setConfigType(initialType);
      setConfigKey(initialKey);
      setJsonValue(
        initialValue && Object.keys(initialValue).length > 0
          ? JSON.stringify(initialValue, null, 2)
          : '{\n  \n}'
      );
      setJsonError(null);
      setTouched(false);
    }
  }, [isOpen, initialType, initialKey, initialValue]);

  // Validate JSON on change
  const validateJson = useCallback((value: string): boolean => {
    if (!value.trim()) {
      setJsonError('JSON value is required');
      return false;
    }
    try {
      JSON.parse(value);
      setJsonError(null);
      return true;
    } catch (e) {
      setJsonError(e instanceof Error ? e.message : 'Invalid JSON');
      return false;
    }
  }, []);

  const handleJsonChange = (value: string) => {
    setJsonValue(value);
    setTouched(true);
    validateJson(value);
  };

  const handleSubmit = async () => {
    if (!configType.trim() || !configKey.trim()) {
      return;
    }
    if (!validateJson(jsonValue)) {
      return;
    }

    try {
      const parsedValue = JSON.parse(jsonValue);
      const success = await onSave(configType.trim(), configKey.trim(), parsedValue);
      if (success) {
        onClose();
      }
    } catch {
      setJsonError('Failed to parse JSON');
    }
  };

  const canSubmit =
    configType.trim() !== '' &&
    configKey.trim() !== '' &&
    !jsonError &&
    !saving &&
    (touched || mode === 'edit');

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40 animate-fade-in"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] max-w-[90vw] max-h-[80vh] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-editor-border">
          <h3 className="text-lg font-semibold">
            {mode === 'create' ? 'Create Configuration' : 'Edit Configuration'}
          </h3>
          <button
            onClick={onClose}
            className="p-1 hover:bg-editor-surface rounded transition-colors"
          >
            <X className="w-5 h-5 text-editor-muted" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4 overflow-y-auto flex-1">
          {/* Config Type */}
          <div>
            <label className="block text-sm font-medium mb-2">Config Type</label>
            <input
              type="text"
              value={configType}
              onChange={(e) => setConfigType(e.target.value)}
              placeholder="e.g., credentials, settings, api-keys"
              disabled={mode === 'edit'}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <p className="text-xs text-editor-muted mt-1">
              Category for grouping related configurations
            </p>
          </div>

          {/* Config Key */}
          <div>
            <label className="block text-sm font-medium mb-2">Config Key</label>
            <input
              type="text"
              value={configKey}
              onChange={(e) => setConfigKey(e.target.value)}
              placeholder="e.g., stripe, github, default"
              disabled={mode === 'edit'}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <p className="text-xs text-editor-muted mt-1">
              Unique identifier within the config type
            </p>
          </div>

          {/* JSON Value */}
          <div>
            <label className="block text-sm font-medium mb-2">Value (JSON)</label>
            <textarea
              value={jsonValue}
              onChange={(e) => handleJsonChange(e.target.value)}
              placeholder='{\n  "key": "value"\n}'
              rows={10}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-sm font-mono resize-y ${
                jsonError && touched
                  ? 'border-red-500'
                  : 'border-editor-border'
              }`}
              spellCheck={false}
            />
            {jsonError && touched && (
              <div className="flex items-center gap-2 mt-2 text-xs text-red-400">
                <AlertTriangle className="w-3 h-3" />
                <span>{jsonError}</span>
              </div>
            )}
            <p className="text-xs text-editor-muted mt-1">
              Configuration value stored as encrypted JSON
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 p-4 border-t border-editor-border">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? 'Saving...' : mode === 'create' ? 'Create' : 'Save'}
          </button>
        </div>
      </div>
    </>
  );
}
