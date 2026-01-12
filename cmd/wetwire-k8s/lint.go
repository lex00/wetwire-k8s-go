package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/spf13/cobra"
)

// newLintCmd creates the lint subcommand
func newLintCmd() *cobra.Command {
	var fix bool
	var format string
	var severity string
	var disable string

	cmd := &cobra.Command{
		Use:   "lint [PATH]",
		Short: "Lint Kubernetes resource declarations in Go code",
		Long: `Lint parses Go source files and applies best practice rules
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
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine source path
			sourcePath := "."
			if len(args) > 0 {
				sourcePath = args[0]
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
			format = strings.ToLower(format)
			if format != "text" && format != "json" && format != "github" {
				return fmt.Errorf("invalid format %q: must be 'text', 'json', or 'github'", format)
			}

			// Parse severity
			severityStr := strings.ToLower(severity)
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
			if disable != "" {
				disabledRules = strings.Split(disable, ",")
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
			writer := cmd.OutOrStdout()

			// Format and output results
			outputFormat := lint.OutputFormat(format)
			formatter := lint.NewFormatter(outputFormat)
			if err := formatter.Format(result, writer); err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			// Handle auto-fix if requested
			if fix && len(result.Issues) > 0 {
				errWriter := cmd.ErrOrStderr()

				// Check if there are any fixable issues
				hasFixable := false
				for _, issue := range result.Issues {
					if isFixableIssue(issue.Rule) {
						hasFixable = true
						break
					}
				}

				if hasFixable {
					fmt.Fprintln(errWriter, "\nApplying auto-fixes...")

					// Create fixer and apply fixes
					fixer := lint.NewFixer(config)
					fixResults, err := fixer.FixDirectory(absPath)
					if err != nil {
						fmt.Fprintf(errWriter, "Error applying fixes: %v\n", err)
					} else {
						// Report fixed issues
						fixedCount := 0
						for _, fr := range fixResults {
							if fr.Fixed {
								fixedCount++
								fmt.Fprintf(errWriter, "  Fixed: [%s] %s\n", fr.Rule, fr.Description)
							} else if fr.Error != nil {
								fmt.Fprintf(errWriter, "  Error: %s: %v\n", fr.File, fr.Error)
							}
						}
						if fixedCount > 0 {
							fmt.Fprintf(errWriter, "\nFixed %d issue(s). Re-run lint to verify.\n", fixedCount)
						}
					}
				} else {
					fmt.Fprintln(errWriter, "\nNote: No auto-fixable issues found. Fixable rules: WK8002, WK8105")
				}
			}

			// Exit with error if there are violations
			if result.ErrorCount > 0 {
				return fmt.Errorf("found %d error(s), %d warning(s), %d info", result.ErrorCount, result.WarningCount, result.InfoCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Auto-fix violations where possible")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json, or github")
	cmd.Flags().StringVarP(&severity, "severity", "s", "info", "Minimum severity to report: error, warning, or info")
	cmd.Flags().StringVarP(&disable, "disable", "d", "", "Disable specific rules (comma-separated, e.g., WK8001,WK8002)")

	return cmd
}

// isFixableIssue returns true if the rule supports auto-fix.
func isFixableIssue(ruleID string) bool {
	fixableRules := map[string]bool{
		"WK8105": true, // ImagePullPolicy
		"WK8002": true, // Deeply nested structures
	}
	return fixableRules[ruleID]
}
