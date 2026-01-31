// Workflow types matching backend types.go

export type WorkflowStatus = 'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'cancelled';
export type StepStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
export type StepType = 'agent' | 'tool' | 'condition' | 'parallel' | 'wait' | 'transform';

// Step configuration types
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

export interface ToolStepConfig {
  toolName: string;
  parameters: Record<string, unknown>;
  outputKey?: string;
}

export interface ConditionConfig {
  expression: string;
  trueBranch: string;
  falseBranch: string;
}

export interface ParallelStepConfig {
  steps: WorkflowStep[];
  waitForAll: boolean;
  failOnFirst: boolean;
}

export type WaitType = 'user_input' | 'webhook' | 'timeout';

export interface WaitStepConfig {
  waitType: WaitType;
  timeout?: number; // milliseconds
  promptText?: string;
  webhookPath?: string;
  outputKey?: string;
}

export type TransformType = 'jq' | 'template' | 'script';

export interface TransformStepConfig {
  type: TransformType;
  template?: string;
  script?: string;
  inputKey?: string;
  outputKey?: string;
  mapping?: Record<string, string>;
}

// Combined step config - only one should be set based on step type
export interface StepConfig {
  agentConfig?: AgentStepConfig;
  toolConfig?: ToolStepConfig;
  conditionConfig?: ConditionConfig;
  parallelConfig?: ParallelStepConfig;
  waitConfig?: WaitStepConfig;
  transformConfig?: TransformStepConfig;
}

// Condition for step execution
export interface Condition {
  type: 'expression' | 'state_check';
  expression?: string;
  stateKey?: string;
  operator?: 'equals' | 'not_equals' | 'exists' | 'contains';
  value?: string;
}

// Retry policy
export interface RetryPolicy {
  maxRetries: number;
  delay: number; // milliseconds
  backoffType?: 'fixed' | 'exponential';
  maxDelay?: number;
}

// Workflow step
export interface WorkflowStep {
  id: string;
  name: string;
  description?: string;
  type: StepType;
  config: StepConfig;
  condition?: Condition;
  onSuccess?: string;
  onFailure?: string;
  timeout?: number;
  retryPolicy?: RetryPolicy;
}

// Workflow definition
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

// Step result from execution
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

// Workflow node for the canvas (extended step with position)
export interface WorkflowNode {
  id: string;
  type: StepType;
  position: { x: number; y: number };
  data: {
    name: string;
    description?: string;
    config: StepConfig;
    condition?: Condition;
    onSuccess?: string;
    onFailure?: string;
    timeout?: number;
    retryPolicy?: RetryPolicy;
  };
}

// Edge connecting nodes
export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
  label?: string;
  type?: 'success' | 'failure' | 'default';
}

// State variable info for the picker
export interface StateVariable {
  key: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array' | 'unknown';
  source: string; // step name that produces this variable
  description?: string;
}

// Common operators for conditions
export const CONDITION_OPERATORS = [
  { value: '==', label: 'equals' },
  { value: '!=', label: 'not equals' },
  { value: 'contains', label: 'contains' },
  { value: 'exists', label: 'exists' },
  { value: '>', label: 'greater than' },
  { value: '<', label: 'less than' },
  { value: '>=', label: 'greater than or equal' },
  { value: '<=', label: 'less than or equal' },
] as const;

// Provider options for agent step
export const DEFAULT_PROVIDERS = [
  'anthropic',
  'openai',
  'google',
  'groq',
  'deepseek',
  'ollama',
  'openrouter',
] as const;
