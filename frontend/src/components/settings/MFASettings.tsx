import { useState, useEffect, useRef } from 'react';
import { useMFAStore } from '../../store/mfaStore';
import { useAuthStore } from '../../store/authStore';
import {
  Shield,
  ShieldCheck,
  ShieldOff,
  Loader2,
  AlertTriangle,
  Copy,
  CheckCircle,
  Download,
  Eye,
  EyeOff,
  RefreshCw,
  X,
} from 'lucide-react';

type SetupStep = 'disabled' | 'qr-code' | 'verify' | 'backup-codes' | 'enabled';

export function MFASettings() {
  const { accessToken } = useAuthStore();
  const {
    isEnabled,
    isLoading,
    setupData,
    backupCodes,
    error,
    setToken,
    fetchStatus,
    startSetup,
    verifySetup,
    disable,
    regenerateBackupCodes,
    clearSetupData,
    clearBackupCodes,
    clearError,
  } = useMFAStore();

  const [step, setStep] = useState<SetupStep>('disabled');
  const [verificationCode, setVerificationCode] = useState(['', '', '', '', '', '']);
  const [showManualEntry, setShowManualEntry] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);
  const [copiedBackupCodes, setCopiedBackupCodes] = useState(false);
  const [showDisableModal, setShowDisableModal] = useState(false);
  const [showRegenerateModal, setShowRegenerateModal] = useState(false);
  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [regenerateCode, setRegenerateCode] = useState('');
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    if (accessToken) {
      setToken(accessToken);
      fetchStatus();
    }
  }, [accessToken, setToken, fetchStatus]);

  useEffect(() => {
    if (isEnabled) {
      setStep('enabled');
    } else if (setupData) {
      setStep('qr-code');
    } else if (backupCodes && backupCodes.length > 0) {
      setStep('backup-codes');
    } else {
      setStep('disabled');
    }
  }, [isEnabled, setupData, backupCodes]);

  const handleStartSetup = async () => {
    clearError();
    await startSetup();
  };

  const handleCodeChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;

    const newCode = [...verificationCode];
    newCode[index] = value.slice(-1);
    setVerificationCode(newCode);

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === 'Backspace' && !verificationCode[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const paste = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const newCode = [...verificationCode];
    for (let i = 0; i < paste.length; i++) {
      newCode[i] = paste[i];
    }
    setVerificationCode(newCode);
    if (paste.length > 0) {
      inputRefs.current[Math.min(paste.length, 5)]?.focus();
    }
  };

  const handleVerifySetup = async () => {
    const code = verificationCode.join('');
    if (code.length !== 6) return;

    clearError();
    const success = await verifySetup(code);
    if (success) {
      setVerificationCode(['', '', '', '', '', '']);
    }
  };

  const handleCancelSetup = () => {
    clearSetupData();
    clearError();
    setVerificationCode(['', '', '', '', '', '']);
    setShowManualEntry(false);
    setStep('disabled');
  };

  const handleDismissBackupCodes = () => {
    clearBackupCodes();
  };

  const handleCopySecret = async () => {
    if (setupData?.secret) {
      await navigator.clipboard.writeText(setupData.secret);
      setCopiedSecret(true);
      setTimeout(() => setCopiedSecret(false), 2000);
    }
  };

  const handleCopyBackupCodes = async () => {
    if (backupCodes) {
      await navigator.clipboard.writeText(backupCodes.join('\n'));
      setCopiedBackupCodes(true);
      setTimeout(() => setCopiedBackupCodes(false), 2000);
    }
  };

  const handleDownloadBackupCodes = () => {
    if (backupCodes) {
      const content = `Prism MFA Backup Codes\n${'='.repeat(25)}\n\nKeep these codes safe. Each code can only be used once.\n\n${backupCodes.join('\n')}\n\nGenerated: ${new Date().toISOString()}`;
      const blob = new Blob([content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'prism-mfa-backup-codes.txt';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  const handleDisable = async () => {
    clearError();
    const success = await disable(disablePassword, disableCode);
    if (success) {
      setShowDisableModal(false);
      setDisablePassword('');
      setDisableCode('');
    }
  };

  const handleRegenerate = async () => {
    clearError();
    await regenerateBackupCodes(regenerateCode);
    if (!error) {
      setShowRegenerateModal(false);
      setRegenerateCode('');
    }
  };

  if (isLoading && step === 'disabled' && !setupData) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="w-6 h-6 animate-spin text-editor-muted" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-3">
        {isEnabled ? (
          <ShieldCheck className="w-5 h-5 text-green-500" />
        ) : (
          <Shield className="w-5 h-5 text-editor-muted" />
        )}
        <div>
          <p className="font-medium">Two-Factor Authentication</p>
          <p className="text-sm text-editor-muted">
            {isEnabled
              ? 'Your account is protected with 2FA'
              : 'Add an extra layer of security to your account'}
          </p>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
          <AlertTriangle className="w-4 h-4 flex-shrink-0" />
          {error}
        </div>
      )}

      {/* Disabled State - Enable Button */}
      {step === 'disabled' && (
        <div className="p-4 bg-editor-bg rounded-lg">
          <div className="flex items-start gap-3 mb-4">
            <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
            <div className="text-sm">
              <p className="font-medium text-yellow-500">Recommended</p>
              <p className="text-editor-muted mt-1">
                Two-factor authentication adds an extra layer of security by requiring
                a code from your authenticator app in addition to your password.
              </p>
            </div>
          </div>
          <button
            onClick={handleStartSetup}
            disabled={isLoading}
            className="w-full px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
          >
            {isLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Shield className="w-4 h-4" />
            )}
            Enable Two-Factor Authentication
          </button>
        </div>
      )}

      {/* QR Code Step */}
      {step === 'qr-code' && setupData && (
        <div className="p-4 bg-editor-bg rounded-lg space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">Step 1: Scan QR Code</h4>
            <button
              onClick={handleCancelSetup}
              className="p-1 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <p className="text-sm text-editor-muted">
            Scan this QR code with your authenticator app (Google Authenticator,
            Authy, 1Password, etc.)
          </p>

          <div className="flex justify-center">
            <div className="bg-white p-4 rounded-lg">
              <img
                src={setupData.qrCodeUrl}
                alt="MFA QR Code"
                className="w-48 h-48"
              />
            </div>
          </div>

          <div className="border-t border-editor-border pt-4">
            <button
              onClick={() => setShowManualEntry(!showManualEntry)}
              className="text-sm text-editor-accent hover:underline flex items-center gap-1"
            >
              {showManualEntry ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              {showManualEntry ? 'Hide manual entry' : "Can't scan? Enter manually"}
            </button>

            {showManualEntry && (
              <div className="mt-3 p-3 bg-editor-surface rounded-lg">
                <p className="text-xs text-editor-muted mb-2">Secret key:</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-sm font-mono bg-editor-bg p-2 rounded border border-editor-border break-all">
                    {setupData.secret}
                  </code>
                  <button
                    onClick={handleCopySecret}
                    className="p-2 hover:bg-editor-bg rounded text-editor-muted hover:text-editor-text"
                  >
                    {copiedSecret ? (
                      <CheckCircle className="w-4 h-4 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>
            )}
          </div>

          <button
            onClick={() => setStep('verify')}
            className="w-full px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors"
          >
            Continue
          </button>
        </div>
      )}

      {/* Verification Step */}
      {step === 'verify' && (
        <div className="p-4 bg-editor-bg rounded-lg space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">Step 2: Verify Code</h4>
            <button
              onClick={handleCancelSetup}
              className="p-1 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <p className="text-sm text-editor-muted">
            Enter the 6-digit code from your authenticator app to verify setup.
          </p>

          <div className="flex justify-center gap-2" onPaste={handlePaste}>
            {verificationCode.map((digit, index) => (
              <input
                key={index}
                ref={(el) => (inputRefs.current[index] = el)}
                type="text"
                inputMode="numeric"
                maxLength={1}
                value={digit}
                onChange={(e) => handleCodeChange(index, e.target.value)}
                onKeyDown={(e) => handleKeyDown(index, e)}
                className="w-12 h-14 text-center text-xl font-mono bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
              />
            ))}
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => setStep('qr-code')}
              className="flex-1 px-4 py-2 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface transition-colors"
            >
              Back
            </button>
            <button
              onClick={handleVerifySetup}
              disabled={isLoading || verificationCode.join('').length !== 6}
              className="flex-1 px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
            >
              {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
              Verify & Enable
            </button>
          </div>
        </div>
      )}

      {/* Backup Codes Display */}
      {step === 'backup-codes' && backupCodes && (
        <div className="p-4 bg-editor-bg rounded-lg space-y-4">
          <div className="flex items-start gap-3">
            <ShieldCheck className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-medium text-green-500">2FA Enabled Successfully!</h4>
              <p className="text-sm text-editor-muted mt-1">
                Save these backup codes in a secure location. You can use them to
                access your account if you lose your authenticator device.
              </p>
            </div>
          </div>

          <div className="p-4 bg-editor-surface rounded-lg">
            <div className="grid grid-cols-2 gap-2">
              {backupCodes.map((code, index) => (
                <code key={index} className="font-mono text-sm p-2 bg-editor-bg rounded border border-editor-border text-center">
                  {code}
                </code>
              ))}
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleCopyBackupCodes}
              className="flex-1 px-4 py-2 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface transition-colors flex items-center justify-center gap-2"
            >
              {copiedBackupCodes ? (
                <CheckCircle className="w-4 h-4 text-green-500" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
              Copy
            </button>
            <button
              onClick={handleDownloadBackupCodes}
              className="flex-1 px-4 py-2 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface transition-colors flex items-center justify-center gap-2"
            >
              <Download className="w-4 h-4" />
              Download
            </button>
          </div>

          <button
            onClick={handleDismissBackupCodes}
            className="w-full px-4 py-2 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors"
          >
            I&apos;ve Saved My Backup Codes
          </button>
        </div>
      )}

      {/* Enabled State */}
      {step === 'enabled' && !backupCodes && (
        <div className="p-4 bg-editor-bg rounded-lg space-y-4">
          <div className="flex items-center gap-3">
            <CheckCircle className="w-5 h-5 text-green-500" />
            <div>
              <p className="font-medium">Two-factor authentication is enabled</p>
              <p className="text-sm text-editor-muted">
                Your account is protected with an authenticator app
              </p>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => setShowRegenerateModal(true)}
              className="flex-1 px-4 py-2 border border-editor-border text-editor-text rounded-lg font-medium hover:bg-editor-surface transition-colors flex items-center justify-center gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              Regenerate Backup Codes
            </button>
            <button
              onClick={() => setShowDisableModal(true)}
              className="flex-1 px-4 py-2 border border-red-500/30 text-red-400 rounded-lg font-medium hover:bg-red-500/10 transition-colors flex items-center justify-center gap-2"
            >
              <ShieldOff className="w-4 h-4" />
              Disable 2FA
            </button>
          </div>
        </div>
      )}

      {/* Disable Modal */}
      {showDisableModal && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={() => setShowDisableModal(false)}
          />
          <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Disable Two-Factor Authentication</h3>
              <button
                onClick={() => setShowDisableModal(false)}
                className="p-1 hover:bg-editor-surface rounded"
              >
                <X className="w-5 h-5 text-editor-muted" />
              </button>
            </div>

            <div className="space-y-4">
              <div className="flex items-start gap-3 p-3 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
                <p className="text-sm text-editor-muted">
                  Disabling 2FA will make your account less secure. You will only
                  need your password to sign in.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Password</label>
                <input
                  type="password"
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                  placeholder="Enter your password"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Verification Code</label>
                <input
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  value={disableCode}
                  onChange={(e) => setDisableCode(e.target.value.replace(/\D/g, ''))}
                  placeholder="Enter 6-digit code"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm font-mono"
                />
              </div>

              {error && (
                <div className="text-sm text-red-400">{error}</div>
              )}

              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setShowDisableModal(false)}
                  className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text"
                >
                  Cancel
                </button>
                <button
                  onClick={handleDisable}
                  disabled={!disablePassword || disableCode.length !== 6 || isLoading}
                  className="px-4 py-2 bg-red-500 text-white text-sm rounded-lg hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                  Disable 2FA
                </button>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Regenerate Modal */}
      {showRegenerateModal && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={() => setShowRegenerateModal(false)}
          />
          <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Regenerate Backup Codes</h3>
              <button
                onClick={() => setShowRegenerateModal(false)}
                className="p-1 hover:bg-editor-surface rounded"
              >
                <X className="w-5 h-5 text-editor-muted" />
              </button>
            </div>

            <div className="space-y-4">
              <div className="flex items-start gap-3 p-3 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
                <p className="text-sm text-editor-muted">
                  This will invalidate all your existing backup codes. Make sure
                  to save the new codes in a secure location.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Verification Code</label>
                <input
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  value={regenerateCode}
                  onChange={(e) => setRegenerateCode(e.target.value.replace(/\D/g, ''))}
                  placeholder="Enter 6-digit code"
                  className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-sm font-mono"
                />
              </div>

              {error && (
                <div className="text-sm text-red-400">{error}</div>
              )}

              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setShowRegenerateModal(false)}
                  className="px-4 py-2 text-sm text-editor-muted hover:text-editor-text"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRegenerate}
                  disabled={regenerateCode.length !== 6 || isLoading}
                  className="px-4 py-2 bg-primary text-white text-sm rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                  Regenerate
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
