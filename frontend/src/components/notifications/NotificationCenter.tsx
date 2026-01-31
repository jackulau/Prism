import { useEffect, useRef, useCallback } from 'react';
import { X, Trash2, ExternalLink } from 'lucide-react';
import { NotificationBell } from './NotificationBell';
import { NotificationList } from './NotificationList';
import { useNotificationStore } from '../../store/notificationStore';

interface NotificationCenterProps {
  className?: string;
}

export function NotificationCenter({ className }: NotificationCenterProps) {
  const { isOpen, setOpen, toggleOpen, markAsRead, clearAll, notifications, getUnreadCount } =
    useNotificationStore();
  const panelRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLDivElement>(null);
  const unreadCount = getUnreadCount();

  // Close on click outside
  const handleClickOutside = useCallback(
    (event: MouseEvent) => {
      if (
        panelRef.current &&
        buttonRef.current &&
        !panelRef.current.contains(event.target as Node) &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setOpen(false);
      }
    },
    [setOpen]
  );

  // Close on Escape key
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen) {
        setOpen(false);
      }
    },
    [isOpen, setOpen]
  );

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, handleClickOutside, handleKeyDown]);

  return (
    <div className={`relative ${className || ''}`}>
      {/* Bell Button */}
      <div ref={buttonRef}>
        <NotificationBell onClick={toggleOpen} isOpen={isOpen} />
      </div>

      {/* Dropdown Panel */}
      {isOpen && (
        <div
          ref={panelRef}
          className="absolute right-0 top-full mt-2 w-80 sm:w-96 max-h-[60vh] bg-editor-bg border border-editor-border rounded-lg shadow-xl overflow-hidden z-50 animate-slide-in-down"
          role="dialog"
          aria-label="Notifications"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-editor-border bg-editor-surface/30">
            <div className="flex items-center gap-2">
              <h2 className="font-semibold text-editor-text">Notifications</h2>
              {unreadCount > 0 && (
                <span className="px-1.5 py-0.5 text-xs font-medium bg-editor-accent/20 text-editor-accent rounded-full">
                  {unreadCount} new
                </span>
              )}
            </div>
            <div className="flex items-center gap-1">
              {notifications.length > 0 && (
                <button
                  onClick={clearAll}
                  className="p-1.5 text-editor-muted hover:text-editor-error hover:bg-editor-error/10 rounded-md transition-colors"
                  title="Clear all notifications"
                >
                  <Trash2 size={14} />
                </button>
              )}
              <button
                onClick={() => setOpen(false)}
                className="p-1.5 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-md transition-colors"
                title="Close"
              >
                <X size={14} />
              </button>
            </div>
          </div>

          {/* Notification List */}
          <div className="max-h-[calc(60vh-100px)] overflow-hidden">
            <NotificationList onMarkAsRead={markAsRead} />
          </div>

          {/* Footer */}
          <div className="px-4 py-2 border-t border-editor-border bg-editor-surface/30">
            <button
              className="w-full flex items-center justify-center gap-2 py-1.5 text-xs text-editor-muted hover:text-editor-accent transition-colors"
              onClick={() => {
                // TODO: Navigate to full notifications page when available
                setOpen(false);
              }}
            >
              <span>View all notifications</span>
              <ExternalLink size={12} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
