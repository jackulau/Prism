import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { OnboardingWizard } from '../components/onboarding/OnboardingWizard';
import { useOnboardingStore } from '../store/onboardingStore';
import { useAuthStore } from '../store/authStore';

export default function Onboarding() {
  const navigate = useNavigate();
  const { hasCompletedOnboarding } = useOnboardingStore();
  const { isAuthenticated } = useAuthStore();

  // Redirect if already completed onboarding or not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    if (hasCompletedOnboarding) {
      navigate('/');
    }
  }, [hasCompletedOnboarding, isAuthenticated, navigate]);

  // Show nothing while redirecting
  if (!isAuthenticated || hasCompletedOnboarding) {
    return null;
  }

  return <OnboardingWizard />;
}
