package workflow

import (
	"fmt"
	"time"
)

// TemplateInfo provides metadata about a workflow template
type TemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	StepCount   int      `json:"step_count"`
}

// GetTemplateInfo returns information about a template
func GetTemplateInfo(id string) *TemplateInfo {
	for _, info := range ListTemplates() {
		if info.ID == id {
			return &info
		}
	}
	return nil
}

// ListTemplates returns all available workflow templates
func ListTemplates() []TemplateInfo {
	return []TemplateInfo{
		{
			ID:          "code-review",
			Name:        "Code Review",
			Description: "Comprehensive code review workflow with analysis, security check, and summary",
			Category:    "development",
			Tags:        []string{"code", "review", "quality"},
			StepCount:   4,
		},
		{
			ID:          "debug",
			Name:        "Debug Workflow",
			Description: "Systematic debugging workflow to identify and fix issues",
			Category:    "development",
			Tags:        []string{"debug", "troubleshoot", "fix"},
			StepCount:   5,
		},
		{
			ID:          "refactor",
			Name:        "Refactoring Workflow",
			Description: "Safe refactoring with analysis, planning, and verification",
			Category:    "development",
			Tags:        []string{"refactor", "improve", "clean"},
			StepCount:   4,
		},
		{
			ID:          "documentation",
			Name:        "Documentation Generator",
			Description: "Generate comprehensive documentation from code",
			Category:    "documentation",
			Tags:        []string{"docs", "readme", "api"},
			StepCount:   3,
		},
		{
			ID:          "test-generation",
			Name:        "Test Generation",
			Description: "Generate unit tests for existing code",
			Category:    "testing",
			Tags:        []string{"test", "unit-test", "coverage"},
			StepCount:   4,
		},
		{
			ID:          "research",
			Name:        "Research Workflow",
			Description: "Research and synthesize information on a topic",
			Category:    "research",
			Tags:        []string{"research", "analyze", "report"},
			StepCount:   4,
		},
	}
}

// GetTemplate returns a workflow definition for the given template ID
func GetTemplate(id string) *WorkflowDefinition {
	switch id {
	case "code-review":
		return CodeReviewTemplate()
	case "debug":
		return DebugTemplate()
	case "refactor":
		return RefactorTemplate()
	case "documentation":
		return DocumentationTemplate()
	case "test-generation":
		return TestGenerationTemplate()
	case "research":
		return ResearchTemplate()
	default:
		return nil
	}
}

// CodeReviewTemplate returns a code review workflow
func CodeReviewTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Code Review",
		Description: "Comprehensive code review workflow",
		Steps: []Step{
			{
				ID:          "analyze",
				Name:        "Analyze Code Structure",
				Description: "Analyze the overall structure and patterns in the code",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are an expert code analyst. Analyze code structure, patterns, and architecture.",
						Prompt: `Analyze the following code and provide a structural overview:

{{state.code}}

Focus on:
1. Overall architecture and design patterns
2. Code organization and modularity
3. Dependencies and coupling
4. Potential areas of concern`,
						OutputKey: "structure_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "security",
				Name:        "Security Analysis",
				Description: "Check for security vulnerabilities and issues",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a security expert. Identify potential security vulnerabilities.",
						Prompt: `Review the following code for security vulnerabilities:

{{state.code}}

Check for:
1. Input validation issues
2. Authentication/authorization problems
3. Data exposure risks
4. Injection vulnerabilities
5. Insecure dependencies`,
						OutputKey: "security_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "quality",
				Name:        "Code Quality Review",
				Description: "Review code quality, readability, and best practices",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a senior developer focused on code quality and best practices.",
						Prompt: `Review the code quality of:

{{state.code}}

Evaluate:
1. Readability and clarity
2. Error handling
3. Test coverage considerations
4. Performance implications
5. Best practice adherence`,
						OutputKey: "quality_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "summarize",
				Name:        "Generate Summary",
				Description: "Synthesize all analyses into a final review summary",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a technical lead writing comprehensive code reviews.",
						Prompt: `Based on the following analyses, generate a comprehensive code review summary:

Structure Analysis:
{{state.structure_analysis}}

Security Analysis:
{{state.security_analysis}}

Quality Analysis:
{{state.quality_analysis}}

Provide:
1. Executive summary
2. Key findings (prioritized)
3. Specific recommendations
4. Overall assessment`,
						OutputKey: "review_summary",
					},
				},
				Timeout: 2 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-5-sonnet-20241022",
		},
	}
}

// DebugTemplate returns a debugging workflow
func DebugTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Debug Workflow",
		Description: "Systematic debugging workflow",
		Steps: []Step{
			{
				ID:          "understand",
				Name:        "Understand the Problem",
				Description: "Analyze the reported issue and gather context",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a debugging expert. Analyze problems systematically.",
						Prompt: `Analyze the following problem:

Issue Description:
{{state.issue}}

Related Code:
{{state.code}}

Error Messages:
{{state.errors}}

Provide:
1. Problem summary
2. Potential root causes
3. Areas to investigate`,
						OutputKey: "problem_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "hypothesize",
				Name:        "Form Hypotheses",
				Description: "Generate possible causes and solutions",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a systematic debugger forming hypotheses.",
						Prompt: `Based on the problem analysis:

{{state.problem_analysis}}

Generate:
1. List of hypotheses (most to least likely)
2. How to test each hypothesis
3. What evidence would confirm/refute each`,
						OutputKey: "hypotheses",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "investigate",
				Name:        "Investigation Plan",
				Description: "Create a detailed investigation plan",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are creating a debugging investigation plan.",
						Prompt: `Create an investigation plan based on:

Hypotheses:
{{state.hypotheses}}

Provide:
1. Step-by-step investigation plan
2. Commands/code to run for debugging
3. What to look for at each step`,
						OutputKey: "investigation_plan",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "wait-results",
				Name:        "Wait for Investigation Results",
				Description: "Pause for user to provide investigation results",
				Type:        StepTypeWait,
				Config: StepConfig{
					WaitConfig: &WaitStepConfig{
						WaitType:   "user_input",
						PromptText: "Please run the investigation plan and provide the results",
						Timeout:    30 * time.Minute,
						OutputKey:  "investigation_results",
					},
				},
			},
			{
				ID:          "solution",
				Name:        "Propose Solution",
				Description: "Propose a fix based on investigation results",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are proposing a bug fix based on investigation results.",
						Prompt: `Based on the investigation:

Original Issue:
{{state.issue}}

Investigation Results:
{{state.investigation_results}}

Provide:
1. Root cause identification
2. Proposed fix (with code if applicable)
3. How to verify the fix
4. Preventive measures`,
						OutputKey: "proposed_solution",
					},
				},
				Timeout: 2 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-5-sonnet-20241022",
		},
	}
}

// RefactorTemplate returns a refactoring workflow
func RefactorTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Refactoring Workflow",
		Description: "Safe refactoring with analysis and verification",
		Steps: []Step{
			{
				ID:          "analyze",
				Name:        "Analyze Current Code",
				Description: "Understand the code to be refactored",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a refactoring expert analyzing code for improvement.",
						Prompt: `Analyze the following code for refactoring:

{{state.code}}

Refactoring Goal:
{{state.goal}}

Identify:
1. Current code structure and patterns
2. Code smells and issues
3. Dependencies and side effects
4. Test coverage (if visible)`,
						OutputKey: "code_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "plan",
				Name:        "Create Refactoring Plan",
				Description: "Plan the refactoring steps",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are creating a safe refactoring plan.",
						Prompt: `Based on the analysis:

{{state.code_analysis}}

Create a refactoring plan that:
1. Lists changes in order of execution
2. Identifies safe stopping points
3. Notes potential risks at each step
4. Suggests tests to run after each step`,
						OutputKey: "refactor_plan",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "implement",
				Name:        "Generate Refactored Code",
				Description: "Generate the refactored code",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are implementing a refactoring plan.",
						Prompt: `Implement the refactoring plan:

Original Code:
{{state.code}}

Plan:
{{state.refactor_plan}}

Provide:
1. Complete refactored code
2. Explanation of each change
3. Any new tests needed`,
						OutputKey: "refactored_code",
					},
				},
				Timeout: 3 * time.Minute,
			},
			{
				ID:          "verify",
				Name:        "Verification Checklist",
				Description: "Create verification checklist",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are creating a verification checklist for refactored code.",
						Prompt: `Create a verification checklist for the refactoring:

Original:
{{state.code}}

Refactored:
{{state.refactored_code}}

Checklist should include:
1. Functional equivalence tests
2. Edge cases to verify
3. Performance considerations
4. Integration points to check`,
						OutputKey: "verification_checklist",
					},
				},
				Timeout: 2 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-5-sonnet-20241022",
		},
	}
}

// DocumentationTemplate returns a documentation generation workflow
func DocumentationTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Documentation Generator",
		Description: "Generate comprehensive documentation",
		Steps: []Step{
			{
				ID:          "analyze",
				Name:        "Analyze Code",
				Description: "Analyze code structure for documentation",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a documentation expert analyzing code.",
						Prompt: `Analyze the following code for documentation:

{{state.code}}

Documentation Type: {{state.doc_type}}

Identify:
1. Public APIs and interfaces
2. Key concepts and data structures
3. Usage patterns
4. Configuration options`,
						OutputKey: "code_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "generate",
				Name:        "Generate Documentation",
				Description: "Generate the documentation content",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a technical writer creating documentation.",
						Prompt: `Generate documentation based on:

Analysis:
{{state.code_analysis}}

Documentation Type: {{state.doc_type}}

Create:
1. Overview/Introduction
2. API Reference (if applicable)
3. Usage examples
4. Configuration guide
5. Troubleshooting section`,
						OutputKey: "documentation",
					},
				},
				Timeout: 3 * time.Minute,
			},
			{
				ID:          "examples",
				Name:        "Generate Examples",
				Description: "Create practical code examples",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are creating practical code examples.",
						Prompt: `Based on the documentation:

{{state.documentation}}

Create practical examples that:
1. Show basic usage
2. Demonstrate common patterns
3. Handle error cases
4. Show advanced features`,
						OutputKey: "examples",
					},
				},
				Timeout: 2 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-5-sonnet-20241022",
			"doc_type": "API",
		},
	}
}

// TestGenerationTemplate returns a test generation workflow
func TestGenerationTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Test Generation",
		Description: "Generate unit tests for code",
		Steps: []Step{
			{
				ID:          "analyze",
				Name:        "Analyze Code",
				Description: "Analyze code for test generation",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a testing expert analyzing code.",
						Prompt: `Analyze the following code for test generation:

{{state.code}}

Testing Framework: {{state.framework}}

Identify:
1. Functions/methods to test
2. Input parameters and types
3. Expected outputs
4. Edge cases and error conditions
5. Dependencies to mock`,
						OutputKey: "test_analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "plan",
				Name:        "Test Plan",
				Description: "Create test cases plan",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are planning comprehensive test coverage.",
						Prompt: `Based on the analysis:

{{state.test_analysis}}

Create a test plan:
1. List all test cases
2. Group by function/feature
3. Include happy path and edge cases
4. Note any mocking requirements`,
						OutputKey: "test_plan",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "generate",
				Name:        "Generate Tests",
				Description: "Generate the actual test code",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are writing comprehensive unit tests.",
						Prompt: `Generate tests based on:

Code:
{{state.code}}

Test Plan:
{{state.test_plan}}

Framework: {{state.framework}}

Generate complete, runnable test code.`,
						OutputKey: "test_code",
					},
				},
				Timeout: 3 * time.Minute,
			},
			{
				ID:          "review",
				Name:        "Review Tests",
				Description: "Review generated tests for completeness",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are reviewing tests for quality and coverage.",
						Prompt: `Review the generated tests:

{{state.test_code}}

Against the original code:
{{state.code}}

Evaluate:
1. Test coverage completeness
2. Edge case coverage
3. Test quality and maintainability
4. Any missing tests`,
						OutputKey: "test_review",
					},
				},
				Timeout: 2 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider":  "anthropic",
			"model":     "claude-3-5-sonnet-20241022",
			"framework": "jest",
		},
	}
}

// ResearchTemplate returns a research workflow
func ResearchTemplate() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "Research Workflow",
		Description: "Research and synthesize information",
		Steps: []Step{
			{
				ID:          "understand",
				Name:        "Understand Topic",
				Description: "Understand the research topic and scope",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are a research analyst understanding a topic.",
						Prompt: `Understand the following research topic:

Topic: {{state.topic}}

Context: {{state.context}}

Provide:
1. Key questions to answer
2. Scope definition
3. Areas to research
4. Expected deliverables`,
						OutputKey: "research_scope",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "gather",
				Name:        "Gather Information",
				Description: "Gather relevant information",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are gathering information for research.",
						Prompt: `Based on the research scope:

{{state.research_scope}}

Gather information on:
1. Current state of knowledge
2. Key facts and data
3. Different perspectives
4. Relevant sources`,
						OutputKey: "gathered_info",
					},
				},
				Timeout: 3 * time.Minute,
			},
			{
				ID:          "analyze",
				Name:        "Analyze Findings",
				Description: "Analyze and synthesize findings",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are analyzing research findings.",
						Prompt: `Analyze the gathered information:

{{state.gathered_info}}

Provide:
1. Key insights
2. Patterns and trends
3. Conflicting information
4. Gaps in knowledge`,
						OutputKey: "analysis",
					},
				},
				Timeout: 2 * time.Minute,
			},
			{
				ID:          "report",
				Name:        "Generate Report",
				Description: "Generate the research report",
				Type:        StepTypeAgent,
				Config: StepConfig{
					AgentConfig: &AgentStepConfig{
						Provider:     "{{state.provider}}",
						Model:        "{{state.model}}",
						SystemPrompt: "You are writing a research report.",
						Prompt: `Generate a research report:

Topic: {{state.topic}}
Scope: {{state.research_scope}}
Analysis: {{state.analysis}}

Create a comprehensive report with:
1. Executive summary
2. Background
3. Findings
4. Analysis
5. Recommendations
6. Conclusion`,
						OutputKey: "report",
					},
				},
				Timeout: 3 * time.Minute,
			},
		},
		InitialState: map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-5-sonnet-20241022",
		},
	}
}

// CreateFromTemplate creates a new workflow definition from a template
func CreateFromTemplate(templateID string, customState map[string]interface{}) (*WorkflowDefinition, error) {
	template := GetTemplate(templateID)
	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Clone the template
	def := &WorkflowDefinition{
		Name:         template.Name,
		Description:  template.Description,
		Steps:        make([]Step, len(template.Steps)),
		InitialState: make(map[string]interface{}),
	}

	// Copy steps
	copy(def.Steps, template.Steps)

	// Merge initial state with custom state
	for k, v := range template.InitialState {
		def.InitialState[k] = v
	}
	for k, v := range customState {
		def.InitialState[k] = v
	}

	return def, nil
}

// Error for template not found
type ErrTemplateNotFound struct {
	ID string
}

func (e ErrTemplateNotFound) Error() string {
	return fmt.Sprintf("template not found: %s", e.ID)
}
