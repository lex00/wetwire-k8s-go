---
title: "Developers"
---


This guide covers everything you need to contribute to wetwire-k8s-go, from setting up your development environment to submitting pull requests.

## Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of Kubernetes resources
- Familiarity with Go syntax and tooling

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/lex00/wetwire-k8s-go.git
cd wetwire-k8s-go
```

### 2. Install Dependencies

```bash
# Download all dependencies
go mod download

# Verify everything works
go build ./...
```

### 3. Install Development Tools

```bash
# Install CLI for local testing
go install ./cmd/wetwire-k8s

# Install MCP server (optional)
go install ./cmd/wetwire-k8s-mcp

# Install useful Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 4. Verify Installation

```bash
# Check CLI works
wetwire-k8s --help

# Run tests
go test ./...

# Build examples
cd examples/guestbook
wetwire-k8s build
```

## Project Structure

```
wetwire-k8s-go/
├── cmd/
│   ├── wetwire-k8s/          # Main CLI application
│   │   ├── main.go           # Entry point
│   │   ├── build.go          # Build command
│   │   ├── lint.go           # Lint command
│   │   ├── import.go         # Import command
│   │   ├── validate.go       # Validate command
│   │   └── ...               # Other commands
│   └── wetwire-k8s-mcp/      # MCP server for Claude Code
│       └── main.go
├── internal/
│   ├── build/                # Build pipeline implementation
│   │   ├── build.go          # Main orchestration
│   │   ├── order.go          # Dependency ordering
│   │   ├── validate.go       # Validation logic
│   │   └── types.go          # Build types
│   ├── discover/             # Resource discovery from AST
│   │   ├── discover.go       # Main discovery logic
│   │   └── types.go          # Discovery types
│   ├── serialize/            # YAML/JSON serialization
│   │   └── serialize.go
│   ├── lint/                 # Linting infrastructure
│   │   ├── lint.go           # Linter core
│   │   ├── rules.go          # Lint rules
│   │   └── formatter.go      # Output formatting
│   ├── importer/             # YAML to Go conversion
│   │   ├── importer.go       # Import logic
│   │   └── types.go
│   └── roundtrip/            # Round-trip testing
│       └── roundtrip.go
├── examples/                  # Example projects
│   ├── guestbook/
│   ├── web-service/
│   └── configmap-secret/
├── docs/                      # Documentation
├── testdata/                  # Test fixtures
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## Running Tests

### Unit Tests

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Run specific package tests:

```bash
go test ./internal/discover
go test ./internal/build
```

### Verbose Test Output

```bash
go test -v ./...
```

### Test Examples

Test that examples build correctly:

```bash
cd examples/guestbook
wetwire-k8s build -o /tmp/guestbook.yaml
cd ../web-service
wetwire-k8s build -o /tmp/web-service.yaml
cd ../configmap-secret
wetwire-k8s build -o /tmp/configmap-secret.yaml
```

### Integration Tests

Run full round-trip tests (YAML -> Go -> YAML):

```bash
go test ./internal/roundtrip -v
```

## Code Style Guidelines

### Go Code Style

Follow standard Go conventions:

- Use `gofmt` for formatting (automatically applied by most editors)
- Use `goimports` for import organization
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

Run formatting:

```bash
# Format all code
gofmt -w .

# Fix imports
goimports -w .
```

### Naming Conventions

- **Package names**: lowercase, single word (e.g., `discover`, `build`)
- **Exported types**: PascalCase (e.g., `ResourceMetadata`)
- **Unexported types**: camelCase (e.g., `resourceCache`)
- **Constants**: PascalCase (e.g., `DefaultNamespace`)
- **Functions**: camelCase or PascalCase depending on export

### Comments

Document all exported types and functions:

```go
// Discover parses Go source files and identifies Kubernetes resources.
// It returns a list of discovered resources and their dependencies.
func Discover(path string) ([]Resource, error) {
    // Implementation...
}
```

Use package comments in doc.go files:

```go
// Package discover implements resource discovery from Go source files.
// It uses the Go AST parser to identify Kubernetes resource declarations
// and build a dependency graph.
package discover
```

### Error Handling

- Return errors, don't panic (except in truly exceptional cases)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use custom error types for errors that need handling

```go
// Good
func process(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open file %s: %w", path, err)
    }
    defer file.Close()
    // ...
}

// Bad - loses error context
func process(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    // ...
}
```

## Adding New Features

### Adding a New CLI Command

1. Create command file in `cmd/wetwire-k8s/`:

```go
// cmd/wetwire-k8s/mycommand.go
package main

import (
    "github.com/spf13/cobra"
)

var myCommandCmd = &cobra.Command{
    Use:   "mycommand [flags] [path]",
    Short: "Brief description",
    Long:  "Detailed description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCommandCmd)

    // Add flags
    myCommandCmd.Flags().StringP("output", "o", "", "Output file")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Add tests:

```go
// cmd/wetwire-k8s/mycommand_test.go
package main

import "testing"

func TestMyCommand(t *testing.T) {
    // Test implementation
}
```

3. Update documentation in `docs/CLI.md`

### Adding a New Lint Rule

1. Add rule definition to `internal/lint/rules.go`:

```go
var rules = []Rule{
    // ... existing rules
    {
        ID:       "WK8999",
        Name:     "my-rule-name",
        Severity: SeverityWarning,
        Message:  "Description of the issue",
        Check:    checkMyRule,
        Fix:      fixMyRule, // Optional
    },
}

func checkMyRule(node ast.Node) []Violation {
    // Check implementation
    var violations []Violation

    // Inspect node and find violations
    if /* condition */ {
        violations = append(violations, Violation{
            Rule:    "WK8999",
            Line:    node.Pos(),
            Message: "Specific violation message",
        })
    }

    return violations
}

func fixMyRule(node ast.Node) (ast.Node, bool) {
    // Fix implementation (optional)
    // Return modified node and true if fixed
    return node, false
}
```

2. Add tests:

```go
// internal/lint/rules_test.go
func TestMyRule(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantViolations bool
    }{
        {
            name: "valid code",
            code: `var X = &appsv1.Deployment{...}`,
            wantViolations: false,
        },
        {
            name: "invalid code",
            code: `func createDeployment() {...}`,
            wantViolations: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            violations := checkMyRule(parseCode(tt.code))
            if (len(violations) > 0) != tt.wantViolations {
                t.Errorf("got %d violations, want violations: %v",
                    len(violations), tt.wantViolations)
            }
        })
    }
}
```

3. Update `docs/LINT_RULES.md`

### Adding Support for New Resource Types

1. Ensure k8s.io/api includes the type:

```bash
go doc k8s.io/api/apps/v1.NewResourceType
```

2. Update discovery patterns in `internal/discover/discover.go` if needed

3. Add example to `examples/` directory

4. Add test cases

## Testing Guidelines

### Writing Good Tests

- **Test one thing per test** - Each test should verify one behavior
- **Use table-driven tests** - Great for testing multiple cases
- **Use meaningful names** - Test names should describe what they test
- **Don't test implementation details** - Test behavior, not internals

Example table-driven test:

```go
func TestDiscover(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        wantResources int
        wantErr       bool
    }{
        {
            name: "single deployment",
            input: `
                package main
                import appsv1 "k8s.io/api/apps/v1"
                var D = &appsv1.Deployment{}
            `,
            wantResources: 1,
            wantErr:       false,
        },
        {
            name:          "empty file",
            input:         `package main`,
            wantResources: 0,
            wantErr:       false,
        },
        {
            name:          "invalid syntax",
            input:         `package main\nvar X = `,
            wantResources: 0,
            wantErr:       true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resources, err := Discover(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if len(resources) != tt.wantResources {
                t.Errorf("got %d resources, want %d", len(resources), tt.wantResources)
            }
        })
    }
}
```

### Test Fixtures

Place test data in `testdata/` directories:

```
internal/discover/testdata/
├── valid_deployment.go
├── multiple_resources.go
└── invalid_syntax.go
```

Load fixtures in tests:

```go
func TestDiscoverFile(t *testing.T) {
    resources, err := DiscoverFile("testdata/valid_deployment.go")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(resources) != 1 {
        t.Errorf("got %d resources, want 1", len(resources))
    }
}
```

## Pull Request Process

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bugfix
```

Branch naming:
- `feature/description` for new features
- `fix/description` for bug fixes
- `docs/description` for documentation
- `refactor/description` for refactoring

### 2. Make Changes

- Write code following style guidelines
- Add tests for new functionality
- Update documentation as needed
- Run tests locally

### 3. Commit Changes

Use conventional commit messages:

```bash
git commit -m "feat: add new command for resource validation"
git commit -m "fix: resolve nil pointer in discovery"
git commit -m "docs: update CLI reference"
git commit -m "test: add tests for import functionality"
```

Commit message format:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

### 4. Run Pre-Commit Checks

```bash
# Format code
gofmt -w .
goimports -w .

# Run linter
golangci-lint run

# Run tests
go test ./...

# Build CLI
go build ./cmd/wetwire-k8s

# Test examples
cd examples/guestbook && wetwire-k8s build
```

### 5. Push and Create PR

```bash
git push origin feature/my-feature
```

Create pull request on GitHub with:
- Clear title describing the change
- Description explaining what and why
- Reference any related issues
- Include test results

### 6. Address Review Comments

- Respond to all comments
- Make requested changes
- Push updates to the same branch
- Re-request review when ready

## CI/CD Pipeline

The project uses GitHub Actions for continuous integration.

### Workflows

**.github/workflows/ci.yml** - Main CI workflow:
- Runs on every push and pull request
- Tests on multiple Go versions
- Runs linters
- Builds examples
- Checks formatting

### Local CI Simulation

Run the same checks locally:

```bash
# Formatting check
gofmt -d .

# Lint
golangci-lint run

# Test
go test -race ./...

# Build
go build ./...
```

## Debugging Tips

### Debugging Discovery

Print AST nodes:

```go
import "go/ast"
import "go/parser"

fset := token.NewFileSet()
node, _ := parser.ParseFile(fset, "file.go", src, parser.ParseComments)
ast.Print(fset, node)
```

### Debugging Build Pipeline

Enable verbose output:

```bash
wetwire-k8s build --verbose
```

Add debug prints in code:

```go
import "log"

log.Printf("DEBUG: discovered %d resources\n", len(resources))
```

### Debugging Tests

Run single test:

```bash
go test -v -run TestSpecificTest ./internal/discover
```

Print test output:

```go
func TestSomething(t *testing.T) {
    t.Logf("Debug info: %v", value)
    // ...
}
```

## Common Development Tasks

### Updating Dependencies

```bash
# Update specific dependency
go get k8s.io/api@v0.30.0

# Update all dependencies
go get -u ./...

# Tidy up
go mod tidy
```

### Regenerating Test Fixtures

```bash
# Generate YAML from examples
cd examples/guestbook
wetwire-k8s build -o testdata/expected.yaml
```

### Benchmarking

Add benchmark tests:

```go
func BenchmarkDiscover(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Discover("testdata/large_file.go")
    }
}
```

Run benchmarks:

```bash
go test -bench=. ./internal/discover
```

## Getting Help

- **Documentation**: Start with docs/ directory
- **Issues**: Check existing GitHub issues
- **Discussions**: Use GitHub Discussions for questions
- **Code**: Read existing code for examples

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

## License

All contributions are licensed under Apache 2.0. By contributing, you agree to license your contributions under the same terms.

## Next Steps

- Read [Internals](/internals/) for architecture details
- Review [Contributing](/contributing/) for contribution guidelines
- Check [examples/](../examples/) for code examples
- See [Quick Start](/quick-start/) for user perspective
