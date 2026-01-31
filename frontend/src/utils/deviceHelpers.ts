import { Laptop, Smartphone, Tablet, Monitor, type LucideIcon } from 'lucide-react';

export type DeviceType = 'desktop' | 'mobile' | 'tablet';

export interface DeviceInfo {
  browser: string;
  os: string;
  device: DeviceType;
}

/**
 * Returns an appropriate icon based on the device name/type
 */
export function getDeviceIcon(deviceName: string): LucideIcon {
  const lowerName = deviceName.toLowerCase();

  if (lowerName.includes('iphone') || lowerName.includes('android') || lowerName.includes('mobile')) {
    return Smartphone;
  }

  if (lowerName.includes('ipad') || lowerName.includes('tablet')) {
    return Tablet;
  }

  if (lowerName.includes('mac') || lowerName.includes('windows') || lowerName.includes('linux')) {
    return Laptop;
  }

  return Monitor;
}

/**
 * Parses a user agent string to extract browser, OS, and device type
 */
export function parseDeviceInfo(userAgent: string): DeviceInfo {
  const ua = userAgent.toLowerCase();

  // Determine device type
  let device: DeviceType = 'desktop';
  if (/iphone|ipod|android.*mobile|windows phone|blackberry|webos/i.test(ua)) {
    device = 'mobile';
  } else if (/ipad|android(?!.*mobile)|tablet/i.test(ua)) {
    device = 'tablet';
  }

  // Determine OS
  let os = 'Unknown';
  if (/windows nt 10/i.test(ua)) {
    os = 'Windows 10/11';
  } else if (/windows nt/i.test(ua)) {
    os = 'Windows';
  } else if (/mac os x/i.test(ua)) {
    if (/iphone|ipod/i.test(ua)) {
      os = 'iOS';
    } else if (/ipad/i.test(ua)) {
      os = 'iPadOS';
    } else {
      os = 'macOS';
    }
  } else if (/android/i.test(ua)) {
    os = 'Android';
  } else if (/linux/i.test(ua)) {
    os = 'Linux';
  } else if (/cros/i.test(ua)) {
    os = 'Chrome OS';
  }

  // Determine browser
  let browser = 'Unknown';
  if (/edg\//i.test(ua)) {
    browser = 'Edge';
  } else if (/opr\//i.test(ua) || /opera/i.test(ua)) {
    browser = 'Opera';
  } else if (/chrome/i.test(ua) && !/edg/i.test(ua)) {
    browser = 'Chrome';
  } else if (/safari/i.test(ua) && !/chrome/i.test(ua)) {
    browser = 'Safari';
  } else if (/firefox/i.test(ua)) {
    browser = 'Firefox';
  } else if (/msie|trident/i.test(ua)) {
    browser = 'Internet Explorer';
  }

  return { browser, os, device };
}

/**
 * Generates a human-readable device name from user agent
 */
export function getDeviceName(userAgent: string): string {
  const { browser, os } = parseDeviceInfo(userAgent);
  return `${browser} on ${os}`;
}

/**
 * Formats a relative time string from a date
 */
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
  } else if (diffDays === 1) {
    return 'Yesterday';
  } else if (diffDays < 7) {
    return `${diffDays} days ago`;
  } else {
    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    });
  }
}

/**
 * Formats a date for session display (e.g., "Today at 9:15 AM")
 */
export function formatSessionDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = date.toDateString() === yesterday.toDateString();

  const timeStr = date.toLocaleTimeString(undefined, {
    hour: 'numeric',
    minute: '2-digit',
  });

  if (isToday) {
    return `Today at ${timeStr}`;
  } else if (isYesterday) {
    return `Yesterday at ${timeStr}`;
  } else {
    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    }) + ` at ${timeStr}`;
  }
}
