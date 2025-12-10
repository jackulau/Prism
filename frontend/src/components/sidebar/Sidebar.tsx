import { useState, useEffect } from 'react'
import { MessageSquare, Settings, FolderTree, ChevronLeft, ChevronRight, Plus, Github, Trash2, Cpu, Loader2, type LucideIcon } from 'lucide-react'
import { useAppStore } from '../../store'
import { ConfirmDialog } from '../ConfirmDialog'
import { toast } from '../../store/toastStore'

interface SidebarProps {
  isCollapsed: boolean
  onToggle: () => void
}

export function Sidebar({ isCollapsed, onToggle }: SidebarProps) {
  const {
    isFileTreeOpen,
    toggleFileTree,
    toggleSettingsPanel,
    toggleChatPanel,
    conversations,
    currentConversationId,
    loadConversations,
    loadMessages,
    createNewConversation,
    deleteConversation,
    isLoadingConversations,
    conversationsError,
    clearConversationsError,
  } = useAppStore()
  const [activeView, setActiveView] = useState<'chat' | 'files' | 'settings'>('chat')
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null)

  // Load conversations on mount
  useEffect(() => {
    loadConversations()
  }, [loadConversations])

  const handleFilesClick = () => {
    setActiveView('files')
    toggleFileTree()
  }

  const handleSettingsClick = () => {
    setActiveView('settings')
    toggleSettingsPanel()
  }

  const handleChatClick = () => {
    setActiveView('chat')
    toggleChatPanel()
  }

  const handleNewChat = async () => {
    await createNewConversation()
  }

  const handleConversationClick = async (id: string) => {
    if (id !== currentConversationId) {
      await loadMessages(id)
    }
  }

  const handleDeleteConversation = (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    setDeleteConfirmId(id)
  }

  const confirmDelete = async () => {
    if (deleteConfirmId) {
      await deleteConversation(deleteConfirmId)
      toast.success('Conversation deleted')
      setDeleteConfirmId(null)
    }
  }

  return (
    <div
      className={`h-full bg-sidebar-bg border-r border-editor-border flex flex-col transition-all duration-200 ${
        isCollapsed ? 'w-14' : 'w-64'
      }`}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-editor-border">
        {!isCollapsed && (
          <div className="flex items-center gap-2">
            <img
              src="/logo.png"
              alt="Prism"
              className="w-8 h-8 rounded-lg object-contain"
            />
          </div>
        )}
        <button
          onClick={onToggle}
          className="p-1.5 rounded-md text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
        >
          {isCollapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-2 overflow-y-auto">
        <div className="space-y-1">
          <NavItem
            icon={MessageSquare}
            label="Chat"
            isCollapsed={isCollapsed}
            active={activeView === 'chat'}
            onClick={handleChatClick}
          />
          <NavItem
            icon={FolderTree}
            label="Files"
            isCollapsed={isCollapsed}
            active={isFileTreeOpen}
            onClick={handleFilesClick}
          />
          <NavItem
            icon={Settings}
            label="Settings"
            isCollapsed={isCollapsed}
            active={activeView === 'settings'}
            onClick={handleSettingsClick}
          />
        </div>

        {/* Conversations List */}
        {!isCollapsed && (
          <div className="mt-4">
            <div className="flex items-center justify-between px-2 mb-2">
              <span className="text-xs font-semibold text-editor-muted uppercase tracking-wider">
                Conversations
              </span>
              <button
                onClick={handleNewChat}
                className="p-1 rounded text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
                title="New Chat"
              >
                <Plus size={14} />
              </button>
            </div>

            {/* Error state */}
            {conversationsError && (
              <div className="px-2 py-2 mx-2 mb-2 bg-red-500/10 border border-red-500/20 rounded text-red-400 text-xs">
                {conversationsError}
                <button
                  onClick={() => {
                    clearConversationsError()
                    loadConversations()
                  }}
                  className="ml-2 underline hover:no-underline"
                >
                  Retry
                </button>
              </div>
            )}

            {/* Loading state */}
            {isLoadingConversations ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 size={16} className="animate-spin text-editor-muted" />
                <span className="ml-2 text-xs text-editor-muted">Loading...</span>
              </div>
            ) : (
              <div className="space-y-1">
                {conversations.length === 0 ? (
                  <button
                    onClick={handleNewChat}
                    className="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors"
                  >
                    <Plus size={14} />
                    <span>Start a new chat</span>
                  </button>
                ) : (
                  conversations.map((conv) => (
                    <ConversationItem
                      key={conv.id}
                      title={conv.title}
                      provider={conv.provider}
                      model={conv.model}
                      active={conv.id === currentConversationId}
                      onClick={() => handleConversationClick(conv.id)}
                      onDelete={(e) => handleDeleteConversation(e, conv.id)}
                    />
                  ))
                )}
              </div>
            )}
          </div>
        )}
      </nav>

      {/* Footer - Open Source Link */}
      <div className="p-3 border-t border-editor-border">
        <a
          href="https://github.com/jackulau/Prism"
          target="_blank"
          rel="noopener noreferrer"
          className={`flex items-center gap-2 px-3 py-2 rounded-md text-editor-muted hover:text-editor-text hover:bg-sidebar-hover transition-colors ${isCollapsed ? 'justify-center' : ''}`}
          title="View on GitHub"
        >
          <Github size={18} />
          {!isCollapsed && <span className="text-sm">Prism</span>}
        </a>
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteConfirmId !== null}
        title="Delete Conversation"
        message="Are you sure you want to delete this conversation? This action cannot be undone."
        confirmText="Delete"
        variant="danger"
        onConfirm={confirmDelete}
        onCancel={() => setDeleteConfirmId(null)}
      />
    </div>
  )
}

interface NavItemProps {
  icon: LucideIcon
  label: string
  isCollapsed: boolean
  active?: boolean
  onClick?: () => void
}

function NavItem({ icon: Icon, label, isCollapsed, active, onClick }: NavItemProps) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2 rounded-md transition-colors ${
        active
          ? 'bg-editor-accent/10 text-editor-accent'
          : 'text-editor-muted hover:text-editor-text hover:bg-sidebar-hover'
      } ${isCollapsed ? 'justify-center' : ''}`}
      title={isCollapsed ? label : undefined}
    >
      <Icon size={18} />
      {!isCollapsed && <span className="text-sm font-medium">{label}</span>}
    </button>
  )
}

interface ConversationItemProps {
  title: string
  provider?: string
  model?: string
  active?: boolean
  onClick?: () => void
  onDelete?: (e: React.MouseEvent) => void
}

function ConversationItem({ title, provider, model, active, onClick, onDelete }: ConversationItemProps) {
  const isOllama = provider === 'ollama'
  return (
    <button
      onClick={onClick}
      className={`w-full flex flex-col gap-0.5 px-3 py-2 rounded-md text-sm transition-colors group ${
        active
          ? 'bg-editor-accent/10 text-editor-accent'
          : 'text-editor-muted hover:text-editor-text hover:bg-sidebar-hover'
      }`}
    >
      <div className="flex items-center gap-2 w-full">
        <MessageSquare size={14} className="flex-shrink-0" />
        <span className="truncate flex-1 text-left">{title}</span>
        {onDelete && (
          <button
            onClick={onDelete}
            className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-editor-error/20 text-editor-muted hover:text-editor-error transition-all"
            title="Delete conversation"
          >
            <Trash2 size={12} />
          </button>
        )}
      </div>
      {provider && (
        <div className="flex items-center gap-1 ml-6 text-xs text-editor-muted/70">
          {isOllama ? (
            <Cpu size={10} className="text-editor-accent" />
          ) : null}
          <span className="truncate">{model || provider}</span>
        </div>
      )}
    </button>
  )
}
