import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import type {
  SystemHealthStatus,
  WsConnectionStatus,
  SseConnectionStatus,
  ActiveAgent,
  Notification,
  ActivityEvent,
  MonitoringMetrics,
  SerializedMonitoringState,
  ActivityEventType,
  ActivityEventSeverity,
  NotificationSource,
  NotificationPriority,
} from '../types/monitoring';

// Constants
const MAX_ACTIVITY_EVENTS = 100;
const STORAGE_KEY = 'prism_monitoring_state';

// Initial metrics state
const initialMetrics: MonitoringMetrics = {
  activeConnections: 0,
  messagesThroughput: 0,
  averageLatency: 0,
  peakLatency: 0,
  messagesReceived: 0,
  messagesSent: 0,
  errorsCount: 0,
  lastUpdated: Date.now(),
};

// Load persisted state from localStorage
function loadPersistedState(): Partial<SerializedMonitoringState> {
  if (typeof window === 'undefined') return {};
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as SerializedMonitoringState;
      return {
        notifications: parsed.notifications || [],
        lastHeartbeat: parsed.lastHeartbeat || null,
      };
    }
  } catch {
    // Ignore parsing errors
  }
  return {};
}

// Save state to localStorage
function persistState(state: { notifications: Notification[]; lastHeartbeat: number | null }) {
  if (typeof window === 'undefined') return;
  try {
    const toStore: SerializedMonitoringState = {
      notifications: state.notifications,
      lastHeartbeat: state.lastHeartbeat,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(toStore));
  } catch {
    // Ignore storage errors
  }
}

// Store interface
interface MonitoringStoreState {
  // State
  systemHealth: SystemHealthStatus;
  wsStatus: WsConnectionStatus;
  sseStatus: SseConnectionStatus;
  lastHeartbeat: number | null;
  activeAgents: Map<string, ActiveAgent>;
  notifications: Notification[];
  activityEvents: ActivityEvent[];
  metrics: MonitoringMetrics;

  // Health actions
  setSystemHealth: (status: SystemHealthStatus) => void;
  computeSystemHealth: () => SystemHealthStatus;

  // Connection status actions
  updateWsStatus: (status: WsConnectionStatus) => void;
  updateSseStatus: (status: SseConnectionStatus) => void;
  recordHeartbeat: () => void;

  // Agent actions
  addAgent: (agent: ActiveAgent) => void;
  updateAgent: (id: string, updates: Partial<ActiveAgent>) => void;
  removeAgent: (id: string) => void;

  // Notification actions
  addNotification: (notification: Omit<Notification, 'id' | 'createdAt' | 'isRead'>) => void;
  markNotificationRead: (id: string) => void;
  markAllNotificationsRead: () => void;
  clearNotifications: () => void;
  removeNotification: (id: string) => void;

  // Activity event actions
  addActivityEvent: (event: Omit<ActivityEvent, 'id' | 'timestamp'>) => void;
  clearOldEvents: (olderThan: number) => void;
  clearAllEvents: () => void;

  // Metrics actions
  updateMetrics: (metrics: Partial<MonitoringMetrics>) => void;
  resetMetrics: () => void;

  // Utility actions
  reset: () => void;
}

// Create the store with subscribeWithSelector middleware for efficient re-renders
export const useMonitoringStore = create<MonitoringStoreState>()(
  subscribeWithSelector((set, get) => {
    // Load persisted state
    const persisted = loadPersistedState();

    return {
      // Initial state
      systemHealth: 'healthy',
      wsStatus: 'disconnected',
      sseStatus: 'disconnected',
      lastHeartbeat: persisted.lastHeartbeat ?? null,
      activeAgents: new Map(),
      notifications: persisted.notifications ?? [],
      activityEvents: [],
      metrics: initialMetrics,

      // Health actions
      setSystemHealth: (status) => set({ systemHealth: status }),

      computeSystemHealth: () => {
        const { wsStatus, sseStatus, lastHeartbeat } = get();

        // Check for error states
        if (wsStatus === 'error' || sseStatus === 'error') {
          return 'unhealthy';
        }

        // Check for disconnected states
        if (wsStatus === 'disconnected' && sseStatus === 'disconnected') {
          return 'unhealthy';
        }

        // Check heartbeat staleness (> 30 seconds)
        if (lastHeartbeat && Date.now() - lastHeartbeat > 30000) {
          return 'degraded';
        }

        // Check for partial connectivity
        if (wsStatus === 'disconnected' || sseStatus === 'disconnected') {
          return 'degraded';
        }

        // Check for connecting states
        if (wsStatus === 'connecting' || sseStatus === 'connecting') {
          return 'degraded';
        }

        return 'healthy';
      },

      // Connection status actions
      updateWsStatus: (status) => {
        const previousStatus = get().wsStatus;
        set({ wsStatus: status });

        // Add activity event for connection state changes
        if (previousStatus !== status) {
          const { addActivityEvent } = get();
          if (status === 'connected' && previousStatus === 'disconnected') {
            addActivityEvent({
              type: 'connection.established',
              title: 'WebSocket Connected',
              description: 'WebSocket connection established',
              severity: 'success',
            });
          } else if (status === 'disconnected' && previousStatus === 'connected') {
            addActivityEvent({
              type: 'connection.lost',
              title: 'WebSocket Disconnected',
              description: 'WebSocket connection lost',
              severity: 'warning',
            });
          } else if (status === 'connected' && previousStatus === 'connecting') {
            addActivityEvent({
              type: 'connection.reconnected',
              title: 'WebSocket Reconnected',
              description: 'WebSocket connection re-established',
              severity: 'success',
            });
          } else if (status === 'error') {
            addActivityEvent({
              type: 'connection.lost',
              title: 'WebSocket Error',
              description: 'WebSocket connection error',
              severity: 'error',
            });
          }
        }

        // Auto-update system health
        const newHealth = get().computeSystemHealth();
        set({ systemHealth: newHealth });
      },

      updateSseStatus: (status) => {
        const previousStatus = get().sseStatus;
        set({ sseStatus: status });

        // Add activity event for connection state changes
        if (previousStatus !== status) {
          const { addActivityEvent } = get();
          if (status === 'connected' && previousStatus === 'disconnected') {
            addActivityEvent({
              type: 'connection.established',
              title: 'SSE Connected',
              description: 'Server-Sent Events connection established',
              severity: 'success',
            });
          } else if (status === 'disconnected' && previousStatus === 'connected') {
            addActivityEvent({
              type: 'connection.lost',
              title: 'SSE Disconnected',
              description: 'Server-Sent Events connection lost',
              severity: 'warning',
            });
          } else if (status === 'error') {
            addActivityEvent({
              type: 'connection.lost',
              title: 'SSE Error',
              description: 'Server-Sent Events connection error',
              severity: 'error',
            });
          }
        }

        // Auto-update system health
        const newHealth = get().computeSystemHealth();
        set({ systemHealth: newHealth });
      },

      recordHeartbeat: () => {
        const timestamp = Date.now();
        set({ lastHeartbeat: timestamp });
        persistState({ notifications: get().notifications, lastHeartbeat: timestamp });

        // Auto-update system health
        const newHealth = get().computeSystemHealth();
        set({ systemHealth: newHealth });
      },

      // Agent actions
      addAgent: (agent) => {
        set((state) => {
          const newAgents = new Map(state.activeAgents);
          newAgents.set(agent.id, agent);
          return { activeAgents: newAgents };
        });

        // Add activity event
        get().addActivityEvent({
          type: 'agent.started',
          title: `Agent Started: ${agent.name}`,
          description: agent.taskDescription,
          severity: 'info',
          agentId: agent.id,
          workspaceId: agent.workspaceId,
        });
      },

      updateAgent: (id, updates) => {
        set((state) => {
          const agent = state.activeAgents.get(id);
          if (!agent) return state;

          const newAgents = new Map(state.activeAgents);
          newAgents.set(id, { ...agent, ...updates, updatedAt: Date.now() });
          return { activeAgents: newAgents };
        });
      },

      removeAgent: (id) => {
        const agent = get().activeAgents.get(id);

        set((state) => {
          const newAgents = new Map(state.activeAgents);
          newAgents.delete(id);
          return { activeAgents: newAgents };
        });

        // Add activity event if agent existed
        if (agent) {
          const eventType: ActivityEventType =
            agent.status === 'completed' ? 'agent.completed' :
            agent.status === 'failed' ? 'agent.failed' :
            agent.status === 'cancelled' ? 'agent.cancelled' : 'agent.completed';

          const severity: ActivityEventSeverity =
            agent.status === 'failed' ? 'error' :
            agent.status === 'cancelled' ? 'warning' : 'success';

          get().addActivityEvent({
            type: eventType,
            title: `Agent ${agent.status === 'completed' ? 'Completed' : agent.status === 'failed' ? 'Failed' : 'Cancelled'}: ${agent.name}`,
            description: agent.taskDescription,
            severity,
            agentId: agent.id,
            workspaceId: agent.workspaceId,
          });
        }
      },

      // Notification actions
      addNotification: (notification) => {
        const newNotification: Notification = {
          ...notification,
          id: `notif-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          createdAt: Date.now(),
          isRead: false,
        };

        set((state) => {
          const notifications = [newNotification, ...state.notifications];
          persistState({ notifications, lastHeartbeat: state.lastHeartbeat });
          return { notifications };
        });

        // Add activity event
        get().addActivityEvent({
          type: 'notification.received',
          title: `Notification: ${notification.title}`,
          description: notification.message,
          severity: notification.priority === 'urgent' || notification.priority === 'high' ? 'warning' : 'info',
        });
      },

      markNotificationRead: (id) => {
        set((state) => {
          const notifications = state.notifications.map((n) =>
            n.id === id ? { ...n, isRead: true } : n
          );
          persistState({ notifications, lastHeartbeat: state.lastHeartbeat });
          return { notifications };
        });
      },

      markAllNotificationsRead: () => {
        set((state) => {
          const notifications = state.notifications.map((n) => ({ ...n, isRead: true }));
          persistState({ notifications, lastHeartbeat: state.lastHeartbeat });
          return { notifications };
        });
      },

      clearNotifications: () => {
        set((state) => {
          persistState({ notifications: [], lastHeartbeat: state.lastHeartbeat });
          return { notifications: [] };
        });
      },

      removeNotification: (id) => {
        set((state) => {
          const notifications = state.notifications.filter((n) => n.id !== id);
          persistState({ notifications, lastHeartbeat: state.lastHeartbeat });
          return { notifications };
        });
      },

      // Activity event actions
      addActivityEvent: (event) => {
        const newEvent: ActivityEvent = {
          ...event,
          id: `event-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          timestamp: Date.now(),
        };

        set((state) => {
          // Add new event and prune to MAX_ACTIVITY_EVENTS
          const activityEvents = [newEvent, ...state.activityEvents].slice(0, MAX_ACTIVITY_EVENTS);
          return { activityEvents };
        });
      },

      clearOldEvents: (olderThan) => {
        set((state) => ({
          activityEvents: state.activityEvents.filter((e) => e.timestamp >= olderThan),
        }));
      },

      clearAllEvents: () => set({ activityEvents: [] }),

      // Metrics actions
      updateMetrics: (metrics) => {
        set((state) => ({
          metrics: { ...state.metrics, ...metrics, lastUpdated: Date.now() },
        }));
      },

      resetMetrics: () => set({ metrics: initialMetrics }),

      // Utility actions
      reset: () => {
        set({
          systemHealth: 'healthy',
          wsStatus: 'disconnected',
          sseStatus: 'disconnected',
          lastHeartbeat: null,
          activeAgents: new Map(),
          notifications: [],
          activityEvents: [],
          metrics: initialMetrics,
        });
        if (typeof window !== 'undefined') {
          localStorage.removeItem(STORAGE_KEY);
        }
      },
    };
  })
);

// ============================================================================
// Selectors - Use these for efficient component re-renders
// ============================================================================

/**
 * Get the count of unread notifications
 */
export function useUnreadNotificationCount(): number {
  return useMonitoringStore((state) =>
    state.notifications.filter((n) => !n.isRead).length
  );
}

/**
 * Get the count of active agents
 */
export function useActiveAgentCount(): number {
  return useMonitoringStore((state) => state.activeAgents.size);
}

/**
 * Get recent activity events with optional limit
 */
export function useRecentActivity(limit = 10): ActivityEvent[] {
  return useMonitoringStore((state) => state.activityEvents.slice(0, limit));
}

/**
 * Get computed system health status based on connection states
 */
export function useSystemHealthStatus(): SystemHealthStatus {
  return useMonitoringStore((state) => {
    const { wsStatus, sseStatus, lastHeartbeat } = state;

    // Check for error states
    if (wsStatus === 'error' || sseStatus === 'error') {
      return 'unhealthy';
    }

    // Check for disconnected states
    if (wsStatus === 'disconnected' && sseStatus === 'disconnected') {
      return 'unhealthy';
    }

    // Check heartbeat staleness (> 30 seconds)
    if (lastHeartbeat && Date.now() - lastHeartbeat > 30000) {
      return 'degraded';
    }

    // Check for partial connectivity
    if (wsStatus === 'disconnected' || sseStatus === 'disconnected') {
      return 'degraded';
    }

    // Check for connecting states
    if (wsStatus === 'connecting' || sseStatus === 'connecting') {
      return 'degraded';
    }

    return 'healthy';
  });
}

/**
 * Get active agents as an array
 */
export function useActiveAgents(): ActiveAgent[] {
  return useMonitoringStore((state) => Array.from(state.activeAgents.values()));
}

/**
 * Get unread notifications
 */
export function useUnreadNotifications(): Notification[] {
  return useMonitoringStore((state) => state.notifications.filter((n) => !n.isRead));
}

/**
 * Get all notifications
 */
export function useNotifications(): Notification[] {
  return useMonitoringStore((state) => state.notifications);
}

/**
 * Get connection statuses
 */
export function useConnectionStatus(): { ws: WsConnectionStatus; sse: SseConnectionStatus } {
  return useMonitoringStore((state) => ({
    ws: state.wsStatus,
    sse: state.sseStatus,
  }));
}

/**
 * Get monitoring metrics
 */
export function useMonitoringMetrics(): MonitoringMetrics {
  return useMonitoringStore((state) => state.metrics);
}

// ============================================================================
// Helper functions for creating notifications and events
// ============================================================================

/**
 * Create a notification from external sources
 */
export function createNotification(
  title: string,
  message: string,
  source: NotificationSource,
  priority: NotificationPriority = 'medium',
  options?: Partial<Omit<Notification, 'id' | 'title' | 'message' | 'source' | 'priority' | 'createdAt' | 'isRead'>>
): Omit<Notification, 'id' | 'createdAt' | 'isRead'> {
  return {
    title,
    message,
    source,
    priority,
    ...options,
  };
}
