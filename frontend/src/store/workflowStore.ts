import { create } from 'zustand';
import type {
  WorkflowNode,
  WorkflowEdge,
  StepConfig,
  StepType,
  StateVariable,
  Workflow,
} from '../types/workflow';

interface WorkflowState {
  // Current workflow being edited
  workflow: Workflow | null;
  setWorkflow: (workflow: Workflow | null) => void;

  // Nodes and edges for the canvas
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  setNodes: (nodes: WorkflowNode[]) => void;
  setEdges: (edges: WorkflowEdge[]) => void;

  // Selected node
  selectedNodeId: string | null;
  setSelectedNodeId: (id: string | null) => void;
  getSelectedNode: () => WorkflowNode | null;

  // Node operations
  addNode: (type: StepType, position: { x: number; y: number }) => WorkflowNode;
  updateNode: (id: string, updates: Partial<WorkflowNode>) => void;
  updateNodeConfig: (id: string, config: StepConfig) => void;
  updateNodeData: (id: string, data: Partial<WorkflowNode['data']>) => void;
  deleteNode: (id: string) => void;

  // Edge operations
  addEdge: (edge: Omit<WorkflowEdge, 'id'>) => void;
  deleteEdge: (id: string) => void;

  // State variables from previous steps
  getAvailableStateVariables: (beforeNodeId?: string) => StateVariable[];

  // Dirty state tracking
  isDirty: boolean;
  setIsDirty: (dirty: boolean) => void;

  // Config panel state
  isConfigPanelOpen: boolean;
  openConfigPanel: (nodeId: string) => void;
  closeConfigPanel: () => void;

  // Validation
  validationErrors: Record<string, string[]>;
  setValidationErrors: (nodeId: string, errors: string[]) => void;
  clearValidationErrors: (nodeId?: string) => void;
}

// Generate a unique ID for nodes
const generateNodeId = () => `step_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

// Default configs for each step type
const getDefaultConfig = (type: StepType): StepConfig => {
  switch (type) {
    case 'agent':
      return {
        agentConfig: {
          provider: 'anthropic',
          model: 'claude-3-sonnet-20240229',
          prompt: '',
          temperature: 0.7,
          maxTokens: 4096,
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
          timeout: 300000, // 5 minutes default
        },
      };
    case 'transform':
      return {
        transformConfig: {
          type: 'template',
        },
      };
    default:
      return {};
  }
};

// Get step type display name
export const getStepTypeLabel = (type: StepType): string => {
  switch (type) {
    case 'agent':
      return 'Agent';
    case 'tool':
      return 'Tool';
    case 'condition':
      return 'Condition';
    case 'parallel':
      return 'Parallel';
    case 'wait':
      return 'Wait';
    case 'transform':
      return 'Transform';
    default:
      return type;
  }
};

export const useWorkflowStore = create<WorkflowState>((set, get) => ({
  // Workflow
  workflow: null,
  setWorkflow: (workflow) => set({ workflow }),

  // Nodes and edges
  nodes: [],
  edges: [],
  setNodes: (nodes) => set({ nodes, isDirty: true }),
  setEdges: (edges) => set({ edges, isDirty: true }),

  // Selected node
  selectedNodeId: null,
  setSelectedNodeId: (id) => set({ selectedNodeId: id }),
  getSelectedNode: () => {
    const { nodes, selectedNodeId } = get();
    return nodes.find((n) => n.id === selectedNodeId) || null;
  },

  // Node operations
  addNode: (type, position) => {
    const id = generateNodeId();
    const newNode: WorkflowNode = {
      id,
      type,
      position,
      data: {
        name: `New ${getStepTypeLabel(type)} Step`,
        config: getDefaultConfig(type),
      },
    };
    set((state) => ({
      nodes: [...state.nodes, newNode],
      isDirty: true,
    }));
    return newNode;
  },

  updateNode: (id, updates) => {
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === id ? { ...node, ...updates } : node
      ),
      isDirty: true,
    }));
  },

  updateNodeConfig: (id, config) => {
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === id
          ? { ...node, data: { ...node.data, config } }
          : node
      ),
      isDirty: true,
    }));
  },

  updateNodeData: (id, data) => {
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === id
          ? { ...node, data: { ...node.data, ...data } }
          : node
      ),
      isDirty: true,
    }));
  },

  deleteNode: (id) => {
    set((state) => ({
      nodes: state.nodes.filter((node) => node.id !== id),
      edges: state.edges.filter(
        (edge) => edge.source !== id && edge.target !== id
      ),
      selectedNodeId: state.selectedNodeId === id ? null : state.selectedNodeId,
      isConfigPanelOpen: state.selectedNodeId === id ? false : state.isConfigPanelOpen,
      isDirty: true,
    }));
    // Clear validation errors for deleted node
    get().clearValidationErrors(id);
  },

  // Edge operations
  addEdge: (edge) => {
    const id = `edge_${edge.source}_${edge.target}`;
    set((state) => ({
      edges: [...state.edges, { ...edge, id }],
      isDirty: true,
    }));
  },

  deleteEdge: (id) => {
    set((state) => ({
      edges: state.edges.filter((edge) => edge.id !== id),
      isDirty: true,
    }));
  },

  // Get state variables from steps before the given node
  getAvailableStateVariables: (beforeNodeId) => {
    const { nodes, edges } = get();
    const variables: StateVariable[] = [];

    // Find all nodes that come before the given node
    const getNodesBeforeId = (targetId: string): string[] => {
      const result: string[] = [];
      const visited = new Set<string>();

      const traverse = (nodeId: string) => {
        if (visited.has(nodeId)) return;
        visited.add(nodeId);

        const incomingEdges = edges.filter((e) => e.target === nodeId);
        for (const edge of incomingEdges) {
          if (edge.source !== targetId) {
            result.push(edge.source);
            traverse(edge.source);
          }
        }
      };

      if (targetId) {
        traverse(targetId);
      }
      return result;
    };

    const precedingNodeIds = beforeNodeId
      ? getNodesBeforeId(beforeNodeId)
      : nodes.map((n) => n.id);

    // Extract output keys from preceding nodes
    for (const nodeId of precedingNodeIds) {
      const node = nodes.find((n) => n.id === nodeId);
      if (!node) continue;

      const { config } = node.data;
      let outputKey: string | undefined;

      if (config.agentConfig?.outputKey) {
        outputKey = config.agentConfig.outputKey;
      } else if (config.toolConfig?.outputKey) {
        outputKey = config.toolConfig.outputKey;
      } else if (config.waitConfig?.outputKey) {
        outputKey = config.waitConfig.outputKey;
      } else if (config.transformConfig?.outputKey) {
        outputKey = config.transformConfig.outputKey;
      }

      if (outputKey) {
        variables.push({
          key: outputKey,
          type: 'unknown',
          source: node.data.name,
        });
      }
    }

    return variables;
  },

  // Dirty state
  isDirty: false,
  setIsDirty: (dirty) => set({ isDirty: dirty }),

  // Config panel
  isConfigPanelOpen: false,
  openConfigPanel: (nodeId) => set({ selectedNodeId: nodeId, isConfigPanelOpen: true }),
  closeConfigPanel: () => set({ isConfigPanelOpen: false }),

  // Validation
  validationErrors: {},
  setValidationErrors: (nodeId, errors) =>
    set((state) => ({
      validationErrors: { ...state.validationErrors, [nodeId]: errors },
    })),
  clearValidationErrors: (nodeId) => {
    if (nodeId) {
      set((state) => {
        const { [nodeId]: _, ...rest } = state.validationErrors;
        return { validationErrors: rest };
      });
    } else {
      set({ validationErrors: {} });
    }
  },
}));
