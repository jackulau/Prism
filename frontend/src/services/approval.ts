import { z } from 'zod';
import type {
  ApprovalRequest,
  ApprovalWorkflow,
  ApprovalListResponse,
  WorkflowListResponse,
  SubmitApprovalDecisionRequest,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  ApprovalStats,
} from '../types/approval';

const API_BASE_URL = '/api/v1';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

// Zod schemas for validation (output type matches the TypeScript interfaces)
const approverSchema = z.object({
  id: z.string(),
  email: z.string(),
  name: z.string().optional(),
  role: z.string().optional(),
});

const approvalDecisionSchema = z.object({
  id: z.string(),
  approvalRequestId: z.string(),
  decidedBy: approverSchema,
  decision: z.enum(['approved', 'rejected']),
  comment: z.string().optional(),
  decidedAt: z.coerce.date(),
});

const approvalRequestSchema = z.object({
  id: z.string(),
  type: z.enum(['tool_execution', 'config_change', 'deployment', 'access_request', 'custom']),
  status: z.enum(['pending', 'approved', 'rejected', 'escalated', 'expired']),
  title: z.string(),
  description: z.string().optional(),
  requestedBy: approverSchema,
  requestedAt: z.coerce.date(),
  expiresAt: z.coerce.date().optional(),
  workflowId: z.string().optional(),
  workflowStepId: z.string().optional(),
  currentStep: z.number().optional(),
  totalSteps: z.number().optional(),
  metadata: z.record(z.unknown()).optional(),
  decisions: z.array(approvalDecisionSchema),
  toolName: z.string().optional(),
  toolParameters: z.record(z.unknown()).optional(),
  configDiff: z.object({
    before: z.string(),
    after: z.string(),
  }).optional(),
  escalatedAt: z.coerce.date().optional(),
  escalatedTo: z.array(approverSchema).optional(),
  conversationId: z.string().optional(),
  workspaceId: z.string().optional(),
});

const escalationPolicySchema = z.object({
  enabled: z.boolean(),
  escalateAfterMinutes: z.number(),
  escalateTo: z.enum(['manager', 'admin', 'specific_users']),
  escalationUserIds: z.array(z.string()).optional(),
  maxEscalations: z.number().optional(),
});

const approvalStepConfigSchema = z.object({
  id: z.string(),
  name: z.string(),
  order: z.number(),
  approverType: z.enum(['role', 'user', 'any']),
  approverRoles: z.array(z.string()).optional(),
  approverUserIds: z.array(z.string()).optional(),
  requiredApprovals: z.number(),
  timeoutMinutes: z.number().optional(),
  escalationPolicy: escalationPolicySchema.optional(),
  parallelWithPrevious: z.boolean().optional(),
});

const approvalWorkflowSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  triggerType: z.enum(['tool_execution', 'config_change', 'deployment', 'access_request', 'custom']),
  triggerConditions: z.record(z.unknown()).optional(),
  steps: z.array(approvalStepConfigSchema),
  isActive: z.boolean(),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
  createdBy: z.string(),
});

const approvalListResponseSchema = z.object({
  approvals: z.array(approvalRequestSchema),
  total: z.number(),
  page: z.number(),
  pageSize: z.number(),
});

const workflowListResponseSchema = z.object({
  workflows: z.array(approvalWorkflowSchema),
  total: z.number(),
});

const approvalStatsSchema = z.object({
  pending: z.number(),
  approvedToday: z.number(),
  rejectedToday: z.number(),
  avgResponseTimeMinutes: z.number(),
});

class ApprovalService {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    schema: z.ZodType<T>,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });

      const contentType = response.headers.get('Content-Type');
      const hasJsonContent = contentType?.includes('application/json');

      let json: unknown;
      if (hasJsonContent) {
        const text = await response.text();
        if (text) {
          json = JSON.parse(text);
        }
      }

      if (!response.ok) {
        return {
          error: (json as { error?: string })?.error || `HTTP ${response.status}: ${response.statusText}`,
        };
      }

      // Parse and validate with schema - this also transforms dates
      const result = schema.safeParse(json);
      if (!result.success) {
        console.warn('API response validation failed:', result.error.issues);
        // Return unparsed data but log warning
        return { data: json as T };
      }

      return { data: result.data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  // Approval Requests
  async listApprovals(params?: {
    status?: string;
    type?: string;
    page?: number;
    pageSize?: number;
  }): Promise<ApiResponse<ApprovalListResponse>> {
    const query = new URLSearchParams();
    if (params?.status) query.set('status', params.status);
    if (params?.type) query.set('type', params.type);
    if (params?.page) query.set('page', params.page.toString());
    if (params?.pageSize) query.set('pageSize', params.pageSize.toString());

    const queryString = query.toString();
    return this.request(
      `/approvals${queryString ? `?${queryString}` : ''}`,
      approvalListResponseSchema as z.ZodType<ApprovalListResponse>
    );
  }

  async getPendingApprovals(): Promise<ApiResponse<ApprovalListResponse>> {
    return this.listApprovals({ status: 'pending' });
  }

  async getApprovalHistory(params?: {
    page?: number;
    pageSize?: number;
  }): Promise<ApiResponse<ApprovalListResponse>> {
    const query = new URLSearchParams();
    query.set('status', 'approved,rejected,expired');
    if (params?.page) query.set('page', params.page.toString());
    if (params?.pageSize) query.set('pageSize', params.pageSize.toString());

    return this.request(
      `/approvals?${query.toString()}`,
      approvalListResponseSchema as z.ZodType<ApprovalListResponse>
    );
  }

  async getApproval(id: string): Promise<ApiResponse<ApprovalRequest>> {
    return this.request(
      `/approvals/${id}`,
      approvalRequestSchema as z.ZodType<ApprovalRequest>
    );
  }

  async submitDecision(request: SubmitApprovalDecisionRequest): Promise<ApiResponse<ApprovalRequest>> {
    return this.request(
      `/approvals/${request.requestId}/decide`,
      approvalRequestSchema as z.ZodType<ApprovalRequest>,
      {
        method: 'POST',
        body: JSON.stringify({
          decision: request.decision,
          comment: request.comment,
        }),
      }
    );
  }

  async getApprovalStats(): Promise<ApiResponse<ApprovalStats>> {
    return this.request('/approvals/stats', approvalStatsSchema);
  }

  // Workflows
  async listWorkflows(): Promise<ApiResponse<WorkflowListResponse>> {
    return this.request(
      '/workflows',
      workflowListResponseSchema as z.ZodType<WorkflowListResponse>
    );
  }

  async getWorkflow(id: string): Promise<ApiResponse<ApprovalWorkflow>> {
    return this.request(
      `/workflows/${id}`,
      approvalWorkflowSchema as z.ZodType<ApprovalWorkflow>
    );
  }

  async createWorkflow(request: CreateWorkflowRequest): Promise<ApiResponse<ApprovalWorkflow>> {
    return this.request(
      '/workflows',
      approvalWorkflowSchema as z.ZodType<ApprovalWorkflow>,
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
  }

  async updateWorkflow(id: string, request: UpdateWorkflowRequest): Promise<ApiResponse<ApprovalWorkflow>> {
    return this.request(
      `/workflows/${id}`,
      approvalWorkflowSchema as z.ZodType<ApprovalWorkflow>,
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      }
    );
  }

  async deleteWorkflow(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request(`/workflows/${id}`, z.object({ success: z.boolean() }), {
      method: 'DELETE',
    });
  }

  async toggleWorkflow(id: string, isActive: boolean): Promise<ApiResponse<ApprovalWorkflow>> {
    return this.updateWorkflow(id, { isActive });
  }
}

export const approvalService = new ApprovalService();
