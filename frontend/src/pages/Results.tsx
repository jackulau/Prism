import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, AlertCircle } from 'lucide-react';
import { ResultsHeader } from '../components/results/ResultsHeader';
import { ResultsMetricsSummary } from '../components/results/ResultsMetricsSummary';
import { ResultsList } from '../components/results/ResultsList';
import { ExecutionDetailPanel } from '../components/results/ExecutionDetailPanel';
import type { ExecutionPanelData } from '../components/results/ExecutionDetailPanel';
import type { ExecutionResult } from '../components/results/types';
import type { AggregatedResults, ExecutionMetrics } from '../types/results';
import {
  useResultsStore,
  type BatchResultSummary,
  type SwarmResultSummary,
} from '../store/resultsStore';
import { toast } from '../store/toastStore';

type ViewTab = 'all' | 'batch' | 'swarm';

interface TabButtonProps {
  tab: ViewTab;
  activeTab: ViewTab;
  label: string;
  count: number;
  onClick: (tab: ViewTab) => void;
}

function TabButton({ tab, activeTab, label, count, onClick }: TabButtonProps) {
  const isActive = tab === activeTab;
  return (
    <button
      onClick={() => onClick(tab)}
      className={`px-4 py-2 text-sm font-medium transition-colors relative ${
        isActive
          ? 'text-editor-accent'
          : 'text-editor-muted hover:text-editor-text'
      }`}
    >
      <span>{label}</span>
      <span
        className={`ml-2 px-2 py-0.5 text-xs rounded-full ${
          isActive
            ? 'bg-editor-accent/20 text-editor-accent'
            : 'bg-editor-surface text-editor-muted'
        }`}
      >
        {count}
      </span>
      {isActive && (
        <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-editor-accent" />
      )}
    </button>
  );
}

function mapBatchToExecutionResult(batch: BatchResultSummary): ExecutionResult {
  return {
    id: batch.id,
    name: batch.name,
    type: 'batch',
    status: batch.status,
    startedAt: batch.startedAt,
    completedAt: batch.completedAt,
    durationMs: batch.completedAt && batch.startedAt
      ? new Date(batch.completedAt).getTime() - new Date(batch.startedAt).getTime()
      : undefined,
    taskCount: batch.totalExecutions,
    agentCount: batch.totalExecutions,
    totalTokens: batch.totalTokens,
    cost: batch.totalCost,
  };
}

function mapSwarmToExecutionResult(swarm: SwarmResultSummary): ExecutionResult {
  return {
    id: swarm.id,
    name: swarm.name,
    type: 'swarm',
    status: swarm.status,
    startedAt: swarm.startedAt,
    completedAt: swarm.completedAt,
    durationMs: swarm.completedAt && swarm.startedAt
      ? new Date(swarm.completedAt).getTime() - new Date(swarm.startedAt).getTime()
      : undefined,
    taskCount: 1,
    agentCount: swarm.agentCount,
    totalTokens: swarm.totalTokens,
    cost: swarm.totalCost,
  };
}

export default function Results() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<ViewTab>('all');
  const [selectedTimeRange, setSelectedTimeRange] = useState('7d');
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [panelData, setPanelData] = useState<ExecutionPanelData | null>(null);

  const {
    batchResults,
    swarmResults,
    selectedExecution,
    aggregatedMetrics,
    isLoading,
    error,
    fetchBatchResults,
    fetchSwarmResults,
    fetchAggregatedMetrics,
    fetchExecutionDetails,
    selectExecution,
    setError,
    executionDetails,
  } = useResultsStore();

  // Fetch data on mount
  useEffect(() => {
    const fetchData = async () => {
      await Promise.all([
        fetchBatchResults(),
        fetchSwarmResults(),
        fetchAggregatedMetrics(),
      ]);
    };
    fetchData();
  }, [fetchBatchResults, fetchSwarmResults, fetchAggregatedMetrics]);

  // Map results to common format
  const allResults = useMemo<ExecutionResult[]>(() => {
    const batch = batchResults.map(mapBatchToExecutionResult);
    const swarm = swarmResults.map(mapSwarmToExecutionResult);
    return [...batch, ...swarm].sort(
      (a, b) => b.startedAt.getTime() - a.startedAt.getTime()
    );
  }, [batchResults, swarmResults]);

  const batchExecutions = useMemo(() => {
    return batchResults.map(mapBatchToExecutionResult);
  }, [batchResults]);

  const swarmExecutions = useMemo(() => {
    return swarmResults.map(mapSwarmToExecutionResult);
  }, [swarmResults]);

  const displayResults = useMemo(() => {
    switch (activeTab) {
      case 'batch':
        return batchExecutions;
      case 'swarm':
        return swarmExecutions;
      default:
        return allResults;
    }
  }, [activeTab, allResults, batchExecutions, swarmExecutions]);

  // Convert aggregated metrics to the format expected by ResultsMetricsSummary
  const metricsData = useMemo<AggregatedResults | null>(() => {
    if (!aggregatedMetrics) return null;

    const metrics: ExecutionMetrics = {
      promptTokens: aggregatedMetrics.totalTokens.input,
      completionTokens: aggregatedMetrics.totalTokens.output,
      totalTokens: aggregatedMetrics.totalTokens.total,
      inputCost: aggregatedMetrics.totalCost * 0.4,
      outputCost: aggregatedMetrics.totalCost * 0.6,
      totalCost: aggregatedMetrics.totalCost,
      currency: 'USD',
      durationMs: aggregatedMetrics.averageDuration,
      tokensPerSecond: aggregatedMetrics.averageDuration > 0
        ? (aggregatedMetrics.averageTokensPerExecution / aggregatedMetrics.averageDuration) * 1000
        : undefined,
    };

    return {
      type: 'batch',
      batch: {
        id: 'aggregated',
        status: 'completed',
        totalCount: aggregatedMetrics.totalExecutions,
        completedCount: aggregatedMetrics.successfulExecutions,
        failedCount: aggregatedMetrics.failedExecutions,
        pendingCount: 0,
        executions: [],
        aggregatedMetrics: metrics,
        createdAt: new Date(),
      },
      status: 'completed',
      totalMetrics: metrics,
      createdAt: new Date(),
    };
  }, [aggregatedMetrics]);

  const handleRefresh = useCallback(async () => {
    await Promise.all([
      fetchBatchResults(),
      fetchSwarmResults(),
      fetchAggregatedMetrics(),
    ]);
    toast.success('Results refreshed');
  }, [fetchBatchResults, fetchSwarmResults, fetchAggregatedMetrics]);

  const handleExport = useCallback((format: 'csv' | 'json') => {
    const data = displayResults;

    if (format === 'json') {
      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: 'application/json',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `results-${new Date().toISOString().slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } else {
      const headers = ['ID', 'Name', 'Type', 'Status', 'Started At', 'Duration', 'Tokens', 'Cost'];
      const rows = data.map((r) => [
        r.id,
        r.name || '',
        r.type,
        r.status,
        r.startedAt.toISOString(),
        r.durationMs ? `${r.durationMs}ms` : '',
        r.totalTokens.toString(),
        r.cost?.toFixed(4) || '',
      ]);
      const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `results-${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    }
    toast.success(`Exported as ${format.toUpperCase()}`);
  }, [displayResults]);

  const handleTimeRangeChange = useCallback((start: Date, end: Date) => {
    const days = Math.round((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
    if (days <= 1) setSelectedTimeRange('24h');
    else if (days <= 7) setSelectedTimeRange('7d');
    else if (days <= 30) setSelectedTimeRange('30d');
    else setSelectedTimeRange('90d');

    fetchBatchResults({ dateRange: { start, end } });
    fetchSwarmResults({ dateRange: { start, end } });
    fetchAggregatedMetrics({ start, end });
  }, [fetchBatchResults, fetchSwarmResults, fetchAggregatedMetrics]);

  const handleSelectExecution = useCallback(async (result: ExecutionResult) => {
    selectExecution(result.id);
    await fetchExecutionDetails(result.id);

    const details = executionDetails.get(result.id);
    if (details) {
      setPanelData({
        execution: {
          id: details.id,
          type: details.type,
          status: details.status,
          name: result.name,
          startedAt: details.startedAt ?? null,
          completedAt: details.completedAt ?? null,
        },
        agentResults: [],
        timelineItems: [
          {
            id: '1',
            label: 'Execution Started',
            status: 'completed',
            startedAt: details.startedAt,
          },
          ...(details.completedAt ? [{
            id: '2',
            label: details.status === 'failed' ? 'Execution Failed' : 'Execution Completed',
            status: details.status === 'failed' ? 'failed' as const : 'completed' as const,
            startedAt: details.completedAt,
            completedAt: details.completedAt,
          }] : []),
        ],
        metrics: {
          promptTokens: details.tokens.input,
          completionTokens: details.tokens.output,
          totalTokens: details.tokens.total,
          inputCost: details.cost * 0.4,
          outputCost: details.cost * 0.6,
          totalCost: details.cost,
          successCount: details.status === 'completed' ? 1 : 0,
          failureCount: details.status === 'failed' ? 1 : 0,
          totalTasks: 1,
          averageDuration: details.duration,
        },
        logs: details.error ? [`Error: ${details.error}`] : undefined,
      });
      setIsPanelOpen(true);
    }
  }, [selectExecution, fetchExecutionDetails, executionDetails]);

  const handleClosePanel = useCallback(() => {
    setIsPanelOpen(false);
    selectExecution(null);
  }, [selectExecution]);

  const handleRunAgents = useCallback(() => {
    navigate('/workspace');
  }, [navigate]);

  if (error) {
    return (
      <div className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto p-6">
          <div className="bg-editor-error/10 border border-editor-error/30 rounded-lg p-6">
            <div className="flex items-center gap-3">
              <AlertCircle className="w-6 h-6 text-editor-error" />
              <div>
                <h3 className="font-medium text-editor-error">Error loading results</h3>
                <p className="text-sm text-editor-muted mt-1">{error}</p>
                <button
                  onClick={() => {
                    setError(null);
                    handleRefresh();
                  }}
                  className="mt-3 text-sm text-editor-accent hover:underline"
                >
                  Try again
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <ResultsHeader
          onRefresh={handleRefresh}
          onExport={handleExport}
          onTimeRangeChange={handleTimeRangeChange}
          isLoading={isLoading}
          selectedTimeRange={selectedTimeRange}
        />

        {/* Metrics Summary */}
        <ResultsMetricsSummary
          data={metricsData}
          loading={isLoading && !metricsData}
        />

        {/* Tabs */}
        <div className="border-b border-editor-border">
          <div className="flex gap-1">
            <TabButton
              tab="all"
              activeTab={activeTab}
              label="All Results"
              count={allResults.length}
              onClick={setActiveTab}
            />
            <TabButton
              tab="batch"
              activeTab={activeTab}
              label="Batch"
              count={batchExecutions.length}
              onClick={setActiveTab}
            />
            <TabButton
              tab="swarm"
              activeTab={activeTab}
              label="Swarm"
              count={swarmExecutions.length}
              onClick={setActiveTab}
            />
          </div>
        </div>

        {/* Results List */}
        {isLoading && displayResults.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-editor-muted" />
            <span className="ml-3 text-editor-muted">Loading results...</span>
          </div>
        ) : (
          <ResultsList
            results={displayResults}
            isLoading={isLoading}
            selectedId={selectedExecution || undefined}
            onSelect={handleSelectExecution}
            onRunAgents={handleRunAgents}
          />
        )}
      </div>

      {/* Execution Detail Panel */}
      <ExecutionDetailPanel
        isOpen={isPanelOpen}
        onClose={handleClosePanel}
        data={panelData}
      />
    </div>
  );
}
