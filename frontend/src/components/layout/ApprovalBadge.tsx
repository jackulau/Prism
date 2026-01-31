import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bell, Check } from 'lucide-react';
import { usePendingCount, useApprovalStore } from '../../store/approvalStore';
import { toast } from '../../store/toastStore';

interface ApprovalBadgeProps {
  showLabel?: boolean;
  className?: string;
}

export function ApprovalBadge({ showLabel = false, className = '' }: ApprovalBadgeProps) {
  const navigate = useNavigate();
  const pendingCount = usePendingCount();
  const loadPendingApprovals = useApprovalStore(state => state.loadPendingApprovals);

  // Load pending approvals on mount and set up polling
  useEffect(() => {
    loadPendingApprovals();

    // Poll every 30 seconds for updates
    const interval = setInterval(() => {
      loadPendingApprovals();
    }, 30000);

    return () => clearInterval(interval);
  }, [loadPendingApprovals]);

  const handleClick = () => {
    navigate('/approvals');
  };

  return (
    <button
      onClick={handleClick}
      className={`relative flex items-center gap-2 p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface transition-colors ${className}`}
      title={pendingCount > 0 ? `${pendingCount} pending approvals` : 'No pending approvals'}
    >
      <Bell size={18} />
      {showLabel && <span className="text-sm">Approvals</span>}

      {/* Badge count */}
      {pendingCount > 0 && (
        <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] flex items-center justify-center px-1 text-xs font-medium text-white bg-editor-error rounded-full animate-pulse-subtle">
          {pendingCount > 99 ? '99+' : pendingCount}
        </span>
      )}
    </button>
  );
}

// Dropdown version for navigation
interface ApprovalDropdownProps {
  isCollapsed?: boolean;
}

export function ApprovalDropdown({ isCollapsed = false }: ApprovalDropdownProps) {
  const navigate = useNavigate();
  const pendingCount = usePendingCount();
  const pendingApprovals = useApprovalStore(state => state.pendingApprovals);
  const loadPendingApprovals = useApprovalStore(state => state.loadPendingApprovals);
  const submitDecision = useApprovalStore(state => state.submitDecision);

  useEffect(() => {
    loadPendingApprovals();
    const interval = setInterval(loadPendingApprovals, 30000);
    return () => clearInterval(interval);
  }, [loadPendingApprovals]);

  const handleQuickApprove = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    const success = await submitDecision(id, 'approved');
    if (success) {
      toast.success('Request approved');
    } else {
      toast.error('Failed to approve request');
    }
  };

  if (isCollapsed) {
    return (
      <button
        onClick={() => navigate('/approvals')}
        className="relative w-full flex items-center justify-center p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
        title="Approvals"
      >
        <Bell size={18} />
        {pendingCount > 0 && (
          <span className="absolute top-1 right-2 min-w-[16px] h-[16px] flex items-center justify-center px-1 text-[10px] font-medium text-white bg-editor-error rounded-full">
            {pendingCount > 9 ? '9+' : pendingCount}
          </span>
        )}
      </button>
    );
  }

  return (
    <div className="px-2">
      <button
        onClick={() => navigate('/approvals')}
        className="w-full flex items-center justify-between gap-2 px-3 py-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
      >
        <div className="flex items-center gap-3">
          <Bell size={18} />
          <span className="text-sm font-medium">Approvals</span>
        </div>
        {pendingCount > 0 && (
          <span className="min-w-[20px] h-[20px] flex items-center justify-center px-1.5 text-xs font-medium text-white bg-editor-error rounded-full">
            {pendingCount > 99 ? '99+' : pendingCount}
          </span>
        )}
      </button>

      {/* Quick preview of pending approvals */}
      {pendingCount > 0 && pendingApprovals.length > 0 && (
        <div className="mt-2 space-y-1">
          {pendingApprovals.slice(0, 3).map((approval) => (
            <div
              key={approval.id}
              onClick={() => navigate('/approvals')}
              className="flex items-center justify-between gap-2 px-3 py-2 rounded-lg bg-editor-surface/50 hover:bg-editor-surface cursor-pointer transition-colors group"
            >
              <div className="flex-1 min-w-0">
                <p className="text-xs text-editor-text truncate">{approval.title}</p>
                <p className="text-[10px] text-editor-muted">
                  {approval.requestedBy.name || approval.requestedBy.email}
                </p>
              </div>
              <button
                onClick={(e) => handleQuickApprove(e, approval.id)}
                className="p-1 rounded opacity-0 group-hover:opacity-100 text-editor-success hover:bg-editor-success/10 transition-all"
                title="Quick approve"
              >
                <Check size={14} />
              </button>
            </div>
          ))}
          {pendingCount > 3 && (
            <p className="px-3 py-1 text-[10px] text-editor-muted text-center">
              +{pendingCount - 3} more
            </p>
          )}
        </div>
      )}
    </div>
  );
}

// Hook for showing toast notifications when new approvals arrive
export function useApprovalNotifications() {
  const pendingApprovals = useApprovalStore(state => state.pendingApprovals);

  useEffect(() => {
    if (pendingApprovals.length === 0) return;

    // Check for new approvals (created within last 30 seconds)
    const thirtySecondsAgo = new Date(Date.now() - 30000);
    const newApprovals = pendingApprovals.filter(
      a => a.requestedAt > thirtySecondsAgo
    );

    newApprovals.forEach((approval) => {
      toast.info(`New approval request: ${approval.title}`);

      // Request desktop notification if supported and permitted
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification('New Approval Request', {
          body: approval.title,
          icon: '/logo.png',
          tag: `approval-${approval.id}`,
        });
      }
    });
  }, [pendingApprovals]);
}

// Request notification permission
export function requestNotificationPermission() {
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission();
  }
}
