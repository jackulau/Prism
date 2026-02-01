// Workflow Types - Frontend type definitions matching backend structures

// Step types matching backend StepType constants
export type StepType = 'agent' | 'tool' | 'condition' | 'parallel' | 'wait' | 'transform';

// Transform types for transform step configuration
export type TransformType = 'jq' | 'template' | 'script';

// Wait types for wait step configuration
export type WaitType = 'user_input' | 'webhook' | 'timeout';

// Condition operators for condition step configuration
export type ConditionOperator = 'equals' | 'not_equals' | 'greater_than' | 'less_than' | 'greater_than_or_equal' | 'less_than_or_equal' | 'contains' | 'not_contains' | 'starts_with' | 'ends_with' | 'matches' | 'exists' | 'not_exists' | 'is_empty' | 'is_not_empty';

export const CONDITION_OPERATORS: ConditionOperator[] = [
  'equals',
  'not_equals',
  'greater_than',
  'less_than',
  'greater_than_or_equal',
  'less_than_or_equal',
  'contains',
  'not_contains',
  'starts_with',
  'ends_with',
  'matches',
  'exists',
  'not_exists',
  'is_empty',
  'is_not_empty',
];

// State variable type for workflow state
export interface StateVariable {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array' | 'any';
  description?: string;
  defaultValue?: unknown;
  sourceStepId?: string;
}

// Workflow status types matching backend WorkflowStatus
export type WorkflowStatus = 'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'cancelled';

// Step status types matching backend StepStatus
export type StepStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';

// Node visual states for the canvas
export type NodeVisualState = 'default' | 'selected' | 'running' | 'completed' | 'failed' | 'hovering';

// Edge labels for conditional branches
export type EdgeLabel = 'success' | 'failure' | 'true' | 'false';

// Position type for node placement
export interface Position {
  x: number;
  y: number;
}

// Canvas viewport state
export interface ViewportState {
  zoom: number;
  pan: Position;
}

// Agent step configuration
export interface AgentStepConfig {
  provider: string;
  model: string;
  systemPrompt?: string;
  prompt: string;
  temperature?: number;
  maxTokens?: number;
  tools?: string[];
  outputKey?: string;
}

// Tool step configuration
export interface ToolStepConfig {
  toolName: string;
  parameters: Record<string, unknown>;
  outputKey?: string;
}

// Condition configuration
export interface ConditionConfig {
  expression: string;
  trueBranch: string;
  falseBranch: string;
}

// Parallel step configuration
export interface ParallelStepConfig {
  steps: WorkflowStep[];
  waitForAll: boolean;
  failOnFirst: boolean;
}

// Wait step configuration
export interface WaitStepConfig {
  waitType: 'user_input' | 'webhook' | 'timeout';
  timeout?: number;
  promptText?: string;
  webhookPath?: string;
  outputKey?: string;
}

// Transform step configuration
export interface TransformStepConfig {
  type: 'jq' | 'template' | 'script';
  template?: string;
  script?: string;
  inputKey?: string;
  outputKey?: string;
  mapping?: Record<string, string>;
}

// Union type for step configurations
export interface StepConfig {
  agentConfig?: AgentStepConfig;
  toolConfig?: ToolStepConfig;
  conditionConfig?: ConditionConfig;
  parallelConfig?: ParallelStepConfig;
  waitConfig?: WaitStepConfig;
  transformConfig?: TransformStepConfig;
}

// Condition for step execution
export interface StepCondition {
  type: 'expression' | 'state_check';
  expression?: string;
  stateKey?: string;
  operator?: 'equals' | 'not_equals' | 'exists' | 'contains';
  value?: string;
}

// Retry policy for steps
export interface RetryPolicy {
  maxRetries: number;
  delay: number;
  backoffType?: 'fixed' | 'exponential';
  maxDelay?: number;
}

// Workflow step definition (backend format)
export interface WorkflowStep {
  id: string;
  name: string;
  description?: string;
  type: StepType;
  config: StepConfig;
  condition?: StepCondition;
  onSuccess?: string;
  onFailure?: string;
  timeout?: number;
  retryPolicy?: RetryPolicy;
}

// Workflow node for canvas (extended with position)
export interface WorkflowNode {
  id: string;
  type: StepType;
  position: Position;
  config: StepConfig;
  name: string;
  description?: string;
  visualState?: NodeVisualState;
  // Step execution state (optional, populated during runtime)
  status?: StepStatus;
  // Connection targets
  onSuccess?: string;
  onFailure?: string;
  // Node data for React Flow compatibility
  data?: Record<string, unknown>;
}

// Edge connecting two nodes
export interface Edge {
  id: string;
  source: string;
  target: string;
  sourcePort?: 'success' | 'failure' | 'default';
  label?: EdgeLabel;
}

// Connection port on a node
export interface ConnectionPort {
  nodeId: string;
  type: 'input' | 'output';
  position: Position;
  portId: 'success' | 'failure' | 'default' | 'input';
}

// Pending connection state during edge creation
export interface PendingConnection {
  sourceNodeId: string;
  sourcePort: 'success' | 'failure' | 'default';
  mousePosition: Position;
}

// Workflow definition for the canvas editor
export interface WorkflowDefinition {
  id?: string;
  name: string;
  description?: string;
  nodes: WorkflowNode[];
  edges: Edge[];
  initialState?: Record<string, unknown>;
}

// Full workflow with execution state
export interface Workflow {
  id: string;
  userId: string;
  name: string;
  description?: string;
  steps: WorkflowStep[];
  status: WorkflowStatus;
  currentStep: number;
  state?: Record<string, unknown>;
  error?: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  completedAt?: string;
}

// Workflow event types for real-time updates
export type WorkflowEventType =
  | 'workflow_started'
  | 'workflow_paused'
  | 'workflow_resumed'
  | 'workflow_completed'
  | 'workflow_failed'
  | 'workflow_cancelled'
  | 'step_started'
  | 'step_completed'
  | 'step_failed'
  | 'step_skipped'
  | 'step_retrying'
  | 'state_updated'
  | 'waiting_input';

// Workflow event
export interface WorkflowEvent {
  workflowId: string;
  type: WorkflowEventType;
  stepId?: string;
  stepName?: string;
  data?: Record<string, unknown>;
  timestamp: string;
}

// Step result
export interface StepResult {
  stepId: string;
  stepName: string;
  status: StepStatus;
  output?: unknown;
  error?: string;
  duration: number;
  retryCount?: number;
  startedAt: string;
  completedAt: string;
  metadata?: Record<string, unknown>;
}

// Utility type for step type to icon mapping
export const STEP_TYPE_ICONS: Record<StepType, string> = {
  agent: 'Bot',
  tool: 'Wrench',
  condition: 'GitBranch',
  parallel: 'GitFork',
  wait: 'Clock',
  transform: 'RefreshCw',
};

// Utility type for step type labels
export const STEP_TYPE_LABELS: Record<StepType, string> = {
  agent: 'AI Agent',
  tool: 'Tool',
  condition: 'Condition',
  parallel: 'Parallel',
  wait: 'Wait',
  transform: 'Transform',
};

// Default node dimensions
export const NODE_DIMENSIONS = {
  width: 200,
  height: 80,
  portRadius: 6,
  portOffset: 12,
};

// Grid settings for canvas
export const CANVAS_GRID = {
  size: 20,
  snapThreshold: 10,
};

// Zoom constraints
export const ZOOM_CONSTRAINTS = {
  min: 0.25,
  max: 2,
  step: 0.1,
};
