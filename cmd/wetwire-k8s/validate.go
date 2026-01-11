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
	"github.com/urfave/cli/v2"
)

// validateCommand creates the validate subcommand
func validateCommand() *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "Validate Kubernetes manifests using kubeconform",
		ArgsUsage: "<file|directory|->...",
		Description: `Validate parses Kubernetes YAML manifests and validates them
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "schema-location",
				Aliases: []string{"s"},
				Usage:   "Override schema location (use 'default' for default schemas)",
				Value:   "",
			},
			&cli.BoolFlag{
				Name:  "strict",
				Usage: "Enable strict validation (reject unknown fields)",
				Value: false,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format: text, json, or tap",
				Value:   "text",
			},
			&cli.BoolFlag{
				Name:  "from-build",
				Usage: "Validate generated output from Go source (run build first)",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "kubernetes-version",
				Usage: "Kubernetes version to validate against (e.g., 1.29.0)",
				Value: "",
			},
		},
		Action: runValidate,
	}
}

// runValidate executes the validate command
func runValidate(c *cli.Context) error {
	fromBuild := c.Bool("from-build")

	// Check for input files (unless using --from-build which can use current dir)
	if c.NArg() < 1 && !fromBuild {
		return fmt.Errorf("missing input file, directory, or use '-' for stdin")
	}

	// Check for kubeconform installation
	kubeconformPath, err := exec.LookPath("kubeconform")
	if err != nil {
		return fmt.Errorf("kubeconform is not installed or not in PATH. Install it from https://github.com/yannh/kubeconform")
	}

	// Handle --from-build mode
	if fromBuild {
		return runValidateFromBuild(c, kubeconformPath)
	}

	// Collect all files to validate
	var filesToValidate []string
	var stdinData []byte

	for _, arg := range c.Args().Slice() {
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
	args := buildKubeconformArgs(c)

	// Run validation
	if stdinData != nil {
		return runKubeconformWithStdin(c, kubeconformPath, args, stdinData)
	}

	if len(filesToValidate) == 0 {
		return fmt.Errorf("no YAML files found to validate")
	}

	return runKubeconformWithFiles(c, kubeconformPath, args, filesToValidate)
}

// runValidateFromBuild runs the build pipeline and validates the output
func runValidateFromBuild(c *cli.Context, kubeconformPath string) error {
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

	// Run the build pipeline
	result, err := build.Build(absPath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		errWriter := c.App.ErrWriter
		if errWriter == nil {
			errWriter = os.Stderr
		}
		fmt.Fprintln(errWriter, "No resources found to validate")
		return nil
	}

	// Generate YAML output
	yamlOutput, err := generateBuildOutput(result.OrderedResources)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	// Build kubeconform arguments
	args := buildKubeconformArgs(c)

	// Validate the generated YAML
	return runKubeconformWithStdin(c, kubeconformPath, args, yamlOutput)
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

// buildKubeconformArgs builds the arguments for kubeconform
func buildKubeconformArgs(c *cli.Context) []string {
	var args []string

	// Output format
	output := c.String("output")
	if output != "" && output != "text" {
		args = append(args, "-output", output)
	}

	// Strict mode
	if c.Bool("strict") {
		args = append(args, "-strict")
	}

	// Schema location
	schemaLocation := c.String("schema-location")
	if schemaLocation != "" && schemaLocation != "default" {
		args = append(args, "-schema-location", schemaLocation)
	}

	// Kubernetes version
	k8sVersion := c.String("kubernetes-version")
	if k8sVersion != "" {
		args = append(args, "-kubernetes-version", k8sVersion)
	}

	// Summary output for better feedback
	args = append(args, "-summary")

	return args
}

// runKubeconformWithStdin runs kubeconform with stdin input
func runKubeconformWithStdin(c *cli.Context, kubeconformPath string, args []string, data []byte) error {
	// Add stdin flag
	args = append(args, "-")

	cmd := exec.Command(kubeconformPath, args...)
	cmd.Stdin = bytes.NewReader(data)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Write output to app writer
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	if stdout.Len() > 0 {
		writer.Write(stdout.Bytes())
	}

	errWriter := c.App.ErrWriter
	if errWriter == nil {
		errWriter = os.Stderr
	}

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

// runKubeconformWithFiles runs kubeconform with file arguments
func runKubeconformWithFiles(c *cli.Context, kubeconformPath string, args []string, files []string) error {
	// Append files to arguments
	args = append(args, files...)

	cmd := exec.Command(kubeconformPath, args...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Write output to app writer
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	if stdout.Len() > 0 {
		writer.Write(stdout.Bytes())
	}

	errWriter := c.App.ErrWriter
	if errWriter == nil {
		errWriter = os.Stderr
	}

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
