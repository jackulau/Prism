import { useState, useEffect, useCallback } from 'react';
import {
  Download,
  Plus,
  RefreshCw,
  FileJson,
  FileSpreadsheet,
  Calendar,
  Shield,
  FileText,
  BarChart3,
  MessageSquare,
  AlertTriangle,
  X,
} from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useAuditStore } from '../../store/auditStore';
import { auditService } from '../../services/audit';
import { ExportJobCard } from '../../components/export/ExportJobCard';
import type { ExportType, ExportFormat, CreateExportRequest } from '../../types/audit';

const exportTypes: {
  value: ExportType;
  label: string;
  description: string;
  icon: typeof Shield;
}[] = [
  {
    value: 'gdpr',
    label: 'GDPR Data Export',
    description: 'Complete export of all personal data for GDPR compliance',
    icon: Shield,
  },
  {
    value: 'audit',
    label: 'Audit Logs',
    description: 'Export audit trail and activity logs',
    icon: FileText,
  },
  {
    value: 'usage',
    label: 'Usage Data',
    description: 'Export usage metrics and analytics data',
    icon: BarChart3,
  },
  {
    value: 'conversations',
    label: 'Conversations',
    description: 'Export chat conversations and messages',
    icon: MessageSquare,
  },
];

const exportFormats: { value: ExportFormat; label: string; icon: typeof FileJson }[] = [
  { value: 'json', label: 'JSON', icon: FileJson },
  { value: 'csv', label: 'CSV', icon: FileSpreadsheet },
];

export default function DataExport() {
  const { accessToken } = useAuthStore();
  const {
    exportJobs,
    exportJobsLoading,
    exportJobsError,
    setExportJobs,
    addExportJob,
    updateExportJob,
    setExportJobsLoading,
    setExportJobsError,
  } = useAuditStore();

  const [showNewExport, setShowNewExport] = useState(false);
  const [selectedType, setSelectedType] = useState<ExportType>('gdpr');
  const [selectedFormat, setSelectedFormat] = useState<ExportFormat>('json');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  useEffect(() => {
    if (accessToken) {
      auditService.setToken(accessToken);
      fetchExportJobs();
    }
  }, [accessToken]);

  // Poll for job status updates
  useEffect(() => {
    const processingJobs = exportJobs.filter(
      (job) => job.status === 'pending' || job.status === 'processing'
    );

    if (processingJobs.length === 0) return;

    const interval = setInterval(async () => {
      for (const job of processingJobs) {
        const response = await auditService.getExportJob(job.id);
        if (response.data) {
          updateExportJob(job.id, response.data);
        }
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [exportJobs, updateExportJob]);

  const fetchExportJobs = useCallback(async () => {
    setExportJobsLoading(true);
    setExportJobsError(null);

    const response = await auditService.listExportJobs();
    if (response.error) {
      setExportJobsError(response.error);
    } else if (response.data) {
      setExportJobs(response.data.jobs);
    }
    setExportJobsLoading(false);
  }, [setExportJobs, setExportJobsLoading, setExportJobsError]);

  const handleCreateExport = async () => {
    setIsCreating(true);
    setCreateError(null);

    const request: CreateExportRequest = {
      type: selectedType,
      format: selectedFormat,
      startDate: startDate ? new Date(startDate) : undefined,
      endDate: endDate ? new Date(endDate) : undefined,
    };

    const response = await auditService.createExportJob(request);
    if (response.error) {
      setCreateError(response.error);
    } else if (response.data) {
      addExportJob(response.data);
      setShowNewExport(false);
      resetForm();
    }
    setIsCreating(false);
  };

  const handleDownload = async (id: string) => {
    try {
      await auditService.downloadExport(id);
    } catch {
      setExportJobsError('Failed to download export');
    }
  };

  const handleCancel = async (id: string) => {
    const response = await auditService.cancelExportJob(id);
    if (response.data?.success) {
      updateExportJob(id, { status: 'failed', error: 'Cancelled by user' });
    }
  };

  const resetForm = () => {
    setSelectedType('gdpr');
    setSelectedFormat('json');
    setStartDate('');
    setEndDate('');
    setCreateError(null);
  };

  const pendingJobs = exportJobs.filter(
    (job) => job.status === 'pending' || job.status === 'processing'
  );
  const completedJobs = exportJobs.filter((job) => job.status === 'complete');
  const failedJobs = exportJobs.filter((job) => job.status === 'failed' || job.status === 'expired');

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-editor-text">Data Export</h1>
            <p className="text-editor-muted mt-1">
              Export your data for compliance, backup, or portability
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={fetchExportJobs}
              disabled={exportJobsLoading}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg transition-colors disabled:opacity-50"
            >
              <RefreshCw size={16} className={exportJobsLoading ? 'animate-spin' : ''} />
              Refresh
            </button>
            <button
              onClick={() => setShowNewExport(true)}
              className="flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              <Plus size={16} />
              New Export
            </button>
          </div>
        </div>

        {/* Error Message */}
        {exportJobsError && (
          <div className="flex items-center gap-2 p-4 bg-editor-error/10 border border-editor-error/20 rounded-lg">
            <AlertTriangle size={18} className="text-editor-error" />
            <span className="text-editor-error">{exportJobsError}</span>
          </div>
        )}

        {/* New Export Modal */}
        {showNewExport && (
          <>
            <div
              className="fixed inset-0 bg-black/50 z-40"
              onClick={() => {
                setShowNewExport(false);
                resetForm();
              }}
            />
            <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-semibold text-editor-text">Create New Export</h2>
                <button
                  onClick={() => {
                    setShowNewExport(false);
                    resetForm();
                  }}
                  className="p-1 hover:bg-editor-surface rounded"
                >
                  <X size={20} className="text-editor-muted" />
                </button>
              </div>

              <div className="space-y-6">
                {/* Export Type Selection */}
                <div className="space-y-3">
                  <label className="text-sm font-medium text-editor-text">Export Type</label>
                  <div className="grid grid-cols-2 gap-3">
                    {exportTypes.map((type) => {
                      const Icon = type.icon;
                      return (
                        <button
                          key={type.value}
                          onClick={() => setSelectedType(type.value)}
                          className={`p-4 rounded-lg border text-left transition-all ${
                            selectedType === type.value
                              ? 'bg-editor-accent/10 border-editor-accent'
                              : 'bg-editor-surface border-editor-border hover:border-editor-muted'
                          }`}
                        >
                          <Icon
                            size={20}
                            className={
                              selectedType === type.value
                                ? 'text-editor-accent'
                                : 'text-editor-muted'
                            }
                          />
                          <p className="font-medium text-editor-text mt-2">{type.label}</p>
                          <p className="text-xs text-editor-muted mt-1">{type.description}</p>
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* Format Selection */}
                <div className="space-y-3">
                  <label className="text-sm font-medium text-editor-text">Format</label>
                  <div className="flex gap-3">
                    {exportFormats.map((format) => {
                      const Icon = format.icon;
                      return (
                        <button
                          key={format.value}
                          onClick={() => setSelectedFormat(format.value)}
                          className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-all ${
                            selectedFormat === format.value
                              ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                              : 'bg-editor-surface border-editor-border text-editor-text hover:border-editor-muted'
                          }`}
                        >
                          <Icon size={16} />
                          {format.label}
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* Date Range */}
                <div className="space-y-3">
                  <label className="flex items-center gap-1.5 text-sm font-medium text-editor-text">
                    <Calendar size={14} />
                    Date Range (Optional)
                  </label>
                  <div className="flex items-center gap-3">
                    <input
                      type="date"
                      value={startDate}
                      onChange={(e) => setStartDate(e.target.value)}
                      className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text"
                      placeholder="Start date"
                    />
                    <span className="text-editor-muted">to</span>
                    <input
                      type="date"
                      value={endDate}
                      onChange={(e) => setEndDate(e.target.value)}
                      className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text"
                      placeholder="End date"
                    />
                  </div>
                  <p className="text-xs text-editor-muted">
                    Leave empty to export all available data
                  </p>
                </div>

                {/* Error */}
                {createError && (
                  <div className="flex items-center gap-2 p-3 bg-editor-error/10 border border-editor-error/20 rounded-lg">
                    <AlertTriangle size={16} className="text-editor-error" />
                    <span className="text-sm text-editor-error">{createError}</span>
                  </div>
                )}

                {/* Actions */}
                <div className="flex items-center justify-end gap-3 pt-4 border-t border-editor-border">
                  <button
                    onClick={() => {
                      setShowNewExport(false);
                      resetForm();
                    }}
                    className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleCreateExport}
                    disabled={isCreating}
                    className="flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50"
                  >
                    {isCreating ? (
                      <RefreshCw size={16} className="animate-spin" />
                    ) : (
                      <Download size={16} />
                    )}
                    {isCreating ? 'Creating...' : 'Create Export'}
                  </button>
                </div>
              </div>
            </div>
          </>
        )}

        {/* In Progress Jobs */}
        {pendingJobs.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-lg font-semibold text-editor-text flex items-center gap-2">
              <RefreshCw size={18} className="text-editor-accent animate-spin" />
              In Progress ({pendingJobs.length})
            </h2>
            <div className="space-y-3">
              {pendingJobs.map((job) => (
                <ExportJobCard
                  key={job.id}
                  job={job}
                  onDownload={handleDownload}
                  onCancel={handleCancel}
                />
              ))}
            </div>
          </section>
        )}

        {/* Completed Exports */}
        {completedJobs.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-lg font-semibold text-editor-text">
              Ready for Download ({completedJobs.length})
            </h2>
            <div className="space-y-3">
              {completedJobs.map((job) => (
                <ExportJobCard
                  key={job.id}
                  job={job}
                  onDownload={handleDownload}
                />
              ))}
            </div>
          </section>
        )}

        {/* Failed/Expired Exports */}
        {failedJobs.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-lg font-semibold text-editor-muted">
              Failed or Expired ({failedJobs.length})
            </h2>
            <div className="space-y-3 opacity-75">
              {failedJobs.map((job) => (
                <ExportJobCard
                  key={job.id}
                  job={job}
                  onDownload={handleDownload}
                />
              ))}
            </div>
          </section>
        )}

        {/* Empty State */}
        {exportJobs.length === 0 && !exportJobsLoading && (
          <div className="text-center py-16 bg-editor-surface border border-editor-border rounded-lg">
            <Download size={48} className="mx-auto mb-4 text-editor-muted opacity-50" />
            <h3 className="text-lg font-medium text-editor-text">No exports yet</h3>
            <p className="text-editor-muted mt-1">
              Create your first data export to get started
            </p>
            <button
              onClick={() => setShowNewExport(true)}
              className="mt-4 flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 mx-auto"
            >
              <Plus size={16} />
              Create Export
            </button>
          </div>
        )}

        {/* Loading State */}
        {exportJobsLoading && exportJobs.length === 0 && (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-40 bg-editor-surface border border-editor-border rounded-lg animate-pulse"
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
