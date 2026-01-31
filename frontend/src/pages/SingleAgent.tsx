import { useCallback } from 'react';
import { Bot } from 'lucide-react';
import { SingleAgentForm } from '../components/agents/SingleAgentForm';
import { AgentExecutionPanel } from '../components/agents/AgentExecutionPanel';
import { RecentExecutions } from '../components/agents/RecentExecutions';
import { useAgentStore } from '../store/agentStore';

export default function SingleAgent() {
  const {
    executionStatus,
    config,
    startExecution,
    addOutput,
    completeExecution,
    stopExecution,
    setExecutionError,
  } = useAgentStore();

  const isRunning = executionStatus === 'running';

  const handleSubmit = useCallback(async () => {
    const executionId = startExecution();
    console.log('Starting agent execution:', executionId);

    // Add initial output showing configuration
    addOutput({
      type: 'text',
      content: `Starting agent "${config.name}" with model ${config.provider}/${config.model}...`,
    });

    try {
      // TODO: Integrate with actual agent execution API
      // For now, simulate a basic execution
      addOutput({
        type: 'thinking',
        content: 'Analyzing the request and preparing response...',
      });

      // Simulate processing time
      await new Promise((resolve) => setTimeout(resolve, 1500));

      // Simulate a response
      addOutput({
        type: 'text',
        content: `Hello! I'm ${config.name}, ready to assist you. This is a placeholder response. The actual agent execution will be connected to the backend API.`,
      });

      completeExecution();
    } catch (error) {
      setExecutionError(error instanceof Error ? error.message : 'Unknown error occurred');
    }
  }, [config, startExecution, addOutput, completeExecution, setExecutionError]);

  const handleStop = useCallback(() => {
    // TODO: Send stop signal to backend
    addOutput({
      type: 'text',
      content: 'Execution stopped by user.',
    });
    stopExecution();
  }, [addOutput, stopExecution]);

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-3 px-6 py-4 border-b border-editor-border bg-editor-surface/30">
        <div className="w-10 h-10 rounded-lg bg-editor-accent/10 flex items-center justify-center">
          <Bot size={20} className="text-editor-accent" />
        </div>
        <div>
          <h1 className="text-lg font-semibold text-editor-text">Single Agent</h1>
          <p className="text-sm text-editor-muted">Configure and run a single AI agent</p>
        </div>
      </div>

      {/* Main Content - Two Panel Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel - Configuration */}
        <div className="w-96 flex-shrink-0 border-r border-editor-border flex flex-col overflow-hidden">
          <div className="flex-1 overflow-y-auto p-6">
            <SingleAgentForm onSubmit={handleSubmit} disabled={isRunning} />
          </div>

          {/* Recent Executions */}
          <RecentExecutions />
        </div>

        {/* Right Panel - Execution Output */}
        <div className="flex-1 flex flex-col overflow-hidden bg-editor-bg">
          <AgentExecutionPanel onStop={isRunning ? handleStop : undefined} />
        </div>
      </div>
    </div>
  );
}
