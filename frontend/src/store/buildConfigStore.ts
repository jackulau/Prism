import { create } from 'zustand';
import {
  buildConfigService,
  type BuildConfig,
  type BuildCommand,
  type BuildEnvVar,
  type CreateBuildConfigInput,
  type UpdateBuildConfigInput,
  type CreateCommandInput,
  type UpdateCommandInput,
  type SetEnvVarInput,
} from '../services/buildConfig';

interface BuildConfigStore {
  // State
  configs: BuildConfig[];
  selectedConfig: BuildConfig | null;
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;

  // Config actions
  fetchConfigs: (workspaceId?: string) => Promise<void>;
  createConfig: (data: CreateBuildConfigInput) => Promise<BuildConfig | null>;
  updateConfig: (id: string, data: UpdateBuildConfigInput) => Promise<void>;
  deleteConfig: (id: string) => Promise<void>;
  setDefault: (id: string) => Promise<void>;
  selectConfig: (config: BuildConfig | null) => void;
  refreshConfig: (id: string) => Promise<void>;

  // Command actions
  addCommand: (configId: string, data: CreateCommandInput) => Promise<void>;
  updateCommand: (configId: string, cmdId: string, data: UpdateCommandInput) => Promise<void>;
  deleteCommand: (configId: string, cmdId: string) => Promise<void>;
  reorderCommands: (configId: string, order: string[]) => Promise<void>;

  // Env var actions
  setEnvVar: (configId: string, data: SetEnvVarInput) => Promise<void>;
  deleteEnvVar: (configId: string, key: string) => Promise<void>;

  // Utility
  setToken: (token: string | null) => void;
  clearError: () => void;
}

export const useBuildConfigStore = create<BuildConfigStore>((set, get) => ({
  // Initial state
  configs: [],
  selectedConfig: null,
  isLoading: false,
  isSaving: false,
  error: null,

  setToken: (token: string | null) => {
    buildConfigService.setToken(token);
  },

  clearError: () => set({ error: null }),

  // Config actions
  fetchConfigs: async (workspaceId?: string) => {
    set({ isLoading: true, error: null });
    const response = await buildConfigService.list(workspaceId);
    if (response.error) {
      set({ isLoading: false, error: response.error });
    } else {
      set({ isLoading: false, configs: response.data?.configs || [] });
    }
  },

  createConfig: async (data: CreateBuildConfigInput) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.create(data);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return null;
    }
    const newConfig = response.data!;
    set((state) => ({
      isSaving: false,
      configs: [...state.configs, newConfig],
      selectedConfig: newConfig,
    }));
    return newConfig;
  },

  updateConfig: async (id: string, data: UpdateBuildConfigInput) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.update(id, data);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    const updatedConfig = response.data!;
    set((state) => ({
      isSaving: false,
      configs: state.configs.map((c) => (c.id === id ? updatedConfig : c)),
      selectedConfig: state.selectedConfig?.id === id ? updatedConfig : state.selectedConfig,
    }));
  },

  deleteConfig: async (id: string) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.delete(id);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    set((state) => ({
      isSaving: false,
      configs: state.configs.filter((c) => c.id !== id),
      selectedConfig: state.selectedConfig?.id === id ? null : state.selectedConfig,
    }));
  },

  setDefault: async (id: string) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.setDefault(id);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Update configs to reflect new default
    set((state) => ({
      isSaving: false,
      configs: state.configs.map((c) => ({
        ...c,
        isDefault: c.id === id,
      })),
      selectedConfig: state.selectedConfig
        ? { ...state.selectedConfig, isDefault: state.selectedConfig.id === id }
        : null,
    }));
  },

  selectConfig: (config: BuildConfig | null) => {
    set({ selectedConfig: config, error: null });
  },

  refreshConfig: async (id: string) => {
    const response = await buildConfigService.get(id);
    if (response.error) {
      set({ error: response.error });
      return;
    }
    const refreshedConfig = response.data!;
    set((state) => ({
      configs: state.configs.map((c) => (c.id === id ? refreshedConfig : c)),
      selectedConfig: state.selectedConfig?.id === id ? refreshedConfig : state.selectedConfig,
    }));
  },

  // Command actions
  addCommand: async (configId: string, data: CreateCommandInput) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.addCommand(configId, data);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Refresh the config to get updated commands
    await get().refreshConfig(configId);
    set({ isSaving: false });
  },

  updateCommand: async (configId: string, cmdId: string, data: UpdateCommandInput) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.updateCommand(configId, cmdId, data);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Update command locally
    set((state) => {
      const updateCommands = (commands: BuildCommand[]) =>
        commands.map((cmd) => (cmd.id === cmdId ? { ...cmd, ...data } : cmd));

      return {
        isSaving: false,
        configs: state.configs.map((c) =>
          c.id === configId ? { ...c, commands: updateCommands(c.commands) } : c
        ),
        selectedConfig:
          state.selectedConfig?.id === configId
            ? { ...state.selectedConfig, commands: updateCommands(state.selectedConfig.commands) }
            : state.selectedConfig,
      };
    });
  },

  deleteCommand: async (configId: string, cmdId: string) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.deleteCommand(configId, cmdId);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Remove command locally
    set((state) => {
      const filterCommands = (commands: BuildCommand[]) => commands.filter((cmd) => cmd.id !== cmdId);

      return {
        isSaving: false,
        configs: state.configs.map((c) =>
          c.id === configId ? { ...c, commands: filterCommands(c.commands) } : c
        ),
        selectedConfig:
          state.selectedConfig?.id === configId
            ? { ...state.selectedConfig, commands: filterCommands(state.selectedConfig.commands) }
            : state.selectedConfig,
      };
    });
  },

  reorderCommands: async (configId: string, order: string[]) => {
    // Optimistically update the order
    set((state) => {
      const reorderCommands = (commands: BuildCommand[]) => {
        const commandMap = new Map(commands.map((cmd) => [cmd.id, cmd]));
        return order
          .map((id, index) => {
            const cmd = commandMap.get(id);
            return cmd ? { ...cmd, runOrder: index } : null;
          })
          .filter((cmd): cmd is BuildCommand => cmd !== null);
      };

      return {
        configs: state.configs.map((c) =>
          c.id === configId ? { ...c, commands: reorderCommands(c.commands) } : c
        ),
        selectedConfig:
          state.selectedConfig?.id === configId
            ? { ...state.selectedConfig, commands: reorderCommands(state.selectedConfig.commands) }
            : state.selectedConfig,
      };
    });

    const response = await buildConfigService.reorderCommands(configId, order);
    if (response.error) {
      // Revert on error by refreshing
      await get().refreshConfig(configId);
      set({ error: response.error });
    }
  },

  // Env var actions
  setEnvVar: async (configId: string, data: SetEnvVarInput) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.setEnvVar(configId, data);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Refresh to get updated env vars (with proper masking)
    await get().refreshConfig(configId);
    set({ isSaving: false });
  },

  deleteEnvVar: async (configId: string, key: string) => {
    set({ isSaving: true, error: null });
    const response = await buildConfigService.deleteEnvVar(configId, key);
    if (response.error) {
      set({ isSaving: false, error: response.error });
      return;
    }
    // Remove env var locally
    set((state) => {
      const filterEnvVars = (envVars: BuildEnvVar[]) => envVars.filter((v) => v.key !== key);

      return {
        isSaving: false,
        configs: state.configs.map((c) =>
          c.id === configId ? { ...c, envVars: filterEnvVars(c.envVars) } : c
        ),
        selectedConfig:
          state.selectedConfig?.id === configId
            ? { ...state.selectedConfig, envVars: filterEnvVars(state.selectedConfig.envVars) }
            : state.selectedConfig,
      };
    });
  },
}));
