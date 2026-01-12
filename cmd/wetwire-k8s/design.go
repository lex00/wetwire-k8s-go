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
	"github.com/lex00/wetwire-k8s-go/internal/kiro"
	"github.com/spf13/cobra"
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

// newDesignCmd creates the design subcommand.
func newDesignCmd() *cobra.Command {
	var prompt string
	var outputDir string
	var maxLintCycles int
	var stream bool
	var provider string

	cmd := &cobra.Command{
		Use:   "design",
		Short: "AI-assisted Kubernetes infrastructure generation",
		Long: `Design uses AI to generate Kubernetes infrastructure code based on
natural language descriptions.

The AI agent will:
1. Ask clarifying questions if needed
2. Generate Go code using wetwire-k8s patterns
3. Run lint and fix any issues
4. Build the final YAML manifests

Examples:
  wetwire-k8s design --prompt "Create a web app with 3 replicas"
  wetwire-k8s design --output-dir ./infra --prompt "Full microservice stack"
  wetwire-k8s design --provider kiro --prompt "Create nginx deployment"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if prompt == "" {
				return fmt.Errorf("--prompt flag is required")
			}

			// Handle kiro provider
			if provider == "kiro" {
				return runDesignKiro(prompt)
			}

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
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Natural language description of the infrastructure to generate")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "Output directory for generated code")
	cmd.Flags().IntVar(&maxLintCycles, "max-lint-cycles", 3, "Maximum lint/fix cycles")
	cmd.Flags().BoolVar(&stream, "stream", true, "Stream AI responses")
	cmd.Flags().StringVar(&provider, "provider", "core", "AI provider to use (core, kiro)")
	_ = cmd.MarkFlagRequired("prompt")

	return cmd
}

// runDesignKiro runs the design command using the Kiro provider.
func runDesignKiro(prompt string) error {
	fmt.Println("Launching Kiro design session...")
	return kiro.LaunchChat(kiro.AgentName, prompt)
}
