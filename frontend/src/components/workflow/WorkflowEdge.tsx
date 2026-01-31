import { memo, useMemo, type MouseEvent } from 'react';
import { Edge, WorkflowNode, NODE_DIMENSIONS, Position } from '../../types/workflow';
import { useWorkflowStore } from '../../store/workflowStore';

interface WorkflowEdgeProps {
  edge: Edge;
  sourceNode: WorkflowNode;
  targetNode: WorkflowNode;
  isSelected: boolean;
}

interface PendingEdgeProps {
  sourceNode: WorkflowNode;
  sourcePort: 'success' | 'failure' | 'default';
  mousePosition: Position;
}

const getPortPosition = (
  node: WorkflowNode,
  portType: 'input' | 'success' | 'failure' | 'default'
): Position => {
  const { x, y } = node.position;
  const { width, height } = NODE_DIMENSIONS;

  switch (portType) {
    case 'input':
      return { x: x + width / 2, y };
    case 'success':
      return { x: x + width / 4, y: y + height };
    case 'failure':
      return { x: x + (width * 3) / 4, y: y + height };
    case 'default':
    default:
      return { x: x + width / 2, y: y + height };
  }
};

const calculateBezierPath = (
  start: Position,
  end: Position,
  sourceIsBottom: boolean = true
): string => {
  const dy = end.y - start.y;

  // Calculate control point offsets based on distance
  const controlOffset = Math.min(Math.abs(dy) * 0.5, 100);

  // Control points for smooth curve
  const cp1 = {
    x: start.x,
    y: sourceIsBottom ? start.y + controlOffset : start.y - controlOffset,
  };
  const cp2 = {
    x: end.x,
    y: end.y - controlOffset,
  };

  return `M ${start.x} ${start.y} C ${cp1.x} ${cp1.y}, ${cp2.x} ${cp2.y}, ${end.x} ${end.y}`;
};

export const WorkflowEdge = memo(function WorkflowEdge({
  edge,
  sourceNode,
  targetNode,
  isSelected,
}: WorkflowEdgeProps) {
  const { selectEdge, removeEdge } = useWorkflowStore();

  const { path, labelPosition, color } = useMemo(() => {
    const sourcePort = edge.sourcePort || 'default';
    const start = getPortPosition(sourceNode, sourcePort);
    const end = getPortPosition(targetNode, 'input');

    const path = calculateBezierPath(start, end);

    // Calculate label position at the middle of the curve
    const labelPosition = {
      x: (start.x + end.x) / 2,
      y: (start.y + end.y) / 2,
    };

    // Color based on label
    let color = 'text-editor-muted';
    if (edge.label === 'success' || edge.label === 'true') {
      color = 'text-green-500';
    } else if (edge.label === 'failure' || edge.label === 'false') {
      color = 'text-red-500';
    }

    return { path, labelPosition, color };
  }, [edge, sourceNode, targetNode]);

  const handleClick = (e: MouseEvent) => {
    e.stopPropagation();
    selectEdge(edge.id);
  };

  const handleDoubleClick = (e: MouseEvent) => {
    e.stopPropagation();
    removeEdge(edge.id);
  };

  const strokeColor = isSelected
    ? 'stroke-editor-accent'
    : edge.label === 'success' || edge.label === 'true'
    ? 'stroke-green-500'
    : edge.label === 'failure' || edge.label === 'false'
    ? 'stroke-red-500'
    : 'stroke-editor-muted';

  return (
    <g className="cursor-pointer" onClick={handleClick} onDoubleClick={handleDoubleClick}>
      {/* Invisible wider path for easier clicking */}
      <path
        d={path}
        fill="none"
        stroke="transparent"
        strokeWidth={20}
        className="cursor-pointer"
      />

      {/* Visible edge path */}
      <path
        d={path}
        fill="none"
        className={`${strokeColor} transition-colors duration-150`}
        strokeWidth={isSelected ? 3 : 2}
        strokeLinecap="round"
        markerEnd={`url(#arrowhead-${edge.label || 'default'})`}
      />

      {/* Edge label */}
      {edge.label && (
        <g transform={`translate(${labelPosition.x}, ${labelPosition.y})`}>
          <rect
            x={-20}
            y={-10}
            width={40}
            height={20}
            rx={4}
            className="fill-editor-bg stroke-editor-border"
          />
          <text
            textAnchor="middle"
            dominantBaseline="central"
            className={`text-xs font-medium ${color} select-none`}
            fill="currentColor"
          >
            {edge.label}
          </text>
        </g>
      )}
    </g>
  );
});

export const PendingEdge = memo(function PendingEdge({
  sourceNode,
  sourcePort,
  mousePosition,
}: PendingEdgeProps) {
  const start = getPortPosition(sourceNode, sourcePort);
  const path = calculateBezierPath(start, mousePosition);

  const strokeColor =
    sourcePort === 'success'
      ? 'stroke-green-500'
      : sourcePort === 'failure'
      ? 'stroke-red-500'
      : 'stroke-editor-accent';

  return (
    <path
      d={path}
      fill="none"
      className={`${strokeColor} opacity-50`}
      strokeWidth={2}
      strokeDasharray="8 4"
      strokeLinecap="round"
    />
  );
});

// SVG definitions for arrow markers
export const EdgeArrowDefs = memo(function EdgeArrowDefs() {
  return (
    <defs>
      <marker
        id="arrowhead-default"
        markerWidth="10"
        markerHeight="10"
        refX="8"
        refY="5"
        orient="auto"
      >
        <path
          d="M 0 0 L 10 5 L 0 10 z"
          className="fill-editor-muted"
        />
      </marker>
      <marker
        id="arrowhead-success"
        markerWidth="10"
        markerHeight="10"
        refX="8"
        refY="5"
        orient="auto"
      >
        <path
          d="M 0 0 L 10 5 L 0 10 z"
          className="fill-green-500"
        />
      </marker>
      <marker
        id="arrowhead-true"
        markerWidth="10"
        markerHeight="10"
        refX="8"
        refY="5"
        orient="auto"
      >
        <path
          d="M 0 0 L 10 5 L 0 10 z"
          className="fill-green-500"
        />
      </marker>
      <marker
        id="arrowhead-failure"
        markerWidth="10"
        markerHeight="10"
        refX="8"
        refY="5"
        orient="auto"
      >
        <path
          d="M 0 0 L 10 5 L 0 10 z"
          className="fill-red-500"
        />
      </marker>
      <marker
        id="arrowhead-false"
        markerWidth="10"
        markerHeight="10"
        refX="8"
        refY="5"
        orient="auto"
      >
        <path
          d="M 0 0 L 10 5 L 0 10 z"
          className="fill-red-500"
        />
      </marker>
    </defs>
  );
});

export default WorkflowEdge;
