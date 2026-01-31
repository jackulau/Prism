import { useState, useEffect, useCallback } from 'react';
import { X, Wrench, Loader2, Save } from 'lucide-react';
import { apiService } from '../../services/api';
import { SchemaEditor } from './SchemaEditor';

interface Tool {
  id: string;
  display_name: string;
  slug_name: string;
  description?: string;
  is_model: boolean;
  is_builtin: boolean;
  provider_id?: string;
  parameters_schema?: string;
}

interface CustomToolFormProps {
  tool?: Tool | null;
  onClose: () => void;
  onSave: () => void;
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

const DEFAULT_SCHEMA = `{
  "type": "object",
  "properties": {
    "example_param": {
      "type": "string",
      "description": "An example parameter"
    }
  },
  "required": []
}`;

export function CustomToolForm({ tool, onClose, onSave }: CustomToolFormProps) {
  const isEditMode = !!tool;

  const [displayName, setDisplayName] = useState(tool?.display_name || '');
  const [slugName, setSlugName] = useState(tool?.slug_name || '');
  const [description, setDescription] = useState(tool?.description || '');
  const [isModel, setIsModel] = useState(tool?.is_model || false);
  const [providerId, setProviderId] = useState(tool?.provider_id || '');
  const [parametersSchema, setParametersSchema] = useState(tool?.parameters_schema || '');

  const [slugManuallyEdited, setSlugManuallyEdited] = useState(isEditMode);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  useEffect(() => {
    if (!slugManuallyEdited && displayName) {
      setSlugName(slugify(displayName));
    }
  }, [displayName, slugManuallyEdited]);

  const validateForm = useCallback((): boolean => {
    const errors: string[] = [];

    if (!displayName.trim()) {
      errors.push('Display name is required');
    }

    if (!slugName.trim()) {
      errors.push('Slug name is required');
    } else if (!/^[a-z0-9-]+$/.test(slugName)) {
      errors.push('Slug name can only contain lowercase letters, numbers, and hyphens');
    }

    if (parametersSchema.trim()) {
      try {
        JSON.parse(parametersSchema);
      } catch {
        errors.push('Parameters schema must be valid JSON');
      }
    }

    setValidationErrors(errors);
    return errors.length === 0;
  }, [displayName, slugName, parametersSchema]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const toolData = {
        display_name: displayName.trim(),
        slug_name: slugName.trim(),
        description: description.trim() || undefined,
        is_model: isModel,
        provider_id: providerId.trim() || undefined,
        parameters_schema: parametersSchema.trim() || undefined,
      };

      let response;
      if (isEditMode && tool) {
        response = await apiService.updateTool(tool.id, toolData);
      } else {
        response = await apiService.createTool(toolData);
      }

      if (response.error) {
        setError(response.error);
      } else {
        onSave();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save tool');
    } finally {
      setSaving(false);
    }
  };

  const handleSlugChange = (value: string) => {
    setSlugManuallyEdited(true);
    setSlugName(value.toLowerCase().replace(/[^a-z0-9-]/g, ''));
  };

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] max-w-[95vw] max-h-[90vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Wrench className="w-5 h-5 text-editor-accent" />
            </div>
            <h2 className="text-lg font-semibold text-editor-text">
              {isEditMode ? 'Edit Custom Tool' : 'Add Custom Tool'}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-5">
          {/* Error display */}
          {(error || validationErrors.length > 0) && (
            <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-lg">
              {error && <p className="text-sm text-red-400">{error}</p>}
              {validationErrors.length > 0 && (
                <ul className="text-sm text-red-400 space-y-1">
                  {validationErrors.map((err, i) => (
                    <li key={i}>{err}</li>
                  ))}
                </ul>
              )}
            </div>
          )}

          {/* Display Name */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Display Name <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="My Custom Tool"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            />
          </div>

          {/* Slug Name */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">
              Slug Name <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={slugName}
              onChange={(e) => handleSlugChange(e.target.value)}
              placeholder="my-custom-tool"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
            />
            <p className="text-xs text-editor-muted">
              Unique identifier for this tool. Auto-generated from display name.
            </p>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what this tool does..."
              rows={3}
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
            />
          </div>

          {/* Provider ID */}
          <div className="space-y-2">
            <label className="block text-sm font-medium text-editor-text">Provider</label>
            <input
              type="text"
              value={providerId}
              onChange={(e) => setProviderId(e.target.value)}
              placeholder="custom"
              className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
            />
            <p className="text-xs text-editor-muted">
              Optional provider identifier (e.g., "openai", "anthropic", or "custom")
            </p>
          </div>

          {/* Is Model Checkbox */}
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="isModel"
              checked={isModel}
              onChange={(e) => setIsModel(e.target.checked)}
              className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
            />
            <label htmlFor="isModel" className="text-sm text-editor-text">
              This is an LLM model (not a tool)
            </label>
          </div>

          {/* Parameters Schema */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="block text-sm font-medium text-editor-text">
                Parameters Schema (JSON)
              </label>
              {!parametersSchema.trim() && (
                <button
                  type="button"
                  onClick={() => setParametersSchema(DEFAULT_SCHEMA)}
                  className="text-xs text-editor-accent hover:text-editor-accent/80 transition-colors"
                >
                  Use template
                </button>
              )}
            </div>
            <SchemaEditor value={parametersSchema} onChange={setParametersSchema} height="200px" />
            <p className="text-xs text-editor-muted">
              Define the parameters this tool accepts using JSON Schema format.
            </p>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-editor-border bg-editor-surface/50 shrink-0">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={saving}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? (
              <>
                <Loader2 size={18} className="animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save size={18} />
                {isEditMode ? 'Update Tool' : 'Create Tool'}
              </>
            )}
          </button>
        </div>
      </div>
    </>
  );
}

export default CustomToolForm;
