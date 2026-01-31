import { useEffect, useCallback } from 'react';
import { X, Trash2, Save, Bot, Wrench, GitBranch, Layers, Clock, Shuffle } from 'lucide-react';
import { useWorkflowStore, getStepTypeLabel } from '../../../store/workflowStore';
import type { StepType } from '../../../types/workflow';
import { AgentStepConfig } from './AgentStepConfig';
import { ToolStepConfig } from './ToolStepConfig';
import { ConditionStepConfig } from './ConditionStepConfig';
import { ParallelStepConfig } from './ParallelStepConfig';
import { WaitStepConfig } from './WaitStepConfig';
import { TransformStepConfig } from './TransformStepConfig';

// Icons for each step type
const stepTypeIcons: Record<StepType, React.ReactNode> = {
  agent: <Bot size={20} />,
  tool: <Wrench size={20} />,
  condition: <GitBranch size={20} />,
  parallel: <Layers size={20} />,
  wait: <Clock size={20} />,
  transform: <Shuffle size={20} />,
};

// Colors for each step type
const stepTypeColors: Record<StepType, string> = {
  agent: 'text-blue-400',
  tool: 'text-green-400',
  condition: 'text-yellow-400',
  parallel: 'text-purple-400',
  wait: 'text-orange-400',
  transform: 'text-cyan-400',
};

export function ConfigPanel() {
  const {
    isConfigPanelOpen,
    closeConfigPanel,
    getSelectedNode,
    deleteNode,
    updateNodeData,
    validationErrors,
  } = useWorkflowStore();

  const selectedNode = getSelectedNode();

  // Handle escape key to close panel
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isConfigPanelOpen) {
        closeConfigPanel();
      }
    },
    [isConfigPanelOpen, closeConfigPanel]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  if (!isConfigPanelOpen || !selectedNode) {
    return null;
  }

  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this step?')) {
      deleteNode(selectedNode.id);
    }
  };

  const handleNameChange = (name: string) => {
    updateNodeData(selectedNode.id, { name });
  };

  const handleDescriptionChange = (description: string) => {
    updateNodeData(selectedNode.id, { description });
  };

  const nodeErrors = validationErrors[selectedNode.id] || [];

  // Render the appropriate config form based on step type
  const renderConfigForm = () => {
    switch (selectedNode.type) {
      case 'agent':
        return <AgentStepConfig nodeId={selectedNode.id} />;
      case 'tool':
        return <ToolStepConfig nodeId={selectedNode.id} />;
      case 'condition':
        return <ConditionStepConfig nodeId={selectedNode.id} />;
      case 'parallel':
        return <ParallelStepConfig nodeId={selectedNode.id} />;
      case 'wait':
        return <WaitStepConfig nodeId={selectedNode.id} />;
      case 'transform':
        return <TransformStepConfig nodeId={selectedNode.id} />;
      default:
        return (
          <div className="p-4 text-editor-muted">
            Unknown step type: {selectedNode.type}
          </div>
        );
    }
  };

  return (
    <>
      {/* Backdrop for mobile */}
      <div
        className="fixed inset-0 bg-black/20 z-30 lg:hidden"
        onClick={closeConfigPanel}
      />

      {/* Panel */}
      <div className="fixed right-0 top-0 bottom-0 w-[400px] max-w-[90vw] bg-editor-bg border-l border-editor-border z-40 flex flex-col shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/30">
          <div className="flex items-center gap-3 min-w-0">
            <div className={`p-2 rounded-lg bg-editor-surface ${stepTypeColors[selectedNode.type]}`}>
              {stepTypeIcons[selectedNode.type]}
            </div>
            <div className="min-w-0">
              <span className="text-xs text-editor-muted uppercase tracking-wide">
                {getStepTypeLabel(selectedNode.type)} Step
              </span>
            </div>
          </div>
          <button
            onClick={closeConfigPanel}
            className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
            title="Close (Esc)"
          >
            <X size={20} />
          </button>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto">
          <div className="p-4 space-y-4">
            {/* Step Name */}
            <div className="space-y-2">
              <label className="block text-sm font-medium text-editor-text">
                Step Name
              </label>
              <input
                type="text"
                value={selectedNode.data.name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="Enter step name..."
                className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
              />
            </div>

            {/* Step Description */}
            <div className="space-y-2">
              <label className="block text-sm font-medium text-editor-text">
                Description
                <span className="text-editor-muted font-normal ml-1">(optional)</span>
              </label>
              <textarea
                value={selectedNode.data.description || ''}
                onChange={(e) => handleDescriptionChange(e.target.value)}
                placeholder="Describe what this step does..."
                rows={2}
                className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
              />
            </div>

            {/* Validation Errors */}
            {nodeErrors.length > 0 && (
              <div className="p-3 bg-editor-error/10 border border-editor-error/30 rounded-lg">
                <div className="text-sm font-medium text-editor-error mb-1">
                  Validation Errors
                </div>
                <ul className="text-sm text-editor-error/80 list-disc list-inside space-y-1">
                  {nodeErrors.map((error, i) => (
                    <li key={i}>{error}</li>
                  ))}
                </ul>
              </div>
            )}

            {/* Divider */}
            <div className="border-t border-editor-border" />

            {/* Step-specific Configuration */}
            {renderConfigForm()}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-editor-border bg-editor-surface/50">
          <button
            onClick={handleDelete}
            className="flex items-center gap-2 px-3 py-2 text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
            title="Delete step"
          >
            <Trash2 size={16} />
            <span className="text-sm">Delete</span>
          </button>
          <button
            onClick={closeConfigPanel}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Save size={16} />
            <span className="text-sm">Done</span>
          </button>
        </div>
      </div>
    </>
  );
}
