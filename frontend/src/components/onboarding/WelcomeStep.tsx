import { Sparkles, MessageSquare, Code, Zap } from 'lucide-react';

interface WelcomeStepProps {
  onNext: () => void;
}

export function WelcomeStep({ onNext }: WelcomeStepProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] px-6 py-8 animate-fade-in">
      {/* Logo/Icon */}
      <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center mb-8 shadow-lg shadow-primary/20">
        <Sparkles className="w-10 h-10 text-white" />
      </div>

      {/* Welcome text */}
      <h1 className="text-3xl font-bold text-editor-text mb-3 text-center">
        Welcome to Prism
      </h1>
      <p className="text-editor-muted text-center max-w-md mb-10">
        Your AI-powered development environment. Let&apos;s get you set up in just a few steps.
      </p>

      {/* Feature highlights */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 w-full max-w-2xl mb-10">
        <FeatureCard
          icon={<MessageSquare className="w-5 h-5" />}
          title="AI Chat"
          description="Natural conversations with AI assistants"
        />
        <FeatureCard
          icon={<Code className="w-5 h-5" />}
          title="Code Tools"
          description="Integrated sandbox and code editing"
        />
        <FeatureCard
          icon={<Zap className="w-5 h-5" />}
          title="Multiple Models"
          description="Connect to various LLM providers"
        />
      </div>

      {/* Get Started button */}
      <button
        onClick={onNext}
        className="px-8 py-3 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-all duration-200 shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5"
      >
        Get Started
      </button>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="p-4 bg-editor-surface border border-editor-border rounded-lg">
      <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center mb-3">
        {icon}
      </div>
      <h3 className="font-medium text-editor-text mb-1">{title}</h3>
      <p className="text-sm text-editor-muted">{description}</p>
    </div>
  );
}
