import { Clock, Zap, Trash2, MessageSquare } from 'lucide-react';

export interface TestResult {
  id: string;
  provider: string;
  model: string;
  prompt: string;
  response: string;
  latencyMs: number;
  tokensUsed: {
    input: number;
    output: number;
  };
  timestamp: Date;
}

interface ProviderTestHistoryProps {
  results: TestResult[];
  onClear: () => void;
}

export function ProviderTestHistory({ results, onClear }: ProviderTestHistoryProps) {
  if (results.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 border-t border-editor-border pt-3">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-xs font-medium text-editor-muted flex items-center gap-1">
          <MessageSquare className="w-3 h-3" />
          Recent Tests ({results.length})
        </h4>
        <button
          onClick={onClear}
          className="text-xs text-editor-muted hover:text-red-400 flex items-center gap-1 transition-colors"
        >
          <Trash2 className="w-3 h-3" />
          Clear
        </button>
      </div>

      <div className="space-y-2">
        {results.map((result) => (
          <TestResultCard key={result.id} result={result} />
        ))}
      </div>
    </div>
  );
}

function TestResultCard({ result }: { result: TestResult }) {
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="p-2 bg-editor-surface rounded-lg border border-editor-border">
      {/* Header with model and metrics */}
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-editor-accent truncate max-w-[50%]">
          {result.model}
        </span>
        <div className="flex items-center gap-3 text-xs text-editor-muted">
          <span className="flex items-center gap-1" title="Latency">
            <Clock className="w-3 h-3" />
            {result.latencyMs}ms
          </span>
          <span className="flex items-center gap-1" title="Tokens (in/out)">
            <Zap className="w-3 h-3" />
            {result.tokensUsed.input}/{result.tokensUsed.output}
          </span>
          <span>{formatTime(result.timestamp)}</span>
        </div>
      </div>

      {/* Prompt */}
      <div className="mb-2">
        <span className="text-[10px] uppercase tracking-wider text-editor-muted">Prompt</span>
        <p className="text-xs text-editor-muted truncate" title={result.prompt}>
          {result.prompt}
        </p>
      </div>

      {/* Response */}
      <div>
        <span className="text-[10px] uppercase tracking-wider text-editor-muted">Response</span>
        <p className="text-xs text-editor-text whitespace-pre-wrap break-words max-h-24 overflow-y-auto">
          {result.response}
        </p>
      </div>
    </div>
  );
}

export default ProviderTestHistory;
