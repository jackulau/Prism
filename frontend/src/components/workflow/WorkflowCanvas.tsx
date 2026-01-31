import { useCallback, useEffect, useRef, useState, memo, type MouseEvent, type WheelEvent, type DragEvent } from 'react';
import { useWorkflowStore } from '../../store/workflowStore';
import { WorkflowNode } from './WorkflowNode';
import { WorkflowEdge, PendingEdge, EdgeArrowDefs } from './WorkflowEdge';
import { CanvasControls } from './CanvasControls';
import { CANVAS_GRID, ZOOM_CONSTRAINTS, StepType } from '../../types/workflow';

interface WorkflowCanvasProps {
  onNodeDoubleClick?: (nodeId: string) => void;
  className?: string;
}

export const WorkflowCanvas = memo(function WorkflowCanvas({
  onNodeDoubleClick,
  className = '',
}: WorkflowCanvasProps) {
  const canvasRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const [canvasSize, setCanvasSize] = useState({ width: 800, height: 600 });

  const {
    nodes,
    edges,
    selectedNodeId,
    selectedEdgeId,
    zoom,
    pan,
    setPan,
    setZoom,
    pendingConnection,
    updateConnectionPosition,
    completeConnection,
    cancelConnection,
    clearSelection,
    deleteSelected,
    undo,
    redo,
    setPanning,
    isDragging,
    isPanning,
  } = useWorkflowStore();

  // Update canvas size on mount and resize
  useEffect(() => {
    const updateSize = () => {
      if (canvasRef.current) {
        const rect = canvasRef.current.getBoundingClientRect();
        setCanvasSize({ width: rect.width, height: rect.height });
      }
    };

    updateSize();
    window.addEventListener('resize', updateSize);
    return () => window.removeEventListener('resize', updateSize);
  }, []);

  // Handle canvas pan (middle mouse or space + drag)
  const handleCanvasMouseDown = useCallback(
    (e: MouseEvent<HTMLDivElement>) => {
      // Only pan with left click on empty space
      if (e.button !== 0) return;
      if ((e.target as HTMLElement).closest('[data-node-id]')) return;
      if ((e.target as HTMLElement).closest('[data-port]')) return;

      // Clear selection when clicking empty space
      clearSelection();

      const startX = e.clientX;
      const startY = e.clientY;
      const startPanX = pan.x;
      const startPanY = pan.y;

      setPanning(true);

      const handleMouseMove = (moveEvent: globalThis.MouseEvent) => {
        const dx = moveEvent.clientX - startX;
        const dy = moveEvent.clientY - startY;
        setPan({ x: startPanX + dx, y: startPanY + dy });
      };

      const handleMouseUp = () => {
        setPanning(false);
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };

      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    },
    [pan, setPan, clearSelection, setPanning]
  );

  // Handle mouse move for pending connections
  const handleCanvasMouseMove = useCallback(
    (e: MouseEvent<HTMLDivElement>) => {
      if (!pendingConnection) return;

      const rect = canvasRef.current?.getBoundingClientRect();
      if (!rect) return;

      const x = (e.clientX - rect.left - pan.x) / zoom;
      const y = (e.clientY - rect.top - pan.y) / zoom;
      updateConnectionPosition({ x, y });
    },
    [pendingConnection, pan, zoom, updateConnectionPosition]
  );

  // Handle connection completion on mouse up
  const handleCanvasMouseUp = useCallback(
    (e: MouseEvent<HTMLDivElement>) => {
      if (!pendingConnection) return;

      // Find if mouse is over an input port
      const target = e.target as HTMLElement;
      const portElement = target.closest('[data-port="input"]');
      if (portElement) {
        const nodeId = portElement.getAttribute('data-node-id');
        if (nodeId && nodeId !== pendingConnection.sourceNodeId) {
          completeConnection(nodeId);
          return;
        }
      }

      cancelConnection();
    },
    [pendingConnection, completeConnection, cancelConnection]
  );

  // Handle zoom with scroll wheel
  const handleWheel = useCallback(
    (e: WheelEvent<HTMLDivElement>) => {
      e.preventDefault();

      const rect = canvasRef.current?.getBoundingClientRect();
      if (!rect) return;

      // Mouse position relative to canvas
      const mouseX = e.clientX - rect.left;
      const mouseY = e.clientY - rect.top;

      // Calculate new zoom level
      const delta = e.deltaY > 0 ? -ZOOM_CONSTRAINTS.step : ZOOM_CONSTRAINTS.step;
      const newZoom = Math.max(
        ZOOM_CONSTRAINTS.min,
        Math.min(ZOOM_CONSTRAINTS.max, zoom + delta)
      );

      if (newZoom === zoom) return;

      // Adjust pan to zoom toward mouse position
      const zoomRatio = newZoom / zoom;
      const newPanX = mouseX - (mouseX - pan.x) * zoomRatio;
      const newPanY = mouseY - (mouseY - pan.y) * zoomRatio;

      setZoom(newZoom);
      setPan({ x: newPanX, y: newPanY });
    },
    [zoom, pan, setZoom, setPan]
  );

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if focus is on an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      // Delete selected
      if (e.key === 'Delete' || e.key === 'Backspace') {
        e.preventDefault();
        deleteSelected();
      }

      // Escape to deselect
      if (e.key === 'Escape') {
        if (pendingConnection) {
          cancelConnection();
        } else {
          clearSelection();
        }
      }

      // Undo: Ctrl/Cmd + Z
      if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
        e.preventDefault();
        undo();
      }

      // Redo: Ctrl/Cmd + Shift + Z or Ctrl/Cmd + Y
      if (
        (e.ctrlKey || e.metaKey) &&
        ((e.key === 'z' && e.shiftKey) || e.key === 'y')
      ) {
        e.preventDefault();
        redo();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [deleteSelected, clearSelection, cancelConnection, pendingConnection, undo, redo]);

  // Handle drop for adding nodes
  const handleDrop = useCallback(
    (e: DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      const stepType = e.dataTransfer.getData('application/workflow-step') as StepType;
      if (!stepType) return;

      const rect = canvasRef.current?.getBoundingClientRect();
      if (!rect) return;

      const x = (e.clientX - rect.left - pan.x) / zoom;
      const y = (e.clientY - rect.top - pan.y) / zoom;

      useWorkflowStore.getState().addNode(stepType, { x, y });
    },
    [pan, zoom]
  );

  const handleDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
  }, []);

  // Get source node for pending connection
  const pendingSourceNode = pendingConnection
    ? nodes.find((n) => n.id === pendingConnection.sourceNodeId)
    : null;

  return (
    <div
      ref={canvasRef}
      className={`relative w-full h-full overflow-hidden bg-editor-bg ${className}`}
      onMouseDown={handleCanvasMouseDown}
      onMouseMove={handleCanvasMouseMove}
      onMouseUp={handleCanvasMouseUp}
      onWheel={handleWheel}
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      style={{ cursor: isPanning ? 'grabbing' : isDragging ? 'grabbing' : 'default' }}
    >
      {/* Grid background */}
      <svg
        className="absolute inset-0 w-full h-full pointer-events-none"
        style={{ opacity: 0.3 }}
      >
        <defs>
          <pattern
            id="grid-pattern"
            width={CANVAS_GRID.size * zoom}
            height={CANVAS_GRID.size * zoom}
            patternUnits="userSpaceOnUse"
            x={pan.x % (CANVAS_GRID.size * zoom)}
            y={pan.y % (CANVAS_GRID.size * zoom)}
          >
            <circle
              cx={CANVAS_GRID.size * zoom / 2}
              cy={CANVAS_GRID.size * zoom / 2}
              r={1}
              className="fill-editor-border"
            />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid-pattern)" />
      </svg>

      {/* Transformed content container */}
      <div
        className="absolute origin-top-left"
        style={{
          transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
          width: canvasSize.width / zoom,
          height: canvasSize.height / zoom,
        }}
      >
        {/* SVG layer for edges */}
        <svg
          ref={svgRef}
          className="absolute inset-0 overflow-visible pointer-events-none"
          style={{
            width: '100%',
            height: '100%',
          }}
        >
          <EdgeArrowDefs />

          {/* Render edges */}
          <g className="pointer-events-auto">
            {edges.map((edge) => {
              const sourceNode = nodes.find((n) => n.id === edge.source);
              const targetNode = nodes.find((n) => n.id === edge.target);
              if (!sourceNode || !targetNode) return null;

              return (
                <WorkflowEdge
                  key={edge.id}
                  edge={edge}
                  sourceNode={sourceNode}
                  targetNode={targetNode}
                  isSelected={selectedEdgeId === edge.id}
                />
              );
            })}
          </g>

          {/* Render pending connection */}
          {pendingConnection && pendingSourceNode && (
            <PendingEdge
              sourceNode={pendingSourceNode}
              sourcePort={pendingConnection.sourcePort}
              mousePosition={pendingConnection.mousePosition}
            />
          )}
        </svg>

        {/* Nodes layer */}
        {nodes.map((node) => (
          <WorkflowNode
            key={node.id}
            node={node}
            isSelected={selectedNodeId === node.id}
            onDoubleClick={onNodeDoubleClick}
          />
        ))}
      </div>

      {/* Canvas controls */}
      <CanvasControls className="absolute bottom-4 left-1/2 -translate-x-1/2" />

      {/* Empty state */}
      {nodes.length === 0 && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center text-editor-muted">
            <p className="text-lg font-medium">No steps yet</p>
            <p className="text-sm mt-1">Drag steps from the sidebar to get started</p>
          </div>
        </div>
      )}
    </div>
  );
});

export default WorkflowCanvas;
