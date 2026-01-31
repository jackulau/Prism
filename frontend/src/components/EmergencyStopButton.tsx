import { useState } from 'react';
import { StopCircle, Loader2 } from 'lucide-react';
import { useEmergencyStopStore } from '../store/emergencyStopStore';
import { ConfirmDialog } from './ConfirmDialog';

interface EmergencyStopButtonProps {
  className?: string;
  showCount?: boolean;
  requireConfirmation?: boolean;
}

export function EmergencyStopButton({
  className,
  showCount = true,
  requireConfirmation = false,
}: EmergencyStopButtonProps) {
  const [showConfirm, setShowConfirm] = useState(false);
  const { activeOperations, isEmergencyStopActive, emergencyStopAll } =
    useEmergencyStopStore();

  const operationCount = activeOperations.length;

  if (operationCount === 0) {
    return null;
  }

  const handleClick = () => {
    if (requireConfirmation) {
      setShowConfirm(true);
    } else {
      emergencyStopAll();
    }
  };

  const handleConfirm = () => {
    setShowConfirm(false);
    emergencyStopAll();
  };

  return (
    <>
      <button
        onClick={handleClick}
        disabled={isEmergencyStopActive}
        className={`
          flex items-center gap-2 px-4 py-2 rounded-lg
          bg-red-500 text-white font-medium
          hover:bg-red-600
          active:scale-95
          transition-all duration-150
          disabled:opacity-50 disabled:cursor-not-allowed
          shadow-lg shadow-red-500/25
          ${className || ''}
        `}
        title="Stop all active operations (Shift+Escape)"
      >
        {isEmergencyStopActive ? (
          <Loader2 className="w-5 h-5 animate-spin" />
        ) : (
          <StopCircle className="w-5 h-5" />
        )}
        <span>Stop All</span>
        {showCount && operationCount > 1 && (
          <span className="ml-1 px-1.5 py-0.5 text-xs rounded-full bg-white/20">
            {operationCount}
          </span>
        )}
      </button>

      {requireConfirmation && (
        <ConfirmDialog
          isOpen={showConfirm}
          title="Stop All Operations?"
          message={`This will immediately stop ${operationCount} active operation${operationCount > 1 ? 's' : ''}. This action cannot be undone.`}
          confirmText="Stop All"
          variant="danger"
          onConfirm={handleConfirm}
          onCancel={() => setShowConfirm(false)}
        />
      )}
    </>
  );
}
