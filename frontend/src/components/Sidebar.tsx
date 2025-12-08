import React from 'react';
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
} from 'lucide-react';
import { useAppStore } from '../store';

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

  const handleNewChat = () => {
    setCurrentConversationId(null);
    clearMessages();
  };

  // Demo conversations
  const demoConversations = conversations.length > 0 ? conversations : [
    { id: '1', title: 'React Component Help', createdAt: new Date(), updatedAt: new Date(), messageCount: 5 },
    { id: '2', title: 'API Integration', createdAt: new Date(Date.now() - 86400000), updatedAt: new Date(Date.now() - 86400000), messageCount: 12 },
    { id: '3', title: 'TypeScript Types', createdAt: new Date(Date.now() - 172800000), updatedAt: new Date(Date.now() - 172800000), messageCount: 8 },
  ];

  const formatDate = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return 'Today';
    if (days === 1) return 'Yesterday';
    if (days < 7) return `${days} days ago`;
    return date.toLocaleDateString();
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

      {/* Conversations List */}
      <div className="flex-1 overflow-y-auto px-3">
        <div className="mb-2">
          <span className="text-xs text-editor-muted uppercase tracking-wide px-2">
            Recent Chats
          </span>
        </div>
        <div className="space-y-1">
          {demoConversations.map((conv) => (
            <button
              key={conv.id}
              onClick={() => setCurrentConversationId(conv.id)}
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
                  <span>{formatDate(conv.updatedAt)}</span>
                  <span>·</span>
                  <span>{conv.messageCount} msgs</span>
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

      {/* Footer */}
      <div className="p-3 border-t border-editor-border space-y-1">
        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth">
          <Settings className="w-4 h-4" />
          <span className="text-sm">Settings</span>
        </button>
        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-text transition-smooth">
          <User className="w-4 h-4" />
          <span className="text-sm">Profile</span>
        </button>
        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-sidebar-hover text-editor-muted hover:text-editor-error transition-smooth">
          <LogOut className="w-4 h-4" />
          <span className="text-sm">Sign Out</span>
        </button>
      </div>
    </div>
  );
};
