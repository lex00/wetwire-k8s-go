# wetwire-k8s-go

[![CI](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-k8s-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-k8s-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-k8s-go)](https://goreportcard.com/report/github.com/lex00/wetwire-k8s-go)
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
    appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"
    corev1 "github.com/lex00/wetwire-k8s-go/resources/core/v1"
)

var NginxDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name:      "nginx",
        Namespace: "default",
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(3),
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{"app": "nginx"},
        },
        Template: corev1.PodTemplateSpec{
            Metadata: corev1.ObjectMeta{
                Labels: map[string]string{"app": "nginx"},
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "nginx",
                        Image: "nginx:latest",
                        Ports: []corev1.ContainerPort{
                            {ContainerPort: 80},
                        },
                    },
                },
            },
        },
    },
}

func ptrInt32(i int32) *int32 { return &i }
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

- [Quick Start Guide](docs/QUICK_START.md) - Step-by-step tutorial (coming soon)
- [CLI Reference](docs/CLI.md) - Complete command documentation
- [FAQ](docs/FAQ.md) - Common questions and answers
- [Lint Rules](docs/LINT_RULES.md) - All lint rules with examples
- [CLAUDE.md](CLAUDE.md) - AI assistant context for development

## Why wetwire-k8s-go?

| Traditional YAML | wetwire-k8s-go |
|------------------|----------------|
| No type checking | Compile-time validation |
| Manual editing | IDE autocomplete |
| Copy-paste reuse | Functions and variables |
| Hard to test | Unit testable |
| Brittle for AI | AI-optimized patterns |

## AI-assisted development

Use with [Claude Code](https://claude.ai/claude-code) for AI-assisted development:

1. Install the MCP server:
```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s-mcp@latest
```

2. Configure in Claude Code settings:
```json
{
  "mcpServers": {
    "wetwire-k8s": {
      "command": "wetwire-k8s-mcp"
    }
  }
}
```

3. Ask Claude to generate Kubernetes resources - it has access to build, lint, validate, and import tools.

## Examples

See the [examples/](examples/) directory for complete example projects (coming soon).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines (coming soon).

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
