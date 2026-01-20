package main

import (
	"fmt"
	"os"

	"github.com/lex00/wetwire-k8s-go/domain"
)

// Version is set at build time
var Version = "dev"

func main() {
	// Set the version in the domain
	domain.Version = Version

	// Create the domain and root command
	d := &domain.K8sDomain{}
	rootCmd := domain.CreateRootCommand(d)

	// Add custom commands that are not part of the standard domain interface
	rootCmd.AddCommand(
		newImportCmd(),
		newDiffCmd(),
		newWatchCmd(),
		newTestCmd(),
		newDesignCmd(),
		newMCPCmd(),
		newCodegenCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
