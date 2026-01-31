/**
 * Zod validation schemas for tool catalog API responses.
 * These schemas provide runtime validation for tool-related API responses.
 */

import { z } from 'zod';

// ============================================================================
// Tool Schemas
// ============================================================================

/** Schema for a single tool */
export const toolSchema = z.object({
  id: z.string(),
  display_name: z.string(),
  slug_name: z.string(),
  description: z.string().optional(),
  is_model: z.boolean(),
  provider_id: z.string().optional(),
  parameters_schema: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

/** Schema for tool list response */
export const toolListResponseSchema = z.object({
  tools: z.array(toolSchema),
});

/** Schema for tool list params */
export const toolListParamsSchema = z.object({
  provider_id: z.string().optional(),
  models_only: z.boolean().optional(),
});

/** Schema for creating a tool */
export const toolCreateInputSchema = z.object({
  display_name: z.string().min(1, 'Display name is required'),
  slug_name: z.string().min(1, 'Slug name is required'),
  description: z.string().optional(),
  is_model: z.boolean().optional(),
  provider_id: z.string().optional(),
  parameters_schema: z.string().optional(),
});

/** Schema for updating a tool */
export const toolUpdateInputSchema = z.object({
  display_name: z.string().min(1).optional(),
  slug_name: z.string().min(1).optional(),
  description: z.string().optional(),
  is_model: z.boolean().optional(),
  provider_id: z.string().optional(),
  parameters_schema: z.string().optional(),
});

// ============================================================================
// Type Exports (inferred from schemas)
// ============================================================================

export type Tool = z.infer<typeof toolSchema>;
export type ToolListResponse = z.infer<typeof toolListResponseSchema>;
export type ToolListParams = z.infer<typeof toolListParamsSchema>;
export type ToolCreateInput = z.infer<typeof toolCreateInputSchema>;
export type ToolUpdateInput = z.infer<typeof toolUpdateInputSchema>;
