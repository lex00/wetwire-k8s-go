package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
)

// Build runs the complete build pipeline on the given path.
// The path can be either a single Go file or a directory.
//
// Pipeline stages:
// 1. DISCOVER - Parse source files and find resource declarations
// 2. VALIDATE - Check references exist, detect cycles
// 3. EXTRACT - Execute source to get runtime values (placeholder for now)
// 4. ORDER - Topological sort by dependencies
// 5. SERIALIZE - Convert to YAML/JSON (depends on #3, stub for now)
// 6. EMIT - Write output file(s)
func Build(path string, opts Options) (*Result, error) {
	// Stage 1: DISCOVER
	resources, err := discoverResources(path)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Stage 2: VALIDATE
	if err := ValidateReferences(resources); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := DetectCycles(resources); err != nil {
		return nil, fmt.Errorf("cycle detection failed: %w", err)
	}

	// Stage 3: EXTRACT (placeholder for now)
	// This will be implemented in issue #3 to execute the source code
	// and get the actual runtime values of the resources.

	// Stage 4: ORDER
	orderedResources, err := TopologicalSort(resources)
	if err != nil {
		return nil, fmt.Errorf("ordering failed: %w", err)
	}

	// Stage 5: SERIALIZE (stub for now)
	// This will be implemented when we have the runtime values from stage 3.
	// For now, we just prepare the output paths.

	// Stage 6: EMIT
	result := &Result{
		Resources:        resources,
		OrderedResources: orderedResources,
	}

	if opts.OutputPath != "" {
		if err := emitOutput(result, opts); err != nil {
			return nil, fmt.Errorf("output failed: %w", err)
		}
	}

	return result, nil
}

// discoverResources discovers resources from the given path.
// The path can be either a single file or a directory.
func discoverResources(path string) ([]discover.Resource, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %q: %w", path, err)
	}

	// Discover resources based on whether it's a file or directory
	if info.IsDir() {
		return discover.DiscoverDirectory(path)
	}
	return discover.DiscoverFile(path)
}

// emitOutput writes the build output according to the specified options.
func emitOutput(result *Result, opts Options) error {
	switch opts.OutputMode {
	case SingleFile:
		return emitSingleFile(result, opts)
	case SeparateFiles:
		return emitSeparateFiles(result, opts)
	default:
		return fmt.Errorf("unknown output mode: %v", opts.OutputMode)
	}
}

// emitSingleFile writes all resources to a single YAML file.
func emitSingleFile(result *Result, opts Options) error {
	outputPath := opts.OutputPath

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// For now, we just set the output path in the result
	// The actual serialization will be implemented in issue #3
	result.OutputPath = outputPath

	// TODO: When serializer is available:
	// 1. Serialize each resource in orderedResources
	// 2. Join with "---\n" separator
	// 3. Write to outputPath

	return nil
}

// emitSeparateFiles writes each resource to its own file.
func emitSeparateFiles(result *Result, opts Options) error {
	outputDir := opts.OutputPath

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output paths for each resource
	outputPaths := make([]string, len(result.OrderedResources))
	for i, r := range result.OrderedResources {
		// Use resource name as filename (with .yaml extension)
		filename := fmt.Sprintf("%s.yaml", r.Name)
		outputPaths[i] = filepath.Join(outputDir, filename)
	}

	result.OutputPaths = outputPaths

	// TODO: When serializer is available:
	// 1. For each resource in orderedResources:
	//    - Serialize the resource
	//    - Write to its corresponding output path

	return nil
}
