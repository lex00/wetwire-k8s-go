package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/roundtrip"
	"github.com/spf13/cobra"
)

// ANSI color codes for diff output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
)

// newDiffCmd creates the diff subcommand
func newDiffCmd() *cobra.Command {
	var against string
	var semantic bool
	var useColor bool

	cmd := &cobra.Command{
		Use:   "diff [PATH]",
		Short: "Compare generated manifests against existing manifests",
		Long: `Diff compares the generated Kubernetes manifests from Go code
against an existing manifest file.

If PATH is not specified, the current directory is used.

The --against flag is required and specifies the existing manifest to compare against.

Examples:
  wetwire-k8s diff ./k8s --against manifest.yaml           # Compare generated output vs existing
  wetwire-k8s diff --against manifest.yaml                 # Compare from current directory
  wetwire-k8s diff ./k8s --against manifest.yaml --semantic # Use semantic comparison
  wetwire-k8s diff ./k8s --against manifest.yaml --color   # Colorized output`,
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

			// Get the --against manifest file
			if against == "" {
				return fmt.Errorf("--against flag is required")
			}

			// Resolve against path
			againstPath, err := filepath.Abs(against)
			if err != nil {
				return fmt.Errorf("failed to resolve against path: %w", err)
			}

			// Read the existing manifest
			existingData, err := os.ReadFile(againstPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest file: %w", err)
			}

			// Run the build pipeline
			result, err := build.Build(absPath, build.Options{
				OutputMode: build.SingleFile,
			})
			if err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			// Generate YAML output from build
			var generatedData []byte
			if len(result.OrderedResources) > 0 {
				generatedData, err = generateOutput(result.OrderedResources, "yaml")
				if err != nil {
					return fmt.Errorf("failed to generate output: %w", err)
				}
			}

			// Get output writer
			writer := cmd.OutOrStdout()

			if semantic {
				return runSemanticDiff(writer, existingData, generatedData, useColor)
			}

			return runTextDiff(writer, existingData, generatedData, useColor)
		},
	}

	cmd.Flags().StringVarP(&against, "against", "a", "", "Existing manifest file to compare against (required)")
	cmd.Flags().BoolVar(&semantic, "semantic", false, "Use semantic comparison (ignores ordering and whitespace)")
	cmd.Flags().BoolVar(&useColor, "color", false, "Enable colorized output")
	_ = cmd.MarkFlagRequired("against")

	return cmd
}

// runSemanticDiff performs semantic comparison of YAML documents
func runSemanticDiff(writer io.Writer, existing, generated []byte, useColor bool) error {
	// Parse both YAML documents
	existingDocs, err := roundtrip.ParseMultiDocYAML(existing)
	if err != nil {
		return fmt.Errorf("failed to parse existing manifest: %w", err)
	}

	generatedDocs, err := roundtrip.ParseMultiDocYAML(generated)
	if err != nil {
		return fmt.Errorf("failed to parse generated manifest: %w", err)
	}

	// Compare documents semantically
	equivalent, differences := roundtrip.Compare(existingDocs, generatedDocs, true)

	if equivalent {
		fmt.Fprintln(writer, "No differences found")
		return nil
	}

	// Format and output differences
	formatSemanticDiff(writer, differences, useColor)
	return nil
}

// formatSemanticDiff formats semantic differences for output
func formatSemanticDiff(writer io.Writer, differences []roundtrip.Difference, useColor bool) {
	fmt.Fprintf(writer, "Found %d difference(s):\n\n", len(differences))

	for _, diff := range differences {
		switch diff.Type {
		case roundtrip.DiffTypeMissing:
			if useColor {
				fmt.Fprintf(writer, "%s- MISSING:%s %s\n", colorRed, colorReset, diff.Path)
				fmt.Fprintf(writer, "  was: %v\n", diff.Original)
			} else {
				fmt.Fprintf(writer, "- MISSING: %s\n", diff.Path)
				fmt.Fprintf(writer, "  was: %v\n", diff.Original)
			}
		case roundtrip.DiffTypeAdded:
			if useColor {
				fmt.Fprintf(writer, "%s+ ADDED:%s %s\n", colorGreen, colorReset, diff.Path)
				fmt.Fprintf(writer, "  now: %v\n", diff.Result)
			} else {
				fmt.Fprintf(writer, "+ ADDED: %s\n", diff.Path)
				fmt.Fprintf(writer, "  now: %v\n", diff.Result)
			}
		case roundtrip.DiffTypeModified:
			if useColor {
				fmt.Fprintf(writer, "%s~ MODIFIED:%s %s\n", colorYellow, colorReset, diff.Path)
				fmt.Fprintf(writer, "  %swas:%s %v\n", colorRed, colorReset, diff.Original)
				fmt.Fprintf(writer, "  %snow:%s %v\n", colorGreen, colorReset, diff.Result)
			} else {
				fmt.Fprintf(writer, "~ MODIFIED: %s\n", diff.Path)
				fmt.Fprintf(writer, "  was: %v\n", diff.Original)
				fmt.Fprintf(writer, "  now: %v\n", diff.Result)
			}
		}
		fmt.Fprintln(writer)
	}
}

// runTextDiff performs line-by-line text comparison
func runTextDiff(writer io.Writer, existing, generated []byte, useColor bool) error {
	existingLines := strings.Split(string(existing), "\n")
	generatedLines := strings.Split(string(generated), "\n")

	// Simple line-by-line diff using LCS algorithm
	diff := computeLineDiff(existingLines, generatedLines)

	if len(diff) == 0 {
		fmt.Fprintln(writer, "No differences found")
		return nil
	}

	// Output diff
	for _, line := range diff {
		if useColor {
			switch {
			case strings.HasPrefix(line, "-"):
				fmt.Fprintf(writer, "%s%s%s\n", colorRed, line, colorReset)
			case strings.HasPrefix(line, "+"):
				fmt.Fprintf(writer, "%s%s%s\n", colorGreen, line, colorReset)
			case strings.HasPrefix(line, "@"):
				fmt.Fprintf(writer, "%s%s%s\n", colorCyan, line, colorReset)
			default:
				fmt.Fprintln(writer, line)
			}
		} else {
			fmt.Fprintln(writer, line)
		}
	}

	return nil
}

// computeLineDiff computes a unified diff between two sets of lines
func computeLineDiff(a, b []string) []string {
	// Use a simple diff algorithm
	lcs := computeLCS(a, b)
	return formatUnifiedDiff(a, b, lcs)
}

// computeLCS computes the Longest Common Subsequence indices
func computeLCS(a, b []string) [][2]int {
	m, n := len(a), len(b)

	// Build LCS table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Backtrack to find LCS indices
	var result [][2]int
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append([][2]int{{i - 1, j - 1}}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return result
}

// formatUnifiedDiff formats the diff in unified diff style
func formatUnifiedDiff(a, b []string, lcs [][2]int) []string {
	var result []string

	// Track whether there are any actual differences
	hasDiff := false

	aIdx, bIdx := 0, 0
	lcsIdx := 0

	for aIdx < len(a) || bIdx < len(b) {
		// Find the next common line
		var nextACommon, nextBCommon int
		if lcsIdx < len(lcs) {
			nextACommon = lcs[lcsIdx][0]
			nextBCommon = lcs[lcsIdx][1]
		} else {
			nextACommon = len(a)
			nextBCommon = len(b)
		}

		// Output lines only in a (removed)
		for aIdx < nextACommon {
			if strings.TrimSpace(a[aIdx]) != "" {
				result = append(result, "- "+a[aIdx])
				hasDiff = true
			}
			aIdx++
		}

		// Output lines only in b (added)
		for bIdx < nextBCommon {
			if strings.TrimSpace(b[bIdx]) != "" {
				result = append(result, "+ "+b[bIdx])
				hasDiff = true
			}
			bIdx++
		}

		// Skip the common line
		if lcsIdx < len(lcs) {
			// Optionally include context lines
			// result = append(result, "  "+a[aIdx])
			aIdx++
			bIdx++
			lcsIdx++
		}
	}

	if !hasDiff {
		return nil
	}

	return result
}
