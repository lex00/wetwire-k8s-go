// Command test runs automated persona-based testing.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/agent/orchestrator"
	"github.com/lex00/wetwire-core-go/agent/personas"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/urfave/cli/v2"
)

// testCommand creates the test subcommand.
func testCommand() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Run automated persona-based testing",
		Description: `Run automated testing with AI personas to evaluate code generation quality.

Available personas:
  - beginner: New to Kubernetes, asks many clarifying questions
  - intermediate: Familiar with Kubernetes basics, asks targeted questions
  - expert: Deep Kubernetes knowledge, asks advanced questions
  - terse: Gives minimal responses
  - verbose: Provides detailed context

Example:
  wetwire-k8s test --persona beginner --prompt "Create a deployment with 3 replicas"
  wetwire-k8s test --all-personas --prompt "Create an nginx deployment"`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prompt",
				Usage:    "Natural language description of infrastructure to generate",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "persona",
				Aliases: []string{"p"},
				Usage:   "Persona to use (beginner, intermediate, expert, terse, verbose)",
				Value:   "intermediate",
			},
			&cli.BoolFlag{
				Name:  "all-personas",
				Usage: "Run test with all personas",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory for generated files",
				Value:   ".",
			},
			&cli.StringFlag{
				Name:  "scenario",
				Usage: "Scenario name for tracking",
				Value: "default",
			},
			&cli.IntFlag{
				Name:  "max-lint-cycles",
				Usage: "Maximum lint/fix cycles",
				Value: 3,
			},
			&cli.BoolFlag{
				Name:  "stream",
				Usage: "Stream AI responses",
			},
		},
		Action: runTest,
	}
}

// runTest executes the test command.
func runTest(c *cli.Context) error {
	prompt := c.String("prompt")
	outputDir := c.String("output-dir")
	personaName := c.String("persona")
	scenario := c.String("scenario")
	maxLintCycles := c.Int("max-lint-cycles")
	stream := c.Bool("stream")
	allPersonas := c.Bool("all-personas")

	if allPersonas {
		return runTestAllPersonas(c, prompt, outputDir, scenario, maxLintCycles, stream)
	}

	return runTestSinglePersona(c, prompt, outputDir, personaName, scenario, maxLintCycles, stream)
}

// runTestAllPersonas runs the test with all available personas.
func runTestAllPersonas(c *cli.Context, prompt, outputDir, scenario string, maxLintCycles int, stream bool) error {
	personaNames := personas.Names()
	var failed []string

	fmt.Printf("Running tests with all %d personas\n\n", len(personaNames))

	for _, personaName := range personaNames {
		// Create persona-specific output directory
		personaOutputDir := fmt.Sprintf("%s/%s", outputDir, personaName)

		fmt.Printf("=== Running persona: %s ===\n", personaName)

		err := runTestSinglePersona(c, prompt, personaOutputDir, personaName, scenario, maxLintCycles, stream)
		if err != nil {
			fmt.Printf("Persona %s: FAILED - %v\n\n", personaName, err)
			failed = append(failed, personaName)
		} else {
			fmt.Printf("Persona %s: PASSED\n\n", personaName)
		}
	}

	// Print summary
	fmt.Println("\n=== All Personas Summary ===")
	fmt.Printf("Total: %d\n", len(personaNames))
	fmt.Printf("Passed: %d\n", len(personaNames)-len(failed))
	fmt.Printf("Failed: %d\n", len(failed))
	if len(failed) > 0 {
		fmt.Printf("Failed personas: %v\n", failed)
		return fmt.Errorf("%d personas failed", len(failed))
	}

	return nil
}

// runTestSinglePersona runs a test with a single persona.
func runTestSinglePersona(c *cli.Context, prompt, outputDir, personaName, scenario string, maxLintCycles int, stream bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, cleaning up...")
		cancel()
	}()

	// Get persona
	persona, err := personas.Get(personaName)
	if err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}

	// Create session for tracking
	session := results.NewSession(personaName, scenario)

	// Create AI developer with persona
	responder := agents.CreateDeveloperResponder("")
	developer := orchestrator.NewAIDeveloper(persona, responder)

	// Create stream handler if streaming enabled
	var streamHandler agents.StreamHandler
	if stream {
		streamHandler = func(text string) {
			fmt.Print(text)
		}
	}

	// Create runner agent with K8s domain config
	runner, err := agents.NewRunnerAgent(agents.RunnerConfig{
		Domain:        K8sDomain(),
		WorkDir:       outputDir,
		MaxLintCycles: maxLintCycles,
		Session:       session,
		Developer:     developer,
		StreamHandler: streamHandler,
	})
	if err != nil {
		return fmt.Errorf("creating runner: %w", err)
	}

	fmt.Printf("Running test with persona '%s' and scenario '%s'\n", personaName, scenario)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Run the agent
	if err := runner.Run(ctx, prompt); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	// Complete session
	session.Complete()

	// Write results
	writer := results.NewWriter(outputDir)
	if err := writer.Write(session); err != nil {
		fmt.Printf("Warning: failed to write results: %v\n", err)
	} else {
		fmt.Printf("\nResults written to: %s\n", outputDir)
	}

	// Print summary
	fmt.Println("\n--- Test Summary ---")
	fmt.Printf("Persona: %s\n", personaName)
	fmt.Printf("Scenario: %s\n", scenario)
	fmt.Printf("Generated files: %d\n", len(runner.GetGeneratedFiles()))
	for _, f := range runner.GetGeneratedFiles() {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Printf("Lint cycles: %d\n", runner.GetLintCycles())
	fmt.Printf("Lint passed: %v\n", runner.LintPassed())
	fmt.Printf("Questions asked: %d\n", len(session.Questions))

	return nil
}
