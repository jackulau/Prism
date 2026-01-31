import { useState } from 'react';
import { Plus, Save, Play, Trash2, Bot, Wrench, GitBranch, Layers, Clock, Shuffle, Workflow, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useWorkflowStore, getStepTypeLabel } from '../store/workflowStore';
import { ConfigPanel } from '../components/workflow/config';
import type { StepType } from '../types/workflow';

// Step type palette
const STEP_TYPES: { type: StepType; icon: React.ReactNode; color: string }[] = [
  { type: 'agent', icon: <Bot size={20} />, color: 'text-blue-400' },
  { type: 'tool', icon: <Wrench size={20} />, color: 'text-green-400' },
  { type: 'condition', icon: <GitBranch size={20} />, color: 'text-yellow-400' },
  { type: 'parallel', icon: <Layers size={20} />, color: 'text-purple-400' },
  { type: 'wait', icon: <Clock size={20} />, color: 'text-orange-400' },
  { type: 'transform', icon: <Shuffle size={20} />, color: 'text-cyan-400' },
];

// Step type icons map
const stepTypeIcons: Record<StepType, React.ReactNode> = {
  agent: <Bot size={16} />,
  tool: <Wrench size={16} />,
  condition: <GitBranch size={16} />,
  parallel: <Layers size={16} />,
  wait: <Clock size={16} />,
  transform: <Shuffle size={16} />,
};

const stepTypeColors: Record<StepType, string> = {
  agent: 'border-blue-400/30 bg-blue-400/5',
  tool: 'border-green-400/30 bg-green-400/5',
  condition: 'border-yellow-400/30 bg-yellow-400/5',
  parallel: 'border-purple-400/30 bg-purple-400/5',
  wait: 'border-orange-400/30 bg-orange-400/5',
  transform: 'border-cyan-400/30 bg-cyan-400/5',
};

export default function WorkflowDesigner() {
  const {
    nodes,
    selectedNodeId,
    addNode,
    deleteNode,
    openConfigPanel,
    isDirty,
  } = useWorkflowStore();

  const [workflowName, setWorkflowName] = useState('New Workflow');

  const handleAddStep = (type: StepType) => {
    // Add node at a simple position based on existing nodes
    const position = {
      x: 100 + (nodes.length % 3) * 300,
      y: 100 + Math.floor(nodes.length / 3) * 150,
    };
    const node = addNode(type, position);
    openConfigPanel(node.id);
  };

  const handleSelectNode = (nodeId: string) => {
    openConfigPanel(nodeId);
  };

  const handleDeleteNode = (nodeId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Delete this step?')) {
      deleteNode(nodeId);
    }
  };

  return (
    <div className="h-full flex flex-col bg-editor-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/30">
        <div className="flex items-center gap-4">
          <Link
            to="/"
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <ArrowLeft size={20} />
          </Link>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-editor-accent/10 rounded-lg">
              <Workflow size={20} className="text-editor-accent" />
            </div>
            <input
              type="text"
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
              className="text-lg font-semibold text-editor-text bg-transparent border-none focus:outline-none focus:ring-0"
              placeholder="Workflow name..."
            />
            {isDirty && (
              <span className="px-2 py-0.5 text-xs bg-editor-warning/20 text-editor-warning rounded">
                Unsaved
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button className="flex items-center gap-2 px-3 py-2 text-sm text-editor-text bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent transition-colors">
            <Save size={16} />
            Save
          </button>
          <button className="flex items-center gap-2 px-3 py-2 text-sm text-white bg-editor-accent rounded-lg hover:bg-editor-accent/90 transition-colors">
            <Play size={16} />
            Run
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Step Palette */}
        <div className="w-64 border-r border-editor-border p-4 space-y-4 overflow-y-auto">
          <h3 className="text-sm font-medium text-editor-text">Add Step</h3>
          <div className="grid grid-cols-2 gap-2">
            {STEP_TYPES.map(({ type, icon, color }) => (
              <button
                key={type}
                onClick={() => handleAddStep(type)}
                className={`flex flex-col items-center gap-2 p-3 rounded-lg border border-editor-border bg-editor-surface/50 hover:border-editor-accent transition-colors`}
              >
                <div className={color}>{icon}</div>
                <span className="text-xs text-editor-text">{getStepTypeLabel(type)}</span>
              </button>
            ))}
          </div>

          {/* Workflow Info */}
          <div className="pt-4 border-t border-editor-border">
            <h3 className="text-sm font-medium text-editor-text mb-2">Workflow</h3>
            <div className="text-xs text-editor-muted space-y-1">
              <div>Steps: {nodes.length}</div>
            </div>
          </div>
        </div>

        {/* Canvas Area */}
        <div className="flex-1 overflow-auto p-6 bg-editor-bg/50">
          {nodes.length === 0 ? (
            <div className="h-full flex items-center justify-center">
              <div className="text-center">
                <div className="p-4 bg-editor-surface/50 rounded-full inline-block mb-4">
                  <Workflow size={48} className="text-editor-muted" />
                </div>
                <h3 className="text-lg font-medium text-editor-text mb-2">No steps yet</h3>
                <p className="text-sm text-editor-muted mb-4">
                  Add steps from the palette on the left to build your workflow.
                </p>
                <button
                  onClick={() => handleAddStep('agent')}
                  className="flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors mx-auto"
                >
                  <Plus size={16} />
                  Add First Step
                </button>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {nodes.map((node, index) => (
                <div
                  key={node.id}
                  onClick={() => handleSelectNode(node.id)}
                  className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${stepTypeColors[node.type]} ${
                    selectedNodeId === node.id
                      ? 'ring-2 ring-editor-accent ring-offset-2 ring-offset-editor-bg'
                      : 'hover:border-editor-accent/50'
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <span className="text-editor-accent">{stepTypeIcons[node.type]}</span>
                      <span className="text-xs text-editor-muted uppercase">
                        {getStepTypeLabel(node.type)}
                      </span>
                    </div>
                    <button
                      onClick={(e) => handleDeleteNode(node.id, e)}
                      className="p-1 text-editor-muted hover:text-editor-error rounded transition-colors"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                  <h4 className="font-medium text-editor-text truncate mb-1">
                    {node.data.name}
                  </h4>
                  {node.data.description && (
                    <p className="text-xs text-editor-muted line-clamp-2">
                      {node.data.description}
                    </p>
                  )}
                  <div className="mt-2 text-xs text-editor-muted">
                    Step {index + 1}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Config Panel */}
      <ConfigPanel />
    </div>
  );
}
