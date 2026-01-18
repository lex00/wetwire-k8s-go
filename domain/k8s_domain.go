// Package domain provides the K8sDomain implementation for wetwire-core-go.
package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	coredomain "github.com/lex00/wetwire-core-go/domain"
	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/lex00/wetwire-k8s-go/internal/lint"
	"github.com/lex00/wetwire-k8s-go/internal/serialize"
	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "dev"

// Re-export core types for convenience
type (
	Context      = coredomain.Context
	BuildOpts    = coredomain.BuildOpts
	LintOpts     = coredomain.LintOpts
	InitOpts     = coredomain.InitOpts
	ValidateOpts = coredomain.ValidateOpts
	ListOpts     = coredomain.ListOpts
	GraphOpts    = coredomain.GraphOpts
	Result       = coredomain.Result
	Error        = coredomain.Error
)

var (
	NewResult              = coredomain.NewResult
	NewResultWithData      = coredomain.NewResultWithData
	NewErrorResult         = coredomain.NewErrorResult
	NewErrorResultMultiple = coredomain.NewErrorResultMultiple
)

// K8sDomain implements the Domain interface for Kubernetes manifest generation.
type K8sDomain struct{}

// Name returns "k8s"
func (d *K8sDomain) Name() string {
	return "k8s"
}

// Version returns the current version
func (d *K8sDomain) Version() string {
	return Version
}

// Builder returns the K8s builder implementation
func (d *K8sDomain) Builder() coredomain.Builder {
	return &k8sBuilder{}
}

// Linter returns the K8s linter implementation
func (d *K8sDomain) Linter() coredomain.Linter {
	return &k8sLinter{}
}

// Initializer returns the K8s initializer implementation
func (d *K8sDomain) Initializer() coredomain.Initializer {
	return &k8sInitializer{}
}

// Validator returns the K8s validator implementation
func (d *K8sDomain) Validator() coredomain.Validator {
	return &k8sValidator{}
}

// Lister returns the K8s lister implementation
func (d *K8sDomain) Lister() coredomain.Lister {
	return &k8sLister{}
}

// Grapher returns the K8s grapher implementation
func (d *K8sDomain) Grapher() coredomain.Grapher {
	return &k8sGrapher{}
}

// CreateRootCommand creates the root command using the domain interface.
func CreateRootCommand(d coredomain.Domain) *cobra.Command {
	return coredomain.Run(d)
}

// k8sBuilder implements domain.Builder
type k8sBuilder struct{}

func (b *k8sBuilder) Build(ctx *Context, path string, opts BuildOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	resources, err := discoverResources(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	if len(resources) == 0 {
		return NewErrorResult("no resources found", Error{
			Path:    absPath,
			Message: "no Kubernetes resources found",
		}), nil
	}

	// Validate references
	if err := build.ValidateReferences(resources); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Detect cycles
	if err := build.DetectCycles(resources); err != nil {
		return nil, fmt.Errorf("cycle detection failed: %w", err)
	}

	// Topological sort
	orderedResources, err := build.TopologicalSort(resources)
	if err != nil {
		return nil, fmt.Errorf("ordering failed: %w", err)
	}

	// Serialize resources
	var outputData []byte
	if opts.Format == "json" {
		outputData, err = serializeToJSON(orderedResources)
	} else {
		outputData, err = serializeToYAML(orderedResources)
	}
	if err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}

	// Handle output file
	if !opts.DryRun && opts.Output != "" {
		if err := os.WriteFile(opts.Output, outputData, 0644); err != nil {
			return nil, fmt.Errorf("write output: %w", err)
		}
		return NewResult(fmt.Sprintf("Wrote %s", opts.Output)), nil
	}

	return NewResultWithData("Build completed", string(outputData)), nil
}

// k8sLinter implements domain.Linter
type k8sLinter struct{}

func (l *k8sLinter) Lint(ctx *Context, path string, opts LintOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Create linter config
	config := &lint.Config{
		MinSeverity:   lint.SeverityInfo,
		DisabledRules: opts.Disable,
	}

	// If Fix mode is enabled, run the fixer first
	if opts.Fix {
		fixer := lint.NewFixer(config)
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("stat path: %w", err)
		}

		if info.IsDir() {
			_, err = fixer.FixDirectory(absPath)
		} else {
			_, err = fixer.FixFile(absPath)
		}
		if err != nil {
			return nil, fmt.Errorf("fix failed: %w", err)
		}
	}

	// Create linter
	linter := lint.NewLinter(config)

	// Run lint
	issues, err := linter.Lint(absPath)
	if err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	if len(issues) == 0 {
		return NewResult("No lint issues found"), nil
	}

	// Convert to domain errors
	errs := make([]Error, 0, len(issues))
	for _, issue := range issues {
		errs = append(errs, Error{
			Path:     issue.File,
			Line:     issue.Line,
			Severity: issue.Severity.String(),
			Message:  issue.Message,
			Code:     issue.Rule,
		})
	}

	// If Fix mode was enabled but issues remain, note that in the message
	if opts.Fix {
		return NewErrorResultMultiple("lint issues found (some issues could not be auto-fixed)", errs), nil
	}

	return NewErrorResultMultiple("lint issues found", errs), nil
}

// k8sInitializer implements domain.Initializer
type k8sInitializer struct{}

func (i *k8sInitializer) Init(ctx *Context, path string, opts InitOpts) (*Result, error) {
	// Use opts.Path if provided, otherwise fall back to path argument
	targetPath := opts.Path
	if targetPath == "" || targetPath == "." {
		targetPath = path
	}

	// Handle scenario initialization
	if opts.Scenario {
		return i.initScenario(ctx, targetPath, opts)
	}

	// Basic project initialization
	return i.initProject(ctx, targetPath, opts)
}

// initScenario creates a full scenario structure with prompts and expected outputs
func (i *k8sInitializer) initScenario(ctx *Context, path string, opts InitOpts) (*Result, error) {
	name := opts.Name
	if name == "" {
		name = filepath.Base(path)
	}

	description := opts.Description
	if description == "" {
		description = "Kubernetes manifest scenario"
	}

	// Use core's scenario scaffolding
	scenario := coredomain.ScaffoldScenario(name, description, "k8s")
	created, err := coredomain.WriteScenario(path, scenario)
	if err != nil {
		return nil, fmt.Errorf("write scenario: %w", err)
	}

	// Create k8s-specific expected directory structure
	expectedK8sDir := filepath.Join(path, "expected", "k8s")
	if err := os.MkdirAll(expectedK8sDir, 0755); err != nil {
		return nil, fmt.Errorf("create expected/k8s directory: %w", err)
	}

	// Create example namespace in expected/k8s/
	exampleNamespace := `package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppNamespace defines the namespace for the application
var AppNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-app",
		Labels: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
	},
}
`
	nsPath := filepath.Join(expectedK8sDir, "namespace.go")
	if err := os.WriteFile(nsPath, []byte(exampleNamespace), 0644); err != nil {
		return nil, fmt.Errorf("write example namespace: %w", err)
	}
	created = append(created, "expected/k8s/namespace.go")

	return NewResultWithData(
		fmt.Sprintf("Created scenario %s with %d files", name, len(created)),
		created,
	), nil
}

// initProject creates a basic project with example Kubernetes resources
func (i *k8sInitializer) initProject(ctx *Context, path string, opts InitOpts) (*Result, error) {
	// Create directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// Create k8s directory
	k8sDir := filepath.Join(path, "k8s")
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		return nil, fmt.Errorf("create k8s directory: %w", err)
	}

	// Create namespace.go
	namespaceContent := `package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppNamespace defines the namespace for the application
var AppNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-app",
		Labels: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
	},
}
`
	nsPath := filepath.Join(k8sDir, "namespace.go")
	if err := os.WriteFile(nsPath, []byte(namespaceContent), 0644); err != nil {
		return nil, fmt.Errorf("write namespace.go: %w", err)
	}

	// Create .wetwire.yaml
	configContent := `# wetwire-k8s configuration
# See https://github.com/lex00/wetwire-k8s-go for documentation

# Source directory containing Go files with Kubernetes resources
source: k8s

# Output configuration
output:
  # Output format: yaml or json
  format: yaml
  # Output path (use '-' for stdout)
  path: manifests.yaml

# Build options
build:
  # Skip validation checks
  skip_validation: false
`
	configPath := filepath.Join(path, ".wetwire.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, fmt.Errorf("write .wetwire.yaml: %w", err)
	}

	return NewResult(fmt.Sprintf("Initialized k8s project in %s", path)), nil
}

// k8sValidator implements domain.Validator
type k8sValidator struct{}

func (v *k8sValidator) Validate(ctx *Context, path string, opts ValidateOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	resources, err := discoverResources(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Validate references
	if err := build.ValidateReferences(resources); err != nil {
		return NewErrorResult("validation failed", Error{
			Path:    absPath,
			Message: err.Error(),
		}), nil
	}

	// Detect cycles
	if err := build.DetectCycles(resources); err != nil {
		return NewErrorResult("cycle detected", Error{
			Path:    absPath,
			Message: err.Error(),
		}), nil
	}

	return NewResult("Validation passed"), nil
}

// k8sLister implements domain.Lister
type k8sLister struct{}

func (l *k8sLister) List(ctx *Context, path string, opts ListOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	resources, err := discoverResources(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Build list
	list := make([]map[string]any, 0)
	for _, r := range resources {
		item := map[string]any{
			"name": r.Name,
			"type": r.Type,
			"file": r.File,
			"line": r.Line,
		}
		if len(r.Dependencies) > 0 {
			item["dependencies"] = r.Dependencies
		}
		list = append(list, item)
	}

	return NewResultWithData(fmt.Sprintf("Discovered %d resources", len(list)), list), nil
}

// k8sGrapher implements domain.Grapher
type k8sGrapher struct{}

func (g *k8sGrapher) Graph(ctx *Context, path string, opts GraphOpts) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Discover all resources
	resources, err := discoverResources(absPath)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Generate graph
	var graph string
	switch opts.Format {
	case "dot", "":
		graph = generateDOTGraph(resources)
	case "mermaid":
		graph = generateMermaidGraph(resources)
	default:
		return nil, fmt.Errorf("unknown format: %s", opts.Format)
	}

	return NewResultWithData("Graph generated", graph), nil
}

// Helper functions

// discoverResources discovers resources from the given path
func discoverResources(path string) ([]discover.Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %q: %w", path, err)
	}

	if info.IsDir() {
		return discover.DiscoverDirectory(path)
	}
	return discover.DiscoverFile(path)
}

// serializeToYAML serializes resources to YAML format
func serializeToYAML(resources []discover.Resource) ([]byte, error) {
	// Convert resources to manifests
	var manifests []interface{}
	for _, r := range resources {
		manifest := createManifestFromResource(r)
		manifests = append(manifests, manifest)
	}
	return serialize.ToMultiYAML(manifests)
}

// serializeToJSON serializes resources to JSON format
func serializeToJSON(resources []discover.Resource) ([]byte, error) {
	// Convert resources to manifests
	var manifests []interface{}
	for _, r := range resources {
		manifest := createManifestFromResource(r)
		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		return []byte("[]"), nil
	}

	if len(manifests) == 1 {
		return serialize.ToJSON(manifests[0])
	}

	// For multiple resources, create a JSON array
	var result []byte
	result = append(result, '[')
	for i, m := range manifests {
		if i > 0 {
			result = append(result, ',', '\n')
		}
		jsonBytes, err := serialize.ToJSON(m)
		if err != nil {
			return nil, err
		}
		result = append(result, jsonBytes...)
	}
	result = append(result, ']')
	return result, nil
}

// createManifestFromResource creates a basic manifest map from discovered resource
func createManifestFromResource(r discover.Resource) map[string]interface{} {
	// Parse the resource type to determine apiVersion and kind
	apiVersion, kind := parseResourceType(r.Type)

	// Create a manifest with the discovered information
	manifest := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": toKubernetesName(r.Name),
		},
	}

	return manifest
}

// parseResourceType extracts apiVersion and kind from a Go type string
// e.g., "appsv1.Deployment" -> ("apps/v1", "Deployment")
func parseResourceType(typeStr string) (string, string) {
	// Default values
	apiVersion := "v1"
	kind := "Unknown"

	// Split by dot to separate package from type
	parts := strings.Split(typeStr, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		kind = parts[1]

		// Map package aliases to API versions
		apiVersion = mapPackageToAPIVersion(pkg)
	} else if len(parts) == 1 {
		kind = parts[0]
	}

	return apiVersion, kind
}

// mapPackageToAPIVersion maps Go package aliases to Kubernetes API versions
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

// toKubernetesName converts a Go variable name to a Kubernetes resource name
// e.g., "MyDeployment" -> "my-deployment"
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

// generateDOTGraph generates a DOT format dependency graph
func generateDOTGraph(resources []discover.Resource) string {
	var b strings.Builder
	b.WriteString("digraph dependencies {\n")
	b.WriteString("  rankdir=TB;\n")
	b.WriteString("  node [shape=box];\n\n")

	// Build resource set
	resourceSet := make(map[string]bool)
	for _, r := range resources {
		resourceSet[r.Name] = true
	}

	// Output nodes
	for _, r := range resources {
		label := fmt.Sprintf("%s\\n(%s)", r.Name, r.Type)
		fmt.Fprintf(&b, "  \"%s\" [label=\"%s\"];\n", r.Name, label)
	}

	b.WriteString("\n")

	// Output edges
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			if resourceSet[dep] {
				fmt.Fprintf(&b, "  \"%s\" -> \"%s\";\n", r.Name, dep)
			}
		}
	}

	b.WriteString("}")
	return b.String()
}

// generateMermaidGraph generates a Mermaid format dependency graph
func generateMermaidGraph(resources []discover.Resource) string {
	var b strings.Builder
	b.WriteString("graph TD\n")

	// Build resource set
	resourceSet := make(map[string]bool)
	for _, r := range resources {
		resourceSet[r.Name] = true
	}

	// Output nodes
	for _, r := range resources {
		fmt.Fprintf(&b, "  %s[%s: %s]\n", r.Name, r.Name, r.Type)
	}

	// Output edges
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			if resourceSet[dep] {
				fmt.Fprintf(&b, "  %s --> %s\n", r.Name, dep)
			}
		}
	}

	return b.String()
}
