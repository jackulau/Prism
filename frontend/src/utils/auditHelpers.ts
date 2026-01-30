import {
  LogIn,
  LogOut,
  UserPlus,
  KeyRound,
  Trash2,
  Key,
  Shield,
  ShieldOff,
  ShieldCheck,
  ShieldAlert,
  Settings,
  RefreshCw,
  Clock,
  AlertCircle,
  CheckCircle,
  XCircle,
  type LucideIcon,
  Github,
  Unlink,
} from 'lucide-react';

// Event type to icon mapping
const eventIcons: Record<string, LucideIcon> = {
  // Auth events
  login: LogIn,
  login_failed: AlertCircle,
  logout: LogOut,
  register: UserPlus,
  password_change: KeyRound,
  token_refresh: RefreshCw,

  // API Key events
  api_key_created: Key,
  api_key_deleted: Trash2,
  api_key_used: Key,

  // Session events
  session_created: Clock,
  session_terminated: XCircle,
  session_expired: Clock,

  // MFA events
  mfa_enabled: Shield,
  mfa_disabled: ShieldOff,
  mfa_verified: ShieldCheck,
  mfa_failed: ShieldAlert,
  backup_code_used: KeyRound,

  // Provider events
  provider_key_set: Key,
  provider_key_deleted: Trash2,

  // Settings events
  settings_changed: Settings,

  // GitHub events
  github_connected: Github,
  github_disconnected: Unlink,
};

// Event type labels
const eventLabels: Record<string, string> = {
  // Auth events
  login: 'Login successful',
  login_failed: 'Login failed',
  logout: 'Logged out',
  register: 'Account created',
  password_change: 'Password changed',
  token_refresh: 'Session refreshed',

  // API Key events
  api_key_created: 'API key created',
  api_key_deleted: 'API key deleted',
  api_key_used: 'API key used',

  // Session events
  session_created: 'Session started',
  session_terminated: 'Session ended',
  session_expired: 'Session expired',

  // MFA events
  mfa_enabled: 'MFA enabled',
  mfa_disabled: 'MFA disabled',
  mfa_verified: 'MFA verified',
  mfa_failed: 'MFA verification failed',
  backup_code_used: 'Backup code used',

  // Provider events
  provider_key_set: 'Provider key added',
  provider_key_deleted: 'Provider key removed',

  // Settings events
  settings_changed: 'Settings updated',

  // GitHub events
  github_connected: 'GitHub connected',
  github_disconnected: 'GitHub disconnected',
};

// Category labels
const categoryLabels: Record<string, string> = {
  authentication: 'Authentication',
  api_key: 'API Keys',
  session: 'Sessions',
  settings: 'Settings',
  mfa: 'Multi-Factor Auth',
  provider: 'LLM Providers',
  github: 'GitHub',
};

// Get icon for an event type
export function getEventIcon(eventType: string): LucideIcon {
  return eventIcons[eventType] || AlertCircle;
}

// Get human-readable label for an event type
export function getEventLabel(eventType: string): string {
  return eventLabels[eventType] || eventType.replace(/_/g, ' ');
}

// Get human-readable label for a category
export function getCategoryLabel(category: string): string {
  return categoryLabels[category] || category.replace(/_/g, ' ');
}

// Get color class based on success status and event type
export function getEventColor(success: boolean, eventType?: string): string {
  if (!success) {
    return 'text-red-400';
  }

  // Warning events even when successful
  const warningEvents = ['mfa_disabled', 'api_key_deleted', 'provider_key_deleted', 'session_terminated'];
  if (eventType && warningEvents.includes(eventType)) {
    return 'text-yellow-400';
  }

  // Success events
  const successEvents = ['login', 'mfa_enabled', 'mfa_verified', 'register'];
  if (eventType && successEvents.includes(eventType)) {
    return 'text-green-400';
  }

  return 'text-editor-text';
}

// Get background color class for success/failure indicator
export function getEventBgColor(success: boolean): string {
  return success ? 'bg-green-500/10' : 'bg-red-500/10';
}

// Format audit log details for display
export function formatAuditDetails(details: Record<string, unknown> | null): string {
  if (!details) return '';

  const parts: string[] = [];

  // Handle common detail fields
  if (details.provider) {
    parts.push(`Provider: ${details.provider}`);
  }
  if (details.model) {
    parts.push(`Model: ${details.model}`);
  }
  if (details.key_name) {
    parts.push(`Key: ${details.key_name}`);
  }
  if (details.setting) {
    parts.push(`Setting: ${details.setting}`);
  }
  if (details.reason) {
    parts.push(`Reason: ${details.reason}`);
  }

  return parts.join(' • ');
}

// Format relative time
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) {
    return 'Just now';
  } else if (diffMinutes < 60) {
    return `${diffMinutes} minute${diffMinutes === 1 ? '' : 's'} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  } else {
    return date.toLocaleDateString();
  }
}

// Parse user agent to get browser and OS
export function parseUserAgent(userAgent: string | null): { browser: string; os: string } {
  if (!userAgent) {
    return { browser: 'Unknown', os: 'Unknown' };
  }

  let browser = 'Unknown';
  let os = 'Unknown';

  // Detect browser
  if (userAgent.includes('Firefox')) {
    browser = 'Firefox';
  } else if (userAgent.includes('Edg/')) {
    browser = 'Edge';
  } else if (userAgent.includes('Chrome')) {
    browser = 'Chrome';
  } else if (userAgent.includes('Safari')) {
    browser = 'Safari';
  }

  // Detect OS
  if (userAgent.includes('Windows')) {
    os = 'Windows';
  } else if (userAgent.includes('Mac OS')) {
    os = 'macOS';
  } else if (userAgent.includes('Linux')) {
    os = 'Linux';
  } else if (userAgent.includes('Android')) {
    os = 'Android';
  } else if (userAgent.includes('iOS') || userAgent.includes('iPhone') || userAgent.includes('iPad')) {
    os = 'iOS';
  }

  return { browser, os };
}

// Get all event categories for filter dropdown
export function getEventCategories(): Array<{ value: string; label: string }> {
  return Object.entries(categoryLabels).map(([value, label]) => ({
    value,
    label,
  }));
}

// Get event types for a category
export function getEventTypesByCategory(category: string): Array<{ value: string; label: string }> {
  const categoryEvents: Record<string, string[]> = {
    authentication: ['login', 'login_failed', 'logout', 'register', 'password_change', 'token_refresh'],
    api_key: ['api_key_created', 'api_key_deleted', 'api_key_used'],
    session: ['session_created', 'session_terminated', 'session_expired'],
    mfa: ['mfa_enabled', 'mfa_disabled', 'mfa_verified', 'mfa_failed', 'backup_code_used'],
    provider: ['provider_key_set', 'provider_key_deleted'],
    settings: ['settings_changed'],
    github: ['github_connected', 'github_disconnected'],
  };

  const events = categoryEvents[category] || [];
  return events.map((value) => ({
    value,
    label: getEventLabel(value),
  }));
}

// Status indicator component props helper
export function getStatusIndicator(success: boolean): { icon: LucideIcon; className: string } {
  return success
    ? { icon: CheckCircle, className: 'text-green-500' }
    : { icon: XCircle, className: 'text-red-500' };
}
