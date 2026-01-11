package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// listCommand creates the list subcommand
func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List discovered Kubernetes resources",
		ArgsUsage: "[PATH]",
		Description: `List parses Go source files and displays discovered Kubernetes resources.

If PATH is not specified, the current directory is used.

Examples:
  wetwire-k8s list                      # List from current directory
  wetwire-k8s list ./k8s                # List from specific directory
  wetwire-k8s list -f json              # Output as JSON
  wetwire-k8s list --all                # Include dependency information`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: table, json, or yaml",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show dependency information",
				Value:   false,
			},
		},
		Action: runList,
	}
}

// resourceInfo represents a discovered resource for output
type resourceInfo struct {
	Name         string   `json:"name" yaml:"name"`
	Type         string   `json:"type" yaml:"type"`
	File         string   `json:"file" yaml:"file"`
	Line         int      `json:"line" yaml:"line"`
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// runList executes the list command
func runList(c *cli.Context) error {
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
	resources, err := discoverResourcesForList(absPath)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	// Get output writer
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Check if we have resources
	if len(resources) == 0 {
		fmt.Fprintln(writer, "No resources found")
		return nil
	}

	// Convert to resourceInfo
	showDeps := c.Bool("all")
	infos := make([]resourceInfo, len(resources))
	for i, r := range resources {
		infos[i] = resourceInfo{
			Name: r.Name,
			Type: r.Type,
			File: r.File,
			Line: r.Line,
		}
		if showDeps {
			infos[i].Dependencies = r.Dependencies
		}
	}

	// Output in requested format
	format := strings.ToLower(c.String("format"))
	switch format {
	case "json":
		return outputJSON(writer, infos)
	case "yaml":
		return outputYAML(writer, infos)
	case "table":
		return outputTable(writer, infos, showDeps)
	default:
		return fmt.Errorf("invalid format %q: must be 'table', 'json', or 'yaml'", format)
	}
}

// discoverResourcesForList discovers resources from the given path
func discoverResourcesForList(path string) ([]discover.Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %q: %w", path, err)
	}

	if info.IsDir() {
		return discover.DiscoverDirectory(path)
	}
	return discover.DiscoverFile(path)
}

// outputJSON outputs resources as JSON
func outputJSON(w io.Writer, infos []resourceInfo) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

// outputYAML outputs resources as YAML
func outputYAML(w io.Writer, infos []resourceInfo) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	return encoder.Encode(infos)
}

// outputTable outputs resources as a formatted table
func outputTable(w io.Writer, infos []resourceInfo, showDeps bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Print header
	if showDeps {
		fmt.Fprintln(tw, "NAME\tTYPE\tFILE\tLINE\tDEPENDENCIES")
	} else {
		fmt.Fprintln(tw, "NAME\tTYPE\tFILE\tLINE")
	}

	// Print resources
	for _, info := range infos {
		// Shorten file path for display
		displayFile := shortenPath(info.File)

		if showDeps {
			deps := "-"
			if len(info.Dependencies) > 0 {
				deps = strings.Join(info.Dependencies, ", ")
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", info.Name, info.Type, displayFile, info.Line, deps)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", info.Name, info.Type, displayFile, info.Line)
		}
	}

	return tw.Flush()
}

// shortenPath shortens a file path for display
func shortenPath(path string) string {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}

	// Try to make path relative to current directory
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}

	// If relative path is shorter, use it
	if len(rel) < len(path) {
		return rel
	}

	return path
}
