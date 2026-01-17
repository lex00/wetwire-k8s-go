package main

import (
	"bytes"

	"github.com/spf13/cobra"
)

// newTestRootCmd creates a new root command for testing.
// This replicates the main() setup for use in tests.
func newTestRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "wetwire-k8s",
		Short:   "Kubernetes manifest synthesis from Go code",
		Version: Version,
	}

	rootCmd.AddCommand(
		newImportCmd(),
		newDiffCmd(),
		newWatchCmd(),
		newTestCmd(),
		newDesignCmd(),
	)

	return rootCmd
}

// runTestCommand is a helper to run a command with args and capture output.
func runTestCommand(args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := newTestRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout, stderr, err
}
