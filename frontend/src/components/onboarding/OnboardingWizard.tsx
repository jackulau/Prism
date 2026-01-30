import { useCallback } from 'react';
import { useOnboardingStore, type OnboardingStep } from '../../store/onboardingStore';
import { OnboardingProgress } from './OnboardingProgress';
import { WelcomeStep } from './WelcomeStep';
import { ProviderSetupStep } from './ProviderSetupStep';
import { FirstWorkspaceStep } from './FirstWorkspaceStep';
import { FeatureTourStep } from './FeatureTourStep';
import { CompletionStep } from './CompletionStep';

export function OnboardingWizard() {
  const {
    currentStep,
    skippedSteps,
    nextStep,
    skipStep,
    completeOnboarding,
  } = useOnboardingStore();

  const handleNext = useCallback(() => {
    nextStep();
  }, [nextStep]);

  const handleSkip = useCallback(
    (step: OnboardingStep) => {
      skipStep(step);
    },
    [skipStep]
  );

  const handleComplete = useCallback(() => {
    completeOnboarding();
  }, [completeOnboarding]);

  const renderStep = () => {
    switch (currentStep) {
      case 'welcome':
        return <WelcomeStep onNext={handleNext} />;
      case 'provider-setup':
        return (
          <ProviderSetupStep
            onNext={handleNext}
            onSkip={() => handleSkip('provider-setup')}
          />
        );
      case 'first-workspace':
        return (
          <FirstWorkspaceStep
            onNext={handleNext}
            onSkip={() => handleSkip('first-workspace')}
          />
        );
      case 'feature-tour':
        return (
          <FeatureTourStep
            onNext={handleNext}
            onSkip={() => handleSkip('feature-tour')}
          />
        );
      case 'completion':
        return <CompletionStep onComplete={handleComplete} />;
      default:
        return <WelcomeStep onNext={handleNext} />;
    }
  };

  return (
    <div className="min-h-screen bg-editor-bg flex flex-col">
      {/* Progress indicator */}
      <OnboardingProgress currentStep={currentStep} skippedSteps={skippedSteps} />

      {/* Step content */}
      <div className="flex-1 flex items-center justify-center overflow-y-auto">
        <div className="w-full max-w-3xl px-4 pb-12">{renderStep()}</div>
      </div>

      {/* Footer */}
      <footer className="py-4 text-center text-xs text-editor-muted border-t border-editor-border">
        <p>
          Need help?{' '}
          <a
            href="https://github.com/anthropics/claude-code/issues"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            Visit our GitHub
          </a>
        </p>
      </footer>
    </div>
  );
}
