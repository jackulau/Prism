import React, { useState, useCallback, useRef, useEffect } from 'react';

interface ResizablePanelGroupProps {
  children: React.ReactNode;
  direction?: 'horizontal' | 'vertical';
  className?: string;
}

interface ResizablePanelProps {
  children: React.ReactNode;
  defaultSize?: number;
  minSize?: number;
  maxSize?: number;
  collapsible?: boolean;
  collapsed?: boolean;
  onCollapse?: () => void;
  onExpand?: () => void;
  onResize?: (size: number) => void;
  className?: string;
  order?: number;
}

interface ResizableHandleProps {
  direction?: 'horizontal' | 'vertical';
  className?: string;
  onDoubleClick?: () => void;
}

const ResizablePanelContext = React.createContext<{
  direction: 'horizontal' | 'vertical';
  registerPanel: (id: string, panel: { minSize: number; maxSize: number }) => void;
  unregisterPanel: (id: string) => void;
}>({
  direction: 'horizontal',
  registerPanel: () => {},
  unregisterPanel: () => {},
});

export function ResizablePanelGroup({
  children,
  direction = 'horizontal',
  className = '',
}: ResizablePanelGroupProps) {
  const [panels] = useState<Map<string, { minSize: number; maxSize: number }>>(new Map());

  const registerPanel = useCallback((id: string, panel: { minSize: number; maxSize: number }) => {
    panels.set(id, panel);
  }, [panels]);

  const unregisterPanel = useCallback((id: string) => {
    panels.delete(id);
  }, [panels]);

  return (
    <ResizablePanelContext.Provider value={{ direction, registerPanel, unregisterPanel }}>
      <div
        className={`flex ${direction === 'horizontal' ? 'flex-row' : 'flex-col'} h-full w-full ${className}`}
      >
        {children}
      </div>
    </ResizablePanelContext.Provider>
  );
}

export function ResizablePanel({
  children,
  defaultSize,
  minSize = 0,
  maxSize = Infinity,
  collapsible = false,
  collapsed = false,
  onCollapse: _onCollapse,
  onExpand: _onExpand,
  onResize,
  className = '',
  order,
}: ResizablePanelProps) {
  // onCollapse and onExpand are reserved for future use
  void _onCollapse;
  void _onExpand;

  const { direction, registerPanel, unregisterPanel } = React.useContext(ResizablePanelContext);
  const [size] = useState(defaultSize);
  const panelId = useRef(`panel-${Math.random().toString(36).slice(2)}`);

  useEffect(() => {
    registerPanel(panelId.current, { minSize, maxSize });
    return () => unregisterPanel(panelId.current);
  }, [minSize, maxSize, registerPanel, unregisterPanel]);

  useEffect(() => {
    if (size !== undefined && onResize) {
      onResize(size);
    }
  }, [size, onResize]);

  const style: React.CSSProperties = {
    order,
  };

  if (collapsed && collapsible) {
    style[direction === 'horizontal' ? 'width' : 'height'] = 0;
    style.overflow = 'hidden';
  } else if (size !== undefined) {
    style[direction === 'horizontal' ? 'width' : 'height'] = size;
    style.flexShrink = 0;
  } else {
    style.flex = 1;
  }

  return (
    <div
      data-panel-id={panelId.current}
      className={`overflow-hidden ${className}`}
      style={style}
    >
      {children}
    </div>
  );
}

export function ResizableHandle({
  direction: propDirection,
  className = '',
  onDoubleClick,
}: ResizableHandleProps) {
  const { direction: contextDirection } = React.useContext(ResizablePanelContext);
  const direction = propDirection || contextDirection;

  const [isDragging, setIsDragging] = useState(false);
  const handleRef = useRef<HTMLDivElement>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);

    const handle = handleRef.current;
    if (!handle) return;

    const parent = handle.parentElement;
    if (!parent) return;

    const prevPanel = handle.previousElementSibling as HTMLElement;
    const nextPanel = handle.nextElementSibling as HTMLElement;

    if (!prevPanel || !nextPanel) return;

    const startPos = direction === 'horizontal' ? e.clientX : e.clientY;
    const prevSize = direction === 'horizontal' ? prevPanel.offsetWidth : prevPanel.offsetHeight;
    const nextSize = direction === 'horizontal' ? nextPanel.offsetWidth : nextPanel.offsetHeight;

    const handleMouseMove = (moveEvent: MouseEvent) => {
      const currentPos = direction === 'horizontal' ? moveEvent.clientX : moveEvent.clientY;
      const delta = currentPos - startPos;

      const newPrevSize = Math.max(0, prevSize + delta);
      const newNextSize = Math.max(0, nextSize - delta);

      prevPanel.style[direction === 'horizontal' ? 'width' : 'height'] = `${newPrevSize}px`;
      prevPanel.style.flexShrink = '0';
      nextPanel.style[direction === 'horizontal' ? 'width' : 'height'] = `${newNextSize}px`;
      nextPanel.style.flexShrink = '0';
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }, [direction]);

  const isHorizontal = direction === 'horizontal';

  return (
    <div
      ref={handleRef}
      className={`
        ${isHorizontal ? 'w-1 cursor-col-resize' : 'h-1 cursor-row-resize'}
        ${isDragging ? 'bg-editor-accent' : 'bg-editor-border hover:bg-editor-accent/50'}
        transition-colors flex-shrink-0 relative group
        ${className}
      `}
      onMouseDown={handleMouseDown}
      onDoubleClick={onDoubleClick}
    >
      {/* Hit area - larger invisible area for easier grabbing */}
      <div
        className={`
          absolute
          ${isHorizontal ? '-left-1 -right-1 top-0 bottom-0' : 'left-0 right-0 -top-1 -bottom-1'}
        `}
      />
      {/* Visual indicator on hover */}
      <div
        className={`
          absolute opacity-0 group-hover:opacity-100 transition-opacity
          ${isHorizontal
            ? 'left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-1 h-8 rounded-full bg-editor-accent/50'
            : 'top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 h-1 w-8 rounded-full bg-editor-accent/50'
          }
        `}
      />
    </div>
  );
}
