import { useState, useEffect, useCallback } from 'react';
import {
  FileText,
  Plus,
  RefreshCw,
  Calendar,
  Download,
  Eye,
  Shield,
  Activity,
  Lock,
  BarChart3,
  AlertTriangle,
  X,
  ExternalLink,
} from 'lucide-react';
import { useAuthStore } from '../../store/authStore';
import { useAuditStore } from '../../store/auditStore';
import { auditService } from '../../services/audit';
import { RetentionPolicy } from '../../components/settings/RetentionPolicy';
import type { ReportType, ReportFormat, ComplianceReport, GenerateReportRequest } from '../../types/audit';

const reportTypes: {
  value: ReportType;
  label: string;
  description: string;
  icon: typeof Shield;
}[] = [
  {
    value: 'access',
    label: 'Access Report',
    description: 'User access patterns and authentication events',
    icon: Lock,
  },
  {
    value: 'changes',
    label: 'Changes Report',
    description: 'Data modifications and configuration changes',
    icon: Activity,
  },
  {
    value: 'security',
    label: 'Security Report',
    description: 'Security events, failed logins, and anomalies',
    icon: Shield,
  },
  {
    value: 'summary',
    label: 'Summary Report',
    description: 'Executive summary of all compliance metrics',
    icon: BarChart3,
  },
];

const reportFormats: { value: ReportFormat; label: string }[] = [
  { value: 'pdf', label: 'PDF' },
  { value: 'html', label: 'HTML' },
  { value: 'json', label: 'JSON' },
];

function formatDate(date: Date): string {
  return new Date(date).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

interface ReportCardProps {
  report: ComplianceReport;
  onDownload: (id: string) => void;
  onPreview: (report: ComplianceReport) => void;
}

function ReportCard({ report, onDownload, onPreview }: ReportCardProps) {
  const typeConfig = reportTypes.find((t) => t.value === report.type) || reportTypes[0];
  const Icon = typeConfig.icon;

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-editor-accent/10">
            <Icon size={20} className="text-editor-accent" />
          </div>
          <div>
            <h3 className="font-medium text-editor-text">{report.title}</h3>
            <p className="text-sm text-editor-muted">{typeConfig.description}</p>
          </div>
        </div>
        <span className="px-2 py-1 text-xs font-medium bg-editor-bg border border-editor-border rounded uppercase">
          {report.format}
        </span>
      </div>

      <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <span className="text-editor-muted block">Period</span>
          <span className="text-editor-text">
            {formatDate(report.period.start)} - {formatDate(report.period.end)}
          </span>
        </div>
        <div>
          <span className="text-editor-muted block">Generated</span>
          <span className="text-editor-text">
            {new Date(report.generatedAt).toLocaleString(undefined, {
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </span>
        </div>
        {report.metrics && (
          <>
            <div>
              <span className="text-editor-muted block">Total Actions</span>
              <span className="text-editor-text">
                {report.metrics.totalActions?.toLocaleString() || 'N/A'}
              </span>
            </div>
            <div>
              <span className="text-editor-muted block">Active Users</span>
              <span className="text-editor-text">
                {report.metrics.activeUsers?.toLocaleString() || 'N/A'}
              </span>
            </div>
          </>
        )}
      </div>

      <div className="mt-4 flex items-center justify-end gap-2 pt-3 border-t border-editor-border">
        {report.previewHtml && (
          <button
            onClick={() => onPreview(report)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text"
          >
            <Eye size={14} />
            Preview
          </button>
        )}
        {report.downloadUrl && (
          <button
            onClick={() => onDownload(report.id)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90"
          >
            <Download size={14} />
            Download
          </button>
        )}
      </div>
    </div>
  );
}

export default function ComplianceReports() {
  const { accessToken } = useAuthStore();
  const {
    reports,
    reportsLoading,
    reportsError,
    setReports,
    addReport,
    setReportsLoading,
    setReportsError,
  } = useAuditStore();

  const [showNewReport, setShowNewReport] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const [previewReport, setPreviewReport] = useState<ComplianceReport | null>(null);
  const [selectedType, setSelectedType] = useState<ReportType>('summary');
  const [selectedFormat, setSelectedFormat] = useState<ReportFormat>('pdf');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [generateError, setGenerateError] = useState<string | null>(null);

  useEffect(() => {
    if (accessToken) {
      auditService.setToken(accessToken);
      fetchReports();
    }
  }, [accessToken]);

  const fetchReports = useCallback(async () => {
    setReportsLoading(true);
    setReportsError(null);

    const response = await auditService.listReports();
    if (response.error) {
      setReportsError(response.error);
    } else if (response.data) {
      setReports(response.data.reports);
    }
    setReportsLoading(false);
  }, [setReports, setReportsLoading, setReportsError]);

  const handleGenerateReport = async () => {
    if (!startDate || !endDate) {
      setGenerateError('Please select a date range');
      return;
    }

    setIsGenerating(true);
    setGenerateError(null);

    const request: GenerateReportRequest = {
      type: selectedType,
      format: selectedFormat,
      startDate: new Date(startDate),
      endDate: new Date(endDate),
    };

    const response = await auditService.generateReport(request);
    if (response.error) {
      setGenerateError(response.error);
    } else if (response.data) {
      addReport(response.data);
      setShowNewReport(false);
      resetForm();
    }
    setIsGenerating(false);
  };

  const handleDownload = async (id: string) => {
    try {
      await auditService.downloadReport(id);
    } catch {
      setReportsError('Failed to download report');
    }
  };

  const handlePreview = (report: ComplianceReport) => {
    setPreviewReport(report);
    setShowPreview(true);
  };

  const resetForm = () => {
    setSelectedType('summary');
    setSelectedFormat('pdf');
    setStartDate('');
    setEndDate('');
    setGenerateError(null);
  };

  // Set default date range to last 30 days
  useEffect(() => {
    if (showNewReport && !startDate && !endDate) {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 30);
      setStartDate(start.toISOString().split('T')[0]);
      setEndDate(end.toISOString().split('T')[0]);
    }
  }, [showNewReport, startDate, endDate]);

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-editor-text">Compliance Reports</h1>
            <p className="text-editor-muted mt-1">
              Generate and view compliance reports for your organization
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={fetchReports}
              disabled={reportsLoading}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg transition-colors disabled:opacity-50"
            >
              <RefreshCw size={16} className={reportsLoading ? 'animate-spin' : ''} />
              Refresh
            </button>
            <button
              onClick={() => setShowNewReport(true)}
              className="flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              <Plus size={16} />
              Generate Report
            </button>
          </div>
        </div>

        {/* Error Message */}
        {reportsError && (
          <div className="flex items-center gap-2 p-4 bg-editor-error/10 border border-editor-error/20 rounded-lg">
            <AlertTriangle size={18} className="text-editor-error" />
            <span className="text-editor-error">{reportsError}</span>
          </div>
        )}

        {/* Generate Report Modal */}
        {showNewReport && (
          <>
            <div
              className="fixed inset-0 bg-black/50 z-40"
              onClick={() => {
                setShowNewReport(false);
                resetForm();
              }}
            />
            <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-semibold text-editor-text">Generate Report</h2>
                <button
                  onClick={() => {
                    setShowNewReport(false);
                    resetForm();
                  }}
                  className="p-1 hover:bg-editor-surface rounded"
                >
                  <X size={20} className="text-editor-muted" />
                </button>
              </div>

              <div className="space-y-6">
                {/* Report Type Selection */}
                <div className="space-y-3">
                  <label className="text-sm font-medium text-editor-text">Report Type</label>
                  <div className="grid grid-cols-2 gap-3">
                    {reportTypes.map((type) => {
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
                    {reportFormats.map((format) => (
                      <button
                        key={format.value}
                        onClick={() => setSelectedFormat(format.value)}
                        className={`px-4 py-2 rounded-lg border transition-all ${
                          selectedFormat === format.value
                            ? 'bg-editor-accent/10 border-editor-accent text-editor-accent'
                            : 'bg-editor-surface border-editor-border text-editor-text hover:border-editor-muted'
                        }`}
                      >
                        {format.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Date Range */}
                <div className="space-y-3">
                  <label className="flex items-center gap-1.5 text-sm font-medium text-editor-text">
                    <Calendar size={14} />
                    Report Period
                  </label>
                  <div className="flex items-center gap-3">
                    <input
                      type="date"
                      value={startDate}
                      onChange={(e) => setStartDate(e.target.value)}
                      className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text"
                    />
                    <span className="text-editor-muted">to</span>
                    <input
                      type="date"
                      value={endDate}
                      onChange={(e) => setEndDate(e.target.value)}
                      className="flex-1 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text"
                    />
                  </div>
                </div>

                {/* Error */}
                {generateError && (
                  <div className="flex items-center gap-2 p-3 bg-editor-error/10 border border-editor-error/20 rounded-lg">
                    <AlertTriangle size={16} className="text-editor-error" />
                    <span className="text-sm text-editor-error">{generateError}</span>
                  </div>
                )}

                {/* Actions */}
                <div className="flex items-center justify-end gap-3 pt-4 border-t border-editor-border">
                  <button
                    onClick={() => {
                      setShowNewReport(false);
                      resetForm();
                    }}
                    className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleGenerateReport}
                    disabled={isGenerating || !startDate || !endDate}
                    className="flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50"
                  >
                    {isGenerating ? (
                      <RefreshCw size={16} className="animate-spin" />
                    ) : (
                      <FileText size={16} />
                    )}
                    {isGenerating ? 'Generating...' : 'Generate Report'}
                  </button>
                </div>
              </div>
            </div>
          </>
        )}

        {/* Preview Modal */}
        {showPreview && previewReport && (
          <>
            <div
              className="fixed inset-0 bg-black/50 z-40"
              onClick={() => setShowPreview(false)}
            />
            <div className="fixed top-4 bottom-4 left-1/2 -translate-x-1/2 w-[800px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 flex flex-col">
              <div className="flex items-center justify-between p-4 border-b border-editor-border">
                <h2 className="text-lg font-semibold text-editor-text">{previewReport.title}</h2>
                <div className="flex items-center gap-2">
                  {previewReport.downloadUrl && (
                    <button
                      onClick={() => handleDownload(previewReport.id)}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90"
                    >
                      <Download size={14} />
                      Download
                    </button>
                  )}
                  <button
                    onClick={() => setShowPreview(false)}
                    className="p-1 hover:bg-editor-surface rounded"
                  >
                    <X size={20} className="text-editor-muted" />
                  </button>
                </div>
              </div>
              <div className="flex-1 overflow-auto p-6">
                {previewReport.previewHtml ? (
                  <div
                    className="prose prose-invert max-w-none"
                    dangerouslySetInnerHTML={{ __html: previewReport.previewHtml }}
                  />
                ) : (
                  <div className="text-center py-12 text-editor-muted">
                    <Eye size={48} className="mx-auto mb-4 opacity-50" />
                    <p>Preview not available for this report format.</p>
                    <button
                      onClick={() => handleDownload(previewReport.id)}
                      className="mt-4 flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 mx-auto"
                    >
                      <ExternalLink size={14} />
                      Download to View
                    </button>
                  </div>
                )}
              </div>
            </div>
          </>
        )}

        {/* Reports List */}
        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-editor-text">Generated Reports</h2>

          {reportsLoading && reports.length === 0 ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="h-40 bg-editor-surface border border-editor-border rounded-lg animate-pulse"
                />
              ))}
            </div>
          ) : reports.length === 0 ? (
            <div className="text-center py-16 bg-editor-surface border border-editor-border rounded-lg">
              <FileText size={48} className="mx-auto mb-4 text-editor-muted opacity-50" />
              <h3 className="text-lg font-medium text-editor-text">No reports yet</h3>
              <p className="text-editor-muted mt-1">
                Generate your first compliance report to get started
              </p>
              <button
                onClick={() => setShowNewReport(true)}
                className="mt-4 flex items-center gap-2 px-4 py-2 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 mx-auto"
              >
                <Plus size={16} />
                Generate Report
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              {reports.map((report) => (
                <ReportCard
                  key={report.id}
                  report={report}
                  onDownload={handleDownload}
                  onPreview={handlePreview}
                />
              ))}
            </div>
          )}
        </section>

        {/* Data Retention Section */}
        <section className="pt-6 border-t border-editor-border">
          <RetentionPolicy />
        </section>
      </div>
    </div>
  );
}
