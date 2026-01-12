package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/agent/personas"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/agent/scoring"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/urfave/cli/v2"
)

// Supported providers for LLM testing
var supportedProviders = []string{"anthropic", "openai", "mock"}

// testCommand creates the test subcommand
func testCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "Run persona-based testing for Kubernetes manifest generation",
		ArgsUsage: "",
		Description: `Test runs AI persona-based testing sessions to evaluate the quality
of Kubernetes manifest generation.

Personas simulate different types of users (beginner, expert, terse, etc.)
interacting with the tool. Each session is scored on 5 dimensions:
  - Completeness: Were all required resources generated?
  - Lint Quality: Did the code pass linting?
  - Code Quality: Does the code follow idiomatic patterns?
  - Output Validity: Is the generated manifest valid?
  - Question Efficiency: Did the agent ask appropriate questions?

Examples:
  wetwire-k8s test --persona beginner                    # Test with beginner persona
  wetwire-k8s test --all-personas                        # Test with all personas
  wetwire-k8s test --persona expert --scenario deploy-app # Custom scenario
  wetwire-k8s test --persona terse --provider anthropic  # Use specific provider
  wetwire-k8s test --list-personas                       # List available personas`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "persona",
				Aliases: []string{"p"},
				Usage:   "Persona to use for testing (beginner, intermediate, expert, terse, verbose)",
			},
			&cli.BoolFlag{
				Name:  "all-personas",
				Usage: "Run tests with all available personas",
			},
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"P"},
				Usage:   "LLM provider to use (anthropic, openai, mock)",
				Value:   "anthropic",
			},
			&cli.StringFlag{
				Name:    "scenario",
				Aliases: []string{"s"},
				Usage:   "Test scenario to run",
				Value:   "default",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output directory for results",
				Value:   "./test-results",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be tested without running",
			},
			&cli.BoolFlag{
				Name:  "mock",
				Usage: "Use mock LLM responses (for testing)",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show verbose output",
			},
			&cli.BoolFlag{
				Name:  "list-personas",
				Usage: "List available personas and exit",
			},
		},
		Action: runTest,
	}
}

// runTest executes the test command
func runTest(c *cli.Context) error {
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Handle --list-personas
	if c.Bool("list-personas") {
		return listPersonas(writer)
	}

	// Validate that either --persona or --all-personas is specified
	personaName := c.String("persona")
	allPersonas := c.Bool("all-personas")

	if personaName == "" && !allPersonas {
		return fmt.Errorf("either --persona or --all-personas must be specified")
	}

	// Validate provider
	provider := c.String("provider")
	if c.Bool("mock") {
		provider = "mock"
	}
	if !isValidProvider(provider) {
		return fmt.Errorf("invalid provider %q: must be one of: %s", provider, strings.Join(supportedProviders, ", "))
	}

	// Get scenario
	scenario := c.String("scenario")

	// Get output directory
	outputDir := c.String("output")

	// Get options
	dryRun := c.Bool("dry-run")
	verbose := c.Bool("verbose")

	// Build list of personas to test
	var personasToTest []personas.Persona
	if allPersonas {
		personasToTest = personas.All()
	} else {
		p, err := personas.Get(personaName)
		if err != nil {
			return err
		}
		personasToTest = []personas.Persona{p}
	}

	// Run tests for each persona
	for _, persona := range personasToTest {
		err := runPersonaTest(c, writer, persona, provider, scenario, outputDir, dryRun, verbose)
		if err != nil {
			return fmt.Errorf("test failed for persona %s: %w", persona.Name, err)
		}
	}

	return nil
}

// listPersonas displays available personas
func listPersonas(writer io.Writer) error {
	fmt.Fprintln(writer, "Available Personas:")
	fmt.Fprintln(writer, "")

	for _, p := range personas.All() {
		fmt.Fprintf(writer, "  %s\n", p.Name)
		fmt.Fprintf(writer, "    Description: %s\n", p.Description)
		fmt.Fprintf(writer, "    Traits: %s\n", strings.Join(p.Traits, ", "))
		fmt.Fprintln(writer, "")
	}

	return nil
}

// isValidProvider checks if the provider is supported
func isValidProvider(provider string) bool {
	for _, p := range supportedProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// runPersonaTest runs a test session for a single persona
func runPersonaTest(c *cli.Context, writer io.Writer, persona personas.Persona, provider, scenario, outputDir string, dryRun, verbose bool) error {
	if dryRun {
		return runDryRun(writer, persona, provider, scenario, outputDir, verbose)
	}

	// Create output directory
	personaDir := filepath.Join(outputDir, persona.Name)
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create session
	session := results.NewSession(persona.Name, scenario)
	session.InitialPrompt = generateInitialPrompt(scenario)

	// Run the test based on provider
	var err error
	if provider == "mock" {
		err = runMockTest(session, persona, scenario)
	} else {
		// Use real LLM provider
		ctx := c.Context
		if ctx == nil {
			ctx = context.Background()
		}
		err = runRealTest(ctx, writer, session, persona, scenario, personaDir, provider)
	}

	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Calculate score
	score := calculateScore(session, persona, scenario)
	session.Score = score

	// Complete session
	session.Complete()

	// Write results
	resultsWriter := results.NewWriter(outputDir)
	if err := resultsWriter.Write(session); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	// Print summary
	fmt.Fprintf(writer, "Test completed for persona: %s\n", persona.Name)
	fmt.Fprintf(writer, "Score: %d/15 (%s)\n", score.Total(), score.Threshold())
	fmt.Fprintf(writer, "Results written to: %s\n", personaDir)

	return nil
}

// runDryRun displays what would be tested
func runDryRun(writer io.Writer, persona personas.Persona, provider, scenario, outputDir string, verbose bool) error {
	fmt.Fprintf(writer, "=== dry-run: Test Configuration ===\n\n")
	fmt.Fprintf(writer, "Persona: %s\n", persona.Name)
	fmt.Fprintf(writer, "Provider: %s\n", provider)
	fmt.Fprintf(writer, "Scenario: %s\n", scenario)
	fmt.Fprintf(writer, "Output Directory: %s\n", outputDir)
	fmt.Fprintln(writer, "")

	if verbose {
		fmt.Fprintf(writer, "Persona Details:\n")
		fmt.Fprintf(writer, "  Description: %s\n", persona.Description)
		fmt.Fprintf(writer, "  Traits: %s\n", strings.Join(persona.Traits, ", "))
		fmt.Fprintf(writer, "  Expected Behavior: %s\n", persona.ExpectedBehavior)
		fmt.Fprintln(writer, "")
	}

	fmt.Fprintf(writer, "Would score on dimensions:\n")
	fmt.Fprintln(writer, "  - Completeness: Were all required resources generated?")
	fmt.Fprintln(writer, "  - Lint Quality: Did the code pass linting?")
	fmt.Fprintln(writer, "  - Code Quality: Does the code follow idiomatic patterns?")
	fmt.Fprintln(writer, "  - Output Validity: Is the generated manifest valid?")
	fmt.Fprintln(writer, "  - Question Efficiency: Did the agent ask appropriate questions?")
	fmt.Fprintln(writer, "")

	fmt.Fprintf(writer, "Would generate files in: %s/%s/\n", outputDir, persona.Name)
	fmt.Fprintln(writer, "  - RESULTS.md")
	fmt.Fprintln(writer, "  - session.json")
	fmt.Fprintln(writer, "  - score.json")

	return nil
}

// runMockTest runs a test with mocked LLM responses
func runMockTest(session *results.Session, persona personas.Persona, scenario string) error {
	// Add mock conversation
	session.AddMessage("developer", fmt.Sprintf("I need help with: %s", scenario))
	session.AddMessage("runner", "I'll help you with that. Let me generate the Kubernetes manifests.")

	// Add mock questions for non-expert personas
	if persona.Name != "expert" && persona.Name != "terse" {
		session.AddQuestion("What namespace should I use?", "default")
		session.AddQuestion("How many replicas do you need?", "3")
	}

	// Add mock lint cycle
	session.AddLintCycle([]string{}, 0, true)

	// Add mock generated files
	session.GeneratedFiles = []string{
		"deployment.yaml",
		"service.yaml",
	}

	return nil
}

// generateInitialPrompt creates the initial prompt for a scenario
func generateInitialPrompt(scenario string) string {
	switch scenario {
	case "deploy-nginx":
		return "Create a Kubernetes deployment for nginx with a service"
	case "deploy-app":
		return "Create a complete application deployment with configmap and secrets"
	default:
		return "Generate Kubernetes manifests for a sample application"
	}
}

// calculateScore calculates the score for a session
func calculateScore(session *results.Session, persona personas.Persona, scenario string) *scoring.Score {
	score := scoring.NewScore(persona.Name, scenario)

	// Score completeness based on generated files
	expectedResources := 2 // For mock: deployment + service
	actualResources := len(session.GeneratedFiles)
	rating, notes := scoring.ScoreCompleteness(expectedResources, actualResources)
	score.Completeness.Rating = rating
	score.Completeness.Notes = notes

	// Score lint quality
	lintPassed := len(session.LintCycles) > 0 && session.LintCycles[len(session.LintCycles)-1].Passed
	lintCycles := len(session.LintCycles)
	rating, notes = scoring.ScoreLintQuality(lintCycles, lintPassed)
	score.LintQuality.Rating = rating
	score.LintQuality.Notes = notes

	// Score code quality (mock: no issues)
	rating, notes = scoring.ScoreCodeQuality([]string{})
	score.CodeQuality.Rating = rating
	score.CodeQuality.Notes = notes

	// Score output validity (mock: valid)
	rating, notes = scoring.ScoreOutputValidity(0, 0)
	score.OutputValidity.Rating = rating
	score.OutputValidity.Notes = notes

	// Score question efficiency
	questionCount := len(session.Questions)
	rating, notes = scoring.ScoreQuestionEfficiency(questionCount)
	score.QuestionEfficiency.Rating = rating
	score.QuestionEfficiency.Notes = notes

	// Store metadata
	score.LintCycles = lintCycles
	score.QuestionCount = questionCount

	return score
}

// runRealTest runs a test with a real LLM provider.
func runRealTest(ctx context.Context, writer io.Writer, session *results.Session, persona personas.Persona, scenario, workDir, providerName string) error {
	// Get API key based on provider
	var apiKey string
	switch providerName {
	case "anthropic":
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required for anthropic provider")
		}
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is required for openai provider")
		}
		// OpenAI not yet implemented
		return fmt.Errorf("openai provider not yet implemented")
	default:
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Create provider
	provider, err := NewAnthropicProvider(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Create the K8s runner agent
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir:       workDir,
		MaxLintCycles: 3,
		Developer:     &PersonaDeveloper{persona: persona, provider: provider},
	})

	// Record session start
	session.AddMessage("developer", session.InitialPrompt)

	// Run the agent loop
	err = runTestAgentLoop(ctx, writer, session, agent, provider, persona)
	if err != nil {
		return fmt.Errorf("agent loop failed: %w", err)
	}

	// Record generated files from agent
	session.GeneratedFiles = agent.GetGeneratedFiles()

	// Record lint state
	if agent.LintCalled() {
		session.AddLintCycle([]string{}, 0, agent.LintPassed())
	}

	return nil
}

// PersonaDeveloper implements the Developer interface using an AI with a persona.
type PersonaDeveloper struct {
	persona  personas.Persona
	provider providers.Provider
}

// Respond generates a response using the persona's system prompt.
func (d *PersonaDeveloper) Respond(ctx context.Context, question string) (string, error) {
	req := providers.MessageRequest{
		System:    d.persona.SystemPrompt,
		Messages:  []providers.Message{providers.NewUserMessage(question)},
		MaxTokens: 1024,
	}

	resp, err := d.provider.CreateMessage(ctx, req)
	if err != nil {
		return "", err
	}

	// Extract text from response
	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text in response")
}

// runTestAgentLoop runs the agent loop for testing, recording interactions in the session.
func runTestAgentLoop(ctx context.Context, writer io.Writer, session *results.Session, agent *K8sRunnerAgent, provider providers.Provider, persona personas.Persona) error {
	// Build the initial message
	messages := []providers.Message{
		providers.NewUserMessage(session.InitialPrompt),
	}

	// Maximum iterations to prevent infinite loops
	maxIterations := 30
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Create the request
		req := providers.MessageRequest{
			System:    agent.GetSystemPrompt(),
			Messages:  messages,
			Tools:     agent.GetProviderTools(),
			MaxTokens: 4096,
		}

		// Call the provider
		resp, err := provider.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("provider call failed: %w", err)
		}

		// Process the response
		var toolsCalled []string
		var toolResults []providers.ContentBlock
		var responseText string

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				responseText += block.Text
				fmt.Fprintf(writer, "[Runner]: %s\n", block.Text)
			case "tool_use":
				toolsCalled = append(toolsCalled, block.Name)
				fmt.Fprintf(writer, "[Tool]: %s\n", block.Name)

				// Handle ask_developer specially to record questions
				if block.Name == "ask_developer" {
					var params map[string]string
					if err := json.Unmarshal(block.Input, &params); err == nil {
						question := params["question"]
						// Get response from persona
						answer := agent.ExecuteTool(ctx, block.Name, block.Input)
						session.AddQuestion(question, answer)
						toolResults = append(toolResults, providers.NewToolResult(block.ID, answer, false))
						continue
					}
				}

				// Execute other tools
				result := agent.ExecuteTool(ctx, block.Name, block.Input)
				toolResults = append(toolResults, providers.NewToolResult(block.ID, result, false))
			}
		}

		// Record assistant message
		if responseText != "" {
			session.AddMessage("runner", responseText)
		}

		// If the model stopped because it wants to use tools, continue the loop
		if resp.StopReason == providers.StopReasonToolUse {
			messages = append(messages, providers.NewAssistantMessage(resp.Content))
			messages = append(messages, providers.NewToolResultMessage(toolResults))
			continue
		}

		// Model finished
		if resp.StopReason == providers.StopReasonEndTurn {
			// Check completion gate
			gate := agent.CheckCompletionGate()
			if gate != "" {
				messages = append(messages, providers.NewAssistantMessage(resp.Content))
				messages = append(messages, providers.NewUserMessage(gate))
				continue
			}
			return nil
		}

		// Handle max tokens
		if resp.StopReason == providers.StopReasonMaxTokens {
			messages = append(messages, providers.NewAssistantMessage(resp.Content))
			messages = append(messages, providers.NewUserMessage("Please continue."))
			continue
		}

		return nil
	}

	return fmt.Errorf("agent loop exceeded maximum iterations (%d)", maxIterations)
}
