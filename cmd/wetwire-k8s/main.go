package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "wetwire-k8s",
		Short:   "Kubernetes manifest synthesis from Go code",
		Version: Version,
		Long: `wetwire-k8s generates Kubernetes YAML manifests from Go code.

Define your Kubernetes resources using native Go syntax:

    var MyDeployment = &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: "my-app",
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: replicas(3),
            ...
        },
    }

Then generate manifests:

    wetwire-k8s build ./k8s`,
	}

	rootCmd.AddCommand(
		newBuildCmd(),
		newImportCmd(),
		newValidateCmd(),
		newLintCmd(),
		newListCmd(),
		newInitCmd(),
		newGraphCmd(),
		newDiffCmd(),
		newWatchCmd(),
		newTestCmd(),
		newDesignCmd(),
		newMCPCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
