package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newInitCmd creates the init subcommand
func newInitCmd() *cobra.Command {
	var example bool

	cmd := &cobra.Command{
		Use:   "init [DIR]",
		Short: "Initialize a new wetwire-k8s project",
		Long: `Init creates a new wetwire-k8s project structure with:
- k8s/ directory for Kubernetes resource definitions
- Sample namespace.go file
- .wetwire.yaml configuration file

If DIR is not specified, the current directory is used.

Examples:
  wetwire-k8s init                    # Initialize in current directory
  wetwire-k8s init ./myproject        # Initialize in specific directory
  wetwire-k8s init --example ./app    # Include example resources`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine target directory
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			// Resolve to absolute path
			absDir, err := filepath.Abs(targetDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Create target directory if it doesn't exist
			if err := os.MkdirAll(absDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create k8s directory
			k8sDir := filepath.Join(absDir, "k8s")
			if err := os.MkdirAll(k8sDir, 0755); err != nil {
				return fmt.Errorf("failed to create k8s directory: %w", err)
			}

			// Get output writer
			writer := cmd.OutOrStdout()

			// Create namespace.go (always)
			nsFile := filepath.Join(k8sDir, "namespace.go")
			if err := writeFileIfNotExists(nsFile, namespaceTemplate); err != nil {
				return fmt.Errorf("failed to create namespace.go: %w", err)
			}
			fmt.Fprintf(writer, "Created %s\n", nsFile)

			// Create example files if requested
			if example {
				// Create deployment.go
				deployFile := filepath.Join(k8sDir, "deployment.go")
				if err := writeFileIfNotExists(deployFile, deploymentTemplate); err != nil {
					return fmt.Errorf("failed to create deployment.go: %w", err)
				}
				fmt.Fprintf(writer, "Created %s\n", deployFile)

				// Create service.go
				svcFile := filepath.Join(k8sDir, "service.go")
				if err := writeFileIfNotExists(svcFile, serviceTemplate); err != nil {
					return fmt.Errorf("failed to create service.go: %w", err)
				}
				fmt.Fprintf(writer, "Created %s\n", svcFile)
			}

			// Create .wetwire.yaml
			configFile := filepath.Join(absDir, ".wetwire.yaml")
			if err := writeFileIfNotExists(configFile, configTemplate); err != nil {
				return fmt.Errorf("failed to create .wetwire.yaml: %w", err)
			}
			fmt.Fprintf(writer, "Created %s\n", configFile)

			fmt.Fprintf(writer, "\nInitialized wetwire-k8s project in %s\n", absDir)
			fmt.Fprintf(writer, "Run 'wetwire-k8s build' to generate Kubernetes manifests\n")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&example, "example", "e", false, "Include example resources (deployment, service)")

	return cmd
}

// writeFileIfNotExists writes content to a file only if it doesn't exist
func writeFileIfNotExists(path string, content string) error {
	if _, err := os.Stat(path); err == nil {
		// File exists, don't overwrite
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// Template for namespace.go
const namespaceTemplate = `package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppNamespace defines the namespace for the application
var AppNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-app",
		Labels: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
	},
}
`

// Template for deployment.go (example)
const deploymentTemplate = `package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// replicas is a helper for creating int32 pointers
func replicas(n int32) *int32 {
	return &n
}

// AppDeployment defines the main application deployment
var AppDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "my-app",
		Namespace: AppNamespace.Name,
		Labels: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: replicas(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/name": "my-app",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "my-app",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: "my-app:latest",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8080,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
			},
		},
	},
}
`

// Template for service.go (example)
const serviceTemplate = `package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppService exposes the application deployment
var AppService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "my-app",
		Namespace: AppNamespace.Name,
		Labels: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
	},
	Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			"app.kubernetes.io/name": "my-app",
		},
		Ports: []corev1.ServicePort{
			{
				Name:     "http",
				Port:     80,
				Protocol: corev1.ProtocolTCP,
			},
		},
	},
}
`

// Template for .wetwire.yaml
const configTemplate = `# wetwire-k8s configuration
# See https://github.com/lex00/wetwire-k8s-go for documentation

# Source directory containing Go files with Kubernetes resources
source: k8s

# Output configuration
output:
  # Output format: yaml or json
  format: yaml
  # Output path (use '-' for stdout)
  path: manifests.yaml

# Build options
build:
  # Skip validation checks
  skip_validation: false
`
