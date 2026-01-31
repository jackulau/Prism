import { useState, useEffect, useCallback } from 'react';
import type { StepType } from '../config/workflowStepTypes';

export interface WorkflowTemplateStep {
  id: string;
  name: string;
  type: StepType;
  config: Record<string, unknown>;
}

export interface WorkflowTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  steps: WorkflowTemplateStep[];
  createdAt?: string;
  updatedAt?: string;
}

const BUILTIN_TEMPLATES: WorkflowTemplate[] = [
  {
    id: 'code-review',
    name: 'Code Review',
    description: 'Automated code review with quality checks and suggestions',
    category: 'development',
    steps: [
      {
        id: 'analyze',
        name: 'Analyze Code',
        type: 'agent',
        config: {
          prompt: 'Analyze the provided code for quality, best practices, and potential issues.',
          outputKey: 'analysis',
        },
      },
      {
        id: 'check-security',
        name: 'Security Check',
        type: 'tool',
        config: {
          toolName: 'security_scan',
          outputKey: 'security_results',
        },
      },
      {
        id: 'has-issues',
        name: 'Has Issues?',
        type: 'condition',
        config: {
          expression: 'state.security_results.issues.length > 0',
          trueBranch: 'generate-fixes',
          falseBranch: 'approve',
        },
      },
      {
        id: 'generate-fixes',
        name: 'Generate Fixes',
        type: 'agent',
        config: {
          prompt: 'Generate fix suggestions for the identified issues.',
          outputKey: 'fixes',
        },
      },
      {
        id: 'approve',
        name: 'Approve',
        type: 'transform',
        config: {
          type: 'template',
          template: '{"status": "approved", "summary": "{{state.analysis.summary}}"}',
        },
      },
    ],
  },
  {
    id: 'debug-assistant',
    name: 'Debug Assistant',
    description: 'Step-by-step debugging workflow with root cause analysis',
    category: 'development',
    steps: [
      {
        id: 'gather-context',
        name: 'Gather Context',
        type: 'parallel',
        config: {
          steps: ['get-logs', 'get-stack-trace'],
          waitForAll: true,
        },
      },
      {
        id: 'analyze-error',
        name: 'Analyze Error',
        type: 'agent',
        config: {
          prompt: 'Analyze the error logs and stack trace to identify the root cause.',
          outputKey: 'root_cause',
        },
      },
      {
        id: 'suggest-fix',
        name: 'Suggest Fix',
        type: 'agent',
        config: {
          prompt: 'Based on the root cause analysis, suggest potential fixes.',
          outputKey: 'suggested_fixes',
        },
      },
      {
        id: 'confirm-fix',
        name: 'Confirm Fix',
        type: 'wait',
        config: {
          waitType: 'user_input',
          promptText: 'Select a fix to apply or provide custom instructions.',
          outputKey: 'selected_fix',
        },
      },
    ],
  },
  {
    id: 'refactoring',
    name: 'Refactoring',
    description: 'Guided code refactoring with before/after comparison',
    category: 'development',
    steps: [
      {
        id: 'identify-smells',
        name: 'Identify Code Smells',
        type: 'agent',
        config: {
          prompt: 'Identify code smells and areas that need refactoring.',
          outputKey: 'code_smells',
        },
      },
      {
        id: 'plan-refactor',
        name: 'Plan Refactoring',
        type: 'agent',
        config: {
          prompt: 'Create a refactoring plan addressing identified issues.',
          outputKey: 'refactor_plan',
        },
      },
      {
        id: 'apply-changes',
        name: 'Apply Changes',
        type: 'tool',
        config: {
          toolName: 'apply_refactoring',
          outputKey: 'refactored_code',
        },
      },
      {
        id: 'validate',
        name: 'Validate Changes',
        type: 'tool',
        config: {
          toolName: 'run_tests',
          outputKey: 'test_results',
        },
      },
    ],
  },
  {
    id: 'documentation',
    name: 'Documentation',
    description: 'Generate comprehensive documentation from code',
    category: 'documentation',
    steps: [
      {
        id: 'extract-structure',
        name: 'Extract Structure',
        type: 'tool',
        config: {
          toolName: 'parse_code',
          outputKey: 'code_structure',
        },
      },
      {
        id: 'generate-docs',
        name: 'Generate Documentation',
        type: 'agent',
        config: {
          prompt: 'Generate comprehensive documentation based on the code structure.',
          outputKey: 'documentation',
        },
      },
      {
        id: 'format-output',
        name: 'Format Output',
        type: 'transform',
        config: {
          type: 'template',
          template: '# {{state.code_structure.name}}\n\n{{state.documentation}}',
          outputKey: 'formatted_docs',
        },
      },
    ],
  },
  {
    id: 'test-generation',
    name: 'Test Generation',
    description: 'Generate unit tests with coverage analysis',
    category: 'testing',
    steps: [
      {
        id: 'analyze-code',
        name: 'Analyze Code',
        type: 'agent',
        config: {
          prompt: 'Analyze the code to identify testable units and edge cases.',
          outputKey: 'test_plan',
        },
      },
      {
        id: 'generate-tests',
        name: 'Generate Tests',
        type: 'agent',
        config: {
          prompt: 'Generate comprehensive unit tests based on the test plan.',
          outputKey: 'generated_tests',
        },
      },
      {
        id: 'run-tests',
        name: 'Run Tests',
        type: 'tool',
        config: {
          toolName: 'run_tests',
          outputKey: 'test_results',
        },
      },
      {
        id: 'check-coverage',
        name: 'Check Coverage',
        type: 'condition',
        config: {
          expression: 'state.test_results.coverage >= 80',
          trueBranch: 'complete',
          falseBranch: 'add-more-tests',
        },
      },
    ],
  },
  {
    id: 'research',
    name: 'Research',
    description: 'Multi-source research with synthesis and citations',
    category: 'research',
    steps: [
      {
        id: 'search-sources',
        name: 'Search Sources',
        type: 'parallel',
        config: {
          steps: ['search-web', 'search-docs', 'search-code'],
          waitForAll: true,
        },
      },
      {
        id: 'analyze-results',
        name: 'Analyze Results',
        type: 'agent',
        config: {
          prompt: 'Analyze and synthesize information from all sources.',
          outputKey: 'analysis',
        },
      },
      {
        id: 'generate-report',
        name: 'Generate Report',
        type: 'agent',
        config: {
          prompt: 'Create a comprehensive research report with citations.',
          outputKey: 'report',
        },
      },
      {
        id: 'format-citations',
        name: 'Format Citations',
        type: 'transform',
        config: {
          type: 'script',
          script: 'formatCitations(state.report, state.sources)',
          outputKey: 'final_report',
        },
      },
    ],
  },
];

interface UseWorkflowTemplatesResult {
  templates: WorkflowTemplate[];
  isLoading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useWorkflowTemplates(): UseWorkflowTemplatesResult {
  const [templates, setTemplates] = useState<WorkflowTemplate[]>(BUILTIN_TEMPLATES);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchTemplates = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      // TODO: Replace with actual API call when endpoint is available
      // const response = await fetch('/api/v1/workflows/templates');
      // const data = await response.json();
      // setTemplates([...BUILTIN_TEMPLATES, ...data]);

      // For now, just use builtin templates with a simulated delay
      await new Promise((resolve) => setTimeout(resolve, 100));
      setTemplates(BUILTIN_TEMPLATES);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch templates');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  return {
    templates,
    isLoading,
    error,
    refetch: fetchTemplates,
  };
}

export function getTemplateById(templates: WorkflowTemplate[], id: string): WorkflowTemplate | undefined {
  return templates.find((t) => t.id === id);
}
