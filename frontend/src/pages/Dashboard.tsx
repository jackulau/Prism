import { RecentWorkspaces } from '../components/dashboard/RecentWorkspaces';
import { QuickActions } from '../components/dashboard/QuickActions';
import { UsageSummary } from '../components/dashboard/UsageSummary';
import { ActiveAgents } from '../components/dashboard/ActiveAgents';

export default function Dashboard() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-editor-text">Dashboard</h1>
          <p className="text-editor-muted">
            Welcome to Prism. Start a new workspace or continue where you left off.
          </p>
        </div>

        {/* Quick Actions */}
        <QuickActions />

        {/* Active Agents - shows running AI agents and swarms */}
        <ActiveAgents />

        {/* Usage Summary */}
        <UsageSummary />

        {/* Recent Workspaces */}
        <RecentWorkspaces />
      </div>
    </div>
  );
}
