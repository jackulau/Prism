import { Clock, MessageSquare, Webhook, Timer } from 'lucide-react';
import { useWorkflowStore } from '../../../store/workflowStore';
import type { WaitStepConfig as WaitConfig, WaitType } from '../../../types/workflow';

interface WaitStepConfigProps {
  nodeId: string;
}

const WAIT_TYPES: { value: WaitType; label: string; icon: React.ReactNode; description: string }[] = [
  {
    value: 'user_input',
    label: 'User Input',
    icon: <MessageSquare size={16} />,
    description: 'Wait for user to provide input',
  },
  {
    value: 'webhook',
    label: 'Webhook',
    icon: <Webhook size={16} />,
    description: 'Wait for an external webhook call',
  },
  {
    value: 'timeout',
    label: 'Timeout',
    icon: <Timer size={16} />,
    description: 'Wait for a specified duration',
  },
];

// Duration presets in milliseconds
const DURATION_PRESETS = [
  { label: '30 seconds', value: 30000 },
  { label: '1 minute', value: 60000 },
  { label: '5 minutes', value: 300000 },
  { label: '15 minutes', value: 900000 },
  { label: '30 minutes', value: 1800000 },
  { label: '1 hour', value: 3600000 },
];

export function WaitStepConfig({ nodeId }: WaitStepConfigProps) {
  const { getSelectedNode, updateNodeConfig } = useWorkflowStore();

  const node = getSelectedNode();
  const config = node?.data.config.waitConfig;

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<WaitConfig>) => {
    updateNodeConfig(nodeId, {
      waitConfig: { ...config, ...updates },
    });
  };

  // Format milliseconds to readable string
  const formatDuration = (ms: number): string => {
    if (ms < 60000) return `${Math.round(ms / 1000)} seconds`;
    if (ms < 3600000) return `${Math.round(ms / 60000)} minutes`;
    return `${Math.round(ms / 3600000)} hours`;
  };

  // Parse duration string to milliseconds
  const parseDuration = (value: string): number => {
    const num = parseInt(value);
    if (isNaN(num)) return 300000; // Default 5 minutes
    return Math.max(1000, Math.min(86400000, num)); // Clamp between 1 second and 24 hours
  };

  return (
    <div className="space-y-4">
      {/* Wait Type Selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          Wait Type
        </label>
        <div className="grid grid-cols-1 gap-2">
          {WAIT_TYPES.map((type) => (
            <button
              key={type.value}
              type="button"
              onClick={() => updateConfig({ waitType: type.value })}
              className={`flex items-start gap-3 p-3 rounded-lg border transition-colors text-left ${
                config.waitType === type.value
                  ? 'border-editor-accent bg-editor-accent/10'
                  : 'border-editor-border bg-editor-surface/50 hover:border-editor-accent/50'
              }`}
            >
              <div
                className={`p-2 rounded-lg ${
                  config.waitType === type.value
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

      {/* Conditional Fields Based on Wait Type */}
      {config.waitType === 'user_input' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            Prompt Text
          </label>
          <textarea
            value={config.promptText || ''}
            onChange={(e) => updateConfig({ promptText: e.target.value })}
            placeholder="Enter the message to display to the user..."
            rows={3}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
          />
          <p className="text-xs text-editor-muted">
            This message will be shown to the user when waiting for their input.
          </p>
        </div>
      )}

      {config.waitType === 'webhook' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            Webhook Path
          </label>
          <div className="flex items-center gap-2">
            <span className="text-sm text-editor-muted">/api/v1/workflows/webhook/</span>
            <input
              type="text"
              value={config.webhookPath || ''}
              onChange={(e) => updateConfig({ webhookPath: e.target.value })}
              placeholder="custom-path"
              className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
            />
          </div>
          <p className="text-xs text-editor-muted">
            External services can POST to this endpoint to resume the workflow.
          </p>
        </div>
      )}

      {config.waitType === 'timeout' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-editor-text">
            Wait Duration
          </label>
          <div className="flex flex-wrap gap-2">
            {DURATION_PRESETS.map((preset) => (
              <button
                key={preset.value}
                type="button"
                onClick={() => updateConfig({ timeout: preset.value })}
                className={`px-3 py-1.5 text-xs rounded-lg border transition-colors ${
                  config.timeout === preset.value
                    ? 'border-editor-accent bg-editor-accent/10 text-editor-accent'
                    : 'border-editor-border bg-editor-surface text-editor-text hover:border-editor-accent/50'
                }`}
              >
                {preset.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Timeout (for all wait types) */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-editor-text">
          <div className="flex items-center gap-2">
            <Clock size={14} className="text-editor-muted" />
            {config.waitType === 'timeout' ? 'Duration (ms)' : 'Maximum Wait Time (ms)'}
          </div>
        </label>
        <div className="flex items-center gap-2">
          <input
            type="number"
            value={config.timeout || 300000}
            onChange={(e) => updateConfig({ timeout: parseDuration(e.target.value) })}
            min={1000}
            max={86400000}
            step={1000}
            className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text focus:outline-none focus:border-editor-accent text-sm"
          />
          <span className="text-sm text-editor-muted whitespace-nowrap">
            = {formatDuration(config.timeout || 300000)}
          </span>
        </div>
        {config.waitType !== 'timeout' && (
          <p className="text-xs text-editor-muted">
            If the wait exceeds this time, the step will fail with a timeout error.
          </p>
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
          placeholder={config.waitType === 'user_input' ? 'userResponse' : 'waitResult'}
          className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent font-mono text-sm"
        />
        <p className="text-xs text-editor-muted">
          {config.waitType === 'user_input'
            ? 'Store the user\'s response in workflow state with this key'
            : config.waitType === 'webhook'
            ? 'Store the webhook payload in workflow state with this key'
            : 'Store a timestamp when the timeout completes'}
        </p>
      </div>
    </div>
  );
}
