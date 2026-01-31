import { Cpu, Wrench, Zap } from 'lucide-react';
import type { Tool } from '../../types/tools';
import { getToolType } from '../../types/tools';

interface ToolCardProps {
  tool: Tool;
  viewMode: 'grid' | 'list';
  onClick: () => void;
}

export function ToolCard({ tool, viewMode, onClick }: ToolCardProps) {
  const toolType = getToolType(tool);

  const TypeIcon = toolType === 'model' ? Cpu : toolType === 'custom' ? Wrench : Zap;

  const typeLabel = toolType === 'model' ? 'Model' : toolType === 'custom' ? 'Custom' : 'Builtin';
  const typeColorClass =
    toolType === 'model'
      ? 'bg-purple-500/10 text-purple-400'
      : toolType === 'custom'
        ? 'bg-blue-500/10 text-blue-400'
        : 'bg-green-500/10 text-green-400';

  if (viewMode === 'list') {
    return (
      <button
        onClick={onClick}
        className="w-full flex items-center gap-4 p-4 bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors text-left"
      >
        <div className="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-editor-bg rounded-lg">
          <TypeIcon size={20} className="text-editor-muted" />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-editor-text truncate">{tool.display_name}</h3>
            <span
              className={`flex-shrink-0 px-2 py-0.5 rounded-full text-xs ${typeColorClass}`}
            >
              {typeLabel}
            </span>
          </div>
          <p className="text-sm text-editor-muted truncate">
            {tool.description || tool.slug_name}
          </p>
        </div>

        {tool.provider_id && (
          <span className="flex-shrink-0 px-2 py-1 bg-editor-bg border border-editor-border rounded text-xs text-editor-muted">
            {tool.provider_id}
          </span>
        )}
      </button>
    );
  }

  return (
    <button
      onClick={onClick}
      className="flex flex-col p-5 bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors text-left h-full"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="w-12 h-12 flex items-center justify-center bg-editor-bg rounded-lg">
          <TypeIcon size={24} className="text-editor-muted" />
        </div>
        <span className={`px-2 py-1 rounded-full text-xs ${typeColorClass}`}>{typeLabel}</span>
      </div>

      <h3 className="font-medium text-editor-text mb-1 line-clamp-1">{tool.display_name}</h3>

      <p className="text-sm text-editor-muted mb-3 line-clamp-2 flex-1">
        {tool.description || `Tool: ${tool.slug_name}`}
      </p>

      {tool.provider_id && (
        <span className="inline-block px-2 py-1 bg-editor-bg border border-editor-border rounded text-xs text-editor-muted w-fit">
          {tool.provider_id}
        </span>
      )}
    </button>
  );
}

export function ToolCardSkeleton({ viewMode }: { viewMode: 'grid' | 'list' }) {
  if (viewMode === 'list') {
    return (
      <div className="flex items-center gap-4 p-4 bg-editor-surface border border-editor-border rounded-lg animate-pulse">
        <div className="w-10 h-10 bg-editor-border rounded-lg" />
        <div className="flex-1 space-y-2">
          <div className="h-4 bg-editor-border rounded w-1/3" />
          <div className="h-3 bg-editor-border rounded w-2/3" />
        </div>
      </div>
    );
  }

  return (
    <div className="p-5 bg-editor-surface border border-editor-border rounded-lg animate-pulse">
      <div className="flex items-start justify-between mb-3">
        <div className="w-12 h-12 bg-editor-border rounded-lg" />
        <div className="w-16 h-5 bg-editor-border rounded-full" />
      </div>
      <div className="h-4 bg-editor-border rounded w-2/3 mb-2" />
      <div className="h-3 bg-editor-border rounded w-full mb-1" />
      <div className="h-3 bg-editor-border rounded w-3/4" />
    </div>
  );
}
