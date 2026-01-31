import { memo, useCallback, useRef } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  Bot,
  Wrench,
  GitBranch,
  GitFork,
  Clock,
  RefreshCw,
  GripVertical,
} from 'lucide-react';
import {
  WorkflowNode as WorkflowNodeType,
  StepType,
  NODE_DIMENSIONS,
  NodeVisualState,
} from '../../types/workflow';
import { useWorkflowStore } from '../../store/workflowStore';

interface WorkflowNodeProps {
  node: WorkflowNodeType;
  isSelected: boolean;
  onDoubleClick?: (nodeId: string) => void;
}

const stepTypeIcons: Record<StepType, LucideIcon> = {
  agent: Bot,
  tool: Wrench,
  condition: GitBranch,
  parallel: GitFork,
  wait: Clock,
  transform: RefreshCw,
};

const stepTypeColors: Record<StepType, string> = {
  agent: 'border-blue-500/50 bg-blue-500/10',
  tool: 'border-purple-500/50 bg-purple-500/10',
  condition: 'border-yellow-500/50 bg-yellow-500/10',
  parallel: 'border-green-500/50 bg-green-500/10',
  wait: 'border-orange-500/50 bg-orange-500/10',
  transform: 'border-cyan-500/50 bg-cyan-500/10',
};

const stepTypeIconColors: Record<StepType, string> = {
  agent: 'text-blue-400',
  tool: 'text-purple-400',
  condition: 'text-yellow-400',
  parallel: 'text-green-400',
  wait: 'text-orange-400',
  transform: 'text-cyan-400',
};

const visualStateStyles: Record<NodeVisualState, string> = {
  default: '',
  selected: 'ring-2 ring-editor-accent ring-offset-2 ring-offset-editor-bg',
  running: 'ring-2 ring-yellow-500 animate-pulse',
  completed: 'ring-2 ring-editor-success',
  failed: 'ring-2 ring-editor-error',
  hovering: 'ring-1 ring-editor-border',
};

export const WorkflowNode = memo(function WorkflowNode({
  node,
  isSelected,
  onDoubleClick,
}: WorkflowNodeProps) {
  const nodeRef = useRef<HTMLDivElement>(null);
  const dragStartRef = useRef<{ x: number; y: number } | null>(null);

  const {
    selectNode,
    moveNode,
    setDragging,
    startConnection,
    zoom,
  } = useWorkflowStore();

  const Icon = stepTypeIcons[node.type];
  const colorClass = stepTypeColors[node.type];
  const iconColorClass = stepTypeIconColors[node.type];
  const visualState = isSelected ? 'selected' : (node.visualState || 'default');
  const visualStateClass = visualStateStyles[visualState];

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      if (e.button !== 0) return; // Only handle left click
      e.stopPropagation();

      selectNode(node.id);
      setDragging(true);

      const startX = e.clientX;
      const startY = e.clientY;
      const startNodeX = node.position.x;
      const startNodeY = node.position.y;

      dragStartRef.current = { x: startX, y: startY };

      const handleMouseMove = (moveEvent: MouseEvent) => {
        const dx = (moveEvent.clientX - startX) / zoom;
        const dy = (moveEvent.clientY - startY) / zoom;

        moveNode(node.id, {
          x: startNodeX + dx,
          y: startNodeY + dy,
        });
      };

      const handleMouseUp = () => {
        setDragging(false);
        dragStartRef.current = null;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };

      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    },
    [node.id, node.position, selectNode, moveNode, setDragging, zoom]
  );

  const handleDoubleClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onDoubleClick?.(node.id);
    },
    [node.id, onDoubleClick]
  );

  const handlePortMouseDown = useCallback(
    (e: React.MouseEvent, portType: 'success' | 'failure' | 'default') => {
      e.stopPropagation();
      startConnection(node.id, portType);
    },
    [node.id, startConnection]
  );

  // Determine which output ports to show based on node type
  const showConditionPorts = node.type === 'condition';

  return (
    <div
      ref={nodeRef}
      className={`
        absolute cursor-grab active:cursor-grabbing
        border rounded-lg shadow-lg
        bg-editor-surface
        transition-all duration-150
        ${colorClass}
        ${visualStateClass}
      `}
      style={{
        left: node.position.x,
        top: node.position.y,
        width: NODE_DIMENSIONS.width,
        minHeight: NODE_DIMENSIONS.height,
      }}
      onMouseDown={handleMouseDown}
      onDoubleClick={handleDoubleClick}
    >
      {/* Drag handle */}
      <div className="absolute -left-1 top-1/2 -translate-y-1/2 p-1 cursor-grab opacity-50 hover:opacity-100">
        <GripVertical size={12} className="text-editor-muted" />
      </div>

      {/* Node content */}
      <div className="p-3 flex items-start gap-3">
        <div className={`p-2 rounded-md bg-editor-bg ${iconColorClass}`}>
          <Icon size={20} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-medium text-sm text-editor-text truncate">
            {node.name}
          </div>
          <div className="text-xs text-editor-muted capitalize">
            {node.type}
          </div>
        </div>
      </div>

      {/* Input port (top center) */}
      <div
        className="absolute -top-2 left-1/2 -translate-x-1/2 w-4 h-4 rounded-full bg-editor-surface border-2 border-editor-border hover:border-editor-accent cursor-crosshair z-10"
        data-port="input"
        data-node-id={node.id}
      />

      {/* Output ports */}
      {showConditionPorts ? (
        <>
          {/* True branch (left) */}
          <div
            className="absolute -bottom-2 left-1/4 -translate-x-1/2 w-4 h-4 rounded-full bg-editor-surface border-2 border-green-500 hover:bg-green-500/20 cursor-crosshair z-10"
            data-port="success"
            data-node-id={node.id}
            onMouseDown={(e) => handlePortMouseDown(e, 'success')}
            title="True branch"
          />
          {/* False branch (right) */}
          <div
            className="absolute -bottom-2 left-3/4 -translate-x-1/2 w-4 h-4 rounded-full bg-editor-surface border-2 border-red-500 hover:bg-red-500/20 cursor-crosshair z-10"
            data-port="failure"
            data-node-id={node.id}
            onMouseDown={(e) => handlePortMouseDown(e, 'failure')}
            title="False branch"
          />
        </>
      ) : (
        /* Default output port (bottom center) */
        <div
          className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-4 h-4 rounded-full bg-editor-surface border-2 border-editor-border hover:border-editor-accent cursor-crosshair z-10"
          data-port="default"
          data-node-id={node.id}
          onMouseDown={(e) => handlePortMouseDown(e, 'default')}
        />
      )}

      {/* Status indicator */}
      {node.status && (
        <div
          className={`
            absolute -top-1 -right-1 w-3 h-3 rounded-full border border-editor-bg
            ${node.status === 'running' ? 'bg-yellow-500 animate-pulse' : ''}
            ${node.status === 'completed' ? 'bg-editor-success' : ''}
            ${node.status === 'failed' ? 'bg-editor-error' : ''}
            ${node.status === 'pending' ? 'bg-editor-muted' : ''}
            ${node.status === 'skipped' ? 'bg-editor-border' : ''}
          `}
        />
      )}
    </div>
  );
});

export default WorkflowNode;
