# CLAUDE.md

This file provides context for AI assistants working with wetwire-k8s-go.

## Overview

wetwire-k8s-go is a Go implementation of the wetwire pattern for Kubernetes manifests. It enables defining Kubernetes resources using Go code with full type safety, IDE support, and AI-assisted development.

This package is part of the wetwire ecosystem, which follows a flat, declarative coding pattern optimized for AI-human collaboration.

## Repository structure

```
wetwire-k8s-go/
├── cmd/
│   ├── wetwire-k8s/      # Main CLI binary (build, lint, import, validate, design, test, etc.)
│   └── wetwire-k8s-mcp/  # MCP server for Claude Code integration
├── codegen/              # Schema fetching and code generation
│   ├── fetch/           # Kubernetes schema fetching
│   ├── parse/           # Schema parsing
│   └── generate/        # Go code generation
├── docs/                 # Documentation
│   ├── CLI.md           # Command reference
│   ├── FAQ.md           # Common questions
│   ├── LINT_RULES.md    # Lint rule documentation
│   └── .../             # Additional docs (QUICK_START, INTERNALS, etc.)
├── examples/            # Example projects (guestbook, web-service, etc.)
├── internal/            # Internal packages
│   ├── build/          # Build pipeline (6-stage: discover → validate → extract → order → serialize → emit)
│   ├── discover/       # AST-based resource discovery
│   ├── importer/       # YAML to Go code converter
│   ├── lint/           # Lint engine and 25 lint rules
│   ├── roundtrip/      # Round-trip testing infrastructure
│   └── serialize/      # YAML/JSON serialization
└── testdata/            # Test data files

```

## Key concepts

### Wetwire pattern

Wetwire is a flat, declarative coding pattern for infrastructure-as-code:

- **Flat declarations** - Resources are top-level variables/structs, not nested constructors
- **Direct references** - Use `MyPod.Metadata.Name`, not function calls
- **AI-optimized** - Simple, predictable patterns easy for AI to generate
- **Type-safe** - Full IDE support with autocomplete and validation

### Kubernetes resources

All Kubernetes resource types are available as Go structs in the `resources/` package:

```go
import (
    appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"
    corev1 "github.com/lex00/wetwire-k8s-go/resources/core/v1"
)

// Flat declarations
var MyDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name: "my-app",
        Namespace: "default",
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: 3,
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{
                "app": "my-app",
            },
        },
        Template: MyPodTemplate,  // Direct reference
    },
}

var MyPodTemplate = corev1.PodTemplateSpec{
    Metadata: corev1.ObjectMeta{
        Labels: map[string]string{
            "app": "my-app",
        },
    },
    Spec: corev1.PodSpec{
        Containers: []corev1.Container{MyContainer},  // Direct reference
    },
}

var MyContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    Ports: []corev1.ContainerPort{
        {ContainerPort: 80},
    },
}
```

### Resource discovery

Resources are discovered automatically by parsing Go source files:

1. AST parser finds top-level variable declarations
2. Type checker identifies Kubernetes resource types
3. Dependency graph is built from field references
4. Output is generated in dependency order

No manual registration or framework invocation required.

## Common tasks

### Building manifests

```bash
# Build from current directory
wetwire-k8s build

# Build from specific path
wetwire-k8s build ./k8s

# Save to file
wetwire-k8s build -o manifests.yaml
```

### Linting

```bash
# Check for issues
wetwire-k8s lint

# Auto-fix issues
wetwire-k8s lint --fix

# Lint specific path
wetwire-k8s lint ./k8s
```

### Validation

```bash
# Validate against Kubernetes schemas
wetwire-k8s validate

# Validate specific version
wetwire-k8s validate --k8s-version 1.28
```

### Importing existing manifests

```bash
# Import YAML to Go code
wetwire-k8s import -o k8s.go manifests.yaml

# Import with custom package name
wetwire-k8s import --package myapp -o k8s.go manifests.yaml
```

### Design mode (AI-assisted)

```bash
# Interactive design session
wetwire-k8s design "Create a deployment for nginx with 3 replicas"

# Or use Claude Code with MCP server (recommended)
```

### Dependency graph

```bash
# Generate Mermaid diagram
wetwire-k8s graph

# Save to file
wetwire-k8s graph -o graph.md
```

## Kubernetes-specific patterns

### Resource naming

Resources MUST be declared as top-level variables with meaningful names:

```go
// Good
var NginxDeployment = appsv1.Deployment{...}
var RedisService = corev1.Service{...}

// Bad - not top-level
func createDeployment() appsv1.Deployment {
    return appsv1.Deployment{...}
}
```

### Label selectors

Label selectors MUST match pod template labels:

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{
                "app": "myapp",  // Must match template labels
            },
        },
        Template: corev1.PodTemplateSpec{
            Metadata: corev1.ObjectMeta{
                Labels: map[string]string{
                    "app": "myapp",  // Must match selector
                },
            },
        },
    },
}
```

### References between resources

Use direct struct references for relationships:

```go
// Service references Deployment's selector labels
var MyService = corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: MyDeployment.Spec.Selector.MatchLabels,
    },
}

// ConfigMap mounted in Pod
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        Volumes: []corev1.Volume{
            {
                Name: "config",
                VolumeSource: corev1.VolumeSource{
                    ConfigMap: &corev1.ConfigMapVolumeSource{
                        Name: MyConfigMap.Metadata.Name,
                    },
                },
            },
        },
    },
}
```

### Namespace handling

Resources without an explicit namespace default to "default":

```go
// Explicit namespace
var MyPod = corev1.Pod{
    Metadata: corev1.ObjectMeta{
        Name: "my-pod",
        Namespace: "production",
    },
}

// Implicit default namespace
var MyService = corev1.Service{
    Metadata: corev1.ObjectMeta{
        Name: "my-service",
        // No namespace = "default"
    },
}
```

## Gotchas

### Pointer fields

Many Kubernetes fields are pointers. Use `&` for literal values:

```go
// Correct
Replicas: ptrInt32(3),  // Helper function
// Or
replicas := int32(3)
Replicas: &replicas

// Common mistake - type error
Replicas: 3  // Wrong - expects *int32
```

### API versions

Always use the correct API version package:

```go
// Correct
import appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"

// Wrong - different API version
import appsv1beta1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1beta1"
```

### Container ports vs service ports

Container ports and service ports are different types:

```go
// Container port
Container: corev1.Container{
    Ports: []corev1.ContainerPort{
        {ContainerPort: 8080},  // ContainerPort field
    },
}

// Service port
Service: corev1.Service{
    Spec: corev1.ServiceSpec{
        Ports: []corev1.ServicePort{
            {Port: 80, TargetPort: 8080},  // Port and TargetPort fields
        },
    },
}
```

### Immutable fields

Some Kubernetes fields are immutable after creation. The linter flags these with warnings, but cannot prevent runtime errors.

### Required fields

Some Kubernetes resources have required fields. Validation will catch missing fields before deployment.

## Lint rules

All lint rules use the `WK8xxx` prefix (W=Wetwire, K8=Kubernetes):

- **WK8001-WK8099**: General structure and patterns
- **WK8100-WK8199**: Resource-specific rules
- **WK8200-WK8299**: Security and best practices
- **WK8300-WK8399**: Performance and optimization

See [docs/LINT_RULES.md](docs/LINT_RULES.md) for complete rule documentation.

## Testing

Run synthesis tests with different AI personas:

```bash
# Test with all personas
wetwire-k8s test

# Test specific persona
wetwire-k8s test --persona beginner

# Test specific scenario
wetwire-k8s test --scenario deployment
```

## Development workflow

1. Write or generate Go code defining Kubernetes resources
2. Run `wetwire-k8s lint --fix` to auto-fix common issues
3. Run `wetwire-k8s validate` to check against Kubernetes schemas
4. Run `wetwire-k8s build` to generate YAML manifests
5. Deploy with `kubectl apply -f manifests.yaml`

## MCP server integration

The MCP (Model Context Protocol) server is built into the `wetwire-k8s` CLI as a subcommand. It can be used with:
- **Claude Code** - For direct IDE integration
- **Kiro CLI** - For AI-assisted infrastructure design sessions (recommended)

### Claude Code

For Claude Code integration, configure the MCP server:

```json
{
  "mcpServers": {
    "wetwire-k8s": {
      "command": "wetwire-k8s",
      "args": ["mcp"]
    }
  }
}
```

This gives Claude Code access to:
- `wetwire_build` - Generate Kubernetes manifests
- `wetwire_lint` - Check and fix code
- `wetwire_validate` - Validate schemas
- `wetwire_import` - Convert YAML to Go

### Kiro CLI

For AI-assisted design sessions with Kiro CLI, see [K8S-KIRO-CLI.md](docs/K8S-KIRO-CLI.md) for full setup instructions.

Quick start:

```bash
# Install Kiro CLI
curl -fsSL https://cli.kiro.dev/install | bash
kiro-cli login

# Run design session (auto-configures MCP)
wetwire-k8s design --provider kiro "Create a deployment for nginx"
```

## See also

- [Wetwire Specification](https://github.com/lex00/wetwire/docs/WETWIRE_SPEC.md) - Core patterns and philosophy
- [CLI Reference](docs/CLI.md) - Complete command documentation
- [FAQ](docs/FAQ.md) - Common questions
- [Lint Rules](docs/LINT_RULES.md) - All lint rules with examples
