package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/urfave/cli/v2"
)

// lintCommand creates the lint subcommand
func lintCommand() *cli.Command {
	return &cli.Command{
		Name:      "lint",
		Usage:     "Lint Kubernetes resource declarations in Go code",
		ArgsUsage: "[PATH]",
		Description: `Lint parses Go source files and applies best practice rules
to Kubernetes resource declarations.

If PATH is not specified, the current directory is used.

Examples:
  wetwire-k8s lint                      # Lint current directory
  wetwire-k8s lint ./k8s                # Lint specific directory
  wetwire-k8s lint --format json        # Output as JSON
  wetwire-k8s lint --format github      # Output for GitHub Actions
  wetwire-k8s lint --severity error     # Show only errors
  wetwire-k8s lint --disable WK8006     # Disable specific rules
  wetwire-k8s lint --fix                # Auto-fix violations where possible`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "Auto-fix violations where possible",
				Value: false,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: text, json, or github",
				Value:   "text",
			},
			&cli.StringFlag{
				Name:    "severity",
				Aliases: []string{"s"},
				Usage:   "Minimum severity to report: error, warning, or info",
				Value:   "info",
			},
			&cli.StringFlag{
				Name:    "disable",
				Aliases: []string{"d"},
				Usage:   "Disable specific rules (comma-separated, e.g., WK8001,WK8002)",
				Value:   "",
			},
		},
		Action: runLint,
	}
}

// runLint executes the lint command
func runLint(c *cli.Context) error {
	// Determine source path
	sourcePath := c.Args().First()
	if sourcePath == "" {
		sourcePath = "."
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate source path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("source path does not exist: %s", absPath)
	}

	// Validate format
	format := strings.ToLower(c.String("format"))
	if format != "text" && format != "json" && format != "github" {
		return fmt.Errorf("invalid format %q: must be 'text', 'json', or 'github'", format)
	}

	// Parse severity
	severityStr := strings.ToLower(c.String("severity"))
	var minSeverity lint.Severity
	switch severityStr {
	case "error":
		minSeverity = lint.SeverityError
	case "warning":
		minSeverity = lint.SeverityWarning
	case "info":
		minSeverity = lint.SeverityInfo
	default:
		return fmt.Errorf("invalid severity %q: must be 'error', 'warning', or 'info'", severityStr)
	}

	// Parse disabled rules
	var disabledRules []string
	disableStr := c.String("disable")
	if disableStr != "" {
		disabledRules = strings.Split(disableStr, ",")
		// Trim whitespace from each rule
		for i, rule := range disabledRules {
			disabledRules[i] = strings.TrimSpace(rule)
		}
	}

	// Create linter configuration
	config := &lint.Config{
		DisabledRules: disabledRules,
		MinSeverity:   minSeverity,
	}

	// Create linter
	linter := lint.NewLinter(config)

	// Run lint
	result, err := linter.LintWithResult(absPath)
	if err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}

	// Get the writer for output
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Format and output results
	outputFormat := lint.OutputFormat(format)
	formatter := lint.NewFormatter(outputFormat)
	if err := formatter.Format(result, writer); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Handle auto-fix if requested
	fix := c.Bool("fix")
	if fix {
		// Auto-fix is not fully implemented yet
		// This is a placeholder for future implementation
		// For now, just inform the user
		errWriter := c.App.ErrWriter
		if errWriter == nil {
			errWriter = os.Stderr
		}
		if len(result.Issues) > 0 {
			fmt.Fprintln(errWriter, "\nNote: Auto-fix is not yet implemented for all rules.")
		}
	}

	// Exit with error if there are violations
	if result.ErrorCount > 0 {
		return fmt.Errorf("found %d error(s), %d warning(s), %d info", result.ErrorCount, result.WarningCount, result.InfoCount)
	}

	return nil
}
