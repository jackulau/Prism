// Approval workflow types

export type ApprovalStatus = 'pending' | 'approved' | 'rejected' | 'escalated' | 'expired';

export type ApprovalType = 'tool_execution' | 'config_change' | 'deployment' | 'access_request' | 'custom';

export interface Approver {
  id: string;
  email: string;
  name?: string;
  role?: string;
}

export interface ApprovalDecision {
  id: string;
  approvalRequestId: string;
  decidedBy: Approver;
  decision: 'approved' | 'rejected';
  comment?: string;
  decidedAt: Date;
}

export interface ApprovalRequest {
  id: string;
  type: ApprovalType;
  status: ApprovalStatus;
  title: string;
  description?: string;
  requestedBy: Approver;
  requestedAt: Date;
  expiresAt?: Date;
  workflowId?: string;
  workflowStepId?: string;
  currentStep?: number;
  totalSteps?: number;
  metadata?: Record<string, unknown>;
  decisions: ApprovalDecision[];
  // For tool execution approvals
  toolName?: string;
  toolParameters?: Record<string, unknown>;
  // For config changes
  configDiff?: {
    before: string;
    after: string;
  };
  // Escalation info
  escalatedAt?: Date;
  escalatedTo?: Approver[];
  // Conversation context
  conversationId?: string;
  workspaceId?: string;
}

export interface ApprovalStepConfig {
  id: string;
  name: string;
  order: number;
  approverType: 'role' | 'user' | 'any';
  approverRoles?: string[];
  approverUserIds?: string[];
  requiredApprovals: number;
  timeoutMinutes?: number;
  escalationPolicy?: EscalationPolicy;
  parallelWithPrevious?: boolean;
}

export interface EscalationPolicy {
  enabled: boolean;
  escalateAfterMinutes: number;
  escalateTo: 'manager' | 'admin' | 'specific_users';
  escalationUserIds?: string[];
  maxEscalations?: number;
}

export interface ApprovalWorkflow {
  id: string;
  name: string;
  description?: string;
  triggerType: ApprovalType;
  triggerConditions?: Record<string, unknown>;
  steps: ApprovalStepConfig[];
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}

export interface ApprovalStats {
  pending: number;
  approvedToday: number;
  rejectedToday: number;
  avgResponseTimeMinutes: number;
}

// API request/response types
export interface SubmitApprovalDecisionRequest {
  requestId: string;
  decision: 'approved' | 'rejected';
  comment?: string;
}

export interface CreateWorkflowRequest {
  name: string;
  description?: string;
  triggerType: ApprovalType;
  triggerConditions?: Record<string, unknown>;
  steps: Omit<ApprovalStepConfig, 'id'>[];
}

export interface UpdateWorkflowRequest {
  name?: string;
  description?: string;
  triggerConditions?: Record<string, unknown>;
  steps?: Omit<ApprovalStepConfig, 'id'>[];
  isActive?: boolean;
}

export interface ApprovalListResponse {
  approvals: ApprovalRequest[];
  total: number;
  page: number;
  pageSize: number;
}

export interface WorkflowListResponse {
  workflows: ApprovalWorkflow[];
  total: number;
}

// WebSocket event types for real-time updates
export type ApprovalEventType =
  | 'approval.created'
  | 'approval.updated'
  | 'approval.decided'
  | 'approval.escalated'
  | 'approval.expired';

export interface ApprovalEvent {
  type: ApprovalEventType;
  approval: ApprovalRequest;
  timestamp: Date;
}
