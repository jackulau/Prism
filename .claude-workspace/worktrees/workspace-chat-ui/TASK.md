---
id: workspace-chat-ui
name: Workspace Chat UI
wave: 1
priority: 1
dependencies: []
estimated_hours: 6
tags:
- frontend
- ui
- chat
---

## Objective

Create a comprehensive workspace-aware chat UI that integrates file context, project navigation, and AI assistance in a unified interface similar to modern AI coding assistants.

## Context

The codebase has existing chat components in `frontend/src/components/chat/`:
- `EnhancedChatPanel.tsx` - Main chat panel with streaming, tool calls, commands
- `ToolCallCard.tsx` - Tool execution display
- `MessageQueue.tsx` - Message queue management
- `ChatInterface.tsx` - Basic chat interface

We need to create a workspace-focused chat experience with:
- File tree integration
- Code context awareness
- Split-pane layouts
- File preview in chat
- Multi-file editing suggestions
- Project-level operations

## Implementation

### 1. Create Workspace Chat Layout

**File**: `frontend/src/components/workspace/WorkspaceChatLayout.tsx`

```tsx
interface WorkspaceChatLayoutProps {
    workspace: Workspace;
    onFileSelect: (path: string) => void;
}

export function WorkspaceChatLayout({ workspace, onFileSelect }: WorkspaceChatLayoutProps) {
    return (
        <div className="flex h-full">
            {/* Left: File Tree */}
            <ResizablePanel defaultSize={20} minSize={15} maxSize={35}>
                <FileTree
                    files={workspace.files}
                    onSelect={onFileSelect}
                    selectedFiles={selectedContext}
                />
            </ResizablePanel>

            {/* Center: Chat + Context */}
            <ResizablePanel defaultSize={50}>
                <WorkspaceChat
                    workspace={workspace}
                    contextFiles={selectedContext}
                />
            </ResizablePanel>

            {/* Right: Preview/Editor */}
            <ResizablePanel defaultSize={30}>
                <FilePreview file={activeFile} />
            </ResizablePanel>
        </div>
    );
}
```

### 2. Create Workspace Chat Component

**File**: `frontend/src/components/workspace/WorkspaceChat.tsx`

```tsx
interface WorkspaceChatProps {
    workspace: Workspace;
    contextFiles: string[];
}

export function WorkspaceChat({ workspace, contextFiles }: WorkspaceChatProps) {
    const [messages, setMessages] = useState<Message[]>([]);
    const [isStreaming, setIsStreaming] = useState(false);

    return (
        <div className="flex flex-col h-full">
            {/* Context Bar */}
            <ContextBar files={contextFiles} onRemove={removeContext} />

            {/* Messages */}
            <MessageList
                messages={messages}
                onEditSuggestion={handleEditSuggestion}
                onApplyDiff={handleApplyDiff}
            />

            {/* Input */}
            <ChatInput
                onSend={handleSend}
                onFileAdd={addFileContext}
                placeholder="Ask about your code..."
            />
        </div>
    );
}
```

### 3. Create Context Bar Component

**File**: `frontend/src/components/workspace/ContextBar.tsx`

```tsx
interface ContextBarProps {
    files: FileContext[];
    onRemove: (path: string) => void;
    onAdd: () => void;
}

export function ContextBar({ files, onRemove, onAdd }: ContextBarProps) {
    return (
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-muted/50">
            <span className="text-sm text-muted-foreground">Context:</span>
            {files.map(file => (
                <ContextChip
                    key={file.path}
                    file={file}
                    onRemove={() => onRemove(file.path)}
                />
            ))}
            <AddContextButton onClick={onAdd} />
        </div>
    );
}
```

### 4. Create Code Diff Component

**File**: `frontend/src/components/workspace/CodeDiff.tsx`

```tsx
interface CodeDiffProps {
    original: string;
    modified: string;
    language: string;
    onAccept: () => void;
    onReject: () => void;
}

export function CodeDiff({ original, modified, language, onAccept, onReject }: CodeDiffProps) {
    return (
        <div className="rounded-lg border overflow-hidden">
            <div className="flex items-center justify-between px-3 py-2 bg-muted/50">
                <span className="text-sm font-medium">Suggested Changes</span>
                <div className="flex gap-2">
                    <Button size="sm" variant="ghost" onClick={onReject}>
                        Reject
                    </Button>
                    <Button size="sm" variant="default" onClick={onAccept}>
                        Accept
                    </Button>
                </div>
            </div>
            <DiffEditor
                original={original}
                modified={modified}
                language={language}
                options={{ readOnly: true }}
            />
        </div>
    );
}
```

### 5. Create File Preview Component

**File**: `frontend/src/components/workspace/FilePreview.tsx`

```tsx
interface FilePreviewProps {
    file: FileContext | null;
    highlights?: Range[];
    onEdit?: (content: string) => void;
}

export function FilePreview({ file, highlights, onEdit }: FilePreviewProps) {
    if (!file) {
        return <EmptyState message="Select a file to preview" />;
    }

    return (
        <div className="h-full flex flex-col">
            <div className="flex items-center justify-between px-4 py-2 border-b">
                <div className="flex items-center gap-2">
                    <FileIcon path={file.path} />
                    <span className="text-sm font-medium">{file.path}</span>
                </div>
                <CopyButton content={file.content} />
            </div>
            <CodeEditor
                value={file.content}
                language={detectLanguage(file.path)}
                highlights={highlights}
                onChange={onEdit}
                readOnly={!onEdit}
            />
        </div>
    );
}
```

### 6. Create Message with Code References

**File**: `frontend/src/components/workspace/WorkspaceMessage.tsx`

```tsx
interface WorkspaceMessageProps {
    message: Message;
    onFileClick: (path: string, line?: number) => void;
    onApplyEdit: (edit: EditSuggestion) => void;
}

export function WorkspaceMessage({ message, onFileClick, onApplyEdit }: WorkspaceMessageProps) {
    const parts = parseMessageWithRefs(message.content);

    return (
        <div className="message">
            {parts.map((part, i) => {
                if (part.type === 'text') {
                    return <Markdown key={i}>{part.content}</Markdown>;
                }
                if (part.type === 'file-ref') {
                    return (
                        <FileReference
                            key={i}
                            path={part.path}
                            line={part.line}
                            onClick={() => onFileClick(part.path, part.line)}
                        />
                    );
                }
                if (part.type === 'code-edit') {
                    return (
                        <CodeDiff
                            key={i}
                            original={part.original}
                            modified={part.modified}
                            language={part.language}
                            onAccept={() => onApplyEdit(part)}
                        />
                    );
                }
            })}
        </div>
    );
}
```

### 7. Create Resizable Panel System

**File**: `frontend/src/components/ui/ResizablePanel.tsx`

```tsx
interface ResizablePanelGroupProps {
    children: React.ReactNode;
    direction: 'horizontal' | 'vertical';
}

interface ResizablePanelProps {
    children: React.ReactNode;
    defaultSize: number;
    minSize?: number;
    maxSize?: number;
    collapsible?: boolean;
}

export function ResizablePanelGroup({ children, direction }: ResizablePanelGroupProps)
export function ResizablePanel({ children, defaultSize, minSize, maxSize }: ResizablePanelProps)
export function ResizableHandle()
```

### 8. Create Workspace Store

**File**: `frontend/src/store/workspaceStore.ts`

```typescript
interface WorkspaceState {
    currentWorkspace: Workspace | null;
    contextFiles: FileContext[];
    activeFile: FileContext | null;
    openFiles: FileContext[];
    pendingEdits: Map<string, EditSuggestion>;

    // Actions
    setWorkspace: (workspace: Workspace) => void;
    addContextFile: (path: string) => void;
    removeContextFile: (path: string) => void;
    setActiveFile: (file: FileContext) => void;
    applyEdit: (path: string, edit: EditSuggestion) => void;
    rejectEdit: (path: string) => void;
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
    // ... implementation
}));
```

### 9. Add Workspace Types

**File**: `frontend/src/types/workspace.ts`

```typescript
export interface Workspace {
    id: string;
    name: string;
    path: string;
    files: FileNode[];
}

export interface FileContext {
    path: string;
    content: string;
    language: string;
    selection?: Range;
}

export interface EditSuggestion {
    path: string;
    original: string;
    modified: string;
    startLine: number;
    endLine: number;
    description: string;
}

export interface Range {
    startLine: number;
    startColumn: number;
    endLine: number;
    endColumn: number;
}
```

### 10. Add Keyboard Shortcuts

**File**: `frontend/src/hooks/useWorkspaceShortcuts.ts`

```typescript
export function useWorkspaceShortcuts() {
    useHotkeys('mod+k', () => openCommandPalette());
    useHotkeys('mod+p', () => openFilePicker());
    useHotkeys('mod+shift+e', () => toggleFileTree());
    useHotkeys('mod+enter', () => sendMessage());
    useHotkeys('escape', () => closePreview());
}
```

## Acceptance Criteria

- [ ] Three-pane layout (file tree, chat, preview)
- [ ] Resizable panels with persistence
- [ ] File context bar showing attached files
- [ ] Click-to-add file context from tree
- [ ] Code diff display with accept/reject
- [ ] File references in messages are clickable
- [ ] File preview with syntax highlighting
- [ ] Keyboard shortcuts for navigation
- [ ] Workspace state management with Zustand
- [ ] Responsive layout for different screen sizes

## Files to Create/Modify

- `frontend/src/components/workspace/WorkspaceChatLayout.tsx`
- `frontend/src/components/workspace/WorkspaceChat.tsx`
- `frontend/src/components/workspace/ContextBar.tsx`
- `frontend/src/components/workspace/CodeDiff.tsx`
- `frontend/src/components/workspace/FilePreview.tsx`
- `frontend/src/components/workspace/WorkspaceMessage.tsx`
- `frontend/src/components/ui/ResizablePanel.tsx`
- `frontend/src/store/workspaceStore.ts`
- `frontend/src/types/workspace.ts`
- `frontend/src/hooks/useWorkspaceShortcuts.ts`

## Integration Points

- **Provides**: Workspace-aware chat UI
- **Provides**: File context management
- **Consumes**: WebSocket/SSE for streaming
- **Consumes**: File API for workspace access
- **Consumes**: Existing chat components (can extend EnhancedChatPanel)
- **Conflicts**: May overlap with sandbox components - coordinate file tree usage
