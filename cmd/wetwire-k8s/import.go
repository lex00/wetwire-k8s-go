package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-k8s-go/internal/importer"
	"github.com/urfave/cli/v2"
)

// importCommand creates the import subcommand
func importCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Convert Kubernetes YAML manifests to Go code",
		ArgsUsage: "<file>",
		Description: `Import reads YAML manifests and generates Go source code
using the wetwire pattern.

Use '-' as the file path to read from stdin.

Examples:
  wetwire-k8s import deployment.yaml           # Convert YAML to Go
  wetwire-k8s import -o k8s.go deployment.yaml # Save to file
  wetwire-k8s import -p myapp deployment.yaml  # Use custom package name
  cat manifests.yaml | wetwire-k8s import -    # Read from stdin`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (default: stdout)",
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "package",
				Aliases: []string{"p"},
				Usage:   "Go package name",
				Value:   "main",
			},
			&cli.StringFlag{
				Name:  "var-prefix",
				Usage: "Prefix for generated variable names",
				Value: "",
			},
		},
		Action: runImport,
	}
}

// runImport executes the import command
func runImport(c *cli.Context) error {
	// Check for input file argument
	if c.NArg() < 1 {
		return fmt.Errorf("missing input file (use '-' for stdin)")
	}

	inputFile := c.Args().First()

	// Read input
	var inputData []byte
	var err error

	if inputFile == "-" {
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	} else {
		inputData, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Configure importer
	opts := importer.Options{
		PackageName: c.String("package"),
		VarPrefix:   c.String("var-prefix"),
	}

	// Run import
	result, err := importer.ImportBytes(inputData, opts)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Output warnings to stderr
	errWriter := c.App.ErrWriter
	if errWriter == nil {
		errWriter = os.Stderr
	}
	for _, warn := range result.Warnings {
		fmt.Fprintf(errWriter, "warning: %s\n", warn)
	}

	// Determine output destination
	outputPath := c.String("output")

	if outputPath == "" || outputPath == "-" {
		// Write to stdout
		writer := c.App.Writer
		if writer == nil {
			writer = os.Stdout
		}
		fmt.Fprint(writer, result.GoCode)
	} else {
		// Create output directory if needed
		if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Write to file
		if err := os.WriteFile(outputPath, []byte(result.GoCode), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		fmt.Fprintf(errWriter, "Imported %d resources to %s\n", result.ResourceCount, outputPath)
	}

	return nil
}
