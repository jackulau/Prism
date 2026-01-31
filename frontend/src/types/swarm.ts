/**
 * TypeScript types for swarm API communication.
 * These types align with the backend orchestrator types in /backend/internal/agent/orchestrator.go
 */

// ============================================================================
// Enums and Literal Types
// ============================================================================

/** How agents in a swarm coordinate their work */
export type SwarmStrategy =
  | 'parallel'    // All agents work independently in parallel
  | 'pipeline'    // Agents work in sequence, each building on previous output
  | 'debate'      // Agents debate/critique each other's outputs
  | 'consensus'   // Agents work to reach consensus on a solution
  | 'map_reduce'  // Split task, parallel execution, then combine results
  | 'specialist'; // Route tasks to specialized agents

/** Specialization role for an agent */
export type AgentRole =
  | 'general'     // General-purpose assistant
  | 'planner'     // Task planning and breakdown
  | 'coder'       // Software development
  | 'reviewer'    // Code review and quality checks
  | 'researcher'  // Research and information gathering
  | 'writer'      // Technical writing and documentation
  | 'analyst'     // Data analysis and insights
  | 'debugger'    // Debugging and issue resolution
  | 'tester'      // Testing and quality assurance
  | 'synthesizer'; // Combining multiple outputs

/** Status of a swarm execution */
export type SwarmStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled';

/** Status of an individual agent within a swarm */
export type AgentStatus =
  | 'pending'
  | 'idle'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled';

/** Types of events emitted by a swarm */
export type SwarmEventType =
  | 'swarm_started'
  | 'agent_started'
  | 'agent_output'
  | 'agent_completed'
  | 'agent_failed'
  | 'message'
  | 'synthesizing'
  | 'swarm_completed'
  | 'swarm_failed'
  | 'swarm_cancelled'
  | 'progress';

/** Types of inter-agent messages */
export type SwarmMessageType = 'output' | 'request' | 'feedback' | 'critique';

// ============================================================================
// Configuration Types
// ============================================================================

/** Configuration for an individual agent */
export interface AgentConfig {
  id?: string;
  user_id?: string;
  conversation_id?: string;
  name?: string;
  description?: string;
  provider: string;
  model: string;
  system_prompt?: string;
  temperature?: number;
  max_tokens?: number;
  tools?: ToolDefinition[];
  metadata?: Record<string, string>;
}

/** Tool definition for agent configuration */
export interface ToolDefinition {
  name: string;
  description: string;
  parameters?: Record<string, unknown>;
}

/** Configuration for an agent with a specific role in a swarm */
export interface AgentRoleConfig {
  role: AgentRole;
  config: AgentConfig;
  /** Number of agents with this role */
  count?: number;
  /** Override system prompt for this role */
  system_prompt?: string;
}

/** Configuration for creating a multi-agent swarm */
export interface SwarmConfig {
  id?: string;
  name: string;
  strategy: SwarmStrategy;
  /** Maximum number of agents allowed */
  max_agents?: number;
  /** Timeout in milliseconds */
  timeout?: number;
  /** Agent configurations for the swarm */
  agent_configs: AgentRoleConfig[];
  /** Configuration for the synthesizer agent that combines results */
  synthesizer_config?: AgentConfig;
}

// ============================================================================
// Runtime Types
// ============================================================================

/** Represents an agent within a running swarm */
export interface SwarmAgent {
  id: string;
  role: AgentRole;
  status: AgentStatus;
  input?: string;
  output?: string;
  started_at?: string;
  completed_at?: string;
}

/** Communication between agents in a swarm */
export interface SwarmMessage {
  id: string;
  from_agent: string;
  /** Empty means broadcast to all agents */
  to_agent?: string;
  type: SwarmMessageType;
  content: string;
  timestamp: string;
}

/** Result from an individual agent's execution */
export interface SwarmResult {
  agent_id: string;
  role: AgentRole;
  output: string;
  success: boolean;
  error?: string;
  /** Duration in milliseconds */
  duration: number;
  metadata?: Record<string, unknown>;
}

/** Event emitted during swarm execution */
export interface SwarmEvent {
  swarm_id: string;
  type: SwarmEventType;
  agent_id?: string;
  role?: AgentRole;
  data?: Record<string, unknown>;
  timestamp: string;
}

/** Full swarm state */
export interface Swarm {
  id: string;
  config: SwarmConfig;
  status: SwarmStatus;
  agents: SwarmAgent[];
  messages: SwarmMessage[];
  results: SwarmResult[];
  final_output?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
}

// ============================================================================
// Real-time Update Types
// ============================================================================

/** Progress information for real-time updates */
export interface SwarmProgressInfo {
  swarm_id: string;
  status: SwarmStatus;
  total_agents: number;
  completed_agents: number;
  failed_agents: number;
  current_phase?: string;
  progress_percentage: number;
  estimated_time_remaining?: number;
}

/** Agent status for real-time updates */
export interface SwarmAgentInfo {
  id: string;
  role: AgentRole;
  status: AgentStatus;
  current_task?: string;
  output_preview?: string;
  tokens_used?: number;
  started_at?: string;
  completed_at?: string;
}

// ============================================================================
// API Request/Response Types
// ============================================================================

/** Request to create a new swarm */
export interface CreateSwarmRequest {
  config: SwarmConfig;
}

/** Response from creating a swarm */
export interface CreateSwarmResponse {
  swarm: Swarm;
}

/** Request to run a swarm with a task */
export interface RunSwarmRequest {
  swarm_id: string;
  task: string;
}

/** Response from running a swarm */
export interface RunSwarmResponse {
  success: boolean;
  error?: string;
}

/** Request to stop a running swarm */
export interface StopSwarmRequest {
  swarm_id: string;
}

/** Response from stopping a swarm */
export interface StopSwarmResponse {
  success: boolean;
  error?: string;
}

/** Response from getting swarm status */
export interface GetSwarmResponse {
  swarm: Swarm;
}

/** Response from listing all swarms */
export interface ListSwarmsResponse {
  swarms: Swarm[];
}

// ============================================================================
// WebSocket Message Types for Swarm
// ============================================================================

/** WebSocket message for swarm events */
export interface SwarmWSMessage {
  type: 'swarm.event' | 'swarm.progress' | 'swarm.complete' | 'swarm.error';
  swarm_id: string;
  event?: SwarmEvent;
  progress?: SwarmProgressInfo;
  result?: Swarm;
  error?: string;
}
