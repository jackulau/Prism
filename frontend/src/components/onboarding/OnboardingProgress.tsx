import { Check } from 'lucide-react';
import { ONBOARDING_STEPS, type OnboardingStep } from '../../store/onboardingStore';

interface OnboardingProgressProps {
  currentStep: OnboardingStep;
  skippedSteps: OnboardingStep[];
}

const STEP_LABELS: Record<OnboardingStep, string> = {
  'welcome': 'Welcome',
  'provider-setup': 'Providers',
  'first-workspace': 'Workspace',
  'feature-tour': 'Tour',
  'completion': 'Done',
};

export function OnboardingProgress({ currentStep, skippedSteps }: OnboardingProgressProps) {
  const currentIndex = ONBOARDING_STEPS.indexOf(currentStep);

  return (
    <div className="w-full px-4 py-6">
      <div className="flex items-center justify-between max-w-md mx-auto">
        {ONBOARDING_STEPS.map((step, index) => {
          const isCompleted = index < currentIndex;
          const isCurrent = step === currentStep;
          const isSkipped = skippedSteps.includes(step);

          return (
            <div key={step} className="flex items-center">
              {/* Step indicator */}
              <div className="flex flex-col items-center">
                <div
                  className={`
                    w-10 h-10 rounded-full flex items-center justify-center
                    transition-all duration-300
                    ${isCompleted
                      ? 'bg-green-500 text-white'
                      : isCurrent
                        ? 'bg-primary text-white ring-4 ring-primary/20'
                        : 'bg-editor-surface border-2 border-editor-border text-editor-muted'
                    }
                    ${isSkipped ? 'opacity-60' : ''}
                  `}
                >
                  {isCompleted ? (
                    <Check className="w-5 h-5" />
                  ) : (
                    <span className="text-sm font-medium">{index + 1}</span>
                  )}
                </div>
                <span
                  className={`
                    mt-2 text-xs font-medium
                    ${isCurrent ? 'text-primary' : isCompleted ? 'text-editor-text' : 'text-editor-muted'}
                  `}
                >
                  {STEP_LABELS[step]}
                </span>
              </div>

              {/* Connector line */}
              {index < ONBOARDING_STEPS.length - 1 && (
                <div
                  className={`
                    w-8 sm:w-12 h-0.5 mx-1 sm:mx-2 mb-6
                    transition-colors duration-300
                    ${index < currentIndex ? 'bg-green-500' : 'bg-editor-border'}
                  `}
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
