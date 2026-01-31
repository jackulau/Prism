import { Trash2, GripVertical, Bot, Wrench, GitBranch, Clock, Shuffle, ChevronDown, ChevronUp } from 'lucide-react';
import { useWorkflowStore, getStepTypeLabel } from '../../../store/workflowStore';
import type { ParallelStepConfig as ParallelConfig, WorkflowStep, StepType } from '../../../types/workflow';
import { useState } from 'react';

interface ParallelStepConfigProps {
  nodeId: string;
}

// Icons for each step type
const stepTypeIcons: Record<StepType, React.ReactNode> = {
  agent: <Bot size={14} />,
  tool: <Wrench size={14} />,
  condition: <GitBranch size={14} />,
  parallel: <Bot size={14} />,
  wait: <Clock size={14} />,
  transform: <Shuffle size={14} />,
};

// Available step types for nested steps (excluding parallel to prevent infinite nesting)
const AVAILABLE_STEP_TYPES: StepType[] = ['agent', 'tool', 'wait', 'transform'];

export function ParallelStepConfig({ nodeId }: ParallelStepConfigProps) {
  const { getSelectedNode, updateNodeConfig } = useWorkflowStore();
  const [expandedStep, setExpandedStep] = useState<number | null>(null);

  const node = getSelectedNode();
  const config = node?.data.config.parallelConfig;

  if (!node || !config) return null;

  const updateConfig = (updates: Partial<ParallelConfig>) => {
    updateNodeConfig(nodeId, {
      parallelConfig: { ...config, ...updates },
    });
  };

  const handleAddStep = (type: StepType) => {
    const newStep: WorkflowStep = {
      id: `parallel_step_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name: `New ${getStepTypeLabel(type)} Step`,
      type,
      config: getDefaultConfigForType(type),
    };
    updateConfig({
      steps: [...config.steps, newStep],
    });
    setExpandedStep(config.steps.length);
  };

  const handleRemoveStep = (index: number) => {
    updateConfig({
      steps: config.steps.filter((_, i) => i !== index),
    });
    if (expandedStep === index) {
      setExpandedStep(null);
    }
  };

  const handleUpdateStep = (index: number, updates: Partial<WorkflowStep>) => {
    const newSteps = [...config.steps];
    newSteps[index] = { ...newSteps[index], ...updates };
    updateConfig({ steps: newSteps });
  };

  const handleMoveStep = (fromIndex: number, toIndex: number) => {
    if (toIndex < 0 || toIndex >= config.steps.length) return;
    const newSteps = [...config.steps];
    const [removed] = newSteps.splice(fromIndex, 1);
    newSteps.splice(toIndex, 0, removed);
    updateConfig({ steps: newSteps });
    setExpandedStep(toIndex);
  };

  return (
    <div className="space-y-4">
      {/* Parallel Options */}
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="waitForAll"
            checked={config.waitForAll}
            onChange={(e) => updateConfig({ waitForAll: e.target.checked })}
            className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
          />
          <label htmlFor="waitForAll" className="text-sm text-editor-text">
            Wait for all steps to complete
          </label>
        </div>

        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="failOnFirst"
            checked={config.failOnFirst}
            onChange={(e) => updateConfig({ failOnFirst: e.target.checked })}
            className="w-4 h-4 rounded border-editor-border bg-editor-surface text-editor-accent focus:ring-editor-accent"
          />
          <label htmlFor="failOnFirst" className="text-sm text-editor-text">
            Fail immediately on first error
          </label>
        </div>
      </div>

      {/* Nested Steps */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-editor-text">
            Parallel Steps ({config.steps.length})
          </label>
        </div>

        {/* Steps List */}
        <div className="space-y-2">
          {config.steps.map((step, index) => (
            <div
              key={step.id}
              className="border border-editor-border rounded-lg overflow-hidden bg-editor-surface/30"
            >
              {/* Step Header */}
              <div
                className="flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-editor-surface/50"
                onClick={() => setExpandedStep(expandedStep === index ? null : index)}
              >
                <div className="text-editor-muted cursor-grab">
                  <GripVertical size={14} />
                </div>
                <div className="text-editor-accent">{stepTypeIcons[step.type]}</div>
                <span className="flex-1 text-sm text-editor-text truncate">
                  {step.name}
                </span>
                <span className="text-xs text-editor-muted px-2 py-0.5 bg-editor-surface rounded">
                  {getStepTypeLabel(step.type)}
                </span>
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleMoveStep(index, index - 1);
                    }}
                    disabled={index === 0}
                    className="p-1 text-editor-muted hover:text-editor-text disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    <ChevronUp size={14} />
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleMoveStep(index, index + 1);
                    }}
                    disabled={index === config.steps.length - 1}
                    className="p-1 text-editor-muted hover:text-editor-text disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    <ChevronDown size={14} />
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleRemoveStep(index);
                    }}
                    className="p-1 text-editor-muted hover:text-editor-error"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              {/* Step Config (Expanded) */}
              {expandedStep === index && (
                <div className="px-3 py-3 border-t border-editor-border space-y-3">
                  <div className="space-y-2">
                    <label className="block text-xs text-editor-muted">Step Name</label>
                    <input
                      type="text"
                      value={step.name}
                      onChange={(e) => handleUpdateStep(index, { name: e.target.value })}
                      className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-editor-text text-sm focus:outline-none focus:border-editor-accent"
                    />
                  </div>

                  <div className="space-y-2">
                    <label className="block text-xs text-editor-muted">Description</label>
                    <textarea
                      value={step.description || ''}
                      onChange={(e) => handleUpdateStep(index, { description: e.target.value })}
                      placeholder="What this step does..."
                      rows={2}
                      className="w-full px-2 py-1.5 bg-editor-surface border border-editor-border rounded text-editor-text text-sm focus:outline-none focus:border-editor-accent resize-none"
                    />
                  </div>

                  <p className="text-xs text-editor-muted italic">
                    Configure detailed step settings in the main workflow canvas after saving.
                  </p>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Add Step Buttons */}
        <div className="pt-2">
          <div className="text-xs text-editor-muted mb-2">Add parallel step:</div>
          <div className="flex flex-wrap gap-2">
            {AVAILABLE_STEP_TYPES.map((type) => (
              <button
                key={type}
                type="button"
                onClick={() => handleAddStep(type)}
                className="flex items-center gap-1.5 px-2 py-1.5 text-xs text-editor-text bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent transition-colors"
              >
                {stepTypeIcons[type]}
                {getStepTypeLabel(type)}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Info */}
      <div className="p-3 bg-editor-surface/50 rounded-lg border border-editor-border">
        <p className="text-xs text-editor-muted">
          All steps in this parallel group will execute simultaneously. Use "Wait for all" to ensure all steps complete before continuing.
        </p>
      </div>
    </div>
  );
}

// Helper to get default config for step types
function getDefaultConfigForType(type: StepType) {
  switch (type) {
    case 'agent':
      return {
        agentConfig: {
          provider: 'anthropic',
          model: 'claude-3-sonnet-20240229',
          prompt: '',
          temperature: 0.7,
          maxTokens: 4096,
        },
      };
    case 'tool':
      return {
        toolConfig: {
          toolName: '',
          parameters: {},
        },
      };
    case 'wait':
      return {
        waitConfig: {
          waitType: 'user_input' as const,
          timeout: 300000,
        },
      };
    case 'transform':
      return {
        transformConfig: {
          type: 'template' as const,
        },
      };
    default:
      return {};
  }
}
