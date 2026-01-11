package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/urfave/cli/v2"
)

// DesignConfig holds configuration for the design command.
type DesignConfig struct {
	Provider  string
	Prompt    string
	OutputDir string
	Model     string
	APIKey    string
}

// Validate validates the design configuration.
func (c *DesignConfig) Validate() error {
	if c.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if c.Provider != "anthropic" && c.Provider != "kiro" {
		return fmt.Errorf("invalid provider %q: must be 'anthropic' or 'kiro'", c.Provider)
	}
	return nil
}

// Developer interface for asking clarifying questions.
type Developer interface {
	Respond(ctx context.Context, question string) (string, error)
}

// K8sRunnerConfig configures the K8sRunnerAgent.
type K8sRunnerConfig struct {
	WorkDir       string
	Developer     Developer
	MaxLintCycles int
}

// K8sRunnerAgent generates Kubernetes infrastructure code.
type K8sRunnerAgent struct {
	workDir        string
	developer      Developer
	maxLintCycles  int
	generatedFiles []string

	// Lint enforcement state
	lintCalled  bool
	lintPassed  bool
	pendingLint bool
	lintCycles  int
}

// Tool represents a tool definition for the agent.
type Tool struct {
	Name        string
	Description string
	InputSchema ToolInputSchema
}

// ToolInputSchema defines the JSON schema for tool parameters.
type ToolInputSchema struct {
	Properties map[string]interface{}
	Required   []string
}

// NewK8sRunnerAgent creates a new K8s-specific runner agent.
func NewK8sRunnerAgent(config K8sRunnerConfig) *K8sRunnerAgent {
	if config.WorkDir == "" {
		config.WorkDir = "."
	}
	if config.MaxLintCycles == 0 {
		config.MaxLintCycles = 3
	}

	return &K8sRunnerAgent{
		workDir:       config.WorkDir,
		developer:     config.Developer,
		maxLintCycles: config.MaxLintCycles,
	}
}

// GetTools returns the domain-specific tools for K8s generation.
func (r *K8sRunnerAgent) GetTools() []Tool {
	return []Tool{
		{
			Name:        "init_package",
			Description: "Initialize a new wetwire-k8s package directory",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Package name (directory name)",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a Go file",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path relative to work directory",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "File content",
					},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read a file's contents",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path relative to work directory",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "run_lint",
			Description: "Run the wetwire-k8s linter on the package",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Package path to lint",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "run_build",
			Description: "Build the Kubernetes YAML manifests from the package",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Package path to build",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "ask_developer",
			Description: "Ask the developer a clarifying question",
			InputSchema: ToolInputSchema{
				Properties: map[string]interface{}{
					"question": map[string]interface{}{
						"type":        "string",
						"description": "The question to ask",
					},
				},
				Required: []string{"question"},
			},
		},
	}
}

// ExecuteTool executes a tool and returns the result.
func (r *K8sRunnerAgent) ExecuteTool(ctx context.Context, name string, input json.RawMessage) string {
	var params map[string]string
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	switch name {
	case "init_package":
		return r.toolInitPackage(params["name"])
	case "write_file":
		return r.toolWriteFile(params["path"], params["content"])
	case "read_file":
		return r.toolReadFile(params["path"])
	case "run_lint":
		return r.toolRunLint(params["path"])
	case "run_build":
		return r.toolRunBuild(params["path"])
	case "ask_developer":
		answer, err := r.askDeveloper(ctx, params["question"])
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return answer
	default:
		return fmt.Sprintf("Unknown tool: %s", name)
	}
}

func (r *K8sRunnerAgent) toolInitPackage(name string) string {
	dir := filepath.Join(r.workDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}
	return fmt.Sprintf("Created package directory: %s", dir)
}

func (r *K8sRunnerAgent) toolWriteFile(path, content string) string {
	fullPath := filepath.Join(r.workDir, path)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err)
	}

	r.generatedFiles = append(r.generatedFiles, path)

	// Update lint enforcement state: code needs linting
	r.pendingLint = true
	r.lintPassed = false

	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
}

func (r *K8sRunnerAgent) toolReadFile(path string) string {
	fullPath := filepath.Join(r.workDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(content)
}

func (r *K8sRunnerAgent) toolRunLint(path string) string {
	fullPath := filepath.Join(r.workDir, path)

	// Use the internal lint package
	linter := lint.NewLinter(nil)
	result, err := linter.LintWithResult(fullPath)

	// Update lint enforcement state
	r.lintCalled = true
	r.pendingLint = false
	r.lintCycles++

	if err != nil {
		r.lintPassed = false
		return fmt.Sprintf(`{"success": false, "error": %q}`, err.Error())
	}

	// Format result as JSON
	type lintIssue struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		RuleID   string `json:"rule_id"`
	}

	type lintOutput struct {
		Success      bool        `json:"success"`
		TotalFiles   int         `json:"total_files"`
		IssueCount   int         `json:"issue_count"`
		ErrorCount   int         `json:"error_count"`
		WarningCount int         `json:"warning_count"`
		Issues       []lintIssue `json:"issues"`
	}

	output := lintOutput{
		Success:      result.ErrorCount == 0,
		TotalFiles:   result.TotalFiles,
		IssueCount:   len(result.Issues),
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		Issues:       make([]lintIssue, len(result.Issues)),
	}

	for i, issue := range result.Issues {
		output.Issues[i] = lintIssue{
			File:     issue.File,
			Line:     issue.Line,
			Column:   issue.Column,
			Message:  issue.Message,
			Severity: issue.Severity.String(),
			RuleID:   issue.Rule,
		}
	}

	r.lintPassed = output.Success

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes)
}

func (r *K8sRunnerAgent) toolRunBuild(path string) string {
	fullPath := filepath.Join(r.workDir, path)

	// Use the internal build package
	result, err := build.Build(fullPath, build.Options{
		OutputMode: build.SingleFile,
	})

	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": %q}`, err.Error())
	}

	if len(result.OrderedResources) == 0 {
		return `{"success": true, "resources": 0, "output": ""}`
	}

	// Generate YAML output
	output, err := generateOutput(result.OrderedResources, "yaml")
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": %q}`, err.Error())
	}

	return fmt.Sprintf(`{"success": true, "resources": %d, "output": %s}`,
		len(result.OrderedResources), string(output))
}

func (r *K8sRunnerAgent) askDeveloper(ctx context.Context, question string) (string, error) {
	if r.developer == nil {
		return "", fmt.Errorf("no developer configured")
	}
	return r.developer.Respond(ctx, question)
}

// GetGeneratedFiles returns the list of generated file paths.
func (r *K8sRunnerAgent) GetGeneratedFiles() []string {
	return r.generatedFiles
}

// AddGeneratedFile adds a file to the generated files list.
func (r *K8sRunnerAgent) AddGeneratedFile(path string) {
	r.generatedFiles = append(r.generatedFiles, path)
}

// PendingLint returns whether code needs linting.
func (r *K8sRunnerAgent) PendingLint() bool {
	return r.pendingLint
}

// SetPendingLint sets the pending lint state.
func (r *K8sRunnerAgent) SetPendingLint(pending bool) {
	r.pendingLint = pending
}

// LintCalled returns whether lint has been called.
func (r *K8sRunnerAgent) LintCalled() bool {
	return r.lintCalled
}

// SetLintCalled sets the lint called state.
func (r *K8sRunnerAgent) SetLintCalled(called bool) {
	r.lintCalled = called
}

// LintPassed returns whether the last lint passed.
func (r *K8sRunnerAgent) LintPassed() bool {
	return r.lintPassed
}

// SetLintPassed sets the lint passed state.
func (r *K8sRunnerAgent) SetLintPassed(passed bool) {
	r.lintPassed = passed
}

// CheckLintEnforcement checks if the agent violated lint enforcement rules.
// Returns an enforcement message if a violation occurred, empty string otherwise.
func (r *K8sRunnerAgent) CheckLintEnforcement(toolsCalled []string) string {
	wroteFile := false
	ranLint := false

	for _, tool := range toolsCalled {
		if tool == "write_file" {
			wroteFile = true
		}
		if tool == "run_lint" {
			ranLint = true
		}
	}

	// Enforcement: If write_file was called but run_lint wasn't in the same turn
	if wroteFile && !ranLint {
		return `ENFORCEMENT: You wrote a file but did not call run_lint in the same turn.
You MUST call run_lint immediately after writing code to check for issues.
Call run_lint now before proceeding.`
	}

	return ""
}

// CheckCompletionGate checks if the agent can complete.
// Returns an enforcement message if completion is not allowed.
func (r *K8sRunnerAgent) CheckCompletionGate() string {
	// Gate 1: Must have called lint at least once
	if !r.lintCalled {
		return `ENFORCEMENT: You cannot complete without running the linter.
You MUST call run_lint to validate your code before finishing.
Call run_lint now.`
	}

	// Gate 2: Code must not be pending lint (written since last lint)
	if r.pendingLint {
		return `ENFORCEMENT: You have written code since the last lint run.
You MUST call run_lint to validate your latest changes before finishing.
Call run_lint now.`
	}

	// Gate 3: Lint must have passed
	if !r.lintPassed {
		return `ENFORCEMENT: The linter found issues that have not been resolved.
You MUST fix the lint errors and run_lint again until it passes.
Review the lint output and fix the issues.`
	}

	// All gates passed
	return ""
}

// GetSystemPrompt returns the K8s-specific system prompt.
func (r *K8sRunnerAgent) GetSystemPrompt() string {
	return `You are an infrastructure code generator using the wetwire-k8s framework.
Your job is to generate Go code that defines Kubernetes resources.

The user will describe what infrastructure they need. You will:
1. Ask clarifying questions if the requirements are unclear
2. Generate Go code using the wetwire-k8s patterns
3. Run the linter and fix any issues
4. Build the Kubernetes YAML manifests

Use the wrapper pattern for all resources:

    import (
        corev1 "k8s.io/api/core/v1"
        appsv1 "k8s.io/api/apps/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    )

    var MyDeployment = &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: "my-deployment",
            Namespace: "default",
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: ptr.To[int32](3),
            // ...
        },
    }

    var MyService = &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name: "my-service",
        },
        Spec: corev1.ServiceSpec{
            Selector: MyDeployment.Spec.Selector.MatchLabels,
        },
    }

Available tools:
- init_package: Create a new package directory
- write_file: Write a Go file
- read_file: Read a file's contents
- run_lint: Run the linter on the package
- run_build: Build the Kubernetes YAML manifests
- ask_developer: Ask the developer a clarifying question

IMPORTANT: Always run_lint after writing files, and fix any issues before running build.
The lint-after-write rule is enforced - you MUST call run_lint after every write_file call.`
}

// designCommand creates the design subcommand
func designCommand() *cli.Command {
	return &cli.Command{
		Name:  "design",
		Usage: "AI-assisted Kubernetes infrastructure generation",
		Description: `Design uses AI to generate Kubernetes infrastructure code based on
natural language descriptions.

The AI agent will:
1. Ask clarifying questions if needed
2. Generate Go code using wetwire-k8s patterns
3. Run lint and fix any issues
4. Build the final YAML manifests

Examples:
  wetwire-k8s design --prompt "Create a web app with 3 replicas"
  wetwire-k8s design --provider kiro --prompt "Add a Redis cache"
  wetwire-k8s design --output-dir ./infra --prompt "Full microservice stack"`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "AI provider: anthropic or kiro",
				Value:   "anthropic",
			},
			&cli.StringFlag{
				Name:     "prompt",
				Usage:    "Natural language description of the infrastructure to generate",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory for generated code",
				Value:   ".",
			},
			&cli.StringFlag{
				Name:  "model",
				Usage: "AI model to use (provider-specific)",
				Value: "",
			},
		},
		Action: runDesign,
	}
}

// runDesign executes the design command
func runDesign(c *cli.Context) error {
	config := DesignConfig{
		Provider:  strings.ToLower(c.String("provider")),
		Prompt:    c.String("prompt"),
		OutputDir: c.String("output-dir"),
		Model:     c.String("model"),
		APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
	}

	if err := config.Validate(); err != nil {
		return err
	}

	// Resolve output directory to absolute path
	absDir, err := filepath.Abs(config.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the K8s runner agent
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir:       absDir,
		MaxLintCycles: 3,
		// Developer will be set up based on whether we're in interactive mode
	})

	// Get the writer for output
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// For now, print configuration and instructions
	// Full AI integration would require importing wetwire-core-go
	fmt.Fprintf(writer, "Design Command Configuration:\n")
	fmt.Fprintf(writer, "  Provider:   %s\n", config.Provider)
	fmt.Fprintf(writer, "  Output Dir: %s\n", absDir)
	fmt.Fprintf(writer, "  Prompt:     %s\n", config.Prompt)
	fmt.Fprintf(writer, "\n")

	if config.Provider == "kiro" {
		fmt.Fprintf(writer, "Kiro provider selected. Run with:\n")
		fmt.Fprintf(writer, "  kiro agent wetwire-k8s-runner\n")
		fmt.Fprintf(writer, "\n")
		return nil
	}

	// Anthropic provider
	if config.APIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required for anthropic provider")
	}

	// Display system prompt and tools for debugging/transparency
	fmt.Fprintf(writer, "System Prompt:\n%s\n\n", agent.GetSystemPrompt())
	fmt.Fprintf(writer, "Available Tools:\n")
	for _, tool := range agent.GetTools() {
		fmt.Fprintf(writer, "  - %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "Note: Full AI agent loop requires wetwire-core-go integration.\n")
	fmt.Fprintf(writer, "To use the agent infrastructure, add wetwire-core-go as a dependency.\n")

	return nil
}
