package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Version is set at build time
var Version = "dev"

func main() {
	app := newApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// newApp creates and configures the CLI application
func newApp() *cli.App {
	return &cli.App{
		Name:    "wetwire-k8s",
		Usage:   "Kubernetes manifest synthesis from Go code",
		Version: Version,
		Commands: []*cli.Command{
			buildCommand(),
			importCommand(),
			validateCommand(),
			listCommand(),
			initCommand(),
			graphCommand(),
			diffCommand(),
			watchCommand(),
		},
	}
}
