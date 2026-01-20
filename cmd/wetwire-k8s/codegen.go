package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-k8s-go/codegen/crd"
	"github.com/lex00/wetwire-k8s-go/codegen/generate"
	"github.com/spf13/cobra"
)

// newCodegenCmd creates the codegen command with subcommands
func newCodegenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codegen",
		Short: "Generate Go code from Kubernetes schemas",
		Long: `Code generation commands for creating Go types from Kubernetes schemas.

Subcommands:
  crd     Generate Go types from CRD (Custom Resource Definition) files`,
	}

	cmd.AddCommand(newCodegenCRDCmd())

	return cmd
}

// newCodegenCRDCmd creates the codegen crd subcommand
func newCodegenCRDCmd() *cobra.Command {
	var outputDir string
	var domain string

	cmd := &cobra.Command{
		Use:   "crd <source>",
		Short: "Generate Go types from CRD files",
		Long: `Generate Go types from Kubernetes CRD (Custom Resource Definition) files.

The source can be:
  - A directory containing CRD YAML files (e.g., ./crds)
  - "config-connector" to fetch Google Config Connector CRDs
  - A URL to fetch CRDs from

Examples:
  # Generate from local CRD directory
  wetwire-k8s codegen crd ./crds --domain myapp --output ./resources

  # Generate from Config Connector CRDs
  wetwire-k8s codegen crd config-connector --domain cnrm --output ./resources

The output structure will be:
  {output}/{domain}/{group}/{version}/{kind}.go

For example:
  ./resources/cnrm/compute/v1beta1/computeinstance.go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			errWriter := cmd.ErrOrStderr()

			// Parse the source
			crdSource := crd.ParseCRDSource(source)

			var crdDir string
			var err error

			switch crdSource.Type {
			case "directory":
				crdDir = crdSource.Path
				// Verify directory exists
				if _, err := os.Stat(crdDir); os.IsNotExist(err) {
					return fmt.Errorf("CRD directory does not exist: %s", crdDir)
				}

			case "github":
				fmt.Fprintf(errWriter, "Fetching CRDs from GitHub...\n")
				fetcher := crd.NewFetcher("")
				crdDir, err = fetcher.FetchConfigConnector(cmd.Context())
				if err != nil {
					return fmt.Errorf("fetch CRDs: %w", err)
				}
				fmt.Fprintf(errWriter, "Downloaded CRDs to %s\n", crdDir)

			case "url":
				return fmt.Errorf("URL source not yet implemented: %s", source)

			default:
				return fmt.Errorf("unknown source type: %s", source)
			}

			// Create output directory
			if outputDir == "" {
				outputDir = "resources"
			}
			absOutput, err := filepath.Abs(outputDir)
			if err != nil {
				return fmt.Errorf("resolve output path: %w", err)
			}

			if err := os.MkdirAll(absOutput, 0755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			// Generate code
			fmt.Fprintf(errWriter, "Generating Go types from CRDs in %s...\n", crdDir)
			generator := generate.NewCRDGenerator(absOutput, domain)
			if err := generator.GenerateFromCRDDirectory(crdDir); err != nil {
				return fmt.Errorf("generate code: %w", err)
			}

			fmt.Fprintf(errWriter, "Output written to %s\n", absOutput)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "resources", "Output directory for generated code")
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain name for generated packages (required)")
	cmd.MarkFlagRequired("domain")

	return cmd
}
