import { Plus, Trash2, Code, FileCode, Terminal, ArrowRight } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';
import type { TransformStepConfig as TransformConfig, TransformType } from '../../../types/workflow';
import { StateVariablePicker } from './StateVariablePicker';

interface TransformStepConfigProps {
  nodeId: string;
}

const TRANSFORM_TYPES: { value: TransformType; label: string; icon: React.ReactNode; description: string }[] = [
  {
    value: 'template',
    label: 'Template',
    icon: <FileCode size={16} />,
    description: 'Use Go template syntax to transform data',
  },
  {
    value: 'jq',
    label: 'JQ Expression',
    icon: <Code size={16} />,
    description: 'Use jq-like expressions for JSON manipulation',
  },
  {
    value: 'script',
    label: 'Script',
    icon: <Terminal size={16} />,
    description: 'Write custom transformation logic',
  },
];

export function TransformStepConfig({ nodeId }: TransformStepConfigProps) {
  const { getSelectedNode, updateNodeConfig } = useWorkflowStore();

  const node = getSelectedNode();
  const config = node?.data.config.transformConfig;

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<TransformConfig>) => {
    updateNodeConfig(nodeId, {
      transformConfig: { ...config, ...updates },
    });
  };

  const handleMappingChange = (key: string, value: string) => {
    updateConfig({
      mapping: { ...config.mapping, [key]: value },
    });
  };

  const handleAddMapping = () => {
    const newKey = `key_${Object.keys(config.mapping || {}).length + 1}`;
    updateConfig({
      mapping: { ...config.mapping, [newKey]: '' },
    });
  };

  const handleRemoveMapping = (key: string) => {
    const { [key]: _, ...rest } = config.mapping || {};
    updateConfig({ mapping: rest });
  };

  const handleRenameMappingKey = (oldKey: string, newKey: string) => {
    if (oldKey === newKey || !newKey.trim()) return;
    const entries = Object.entries(config.mapping || {});
    const newMapping: Record<string, string> = {};
    for (const [k, v] of entries) {
      newMapping[k === oldKey ? newKey : k] = v;
    }
    updateConfig({ mapping: newMapping });
  };

  const handleInsertVariable = (variable: string) => {
    // Insert into the current template/script field
    if (config.type === 'template') {
      updateConfig({ template: (config.template || '') + variable });
    } else if (config.type === 'script') {
      updateConfig({ script: (config.script || '') + variable });
    }
  };

  return (
    <div className="space-y-4">
      {/* Transform Type Selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Transform Type
        </label>
        <div className="grid grid-cols-1 gap-2">
          {TRANSFORM_TYPES.map((type) => (
            <button
              key={type.value}
              type="button"
              onClick={() => updateConfig({ type: type.value })}
              className={`flex items-start gap-3 p-3 rounded-lg border transition-colors text-left ${
                config.type === type.value
                  ? 'border-editor-accent bg-editor-accent/10'
                  : 'border-editor-border bg-editor-surface/50 hover:border-editor-accent/50'
              }`}
            >
              <div
                className={`p-2 rounded-lg ${
                  config.type === type.value
                    ? 'bg-editor-accent text-white'
                    : 'bg-editor-surface text-editor-muted'
                }`}
              >
                {type.icon}
              </div>
              <div>
                <div className="text-sm font-medium text-editor-text">{type.label}</div>
                <div className="text-xs text-editor-muted">{type.description}</div>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Input Key */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Input Key
          <span className="text-editor-muted font-normal ml-1">(optional)</span>
        </label>
        <input
          type="text"
          value={config.inputKey || ''}
          onChange={(e) => updateConfig({ inputKey: e.target.value })}
          placeholder="sourceData"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
        <p className="text-xs text-editor-muted">
          State key to use as input for the transformation. Leave empty to use entire state.
        </p>
      </div>

      {/* Type-specific Configuration */}
      {config.type === 'template' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            Template
          </label>
          <textarea
            value={config.template || ''}
            onChange={(e) => updateConfig({ template: e.target.value })}
            placeholder={'Hello {{.name}}, your order #{{.orderId}} is ready!'}
            rows={6}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm resize-none"
          />
          <StateVariablePicker nodeId={nodeId} onInsert={handleInsertVariable} />
          <div className="p-2 bg-editor-surface/50 rounded-lg text-xs text-editor-muted">
            <strong>Template Syntax:</strong>
            <ul className="mt-1 space-y-0.5 list-disc list-inside">
              <li><code className="bg-editor-surface px-1 rounded">{'{{.field}}'}</code> - Access a field</li>
              <li><code className="bg-editor-surface px-1 rounded">{'{{if .condition}}...{{end}}'}</code> - Conditional</li>
              <li><code className="bg-editor-surface px-1 rounded">{'{{range .items}}...{{end}}'}</code> - Loop</li>
            </ul>
          </div>
        </div>
      )}

      {config.type === 'jq' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            JQ Expression
          </label>
          <input
            type="text"
            value={config.template || ''}
            onChange={(e) => updateConfig({ template: e.target.value })}
            placeholder={'.data | map(select(.status == "active"))'}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
          />
          <div className="p-2 bg-editor-surface/50 rounded-lg text-xs text-editor-muted">
            <strong>JQ Examples:</strong>
            <ul className="mt-1 space-y-0.5 list-disc list-inside">
              <li><code className="bg-editor-surface px-1 rounded">.field</code> - Get a field</li>
              <li><code className="bg-editor-surface px-1 rounded">.[] | .name</code> - Get all names from array</li>
              <li><code className="bg-editor-surface px-1 rounded">{'select(.age > 18)'}</code> - Filter</li>
            </ul>
          </div>
        </div>
      )}

      {config.type === 'script' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            Script
          </label>
          <textarea
            value={config.script || ''}
            onChange={(e) => updateConfig({ script: e.target.value })}
            placeholder={'// JavaScript-like syntax\nconst result = input.items.filter(i => i.active);\nreturn { processed: result };'}
            rows={8}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm resize-none"
          />
          <StateVariablePicker nodeId={nodeId} onInsert={handleInsertVariable} />
          <p className="text-xs text-editor-muted">
            Write custom transformation logic. <code>input</code> contains the input data, return the transformed result.
          </p>
        </div>
      )}

      {/* Simple Key Mapping */}
      <div className="space-y-2 pt-2">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-editor-text">
            Key Mapping
            <span className="text-editor-muted font-normal ml-1">(optional)</span>
          </label>
        </div>
        <p className="text-xs text-editor-muted mb-2">
          Rename or remap fields from input to output.
        </p>

        <div className="space-y-2">
          {Object.entries(config.mapping || {}).map(([key, value]) => (
            <div key={key} className="flex items-center gap-2">
              <input
                type="text"
                value={key}
                onChange={(e) => handleRenameMappingKey(key, e.target.value)}
                placeholder="outputKey"
                className="w-28 px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-xs"
              />
              <ArrowRight size={14} className="text-editor-muted flex-shrink-0" />
              <input
                type="text"
                value={value}
                onChange={(e) => handleMappingChange(key, e.target.value)}
                placeholder="inputKey or {{state.var}}"
                className="flex-1 px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-xs"
              />
              <button
                type="button"
                onClick={() => handleRemoveMapping(key)}
                className="p-1.5 text-editor-muted hover:text-editor-error transition-colors"
              >
                <Trash2 size={14} />
              </button>
            </div>
          ))}

          <button
            type="button"
            onClick={handleAddMapping}
            className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors w-full border border-dashed border-editor-border"
          >
            <Plus size={14} />
            Add Mapping
          </button>
        </div>
      </div>

      {/* Output Key */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Output Key
          <span className="text-editor-muted font-normal ml-1">(optional)</span>
        </label>
        <input
          type="text"
          value={config.outputKey || ''}
          onChange={(e) => updateConfig({ outputKey: e.target.value })}
          placeholder="transformedData"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
        <p className="text-xs text-editor-muted">
          Store the transformation result in workflow state with this key
        </p>
      </div>
    </div>
  );
}
