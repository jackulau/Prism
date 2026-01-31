/**
 * Agent configuration and execution types for the Single Agent Form component.
 */

/**
 * Configuration for agent execution parameters
 */
export interface AgentExecutionConfig {
  /** Temperature for LLM sampling (0-2) */
  temperature: number;
  /** Maximum tokens for response */
  maxTokens: number;
  /** Optional system prompt override */
  systemPrompt?: string;
  /** List of enabled tool IDs */
  enabledTools?: string[];
}

/**
 * Main agent configuration for form submission
 */
export interface AgentConfig {
  /** Provider name (e.g., 'anthropic', 'openai', 'ollama') */
  provider: string;
  /** Model ID (e.g., 'claude-3-opus-20240229') */
  model: string;
  /** Task prompt/instruction */
  prompt: string;
  /** Execution configuration */
  executionConfig: AgentExecutionConfig;
}

/**
 * Default values for agent execution configuration
 */
export const DEFAULT_AGENT_EXECUTION_CONFIG: AgentExecutionConfig = {
  temperature: 0.7,
  maxTokens: 4096,
  systemPrompt: undefined,
  enabledTools: [],
};

/**
 * Available tool definition for tool selection UI
 */
export interface AvailableTool {
  /** Unique tool identifier */
  id: string;
  /** Display name */
  name: string;
  /** Tool description */
  description: string;
  /** Category for grouping (e.g., 'file', 'code', 'web') */
  category?: string;
  /** Whether the tool is enabled by default */
  defaultEnabled?: boolean;
}

/**
 * Form state for the SingleAgentForm component
 */
export interface SingleAgentFormState {
  /** Selected provider name */
  provider: string;
  /** Selected model ID */
  model: string;
  /** Task prompt text */
  prompt: string;
  /** Temperature setting */
  temperature: number;
  /** Max tokens setting */
  maxTokens: number;
  /** System prompt override */
  systemPrompt: string;
  /** List of enabled tool IDs */
  enabledTools: string[];
  /** Whether advanced settings section is expanded */
  showAdvanced: boolean;
}

/**
 * Props for the SingleAgentForm component
 */
export interface SingleAgentFormProps {
  /** Callback when form is submitted with valid configuration */
  onSubmit: (config: AgentConfig) => void | Promise<void>;
  /** Callback when form is closed/cancelled */
  onClose: () => void;
  /** Initial values for the form (optional) */
  initialValues?: Partial<SingleAgentFormState>;
  /** Whether the form is currently submitting */
  isSubmitting?: boolean;
  /** List of available tools for selection */
  availableTools?: AvailableTool[];
}

/**
 * Props for the AgentConfigSection component
 */
export interface AgentConfigSectionProps {
  /** Current temperature value */
  temperature: number;
  /** Callback when temperature changes */
  onTemperatureChange: (value: number) => void;
  /** Current max tokens value */
  maxTokens: number;
  /** Callback when max tokens changes */
  onMaxTokensChange: (value: number) => void;
  /** Current system prompt value */
  systemPrompt: string;
  /** Callback when system prompt changes */
  onSystemPromptChange: (value: string) => void;
  /** List of enabled tool IDs */
  enabledTools: string[];
  /** Callback when tool selection changes */
  onToolsChange: (tools: string[]) => void;
  /** Available tools for selection */
  availableTools?: AvailableTool[];
  /** Whether the section is initially collapsed */
  defaultCollapsed?: boolean;
}
