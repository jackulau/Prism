import {
  Bot,
  Wrench,
  GitBranch,
  Layers,
  Clock,
  ArrowRightLeft,
  type LucideIcon,
} from 'lucide-react';

export type StepType = 'agent' | 'tool' | 'condition' | 'parallel' | 'wait' | 'transform';
export type StepCategory = 'ai' | 'logic' | 'control' | 'data';

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

export interface ConditionStepConfig {
  expression: string;
  trueBranch: string;
  falseBranch: string;
}

export interface ParallelStepConfig {
  steps: string[];
  waitForAll: boolean;
  failOnFirst: boolean;
}

export interface WaitStepConfig {
  waitType: 'user_input' | 'webhook' | 'timeout';
  timeout?: number;
  promptText?: string;
  webhookPath?: string;
  outputKey?: string;
}

export interface TransformStepConfig {
  type: 'jq' | 'template' | 'script';
  template?: string;
  script?: string;
  inputKey?: string;
  outputKey?: string;
  mapping?: Record<string, string>;
}

export type StepConfig =
  | { agent: AgentStepConfig }
  | { tool: ToolStepConfig }
  | { condition: ConditionStepConfig }
  | { parallel: ParallelStepConfig }
  | { wait: WaitStepConfig }
  | { transform: TransformStepConfig };

export interface StepTypeDefinition {
  type: StepType;
  name: string;
  description: string;
  detailedDescription: string;
  icon: LucideIcon;
  color: string;
  bgColor: string;
  category: StepCategory;
  defaultConfig: Partial<StepConfig>;
}

export const STEP_CATEGORIES: Record<StepCategory, { label: string; description: string }> = {
  ai: { label: 'AI', description: 'AI-powered steps that use language models' },
  logic: { label: 'Logic', description: 'Conditional branching and decision making' },
  control: { label: 'Control', description: 'Control flow and execution management' },
  data: { label: 'Data', description: 'Data transformation and manipulation' },
};

export const STEP_TYPE_DEFINITIONS: StepTypeDefinition[] = [
  {
    type: 'agent',
    name: 'Agent',
    description: 'Execute LLM agent with prompt',
    detailedDescription:
      'Run an AI agent with a custom prompt. The agent can use enabled tools and produce structured output that gets stored in workflow state.',
    icon: Bot,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/20',
    category: 'ai',
    defaultConfig: {
      agent: {
        provider: 'anthropic',
        model: 'claude-3-5-sonnet-20241022',
        prompt: '',
        temperature: 0.7,
        maxTokens: 4096,
        tools: [],
        outputKey: 'agent_output',
      },
    },
  },
  {
    type: 'tool',
    name: 'Tool',
    description: 'Run a registered tool with parameters',
    detailedDescription:
      'Execute a specific tool with defined parameters. Useful for API calls, file operations, or integrations with external services.',
    icon: Wrench,
    color: 'text-purple-400',
    bgColor: 'bg-purple-500/20',
    category: 'ai',
    defaultConfig: {
      tool: {
        toolName: '',
        parameters: {},
        outputKey: 'tool_output',
      },
    },
  },
  {
    type: 'condition',
    name: 'Condition',
    description: 'Evaluate expression for branching',
    detailedDescription:
      'Create conditional logic in your workflow. Evaluate expressions against workflow state to determine which path to take.',
    icon: GitBranch,
    color: 'text-yellow-400',
    bgColor: 'bg-yellow-500/20',
    category: 'logic',
    defaultConfig: {
      condition: {
        expression: '',
        trueBranch: '',
        falseBranch: '',
      },
    },
  },
  {
    type: 'parallel',
    name: 'Parallel',
    description: 'Execute multiple steps concurrently',
    detailedDescription:
      'Run multiple workflow steps at the same time. Configure whether to wait for all steps or fail on the first error.',
    icon: Layers,
    color: 'text-green-400',
    bgColor: 'bg-green-500/20',
    category: 'control',
    defaultConfig: {
      parallel: {
        steps: [],
        waitForAll: true,
        failOnFirst: false,
      },
    },
  },
  {
    type: 'wait',
    name: 'Wait',
    description: 'Pause for user input, webhook, or timeout',
    detailedDescription:
      'Pause workflow execution until a condition is met. Wait for user input, webhook callback, or a specified timeout duration.',
    icon: Clock,
    color: 'text-orange-400',
    bgColor: 'bg-orange-500/20',
    category: 'control',
    defaultConfig: {
      wait: {
        waitType: 'user_input',
        timeout: 300000,
        promptText: 'Please provide input to continue',
        outputKey: 'user_input',
      },
    },
  },
  {
    type: 'transform',
    name: 'Transform',
    description: 'Transform data without LLM',
    detailedDescription:
      'Manipulate and transform data in workflow state. Use JQ expressions, templates, or scripts to reshape data between steps.',
    icon: ArrowRightLeft,
    color: 'text-cyan-400',
    bgColor: 'bg-cyan-500/20',
    category: 'data',
    defaultConfig: {
      transform: {
        type: 'template',
        template: '',
        inputKey: '',
        outputKey: 'transformed_output',
      },
    },
  },
];

export function getStepTypeDefinition(type: StepType): StepTypeDefinition | undefined {
  return STEP_TYPE_DEFINITIONS.find((def) => def.type === type);
}

export function getStepsByCategory(category: StepCategory): StepTypeDefinition[] {
  return STEP_TYPE_DEFINITIONS.filter((def) => def.category === category);
}

export function getCategoryForStep(type: StepType): StepCategory | undefined {
  return getStepTypeDefinition(type)?.category;
}
