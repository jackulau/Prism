import type {
  Task,
  AgentConfig,
  Execution,
  Swarm,
  SwarmStrategy,
  AgentRoleConfig,
  AgentResult,
  SwarmAgent,
  SwarmResult,
  ManagerStats,
} from '../routers/workers/schemas.js';

// In-memory storage for executions and swarms
// In production, this would be backed by a database or the Go backend
const executions = new Map<string, Execution>();
const swarms = new Map<string, Swarm>();

// Counter for generating IDs
let executionCounter = 0;
let swarmCounter = 0;

function generateExecutionId(): string {
  return `exec_${++executionCounter}_${Date.now()}`;
}

function generateSwarmId(): string {
  return `swarm_${++swarmCounter}_${Date.now()}`;
}

// Simulated delay for async operations
async function simulateProcessing(minMs: number, maxMs: number): Promise<void> {
  const delay = Math.floor(Math.random() * (maxMs - minMs + 1)) + minMs;
  await new Promise((resolve) => setTimeout(resolve, delay));
}

export const workerService = {
  /**
   * Run a single task with an agent
   */
  async runTask(
    _userId: string,
    task: Task,
    _config: AgentConfig
  ): Promise<Execution> {
    const executionId = generateExecutionId();
    const now = new Date();

    const execution: Execution = {
      id: executionId,
      type: 'single',
      status: 'running',
      tasks: [task],
      startedAt: now,
    };

    executions.set(executionId, execution);

    // Simulate async task processing
    // In production, this would call the Go backend or execute the agent directly
    setImmediate(async () => {
      await simulateProcessing(1000, 3000);

      const completedExecution = executions.get(executionId);
      if (completedExecution && completedExecution.status === 'running') {
        const result: AgentResult = {
          agentId: `agent_${executionId}`,
          taskId: task.id,
          success: true,
          output: `Processed task: ${task.prompt.substring(0, 100)}...`,
          durationMs: Date.now() - now.getTime(),
          completedAt: new Date(),
        };

        completedExecution.status = 'completed';
        completedExecution.results = [result];
        completedExecution.completedAt = new Date();
        executions.set(executionId, completedExecution);
      }
    });

    return execution;
  },

  /**
   * Run multiple tasks in parallel
   */
  async runParallel(
    _userId: string,
    tasks: Task[],
    _config: AgentConfig
  ): Promise<Execution> {
    const executionId = generateExecutionId();
    const now = new Date();

    const execution: Execution = {
      id: executionId,
      type: 'parallel',
      status: 'running',
      tasks,
      startedAt: now,
    };

    executions.set(executionId, execution);

    // Simulate async parallel task processing
    setImmediate(async () => {
      await simulateProcessing(2000, 5000);

      const completedExecution = executions.get(executionId);
      if (completedExecution && completedExecution.status === 'running') {
        const results: AgentResult[] = tasks.map((task, index) => ({
          agentId: `agent_${executionId}_${index}`,
          taskId: task.id,
          success: true,
          output: `Processed parallel task ${index + 1}: ${task.prompt.substring(0, 50)}...`,
          durationMs: Date.now() - now.getTime(),
          completedAt: new Date(),
        }));

        completedExecution.status = 'completed';
        completedExecution.results = results;
        completedExecution.completedAt = new Date();
        executions.set(executionId, completedExecution);
      }
    });

    return execution;
  },

  /**
   * Run multiple tasks sequentially
   */
  async runSequential(
    _userId: string,
    tasks: Task[],
    _config: AgentConfig
  ): Promise<Execution> {
    const executionId = generateExecutionId();
    const now = new Date();

    const execution: Execution = {
      id: executionId,
      type: 'sequential',
      status: 'running',
      tasks,
      startedAt: now,
    };

    executions.set(executionId, execution);

    // Simulate async sequential task processing
    setImmediate(async () => {
      const results: AgentResult[] = [];

      for (let i = 0; i < tasks.length; i++) {
        const task = tasks[i]!;
        await simulateProcessing(500, 1500);

        const completedExecution = executions.get(executionId);
        if (!completedExecution || completedExecution.status !== 'running') {
          break;
        }

        results.push({
          agentId: `agent_${executionId}_${i}`,
          taskId: task.id,
          success: true,
          output: `Processed sequential task ${i + 1}: ${task.prompt.substring(0, 50)}...`,
          durationMs: Date.now() - now.getTime(),
          completedAt: new Date(),
        });
      }

      const finalExecution = executions.get(executionId);
      if (finalExecution && finalExecution.status === 'running') {
        finalExecution.status = 'completed';
        finalExecution.results = results;
        finalExecution.completedAt = new Date();
        executions.set(executionId, finalExecution);
      }
    });

    return execution;
  },

  /**
   * Run a multi-agent swarm
   */
  async runSwarm(
    _userId: string,
    prompt: string,
    name: string | undefined,
    strategy: SwarmStrategy,
    roles: AgentRoleConfig[],
    _config: AgentConfig,
    _timeout?: number
  ): Promise<Swarm> {
    const swarmId = generateSwarmId();
    const now = new Date();

    // Create swarm agents from role configs
    const agents: SwarmAgent[] = [];
    let agentCounter = 0;

    for (const roleConfig of roles) {
      const count = roleConfig.count ?? 1;
      for (let i = 0; i < count; i++) {
        agents.push({
          id: `agent_${swarmId}_${++agentCounter}`,
          role: roleConfig.role,
          status: 'idle',
        });
      }
    }

    const swarm: Swarm = {
      id: swarmId,
      name: name ?? `Swarm ${swarmId}`,
      strategy,
      status: 'running',
      agents,
      messages: [],
      results: [],
      createdAt: now,
      startedAt: now,
    };

    swarms.set(swarmId, swarm);

    // Simulate async swarm processing
    setImmediate(async () => {
      // Update agents to running
      const runningSwarm = swarms.get(swarmId);
      if (runningSwarm) {
        runningSwarm.agents = runningSwarm.agents.map((a) => ({
          ...a,
          status: 'running' as const,
          startedAt: new Date(),
        }));
        swarms.set(swarmId, runningSwarm);
      }

      await simulateProcessing(3000, 8000);

      const completedSwarm = swarms.get(swarmId);
      if (completedSwarm && completedSwarm.status === 'running') {
        // Generate results for each agent
        const results: SwarmResult[] = completedSwarm.agents.map((agent) => ({
          agentId: agent.id,
          role: agent.role,
          output: `Output from ${agent.role} agent for: ${prompt.substring(0, 50)}...`,
          success: true,
          durationMs: Date.now() - now.getTime(),
        }));

        // Update agent statuses
        const completedAgents: SwarmAgent[] = completedSwarm.agents.map(
          (a) => ({
            ...a,
            status: 'completed' as const,
            output: results.find((r) => r.agentId === a.id)?.output,
            completedAt: new Date(),
          })
        );

        completedSwarm.status = 'completed';
        completedSwarm.agents = completedAgents;
        completedSwarm.results = results;
        completedSwarm.finalOutput = `Synthesized output from ${completedAgents.length} agents using ${strategy} strategy.`;
        completedSwarm.completedAt = new Date();
        swarms.set(swarmId, completedSwarm);
      }
    });

    return swarm;
  },

  /**
   * Get execution by ID
   */
  async getExecution(executionId: string): Promise<Execution | null> {
    return executions.get(executionId) ?? null;
  },

  /**
   * List executions for a user
   */
  async listExecutions(_userId: string): Promise<Execution[]> {
    // In production, filter by userId
    return Array.from(executions.values());
  },

  /**
   * Cancel a running execution
   */
  async cancelExecution(executionId: string): Promise<void> {
    const execution = executions.get(executionId);
    if (!execution) {
      throw new Error(`Execution not found: ${executionId}`);
    }

    if (
      execution.status !== 'running' &&
      execution.status !== 'pending'
    ) {
      throw new Error(`Cannot cancel execution with status: ${execution.status}`);
    }

    execution.status = 'cancelled';
    execution.completedAt = new Date();
    executions.set(executionId, execution);
  },

  /**
   * Get swarm by ID
   */
  async getSwarm(swarmId: string): Promise<Swarm | null> {
    return swarms.get(swarmId) ?? null;
  },

  /**
   * List swarms for a user
   */
  async listSwarms(_userId: string): Promise<Swarm[]> {
    // In production, filter by userId
    return Array.from(swarms.values());
  },

  /**
   * Cancel a running swarm
   */
  async cancelSwarm(swarmId: string): Promise<void> {
    const swarm = swarms.get(swarmId);
    if (!swarm) {
      throw new Error(`Swarm not found: ${swarmId}`);
    }

    if (swarm.status !== 'running' && swarm.status !== 'pending') {
      throw new Error(`Cannot cancel swarm with status: ${swarm.status}`);
    }

    swarm.status = 'cancelled';
    swarm.agents = swarm.agents.map((a) =>
      a.status === 'running' ? { ...a, status: 'cancelled' as const } : a
    );
    swarm.completedAt = new Date();
    swarms.set(swarmId, swarm);
  },

  /**
   * Get manager statistics
   */
  async getStats(_userId: string): Promise<ManagerStats> {
    const allExecutions = Array.from(executions.values());
    const allSwarms = Array.from(swarms.values());

    return {
      totalExecutions: allExecutions.length,
      runningExecutions: allExecutions.filter((e) => e.status === 'running')
        .length,
      completedExecutions: allExecutions.filter((e) => e.status === 'completed')
        .length,
      failedExecutions: allExecutions.filter((e) => e.status === 'failed')
        .length,
      cancelledExecutions: allExecutions.filter((e) => e.status === 'cancelled')
        .length,
      activeAgents: allSwarms
        .flatMap((s) => s.agents)
        .filter((a) => a.status === 'running').length,
      queuedTasks: allExecutions.filter((e) => e.status === 'pending').length,
      registeredConfigs: 0, // Would be tracked separately
      activeSwarms: allSwarms.filter((s) => s.status === 'running').length,
    };
  },

  /**
   * Clean up old executions and swarms
   */
  async cleanup(maxAgeMs: number = 24 * 60 * 60 * 1000): Promise<number> {
    const cutoff = new Date(Date.now() - maxAgeMs);
    let cleaned = 0;

    for (const [id, execution] of executions.entries()) {
      if (execution.completedAt && execution.completedAt < cutoff) {
        executions.delete(id);
        cleaned++;
      }
    }

    for (const [id, swarm] of swarms.entries()) {
      if (swarm.completedAt && swarm.completedAt < cutoff) {
        swarms.delete(id);
        cleaned++;
      }
    }

    return cleaned;
  },
};
