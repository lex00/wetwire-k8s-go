package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-k8s-go/internal/importer"
	"github.com/spf13/cobra"
)

// newImportCmd creates the import subcommand
func newImportCmd() *cobra.Command {
	var output string
	var pkgName string
	var varPrefix string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Convert Kubernetes YAML manifests to Go code",
		Long: `Import reads YAML manifests and generates Go source code
using the wetwire pattern.

Use '-' as the file path to read from stdin.

Examples:
  wetwire-k8s import deployment.yaml           # Convert YAML to Go
  wetwire-k8s import -o k8s.go deployment.yaml # Save to file
  wetwire-k8s import -p myapp deployment.yaml  # Use custom package name
  cat manifests.yaml | wetwire-k8s import -    # Read from stdin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]

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
				PackageName: pkgName,
				VarPrefix:   varPrefix,
			}

			// Run import
			result, err := importer.ImportBytes(inputData, opts)
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			// Output warnings to stderr
			errWriter := cmd.ErrOrStderr()
			for _, warn := range result.Warnings {
				fmt.Fprintf(errWriter, "warning: %s\n", warn)
			}

			// Determine output destination
			if output == "" || output == "-" {
				// Write to stdout
				writer := cmd.OutOrStdout()
				fmt.Fprint(writer, result.GoCode)
			} else {
				// Create output directory if needed
				if dir := filepath.Dir(output); dir != "" && dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("failed to create output directory: %w", err)
					}
				}

				// Write to file
				if err := os.WriteFile(output, []byte(result.GoCode), 0644); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}

				fmt.Fprintf(errWriter, "Imported %d resources to %s\n", result.ResourceCount, output)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&pkgName, "package", "p", "main", "Go package name")
	cmd.Flags().StringVar(&varPrefix, "var-prefix", "", "Prefix for generated variable names")

	return cmd
}
