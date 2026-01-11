package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/urfave/cli/v2"
)

// graphCommand creates the graph subcommand
func graphCommand() *cli.Command {
	return &cli.Command{
		Name:      "graph",
		Usage:     "Show resource dependency graph",
		ArgsUsage: "[PATH]",
		Description: `Graph displays the dependency relationships between Kubernetes resources.

If PATH is not specified, the current directory is used.

Output formats:
- ascii: Text-based tree representation (default)
- dot: Graphviz DOT format for visualization

Examples:
  wetwire-k8s graph                      # Show graph from current directory
  wetwire-k8s graph ./k8s                # Show graph from specific directory
  wetwire-k8s graph -f dot               # Output as DOT format
  wetwire-k8s graph -f dot -o graph.dot  # Save DOT output to file`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: ascii or dot",
				Value:   "ascii",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (default: stdout)",
				Value:   "",
			},
		},
		Action: runGraph,
	}
}

// runGraph executes the graph command
func runGraph(c *cli.Context) error {
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

	// Discover resources
	resources, err := discoverResourcesForGraph(absPath)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	// Determine output writer
	var writer io.Writer
	outputPath := c.String("output")
	if outputPath != "" {
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
	} else {
		writer = c.App.Writer
		if writer == nil {
			writer = os.Stdout
		}
	}

	// Check if we have resources
	if len(resources) == 0 {
		fmt.Fprintln(writer, "No resources found")
		return nil
	}

	// Output in requested format
	format := strings.ToLower(c.String("format"))
	switch format {
	case "dot":
		return outputDOT(writer, resources)
	case "ascii":
		return outputASCII(writer, resources)
	default:
		return fmt.Errorf("invalid format %q: must be 'ascii' or 'dot'", format)
	}
}

// discoverResourcesForGraph discovers resources from the given path
func discoverResourcesForGraph(path string) ([]discover.Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %q: %w", path, err)
	}

	if info.IsDir() {
		return discover.DiscoverDirectory(path)
	}
	return discover.DiscoverFile(path)
}

// outputDOT outputs the dependency graph in Graphviz DOT format
func outputDOT(w io.Writer, resources []discover.Resource) error {
	fmt.Fprintln(w, "digraph dependencies {")
	fmt.Fprintln(w, "  rankdir=TB;")
	fmt.Fprintln(w, "  node [shape=box];")
	fmt.Fprintln(w)

	// Build a set of all resource names for validation
	resourceSet := make(map[string]bool)
	for _, r := range resources {
		resourceSet[r.Name] = true
	}

	// Output nodes with labels
	for _, r := range resources {
		label := fmt.Sprintf("%s\\n(%s)", r.Name, r.Type)
		fmt.Fprintf(w, "  \"%s\" [label=\"%s\"];\n", r.Name, label)
	}

	fmt.Fprintln(w)

	// Output edges (dependencies)
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			// Only show edges to resources that exist in our set
			if resourceSet[dep] {
				fmt.Fprintf(w, "  \"%s\" -> \"%s\";\n", r.Name, dep)
			}
		}
	}

	fmt.Fprintln(w, "}")
	return nil
}

// outputASCII outputs the dependency graph as ASCII art
func outputASCII(w io.Writer, resources []discover.Resource) error {
	// Build dependency map
	deps := make(map[string][]string)
	dependsOn := make(map[string][]string)
	resourceSet := make(map[string]bool)

	for _, r := range resources {
		resourceSet[r.Name] = true
		deps[r.Name] = []string{}
		dependsOn[r.Name] = []string{}
	}

	for _, r := range resources {
		for _, dep := range r.Dependencies {
			if resourceSet[dep] {
				deps[r.Name] = append(deps[r.Name], dep)
				dependsOn[dep] = append(dependsOn[dep], r.Name)
			}
		}
	}

	// Find root nodes (no dependencies)
	var roots []string
	for _, r := range resources {
		if len(deps[r.Name]) == 0 {
			roots = append(roots, r.Name)
		}
	}

	// Sort for consistent output
	sort.Strings(roots)

	// Print header
	fmt.Fprintln(w, "Resource Dependency Graph")
	fmt.Fprintln(w, "=========================")
	fmt.Fprintln(w)

	if len(roots) == 0 && len(resources) > 0 {
		// All resources have dependencies - might be a cycle or complex graph
		// Just list them all
		fmt.Fprintln(w, "Resources (circular or complex dependencies detected):")
		for _, r := range resources {
			fmt.Fprintf(w, "  %s (%s)\n", r.Name, r.Type)
			if len(deps[r.Name]) > 0 {
				fmt.Fprintf(w, "    --> depends on: %s\n", strings.Join(deps[r.Name], ", "))
			}
		}
		return nil
	}

	// Print tree starting from roots
	printed := make(map[string]bool)

	for _, root := range roots {
		printResourceTree(w, root, "", true, resources, dependsOn, printed)
	}

	// Print any remaining resources that weren't reachable from roots
	for _, r := range resources {
		if !printed[r.Name] {
			printResourceTree(w, r.Name, "", true, resources, dependsOn, printed)
		}
	}

	return nil
}

// printResourceTree prints a resource and its dependents in tree format
func printResourceTree(w io.Writer, name string, prefix string, isLast bool, resources []discover.Resource, dependsOn map[string][]string, printed map[string]bool) {
	if printed[name] {
		return
	}
	printed[name] = true

	// Find the resource type
	var resourceType string
	for _, r := range resources {
		if r.Name == name {
			resourceType = r.Type
			break
		}
	}

	// Print this node
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		fmt.Fprintf(w, "%s (%s)\n", name, resourceType)
	} else {
		fmt.Fprintf(w, "%s%s%s (%s)\n", prefix, connector, name, resourceType)
	}

	// Get dependents (resources that depend on this one)
	dependents := dependsOn[name]
	sort.Strings(dependents)

	// Calculate new prefix
	newPrefix := prefix
	if prefix != "" {
		if isLast {
			newPrefix = prefix + "    "
		} else {
			newPrefix = prefix + "│   "
		}
	}

	// Print dependents
	for i, dep := range dependents {
		isLastDep := i == len(dependents)-1
		printResourceTree(w, dep, newPrefix, isLastDep, resources, dependsOn, printed)
	}
}
