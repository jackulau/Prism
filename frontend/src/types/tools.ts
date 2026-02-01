/**
 * Tool catalog TypeScript types.
 * These types match the backend Tool struct and API responses.
 */

// ============================================================================
// Tool Types
// ============================================================================

/** Tool type categories */
export type ToolType = 'file' | 'web' | 'code' | 'api' | 'system' | 'mcp' | 'custom' | 'model' | 'builtin' | 'all' | 'unknown';

/** Get tool type from tool name or Tool object */
export function getToolType(tool: string | Tool): ToolType {
  // Handle Tool object
  if (typeof tool === 'object' && tool !== null) {
    if (tool.is_model) {
      return 'model';
    }
    const toolName = tool.slug_name || tool.display_name || '';
    return getToolTypeFromName(toolName);
  }

  // Handle string
  return getToolTypeFromName(tool);
}

/** Get tool type from tool name string */
function getToolTypeFromName(toolName: string): ToolType {
  const lowerName = toolName.toLowerCase();

  if (lowerName.includes('file') || lowerName.includes('read') || lowerName.includes('write')) {
    return 'file';
  }
  if (lowerName.includes('web') || lowerName.includes('fetch') || lowerName.includes('http')) {
    return 'web';
  }
  if (lowerName.includes('code') || lowerName.includes('execute') || lowerName.includes('run')) {
    return 'code';
  }
  if (lowerName.includes('api') || lowerName.includes('request')) {
    return 'api';
  }
  if (lowerName.includes('system') || lowerName.includes('shell') || lowerName.includes('bash')) {
    return 'system';
  }
  if (lowerName.includes('mcp')) {
    return 'mcp';
  }

  return 'unknown';
}

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
