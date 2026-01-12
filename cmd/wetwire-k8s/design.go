// Command design provides AI-assisted infrastructure design.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/agent/orchestrator"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/urfave/cli/v2"
)

// K8sDomain returns the Kubernetes domain configuration for the RunnerAgent.
func K8sDomain() agents.DomainConfig {
	return agents.DomainConfig{
		Name:       "k8s",
		CLICommand: "wetwire-k8s",
		SystemPrompt: `You are a Kubernetes manifest generator using the wetwire-k8s framework.
Your job is to generate Go code that defines Kubernetes resources.

Use the wrapper pattern for all resources:
    var MyDeployment = apps.Deployment{
        Name: "my-app",
        Replicas: 3,
    }

Available tools: init_package, write_file, read_file, run_lint, run_build, ask_developer

Always run_lint after writing files, and fix any issues before running build.`,
		OutputFormat: "Kubernetes YAML",
	}
}

// designCommand creates the design subcommand.
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
  wetwire-k8s design --output-dir ./infra --prompt "Full microservice stack"`,
		Flags: []cli.Flag{
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
			&cli.IntFlag{
				Name:  "max-lint-cycles",
				Usage: "Maximum lint/fix cycles",
				Value: 3,
			},
			&cli.BoolFlag{
				Name:  "stream",
				Usage: "Stream AI responses",
				Value: true,
			},
		},
		Action: runDesign,
	}
}

// runDesign executes the design command using the core-go RunnerAgent.
func runDesign(c *cli.Context) error {
	prompt := c.String("prompt")
	outputDir := c.String("output-dir")
	maxLintCycles := c.Int("max-lint-cycles")
	stream := c.Bool("stream")

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

	// Create session for tracking
	session := results.NewSession("human", "design")

	// Create human developer (reads from stdin)
	reader := bufio.NewReader(os.Stdin)
	developer := orchestrator.NewHumanDeveloper(func() (string, error) {
		return reader.ReadString('\n')
	})

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

	fmt.Println("Starting AI-assisted design session...")
	fmt.Println("The AI will ask questions and generate infrastructure code.")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// Run the agent
	if err := runner.Run(ctx, prompt); err != nil {
		return fmt.Errorf("design session failed: %w", err)
	}

	// Print summary
	fmt.Println("\n--- Session Summary ---")
	fmt.Printf("Generated files: %d\n", len(runner.GetGeneratedFiles()))
	for _, f := range runner.GetGeneratedFiles() {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Printf("Lint cycles: %d\n", runner.GetLintCycles())
	fmt.Printf("Lint passed: %v\n", runner.LintPassed())

	return nil
}
