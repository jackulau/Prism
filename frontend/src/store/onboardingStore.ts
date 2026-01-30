import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type OnboardingStep =
  | 'welcome'
  | 'provider-setup'
  | 'first-workspace'
  | 'feature-tour'
  | 'completion';

export const ONBOARDING_STEPS: OnboardingStep[] = [
  'welcome',
  'provider-setup',
  'first-workspace',
  'feature-tour',
  'completion',
];

export interface OnboardingState {
  hasCompletedOnboarding: boolean;
  currentStep: OnboardingStep;
  skippedSteps: OnboardingStep[];
  completedAt: string | null;

  // Actions
  setCurrentStep: (step: OnboardingStep) => void;
  skipStep: (step: OnboardingStep) => void;
  nextStep: () => void;
  previousStep: () => void;
  completeOnboarding: () => void;
  resetOnboarding: () => void;
  canSkipStep: (step: OnboardingStep) => boolean;
  getStepIndex: (step: OnboardingStep) => number;
  isFirstStep: () => boolean;
  isLastStep: () => boolean;
}

// Steps that can be skipped
const SKIPPABLE_STEPS: OnboardingStep[] = ['provider-setup', 'first-workspace', 'feature-tour'];

export const useOnboardingStore = create<OnboardingState>()(
  persist(
    (set, get) => ({
      hasCompletedOnboarding: false,
      currentStep: 'welcome',
      skippedSteps: [],
      completedAt: null,

      setCurrentStep: (step) => set({ currentStep: step }),

      skipStep: (step) => {
        const state = get();
        if (!state.skippedSteps.includes(step)) {
          set({ skippedSteps: [...state.skippedSteps, step] });
        }
        state.nextStep();
      },

      nextStep: () => {
        const state = get();
        const currentIndex = ONBOARDING_STEPS.indexOf(state.currentStep);
        if (currentIndex < ONBOARDING_STEPS.length - 1) {
          set({ currentStep: ONBOARDING_STEPS[currentIndex + 1] });
        }
      },

      previousStep: () => {
        const state = get();
        const currentIndex = ONBOARDING_STEPS.indexOf(state.currentStep);
        if (currentIndex > 0) {
          set({ currentStep: ONBOARDING_STEPS[currentIndex - 1] });
        }
      },

      completeOnboarding: () => {
        set({
          hasCompletedOnboarding: true,
          completedAt: new Date().toISOString(),
        });
      },

      resetOnboarding: () => {
        set({
          hasCompletedOnboarding: false,
          currentStep: 'welcome',
          skippedSteps: [],
          completedAt: null,
        });
      },

      canSkipStep: (step) => SKIPPABLE_STEPS.includes(step),

      getStepIndex: (step) => ONBOARDING_STEPS.indexOf(step),

      isFirstStep: () => {
        const state = get();
        return ONBOARDING_STEPS.indexOf(state.currentStep) === 0;
      },

      isLastStep: () => {
        const state = get();
        return ONBOARDING_STEPS.indexOf(state.currentStep) === ONBOARDING_STEPS.length - 1;
      },
    }),
    {
      name: 'prism-onboarding',
      partialize: (state) => ({
        hasCompletedOnboarding: state.hasCompletedOnboarding,
        currentStep: state.currentStep,
        skippedSteps: state.skippedSteps,
        completedAt: state.completedAt,
      }),
    }
  )
);
