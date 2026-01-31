import { useState } from 'react';
import { X, Cpu, Wrench, Zap, Copy, Check, Trash2, Edit2, Loader2 } from 'lucide-react';
import type { Tool } from '../../types/tools';
import { getToolType } from '../../types/tools';
import { apiService } from '../../services/api';
import { toast } from '../../store/toastStore';

interface ToolDetailModalProps {
  tool: Tool;
  onClose: () => void;
  onDelete?: () => void;
  onEdit?: () => void;
}

export function ToolDetailModal({ tool, onClose, onDelete, onEdit }: ToolDetailModalProps) {
  const [copied, setCopied] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const toolType = getToolType(tool);
  const TypeIcon = toolType === 'model' ? Cpu : toolType === 'custom' ? Wrench : Zap;
  const typeLabel = toolType === 'model' ? 'Model' : toolType === 'custom' ? 'Custom Tool' : 'Builtin Tool';
  const typeColorClass =
    toolType === 'model'
      ? 'bg-purple-500/10 text-purple-400'
      : toolType === 'custom'
        ? 'bg-blue-500/10 text-blue-400'
        : 'bg-green-500/10 text-green-400';

  const isCustom = toolType === 'custom';

  const handleCopySlug = async () => {
    await navigator.clipboard.writeText(tool.slug_name);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDelete = async () => {
    if (!showDeleteConfirm) {
      setShowDeleteConfirm(true);
      return;
    }

    setIsDeleting(true);
    const result = await apiService.deleteTool(tool.id);
    setIsDeleting(false);

    if (result.error) {
      toast.error(result.error);
    } else {
      toast.success('Tool deleted successfully');
      onDelete?.();
      onClose();
    }
  };

  // Parse parameters schema for display
  let parsedSchema = null;
  if (tool.parameters_schema) {
    try {
      parsedSchema = JSON.parse(tool.parameters_schema);
    } catch {
      // Invalid JSON, will show raw string
    }
  }

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[560px] max-w-[90vw] max-h-[85vh] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50 flex flex-col">
        {/* Header */}
        <div className="flex items-start justify-between px-6 py-4 border-b border-editor-border">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 flex items-center justify-center bg-editor-surface rounded-lg flex-shrink-0">
              <TypeIcon size={24} className="text-editor-muted" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-editor-text">{tool.display_name}</h2>
              <div className="flex items-center gap-2 mt-1">
                <span className={`px-2 py-0.5 rounded-full text-xs ${typeColorClass}`}>
                  {typeLabel}
                </span>
                {tool.provider_id && (
                  <span className="px-2 py-0.5 bg-editor-surface border border-editor-border rounded text-xs text-editor-muted">
                    {tool.provider_id}
                  </span>
                )}
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Slug */}
          <div>
            <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-2">
              Slug Name
            </label>
            <div className="flex items-center gap-2">
              <code className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text font-mono">
                {tool.slug_name}
              </code>
              <button
                onClick={handleCopySlug}
                className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
                title="Copy slug"
              >
                {copied ? <Check size={16} className="text-editor-success" /> : <Copy size={16} />}
              </button>
            </div>
          </div>

          {/* Description */}
          {tool.description && (
            <div>
              <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-2">
                Description
              </label>
              <p className="text-sm text-editor-text">{tool.description}</p>
            </div>
          )}

          {/* Parameters Schema */}
          {tool.parameters_schema && (
            <div>
              <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-2">
                Parameters Schema
              </label>
              <pre className="px-4 py-3 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text font-mono overflow-x-auto">
                {parsedSchema ? JSON.stringify(parsedSchema, null, 2) : tool.parameters_schema}
              </pre>
            </div>
          )}

          {/* Metadata */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-1">
                Created
              </label>
              <p className="text-sm text-editor-text">
                {new Date(tool.created_at).toLocaleDateString()}
              </p>
            </div>
            <div>
              <label className="block text-xs font-medium text-editor-muted uppercase tracking-wider mb-1">
                Updated
              </label>
              <p className="text-sm text-editor-text">
                {new Date(tool.updated_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>

        {/* Footer - only show for custom tools */}
        {isCustom && (
          <div className="flex items-center justify-between px-6 py-4 border-t border-editor-border bg-editor-surface/50">
            {showDeleteConfirm ? (
              <div className="flex items-center gap-2 text-sm text-editor-error">
                <span>Delete this tool?</span>
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="px-2 py-1 text-editor-muted hover:text-editor-text transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="flex items-center gap-1 px-2 py-1 bg-editor-error text-white rounded hover:bg-editor-error/90 transition-colors disabled:opacity-50"
                >
                  {isDeleting ? <Loader2 size={14} className="animate-spin" /> : <Trash2 size={14} />}
                  Confirm
                </button>
              </div>
            ) : (
              <button
                onClick={handleDelete}
                className="flex items-center gap-2 px-4 py-2 text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
              >
                <Trash2 size={16} />
                Delete
              </button>
            )}
            <button
              onClick={onEdit}
              className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              <Edit2 size={16} />
              Edit Tool
            </button>
          </div>
        )}
      </div>
    </>
  );
}
