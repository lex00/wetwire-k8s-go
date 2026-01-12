// MCP server implementation for embedded design mode.
//
// When mcp subcommand is called, this runs the MCP protocol over stdio,
// providing wetwire_build, wetwire_lint, wetwire_validate, and wetwire_import tools.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/importer"
	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/spf13/cobra"
)

// newMCPCmd creates the mcp subcommand.
func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server for Claude Code integration",
		Long: `Start an MCP (Model Context Protocol) server for integration with Claude Code.

This command runs an MCP server over stdio, providing tools for:
  - wetwire_build: Generate Kubernetes manifests from Go code
  - wetwire_lint: Lint Go code for wetwire-k8s patterns
  - wetwire_validate: Validate manifests using kubeconform
  - wetwire_import: Convert YAML manifests to Go code

This is typically called by Claude Code or other MCP clients, not directly by users.`,
		RunE: runMCPServer,
	}

	return cmd
}

// runMCPServer starts the MCP server on stdio transport.
func runMCPServer(cmd *cobra.Command, args []string) error {
	debugLog("Starting wetwire-k8s MCP server version %s", Version)

	// Create MCP server using wetwire-core-go infrastructure
	server := mcp.NewServer(mcp.Config{
		Name:    "wetwire-k8s",
		Version: Version,
		Debug:   os.Getenv("WETWIRE_MCP_DEBUG") != "",
	})

	// Register standard wetwire tools using core infrastructure
	mcp.RegisterStandardTools(server, "k8s", mcp.StandardToolHandlers{
		Init:     nil, // Not yet implemented
		Build:    buildToolHandler,
		Lint:     lintToolHandler,
		Validate: validateToolHandler,
		Import:   importToolHandler,
		List:     nil, // Not yet implemented
		Graph:    nil, // Not yet implemented
	})

	debugLog("Registered all tools, starting stdio server")

	// Start stdio server
	return server.Start(context.Background())
}

// debugLog logs messages when WETWIRE_MCP_DEBUG is set.
func debugLog(format string, args ...interface{}) {
	if os.Getenv("WETWIRE_MCP_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// Helper functions to parse arguments from the map
func parseString(args map[string]any, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

func parseBoolean(args map[string]any, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

// buildToolHandler handles the wetwire_build tool call using core infrastructure.
// It generates Kubernetes YAML/JSON manifests from Go source code.
func buildToolHandler(ctx context.Context, args map[string]any) (string, error) {
	debugLog("wetwire_build tool called with args: %v", args)

	// Parse arguments using core schema conventions
	pkgPath := parseString(args, "package", ".")
	format := parseString(args, "format", "yaml")
	dryRun := parseBoolean(args, "dry_run", true) // Default to dry_run for safety

	// Resolve to absolute path
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}

	debugLog("building from path: %s, format: %s, dry_run: %v", absPath, format, dryRun)

	// Run the build pipeline
	result, err := build.Build(absPath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return "", fmt.Errorf("build failed: %v", err)
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		return "No Kubernetes resources found in the specified path.", nil
	}

	// Generate output
	output, err := generateOutput(result.OrderedResources, format)
	if err != nil {
		return "", fmt.Errorf("failed to generate output: %v", err)
	}

	return string(output), nil
}

// lintToolHandler handles the wetwire_lint tool call using core infrastructure.
// It lints Go files for wetwire-k8s patterns and best practices.
func lintToolHandler(ctx context.Context, args map[string]any) (string, error) {
	debugLog("wetwire_lint tool called with args: %v", args)

	// Parse arguments using core schema conventions
	pkgPath := parseString(args, "package", ".")
	format := parseString(args, "format", "text")

	// Resolve to absolute path
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}

	debugLog("linting path: %s, format: %s", absPath, format)

	// Create linter with default config
	linter := lint.NewLinter(nil)

	// Run lint
	result, err := linter.LintWithResult(absPath)
	if err != nil {
		return "", fmt.Errorf("lint failed: %v", err)
	}

	// Format output based on requested format
	var output string
	switch format {
	case "json":
		output, err = formatLintJSON(result)
		if err != nil {
			return "", fmt.Errorf("failed to format output: %v", err)
		}
	default:
		output = formatLintText(result)
	}

	return output, nil
}

// importToolHandler handles the wetwire_import tool call using core infrastructure.
// It converts Kubernetes YAML manifests to Go code.
func importToolHandler(ctx context.Context, args map[string]any) (string, error) {
	debugLog("wetwire_import tool called with args: %v", args)

	// Parse arguments - for import, we accept files array or direct yaml content
	// The core schema expects files array, but we also support inline yaml
	var yamlContent string
	if files, ok := args["files"].([]interface{}); ok && len(files) > 0 {
		// Read content from first file
		filePath := fmt.Sprintf("%v", files[0])
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve path: %v", err)
		}
		content, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %v", err)
		}
		yamlContent = string(content)
	} else if yaml, ok := args["yaml"].(string); ok {
		// Backward compatibility: accept yaml parameter directly
		yamlContent = yaml
	}

	if yamlContent == "" {
		return "", fmt.Errorf("files parameter is required (array of file paths)")
	}

	packageName := parseString(args, "package", "main")
	varPrefix := parseString(args, "var_prefix", "")

	debugLog("importing YAML, package: %s, varPrefix: %s", packageName, varPrefix)

	// Configure importer
	opts := importer.Options{
		PackageName: packageName,
		VarPrefix:   varPrefix,
	}

	// Run import
	result, err := importer.ImportBytes([]byte(yamlContent), opts)
	if err != nil {
		return "", fmt.Errorf("import failed: %v", err)
	}

	// Build response with warnings if any
	var output strings.Builder
	if len(result.Warnings) > 0 {
		output.WriteString("// Warnings:\n")
		for _, warn := range result.Warnings {
			output.WriteString(fmt.Sprintf("// - %s\n", warn))
		}
		output.WriteString("\n")
	}
	output.WriteString(result.GoCode)

	return output.String(), nil
}

// validateToolHandler handles the wetwire_validate tool call using core infrastructure.
// It validates Kubernetes manifests using kubeconform.
func validateToolHandler(ctx context.Context, args map[string]any) (string, error) {
	debugLog("wetwire_validate tool called with args: %v", args)

	// Parse arguments using core schema conventions
	path := parseString(args, "path", "")

	if path == "" {
		return "", fmt.Errorf("path parameter is required")
	}

	// Check for kubeconform installation
	kubeconformPath, err := exec.LookPath("kubeconform")
	if err != nil {
		return "", fmt.Errorf("kubeconform is not installed. Install it from https://github.com/yannh/kubeconform")
	}

	debugLog("using kubeconform at: %s", kubeconformPath)

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}

	// Build kubeconform arguments
	cmdArgs := []string{"-summary", absPath}

	debugLog("running kubeconform with args: %v", cmdArgs)

	// Run kubeconform
	kubeCmd := exec.CommandContext(ctx, kubeconformPath, cmdArgs...)

	var stdout, stderr bytes.Buffer
	kubeCmd.Stdout = &stdout
	kubeCmd.Stderr = &stderr

	err = kubeCmd.Run()

	// Build output
	var output strings.Builder
	if stdout.Len() > 0 {
		output.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(stderr.String())
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			output.WriteString(fmt.Sprintf("\nValidation failed with exit code %d", exitErr.ExitCode()))
		} else {
			return "", fmt.Errorf("kubeconform execution failed: %v", err)
		}
	}

	if output.Len() == 0 {
		output.WriteString("Validation passed - all resources are valid.")
	}

	return output.String(), nil
}

// formatLintText formats lint results as plain text.
func formatLintText(result *lint.LintResult) string {
	if len(result.Issues) == 0 {
		return fmt.Sprintf("No issues found. Scanned %d files.", result.TotalFiles)
	}

	var output strings.Builder
	for _, issue := range result.Issues {
		output.WriteString(fmt.Sprintf("%s:%d:%d: %s [%s] %s\n",
			issue.File, issue.Line, issue.Column,
			issue.Severity.String(), issue.Rule, issue.Message))
	}

	output.WriteString(fmt.Sprintf("\nSummary: %d errors, %d warnings, %d info messages in %d files\n",
		result.ErrorCount, result.WarningCount, result.InfoCount, result.FilesWithIssues))

	return output.String()
}

// formatLintJSON formats lint results as JSON.
func formatLintJSON(result *lint.LintResult) (string, error) {
	type jsonIssue struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Severity string `json:"severity"`
		Rule     string `json:"rule"`
		Message  string `json:"message"`
	}

	type jsonResult struct {
		TotalFiles      int         `json:"total_files"`
		FilesWithIssues int         `json:"files_with_issues"`
		ErrorCount      int         `json:"error_count"`
		WarningCount    int         `json:"warning_count"`
		InfoCount       int         `json:"info_count"`
		Issues          []jsonIssue `json:"issues"`
	}

	jr := jsonResult{
		TotalFiles:      result.TotalFiles,
		FilesWithIssues: result.FilesWithIssues,
		ErrorCount:      result.ErrorCount,
		WarningCount:    result.WarningCount,
		InfoCount:       result.InfoCount,
		Issues:          make([]jsonIssue, len(result.Issues)),
	}

	for i, issue := range result.Issues {
		jr.Issues[i] = jsonIssue{
			File:     issue.File,
			Line:     issue.Line,
			Column:   issue.Column,
			Severity: issue.Severity.String(),
			Rule:     issue.Rule,
			Message:  issue.Message,
		}
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
