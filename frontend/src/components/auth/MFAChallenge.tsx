import { useState, useRef, useEffect } from 'react';
import { Shield, Loader2, AlertTriangle, ArrowLeft } from 'lucide-react';
import { useMFAStore } from '../../store/mfaStore';

interface MFAChallengeProps {
  sessionToken: string;
  onSuccess: (authData: {
    access_token: string;
    refresh_token: string;
    user: { id: string; email: string; created_at: string };
  }) => void;
  onBack?: () => void;
}

export function MFAChallenge({ sessionToken, onSuccess, onBack }: MFAChallengeProps) {
  const { isLoading, error, validateLogin, validateBackupCode, clearError } = useMFAStore();

  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [useBackupCode, setUseBackupCode] = useState(false);
  const [backupCode, setBackupCode] = useState('');
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    clearError();
    if (!useBackupCode) {
      inputRefs.current[0]?.focus();
    }
  }, [useBackupCode, clearError]);

  const handleCodeChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;

    const newCode = [...code];
    newCode[index] = value.slice(-1);
    setCode(newCode);

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }

    // Auto-submit when all digits are entered
    const fullCode = [...newCode.slice(0, index), value.slice(-1), ...newCode.slice(index + 1)].join('');
    if (fullCode.length === 6 && !fullCode.includes('')) {
      handleSubmitCode(fullCode);
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const paste = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const newCode = ['', '', '', '', '', ''];
    for (let i = 0; i < paste.length; i++) {
      newCode[i] = paste[i];
    }
    setCode(newCode);
    if (paste.length === 6) {
      handleSubmitCode(paste);
    } else if (paste.length > 0) {
      inputRefs.current[Math.min(paste.length, 5)]?.focus();
    }
  };

  const handleSubmitCode = async (fullCode?: string) => {
    const codeToSubmit = fullCode || code.join('');
    if (codeToSubmit.length !== 6) return;

    clearError();
    const result = await validateLogin(sessionToken, codeToSubmit);
    if (result) {
      onSuccess(result);
    } else {
      // Clear the code on error
      setCode(['', '', '', '', '', '']);
      inputRefs.current[0]?.focus();
    }
  };

  const handleSubmitBackupCode = async () => {
    if (!backupCode.trim()) return;

    clearError();
    const result = await validateBackupCode(sessionToken, backupCode.trim());
    if (result) {
      onSuccess(result);
    } else {
      setBackupCode('');
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (useBackupCode) {
      handleSubmitBackupCode();
    } else {
      handleSubmitCode();
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="text-center mb-6">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <Shield className="w-8 h-8 text-primary" />
        </div>
        <h2 className="text-2xl font-bold">Two-Factor Authentication</h2>
        <p className="text-editor-muted mt-2">
          {useBackupCode
            ? 'Enter one of your backup codes'
            : 'Enter the 6-digit code from your authenticator app'}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
            <AlertTriangle className="w-4 h-4 flex-shrink-0" />
            {error}
          </div>
        )}

        {!useBackupCode ? (
          <div className="flex justify-center gap-2" onPaste={handlePaste}>
            {code.map((digit, index) => (
              <input
                key={index}
                ref={(el) => (inputRefs.current[index] = el)}
                type="text"
                inputMode="numeric"
                maxLength={1}
                value={digit}
                onChange={(e) => handleCodeChange(index, e.target.value)}
                onKeyDown={(e) => handleKeyDown(index, e)}
                disabled={isLoading}
                className="w-12 h-14 text-center text-xl font-mono bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                autoFocus={index === 0}
              />
            ))}
          </div>
        ) : (
          <div>
            <input
              type="text"
              value={backupCode}
              onChange={(e) => setBackupCode(e.target.value)}
              placeholder="Enter backup code"
              disabled={isLoading}
              className="w-full px-4 py-3 bg-editor-surface border border-editor-border rounded-lg text-center font-mono text-lg focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              autoFocus
            />
          </div>
        )}

        <button
          type="submit"
          disabled={isLoading || (useBackupCode ? !backupCode.trim() : code.join('').length !== 6)}
          className="w-full py-3 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
        >
          {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
          Verify
        </button>
      </form>

      <div className="mt-6 text-center space-y-2">
        <button
          type="button"
          onClick={() => {
            setUseBackupCode(!useBackupCode);
            setCode(['', '', '', '', '', '']);
            setBackupCode('');
            clearError();
          }}
          className="text-sm text-editor-accent hover:underline"
        >
          {useBackupCode ? 'Use authenticator code' : 'Use a backup code'}
        </button>

        {onBack && (
          <div>
            <button
              type="button"
              onClick={onBack}
              className="text-sm text-editor-muted hover:text-editor-text flex items-center justify-center gap-1 mx-auto"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to login
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
