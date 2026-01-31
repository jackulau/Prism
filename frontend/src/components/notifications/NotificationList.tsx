import { CheckCheck, Inbox } from 'lucide-react';
import { NotificationItem } from './NotificationItem';
import { useNotificationStore, type Notification, type NotificationSource } from '../../store/notificationStore';

interface NotificationListProps {
  onMarkAsRead: (id: string) => void;
}

type FilterTab = 'all' | NotificationSource;

const FILTER_TABS: { label: string; value: FilterTab }[] = [
  { label: 'All', value: 'all' },
  { label: 'Discord', value: 'discord' },
  { label: 'Slack', value: 'slack' },
  { label: 'System', value: 'system' },
];

function groupNotificationsByTime(notifications: Notification[]): Record<string, Notification[]> {
  const groups: Record<string, Notification[]> = {
    Today: [],
    Yesterday: [],
    'This Week': [],
    Earlier: [],
  };

  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today.getTime() - 86400000);
  const weekAgo = new Date(today.getTime() - 604800000);

  for (const notification of notifications) {
    const notifDate = new Date(notification.timestamp);
    const notifDay = new Date(notifDate.getFullYear(), notifDate.getMonth(), notifDate.getDate());

    if (notifDay.getTime() >= today.getTime()) {
      groups.Today.push(notification);
    } else if (notifDay.getTime() >= yesterday.getTime()) {
      groups.Yesterday.push(notification);
    } else if (notifDay.getTime() >= weekAgo.getTime()) {
      groups['This Week'].push(notification);
    } else {
      groups.Earlier.push(notification);
    }
  }

  return groups;
}

export function NotificationList({ onMarkAsRead }: NotificationListProps) {
  const { filter, setFilter, getFilteredNotifications, markAllAsRead, getUnreadCount } = useNotificationStore();
  const notifications = getFilteredNotifications();
  const unreadCount = getUnreadCount();
  const groupedNotifications = groupNotificationsByTime(notifications);

  return (
    <div className="flex flex-col h-full">
      {/* Filter Tabs */}
      <div className="flex items-center gap-1 px-3 py-2 border-b border-editor-border overflow-x-auto">
        {FILTER_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setFilter(tab.value)}
            className={`px-2.5 py-1 text-xs font-medium rounded-md transition-colors whitespace-nowrap ${
              filter === tab.value
                ? 'bg-editor-accent/20 text-editor-accent'
                : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
            }`}
          >
            {tab.label}
          </button>
        ))}

        {/* Mark All as Read */}
        {unreadCount > 0 && (
          <button
            onClick={markAllAsRead}
            className="ml-auto p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-md transition-colors"
            title="Mark all as read"
          >
            <CheckCheck size={14} />
          </button>
        )}
      </div>

      {/* Notification List */}
      <div className="flex-1 overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full py-12 text-editor-muted">
            <Inbox size={40} className="mb-3 opacity-50" />
            <p className="text-sm font-medium">No notifications</p>
            <p className="text-xs mt-1">You're all caught up!</p>
          </div>
        ) : (
          <div className="divide-y divide-editor-border/50">
            {Object.entries(groupedNotifications).map(([group, items]) => {
              if (items.length === 0) return null;
              return (
                <div key={group}>
                  <div className="px-4 py-2 text-xs font-semibold text-editor-muted uppercase tracking-wider bg-editor-surface/30 sticky top-0">
                    {group}
                  </div>
                  <div className="divide-y divide-editor-border/30">
                    {items.map((notification) => (
                      <NotificationItem
                        key={notification.id}
                        notification={notification}
                        onMarkAsRead={onMarkAsRead}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
