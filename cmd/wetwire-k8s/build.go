package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/lex00/wetwire-k8s-go/internal/serialize"
	"github.com/urfave/cli/v2"
)

// buildCommand creates the build subcommand
func buildCommand() *cli.Command {
	return &cli.Command{
		Name:      "build",
		Usage:     "Generate Kubernetes YAML manifests from Go code",
		ArgsUsage: "[PATH]",
		Description: `Build parses Go source files, discovers Kubernetes resource declarations,
and generates YAML or JSON manifests.

If PATH is not specified, the current directory is used.

Examples:
  wetwire-k8s build                    # Build from current directory
  wetwire-k8s build ./k8s              # Build from specific directory
  wetwire-k8s build -o manifests.yaml  # Save output to file
  wetwire-k8s build -f json            # Output as JSON`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (use '-' for stdout)",
				Value:   "-",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: yaml or json",
				Value:   "yaml",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show output without writing to file",
				Value: false,
			},
		},
		Action: runBuild,
	}
}

// runBuild executes the build command
func runBuild(c *cli.Context) error {
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
	if format != "yaml" && format != "json" {
		return fmt.Errorf("invalid format %q: must be 'yaml' or 'json'", format)
	}

	// Run the build pipeline
	result, err := build.Build(absPath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		return nil
	}

	// Generate output
	output, err := generateOutput(result.OrderedResources, format)
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Determine output destination
	outputPath := c.String("output")
	dryRun := c.Bool("dry-run")

	// Get the writer for output
	var writer io.Writer
	if outputPath == "-" || dryRun {
		// Write to stdout (or app's writer for testing)
		writer = c.App.Writer
		if writer == nil {
			writer = os.Stdout
		}
	} else {
		// Write to file
		if !dryRun {
			// Create output directory if needed
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			writer = file
		}
	}

	// If dry-run with output file, still write to stdout
	if dryRun && outputPath != "-" {
		writer = c.App.Writer
		if writer == nil {
			writer = os.Stdout
		}
	}

	// Write output
	if writer != nil {
		_, err = writer.Write(output)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// generateOutput creates the serialized output from discovered resources
func generateOutput(resources []discover.Resource, format string) ([]byte, error) {
	// Convert discovered resources to manifest maps
	// Since the current build pipeline only discovers metadata (not runtime values),
	// we create stub manifests based on the discovered metadata.
	// In a full implementation, stage 3 (EXTRACT) would execute the Go code
	// to get the actual resource values.

	var manifests []interface{}
	for _, r := range resources {
		manifest := createManifestFromResource(r)
		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		return []byte{}, nil
	}

	// Serialize based on format
	if format == "json" {
		return serializeResourcesJSON(manifests)
	}
	return serializeResourcesYAML(manifests)
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

// serializeResourcesYAML converts resources to multi-document YAML
func serializeResourcesYAML(resources []interface{}) ([]byte, error) {
	return serialize.ToMultiYAML(resources)
}

// serializeResourcesJSON converts resources to JSON array
func serializeResourcesJSON(resources []interface{}) ([]byte, error) {
	if len(resources) == 1 {
		return serialize.ToJSON(resources[0])
	}

	// For multiple resources, create a JSON array
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
