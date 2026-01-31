import { useCallback, useState, useMemo, type DragEvent } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { LucideIcon } from 'lucide-react';
import {
  Save,
  Play,
  Trash2,
  Bot,
  Wrench,
  GitBranch,
  GitFork,
  Clock,
  RefreshCw,
  ChevronLeft,
  Settings,
} from 'lucide-react';
import { WorkflowCanvas } from '../components/workflow/WorkflowCanvas';
import { useWorkflowStore } from '../store/workflowStore';
import { StepType, STEP_TYPE_LABELS } from '../types/workflow';

const stepTypeIcons: Record<StepType, LucideIcon> = {
  agent: Bot,
  tool: Wrench,
  condition: GitBranch,
  parallel: GitFork,
  wait: Clock,
  transform: RefreshCw,
};

const stepTypeDescriptions: Record<StepType, string> = {
  agent: 'Run AI agent with a prompt',
  tool: 'Execute a specific tool',
  condition: 'Branch based on condition',
  parallel: 'Run multiple steps at once',
  wait: 'Wait for input or timeout',
  transform: 'Transform data between steps',
};

export default function WorkflowDesigner() {
  // id is used for loading existing workflows (to be implemented)
  const { id: _workflowId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [selectedStepForConfig, setSelectedStepForConfig] = useState<string | null>(null);

  const {
    workflowName,
    workflowDescription,
    isDirty,
    nodes,
    clearCanvas,
    setWorkflowMeta,
    getWorkflowDefinition,
  } = useWorkflowStore();

  const handleSave = useCallback(() => {
    const definition = getWorkflowDefinition();
    console.log('Saving workflow:', definition);
    // TODO: Implement save API call
    alert('Workflow saved! (API not connected)');
  }, [getWorkflowDefinition]);

  const handleRun = useCallback(() => {
    const definition = getWorkflowDefinition();
    console.log('Running workflow:', definition);
    // TODO: Implement run API call
    alert('Workflow started! (API not connected)');
  }, [getWorkflowDefinition]);

  const handleClear = useCallback(() => {
    if (nodes.length === 0) return;
    if (confirm('Are you sure you want to clear all steps?')) {
      clearCanvas();
    }
  }, [nodes.length, clearCanvas]);

  const handleBack = useCallback(() => {
    if (isDirty && !confirm('You have unsaved changes. Are you sure you want to leave?')) {
      return;
    }
    navigate('/');
  }, [isDirty, navigate]);

  const handleNodeDoubleClick = useCallback((nodeId: string) => {
    setSelectedStepForConfig(nodeId);
  }, []);

  const handleDragStart = useCallback(
    (e: DragEvent<HTMLDivElement>, stepType: StepType) => {
      e.dataTransfer.setData('application/workflow-step', stepType);
      e.dataTransfer.effectAllowed = 'copy';
    },
    []
  );

  const selectedNode = useMemo(() => {
    if (!selectedStepForConfig) return null;
    return nodes.find((n) => n.id === selectedStepForConfig);
  }, [selectedStepForConfig, nodes]);

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {/* Header toolbar */}
      <div className="flex items-center justify-between px-4 py-3 bg-editor-surface border-b border-editor-border">
        <div className="flex items-center gap-4">
          <button
            onClick={handleBack}
            className="p-2 rounded-lg hover:bg-editor-bg transition-colors"
            title="Back to dashboard"
          >
            <ChevronLeft size={20} className="text-editor-muted" />
          </button>

          <div>
            <input
              type="text"
              value={workflowName}
              onChange={(e) => setWorkflowMeta(null, e.target.value, workflowDescription)}
              className="text-lg font-semibold bg-transparent border-none focus:outline-none text-editor-text placeholder:text-editor-muted"
              placeholder="Untitled Workflow"
            />
            <p className="text-xs text-editor-muted">
              {nodes.length} step{nodes.length !== 1 ? 's' : ''}
              {isDirty && <span className="ml-2 text-editor-warning">Unsaved changes</span>}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleClear}
            disabled={nodes.length === 0}
            className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Clear canvas"
          >
            <Trash2 size={16} className="text-editor-muted" />
            <span className="text-sm text-editor-muted">Clear</span>
          </button>

          <div className="w-px h-6 bg-editor-border" />

          <button
            onClick={handleSave}
            disabled={nodes.length === 0}
            className="flex items-center gap-2 px-3 py-2 rounded-lg bg-editor-bg hover:bg-editor-border disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Save workflow"
          >
            <Save size={16} className="text-editor-text" />
            <span className="text-sm text-editor-text">Save</span>
          </button>

          <button
            onClick={handleRun}
            disabled={nodes.length === 0}
            className="flex items-center gap-2 px-3 py-2 rounded-lg bg-editor-accent hover:bg-editor-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Run workflow"
          >
            <Play size={16} className="text-editor-bg" />
            <span className="text-sm text-editor-bg font-medium">Run</span>
          </button>
        </div>
      </div>

      {/* Main content area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Step library sidebar */}
        <div className="w-64 border-r border-editor-border bg-editor-surface overflow-y-auto">
          <div className="p-4">
            <h3 className="text-sm font-semibold text-editor-text mb-4">Step Library</h3>
            <div className="space-y-2">
              {(Object.keys(STEP_TYPE_LABELS) as StepType[]).map((stepType) => {
                const Icon = stepTypeIcons[stepType];
                return (
                  <div
                    key={stepType}
                    draggable
                    onDragStart={(e) => handleDragStart(e, stepType)}
                    className="flex items-start gap-3 p-3 rounded-lg bg-editor-bg border border-editor-border hover:border-editor-accent cursor-grab active:cursor-grabbing transition-colors"
                  >
                    <div className="p-2 rounded-md bg-editor-surface">
                      <Icon size={18} className="text-editor-accent" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm text-editor-text">
                        {STEP_TYPE_LABELS[stepType]}
                      </div>
                      <div className="text-xs text-editor-muted mt-0.5">
                        {stepTypeDescriptions[stepType]}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Canvas area */}
        <div className="flex-1 relative">
          <WorkflowCanvas onNodeDoubleClick={handleNodeDoubleClick} />
        </div>

        {/* Configuration panel placeholder */}
        {selectedNode && (
          <div className="w-80 border-l border-editor-border bg-editor-surface overflow-y-auto">
            <div className="p-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold text-editor-text flex items-center gap-2">
                  <Settings size={16} />
                  Step Configuration
                </h3>
                <button
                  onClick={() => setSelectedStepForConfig(null)}
                  className="text-editor-muted hover:text-editor-text"
                >
                  &times;
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-editor-muted mb-1">
                    Step Name
                  </label>
                  <input
                    type="text"
                    value={selectedNode.name}
                    onChange={(e) => {
                      useWorkflowStore.getState().updateNode(selectedNode.id, {
                        name: e.target.value,
                      });
                    }}
                    className="w-full px-3 py-2 rounded-lg bg-editor-bg border border-editor-border text-editor-text text-sm focus:outline-none focus:border-editor-accent"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-editor-muted mb-1">
                    Description
                  </label>
                  <textarea
                    value={selectedNode.description || ''}
                    onChange={(e) => {
                      useWorkflowStore.getState().updateNode(selectedNode.id, {
                        description: e.target.value,
                      });
                    }}
                    className="w-full px-3 py-2 rounded-lg bg-editor-bg border border-editor-border text-editor-text text-sm focus:outline-none focus:border-editor-accent resize-none"
                    rows={3}
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-editor-muted mb-1">
                    Step Type
                  </label>
                  <div className="px-3 py-2 rounded-lg bg-editor-bg border border-editor-border text-editor-text text-sm capitalize">
                    {selectedNode.type}
                  </div>
                </div>

                {/* Placeholder for step-specific configuration */}
                <div className="p-4 rounded-lg bg-editor-bg border border-dashed border-editor-border">
                  <p className="text-xs text-editor-muted text-center">
                    Step-specific configuration will be added in a future update
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
