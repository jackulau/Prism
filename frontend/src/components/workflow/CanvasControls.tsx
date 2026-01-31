import { memo } from 'react';
import {
  ZoomIn,
  ZoomOut,
  Maximize2,
  RotateCcw,
  Undo2,
  Redo2,
  Trash2,
} from 'lucide-react';
import { useWorkflowStore } from '../../store/workflowStore';
import { ZOOM_CONSTRAINTS } from '../../types/workflow';

interface CanvasControlsProps {
  className?: string;
}

export const CanvasControls = memo(function CanvasControls({
  className = '',
}: CanvasControlsProps) {
  const {
    zoom,
    zoomIn,
    zoomOut,
    resetZoom,
    fitToView,
    undo,
    redo,
    deleteSelected,
    selectedNodeId,
    selectedEdgeId,
    historyIndex,
    history,
  } = useWorkflowStore();

  const canUndo = historyIndex > 0;
  const canRedo = historyIndex < history.length - 1;
  const hasSelection = selectedNodeId !== null || selectedEdgeId !== null;

  const zoomPercentage = Math.round(zoom * 100);

  return (
    <div
      className={`flex items-center gap-1 p-1 bg-editor-surface/90 backdrop-blur-sm border border-editor-border rounded-lg shadow-lg ${className}`}
    >
      {/* Zoom controls */}
      <div className="flex items-center gap-1 px-2">
        <button
          onClick={zoomOut}
          disabled={zoom <= ZOOM_CONSTRAINTS.min}
          className="p-1.5 rounded hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          title="Zoom out"
        >
          <ZoomOut size={16} className="text-editor-muted" />
        </button>

        <span className="w-12 text-center text-xs font-medium text-editor-text tabular-nums">
          {zoomPercentage}%
        </span>

        <button
          onClick={zoomIn}
          disabled={zoom >= ZOOM_CONSTRAINTS.max}
          className="p-1.5 rounded hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          title="Zoom in"
        >
          <ZoomIn size={16} className="text-editor-muted" />
        </button>
      </div>

      <div className="w-px h-6 bg-editor-border" />

      {/* View controls */}
      <div className="flex items-center gap-1 px-2">
        <button
          onClick={fitToView}
          className="p-1.5 rounded hover:bg-editor-bg transition-colors"
          title="Fit to view"
        >
          <Maximize2 size={16} className="text-editor-muted" />
        </button>

        <button
          onClick={resetZoom}
          className="p-1.5 rounded hover:bg-editor-bg transition-colors"
          title="Reset zoom"
        >
          <RotateCcw size={16} className="text-editor-muted" />
        </button>
      </div>

      <div className="w-px h-6 bg-editor-border" />

      {/* History controls */}
      <div className="flex items-center gap-1 px-2">
        <button
          onClick={undo}
          disabled={!canUndo}
          className="p-1.5 rounded hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          title="Undo (Ctrl+Z)"
        >
          <Undo2 size={16} className="text-editor-muted" />
        </button>

        <button
          onClick={redo}
          disabled={!canRedo}
          className="p-1.5 rounded hover:bg-editor-bg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          title="Redo (Ctrl+Shift+Z)"
        >
          <Redo2 size={16} className="text-editor-muted" />
        </button>
      </div>

      <div className="w-px h-6 bg-editor-border" />

      {/* Delete control */}
      <div className="px-2">
        <button
          onClick={deleteSelected}
          disabled={!hasSelection}
          className="p-1.5 rounded hover:bg-editor-error/20 disabled:opacity-50 disabled:cursor-not-allowed transition-colors group"
          title="Delete selected (Delete)"
        >
          <Trash2
            size={16}
            className="text-editor-muted group-hover:text-editor-error transition-colors"
          />
        </button>
      </div>
    </div>
  );
});

export default CanvasControls;
