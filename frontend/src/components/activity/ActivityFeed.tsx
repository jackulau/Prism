import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { useMonitoringStore } from '../../store/monitoringStore';
import type { ActivityEvent, ActivityEventType, ActivityEventSeverity } from '../../types/monitoring';
import { ActivityEventItem } from './ActivityEventItem';
import { ActivityFeedHeader } from './ActivityFeedHeader';
import { ArrowDown } from 'lucide-react';

export type EventTypeFilter = 'all' | 'agent' | 'swarm' | 'build' | 'connection' | 'notification' | 'system';
export type EventStatusFilter = 'all' | 'success' | 'warning' | 'error' | 'info';

interface ActivityFeedProps {
  maxHeight?: string;
  showHeader?: boolean;
  initialTypeFilter?: EventTypeFilter;
  initialStatusFilter?: EventStatusFilter;
}

const MAX_VISIBLE_EVENTS = 100;
const NEW_EVENT_ANIMATION_DURATION = 2000;

function matchesTypeFilter(eventType: ActivityEventType, filter: EventTypeFilter): boolean {
  if (filter === 'all') return true;
  return eventType.startsWith(`${filter}.`);
}

function matchesStatusFilter(severity: ActivityEventSeverity, filter: EventStatusFilter): boolean {
  if (filter === 'all') return true;
  return severity === filter;
}

function matchesSearch(event: ActivityEvent, search: string): boolean {
  if (!search.trim()) return true;
  const searchLower = search.toLowerCase();
  return (
    event.title.toLowerCase().includes(searchLower) ||
    (event.description?.toLowerCase().includes(searchLower) ?? false) ||
    event.type.toLowerCase().includes(searchLower)
  );
}

export function ActivityFeed({
  maxHeight = '400px',
  showHeader = true,
  initialTypeFilter = 'all',
  initialStatusFilter = 'all',
}: ActivityFeedProps) {
  const activityEvents = useMonitoringStore((state) => state.activityEvents);
  const clearAllEvents = useMonitoringStore((state) => state.clearAllEvents);
  const wsStatus = useMonitoringStore((state) => state.wsStatus);

  const [typeFilter, setTypeFilter] = useState<EventTypeFilter>(initialTypeFilter);
  const [statusFilter, setStatusFilter] = useState<EventStatusFilter>(initialStatusFilter);
  const [searchQuery, setSearchQuery] = useState('');
  const [isPaused, setIsPaused] = useState(false);
  const [newEventsCount, setNewEventsCount] = useState(0);
  const [isUserScrolled, setIsUserScrolled] = useState(false);
  const [newEventIds, setNewEventIds] = useState<Set<string>>(new Set());

  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const lastEventIdRef = useRef<string | null>(null);
  const pausedEventsRef = useRef<ActivityEvent[]>([]);

  // Filter events
  const filteredEvents = useMemo(() => {
    let events = isPaused ? pausedEventsRef.current : activityEvents;

    return events.filter((event) => {
      return (
        matchesTypeFilter(event.type, typeFilter) &&
        matchesStatusFilter(event.severity, statusFilter) &&
        matchesSearch(event, searchQuery)
      );
    }).slice(0, MAX_VISIBLE_EVENTS);
  }, [activityEvents, typeFilter, statusFilter, searchQuery, isPaused]);

  // Track new events for animation
  useEffect(() => {
    if (isPaused) return;

    const latestEvent = activityEvents[0];
    if (latestEvent && latestEvent.id !== lastEventIdRef.current) {
      // Mark event as new
      setNewEventIds((prev) => new Set(prev).add(latestEvent.id));
      lastEventIdRef.current = latestEvent.id;

      // Remove animation after duration
      const timeout = setTimeout(() => {
        setNewEventIds((prev) => {
          const next = new Set(prev);
          next.delete(latestEvent.id);
          return next;
        });
      }, NEW_EVENT_ANIMATION_DURATION);

      return () => clearTimeout(timeout);
    }
  }, [activityEvents, isPaused]);

  // Track paused events and count new ones
  useEffect(() => {
    if (isPaused) {
      // Capture current events when pausing
      pausedEventsRef.current = activityEvents;
    } else {
      // Reset new events count when unpausing
      setNewEventsCount(0);
      pausedEventsRef.current = [];
    }
  }, [isPaused, activityEvents]);

  // Count new events while paused
  useEffect(() => {
    if (isPaused) {
      const pausedCount = pausedEventsRef.current.length;
      const currentCount = activityEvents.length;
      const newCount = Math.max(0, currentCount - pausedCount);
      setNewEventsCount(newCount);
    }
  }, [activityEvents.length, isPaused]);

  // Handle scroll to detect if user has scrolled up
  const handleScroll = useCallback(() => {
    const container = scrollContainerRef.current;
    if (!container) return;

    const { scrollTop, scrollHeight, clientHeight } = container;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
    setIsUserScrolled(!isAtBottom);
  }, []);

  // Auto-scroll to top when new events arrive (if not manually scrolled)
  useEffect(() => {
    if (!isPaused && !isUserScrolled && scrollContainerRef.current) {
      scrollContainerRef.current.scrollTop = 0;
    }
  }, [activityEvents.length, isPaused, isUserScrolled]);

  // Scroll to top and resume
  const scrollToLatest = useCallback(() => {
    setIsPaused(false);
    setIsUserScrolled(false);
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollTop = 0;
    }
  }, []);

  const handleClear = useCallback(() => {
    clearAllEvents();
    setNewEventsCount(0);
    pausedEventsRef.current = [];
  }, [clearAllEvents]);

  const handleTogglePause = useCallback(() => {
    if (isPaused) {
      // Resuming
      setNewEventsCount(0);
    }
    setIsPaused(!isPaused);
  }, [isPaused]);

  const isLive = wsStatus === 'connected' && !isPaused;

  return (
    <div className="flex flex-col h-full">
      {showHeader && (
        <ActivityFeedHeader
          typeFilter={typeFilter}
          statusFilter={statusFilter}
          searchQuery={searchQuery}
          isPaused={isPaused}
          isLive={isLive}
          onTypeFilterChange={setTypeFilter}
          onStatusFilterChange={setStatusFilter}
          onSearchChange={setSearchQuery}
          onTogglePause={handleTogglePause}
          onClear={handleClear}
        />
      )}

      <div className="relative flex-1 min-h-0">
        {/* Events list */}
        <div
          ref={scrollContainerRef}
          onScroll={handleScroll}
          className="overflow-y-auto h-full"
          style={{ maxHeight }}
        >
          {filteredEvents.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-32 text-editor-muted">
              <p className="text-sm">No activity events</p>
              <p className="text-xs mt-1">Events will appear here as they occur</p>
            </div>
          ) : (
            <div className="divide-y divide-editor-border">
              {filteredEvents.map((event) => (
                <ActivityEventItem
                  key={event.id}
                  event={event}
                  isNew={newEventIds.has(event.id)}
                />
              ))}
            </div>
          )}
        </div>

        {/* New events badge when scrolled away or paused */}
        {(isUserScrolled || (isPaused && newEventsCount > 0)) && (
          <button
            onClick={scrollToLatest}
            className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-1.5 bg-editor-accent text-white text-xs font-medium rounded-full shadow-lg hover:bg-editor-accent/90 transition-all animate-bounce-subtle"
          >
            <ArrowDown size={12} />
            {newEventsCount > 0 ? (
              <span>{newEventsCount} new {newEventsCount === 1 ? 'event' : 'events'}</span>
            ) : (
              <span>Scroll to latest</span>
            )}
          </button>
        )}
      </div>
    </div>
  );
}
