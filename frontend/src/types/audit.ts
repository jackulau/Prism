// Audit Log Types
export type AuditActionType =
  | 'user.login'
  | 'user.logout'
  | 'user.create'
  | 'user.update'
  | 'user.delete'
  | 'workspace.create'
  | 'workspace.update'
  | 'workspace.delete'
  | 'conversation.create'
  | 'conversation.delete'
  | 'api_key.create'
  | 'api_key.revoke'
  | 'settings.update'
  | 'export.request'
  | 'export.download'
  | 'member.invite'
  | 'member.remove'
  | 'role.change';

export interface AuditLogEntry {
  id: string;
  timestamp: Date;
  action: AuditActionType;
  actor: {
    id: string;
    email: string;
    name?: string;
  };
  resource: {
    type: string;
    id: string;
    name?: string;
  };
  details?: {
    before?: Record<string, unknown>;
    after?: Record<string, unknown>;
    metadata?: Record<string, unknown>;
  };
  ipAddress?: string;
  userAgent?: string;
  sessionId?: string;
  organizationId?: string;
}

export interface AuditLogFilter {
  startDate?: Date;
  endDate?: Date;
  actions?: AuditActionType[];
  actorId?: string;
  resourceType?: string;
  resourceId?: string;
  search?: string;
}

export interface AuditLogPage {
  entries: AuditLogEntry[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

// Export Job Types
export type ExportType = 'gdpr' | 'audit' | 'usage' | 'conversations';

export type ExportFormat = 'json' | 'csv';

export type ExportJobStatus = 'pending' | 'processing' | 'complete' | 'failed' | 'expired';

export interface ExportJob {
  id: string;
  type: ExportType;
  format: ExportFormat;
  status: ExportJobStatus;
  progress: number;
  createdAt: Date;
  completedAt?: Date;
  expiresAt?: Date;
  downloadUrl?: string;
  fileSize?: number;
  error?: string;
  filters?: {
    startDate?: Date;
    endDate?: Date;
  };
  requestedBy: {
    id: string;
    email: string;
  };
}

export interface CreateExportRequest {
  type: ExportType;
  format: ExportFormat;
  startDate?: Date;
  endDate?: Date;
}

// Compliance Report Types
export type ReportType = 'access' | 'changes' | 'security' | 'summary';

export type ReportFormat = 'pdf' | 'html' | 'json';

export interface ComplianceReport {
  id: string;
  type: ReportType;
  format: ReportFormat;
  title: string;
  description?: string;
  generatedAt: Date;
  period: {
    start: Date;
    end: Date;
  };
  downloadUrl?: string;
  previewHtml?: string;
  metrics?: ReportMetrics;
}

export interface ReportMetrics {
  totalUsers?: number;
  activeUsers?: number;
  totalActions?: number;
  securityEvents?: number;
  dataExports?: number;
}

export interface GenerateReportRequest {
  type: ReportType;
  format: ReportFormat;
  startDate: Date;
  endDate: Date;
}

// Retention Policy Types
export type DataType =
  | 'conversations'
  | 'messages'
  | 'audit_logs'
  | 'exports'
  | 'user_data'
  | 'analytics';

export interface RetentionPolicy {
  dataType: DataType;
  retentionDays: number;
  description: string;
  isConfigurable: boolean;
  legalHold: boolean;
  lastUpdated?: Date;
  updatedBy?: {
    id: string;
    email: string;
  };
}

export interface RetentionPolicyUpdate {
  dataType: DataType;
  retentionDays: number;
  legalHold?: boolean;
}

// API Response Types
export interface AuditApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}
