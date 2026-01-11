package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-k8s-go/internal/importer"
)

const version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "import":
		return runImport(cmdArgs, stdout, stderr)
	case "build":
		fmt.Fprintln(stdout, "wetwire-k8s build - Build Kubernetes manifests from Go code")
		fmt.Fprintln(stdout, "(Not yet fully implemented)")
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "wetwire-k8s version %s\n", version)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "wetwire-k8s - Kubernetes manifest synthesis")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  wetwire-k8s <command> [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  build     Build Kubernetes manifests from Go code")
	fmt.Fprintln(w, "  import    Convert YAML manifests to Go code")
	fmt.Fprintln(w, "  help      Show this help message")
	fmt.Fprintln(w, "  version   Show version information")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run 'wetwire-k8s <command> -h' for more information on a command.")
}

func runImport(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		output      string
		packageName string
		varPrefix   string
	)

	fs.StringVar(&output, "o", "", "Output file path (default: stdout)")
	fs.StringVar(&output, "output", "", "Output file path (default: stdout)")
	fs.StringVar(&packageName, "p", "main", "Go package name")
	fs.StringVar(&packageName, "package", "main", "Go package name")
	fs.StringVar(&varPrefix, "var-prefix", "", "Prefix for generated variable names")

	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: wetwire-k8s import [options] <file>")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Convert Kubernetes YAML manifests to Go code.")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Arguments:")
		fmt.Fprintln(stderr, "  <file>    Path to YAML file (use '-' for stdin)")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "error: missing input file")
		fs.Usage()
		return 2
	}

	inputFile := fs.Arg(0)

	var inputData []byte
	var err error

	if inputFile == "-" {
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "error reading stdin: %v\n", err)
			return 1
		}
	} else {
		inputData, err = os.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading file: %v\n", err)
			return 1
		}
	}

	opts := importer.Options{
		PackageName: packageName,
		VarPrefix:   varPrefix,
	}

	result, err := importer.ImportBytes(inputData, opts)
	if err != nil {
		fmt.Fprintf(stderr, "error importing: %v\n", err)
		return 1
	}

	for _, warn := range result.Warnings {
		fmt.Fprintf(stderr, "warning: %s\n", warn)
	}

	if output == "" || output == "-" {
		fmt.Fprint(stdout, result.GoCode)
	} else {
		if dir := filepath.Dir(output); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(stderr, "error creating output directory: %v\n", err)
				return 1
			}
		}

		if err := os.WriteFile(output, []byte(result.GoCode), 0644); err != nil {
			fmt.Fprintf(stderr, "error writing output: %v\n", err)
			return 1
		}

		fmt.Fprintf(stderr, "Imported %d resources to %s\n", result.ResourceCount, output)
	}

	return 0
}
