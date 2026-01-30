import { useState } from 'react';
import { Key, Copy, CheckCircle, AlertTriangle } from 'lucide-react';

interface NewKeyDisplayProps {
  isOpen: boolean;
  keyValue: string;
  onClose: () => void;
}

export function NewKeyDisplay({ isOpen, keyValue, onClose }: NewKeyDisplayProps) {
  const [copied, setCopied] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);

  if (!isOpen || !keyValue) return null;

  const handleCopy = async () => {
    await navigator.clipboard.writeText(keyValue);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleClose = () => {
    if (acknowledged) {
      setAcknowledged(false);
      setCopied(false);
      onClose();
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50 z-50" />

      {/* Modal */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[520px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50">
        {/* Header */}
        <div className="flex items-center gap-3 p-4 border-b border-editor-border bg-editor-surface rounded-t-lg">
          <div className="p-2 bg-editor-accent/20 rounded-lg">
            <Key className="w-5 h-5 text-editor-accent" />
          </div>
          <h3 className="text-lg font-semibold">Your New API Key</h3>
        </div>

        <div className="p-6 space-y-4">
          {/* Key Display */}
          <div className="relative">
            <div className="flex items-center bg-editor-surface border-2 border-editor-accent/50 rounded-lg overflow-hidden">
              <code className="flex-1 px-4 py-3 text-sm font-mono text-editor-text truncate select-all">
                {keyValue}
              </code>
              <button
                onClick={handleCopy}
                className="px-4 py-3 bg-editor-accent/10 hover:bg-editor-accent/20 transition-colors border-l border-editor-accent/30"
                title="Copy to clipboard"
              >
                {copied ? (
                  <CheckCircle className="w-5 h-5 text-green-500" />
                ) : (
                  <Copy className="w-5 h-5 text-editor-accent" />
                )}
              </button>
            </div>
            {copied && (
              <div className="absolute -top-2 right-0 px-2 py-0.5 bg-green-500 text-white text-xs rounded">
                Copied!
              </div>
            )}
          </div>

          {/* Warning */}
          <div className="flex items-start gap-3 p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
            <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
            <div className="text-sm">
              <p className="font-medium text-yellow-500">This key will only be shown once</p>
              <p className="text-editor-muted mt-1">
                Make sure to copy and store it securely. You won&apos;t be able to see it again after closing this dialog.
              </p>
            </div>
          </div>

          {/* Acknowledgment checkbox */}
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={acknowledged}
              onChange={(e) => setAcknowledged(e.target.checked)}
              className="rounded border-editor-border text-editor-accent focus:ring-editor-accent"
            />
            <span className="text-sm">I have copied my API key</span>
          </label>
        </div>

        {/* Footer */}
        <div className="flex justify-end p-4 border-t border-editor-border">
          <button
            onClick={handleClose}
            disabled={!acknowledged}
            className={`px-6 py-2 text-sm rounded-lg transition-colors ${
              acknowledged
                ? 'bg-primary text-white hover:bg-primary/90'
                : 'bg-editor-surface text-editor-muted cursor-not-allowed'
            }`}
          >
            Close
          </button>
        </div>
      </div>
    </>
  );
}
