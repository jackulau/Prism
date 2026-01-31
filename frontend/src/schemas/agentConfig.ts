/**
 * Zod validation schemas for agent configuration forms.
 */

import { z } from 'zod';

// ============================================================================
// Execution Config Schema
// ============================================================================

/**
 * Schema for agent execution configuration parameters
 */
export const agentExecutionConfigSchema = z.object({
  temperature: z
    .number()
    .min(0, 'Temperature must be at least 0')
    .max(2, 'Temperature must be at most 2')
    .default(0.7),
  maxTokens: z
    .number()
    .int('Max tokens must be a whole number')
    .min(1, 'Max tokens must be at least 1')
    .max(200000, 'Max tokens must be at most 200,000')
    .default(4096),
  systemPrompt: z.string().optional(),
  enabledTools: z.array(z.string()).default([]),
});

// ============================================================================
// Task Input Schema
// ============================================================================

/**
 * Schema for validating task prompt input
 */
export const taskInputSchema = z
  .string()
  .min(1, 'Task prompt is required')
  .max(100000, 'Task prompt must be less than 100,000 characters')
  .refine(
    (val) => val.trim().length > 0,
    'Task prompt cannot be empty or whitespace only'
  );

// ============================================================================
// Agent Config Schema
// ============================================================================

/**
 * Full agent configuration schema for form validation
 */
export const agentConfigSchema = z.object({
  provider: z.string().min(1, 'Provider is required'),
  model: z.string().min(1, 'Model is required'),
  prompt: taskInputSchema,
  executionConfig: agentExecutionConfigSchema,
});

// ============================================================================
// Form State Schema
// ============================================================================

/**
 * Schema for the SingleAgentForm internal state
 */
export const singleAgentFormStateSchema = z.object({
  provider: z.string(),
  model: z.string(),
  prompt: z.string(),
  temperature: z.number().min(0).max(2),
  maxTokens: z.number().int().min(1).max(200000),
  systemPrompt: z.string(),
  enabledTools: z.array(z.string()),
  showAdvanced: z.boolean(),
});

// ============================================================================
// Type Exports
// ============================================================================

export type AgentExecutionConfigInput = z.input<typeof agentExecutionConfigSchema>;
export type AgentExecutionConfig = z.output<typeof agentExecutionConfigSchema>;
export type TaskInput = z.output<typeof taskInputSchema>;
export type AgentConfigInput = z.input<typeof agentConfigSchema>;
export type AgentConfig = z.output<typeof agentConfigSchema>;
export type SingleAgentFormState = z.output<typeof singleAgentFormStateSchema>;

// ============================================================================
// Validation Helpers
// ============================================================================

/**
 * Validates agent configuration and returns parsed result or errors
 */
export function validateAgentConfig(data: unknown): {
  success: boolean;
  data?: AgentConfig;
  errors?: z.ZodError;
} {
  const result = agentConfigSchema.safeParse(data);
  if (result.success) {
    return { success: true, data: result.data };
  }
  return { success: false, errors: result.error };
}

/**
 * Validates task input and returns parsed result or error message
 */
export function validateTaskInput(input: string): {
  valid: boolean;
  message?: string;
} {
  const result = taskInputSchema.safeParse(input);
  if (result.success) {
    return { valid: true };
  }
  return { valid: false, message: result.error.errors[0]?.message };
}
