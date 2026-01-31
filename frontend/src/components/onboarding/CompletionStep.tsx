import { useNavigate } from 'react-router-dom';
import { CheckCircle, ArrowRight, Settings, MessageSquare, Sparkles } from 'lucide-react';

interface CompletionStepProps {
  onComplete: () => void;
}

export function CompletionStep({ onComplete }: CompletionStepProps) {
  const navigate = useNavigate();

  const handleGoToDashboard = () => {
    onComplete();
    navigate('/');
  };

  const handleGoToWorkspace = () => {
    onComplete();
    navigate('/workspace');
  };

  const handleGoToSettings = () => {
    onComplete();
    navigate('/settings');
  };

  return (
    <div className="flex flex-col items-center min-h-[400px] px-6 py-8 animate-fade-in">
      {/* Success icon */}
      <div className="w-20 h-20 rounded-full bg-green-500/10 text-green-500 flex items-center justify-center mb-6">
        <CheckCircle className="w-10 h-10" />
      </div>

      <h2 className="text-2xl font-bold text-editor-text mb-2 text-center">
        You&apos;re All Set!
      </h2>
      <p className="text-editor-muted text-center max-w-md mb-10">
        Your Prism workspace is ready. Start chatting with AI or explore more features.
      </p>

      {/* Next steps */}
      <div className="w-full max-w-md space-y-3 mb-10">
        <NextStepCard
          icon={<MessageSquare className="w-5 h-5" />}
          title="Start a conversation"
          description="Open a workspace and start chatting with AI"
          onClick={handleGoToWorkspace}
        />
        <NextStepCard
          icon={<Sparkles className="w-5 h-5" />}
          title="Explore the dashboard"
          description="View recent workspaces and quick actions"
          onClick={handleGoToDashboard}
        />
        <NextStepCard
          icon={<Settings className="w-5 h-5" />}
          title="Customize settings"
          description="Configure themes, add providers, and more"
          onClick={handleGoToSettings}
        />
      </div>

      {/* Primary action */}
      <button
        onClick={handleGoToWorkspace}
        className="px-8 py-3 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-all duration-200 shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/30 flex items-center gap-2"
      >
        Start Using Prism
        <ArrowRight className="w-4 h-4" />
      </button>
    </div>
  );
}

function NextStepCard({
  icon,
  title,
  description,
  onClick,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="w-full p-4 bg-editor-surface border border-editor-border rounded-lg flex items-center gap-4 hover:border-primary/30 hover:bg-editor-surface/80 transition-colors text-left group"
    >
      <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center flex-shrink-0 group-hover:bg-primary/20 transition-colors">
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="font-medium text-editor-text">{title}</h3>
        <p className="text-sm text-editor-muted truncate">{description}</p>
      </div>
      <ArrowRight className="w-4 h-4 text-editor-muted group-hover:text-primary transition-colors flex-shrink-0" />
    </button>
  );
}
