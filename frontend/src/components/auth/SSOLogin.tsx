import { useState } from 'react';
import { ArrowLeft } from 'lucide-react';
import { initiateSSO } from '../../store/authStore';

interface SSOLoginProps {
  onBack?: () => void;
}

export function SSOLogin({ onBack }: SSOLoginProps) {
  const [organization, setOrganization] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await initiateSSO(organization);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'SSO login failed');
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="flex items-center gap-2 mb-6">
        {onBack && (
          <button
            type="button"
            onClick={onBack}
            className="p-2 rounded-lg hover:bg-editor-surface text-editor-muted hover:text-editor-text transition-colors"
            aria-label="Back to login"
          >
            <ArrowLeft size={20} />
          </button>
        )}
        <h2 className="text-2xl font-bold">Sign in with SSO</h2>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label htmlFor="organization" className="block text-sm font-medium mb-1">
            Organization
          </label>
          <input
            id="organization"
            type="text"
            value={organization}
            onChange={(e) => setOrganization(e.target.value)}
            className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="your-company or your-company.com"
            required
            disabled={isLoading}
          />
          <p className="mt-1 text-xs text-editor-muted">
            Enter your organization's domain or identifier
          </p>
        </div>

        <button
          type="submit"
          disabled={isLoading || !organization.trim()}
          className="w-full py-2 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isLoading ? 'Redirecting...' : 'Continue with SSO'}
        </button>
      </form>
    </div>
  );
}
