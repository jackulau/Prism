import { useState, useEffect } from 'react';
import {
  Check,
  X,
  Clock,
  AlertTriangle,
  MessageSquare,
  ChevronDown,
  ChevronUp,
  User,
  Wrench,
  Settings,
  Shield,
  Zap,
} from 'lucide-react';
import type { ApprovalRequest, ApprovalType } from '../../types/approval';

interface ApprovalCardProps {
  approval: ApprovalRequest;
  onApprove: (comment?: string) => void;
  onReject: (comment?: string) => void;
  onViewDetails: () => void;
}

const TYPE_ICONS: Record<ApprovalType, typeof Wrench> = {
  tool_execution: Wrench,
  config_change: Settings,
  deployment: Zap,
  access_request: Shield,
  custom: MessageSquare,
};

const TYPE_LABELS: Record<ApprovalType, string> = {
  tool_execution: 'Tool Execution',
  config_change: 'Config Change',
  deployment: 'Deployment',
  access_request: 'Access Request',
  custom: 'Custom',
};

function formatTimeAgo(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
}

function formatCountdown(expiresAt: Date): { text: string; isUrgent: boolean } {
  const now = new Date();
  const diffMs = expiresAt.getTime() - now.getTime();

  if (diffMs <= 0) return { text: 'Expired', isUrgent: true };

  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);

  if (diffMins < 30) return { text: `${diffMins}m left`, isUrgent: true };
  if (diffHours < 1) return { text: `${diffMins}m left`, isUrgent: false };
  if (diffHours < 24) return { text: `${diffHours}h left`, isUrgent: false };
  return { text: `${Math.floor(diffHours / 24)}d left`, isUrgent: false };
}

export function ApprovalCard({ approval, onApprove, onReject, onViewDetails }: ApprovalCardProps) {
  const [showComment, setShowComment] = useState(false);
  const [comment, setComment] = useState('');
  const [countdown, setCountdown] = useState<{ text: string; isUrgent: boolean } | null>(null);

  const TypeIcon = TYPE_ICONS[approval.type];

  // Update countdown every minute
  useEffect(() => {
    if (!approval.expiresAt) return;

    const updateCountdown = () => {
      setCountdown(formatCountdown(approval.expiresAt!));
    };

    updateCountdown();
    const interval = setInterval(updateCountdown, 60000);
    return () => clearInterval(interval);
  }, [approval.expiresAt]);

  const handleApprove = () => {
    onApprove(showComment ? comment : undefined);
    setComment('');
    setShowComment(false);
  };

  const handleReject = () => {
    onReject(showComment ? comment : undefined);
    setComment('');
    setShowComment(false);
  };

  return (
    <div className="bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors">
      {/* Header */}
      <div className="p-4 border-b border-editor-border/50">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            <div className="p-2 rounded-lg bg-editor-accent/10 text-editor-accent flex-shrink-0">
              <TypeIcon size={18} />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-medium text-editor-text truncate">{approval.title}</h3>
              <div className="flex items-center gap-2 mt-1 text-xs text-editor-muted">
                <span className="px-1.5 py-0.5 bg-editor-bg rounded">
                  {TYPE_LABELS[approval.type]}
                </span>
                {approval.currentStep && approval.totalSteps && (
                  <span>
                    Step {approval.currentStep}/{approval.totalSteps}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Countdown/Status */}
          {countdown && (
            <div
              className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs ${
                countdown.isUrgent
                  ? 'bg-editor-warning/10 text-editor-warning'
                  : 'bg-editor-muted/10 text-editor-muted'
              }`}
            >
              <Clock size={12} />
              {countdown.text}
            </div>
          )}

          {approval.status === 'escalated' && (
            <div className="flex items-center gap-1 px-2 py-1 bg-orange-500/10 text-orange-400 rounded-full text-xs">
              <AlertTriangle size={12} />
              Escalated
            </div>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-3">
        {/* Description */}
        {approval.description && (
          <p className="text-sm text-editor-muted line-clamp-2">{approval.description}</p>
        )}

        {/* Tool info for tool execution */}
        {approval.toolName && (
          <div className="flex items-center gap-2 text-sm">
            <Wrench size={14} className="text-editor-muted" />
            <code className="px-1.5 py-0.5 bg-editor-bg rounded text-editor-text">
              {approval.toolName}
            </code>
          </div>
        )}

        {/* Requestor info */}
        <div className="flex items-center gap-2 text-xs text-editor-muted">
          <User size={12} />
          <span>
            Requested by <span className="text-editor-text">{approval.requestedBy.name || approval.requestedBy.email}</span>
          </span>
          <span>·</span>
          <span>{formatTimeAgo(approval.requestedAt)}</span>
        </div>
      </div>

      {/* Comment input */}
      {showComment && (
        <div className="px-4 pb-3">
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Add a comment (optional)..."
            className="w-full px-3 py-2 bg-editor-bg border border-editor-border rounded-lg text-sm text-editor-text placeholder:text-editor-muted resize-none focus:outline-none focus:border-editor-accent"
            rows={2}
          />
        </div>
      )}

      {/* Actions */}
      <div className="px-4 pb-4 flex items-center justify-between gap-2">
        <button
          onClick={() => setShowComment(!showComment)}
          className="flex items-center gap-1 px-2 py-1.5 text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          <MessageSquare size={14} />
          {showComment ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          Comment
        </button>

        <div className="flex items-center gap-2">
          <button
            onClick={onViewDetails}
            className="px-3 py-1.5 text-sm text-editor-muted hover:text-editor-text hover:bg-editor-bg rounded-lg transition-colors"
          >
            Details
          </button>
          <button
            onClick={handleReject}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-editor-error hover:bg-editor-error/10 rounded-lg transition-colors"
          >
            <X size={14} />
            Reject
          </button>
          <button
            onClick={handleApprove}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-white bg-editor-success hover:bg-editor-success/80 rounded-lg transition-colors"
          >
            <Check size={14} />
            Approve
          </button>
        </div>
      </div>
    </div>
  );
}

// Compact version for lists and sidebar
interface ApprovalCardCompactProps {
  approval: ApprovalRequest;
  onClick: () => void;
}

export function ApprovalCardCompact({ approval, onClick }: ApprovalCardCompactProps) {
  const TypeIcon = TYPE_ICONS[approval.type];

  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-3 p-3 bg-editor-surface border border-editor-border rounded-lg hover:border-editor-accent/30 transition-colors text-left"
    >
      <div className="p-1.5 rounded bg-editor-accent/10 text-editor-accent">
        <TypeIcon size={14} />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-editor-text truncate">{approval.title}</p>
        <p className="text-xs text-editor-muted">
          {formatTimeAgo(approval.requestedAt)}
        </p>
      </div>
      {approval.status === 'escalated' && (
        <AlertTriangle size={14} className="text-orange-400" />
      )}
    </button>
  );
}
