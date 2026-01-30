import { Bell, BellRing } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useNotificationStore } from '../../store/notificationStore';

interface NotificationBellProps {
  onClick: () => void;
  isOpen?: boolean;
}

export function NotificationBell({ onClick, isOpen }: NotificationBellProps) {
  const { getUnreadCount, notifications } = useNotificationStore();
  const unreadCount = getUnreadCount();
  const [isAnimating, setIsAnimating] = useState(false);
  const prevCountRef = useRef(unreadCount);

  // Animate when new notification arrives
  useEffect(() => {
    if (notifications.length > 0 && unreadCount > prevCountRef.current) {
      setIsAnimating(true);
      const timer = setTimeout(() => setIsAnimating(false), 500);
      return () => clearTimeout(timer);
    }
    prevCountRef.current = unreadCount;
  }, [notifications.length, unreadCount]);

  const displayCount = unreadCount > 99 ? '99+' : unreadCount;
  const hasUnread = unreadCount > 0;

  return (
    <button
      onClick={onClick}
      className={`relative p-2 rounded-lg transition-colors ${
        isOpen
          ? 'bg-editor-accent/20 text-editor-accent'
          : 'text-editor-muted hover:text-editor-text hover:bg-editor-surface'
      }`}
      title={hasUnread ? `${unreadCount} unread notifications` : 'Notifications'}
      aria-label={hasUnread ? `${unreadCount} unread notifications` : 'Notifications'}
    >
      {isAnimating || hasUnread ? (
        <BellRing
          size={18}
          className={isAnimating ? 'animate-wiggle' : ''}
        />
      ) : (
        <Bell size={18} />
      )}

      {/* Badge */}
      {hasUnread && (
        <span
          className={`absolute -top-0.5 -right-0.5 flex items-center justify-center min-w-[18px] h-[18px] px-1 text-[10px] font-bold bg-editor-accent text-white rounded-full ${
            isAnimating ? 'animate-bounce' : ''
          }`}
        >
          {displayCount}
        </span>
      )}
    </button>
  );
}
