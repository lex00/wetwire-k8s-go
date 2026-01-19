# wetwire-k8s-go

[![CI](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-k8s-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-k8s-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-k8s-go)](https://goreportcard.com/report/github.com/lex00/wetwire-k8s-go)
[![Coverage](https://codecov.io/gh/lex00/wetwire-k8s-go/branch/main/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-k8s-go)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go implementation of the wetwire pattern for Kubernetes manifests. Define Kubernetes resources using native Go code with full type safety, IDE support, and AI-assisted development.

## Overview

wetwire-k8s-go enables defining Kubernetes resources using Go code with:

- **Type safety** - Compile-time validation of resource definitions
- **IDE support** - Full autocomplete, go-to-definition, and refactoring
- **AI-optimized** - Flat, declarative patterns designed for AI generation
- **No DSL** - Native Go code, same language as your application
- **Static analysis** - Lint rules enforce best practices and patterns

## Installation

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

## Quick start

1. Initialize a new project:
```bash
wetwire-k8s init --example
```

2. Define your resources in Go:
```go
package main

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper for pointer values
func ptr[T any](v T) *T { return &v }

// Shared labels for selector and template matching
var nginxLabels = map[string]string{"app": "nginx"}

// NginxContainer defines the nginx container configuration
var NginxContainer = corev1.Container{
    Name:  "nginx",
    Image: "nginx:1.25-alpine",
    Ports: []corev1.ContainerPort{
        {Name: "http", ContainerPort: 80},
    },
}

// NginxPodSpec defines the pod specification
var NginxPodSpec = corev1.PodSpec{
    Containers: []corev1.Container{NginxContainer},
}

// NginxPodTemplate defines the pod template with labels
var NginxPodTemplate = corev1.PodTemplateSpec{
    ObjectMeta: metav1.ObjectMeta{
        Labels: nginxLabels,
    },
    Spec: NginxPodSpec,
}

// NginxDeployment is the main deployment resource
var NginxDeployment = appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "nginx",
        Namespace: "default",
        Labels:    nginxLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
        Selector: &metav1.LabelSelector{
            MatchLabels: nginxLabels,
        },
        Template: NginxPodTemplate,
    },
}
```

3. Build manifests:
```bash
wetwire-k8s build -o manifests.yaml
```

4. Apply to your cluster:
```bash
kubectl apply -f manifests.yaml
```

## Features

- **build** - Generate Kubernetes YAML/JSON manifests from Go code
- **lint** - Enforce flat, declarative patterns and best practices
- **import** - Convert existing YAML manifests to Go code
- **validate** - Validate resources against Kubernetes schemas
- **design** - AI-assisted interactive design mode (with Claude)
- **graph** - Visualize resource dependencies
- **diff** - Compare generated manifests with deployed resources
- **watch** - Auto-rebuild on file changes
- **test** - Run synthesis tests with different AI personas

## Documentation

**Getting Started:**
- [Quick Start](docs/QUICK_START.md) - 5-minute tutorial
- [FAQ](docs/FAQ.md) - Common questions

**Reference:**
- [CLI Reference](docs/CLI.md) - All commands
- [Lint Rules](docs/LINT_RULES.md) - WK8 rule reference

**Advanced:**
- [Internals](docs/INTERNALS.md) - Architecture and extension points
- [Adoption Guide](docs/ADOPTION.md) - Team migration strategies
- [Import Workflow](docs/IMPORT_WORKFLOW.md) - Migrate existing configs

## Why wetwire-k8s-go?

| Traditional YAML | wetwire-k8s-go |
|------------------|----------------|
| No type checking | Compile-time validation |
| Manual editing | IDE autocomplete |
| Copy-paste reuse | Functions and variables |
| Hard to test | Unit testable |
| Brittle for AI | AI-optimized patterns |

## AI-Assisted Design

Use the `design` command for interactive, AI-assisted Kubernetes configuration:

```bash
# No API key required - uses Claude CLI
wetwire-k8s design "Create a deployment for nginx with 3 replicas"
```

The design command uses [Claude CLI](https://claude.ai/download) by default, which requires no API key setup. It falls back to the Anthropic API if Claude CLI is not installed.

### MCP Server for Claude Code

When working inside Claude Code, the MCP server provides direct tool access:

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

This exposes `wetwire_build`, `wetwire_lint`, `wetwire_import`, and `wetwire_validate` tools.

## Examples

See the [examples/](examples/) directory for complete example projects (coming soon).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines (coming soon).

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
