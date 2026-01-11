package build

import "github.com/lex00/wetwire-k8s-go/internal/discover"

// OutputMode specifies how the built resources should be output.
type OutputMode int

const (
	// SingleFile outputs all resources to a single YAML file with document separators.
	SingleFile OutputMode = iota
	// SeparateFiles outputs each resource to its own file.
	SeparateFiles
)

// Options configures the build pipeline.
type Options struct {
	// OutputMode specifies whether to output to a single file or separate files.
	OutputMode OutputMode

	// OutputPath specifies where to write the output.
	// For SingleFile mode, this is the output file path.
	// For SeparateFiles mode, this is the output directory path.
	OutputPath string

	// Serializer is used to convert resources to YAML/JSON.
	// If nil, a stub serializer will be used (for testing).
	Serializer Serializer
}

// Result contains the output of the build pipeline.
type Result struct {
	// Resources are the discovered resources in original order.
	Resources []discover.Resource

	// OrderedResources are the resources sorted in topological order.
	OrderedResources []discover.Resource

	// OutputPath is the path to the output file (for SingleFile mode).
	OutputPath string

	// OutputPaths are the paths to the output files (for SeparateFiles mode).
	OutputPaths []string
}

// Serializer converts resources to YAML or JSON format.
// This is a stub interface that will be implemented in issue #3.
type Serializer interface {
	// ToYAML converts a resource to YAML bytes.
	ToYAML(resource interface{}) ([]byte, error)
}
