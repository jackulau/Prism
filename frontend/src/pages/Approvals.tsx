import { useState, useEffect } from 'react';
import {
  Clock,
  History,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Loader2,
  RefreshCw,
  Filter,
  Search,
} from 'lucide-react';
import {
  useApprovalStore,
  usePendingApprovals,
  useApprovalHistory,
  useApprovalStats,
} from '../store/approvalStore';
import { ApprovalCard } from '../components/approval/ApprovalCard';
import { ApprovalDetail } from '../components/approval/ApprovalDetail';
import { toast } from '../store/toastStore';
import type { ApprovalRequest, ApprovalStatus, ApprovalType } from '../types/approval';

type TabType = 'pending' | 'history';

const STATUS_FILTERS: { value: ApprovalStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'expired', label: 'Expired' },
];

const TYPE_FILTERS: { value: ApprovalType | 'all'; label: string }[] = [
  { value: 'all', label: 'All Types' },
  { value: 'tool_execution', label: 'Tool Execution' },
  { value: 'config_change', label: 'Config Change' },
  { value: 'deployment', label: 'Deployment' },
  { value: 'access_request', label: 'Access Request' },
  { value: 'custom', label: 'Custom' },
];

export default function Approvals() {
  const [activeTab, setActiveTab] = useState<TabType>('pending');
  const [selectedApproval, setSelectedApproval] = useState<ApprovalRequest | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<ApprovalStatus | 'all'>('all');
  const [typeFilter, setTypeFilter] = useState<ApprovalType | 'all'>('all');

  const { approvals: pendingApprovals, isLoading: isPendingLoading, error: pendingError, refresh: refreshPending } = usePendingApprovals();
  const { approvals: historyApprovals, isLoading: isHistoryLoading, error: historyError, total: historyTotal, page: historyPage, loadPage: loadHistoryPage } = useApprovalHistory();
  const { stats, refresh: refreshStats } = useApprovalStats();
  const submitDecision = useApprovalStore(state => state.submitDecision);

  // Load data on mount
  useEffect(() => {
    refreshPending();
    refreshStats();
    loadHistoryPage(1);
  }, [refreshPending, refreshStats, loadHistoryPage]);

  const handleApprove = async (approval: ApprovalRequest, comment?: string) => {
    const success = await submitDecision(approval.id, 'approved', comment);
    if (success) {
      toast.success(`Approved: ${approval.title}`);
      setSelectedApproval(null);
    } else {
      toast.error('Failed to approve request');
    }
  };

  const handleReject = async (approval: ApprovalRequest, comment?: string) => {
    const success = await submitDecision(approval.id, 'rejected', comment);
    if (success) {
      toast.success(`Rejected: ${approval.title}`);
      setSelectedApproval(null);
    } else {
      toast.error('Failed to reject request');
    }
  };

  const handleRefresh = () => {
    if (activeTab === 'pending') {
      refreshPending();
    } else {
      loadHistoryPage(1);
    }
    refreshStats();
  };

  // Filter approvals based on search and filters
  const filterApprovals = (approvals: ApprovalRequest[]) => {
    return approvals.filter(approval => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesTitle = approval.title.toLowerCase().includes(query);
        const matchesDescription = approval.description?.toLowerCase().includes(query);
        const matchesRequestor = (approval.requestedBy.name || approval.requestedBy.email).toLowerCase().includes(query);
        if (!matchesTitle && !matchesDescription && !matchesRequestor) {
          return false;
        }
      }

      // Type filter
      if (typeFilter !== 'all' && approval.type !== typeFilter) {
        return false;
      }

      // Status filter (only for history)
      if (activeTab === 'history' && statusFilter !== 'all' && approval.status !== statusFilter) {
        return false;
      }

      return true;
    });
  };

  const displayedApprovals = activeTab === 'pending'
    ? filterApprovals(pendingApprovals)
    : filterApprovals(historyApprovals);

  const isLoading = activeTab === 'pending' ? isPendingLoading : isHistoryLoading;
  const error = activeTab === 'pending' ? pendingError : historyError;

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-editor-text">Approvals</h1>
            <p className="text-editor-muted">
              Review and manage approval requests
            </p>
          </div>
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text hover:bg-editor-surface/80 transition-colors disabled:opacity-50"
          >
            <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>

        {/* Stats Cards */}
        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              icon={<Clock className="text-yellow-400" size={20} />}
              label="Pending"
              value={stats.pending}
              color="yellow"
            />
            <StatCard
              icon={<CheckCircle className="text-green-400" size={20} />}
              label="Approved Today"
              value={stats.approvedToday}
              color="green"
            />
            <StatCard
              icon={<XCircle className="text-red-400" size={20} />}
              label="Rejected Today"
              value={stats.rejectedToday}
              color="red"
            />
            <StatCard
              icon={<AlertTriangle className="text-editor-accent" size={20} />}
              label="Avg Response"
              value={`${stats.avgResponseTimeMinutes}m`}
              color="blue"
            />
          </div>
        )}

        {/* Tabs */}
        <div className="flex items-center gap-1 p-1 bg-editor-surface rounded-lg w-fit">
          <button
            onClick={() => setActiveTab('pending')}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === 'pending'
                ? 'bg-editor-bg text-editor-text shadow-sm'
                : 'text-editor-muted hover:text-editor-text'
            }`}
          >
            <Clock size={16} />
            Pending
            {pendingApprovals.length > 0 && (
              <span className="px-1.5 py-0.5 text-xs bg-yellow-500/20 text-yellow-400 rounded-full">
                {pendingApprovals.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === 'history'
                ? 'bg-editor-bg text-editor-text shadow-sm'
                : 'text-editor-muted hover:text-editor-text'
            }`}
          >
            <History size={16} />
            History
          </button>
        </div>

        {/* Filters */}
        <div className="flex flex-wrap items-center gap-4">
          {/* Search */}
          <div className="relative flex-1 min-w-[200px] max-w-md">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search approvals..."
              className="w-full pl-10 pr-4 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text placeholder:text-editor-muted focus:outline-none focus:border-editor-accent"
            />
          </div>

          {/* Type Filter */}
          <div className="flex items-center gap-2">
            <Filter size={16} className="text-editor-muted" />
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value as ApprovalType | 'all')}
              className="px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
            >
              {TYPE_FILTERS.map(filter => (
                <option key={filter.value} value={filter.value}>
                  {filter.label}
                </option>
              ))}
            </select>
          </div>

          {/* Status Filter (History only) */}
          {activeTab === 'history' && (
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as ApprovalStatus | 'all')}
              className="px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm text-editor-text focus:outline-none focus:border-editor-accent"
            >
              {STATUS_FILTERS.map(filter => (
                <option key={filter.value} value={filter.value}>
                  {filter.label}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* Content */}
        {error ? (
          <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400">
            <p>{error}</p>
            <button onClick={handleRefresh} className="mt-2 text-sm underline hover:no-underline">
              Retry
            </button>
          </div>
        ) : isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-editor-muted" />
          </div>
        ) : displayedApprovals.length === 0 ? (
          <EmptyState tab={activeTab} hasFilters={!!searchQuery || typeFilter !== 'all' || statusFilter !== 'all'} />
        ) : (
          <div className="space-y-4">
            {displayedApprovals.map((approval) => (
              <ApprovalCard
                key={approval.id}
                approval={approval}
                onApprove={(comment) => handleApprove(approval, comment)}
                onReject={(comment) => handleReject(approval, comment)}
                onViewDetails={() => setSelectedApproval(approval)}
              />
            ))}

            {/* Pagination for History */}
            {activeTab === 'history' && historyTotal > 20 && (
              <div className="flex items-center justify-center gap-2 pt-4">
                <button
                  onClick={() => loadHistoryPage(historyPage - 1)}
                  disabled={historyPage <= 1}
                  className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text disabled:opacity-50"
                >
                  Previous
                </button>
                <span className="text-sm text-editor-muted">
                  Page {historyPage} of {Math.ceil(historyTotal / 20)}
                </span>
                <button
                  onClick={() => loadHistoryPage(historyPage + 1)}
                  disabled={historyPage >= Math.ceil(historyTotal / 20)}
                  className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Detail Modal */}
      {selectedApproval && (
        <ApprovalDetail
          approval={selectedApproval}
          isOpen={true}
          onClose={() => setSelectedApproval(null)}
          onApprove={(comment) => handleApprove(selectedApproval, comment)}
          onReject={(comment) => handleReject(selectedApproval, comment)}
        />
      )}
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: number | string;
  color: 'yellow' | 'green' | 'red' | 'blue';
}

function StatCard({ icon, label, value, color }: StatCardProps) {
  const colorClasses = {
    yellow: 'bg-yellow-500/5 border-yellow-500/20',
    green: 'bg-green-500/5 border-green-500/20',
    red: 'bg-red-500/5 border-red-500/20',
    blue: 'bg-editor-accent/5 border-editor-accent/20',
  };

  return (
    <div className={`p-4 rounded-lg border ${colorClasses[color]}`}>
      <div className="flex items-center gap-3">
        {icon}
        <div>
          <p className="text-2xl font-bold text-editor-text">{value}</p>
          <p className="text-sm text-editor-muted">{label}</p>
        </div>
      </div>
    </div>
  );
}

interface EmptyStateProps {
  tab: TabType;
  hasFilters: boolean;
}

function EmptyState({ tab, hasFilters }: EmptyStateProps) {
  if (hasFilters) {
    return (
      <div className="text-center py-12">
        <Search size={48} className="mx-auto text-editor-muted mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">No matching approvals</h3>
        <p className="text-editor-muted">
          Try adjusting your filters or search query
        </p>
      </div>
    );
  }

  if (tab === 'pending') {
    return (
      <div className="text-center py-12">
        <CheckCircle size={48} className="mx-auto text-green-400 mb-4" />
        <h3 className="text-lg font-medium text-editor-text mb-2">All caught up!</h3>
        <p className="text-editor-muted">
          You have no pending approval requests
        </p>
      </div>
    );
  }

  return (
    <div className="text-center py-12">
      <History size={48} className="mx-auto text-editor-muted mb-4" />
      <h3 className="text-lg font-medium text-editor-text mb-2">No approval history</h3>
      <p className="text-editor-muted">
        Completed approvals will appear here
      </p>
    </div>
  );
}
