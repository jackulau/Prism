import { useState } from 'react';
import { Plus, Trash2, Code, AlertCircle } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';
import type { ToolStepConfig as ToolConfig } from '../../../types/workflow';
import { StateVariableInput } from './StateVariablePicker';

interface ToolStepConfigProps {
  nodeId: string;
}

// Common tools available in the system
const COMMON_TOOLS = [
  { name: 'read_file', description: 'Read contents of a file' },
  { name: 'write_file', description: 'Write contents to a file' },
  { name: 'list_files', description: 'List files in a directory' },
  { name: 'search_files', description: 'Search for files matching a pattern' },
  { name: 'execute_command', description: 'Execute a shell command' },
  { name: 'http_request', description: 'Make an HTTP request' },
  { name: 'send_email', description: 'Send an email' },
  { name: 'create_github_issue', description: 'Create a GitHub issue' },
  { name: 'query_database', description: 'Query a database' },
];

export function ToolStepConfig({ nodeId }: ToolStepConfigProps) {
  const { getSelectedNode, updateNodeConfig } = useWorkflowStore();
  const [showJsonEditor, setShowJsonEditor] = useState(false);
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [jsonValue, setJsonValue] = useState('');

  const node = getSelectedNode();
  const config = node?.data.config.toolConfig;

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<ToolConfig>) => {
    updateNodeConfig(nodeId, {
      toolConfig: { ...config, ...updates },
    });
  };

  const handleParameterChange = (key: string, value: string) => {
    updateConfig({
      parameters: { ...config.parameters, [key]: value },
    });
  };

  const handleAddParameter = () => {
    const newKey = `param_${Object.keys(config.parameters).length + 1}`;
    updateConfig({
      parameters: { ...config.parameters, [newKey]: '' },
    });
  };

  const handleRemoveParameter = (key: string) => {
    const { [key]: _, ...rest } = config.parameters;
    updateConfig({ parameters: rest });
  };

  const handleRenameParameter = (oldKey: string, newKey: string) => {
    if (oldKey === newKey || !newKey.trim()) return;
    const entries = Object.entries(config.parameters);
    const newParams: Record<string, unknown> = {};
    for (const [k, v] of entries) {
      newParams[k === oldKey ? newKey : k] = v;
    }
    updateConfig({ parameters: newParams });
  };

  const handleJsonEdit = (json: string) => {
    setJsonValue(json);
    try {
      const parsed = JSON.parse(json);
      if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
        updateConfig({ parameters: parsed });
        setJsonError(null);
      } else {
        setJsonError('Parameters must be a JSON object');
      }
    } catch {
      setJsonError('Invalid JSON');
    }
  };

  const toggleJsonEditor = () => {
    if (!showJsonEditor) {
      setJsonValue(JSON.stringify(config.parameters, null, 2));
      setJsonError(null);
    }
    setShowJsonEditor(!showJsonEditor);
  };

  return (
    <div className="space-y-4">
      {/* Tool Name Selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Tool Name
        </label>
        <select
          value={config.toolName}
          onChange={(e) => updateConfig({ toolName: e.target.value })}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent"
        >
          <option value="">Select a tool...</option>
          {COMMON_TOOLS.map((tool) => (
            <option key={tool.name} value={tool.name}>
              {tool.name}
            </option>
          ))}
        </select>
        {config.toolName && (
          <p className="text-xs text-editor-muted">
            {COMMON_TOOLS.find((t) => t.name === config.toolName)?.description}
          </p>
        )}
      </div>

      {/* Custom Tool Name */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Or enter custom tool name
        </label>
        <input
          type="text"
          value={config.toolName}
          onChange={(e) => updateConfig({ toolName: e.target.value })}
          placeholder="custom_tool_name"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
      </div>

      {/* Parameters */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-editor-text">
            Parameters
          </label>
          <button
            type="button"
            onClick={toggleJsonEditor}
            className="flex items-center gap-1 px-2 py-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
          >
            <Code size={12} />
            {showJsonEditor ? 'Form View' : 'JSON View'}
          </button>
        </div>

        {showJsonEditor ? (
          <div className="space-y-2">
            <textarea
              value={jsonValue}
              onChange={(e) => handleJsonEdit(e.target.value)}
              placeholder='{"key": "value"}'
              rows={8}
              className={`w-full px-3 py-2 bg-editor-surface border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none font-mono text-sm resize-none ${
                jsonError ? 'border-editor-error' : 'border-editor-border focus:border-editor-accent'
              }`}
            />
            {jsonError && (
              <div className="flex items-center gap-2 text-sm text-editor-error">
                <AlertCircle size={14} />
                {jsonError}
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {Object.entries(config.parameters).map(([key, value]) => (
              <div key={key} className="flex items-start gap-2">
                <input
                  type="text"
                  value={key}
                  onChange={(e) => handleRenameParameter(key, e.target.value)}
                  placeholder="key"
                  className="w-32 px-2 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
                />
                <div className="flex-1">
                  <StateVariableInput
                    value={String(value)}
                    onChange={(v) => handleParameterChange(key, v)}
                    nodeId={nodeId}
                    placeholder="value or {{state.var}}"
                  />
                </div>
                <button
                  type="button"
                  onClick={() => handleRemoveParameter(key)}
                  className="p-2 text-editor-muted hover:text-editor-error transition-colors"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}

            <button
              type="button"
              onClick={handleAddParameter}
              className="flex items-center gap-2 px-3 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors w-full border border-dashed border-editor-border"
            >
              <Plus size={14} />
              Add Parameter
            </button>
          </div>
        )}
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
          placeholder="toolResult"
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
        <p className="text-xs text-editor-muted">
          Store the tool's output in workflow state with this key
        </p>
      </div>
    </div>
  );
}
