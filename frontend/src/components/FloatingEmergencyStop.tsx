import { useEffect } from 'react';
import { EmergencyStopButton } from './EmergencyStopButton';
import { useEmergencyStopStore } from '../store/emergencyStopStore';

export function FloatingEmergencyStop() {
  const { hasActiveOperations, emergencyStopAll } = useEmergencyStopStore();
  const hasActive = hasActiveOperations();

  // Add keyboard shortcut: Shift+Escape
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && e.shiftKey) {
        const store = useEmergencyStopStore.getState();
        if (store.hasActiveOperations()) {
          e.preventDefault();
          store.emergencyStopAll();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [emergencyStopAll]);

  if (!hasActive) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 animate-fade-in">
      <EmergencyStopButton requireConfirmation />
    </div>
  );
}
