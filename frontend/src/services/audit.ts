import type {
  AuditLogEntry,
  AuditLogFilter,
  AuditLogPage,
  ExportJob,
  CreateExportRequest,
  ComplianceReport,
  GenerateReportRequest,
  RetentionPolicy,
  RetentionPolicyUpdate,
  AuditApiResponse,
} from '../types/audit';

const API_BASE_URL = '/api/v1';

class AuditService {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<AuditApiResponse<T>> {
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

      let data: T | undefined;
      if (hasJsonContent) {
        const text = await response.text();
        if (text) {
          try {
            data = JSON.parse(text);
          } catch {
            if (!response.ok) {
              return { error: text || 'An error occurred' };
            }
          }
        }
      }

      if (!response.ok) {
        return { error: (data as { error?: string })?.error || 'An error occurred' };
      }

      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : 'Network error' };
    }
  }

  // Audit Logs
  async getAuditLogs(
    filter: AuditLogFilter = {},
    page = 1,
    pageSize = 50
  ): Promise<AuditApiResponse<AuditLogPage>> {
    const params = new URLSearchParams();
    params.set('page', page.toString());
    params.set('pageSize', pageSize.toString());

    if (filter.startDate) {
      params.set('startDate', filter.startDate.toISOString());
    }
    if (filter.endDate) {
      params.set('endDate', filter.endDate.toISOString());
    }
    if (filter.actions?.length) {
      params.set('actions', filter.actions.join(','));
    }
    if (filter.actorId) {
      params.set('actorId', filter.actorId);
    }
    if (filter.resourceType) {
      params.set('resourceType', filter.resourceType);
    }
    if (filter.resourceId) {
      params.set('resourceId', filter.resourceId);
    }
    if (filter.search) {
      params.set('search', filter.search);
    }

    return this.request<AuditLogPage>(`/audit/logs?${params.toString()}`);
  }

  async getAuditLogEntry(id: string): Promise<AuditApiResponse<AuditLogEntry>> {
    return this.request<AuditLogEntry>(`/audit/logs/${id}`);
  }

  async exportAuditLogs(filter: AuditLogFilter, format: 'json' | 'csv'): Promise<AuditApiResponse<{ downloadUrl: string }>> {
    return this.request<{ downloadUrl: string }>('/audit/logs/export', {
      method: 'POST',
      body: JSON.stringify({ filter, format }),
    });
  }

  // Export Jobs
  async listExportJobs(): Promise<AuditApiResponse<{ jobs: ExportJob[] }>> {
    return this.request<{ jobs: ExportJob[] }>('/exports');
  }

  async getExportJob(id: string): Promise<AuditApiResponse<ExportJob>> {
    return this.request<ExportJob>(`/exports/${id}`);
  }

  async createExportJob(request: CreateExportRequest): Promise<AuditApiResponse<ExportJob>> {
    return this.request<ExportJob>('/exports', {
      method: 'POST',
      body: JSON.stringify({
        type: request.type,
        format: request.format,
        start_date: request.startDate?.toISOString(),
        end_date: request.endDate?.toISOString(),
      }),
    });
  }

  async cancelExportJob(id: string): Promise<AuditApiResponse<{ success: boolean }>> {
    return this.request<{ success: boolean }>(`/exports/${id}/cancel`, {
      method: 'POST',
    });
  }

  async downloadExport(id: string): Promise<void> {
    const headers: HeadersInit = {};
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE_URL}/exports/${id}/download`, {
      headers,
    });

    if (!response.ok) {
      throw new Error('Failed to download export');
    }

    const blob = await response.blob();
    const disposition = response.headers.get('Content-Disposition');
    const filenameMatch = disposition?.match(/filename="?([^"]+)"?/);
    const filename = filenameMatch?.[1] || `export-${id}.zip`;

    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  }

  // Compliance Reports
  async listReports(): Promise<AuditApiResponse<{ reports: ComplianceReport[] }>> {
    return this.request<{ reports: ComplianceReport[] }>('/compliance/reports');
  }

  async getReport(id: string): Promise<AuditApiResponse<ComplianceReport>> {
    return this.request<ComplianceReport>(`/compliance/reports/${id}`);
  }

  async generateReport(request: GenerateReportRequest): Promise<AuditApiResponse<ComplianceReport>> {
    return this.request<ComplianceReport>('/compliance/reports', {
      method: 'POST',
      body: JSON.stringify({
        type: request.type,
        format: request.format,
        start_date: request.startDate.toISOString(),
        end_date: request.endDate.toISOString(),
      }),
    });
  }

  async downloadReport(id: string): Promise<void> {
    const headers: HeadersInit = {};
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE_URL}/compliance/reports/${id}/download`, {
      headers,
    });

    if (!response.ok) {
      throw new Error('Failed to download report');
    }

    const blob = await response.blob();
    const disposition = response.headers.get('Content-Disposition');
    const filenameMatch = disposition?.match(/filename="?([^"]+)"?/);
    const filename = filenameMatch?.[1] || `report-${id}.pdf`;

    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  }

  // Retention Policies
  async getRetentionPolicies(): Promise<AuditApiResponse<{ policies: RetentionPolicy[] }>> {
    return this.request<{ policies: RetentionPolicy[] }>('/compliance/retention');
  }

  async updateRetentionPolicy(update: RetentionPolicyUpdate): Promise<AuditApiResponse<RetentionPolicy>> {
    return this.request<RetentionPolicy>('/compliance/retention', {
      method: 'PUT',
      body: JSON.stringify({
        data_type: update.dataType,
        retention_days: update.retentionDays,
        legal_hold: update.legalHold,
      }),
    });
  }

  async setLegalHold(dataType: string, enabled: boolean): Promise<AuditApiResponse<{ success: boolean }>> {
    return this.request<{ success: boolean }>('/compliance/retention/legal-hold', {
      method: 'POST',
      body: JSON.stringify({ data_type: dataType, enabled }),
    });
  }
}

export const auditService = new AuditService();
