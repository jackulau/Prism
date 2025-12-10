import React, { useState, useEffect, useMemo } from 'react';
import {
  MessageSquare,
  Plus,
  Settings,
  User,
  LogOut,
  ChevronLeft,
  ChevronRight,
  Trash2,
  Clock,
  Sparkles,
  Search,
  X,
  Loader2,
} from 'lucide-react';
import { useAppStore } from '../store';
import { apiService } from '../services/api';
import { logoutUser } from '../store/authStore';

// Conversation type for sidebar
interface SidebarConversation {
  id: string;
  title: string;
  createdAt: Date;
  updatedAt: Date;
  messageCount?: number;
}

// Group conversations by date
type ConversationGroup = {
  label: string;
  conversations: SidebarConversation[];
};

const groupConversationsByDate = (conversations: SidebarConversation[]): ConversationGroup[] => {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today.getTime() - 86400000);
  const lastWeek = new Date(today.getTime() - 7 * 86400000);

  const groups: { [key: string]: SidebarConversation[] } = {
    Today: [],
    Yesterday: [],
    'Last 7 Days': [],
    Older: [],
  };

  conversations.forEach((conv) => {
    const convDate = new Date(conv.updatedAt);
    if (convDate >= today) {
      groups['Today'].push(conv);
    } else if (convDate >= yesterday) {
      groups['Yesterday'].push(conv);
    } else if (convDate >= lastWeek) {
      groups['Last 7 Days'].push(conv);
    } else {
      groups['Older'].push(conv);
    }
  });

  // Return only non-empty groups
  return Object.entries(groups)
    .filter(([_, convs]) => convs.length > 0)
    .map(([label, conversations]) => ({ label, conversations }));
};

export const Sidebar: React.FC = () => {
  const {
    isSidebarOpen,
    toggleSidebar,
    conversations,
    currentConversationId,
    setCurrentConversationId,
    clearMessages,
    deleteConversation,
  } = useAppStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SidebarConversation[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [debouncedQuery, setDebouncedQuery] = useState('');

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Search when debounced query changes
  useEffect(() => {
    const search = async () => {
      if (!debouncedQuery.trim()) {
        setSearchResults([]);
        return;
      }

      setIsSearching(true);
      try {
        const response = await apiService.searchConversations(debouncedQuery);
        if (response.data?.conversations) {
          setSearchResults(
            response.data.conversations.map((c) => ({
              id: c.id,
              title: c.title || 'Untitled',
              createdAt: new Date(c.created_at),
              updatedAt: new Date(c.updated_at),
            }))
          );
        }
      } catch {
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    };

    search();
  }, [debouncedQuery]);

  const handleNewChat = () => {
    setCurrentConversationId(null);
    clearMessages();
    setSearchQuery('');
    setSearchResults([]);
  };

  const clearSearch = () => {
    setSearchQuery('');
    setSearchResults([]);
  };

  // Convert store conversations to sidebar format
  const sidebarConversations = useMemo((): SidebarConversation[] => {
    if (conversations.length > 0) {
      return conversations.map((c) => ({
        id: c.id,
        title: c.title,
        createdAt: c.createdAt,
        updatedAt: c.updatedAt,
        messageCount: c.messageCount,
      }));
    }
    // Demo conversations for empty state
    return [
      { id: '1', title: 'React Component Help', createdAt: new Date(), updatedAt: new Date(), messageCount: 5 },
      { id: '2', title: 'API Integration', createdAt: new Date(Date.now() - 86400000), updatedAt: new Date(Date.now() - 86400000), messageCount: 12 },
      { id: '3', title: 'TypeScript Types', createdAt: new Date(Date.now() - 172800000), updatedAt: new Date(Date.now() - 172800000), messageCount: 8 },
    ];
  }, [conversations]);

  // Group conversations for display
  const groupedConversations = useMemo(() => {
    const toGroup = searchQuery.trim() ? searchResults : sidebarConversations;
    return groupConversationsByDate(toGroup);
  }, [searchQuery, searchResults, sidebarConversations]);

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (!isSidebarOpen) {
    return (
      <div className="w-12 bg-sidebar-bg border-r border-editor-border flex flex-col items-center py-4">
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth"
          title="Expand sidebar"
        >
          <ChevronRight className="w-5 h-5" />
        </button>
        <div className="mt-4">
          <button
            onClick={handleNewChat}
            className="p-2 rounded-lg bg-editor-accent text-white hover:bg-editor-accent/80 transition-smooth"
            title="New chat"
          >
            <Plus className="w-5 h-5" />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="w-64 bg-sidebar-bg border-r border-editor-border flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-editor-border flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-editor-accent to-purple-500 flex items-center justify-center">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          <span className="font-semibold text-editor-text">Prism</span>
        </div>
        <button
          onClick={toggleSidebar}
          className="p-1.5 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth"
          title="Collapse sidebar"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>
      </div>

      {/* New Chat Button */}
      <div className="p-3">
        <button
          onClick={handleNewChat}
          className="w-full flex items-center gap-2 px-3 py-2.5 rounded-lg border border-editor-border hover:bg-sidebar-hover text-editor-text transition-smooth"
        >
          <Plus className="w-4 h-4" />
          <span className="text-sm">New Chat</span>
        </button>
      </div>

      {/* Search Input */}
      <div className="px-3 pb-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-editor-muted" />
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-editor-surface text-editor-text text-sm pl-8 pr-8 py-2 rounded-lg border border-editor-border focus:border-editor-accent focus:outline-none transition-smooth"
          />
          {searchQuery && (
            <button
              onClick={clearSearch}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 p-0.5 hover:bg-sidebar-hover rounded"
            >
              {isSearching ? (
                <Loader2 className="w-3.5 h-3.5 text-editor-muted animate-spin" />
              ) : (
                <X className="w-3.5 h-3.5 text-editor-muted hover:text-editor-text" />
              )}
            </button>
          )}
        </div>
      </div>

      {/* Conversations List - Grouped by Date */}
      <div className="flex-1 overflow-y-auto px-3">
        {groupedConversations.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-sm text-editor-muted">
              {searchQuery ? 'No conversations found' : 'No conversations yet'}
            </p>
          </div>
        ) : (
          groupedConversations.map((group) => (
            <div key={group.label} className="mb-4">
              <div className="mb-2 sticky top-0 bg-sidebar-bg py-1">
                <span className="text-xs text-editor-muted uppercase tracking-wide px-2">
                  {group.label}
                </span>
              </div>
              <div className="space-y-1">
                {group.conversations.map((conv) => (
                  <button
                    key={conv.id}
                    onClick={() => {
                      setCurrentConversationId(conv.id);
                      clearSearch();
                    }}
                    className={`w-full flex items-start gap-2 px-2 py-2 rounded-lg text-left transition-smooth group ${
                      currentConversationId === conv.id
                        ? 'bg-editor-accent/20 text-editor-accent'
                        : 'hover:bg-sidebar-hover text-editor-text'
                    }`}
                  >
                    <MessageSquare className="w-4 h-4 mt-0.5 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm truncate">{conv.title}</div>
                      <div className="flex items-center gap-2 text-xs text-editor-muted">
                        <Clock className="w-3 h-3" />
                        <span>{formatTime(conv.updatedAt)}</span>
                        {conv.messageCount !== undefined && (
                          <>
                            <span>·</span>
                            <span>{conv.messageCount} msgs</span>
                          </>
                        )}
                      </div>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        deleteConversation(conv.id);
                      }}
                      className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-editor-error/20 text-editor-muted hover:text-editor-error transition-smooth"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </button>
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Footer */}
      <div className="p-3 border-t border-editor-border space-y-1">
        <button
          onClick={() => window.location.href = '/settings'}
          className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth"
        >
          <Settings className="w-4 h-4" />
          <span className="text-sm">Settings</span>
        </button>
        <button
          onClick={() => window.location.href = '/profile'}
          className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth"
        >
          <User className="w-4 h-4" />
          <span className="text-sm">Profile</span>
        </button>
        <button
          onClick={async () => {
            await logoutUser();
            window.location.href = '/login';
          }}
          className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-error transition-smooth"
        >
          <LogOut className="w-4 h-4" />
          <span className="text-sm">Sign Out</span>
        </button>
      </div>
    </div>
  );
};
