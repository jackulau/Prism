import { useNavigate } from 'react-router-dom';
import { Plus, Bot, Plug, BarChart3 } from 'lucide-react';

export function QuickActions() {
  const navigate = useNavigate();

  const actions = [
    {
      icon: Plus,
      label: 'New Workspace',
      description: 'Start a new coding session',
      onClick: () => navigate('/workspace'),
      primary: true,
    },
    {
      icon: Bot,
      label: 'Manage Workers',
      description: 'Configure AI agents',
      onClick: () => navigate('/workers'),
    },
    {
      icon: Plug,
      label: 'Integrations',
      description: 'Connect external services',
      onClick: () => navigate('/integrations'),
    },
    {
      icon: BarChart3,
      label: 'View Usage',
      description: 'Monitor your usage',
      onClick: () => navigate('/usage'),
    },
  ];

  return (
    <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {actions.map((action) => {
        const Icon = action.icon;
        return (
          <button
            key={action.label}
            onClick={action.onClick}
            className={`group flex flex-col items-start gap-3 p-4 rounded-lg border transition-all ${
              action.primary
                ? 'bg-editor-accent text-white border-editor-accent hover:bg-editor-accent/90'
                : 'bg-editor-surface border-editor-border hover:border-editor-accent/50 hover:bg-editor-surface/80'
            }`}
          >
            <div
              className={`p-2 rounded-lg ${
                action.primary
                  ? 'bg-white/20'
                  : 'bg-editor-accent/10 group-hover:bg-editor-accent/20'
              }`}
            >
              <Icon
                size={20}
                className={action.primary ? 'text-white' : 'text-editor-accent'}
              />
            </div>
            <div>
              <h3
                className={`font-medium ${
                  action.primary ? 'text-white' : 'text-editor-text'
                }`}
              >
                {action.label}
              </h3>
              <p
                className={`text-sm ${
                  action.primary ? 'text-white/80' : 'text-editor-muted'
                }`}
              >
                {action.description}
              </p>
            </div>
          </button>
        );
      })}
    </section>
  );
}
