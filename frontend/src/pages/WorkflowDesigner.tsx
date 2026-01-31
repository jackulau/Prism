import { useState, useCallback, useRef } from 'react';
import { Plus, Play, Save, Trash2, ZoomIn, ZoomOut, Maximize2 } from 'lucide-react';
import { StepLibrary } from '../components/workflow/StepLibrary';
import { useWorkflowTemplates, type WorkflowTemplate } from '../hooks/useWorkflowTemplates';
import { getStepTypeDefinition, type StepType, type StepTypeDefinition } from '../config/workflowStepTypes';

interface WorkflowNode {
  id: string;
  type: StepType;
  name: string;
  config: Record<string, unknown>;
  position: { x: number; y: number };
}

export default function WorkflowDesigner() {
  const [isLibraryCollapsed, setIsLibraryCollapsed] = useState(false);
  const [nodes, setNodes] = useState<WorkflowNode[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [isDraggingOver, setIsDraggingOver] = useState(false);
  const [workflowName, setWorkflowName] = useState('Untitled Workflow');
  const canvasRef = useRef<HTMLDivElement>(null);

  const { templates } = useWorkflowTemplates();

  const handleDragStart = useCallback((_stepType: StepTypeDefinition) => {
    // Visual feedback could be added here
  }, []);

  const handleDragEnd = useCallback(() => {
    setIsDraggingOver(false);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    setIsDraggingOver(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setIsDraggingOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDraggingOver(false);

    const data = e.dataTransfer.getData('application/workflow-step');
    if (!data) return;

    try {
      const stepData = JSON.parse(data);
      const canvasRect = canvasRef.current?.getBoundingClientRect();
      if (!canvasRect) return;

      const position = {
        x: e.clientX - canvasRect.left - 100,
        y: e.clientY - canvasRect.top - 40,
      };

      const newNode: WorkflowNode = {
        id: `node-${Date.now()}`,
        type: stepData.type,
        name: stepData.name,
        config: stepData.defaultConfig || {},
        position,
      };

      setNodes((prev) => [...prev, newNode]);
      setSelectedNode(newNode.id);
    } catch (err) {
      console.error('Failed to parse drop data:', err);
    }
  }, []);

  const handleTemplateClick = useCallback((template: WorkflowTemplate) => {
    const newNodes = template.steps.map((step, index) => ({
      id: `node-${Date.now()}-${index}`,
      type: step.type,
      name: step.name,
      config: step.config,
      position: {
        x: 100 + (index % 3) * 250,
        y: 100 + Math.floor(index / 3) * 150,
      },
    }));
    setNodes(newNodes);
    setWorkflowName(template.name);
  }, []);

  const handleTemplatePreview = useCallback((template: WorkflowTemplate) => {
    // TODO: Show template preview modal
    console.log('Preview template:', template);
  }, []);

  const handleNodeClick = useCallback((nodeId: string) => {
    setSelectedNode(nodeId);
  }, []);

  const handleDeleteNode = useCallback((nodeId: string) => {
    setNodes((prev) => prev.filter((n) => n.id !== nodeId));
    if (selectedNode === nodeId) {
      setSelectedNode(null);
    }
  }, [selectedNode]);

  const handleClearCanvas = useCallback(() => {
    setNodes([]);
    setSelectedNode(null);
  }, []);

  return (
    <div className="flex-1 flex h-full overflow-hidden">
      <StepLibrary
        isCollapsed={isLibraryCollapsed}
        onToggle={() => setIsLibraryCollapsed(!isLibraryCollapsed)}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        templates={templates}
        onTemplateClick={handleTemplateClick}
        onTemplatePreview={handleTemplatePreview}
      />

      <div className="flex-1 flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-bg">
          <div className="flex items-center gap-4">
            <input
              type="text"
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
              className="bg-transparent text-lg font-semibold text-editor-text border-b border-transparent hover:border-editor-border focus:border-editor-accent focus:outline-none px-1"
            />
            <span className="text-xs text-editor-muted">
              {nodes.length} step{nodes.length !== 1 ? 's' : ''}
            </span>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleClearCanvas}
              disabled={nodes.length === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Clear canvas"
            >
              <Trash2 size={14} />
              Clear
            </button>
            <button
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors"
              title="Save workflow"
            >
              <Save size={14} />
              Save
            </button>
            <button
              disabled={nodes.length === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent text-white rounded hover:bg-editor-accent/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Run workflow"
            >
              <Play size={14} />
              Run
            </button>
          </div>
        </div>

        <div
          ref={canvasRef}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          className={`flex-1 relative overflow-auto bg-editor-bg ${
            isDraggingOver ? 'ring-2 ring-inset ring-editor-accent/50' : ''
          }`}
          style={{
            backgroundImage: `
              linear-gradient(to right, var(--color-editor-border) 1px, transparent 1px),
              linear-gradient(to bottom, var(--color-editor-border) 1px, transparent 1px)
            `,
            backgroundSize: '40px 40px',
          }}
        >
          {nodes.length === 0 ? (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="text-center p-8">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-editor-surface/50 flex items-center justify-center">
                  <Plus size={32} className="text-editor-muted" />
                </div>
                <h3 className="text-lg font-medium text-editor-text mb-2">
                  Start building your workflow
                </h3>
                <p className="text-sm text-editor-muted max-w-md">
                  Drag steps from the library on the left to add them to your workflow,
                  or select a template to get started quickly.
                </p>
              </div>
            </div>
          ) : (
            nodes.map((node) => (
              <WorkflowNodeCard
                key={node.id}
                node={node}
                isSelected={selectedNode === node.id}
                onClick={() => handleNodeClick(node.id)}
                onDelete={() => handleDeleteNode(node.id)}
              />
            ))
          )}

          {isDraggingOver && (
            <div className="absolute inset-0 bg-editor-accent/5 pointer-events-none flex items-center justify-center">
              <div className="px-4 py-2 bg-editor-surface border border-editor-accent rounded-lg shadow-lg">
                <span className="text-sm text-editor-accent">Drop to add step</span>
              </div>
            </div>
          )}
        </div>

        <div className="flex items-center justify-between px-4 py-2 border-t border-editor-border bg-editor-surface">
          <div className="flex items-center gap-2">
            <button
              className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors"
              title="Zoom out"
            >
              <ZoomOut size={14} />
            </button>
            <span className="text-xs text-editor-muted px-2">100%</span>
            <button
              className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors"
              title="Zoom in"
            >
              <ZoomIn size={14} />
            </button>
            <button
              className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-sidebar-hover rounded transition-colors"
              title="Fit to screen"
            >
              <Maximize2 size={14} />
            </button>
          </div>
          <div className="text-xs text-editor-muted">
            Workflow Designer
          </div>
        </div>
      </div>
    </div>
  );
}

interface WorkflowNodeCardProps {
  node: WorkflowNode;
  isSelected: boolean;
  onClick: () => void;
  onDelete: () => void;
}

function WorkflowNodeCard({ node, isSelected, onClick, onDelete }: WorkflowNodeCardProps) {
  const stepDef = getStepTypeDefinition(node.type);
  const Icon = stepDef?.icon;

  return (
    <div
      className={`absolute w-48 bg-editor-surface border rounded-lg shadow-md cursor-pointer transition-all ${
        isSelected
          ? 'border-editor-accent ring-2 ring-editor-accent/20'
          : 'border-editor-border hover:border-editor-muted'
      }`}
      style={{
        left: node.position.x,
        top: node.position.y,
      }}
      onClick={onClick}
    >
      <div className="flex items-center gap-2 p-3 border-b border-editor-border/50">
        {Icon && stepDef && (
          <div className={`p-1.5 rounded ${stepDef.bgColor}`}>
            <Icon size={14} className={stepDef.color} />
          </div>
        )}
        <span className="flex-1 text-sm font-medium text-editor-text truncate">
          {node.name}
        </span>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="p-1 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded transition-colors opacity-0 group-hover:opacity-100"
        >
          <Trash2 size={12} />
        </button>
      </div>
      <div className="p-2">
        <div className="text-xs text-editor-muted capitalize">{node.type} step</div>
      </div>
    </div>
  );
}
