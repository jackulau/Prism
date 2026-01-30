import { useState } from 'react';
import { Eye, EyeOff, Copy, Check, AlertCircle } from 'lucide-react';

interface EnvVarInputProps {
  value: string;
  isSecret: boolean;
  onChange?: (value: string) => void;
  placeholder?: string;
  readOnly?: boolean;
  error?: string;
}

export function EnvVarInput({
  value,
  isSecret,
  onChange,
  placeholder = 'Value',
  readOnly = false,
  error,
}: EnvVarInputProps) {
  const [isVisible, setIsVisible] = useState(!isSecret);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API not available
    }
  };

  const displayValue = isSecret && !isVisible && value ? '••••••••' : value;

  return (
    <div className="relative flex-1">
      <div className="flex items-center gap-1">
        <div className="relative flex-1">
          <input
            type={isSecret && !isVisible ? 'password' : 'text'}
            value={displayValue}
            onChange={(e) => onChange?.(e.target.value)}
            placeholder={placeholder}
            readOnly={readOnly}
            className={`w-full px-3 py-2 pr-16 bg-editor-surface border rounded-lg text-sm text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent ${
              error ? 'border-red-500' : 'border-editor-border'
            } ${readOnly ? 'cursor-default opacity-75' : ''}`}
          />
          <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
            {isSecret && (
              <button
                type="button"
                onClick={() => setIsVisible(!isVisible)}
                className="p-1 text-editor-muted hover:text-editor-text transition-colors"
                title={isVisible ? 'Hide value' : 'Show value'}
              >
                {isVisible ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
            )}
            <button
              type="button"
              onClick={handleCopy}
              className="p-1 text-editor-muted hover:text-editor-text transition-colors"
              title="Copy to clipboard"
            >
              {copied ? (
                <Check size={14} className="text-green-500" />
              ) : (
                <Copy size={14} />
              )}
            </button>
          </div>
        </div>
      </div>
      {error && (
        <div className="flex items-center gap-1 mt-1 text-xs text-red-400">
          <AlertCircle size={12} />
          {error}
        </div>
      )}
    </div>
  );
}
