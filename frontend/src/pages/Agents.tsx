import { useState, useCallback, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { RefreshCw, Loader2 } from 'lucide-react';
import { AgentList, AgentStats, AgentDetail } from '../components/agents';
import type { AgentExecution } from '../components/agents';
import { useAgents, useAgentActions } from '../hooks/useAgents';
import { toast } from '../store/toastStore';

export default function Agents() {
  const { id: selectedAgentId } = useParams<{ id?: string }>();
  const navigate = useNavigate();

  const { agents, total, isLoading, error, refetch, hasMore } = useAgents({ limit: 50 });
  const { cancelAgent, retryAgent, getAgent } = useAgentActions();

  const [selectedAgent, setSelectedAgent] = useState<AgentExecution | null>(null);
  const [isCancelling, setIsCancelling] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isLoadingAgent, setIsLoadingAgent] = useState(false);

  // Load agent details when URL has an ID
  useEffect(() => {
    if (selectedAgentId && !selectedAgent) {
      setIsLoadingAgent(true);
      getAgent(selectedAgentId)
        .then(setSelectedAgent)
        .catch((err) => {
          toast.error(`Failed to load agent: ${err.message}`);
          navigate('/agents', { replace: true });
        })
        .finally(() => setIsLoadingAgent(false));
    }
  }, [selectedAgentId, selectedAgent, getAgent, navigate]);

  const handleAgentClick = useCallback((agent: AgentExecution) => {
    setSelectedAgent(agent);
    navigate(`/agents/${agent.id}`);
  }, [navigate]);

  const handleCloseDetail = useCallback(() => {
    setSelectedAgent(null);
    navigate('/agents');
  }, [navigate]);

  const handleCancel = useCallback(async (id: string) => {
    setIsCancelling(true);
    try {
      await cancelAgent(id);
      toast.success('Agent cancelled');
      refetch();
      handleCloseDetail();
    } catch (err) {
      toast.error(`Failed to cancel agent: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setIsCancelling(false);
    }
  }, [cancelAgent, refetch, handleCloseDetail]);

  const handleRetry = useCallback(async (id: string) => {
    try {
      await retryAgent(id);
      toast.success('Agent requeued for retry');
      refetch();
      handleCloseDetail();
    } catch (err) {
      toast.error(`Failed to retry agent: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  }, [retryAgent, refetch, handleCloseDetail]);

  const handleDelete = useCallback(async (id: string) => {
    setIsDeleting(true);
    try {
      await cancelAgent(id);
      toast.success('Agent deleted');
      refetch();
      handleCloseDetail();
    } catch (err) {
      toast.error(`Failed to delete agent: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setIsDeleting(false);
    }
  }, [cancelAgent, refetch, handleCloseDetail]);

  // Convert AgentExecution to the format expected by AgentDetail
  const normalizeForDetail = (agent: AgentExecution): AgentExecution => ({
    ...agent,
    messages: agent.messages || [],
    toolCalls: agent.toolCalls || [],
  });

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Agents</h1>
            <p className="text-editor-muted">
              Monitor and manage your AI agent executions
            </p>
          </div>
          <button
            onClick={refetch}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors disabled:opacity-50"
          >
            {isLoading ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <RefreshCw size={16} />
            )}
            Refresh
          </button>
        </div>

        {/* Stats */}
        <AgentStats agents={agents} isLoading={isLoading} />

        {/* Agent List */}
        <AgentList
          agents={agents}
          isLoading={isLoading}
          error={error}
          onRefresh={refetch}
          onAgentClick={handleAgentClick}
          showFilters={true}
          pageSize={12}
          title="Recent Executions"
          emptyMessage="No agent executions"
          emptyDescription="Run an agent to see executions here"
        />

        {/* Load More indicator */}
        {hasMore && !isLoading && (
          <p className="text-center text-sm text-editor-muted">
            Showing {agents.length} of {total} agents
          </p>
        )}

        {/* Loading overlay for direct URL access */}
        {isLoadingAgent && (
          <div className="fixed inset-0 bg-black/50 z-40 flex items-center justify-center">
            <div className="bg-editor-surface border border-editor-border rounded-lg p-6 flex items-center gap-3">
              <Loader2 size={24} className="animate-spin text-editor-accent" />
              <span className="text-editor-text">Loading agent details...</span>
            </div>
          </div>
        )}

        {/* Agent Detail Modal */}
        {selectedAgent && (
          <AgentDetail
            agent={normalizeForDetail(selectedAgent)}
            onClose={handleCloseDetail}
            onCancel={handleCancel}
            onRetry={handleRetry}
            onDelete={handleDelete}
            isCancelling={isCancelling}
            isDeleting={isDeleting}
          />
        )}
      </div>
    </div>
  );
}
