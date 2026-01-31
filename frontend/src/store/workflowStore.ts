import { create } from 'zustand';
import {
  WorkflowNode,
  Edge,
  Position,
  StepType,
  StepConfig,
  PendingConnection,
  WorkflowDefinition,
  ZOOM_CONSTRAINTS,
  CANVAS_GRID,
  NODE_DIMENSIONS,
} from '../types/workflow';

interface WorkflowState {
  // Canvas state
  nodes: WorkflowNode[];
  edges: Edge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  zoom: number;
  pan: Position;

  // Editing state
  pendingConnection: PendingConnection | null;
  isDragging: boolean;
  isPanning: boolean;

  // Workflow metadata
  workflowId: string | null;
  workflowName: string;
  workflowDescription: string;
  isDirty: boolean;

  // History for undo/redo
  history: { nodes: WorkflowNode[]; edges: Edge[] }[];
  historyIndex: number;
}

interface WorkflowActions {
  // Node actions
  addNode: (type: StepType, position: Position, name?: string) => string;
  removeNode: (nodeId: string) => void;
  updateNode: (nodeId: string, updates: Partial<WorkflowNode>) => void;
  moveNode: (nodeId: string, position: Position) => void;
  selectNode: (nodeId: string | null) => void;
  duplicateNode: (nodeId: string) => string | null;

  // Edge actions
  addEdge: (source: string, target: string, sourcePort?: 'success' | 'failure' | 'default', label?: Edge['label']) => string | null;
  removeEdge: (edgeId: string) => void;
  selectEdge: (edgeId: string | null) => void;

  // Connection actions
  startConnection: (nodeId: string, port: 'success' | 'failure' | 'default') => void;
  updateConnectionPosition: (position: Position) => void;
  completeConnection: (targetNodeId: string) => void;
  cancelConnection: () => void;

  // Canvas actions
  setZoom: (zoom: number) => void;
  zoomIn: () => void;
  zoomOut: () => void;
  resetZoom: () => void;
  setPan: (pan: Position) => void;
  fitToView: () => void;
  setDragging: (isDragging: boolean) => void;
  setPanning: (isPanning: boolean) => void;

  // Workflow actions
  setWorkflowMeta: (id: string | null, name: string, description: string) => void;
  loadWorkflow: (definition: WorkflowDefinition) => void;
  clearCanvas: () => void;
  getWorkflowDefinition: () => WorkflowDefinition;

  // History actions
  undo: () => void;
  redo: () => void;
  saveToHistory: () => void;

  // Selection actions
  clearSelection: () => void;
  deleteSelected: () => void;
  selectAll: () => void;
}

type WorkflowStore = WorkflowState & WorkflowActions;

const generateId = () => `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
const generateEdgeId = () => `edge_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

const getDefaultConfig = (type: StepType): StepConfig => {
  switch (type) {
    case 'agent':
      return {
        agentConfig: {
          provider: 'anthropic',
          model: 'claude-3-sonnet-20240229',
          prompt: '',
        },
      };
    case 'tool':
      return {
        toolConfig: {
          toolName: '',
          parameters: {},
        },
      };
    case 'condition':
      return {
        conditionConfig: {
          expression: '',
          trueBranch: '',
          falseBranch: '',
        },
      };
    case 'parallel':
      return {
        parallelConfig: {
          steps: [],
          waitForAll: true,
          failOnFirst: false,
        },
      };
    case 'wait':
      return {
        waitConfig: {
          waitType: 'user_input',
          promptText: '',
        },
      };
    case 'transform':
      return {
        transformConfig: {
          type: 'template',
          template: '',
        },
      };
  }
};

const initialState: WorkflowState = {
  nodes: [],
  edges: [],
  selectedNodeId: null,
  selectedEdgeId: null,
  zoom: 1,
  pan: { x: 0, y: 0 },
  pendingConnection: null,
  isDragging: false,
  isPanning: false,
  workflowId: null,
  workflowName: 'Untitled Workflow',
  workflowDescription: '',
  isDirty: false,
  history: [],
  historyIndex: -1,
};

export const useWorkflowStore = create<WorkflowStore>((set, get) => ({
  ...initialState,

  // Node actions
  addNode: (type, position, name) => {
    const id = generateId();
    const newNode: WorkflowNode = {
      id,
      type,
      position: {
        x: Math.round(position.x / CANVAS_GRID.size) * CANVAS_GRID.size,
        y: Math.round(position.y / CANVAS_GRID.size) * CANVAS_GRID.size,
      },
      config: getDefaultConfig(type),
      name: name || `${type.charAt(0).toUpperCase() + type.slice(1)} Step`,
      visualState: 'default',
    };

    set((state) => ({
      nodes: [...state.nodes, newNode],
      isDirty: true,
    }));

    get().saveToHistory();
    return id;
  },

  removeNode: (nodeId) => {
    set((state) => ({
      nodes: state.nodes.filter((n) => n.id !== nodeId),
      edges: state.edges.filter((e) => e.source !== nodeId && e.target !== nodeId),
      selectedNodeId: state.selectedNodeId === nodeId ? null : state.selectedNodeId,
      isDirty: true,
    }));
    get().saveToHistory();
  },

  updateNode: (nodeId, updates) => {
    set((state) => ({
      nodes: state.nodes.map((n) => (n.id === nodeId ? { ...n, ...updates } : n)),
      isDirty: true,
    }));
  },

  moveNode: (nodeId, position) => {
    const snappedPosition = {
      x: Math.round(position.x / CANVAS_GRID.size) * CANVAS_GRID.size,
      y: Math.round(position.y / CANVAS_GRID.size) * CANVAS_GRID.size,
    };
    set((state) => ({
      nodes: state.nodes.map((n) =>
        n.id === nodeId ? { ...n, position: snappedPosition } : n
      ),
      isDirty: true,
    }));
  },

  selectNode: (nodeId) => {
    set({ selectedNodeId: nodeId, selectedEdgeId: null });
  },

  duplicateNode: (nodeId) => {
    const state = get();
    const node = state.nodes.find((n) => n.id === nodeId);
    if (!node) return null;

    const newId = generateId();
    const newNode: WorkflowNode = {
      ...node,
      id: newId,
      position: {
        x: node.position.x + NODE_DIMENSIONS.width + CANVAS_GRID.size,
        y: node.position.y,
      },
      name: `${node.name} (copy)`,
    };

    set((state) => ({
      nodes: [...state.nodes, newNode],
      selectedNodeId: newId,
      isDirty: true,
    }));

    get().saveToHistory();
    return newId;
  },

  // Edge actions
  addEdge: (source, target, sourcePort = 'default', label) => {
    const state = get();

    // Prevent duplicate edges
    const existingEdge = state.edges.find(
      (e) => e.source === source && e.target === target && e.sourcePort === sourcePort
    );
    if (existingEdge) return null;

    // Prevent self-loops
    if (source === target) return null;

    // Prevent cycles (simple check)
    const wouldCreateCycle = (src: string, tgt: string): boolean => {
      const visited = new Set<string>();
      const stack = [tgt];
      while (stack.length > 0) {
        const current = stack.pop()!;
        if (current === src) return true;
        if (visited.has(current)) continue;
        visited.add(current);
        state.edges
          .filter((e) => e.source === current)
          .forEach((e) => stack.push(e.target));
      }
      return false;
    };

    if (wouldCreateCycle(source, target)) return null;

    const id = generateEdgeId();
    const newEdge: Edge = {
      id,
      source,
      target,
      sourcePort,
      label,
    };

    set((state) => ({
      edges: [...state.edges, newEdge],
      isDirty: true,
    }));

    get().saveToHistory();
    return id;
  },

  removeEdge: (edgeId) => {
    set((state) => ({
      edges: state.edges.filter((e) => e.id !== edgeId),
      selectedEdgeId: state.selectedEdgeId === edgeId ? null : state.selectedEdgeId,
      isDirty: true,
    }));
    get().saveToHistory();
  },

  selectEdge: (edgeId) => {
    set({ selectedEdgeId: edgeId, selectedNodeId: null });
  },

  // Connection actions
  startConnection: (nodeId, port) => {
    const state = get();
    const node = state.nodes.find((n) => n.id === nodeId);
    if (!node) return;

    set({
      pendingConnection: {
        sourceNodeId: nodeId,
        sourcePort: port,
        mousePosition: {
          x: node.position.x + NODE_DIMENSIONS.width,
          y: node.position.y + NODE_DIMENSIONS.height / 2,
        },
      },
    });
  },

  updateConnectionPosition: (position) => {
    set((state) => ({
      pendingConnection: state.pendingConnection
        ? { ...state.pendingConnection, mousePosition: position }
        : null,
    }));
  },

  completeConnection: (targetNodeId) => {
    const state = get();
    if (!state.pendingConnection) return;

    const { sourceNodeId, sourcePort } = state.pendingConnection;
    const label = sourcePort === 'success' ? 'success' : sourcePort === 'failure' ? 'failure' : undefined;

    get().addEdge(sourceNodeId, targetNodeId, sourcePort, label);
    set({ pendingConnection: null });
  },

  cancelConnection: () => {
    set({ pendingConnection: null });
  },

  // Canvas actions
  setZoom: (zoom) => {
    set({
      zoom: Math.max(ZOOM_CONSTRAINTS.min, Math.min(ZOOM_CONSTRAINTS.max, zoom)),
    });
  },

  zoomIn: () => {
    set((state) => ({
      zoom: Math.min(ZOOM_CONSTRAINTS.max, state.zoom + ZOOM_CONSTRAINTS.step),
    }));
  },

  zoomOut: () => {
    set((state) => ({
      zoom: Math.max(ZOOM_CONSTRAINTS.min, state.zoom - ZOOM_CONSTRAINTS.step),
    }));
  },

  resetZoom: () => {
    set({ zoom: 1, pan: { x: 0, y: 0 } });
  },

  setPan: (pan) => {
    set({ pan });
  },

  fitToView: () => {
    const state = get();
    if (state.nodes.length === 0) {
      set({ zoom: 1, pan: { x: 0, y: 0 } });
      return;
    }

    const bounds = state.nodes.reduce(
      (acc, node) => ({
        minX: Math.min(acc.minX, node.position.x),
        minY: Math.min(acc.minY, node.position.y),
        maxX: Math.max(acc.maxX, node.position.x + NODE_DIMENSIONS.width),
        maxY: Math.max(acc.maxY, node.position.y + NODE_DIMENSIONS.height),
      }),
      { minX: Infinity, minY: Infinity, maxX: -Infinity, maxY: -Infinity }
    );

    const padding = 50;
    const contentWidth = bounds.maxX - bounds.minX + padding * 2;
    const contentHeight = bounds.maxY - bounds.minY + padding * 2;

    // Assume a reasonable canvas size (will be adjusted by the actual canvas component)
    const canvasWidth = 800;
    const canvasHeight = 600;

    const scaleX = canvasWidth / contentWidth;
    const scaleY = canvasHeight / contentHeight;
    const zoom = Math.min(scaleX, scaleY, ZOOM_CONSTRAINTS.max);

    const centerX = (bounds.minX + bounds.maxX) / 2;
    const centerY = (bounds.minY + bounds.maxY) / 2;

    set({
      zoom: Math.max(ZOOM_CONSTRAINTS.min, zoom),
      pan: {
        x: canvasWidth / 2 - centerX * zoom,
        y: canvasHeight / 2 - centerY * zoom,
      },
    });
  },

  setDragging: (isDragging) => {
    set({ isDragging });
  },

  setPanning: (isPanning) => {
    set({ isPanning });
  },

  // Workflow actions
  setWorkflowMeta: (id, name, description) => {
    set({
      workflowId: id,
      workflowName: name,
      workflowDescription: description,
    });
  },

  loadWorkflow: (definition) => {
    set({
      workflowId: definition.id || null,
      workflowName: definition.name,
      workflowDescription: definition.description || '',
      nodes: definition.nodes,
      edges: definition.edges,
      selectedNodeId: null,
      selectedEdgeId: null,
      isDirty: false,
      history: [],
      historyIndex: -1,
    });
  },

  clearCanvas: () => {
    set({
      ...initialState,
      history: [],
      historyIndex: -1,
    });
  },

  getWorkflowDefinition: () => {
    const state = get();
    return {
      id: state.workflowId || undefined,
      name: state.workflowName,
      description: state.workflowDescription || undefined,
      nodes: state.nodes,
      edges: state.edges,
    };
  },

  // History actions
  undo: () => {
    const state = get();
    if (state.historyIndex <= 0) return;

    const newIndex = state.historyIndex - 1;
    const historyState = state.history[newIndex];

    set({
      nodes: historyState.nodes,
      edges: historyState.edges,
      historyIndex: newIndex,
      isDirty: true,
    });
  },

  redo: () => {
    const state = get();
    if (state.historyIndex >= state.history.length - 1) return;

    const newIndex = state.historyIndex + 1;
    const historyState = state.history[newIndex];

    set({
      nodes: historyState.nodes,
      edges: historyState.edges,
      historyIndex: newIndex,
      isDirty: true,
    });
  },

  saveToHistory: () => {
    const state = get();
    const newHistory = state.history.slice(0, state.historyIndex + 1);
    newHistory.push({
      nodes: JSON.parse(JSON.stringify(state.nodes)),
      edges: JSON.parse(JSON.stringify(state.edges)),
    });

    // Limit history size
    if (newHistory.length > 50) {
      newHistory.shift();
    }

    set({
      history: newHistory,
      historyIndex: newHistory.length - 1,
    });
  },

  // Selection actions
  clearSelection: () => {
    set({ selectedNodeId: null, selectedEdgeId: null });
  },

  deleteSelected: () => {
    const state = get();
    if (state.selectedNodeId) {
      get().removeNode(state.selectedNodeId);
    } else if (state.selectedEdgeId) {
      get().removeEdge(state.selectedEdgeId);
    }
  },

  selectAll: () => {
    // For now, just select the first node if any exist
    const state = get();
    if (state.nodes.length > 0) {
      set({ selectedNodeId: state.nodes[0].id });
    }
  },
}));
