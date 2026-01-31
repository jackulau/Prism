// Tool catalog types matching backend ToolResponse struct

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

export interface ToolListResponse {
  tools: Tool[];
}

export interface CreateToolInput {
  display_name: string;
  slug_name: string;
  description?: string;
  is_model: boolean;
  provider_id?: string;
  parameters_schema?: string;
}

export interface UpdateToolInput {
  display_name?: string;
  slug_name?: string;
  description?: string;
  is_model?: boolean;
  provider_id?: string;
  parameters_schema?: string;
}

// Utility type for tool type filtering
export type ToolType = 'all' | 'model' | 'builtin' | 'custom';

// Helper to determine tool type for display
export function getToolType(tool: Tool): ToolType {
  if (tool.is_model) return 'model';
  if (tool.provider_id) return 'custom';
  return 'builtin';
}

// Helper to get icon for tool type
export function getToolIcon(tool: Tool): string {
  if (tool.is_model) return '🤖';
  if (tool.provider_id) return '🔧';
  return '⚡';
}
