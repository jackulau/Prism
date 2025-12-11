import { FileContextBadge } from './FileContextBadge';
import { ModeSwitcher } from './ModeSwitcher';
import { ThinkingToggle } from './ThinkingToggle';

export function ChatContextBar() {
  return (
    <div className="flex items-center gap-2 px-3 py-2 border-b border-editor-border bg-editor-bg/50">
      <FileContextBadge />
      <div className="flex-1" />
      <ModeSwitcher />
      <ThinkingToggle />
    </div>
  );
}
