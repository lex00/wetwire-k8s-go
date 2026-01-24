---
title: "CLI"
---


Complete command reference for wetwire-k8s.

## Installation

See [README.md](../README.md#installation) for installation instructions.

## Global flags

All commands support these global flags:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--quiet` | `-q` | Suppress non-error output | `false` |
| `--help` | `-h` | Show help for command | - |

## Commands

### build

Generate Kubernetes YAML manifests from Go code.

```bash
wetwire-k8s build [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory containing Go files (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file path (use `-` for stdout) | stdout |
| `--format` | `-f` | Output format (`yaml` or `json`) | `yaml` |
| `--namespace` | `-n` | Default namespace for resources without explicit namespace | `default` |
| `--validate` | | Validate resources before building | `true` |
| `--k8s-version` | | Target Kubernetes version | `1.28` |

**Exit codes:**

- `0` - Success
- `1` - Build error (parse error, validation error, etc.)
- `2` - Invalid arguments

**Examples:**

```bash
# Build from current directory, output to stdout
wetwire-k8s build

# Build and save to file
wetwire-k8s build -o manifests.yaml

# Build from specific directory
wetwire-k8s build ./k8s

# Build as JSON
wetwire-k8s build -f json -o manifests.json

# Build without validation
wetwire-k8s build --validate=false

# Build for specific Kubernetes version
wetwire-k8s build --k8s-version 1.30
```

**How it works:**

1. Parses Go source files in the specified directory
2. Discovers top-level variable declarations of Kubernetes resource types
3. Builds dependency graph from field references
4. Validates resources against Kubernetes schemas (if enabled)
5. Generates YAML/JSON output in dependency order

---

### lint

Check Go code for wetwire pattern violations and Kubernetes best practices.

```bash
wetwire-k8s lint [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory or file to lint (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--fix` | | Automatically fix issues where possible | `false` |
| `--rules` | | Comma-separated list of rules to enable | all rules |
| `--disable` | | Comma-separated list of rules to disable | none |
| `--severity` | | Minimum severity to report (`error`, `warning`, `info`) | `info` |
| `--format` | `-f` | Output format (`text`, `json`, `github`) | `text` |

**Exit codes:**

- `0` - No issues found or all issues auto-fixed
- `1` - Issues found (with `--fix`, issues that couldn't be auto-fixed)
- `2` - Invalid arguments

**Examples:**

```bash
# Lint current directory
wetwire-k8s lint

# Lint and auto-fix
wetwire-k8s lint --fix

# Lint specific file
wetwire-k8s lint main.go

# Lint with only errors
wetwire-k8s lint --severity error

# Disable specific rules
wetwire-k8s lint --disable WK8001,WK8002

# Output as JSON
wetwire-k8s lint -f json

# GitHub Actions format
wetwire-k8s lint -f github
```

**What it checks:**

- Flat, declarative patterns (no nested constructors, loops, conditionals)
- Top-level resource declarations
- Direct field references (no function calls for dependencies)
- Label selector consistency
- Required field presence
- Security best practices
- Resource limits and requests
- Naming conventions

---

### import

Convert existing Kubernetes YAML manifests to Go code.

```bash
wetwire-k8s import [OPTIONS] FILE
```

**Arguments:**

- `FILE` - Path to YAML or JSON file to import (use `-` for stdin)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output Go file path (use `-` for stdout) | stdout |
| `--package` | `-p` | Go package name | `main` |
| `--var-prefix` | | Prefix for generated variable names | empty |
| `--optimize` | | Apply wetwire pattern optimizations | `true` |

**Exit codes:**

- `0` - Success
- `1` - Import error (parse error, conversion error, etc.)
- `2` - Invalid arguments

**Examples:**

```bash
# Import YAML to stdout
wetwire-k8s import manifests.yaml

# Import and save to file
wetwire-k8s import -o k8s.go manifests.yaml

# Import with custom package
wetwire-k8s import --package myapp -o k8s.go manifests.yaml

# Import from stdin
cat manifests.yaml | wetwire-k8s import -

# Import with variable prefix
wetwire-k8s import --var-prefix Prod -o k8s.go manifests.yaml

# Import without optimizations
wetwire-k8s import --optimize=false -o k8s.go manifests.yaml
```

**How it works:**

1. Parses YAML/JSON Kubernetes manifests
2. Converts to Go struct declarations
3. Extracts shared values to separate variables (if `--optimize`)
4. Generates idiomatic Go code following wetwire patterns
5. Adds necessary imports

**Note:** Import is best-effort. Complex manifests may require manual cleanup. Run `wetwire-k8s lint --fix` after import.

---

### validate

Validate Kubernetes resources against schemas without building.

```bash
wetwire-k8s validate [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory containing Go files (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--k8s-version` | | Kubernetes version for schema validation | `1.28` |
| `--strict` | | Fail on warnings | `false` |
| `--format` | `-f` | Output format (`text`, `json`) | `text` |

**Exit codes:**

- `0` - Valid
- `1` - Validation errors
- `2` - Invalid arguments

**Examples:**

```bash
# Validate current directory
wetwire-k8s validate

# Validate specific directory
wetwire-k8s validate ./k8s

# Validate for specific Kubernetes version
wetwire-k8s validate --k8s-version 1.30

# Strict mode (fail on warnings)
wetwire-k8s validate --strict

# JSON output
wetwire-k8s validate -f json
```

**What it validates:**

- Required fields are present
- Field types match Kubernetes schemas
- API versions are valid
- Resource references are resolvable
- Immutable field constraints
- Field value ranges and patterns

---

### list

List all discovered Kubernetes resources.

```bash
wetwire-k8s list [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory containing Go files (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format (`text`, `json`, `yaml`) | `text` |
| `--kind` | `-k` | Filter by resource kind | all |
| `--namespace` | `-n` | Filter by namespace | all |

**Exit codes:**

- `0` - Success
- `1` - Error discovering resources
- `2` - Invalid arguments

**Examples:**

```bash
# List all resources
wetwire-k8s list

# List Deployments only
wetwire-k8s list --kind Deployment

# List resources in specific namespace
wetwire-k8s list --namespace production

# JSON output
wetwire-k8s list -f json
```

**Output columns (text format):**

- Variable name
- Kind
- Namespace
- Name
- API version

---

### init

Initialize a new wetwire-k8s project.

```bash
wetwire-k8s init [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to initialize project in (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--module` | `-m` | Go module name | derived from directory |
| `--example` | `-e` | Include example resources | `false` |
| `--force` | `-f` | Overwrite existing files | `false` |

**Exit codes:**

- `0` - Success
- `1` - Initialization error
- `2` - Invalid arguments

**Examples:**

```bash
# Initialize in current directory
wetwire-k8s init

# Initialize with module name
wetwire-k8s init --module github.com/myorg/myproject

# Initialize with examples
wetwire-k8s init --example

# Initialize in specific directory
wetwire-k8s init ./my-project

# Force overwrite existing files
wetwire-k8s init --force
```

**What it creates:**

- `go.mod` with wetwire-k8s dependency
- `main.go` with basic example (if `--example`)
- `.gitignore` with common patterns
- `README.md` with quick start guide

---

### graph

Generate a dependency graph visualization of Kubernetes resources.

```bash
wetwire-k8s graph [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory containing Go files (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file path (use `-` for stdout) | stdout |
| `--format` | `-f` | Output format (`mermaid`, `dot`, `json`) | `mermaid` |
| `--include-fields` | | Include field-level dependencies | `false` |

**Exit codes:**

- `0` - Success
- `1` - Graph generation error
- `2` - Invalid arguments

**Examples:**

```bash
# Generate Mermaid graph to stdout
wetwire-k8s graph

# Save to file
wetwire-k8s graph -o graph.md

# Generate DOT format
wetwire-k8s graph -f dot -o graph.dot

# Include field-level dependencies
wetwire-k8s graph --include-fields

# Generate JSON
wetwire-k8s graph -f json
```

**Output formats:**

- `mermaid` - Mermaid diagram (for GitHub, documentation)
- `dot` - Graphviz DOT format (for visualization tools)
- `json` - Structured JSON (for programmatic analysis)

---

### diff

Show differences between generated manifests and deployed resources.

```bash
wetwire-k8s diff [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory containing Go files (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--context` | `-c` | Kubernetes context to use | current context |
| `--namespace` | `-n` | Namespace to compare | all namespaces |
| `--format` | `-f` | Output format (`text`, `json`) | `text` |

**Exit codes:**

- `0` - No differences
- `1` - Differences found or error
- `2` - Invalid arguments

**Examples:**

```bash
# Diff against current cluster
wetwire-k8s diff

# Diff specific namespace
wetwire-k8s diff --namespace production

# Diff with specific context
wetwire-k8s diff --context staging

# JSON output
wetwire-k8s diff -f json
```

**Note:** Requires `kubectl` to be configured and accessible.

---

### watch

Watch Go files for changes and rebuild automatically.

```bash
wetwire-k8s watch [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to directory to watch (default: current directory)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file path for builds | stdout |
| `--lint` | | Run linter on changes | `true` |
| `--validate` | | Run validation on changes | `true` |
| `--debounce` | | Debounce duration for changes | `500ms` |

**Exit codes:**

- `0` - Exited normally (Ctrl+C)
- `1` - Watch error
- `2` - Invalid arguments

**Examples:**

```bash
# Watch current directory
wetwire-k8s watch

# Watch and save to file
wetwire-k8s watch -o manifests.yaml

# Watch without linting
wetwire-k8s watch --lint=false

# Watch with custom debounce
wetwire-k8s watch --debounce 1s
```

**How it works:**

1. Watches `.go` files in the specified directory
2. On change, runs lint (if enabled), validate (if enabled), and build
3. Outputs results to specified location
4. Continues watching until Ctrl+C

---

### design

AI-assisted interactive design mode for generating Kubernetes resources.

```bash
wetwire-k8s design [OPTIONS] [PROMPT]
```

**Arguments:**

- `PROMPT` - Initial design prompt (optional, can be interactive)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output directory for generated code | current directory |
| `--persona` | `-p` | AI persona (`beginner`, `intermediate`, `expert`) | `intermediate` |
| `--max-cycles` | | Maximum lint/fix cycles | `3` |
| `--interactive` | `-i` | Interactive mode | `true` if no prompt |

**Exit codes:**

- `0` - Success
- `1` - Design error
- `2` - Invalid arguments

**Examples:**

```bash
# Interactive design mode
wetwire-k8s design

# Design with prompt
wetwire-k8s design "Create a deployment for nginx with 3 replicas"

# Design with specific persona
wetwire-k8s design --persona expert "Create a production-ready web app"

# Design and save to specific directory
wetwire-k8s design -o ./k8s "Create a StatefulSet for PostgreSQL"

# Non-interactive with max cycles
wetwire-k8s design --interactive=false --max-cycles 5 "Create a complete web stack"
```

**Environment variables:**

- `ANTHROPIC_API_KEY` - Required for design mode

**How it works:**

1. Takes natural language prompt
2. Generates Go code using AI (Claude)
3. Runs linter and auto-fixes issues
4. Validates against Kubernetes schemas
5. Iterates until passing or max cycles reached
6. Saves generated code to output directory

**Note:** For interactive development, using Claude Code with the MCP server is recommended over the CLI design command.

---

### test

Run synthesis tests with different AI personas.

```bash
wetwire-k8s test [OPTIONS] [PATH]
```

**Arguments:**

- `PATH` - Path to test scenarios (default: `./test/scenarios`)

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--persona` | `-p` | Test specific persona only | all personas |
| `--scenario` | `-s` | Test specific scenario only | all scenarios |
| `--output` | `-o` | Output directory for test results | `./test/results` |
| `--fail-fast` | | Stop on first failure | `false` |

**Exit codes:**

- `0` - All tests passed
- `1` - One or more tests failed
- `2` - Invalid arguments

**Examples:**

```bash
# Run all tests
wetwire-k8s test

# Test specific persona
wetwire-k8s test --persona beginner

# Test specific scenario
wetwire-k8s test --scenario deployment

# Fail fast
wetwire-k8s test --fail-fast

# Custom output directory
wetwire-k8s test -o ./results
```

**Test personas:**

- `beginner` - Asks many questions, needs guidance
- `intermediate` - Balanced, knows basics
- `expert` - Deep knowledge, precise requirements

Custom personas can be registered for domain-specific testing.

**Test scoring:**

Tests are scored on 5 dimensions (0-3 scale each):

1. **Correctness** - Syntactic and semantic correctness
2. **Completeness** - All requirements met
3. **Idiomaticity** - Follows Kubernetes and wetwire patterns
4. **Efficiency** - Resource usage and optimization
5. **Clarity** - Code readability and maintainability

Total score: 0-15 points

---

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ANTHROPIC_API_KEY` | Anthropic API key for design mode | required for design |
| `WETWIRE_K8S_VERSION` | Default Kubernetes version | `1.28` |
| `WETWIRE_K8S_NAMESPACE` | Default namespace | `default` |
| `NO_COLOR` | Disable colored output | `false` |

---

## Configuration file

wetwire-k8s can be configured via `.wetwire-k8s.yaml` in the project root:

```yaml
# Kubernetes version for validation
k8s_version: "1.28"

# Default namespace
namespace: default

# Lint configuration
lint:
  auto_fix: true
  severity: info
  disabled_rules:
    - WK8001
    - WK8002

# Build configuration
build:
  format: yaml
  validate: true

# Design mode configuration
design:
  persona: intermediate
  max_cycles: 3
```

---

## See also

- [FAQ](/faq/) - Common questions
- [Lint Rules](/lint-rules/) - Complete lint rule reference
- [CLAUDE.md](../CLAUDE.md) - AI assistant context
