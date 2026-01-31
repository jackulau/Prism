/**
 * Tool catalog TypeScript types.
 * These types match the backend Tool struct and API responses.
 */

// ============================================================================
// Tool Types
// ============================================================================

/** Tool entity matching backend ToolResponse struct */
export interface Tool {
  id: string;
  display_name: string;
  slug_name: string;
  description?: string;
  is_model: boolean;
  provider_id?: string;
  parameters_schema?: string;
  created_at: string;
  updated_at: string;
}

/** Response from listing tools */
export interface ToolListResponse {
  tools: Tool[];
}

/** Parameters for listing tools */
export interface ToolListParams {
  provider_id?: string;
  models_only?: boolean;
}

/** Input for creating a new tool */
export interface ToolCreateInput {
  display_name: string;
  slug_name: string;
  description?: string;
  is_model?: boolean;
  provider_id?: string;
  parameters_schema?: string;
}

/** Input for updating an existing tool (all fields optional) */
export interface ToolUpdateInput {
  display_name?: string;
  slug_name?: string;
  description?: string;
  is_model?: boolean;
  provider_id?: string;
  parameters_schema?: string;
}
