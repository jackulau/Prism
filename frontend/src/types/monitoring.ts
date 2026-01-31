// System health status
export type SystemHealthStatus = 'healthy' | 'degraded' | 'unhealthy';

// WebSocket connection status
export type WsConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

// SSE connection status
export type SseConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

// Agent status
export type AgentStatus = 'starting' | 'running' | 'completed' | 'failed' | 'cancelled';

// Active agent information
export interface ActiveAgent {
  id: string;
  name: string;
  status: AgentStatus;
  startedAt: number;
  updatedAt: number;
  workspaceId?: string;
  taskDescription?: string;
  progress?: number;
  metadata?: Record<string, unknown>;
}

// Notification source type
export type NotificationSource = 'discord' | 'slack' | 'github' | 'system' | 'agent';

// Notification priority
export type NotificationPriority = 'low' | 'medium' | 'high' | 'urgent';

// Notification item
export interface Notification {
  id: string;
  title: string;
  message: string;
  source: NotificationSource;
  priority: NotificationPriority;
  isRead: boolean;
  createdAt: number;
  expiresAt?: number;
  link?: string;
  metadata?: Record<string, unknown>;
}

// Activity event types
export type ActivityEventType =
  | 'agent.started'
  | 'agent.completed'
  | 'agent.failed'
  | 'agent.cancelled'
  | 'swarm.started'
  | 'swarm.completed'
  | 'swarm.failed'
  | 'connection.established'
  | 'connection.lost'
  | 'connection.reconnected'
  | 'build.started'
  | 'build.completed'
  | 'build.failed'
  | 'notification.received'
  | 'system.health_changed'
  | 'user.action';

// Activity event severity
export type ActivityEventSeverity = 'info' | 'success' | 'warning' | 'error';

// Activity event
export interface ActivityEvent {
  id: string;
  type: ActivityEventType;
  title: string;
  description?: string;
  severity: ActivityEventSeverity;
  timestamp: number;
  agentId?: string;
  workspaceId?: string;
  metadata?: Record<string, unknown>;
}

// Real-time metrics
export interface MonitoringMetrics {
  activeConnections: number;
  messagesThroughput: number;
  averageLatency: number;
  peakLatency: number;
  messagesReceived: number;
  messagesSent: number;
  errorsCount: number;
  lastUpdated: number;
}

// Monitoring store state
export interface MonitoringState {
  // Connection state
  systemHealth: SystemHealthStatus;
  wsStatus: WsConnectionStatus;
  sseStatus: SseConnectionStatus;
  lastHeartbeat: number | null;

  // Agents
  activeAgents: Map<string, ActiveAgent>;

  // Notifications
  notifications: Notification[];

  // Activity
  activityEvents: ActivityEvent[];

  // Metrics
  metrics: MonitoringMetrics;
}

// Serialized monitoring state for localStorage
export interface SerializedMonitoringState {
  notifications: Notification[];
  lastHeartbeat: number | null;
}
