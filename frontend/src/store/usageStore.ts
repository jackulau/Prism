import { create } from 'zustand';

interface UsageMetrics {
  used: number;
  limit: number;
}

interface BillingPeriod {
  start: Date;
  end: Date;
}

interface UsageState {
  tokenUsage: UsageMetrics;
  sandboxUsage: UsageMetrics;
  billingPeriod: BillingPeriod | null;
  isLoading: boolean;
  error: string | null;
  lastUpdated: Date | null;

  setTokenUsage: (usage: UsageMetrics) => void;
  setSandboxUsage: (usage: UsageMetrics) => void;
  setBillingPeriod: (period: BillingPeriod | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  updateUsage: (data: {
    tokenUsage?: UsageMetrics;
    sandboxUsage?: UsageMetrics;
    billingPeriod?: BillingPeriod;
  }) => void;
  reset: () => void;
}

const initialState = {
  tokenUsage: { used: 0, limit: 500000 }, // 500K tokens default
  sandboxUsage: { used: 0, limit: 2 }, // 2 hours default
  billingPeriod: null,
  isLoading: false,
  error: null,
  lastUpdated: null,
};

export const useUsageStore = create<UsageState>((set) => ({
  ...initialState,

  setTokenUsage: (usage) => set({ tokenUsage: usage, lastUpdated: new Date() }),

  setSandboxUsage: (usage) => set({ sandboxUsage: usage, lastUpdated: new Date() }),

  setBillingPeriod: (period) => set({ billingPeriod: period }),

  setLoading: (isLoading) => set({ isLoading }),

  setError: (error) => set({ error }),

  updateUsage: (data) =>
    set((state) => ({
      ...state,
      ...(data.tokenUsage && { tokenUsage: data.tokenUsage }),
      ...(data.sandboxUsage && { sandboxUsage: data.sandboxUsage }),
      ...(data.billingPeriod && { billingPeriod: data.billingPeriod }),
      lastUpdated: new Date(),
    })),

  reset: () => set(initialState),
}));

// Selectors for computed values
export const selectTokenPercent = (state: UsageState) => {
  const { used, limit } = state.tokenUsage;
  return limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
};

export const selectSandboxPercent = (state: UsageState) => {
  const { used, limit } = state.sandboxUsage;
  return limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
};

export const selectIsOverTokenLimit = (state: UsageState) => {
  const { used, limit } = state.tokenUsage;
  return used >= limit;
};

export const selectIsOverSandboxLimit = (state: UsageState) => {
  const { used, limit } = state.sandboxUsage;
  return used >= limit;
};
