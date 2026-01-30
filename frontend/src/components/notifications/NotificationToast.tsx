import { useEffect, useState, useCallback } from 'react';
import { X, AlertCircle } from 'lucide-react';
import { useNotificationStore, type Notification, type NotificationSource } from '../../store/notificationStore';
import type { LucideIcon } from 'lucide-react';

const MAX_VISIBLE_TOASTS = 3;
const AUTO_DISMISS_MS = 5000;

interface IconProps {
  className?: string;
  size?: number;
}

const DiscordIcon = ({ className, size }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    fill="currentColor"
    className={className}
    width={size}
    height={size}
  >
    <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
  </svg>
);

const SlackIcon = ({ className, size }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    fill="currentColor"
    className={className}
    width={size}
    height={size}
  >
    <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
  </svg>
);

type IconComponent = React.FC<IconProps> | LucideIcon;

const sourceIcons: Record<NotificationSource, IconComponent> = {
  discord: DiscordIcon,
  slack: SlackIcon,
  system: AlertCircle,
};

const priorityColors = {
  info: 'border-l-blue-500 bg-blue-500/5',
  warning: 'border-l-yellow-500 bg-yellow-500/5',
  error: 'border-l-red-500 bg-red-500/5',
};

interface NotificationToastItemProps {
  notification: Notification;
  onDismiss: (id: string) => void;
  onClick: (notification: Notification) => void;
}

function NotificationToastItem({ notification, onDismiss, onClick }: NotificationToastItemProps) {
  const SourceIcon = sourceIcons[notification.source];

  useEffect(() => {
    const timer = setTimeout(() => {
      onDismiss(notification.id);
    }, AUTO_DISMISS_MS);

    return () => clearTimeout(timer);
  }, [notification.id, onDismiss]);

  return (
    <div
      className={`flex items-start gap-3 px-4 py-3 bg-editor-bg border border-editor-border rounded-lg shadow-lg backdrop-blur-sm cursor-pointer transition-all hover:bg-editor-surface/50 border-l-2 ${
        priorityColors[notification.priority]
      } animate-slide-in-right`}
      onClick={() => onClick(notification)}
      role="alert"
    >
      <SourceIcon
        size={16}
        className={`flex-shrink-0 mt-0.5 ${
          notification.source === 'discord'
            ? 'text-[#5865F2]'
            : notification.source === 'slack'
            ? 'text-[#E01E5A]'
            : 'text-editor-muted'
        }`}
      />

      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-editor-text truncate">
          {notification.title}
        </p>
        <p className="text-xs text-editor-muted mt-0.5 line-clamp-2">
          {notification.message}
        </p>
      </div>

      <button
        onClick={(e) => {
          e.stopPropagation();
          onDismiss(notification.id);
        }}
        className="flex-shrink-0 p-1 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-colors"
        aria-label="Dismiss"
      >
        <X size={14} />
      </button>
    </div>
  );
}

export function NotificationToastContainer() {
  const { notifications, settings, markAsRead, setOpen } = useNotificationStore();
  const [dismissedIds, setDismissedIds] = useState<Set<string>>(new Set());

  // Only show toasts for new notifications that haven't been shown or dismissed
  const visibleToasts = notifications
    .filter((n) => !n.read && !dismissedIds.has(n.id))
    .slice(0, MAX_VISIBLE_TOASTS);

  // Only show toasts that were newly added (not from localStorage)
  const toastsToShow = visibleToasts.filter((n) => {
    // Show if added within the last 10 seconds
    const age = Date.now() - new Date(n.timestamp).getTime();
    return age < 10000;
  });

  const handleDismiss = useCallback((id: string) => {
    setDismissedIds((prev) => new Set(prev).add(id));
  }, []);

  const handleClick = useCallback(
    (notification: Notification) => {
      markAsRead(notification.id);
      handleDismiss(notification.id);
      setOpen(true);
    },
    [markAsRead, handleDismiss, setOpen]
  );

  if (!settings.showToasts || toastsToShow.length === 0) {
    return null;
  }

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
      {toastsToShow.map((notification) => (
        <NotificationToastItem
          key={notification.id}
          notification={notification}
          onDismiss={handleDismiss}
          onClick={handleClick}
        />
      ))}
    </div>
  );
}
