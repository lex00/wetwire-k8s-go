---
title: "Internals"
---


This document provides a deep-dive into wetwire-k8s-go's architecture, build pipeline, and design decisions.

## Overview

wetwire-k8s-go converts Go source code containing Kubernetes resource declarations into YAML/JSON manifests. The system parses Go code statically, discovers resources, validates them, and generates output in dependency order.

```
   Go Code       DISCOVER      VALIDATE      EXTRACT       ORDER        EMIT
    (.go)    -->           -->          -->          -->          -->
                    |             |            |            |           |
                    v             v            v            v           v
               Resources     Validated     Runtime      Ordered     YAML/JSON
               Metadata      Graph         Values       List        Output
```

## Build Pipeline Stages

### Stage 1: DISCOVER

**Purpose:** Parse Go source files and identify Kubernetes resource declarations.

**Implementation:** `internal/discover/discover.go`

The discovery stage uses Go's `go/ast` and `go/parser` packages to:

1. Parse all `.go` files in the target directory (excluding `_test.go` files)
2. Find top-level variable declarations (`var X = ...`)
3. Identify variables with Kubernetes resource types
4. Extract dependency information from field references

**Resource Detection:**

Resources are identified by their type, which must be from a known Kubernetes API package:

```go
// Recognized packages
k8sPackages := []string{
    "corev1", "appsv1", "batchv1", "networkingv1", "rbacv1",
    "storagev1", "policyv1", "autoscalingv1", "autoscalingv2",
}
```

**Dependency Detection:**

Dependencies are found by traversing the AST of each resource's initializer and finding references to other top-level variables:

```go
var ConfigMap = &corev1.ConfigMap{...}

var Deployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        EnvFrom: []corev1.EnvFromSource{
                            {
                                ConfigMapRef: &corev1.ConfigMapEnvSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: ConfigMap.Name, // Dependency detected!
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    },
}
```

**Output:** List of `Resource` structs containing:
- Name (variable name)
- Type (e.g., "appsv1.Deployment")
- File path and line number
- List of dependencies (other variable names)

### Stage 2: VALIDATE

**Purpose:** Ensure resource declarations are valid and references are resolvable.

**Implementation:** `internal/build/validate.go`

Validation checks:

1. **Reference Resolution:** All referenced variables must exist
2. **Cycle Detection:** Dependency graph must be acyclic (DAG)
3. **Type Checking:** Resource types must be valid Kubernetes types

**Cycle Detection Algorithm:**

Uses depth-first search with three states (white/gray/black) to detect back edges:

```go
func DetectCycles(resources []discover.Resource) error {
    // Build adjacency map
    // DFS with color states
    // Report cycles if found
}
```

### Stage 3: EXTRACT (Planned)

**Purpose:** Execute the Go code to obtain runtime values of resources.

**Status:** Currently a placeholder. The build command generates stub manifests based on metadata.

**Planned Implementation:**

1. Use `go/types` for type checking
2. Use a Go interpreter or plugin system to evaluate declarations
3. Extract actual field values at runtime
4. Handle expressions, function calls, and computed values

### Stage 4: ORDER

**Purpose:** Sort resources in dependency order for correct deployment.

**Implementation:** `internal/build/order.go`

Uses Kahn's algorithm for topological sorting:

1. Build in-degree map (count of incoming edges)
2. Start with nodes having zero in-degree
3. Remove edges and add newly zero in-degree nodes
4. Repeat until all nodes processed

This ensures resources are created in the correct order (e.g., ConfigMap before Deployment that uses it).

### Stage 5: SERIALIZE

**Purpose:** Convert Go structs to YAML/JSON.

**Implementation:** `internal/serialize/serialize.go`

The serializer:

1. Uses `encoding/json` to marshal structs (respecting JSON tags)
2. Converts to `map[string]interface{}` for flexibility
3. Cleans zero values to produce minimal output
4. Uses `gopkg.in/yaml.v3` for YAML output

**Multi-Document YAML:**

Multiple resources are joined with `---` separators:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
```

### Stage 6: EMIT

**Purpose:** Write output to file(s) or stdout.

**Implementation:** `internal/build/build.go`

Output modes:
- **SingleFile:** All resources in one YAML file
- **SeparateFiles:** Each resource in its own file

## Package Structure

```
wetwire-k8s-go/
├── cmd/
│   ├── wetwire-k8s/          # Main CLI binary
│   │   ├── main.go           # Entry point, CLI setup
│   │   ├── build.go          # build command
│   │   ├── lint.go           # lint command (planned)
│   │   ├── import.go         # import command
│   │   ├── validate.go       # validate command
│   │   ├── graph.go          # graph command
│   │   ├── init.go           # init command
│   │   └── list.go           # list command
│   └── wetwire-k8s-mcp/      # MCP server for Claude Code
│       └── main.go
├── internal/
│   ├── build/                # Build pipeline
│   │   ├── build.go          # Main pipeline orchestration
│   │   ├── order.go          # Topological sorting
│   │   ├── validate.go       # Validation logic
│   │   └── types.go          # Build types
│   ├── discover/             # Resource discovery
│   │   ├── discover.go       # AST parsing, resource detection
│   │   └── types.go          # Discovery types
│   ├── serialize/            # Serialization
│   │   └── serialize.go      # YAML/JSON conversion
│   ├── lint/                 # Linting infrastructure
│   │   ├── lint.go           # Linter orchestration
│   │   ├── rules.go          # Lint rule definitions
│   │   └── formatter.go      # Output formatting
│   ├── importer/             # YAML to Go conversion
│   │   ├── importer.go       # Import logic
│   │   └── types.go          # Import types
│   └── roundtrip/            # Roundtrip testing
│       └── roundtrip.go      # YAML -> Go -> YAML tests
└── codegen/                  # Code generation utilities
    ├── fetch/                # Schema fetching
    ├── parse/                # Schema parsing
    └── generate/             # Code generation
```

## Key Design Decisions

### Static Analysis Over Runtime Execution

wetwire-k8s-go uses static analysis (AST parsing) rather than runtime execution. This provides:

- **Speed:** No compilation or execution required
- **Safety:** No arbitrary code execution
- **Predictability:** Same output for same input

The tradeoff is that computed values (function results, conditionals) cannot be evaluated. This aligns with the wetwire philosophy of flat, declarative definitions.

### Using Official k8s.io/api Types

Resources use official Kubernetes Go types from `k8s.io/api` rather than custom types:

- **Compatibility:** Types match upstream Kubernetes
- **IDE Support:** Full autocomplete and documentation
- **Updates:** New fields automatically available with updates

### Dependency Graph for Ordering

Resources are sorted by dependencies to ensure correct deployment order. This is critical because:

- ConfigMaps/Secrets must exist before Pods reference them
- Services should exist before Ingresses route to them
- Namespaces must exist before namespaced resources

### Minimal Output

The serializer removes zero values to produce clean, minimal YAML:

```go
// Go struct with defaults
Container: corev1.Container{
    Name:  "app",
    Image: "nginx",
    // Many zero-value fields
}

// Generated YAML (minimal)
containers:
  - name: app
    image: nginx
```

## Linting Architecture

**Implementation:** `internal/lint/`

The linter enforces wetwire patterns and Kubernetes best practices:

### Rule Structure

Each rule has:
- **ID:** Unique identifier (e.g., WK8001)
- **Severity:** Error, Warning, or Info
- **Check:** Function that analyzes AST nodes
- **Fix:** Optional function for auto-repair

### Rule Categories

See [Lint Rules](/lint-rules/) for the complete rule index with categories WK8001-WK8399.

### Auto-Fix

Some rules support automatic fixing:

```bash
wetwire-k8s lint --fix
```

The fixer modifies AST nodes and regenerates source code using `go/format`.

## Import Architecture

**Implementation:** `internal/importer/`

The importer converts YAML manifests to Go code:

1. Parse YAML documents
2. Identify resource type from apiVersion/kind
3. Generate Go variable declarations
4. Add necessary imports
5. Apply wetwire optimizations (extract shared values)

### Optimization Example

Input YAML:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myapp
spec:
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
```

Optimized Go:
```go
var appLabels = map[string]string{"app": "myapp"}

var MyDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Labels: appLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{
            MatchLabels: appLabels,
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: appLabels,
            },
        },
    },
}
```

## MCP Server

**Implementation:** `cmd/wetwire-k8s-mcp/`

The Model Context Protocol (MCP) server enables AI assistants (Claude Code) to interact with wetwire-k8s:

### Available Tools

- **build:** Generate manifests
- **lint:** Check code patterns
- **validate:** Validate against schemas
- **import:** Convert YAML to Go
- **graph:** Visualize dependencies

### Communication

Uses stdio for communication with JSON-RPC 2.0 protocol:

```json
{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "build", "arguments": {"path": "./k8s"}}}
```

## Testing Strategy

### Unit Tests

Each package has `*_test.go` files testing individual functions:

```bash
go test ./...
```

### Integration Tests

The `internal/roundtrip/` package tests the full YAML -> Go -> YAML cycle:

1. Parse example YAML files
2. Convert to Go code
3. Build back to YAML
4. Compare output

### Testdata

Example files in `testdata/` and `internal/*/testdata/` provide:

- Valid Go files for discovery testing
- Example YAML for roundtrip testing
- Good/bad examples for lint rule testing

## Performance Considerations

### File Parsing

Go's parser is fast but can be slow for large codebases. Optimizations:

- Skip `_test.go` files early
- Use filepath.Walk for efficient directory traversal
- Cache parsed files when possible

### Memory Usage

AST parsing can be memory-intensive. The discovery phase:

- Processes files one at a time
- Releases AST nodes after extracting metadata
- Uses lightweight Resource structs

### Output Generation

YAML serialization is optimized by:

- Using streaming for large outputs
- Avoiding intermediate string allocations
- Cleaning zero values during serialization (not after)

## Future Enhancements

### Runtime Value Extraction

Complete the EXTRACT stage to evaluate Go expressions and obtain actual values.

### Watch Mode

Implement efficient file watching with debouncing for development workflow.

### Diff Against Cluster

Compare generated manifests with deployed resources using kubectl.

### Plugin System

Allow custom lint rules and transformations via Go plugins.
