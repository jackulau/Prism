import { create } from 'zustand';
import { wsService } from '../services/websocket';
import { useAppStore } from './index';
import { useSandboxStore } from './sandboxStore';
import { toast } from './toastStore';

export type OperationType = 'chat' | 'agent' | 'workflow' | 'build' | 'swarm';

export interface ActiveOperation {
  id: string;
  type: OperationType;
  startedAt: Date;
  description?: string;
}

interface EmergencyStopState {
  activeOperations: ActiveOperation[];
  isEmergencyStopActive: boolean;

  // Actions
  registerOperation: (op: Omit<ActiveOperation, 'startedAt'>) => void;
  unregisterOperation: (id: string) => void;
  emergencyStopAll: () => Promise<void>;
  hasActiveOperations: () => boolean;
  getOperationCount: () => number;
}

export const useEmergencyStopStore = create<EmergencyStopState>((set, get) => ({
  activeOperations: [],
  isEmergencyStopActive: false,

  registerOperation: (op) => {
    set((state) => ({
      activeOperations: [
        ...state.activeOperations.filter((o) => o.id !== op.id),
        { ...op, startedAt: new Date() },
      ],
    }));
  },

  unregisterOperation: (id) => {
    set((state) => ({
      activeOperations: state.activeOperations.filter((o) => o.id !== id),
    }));
  },

  emergencyStopAll: async () => {
    const operations = get().activeOperations;
    if (operations.length === 0) return;

    set({ isEmergencyStopActive: true });

    try {
      // Stop all operations in parallel
      await Promise.all(
        operations.map(async (op) => {
          try {
            switch (op.type) {
              case 'chat':
                wsService.stopGeneration(op.id);
                break;
              case 'build':
                wsService.stopBuild(op.id);
                break;
              case 'agent':
                wsService.send({ type: 'agent.stop', execution_id: op.id });
                break;
              case 'workflow':
                wsService.send({ type: 'workflow.stop', workflow_id: op.id });
                break;
              case 'swarm':
                wsService.send({ type: 'swarm.stop', swarm_id: op.id });
                break;
            }
          } catch {
            // Continue stopping other operations even if one fails
          }
        })
      );

      // Clear UI state
      useAppStore.getState().endGeneration();
      useSandboxStore.getState().setIsRunning(false);

      const count = operations.length;
      toast.warning(`Stopped ${count} operation${count > 1 ? 's' : ''}`);

      set({ activeOperations: [] });
    } finally {
      set({ isEmergencyStopActive: false });
    }
  },

  hasActiveOperations: () => get().activeOperations.length > 0,

  getOperationCount: () => get().activeOperations.length,
}));
