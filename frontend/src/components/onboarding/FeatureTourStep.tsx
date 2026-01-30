import { useState } from 'react';
import {
  MessageSquare,
  Code,
  Settings,
  Keyboard,
  ChevronLeft,
  ChevronRight,
  Layers,
} from 'lucide-react';

interface FeatureTourStepProps {
  onNext: () => void;
  onSkip: () => void;
}

interface TourFeature {
  icon: React.ReactNode;
  title: string;
  description: string;
  tip: string;
}

const TOUR_FEATURES: TourFeature[] = [
  {
    icon: <MessageSquare className="w-8 h-8" />,
    title: 'AI Chat',
    description:
      'Have natural conversations with AI models. Ask questions, get explanations, and receive help with your code.',
    tip: 'Tip: Use @ mentions to reference files in your conversation.',
  },
  {
    icon: <Code className="w-8 h-8" />,
    title: 'Integrated Sandbox',
    description:
      'Run code directly in the browser with our built-in sandbox. Preview results and iterate quickly.',
    tip: 'Tip: Toggle the sandbox panel with the code icon in the toolbar.',
  },
  {
    icon: <Layers className="w-8 h-8" />,
    title: 'Multiple Providers',
    description:
      'Switch between different AI providers and models. Use OpenAI, Anthropic, local Ollama, and more.',
    tip: 'Tip: Each workspace can use a different model.',
  },
  {
    icon: <Settings className="w-8 h-8" />,
    title: 'Customizable',
    description:
      'Configure themes, connect integrations, and customize your experience in Settings.',
    tip: 'Tip: Access settings from the sidebar gear icon.',
  },
  {
    icon: <Keyboard className="w-8 h-8" />,
    title: 'Keyboard Shortcuts',
    description:
      'Work faster with keyboard shortcuts. Press Cmd/Ctrl + K to open the command palette.',
    tip: 'Tip: Press ? to see all available shortcuts.',
  },
];

export function FeatureTourStep({ onNext, onSkip }: FeatureTourStepProps) {
  const [currentFeature, setCurrentFeature] = useState(0);

  const feature = TOUR_FEATURES[currentFeature];
  const isFirst = currentFeature === 0;
  const isLast = currentFeature === TOUR_FEATURES.length - 1;

  const handlePrevious = () => {
    if (!isFirst) {
      setCurrentFeature((prev) => prev - 1);
    }
  };

  const handleNext = () => {
    if (isLast) {
      onNext();
    } else {
      setCurrentFeature((prev) => prev + 1);
    }
  };

  return (
    <div className="flex flex-col items-center min-h-[400px] px-6 py-8 animate-fade-in">
      {/* Feature card */}
      <div className="w-full max-w-md mb-8">
        <div className="p-8 bg-editor-surface border border-editor-border rounded-xl text-center">
          {/* Icon */}
          <div className="w-16 h-16 rounded-2xl bg-primary/10 text-primary flex items-center justify-center mx-auto mb-6">
            {feature.icon}
          </div>

          {/* Content */}
          <h2 className="text-xl font-bold text-editor-text mb-3">{feature.title}</h2>
          <p className="text-editor-muted mb-4">{feature.description}</p>

          {/* Tip */}
          <div className="px-4 py-3 bg-primary/5 border border-primary/10 rounded-lg">
            <p className="text-sm text-primary">{feature.tip}</p>
          </div>
        </div>
      </div>

      {/* Navigation dots */}
      <div className="flex items-center gap-2 mb-6">
        {TOUR_FEATURES.map((_, index) => (
          <button
            key={index}
            onClick={() => setCurrentFeature(index)}
            className={`w-2 h-2 rounded-full transition-all ${
              index === currentFeature
                ? 'w-6 bg-primary'
                : 'bg-editor-border hover:bg-editor-muted'
            }`}
            aria-label={`Go to feature ${index + 1}`}
          />
        ))}
      </div>

      {/* Navigation buttons */}
      <div className="flex items-center gap-4">
        <button
          onClick={handlePrevious}
          disabled={isFirst}
          className="p-2 rounded-lg text-editor-muted hover:text-editor-text hover:bg-editor-surface disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          aria-label="Previous feature"
        >
          <ChevronLeft className="w-6 h-6" />
        </button>

        <button
          onClick={onSkip}
          className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
        >
          Skip tour
        </button>

        <button
          onClick={handleNext}
          className="px-6 py-2.5 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-colors flex items-center gap-2"
        >
          {isLast ? 'Finish Tour' : 'Next'}
          {!isLast && <ChevronRight className="w-4 h-4" />}
        </button>
      </div>
    </div>
  );
}
