import { ActivityFeed } from '../components/activity';
import { useMonitoringStore } from '../store/monitoringStore';

export default function Activity() {
  const eventCount = useMonitoringStore((state) => state.activityEvents.length);
  const wsStatus = useMonitoringStore((state) => state.wsStatus);

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="max-w-6xl mx-auto w-full p-6 flex flex-col flex-1 min-h-0">
        {/* Header */}
        <div className="space-y-2 mb-6 shrink-0">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-editor-text">Activity</h1>
              <p className="text-editor-muted">
                Real-time stream of all system events and activity
              </p>
            </div>

            {/* Stats */}
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-2">
                <span className="text-editor-muted">Events:</span>
                <span className="font-medium text-editor-text">{eventCount}</span>
              </div>
              <div className="flex items-center gap-2">
                <div
                  className={`w-2 h-2 rounded-full ${
                    wsStatus === 'connected'
                      ? 'bg-editor-success animate-pulse'
                      : wsStatus === 'connecting'
                      ? 'bg-editor-warning animate-pulse'
                      : 'bg-editor-error'
                  }`}
                />
                <span className="text-editor-muted capitalize">{wsStatus}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Activity Feed - Full height */}
        <div className="flex-1 min-h-0 bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
          <ActivityFeed maxHeight="100%" showHeader={true} />
        </div>
      </div>
    </div>
  );
}
