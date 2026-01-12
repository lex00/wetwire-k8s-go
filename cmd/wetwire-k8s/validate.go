package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/spf13/cobra"
)

// newValidateCmd creates the validate subcommand
func newValidateCmd() *cobra.Command {
	var schemaLocation string
	var strict bool
	var output string
	var fromBuild bool
	var kubernetesVersion string

	cmd := &cobra.Command{
		Use:   "validate <file|directory|->...",
		Short: "Validate Kubernetes manifests using kubeconform",
		Long: `Validate parses Kubernetes YAML manifests and validates them
against the Kubernetes OpenAPI specification using kubeconform.

Use '-' as the file path to read from stdin.
Use --from-build to validate generated output from Go source code.

Examples:
  wetwire-k8s validate deployment.yaml          # Validate a single file
  wetwire-k8s validate manifests/               # Validate all YAML files in directory
  wetwire-k8s validate -                        # Validate from stdin
  wetwire-k8s validate --strict deployment.yaml # Strict validation
  wetwire-k8s validate --output json file.yaml  # JSON output format
  wetwire-k8s validate --from-build ./k8s       # Validate build output from Go source`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for input files (unless using --from-build which can use current dir)
			if len(args) < 1 && !fromBuild {
				return fmt.Errorf("missing input file, directory, or use '-' for stdin")
			}

			// Check for kubeconform installation
			kubeconformPath, err := exec.LookPath("kubeconform")
			if err != nil {
				return fmt.Errorf("kubeconform is not installed or not in PATH. Install it from https://github.com/yannh/kubeconform")
			}

			// Handle --from-build mode
			if fromBuild {
				return runValidateFromBuildCobra(cmd, args, kubeconformPath, schemaLocation, strict, output, kubernetesVersion)
			}

			// Collect all files to validate
			var filesToValidate []string
			var stdinData []byte

			for _, arg := range args {
				if arg == "-" {
					// Read from stdin
					data, err := io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read stdin: %w", err)
					}
					stdinData = data
				} else {
					// Check if path exists
					info, err := os.Stat(arg)
					if err != nil {
						return fmt.Errorf("cannot access %s: %w", arg, err)
					}

					if info.IsDir() {
						// Collect all YAML files from directory
						files, err := collectYAMLFiles(arg)
						if err != nil {
							return fmt.Errorf("failed to collect files from %s: %w", arg, err)
						}
						filesToValidate = append(filesToValidate, files...)
					} else {
						filesToValidate = append(filesToValidate, arg)
					}
				}
			}

			// Build kubeconform arguments
			kubeArgs := buildKubeconformArgsCobra(schemaLocation, strict, output, kubernetesVersion)

			// Run validation
			if stdinData != nil {
				return runKubeconformWithStdinCobra(cmd, kubeconformPath, kubeArgs, stdinData)
			}

			if len(filesToValidate) == 0 {
				return fmt.Errorf("no YAML files found to validate")
			}

			return runKubeconformWithFilesCobra(cmd, kubeconformPath, kubeArgs, filesToValidate)
		},
	}

	cmd.Flags().StringVarP(&schemaLocation, "schema-location", "s", "", "Override schema location (use 'default' for default schemas)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation (reject unknown fields)")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format: text, json, or tap")
	cmd.Flags().BoolVar(&fromBuild, "from-build", false, "Validate generated output from Go source (run build first)")
	cmd.Flags().StringVar(&kubernetesVersion, "kubernetes-version", "", "Kubernetes version to validate against (e.g., 1.29.0)")

	return cmd
}

// runValidateFromBuildCobra runs the build pipeline and validates the output
func runValidateFromBuildCobra(cmd *cobra.Command, args []string, kubeconformPath, schemaLocation string, strict bool, output, kubernetesVersion string) error {
	// Determine source path
	sourcePath := "."
	if len(args) > 0 {
		sourcePath = args[0]
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

	// Run the build pipeline
	result, err := build.Build(absPath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		errWriter := cmd.ErrOrStderr()
		fmt.Fprintln(errWriter, "No resources found to validate")
		return nil
	}

	// Generate YAML output
	yamlOutput, err := generateBuildOutput(result.OrderedResources)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	// Build kubeconform arguments
	kubeArgs := buildKubeconformArgsCobra(schemaLocation, strict, output, kubernetesVersion)

	// Validate the generated YAML
	return runKubeconformWithStdinCobra(cmd, kubeconformPath, kubeArgs, yamlOutput)
}

// generateBuildOutput creates YAML from discovered resources
func generateBuildOutput(resources []discover.Resource) ([]byte, error) {
	// Reuse the generateOutput function from build.go
	return generateOutput(resources, "yaml")
}

// collectYAMLFiles collects all YAML files from a directory
func collectYAMLFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// buildKubeconformArgsCobra builds the arguments for kubeconform
func buildKubeconformArgsCobra(schemaLocation string, strict bool, output, kubernetesVersion string) []string {
	var args []string

	// Output format
	if output != "" && output != "text" {
		args = append(args, "-output", output)
	}

	// Strict mode
	if strict {
		args = append(args, "-strict")
	}

	// Schema location
	if schemaLocation != "" && schemaLocation != "default" {
		args = append(args, "-schema-location", schemaLocation)
	}

	// Kubernetes version
	if kubernetesVersion != "" {
		args = append(args, "-kubernetes-version", kubernetesVersion)
	}

	// Summary output for better feedback
	args = append(args, "-summary")

	return args
}

// runKubeconformWithStdinCobra runs kubeconform with stdin input
func runKubeconformWithStdinCobra(cmd *cobra.Command, kubeconformPath string, args []string, data []byte) error {
	// Add stdin flag
	args = append(args, "-")

	execCmd := exec.Command(kubeconformPath, args...)
	execCmd.Stdin = bytes.NewReader(data)

	// Capture output
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	// Write output to app writer
	writer := cmd.OutOrStdout()

	if stdout.Len() > 0 {
		writer.Write(stdout.Bytes())
	}

	errWriter := cmd.ErrOrStderr()

	if stderr.Len() > 0 {
		errWriter.Write(stderr.Bytes())
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("validation failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("kubeconform execution failed: %w", err)
	}

	return nil
}

// runKubeconformWithFilesCobra runs kubeconform with file arguments
func runKubeconformWithFilesCobra(cmd *cobra.Command, kubeconformPath string, args []string, files []string) error {
	// Append files to arguments
	args = append(args, files...)

	execCmd := exec.Command(kubeconformPath, args...)

	// Capture output
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	// Write output to app writer
	writer := cmd.OutOrStdout()

	if stdout.Len() > 0 {
		writer.Write(stdout.Bytes())
	}

	errWriter := cmd.ErrOrStderr()

	if stderr.Len() > 0 {
		errWriter.Write(stderr.Bytes())
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("validation failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("kubeconform execution failed: %w", err)
	}

	return nil
}
