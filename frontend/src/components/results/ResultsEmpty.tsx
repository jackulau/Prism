import { FileSearch, Play, Sparkles } from 'lucide-react';

interface ResultsEmptyProps {
  hasFilters?: boolean;
  onClearFilters?: () => void;
  onRunAgents?: () => void;
}

export function ResultsEmpty({ hasFilters, onClearFilters, onRunAgents }: ResultsEmptyProps) {
  if (hasFilters) {
    return (
      <div className="bg-editor-surface border border-editor-border rounded-lg p-12 text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 bg-editor-bg rounded-full mb-4">
          <FileSearch className="w-8 h-8 text-editor-muted" />
        </div>
        <h3 className="text-lg font-medium text-editor-text mb-2">
          No results match your filters
        </h3>
        <p className="text-editor-muted mb-6 max-w-md mx-auto">
          Try adjusting your search criteria or clearing the filters to see all results.
        </p>
        {onClearFilters && (
          <button
            onClick={onClearFilters}
            className="inline-flex items-center gap-2 px-4 py-2 text-editor-accent hover:bg-editor-accent/10 rounded-lg transition-colors"
          >
            Clear all filters
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-12 text-center">
      <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-editor-accent/20 to-purple-500/20 rounded-full mb-4">
        <Sparkles className="w-8 h-8 text-editor-accent" />
      </div>
      <h3 className="text-lg font-medium text-editor-text mb-2">
        No execution results yet
      </h3>
      <p className="text-editor-muted mb-6 max-w-md mx-auto">
        Run your first batch or swarm execution to see results appear here.
        Track progress, view token usage, and analyze agent performance.
      </p>
      {onRunAgents && (
        <button
          onClick={onRunAgents}
          className="inline-flex items-center gap-2 px-5 py-2.5 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
        >
          <Play size={18} />
          Run Agents
        </button>
      )}
    </div>
  );
}
