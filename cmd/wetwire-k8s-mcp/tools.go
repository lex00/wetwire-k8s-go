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

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/lex00/wetwire-k8s-go/internal/importer"
	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/lex00/wetwire-k8s-go/internal/serialize"
	"github.com/mark3labs/mcp-go/mcp"
)

// buildHandler handles the build tool call.
// It generates Kubernetes YAML/JSON manifests from Go source code.
func buildHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	debugLog("build tool called with args: %v", request.Params.Arguments)

	// Parse arguments
	path := mcp.ParseString(request, "path", ".")
	format := mcp.ParseString(request, "format", "yaml")

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", absPath)), nil
	}

	debugLog("building from path: %s, format: %s", absPath, format)

	// Run the build pipeline
	result, err := build.Build(absPath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("build failed: %v", err)), nil
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		return mcp.NewToolResultText("No Kubernetes resources found in the specified path."), nil
	}

	// Generate output
	output, err := generateOutput(result.OrderedResources, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate output: %v", err)), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

// lintHandler handles the lint tool call.
// It lints Go files for wetwire-k8s patterns and best practices.
func lintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	debugLog("lint tool called with args: %v", request.Params.Arguments)

	// Parse arguments
	path := mcp.ParseString(request, "path", ".")
	format := mcp.ParseString(request, "format", "text")

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", absPath)), nil
	}

	debugLog("linting path: %s, format: %s", absPath, format)

	// Create linter with default config
	linter := lint.NewLinter(nil)

	// Run lint
	result, err := linter.LintWithResult(absPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("lint failed: %v", err)), nil
	}

	// Format output based on requested format
	var output string
	switch format {
	case "json":
		output, err = formatLintJSON(result)
	case "github":
		output = formatLintGitHub(result)
	default:
		output = formatLintText(result)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format output: %v", err)), nil
	}

	return mcp.NewToolResultText(output), nil
}

// importHandler handles the import tool call.
// It converts Kubernetes YAML manifests to Go code.
func importHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	debugLog("import tool called with args: %v", request.Params.Arguments)

	// Parse arguments
	yamlContent := mcp.ParseString(request, "yaml", "")
	packageName := mcp.ParseString(request, "package", "main")
	varPrefix := mcp.ParseString(request, "var_prefix", "")

	if yamlContent == "" {
		return mcp.NewToolResultError("yaml parameter is required"), nil
	}

	debugLog("importing YAML, package: %s, varPrefix: %s", packageName, varPrefix)

	// Configure importer
	opts := importer.Options{
		PackageName: packageName,
		VarPrefix:   varPrefix,
	}

	// Run import
	result, err := importer.ImportBytes([]byte(yamlContent), opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("import failed: %v", err)), nil
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

	return mcp.NewToolResultText(output.String()), nil
}

// validateHandler handles the validate tool call.
// It validates Kubernetes manifests using kubeconform.
func validateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	debugLog("validate tool called with args: %v", request.Params.Arguments)

	// Parse arguments
	yamlContent := mcp.ParseString(request, "yaml", "")
	path := mcp.ParseString(request, "path", "")
	strict := mcp.ParseBoolean(request, "strict", false)
	k8sVersion := mcp.ParseString(request, "kubernetes_version", "")

	// Must have either yaml content or path
	if yamlContent == "" && path == "" {
		return mcp.NewToolResultError("either 'yaml' or 'path' parameter is required"), nil
	}

	// Check for kubeconform installation
	kubeconformPath, err := exec.LookPath("kubeconform")
	if err != nil {
		return mcp.NewToolResultError("kubeconform is not installed. Install it from https://github.com/yannh/kubeconform"), nil
	}

	debugLog("using kubeconform at: %s", kubeconformPath)

	// Build kubeconform arguments
	args := []string{"-summary"}
	if strict {
		args = append(args, "-strict")
	}
	if k8sVersion != "" {
		args = append(args, "-kubernetes-version", k8sVersion)
	}

	var inputData []byte
	if yamlContent != "" {
		// Validate from provided YAML content
		inputData = []byte(yamlContent)
		args = append(args, "-")
	} else {
		// Validate from file/directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
		}

		if _, err := os.Stat(absPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", absPath)), nil
		}

		args = append(args, absPath)
	}

	debugLog("running kubeconform with args: %v", args)

	// Run kubeconform
	cmd := exec.CommandContext(ctx, kubeconformPath, args...)
	if inputData != nil {
		cmd.Stdin = bytes.NewReader(inputData)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

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
			return mcp.NewToolResultError(fmt.Sprintf("kubeconform execution failed: %v", err)), nil
		}
	}

	if output.Len() == 0 {
		output.WriteString("Validation passed - all resources are valid.")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// generateOutput creates YAML or JSON from discovered resources.
func generateOutput(resources []discover.Resource, format string) ([]byte, error) {
	var manifests []interface{}
	for _, r := range resources {
		manifest := createManifestFromResource(r)
		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		return []byte{}, nil
	}

	if format == "json" {
		return serializeResourcesJSON(manifests)
	}
	return serializeResourcesYAML(manifests)
}

// createManifestFromResource creates a basic manifest map from discovered resource.
func createManifestFromResource(r discover.Resource) map[string]interface{} {
	apiVersion, kind := parseResourceType(r.Type)

	manifest := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": toKubernetesName(r.Name),
		},
	}

	return manifest
}

// parseResourceType extracts apiVersion and kind from a Go type string.
func parseResourceType(typeStr string) (string, string) {
	apiVersion := "v1"
	kind := "Unknown"

	parts := strings.Split(typeStr, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		kind = parts[1]
		apiVersion = mapPackageToAPIVersion(pkg)
	} else if len(parts) == 1 {
		kind = parts[0]
	}

	return apiVersion, kind
}

// mapPackageToAPIVersion maps Go package aliases to Kubernetes API versions.
func mapPackageToAPIVersion(pkg string) string {
	packageMap := map[string]string{
		"corev1":         "v1",
		"appsv1":         "apps/v1",
		"batchv1":        "batch/v1",
		"networkingv1":   "networking.k8s.io/v1",
		"rbacv1":         "rbac.authorization.k8s.io/v1",
		"storagev1":      "storage.k8s.io/v1",
		"policyv1":       "policy/v1",
		"autoscalingv1":  "autoscaling/v1",
		"autoscalingv2":  "autoscaling/v2",
		"admissionv1":    "admissionregistration.k8s.io/v1",
		"certificatesv1": "certificates.k8s.io/v1",
	}

	if version, ok := packageMap[pkg]; ok {
		return version
	}
	return "v1"
}

// toKubernetesName converts a Go variable name to a Kubernetes resource name.
func toKubernetesName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// serializeResourcesYAML converts resources to multi-document YAML.
func serializeResourcesYAML(resources []interface{}) ([]byte, error) {
	return serialize.ToMultiYAML(resources)
}

// serializeResourcesJSON converts resources to JSON.
func serializeResourcesJSON(resources []interface{}) ([]byte, error) {
	if len(resources) == 1 {
		return serialize.ToJSON(resources[0])
	}

	var result []byte
	result = append(result, '[')
	for i, r := range resources {
		if i > 0 {
			result = append(result, ',', '\n')
		}
		jsonBytes, err := serialize.ToJSON(r)
		if err != nil {
			return nil, err
		}
		result = append(result, jsonBytes...)
	}
	result = append(result, ']')
	return result, nil
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

// formatLintGitHub formats lint results in GitHub Actions format.
func formatLintGitHub(result *lint.LintResult) string {
	if len(result.Issues) == 0 {
		return ""
	}

	var output strings.Builder
	for _, issue := range result.Issues {
		// GitHub Actions annotation format
		var level string
		switch issue.Severity {
		case lint.SeverityError:
			level = "error"
		case lint.SeverityWarning:
			level = "warning"
		default:
			level = "notice"
		}

		output.WriteString(fmt.Sprintf("::%s file=%s,line=%d,col=%d,title=%s::%s\n",
			level, issue.File, issue.Line, issue.Column, issue.Rule, issue.Message))
	}

	return output.String()
}
