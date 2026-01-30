import { useState, useEffect } from 'react';
import {
  Download,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  FileJson,
  FileSpreadsheet,
  AlertTriangle,
  Trash2,
} from 'lucide-react';
import type { ExportJob, ExportType, ExportFormat, ExportJobStatus } from '../../types/audit';

interface ExportJobCardProps {
  job: ExportJob;
  onDownload: (id: string) => void;
  onCancel?: (id: string) => void;
}

const exportTypeLabels: Record<ExportType, string> = {
  gdpr: 'GDPR Data Export',
  audit: 'Audit Logs Export',
  usage: 'Usage Data Export',
  conversations: 'Conversations Export',
};

const exportTypeDescriptions: Record<ExportType, string> = {
  gdpr: 'Complete data export for GDPR compliance',
  audit: 'Audit trail and activity logs',
  usage: 'Usage metrics and analytics data',
  conversations: 'Chat conversations and messages',
};

const formatIcons: Record<ExportFormat, typeof FileJson> = {
  json: FileJson,
  csv: FileSpreadsheet,
};

function formatFileSize(bytes?: number): string {
  if (!bytes) return 'Unknown size';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

function formatDateRange(start?: Date, end?: Date): string {
  if (!start && !end) return 'All time';
  const format = (d: Date) =>
    new Date(d).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  if (start && end) return `${format(start)} - ${format(end)}`;
  if (start) return `From ${format(start)}`;
  if (end) return `Until ${format(end)}`;
  return 'All time';
}

function useCountdown(targetDate?: Date): string | null {
  const [timeLeft, setTimeLeft] = useState<string | null>(null);

  useEffect(() => {
    if (!targetDate) return;

    const updateCountdown = () => {
      const now = new Date();
      const target = new Date(targetDate);
      const diff = target.getTime() - now.getTime();

      if (diff <= 0) {
        setTimeLeft('Expired');
        return;
      }

      const hours = Math.floor(diff / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

      if (hours > 24) {
        const days = Math.floor(hours / 24);
        setTimeLeft(`${days}d ${hours % 24}h`);
      } else if (hours > 0) {
        setTimeLeft(`${hours}h ${minutes}m`);
      } else {
        setTimeLeft(`${minutes}m`);
      }
    };

    updateCountdown();
    const interval = setInterval(updateCountdown, 60000);
    return () => clearInterval(interval);
  }, [targetDate]);

  return timeLeft;
}

export function ExportJobCard({ job, onDownload, onCancel }: ExportJobCardProps) {
  const expiryCountdown = useCountdown(job.expiresAt);
  const FormatIcon = formatIcons[job.format];

  const statusConfig: Record<
    ExportJobStatus,
    {
      icon: typeof Clock;
      color: string;
      bgColor: string;
      label: string;
      animate?: boolean;
    }
  > = {
    pending: {
      icon: Clock,
      color: 'text-editor-muted',
      bgColor: 'bg-editor-surface',
      label: 'Pending',
    },
    processing: {
      icon: Loader2,
      color: 'text-editor-accent',
      bgColor: 'bg-editor-accent/10',
      label: 'Processing',
      animate: true,
    },
    complete: {
      icon: CheckCircle,
      color: 'text-editor-success',
      bgColor: 'bg-editor-success/10',
      label: 'Complete',
    },
    failed: {
      icon: XCircle,
      color: 'text-editor-error',
      bgColor: 'bg-editor-error/10',
      label: 'Failed',
    },
    expired: {
      icon: AlertTriangle,
      color: 'text-editor-warning',
      bgColor: 'bg-editor-warning/10',
      label: 'Expired',
    },
  };

  const status = statusConfig[job.status];
  const StatusIcon = status.icon;

  const isExpiringSoon =
    job.expiresAt &&
    job.status === 'complete' &&
    new Date(job.expiresAt).getTime() - Date.now() < 24 * 60 * 60 * 1000;

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg p-4 space-y-4">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg ${status.bgColor}`}>
            <StatusIcon
              size={20}
              className={`${status.color} ${status.animate ? 'animate-spin' : ''}`}
            />
          </div>
          <div>
            <h3 className="font-medium text-editor-text">
              {exportTypeLabels[job.type]}
            </h3>
            <p className="text-sm text-editor-muted">
              {exportTypeDescriptions[job.type]}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className={`inline-flex items-center gap-1.5 px-2 py-1 text-xs font-medium rounded-full ${status.bgColor} ${status.color}`}>
            {status.label}
          </span>
          <div className="flex items-center gap-1 px-2 py-1 bg-editor-bg rounded text-xs text-editor-muted">
            <FormatIcon size={12} />
            <span className="uppercase">{job.format}</span>
          </div>
        </div>
      </div>

      {/* Progress Bar (for processing jobs) */}
      {job.status === 'processing' && (
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-editor-muted">Progress</span>
            <span className="text-editor-text">{job.progress}%</span>
          </div>
          <div className="h-2 bg-editor-bg rounded-full overflow-hidden">
            <div
              className="h-full bg-editor-accent rounded-full transition-all duration-300"
              style={{ width: `${job.progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Details */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <span className="text-editor-muted block">Date Range</span>
          <span className="text-editor-text">
            {formatDateRange(job.filters?.startDate, job.filters?.endDate)}
          </span>
        </div>
        <div>
          <span className="text-editor-muted block">Created</span>
          <span className="text-editor-text">
            {new Date(job.createdAt).toLocaleString(undefined, {
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </span>
        </div>
        {job.status === 'complete' && (
          <>
            <div>
              <span className="text-editor-muted block">File Size</span>
              <span className="text-editor-text">{formatFileSize(job.fileSize)}</span>
            </div>
            <div>
              <span className="text-editor-muted block">Expires In</span>
              <span className={isExpiringSoon ? 'text-editor-warning' : 'text-editor-text'}>
                {expiryCountdown || 'N/A'}
              </span>
            </div>
          </>
        )}
        {job.status === 'failed' && job.error && (
          <div className="col-span-2">
            <span className="text-editor-muted block">Error</span>
            <span className="text-editor-error">{job.error}</span>
          </div>
        )}
      </div>

      {/* Expiry Warning */}
      {isExpiringSoon && (
        <div className="flex items-center gap-2 p-3 bg-editor-warning/10 border border-editor-warning/20 rounded-lg">
          <AlertTriangle size={16} className="text-editor-warning flex-shrink-0" />
          <span className="text-sm text-editor-warning">
            Download link expires soon. Download now to avoid losing access.
          </span>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center justify-between pt-2 border-t border-editor-border">
        <span className="text-xs text-editor-muted">
          Requested by {job.requestedBy.email}
        </span>
        <div className="flex items-center gap-2">
          {(job.status === 'pending' || job.status === 'processing') && onCancel && (
            <button
              onClick={() => onCancel(job.id)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
            >
              <Trash2 size={14} />
              Cancel
            </button>
          )}
          {job.status === 'complete' && job.downloadUrl && (
            <button
              onClick={() => onDownload(job.id)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
            >
              <Download size={14} />
              Download
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
