import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type NotificationSource = 'discord' | 'slack' | 'system';
export type NotificationPriority = 'info' | 'warning' | 'error';

export interface Notification {
  id: string;
  source: NotificationSource;
  title: string;
  message: string;
  priority: NotificationPriority;
  timestamp: Date;
  read: boolean;
  link?: string;
  metadata?: Record<string, unknown>;
}

interface NotificationSettings {
  soundEnabled: boolean;
  showToasts: boolean;
}

interface NotificationStore {
  notifications: Notification[];
  settings: NotificationSettings;
  isOpen: boolean;
  filter: NotificationSource | 'all';

  // Actions
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => string;
  removeNotification: (id: string) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
  setFilter: (filter: NotificationSource | 'all') => void;
  toggleOpen: () => void;
  setOpen: (open: boolean) => void;

  // Settings
  setSoundEnabled: (enabled: boolean) => void;
  setShowToasts: (enabled: boolean) => void;

  // Computed helpers
  getUnreadCount: () => number;
  getFilteredNotifications: () => Notification[];
}

// Parse dates when rehydrating from localStorage
const parseNotificationDates = (notifications: Notification[]): Notification[] => {
  return notifications.map(n => ({
    ...n,
    timestamp: new Date(n.timestamp),
  }));
};

export const useNotificationStore = create<NotificationStore>()(
  persist(
    (set, get) => ({
      notifications: [],
      settings: {
        soundEnabled: false,
        showToasts: true,
      },
      isOpen: false,
      filter: 'all',

      addNotification: (notification) => {
        const id = crypto.randomUUID();
        const newNotification: Notification = {
          ...notification,
          id,
          timestamp: new Date(),
          read: false,
        };

        set((state) => ({
          notifications: [newNotification, ...state.notifications],
        }));

        // Play sound if enabled
        if (get().settings.soundEnabled) {
          playNotificationSound();
        }

        return id;
      },

      removeNotification: (id) =>
        set((state) => ({
          notifications: state.notifications.filter((n) => n.id !== id),
        })),

      markAsRead: (id) =>
        set((state) => ({
          notifications: state.notifications.map((n) =>
            n.id === id ? { ...n, read: true } : n
          ),
        })),

      markAllAsRead: () =>
        set((state) => ({
          notifications: state.notifications.map((n) => ({ ...n, read: true })),
        })),

      clearAll: () => set({ notifications: [] }),

      setFilter: (filter) => set({ filter }),

      toggleOpen: () => set((state) => ({ isOpen: !state.isOpen })),

      setOpen: (open) => set({ isOpen: open }),

      setSoundEnabled: (enabled) =>
        set((state) => ({
          settings: { ...state.settings, soundEnabled: enabled },
        })),

      setShowToasts: (enabled) =>
        set((state) => ({
          settings: { ...state.settings, showToasts: enabled },
        })),

      getUnreadCount: () => {
        return get().notifications.filter((n) => !n.read).length;
      },

      getFilteredNotifications: () => {
        const { notifications, filter } = get();
        if (filter === 'all') return notifications;
        return notifications.filter((n) => n.source === filter);
      },
    }),
    {
      name: 'prism-notifications',
      partialize: (state) => ({
        notifications: state.notifications,
        settings: state.settings,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // Parse date strings back to Date objects
          state.notifications = parseNotificationDates(state.notifications);
        }
      },
    }
  )
);

// Sound effect for notifications (muted by default)
function playNotificationSound() {
  try {
    const audio = new Audio('/notification.mp3');
    audio.volume = 0.3;
    audio.play().catch(() => {
      // Ignore autoplay errors
    });
  } catch {
    // Ignore audio errors
  }
}

// Convenience helpers for adding notifications
export const notify = {
  discord: (title: string, message: string, priority: NotificationPriority = 'info') =>
    useNotificationStore.getState().addNotification({ source: 'discord', title, message, priority }),

  slack: (title: string, message: string, priority: NotificationPriority = 'info') =>
    useNotificationStore.getState().addNotification({ source: 'slack', title, message, priority }),

  system: (title: string, message: string, priority: NotificationPriority = 'info') =>
    useNotificationStore.getState().addNotification({ source: 'system', title, message, priority }),

  info: (title: string, message: string, source: NotificationSource = 'system') =>
    useNotificationStore.getState().addNotification({ source, title, message, priority: 'info' }),

  warning: (title: string, message: string, source: NotificationSource = 'system') =>
    useNotificationStore.getState().addNotification({ source, title, message, priority: 'warning' }),

  error: (title: string, message: string, source: NotificationSource = 'system') =>
    useNotificationStore.getState().addNotification({ source, title, message, priority: 'error' }),
};
