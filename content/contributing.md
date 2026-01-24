---
title: "Contributing"
---

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

Thank you for your interest in contributing to wetwire-k8s-go! This guide will help you get started with development.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for convenience scripts)

### Clone and Setup

```bash
git clone https://github.com/lex00/wetwire-k8s-go.git
cd wetwire-k8s-go
go mod download
```

### Build

```bash
go build ./cmd/wetwire-k8s
```

### Run Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feat/your-feature
```

Branch naming conventions:
- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Test additions/changes

### 2. Make Your Changes

Follow the code style and patterns established in the codebase:

- Use standard Go formatting (`go fmt`)
- Follow effective Go guidelines
- Write tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/discover/...

# Run with race detection
go test -race ./...

# Run with verbose output
go test -v ./...
```

### 4. Lint Your Code

```bash
# Format code
go fmt ./...

# Run vet
go vet ./...

# Run staticcheck (if installed)
staticcheck ./...
```

### 5. Commit Your Changes

Write clear commit messages following conventional commits:

```
feat: add support for StatefulSet resources

- Add StatefulSet to recognized types in discover.go
- Add tests for StatefulSet discovery
- Update documentation
```

Commit types:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `style:` - Formatting, no code change
- `refactor:` - Code restructuring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

### 6. Push and Create Pull Request

```bash
git push origin feat/your-feature
```

Then create a pull request on GitHub.

## Code Organization

### Package Structure

```
wetwire-k8s-go/
├── cmd/                    # Command-line binaries
│   ├── wetwire-k8s/       # Main CLI
│   └── wetwire-k8s-mcp/   # MCP server
├── internal/               # Internal packages (not importable)
│   ├── build/             # Build pipeline
│   ├── discover/          # Resource discovery
│   ├── serialize/         # YAML/JSON serialization
│   ├── lint/              # Linting infrastructure
│   ├── importer/          # YAML to Go conversion
│   └── roundtrip/         # Roundtrip testing
├── codegen/               # Code generation utilities
├── docs/                  # Documentation
├── examples/              # Example projects
└── testdata/              # Test data files
```

### Key Files

- `cmd/wetwire-k8s/main.go` - CLI entry point
- `internal/discover/discover.go` - Resource discovery logic
- `internal/build/build.go` - Build pipeline orchestration
- `internal/serialize/serialize.go` - Serialization logic

## Adding New Features

### Adding a New Command

1. Create a new file in `cmd/wetwire-k8s/`:

```go
// cmd/wetwire-k8s/mycommand.go
package main

import "github.com/urfave/cli/v2"

func myCommand() *cli.Command {
    return &cli.Command{
        Name:  "mycommand",
        Usage: "Description of what it does",
        Flags: []cli.Flag{
            // Add flags here
        },
        Action: runMyCommand,
    }
}

func runMyCommand(c *cli.Context) error {
    // Implementation
    return nil
}
```

2. Register in `main.go`:

```go
app.Commands = []*cli.Command{
    buildCommand(),
    lintCommand(),
    myCommand(),  // Add here
}
```

3. Add tests in `cmd/wetwire-k8s/mycommand_test.go`

4. Update CLI documentation in `docs/CLI.md`

### Adding a New Lint Rule

1. Add the rule in `internal/lint/rules.go`:

```go
var WK8XXX = Rule{
    ID:       "WK8XXX",
    Name:     "my-rule-name",
    Severity: Warning,
    Check: func(ctx *Context, node ast.Node) []Issue {
        // Check logic
        return nil
    },
    Fix: func(ctx *Context, issue Issue) error {
        // Optional auto-fix
        return nil
    },
}
```

2. Register the rule:

```go
var AllRules = []Rule{
    // ...existing rules...
    WK8XXX,
}
```

3. Add test cases in `internal/lint/testdata/`:
   - `wk8xxx_bad.go` - Code that should trigger the rule
   - `wk8xxx_good.go` - Code that should pass

4. Add tests in `internal/lint/rules_test.go`

5. Document in `docs/LINT_RULES.md`

### Adding Support for New Resource Types

1. Update `internal/discover/discover.go`:

```go
// Add to k8sPackages if new package
k8sPackages := []string{
    "corev1", "appsv1", "batchv1", "networkingv1", "rbacv1",
    "mynewpkg",  // Add new package
}

// Add to k8sTypes if new type
k8sTypes := []string{
    "Pod", "Service", "Deployment",
    "MyNewType",  // Add new type
}
```

2. Update `cmd/wetwire-k8s/build.go` to map the package:

```go
packageMap := map[string]string{
    "corev1":    "v1",
    "appsv1":    "apps/v1",
    "mynewpkg":  "my.api.group/v1",  // Add mapping
}
```

3. Add test cases in `internal/discover/testdata/`

## Testing Guidelines

### Test File Naming

- `*_test.go` - Unit tests for the corresponding file
- `integration_test.go` - Integration tests

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "valid input",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "invalid input",
            input:   invalidInput,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Using testdata

Place test files in `testdata/` directories:

```go
func TestWithTestdata(t *testing.T) {
    content, err := os.ReadFile("testdata/example.go")
    if err != nil {
        t.Fatal(err)
    }
    // Use content in test
}
```

## Documentation

### Code Comments

- Export comments for all public types, functions, and methods
- Use complete sentences
- Start with the name being documented

```go
// Discover finds all Kubernetes resource declarations in Go source files.
// It returns a slice of Resource structs containing metadata about each
// discovered resource, including dependencies.
func Discover(path string) ([]Resource, error) {
    // ...
}
```

### README Updates

Update README.md for:
- New features
- Changed behavior
- New examples

### Documentation Files

Update relevant docs in `docs/`:
- `CLI.md` - Command reference
- `FAQ.md` - Common questions
- `LINT_RULES.md` - Lint rule documentation
- `TROUBLESHOOTING.md` - Known issues and solutions

## Pull Request Process

### Before Submitting

1. Ensure all tests pass: `go test ./...`
2. Format code: `go fmt ./...`
3. Run vet: `go vet ./...`
4. Update documentation
5. Write a clear PR description

### PR Description Template

```markdown
## Summary

Brief description of changes.

## Changes

- List of specific changes
- Another change

## Testing

How the changes were tested.

## Documentation

- [ ] Updated CLI.md
- [ ] Updated CHANGELOG.md
- [ ] Added/updated tests
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback
3. Once approved, a maintainer will merge

## Release Process

Releases are managed by maintainers:

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag
4. Push tag to trigger release workflow

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Getting Help

- File issues for bugs or feature requests
- Ask questions in discussions
- Check existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
