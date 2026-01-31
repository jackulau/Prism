import { ActivityFeed, type EventTypeFilter, type EventStatusFilter } from './ActivityFeed';

export type PanelSize = 'compact' | 'medium' | 'large' | 'full';

interface ActivityFeedPanelProps {
  size?: PanelSize;
  showHeader?: boolean;
  className?: string;
  initialTypeFilter?: EventTypeFilter;
  initialStatusFilter?: EventStatusFilter;
}

function getPanelHeight(size: PanelSize): string {
  switch (size) {
    case 'compact':
      return '300px';
    case 'medium':
      return '500px';
    case 'large':
      return '700px';
    case 'full':
      return '100%';
    default:
      return '400px';
  }
}

export function ActivityFeedPanel({
  size = 'medium',
  showHeader = true,
  className = '',
  initialTypeFilter = 'all',
  initialStatusFilter = 'all',
}: ActivityFeedPanelProps) {
  const height = getPanelHeight(size);
  const isFullHeight = size === 'full';

  return (
    <div
      className={`
        bg-editor-surface border border-editor-border rounded-lg overflow-hidden
        ${isFullHeight ? 'h-full' : ''}
        ${className}
      `}
      style={!isFullHeight ? { height } : undefined}
    >
      <ActivityFeed
        maxHeight={isFullHeight ? '100%' : height}
        showHeader={showHeader}
        initialTypeFilter={initialTypeFilter}
        initialStatusFilter={initialStatusFilter}
      />
    </div>
  );
}

// Slide-out panel variant for sidebar
interface SlideOutActivityPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SlideOutActivityPanel({ isOpen, onClose }: SlideOutActivityPanelProps) {
  return (
    <>
      {/* Backdrop */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/20 z-40 transition-opacity"
          onClick={onClose}
        />
      )}

      {/* Panel */}
      <div
        className={`
          fixed right-0 top-0 h-full w-96 max-w-full bg-editor-surface border-l border-editor-border z-50
          transform transition-transform duration-300 ease-in-out
          ${isOpen ? 'translate-x-0' : 'translate-x-full'}
        `}
      >
        <div className="h-full flex flex-col">
          {/* Close button */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border">
            <h2 className="text-lg font-semibold text-editor-text">Activity</h2>
            <button
              onClick={onClose}
              className="p-1.5 rounded text-editor-muted hover:text-editor-text hover:bg-editor-bg transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          {/* Activity feed */}
          <div className="flex-1 overflow-hidden">
            <ActivityFeed maxHeight="100%" showHeader={true} />
          </div>
        </div>
      </div>
    </>
  );
}
