---
title: "Wetwire Kubernetes"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-k8s-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-k8s-go)
[![CI](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/lex00/wetwire-k8s-go/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-k8s-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Generate Kubernetes manifests from Go structs with AI-assisted design.

## Philosophy

Wetwire uses typed constraints to reduce the model capability required for accurate code generation.

**Core hypothesis:** Typed input + smaller model ≈ Semantic input + larger model

The type system and lint rules act as a force multiplier — cheaper models produce quality output when guided by schema-generated types and iterative lint feedback.

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference]({{< relref "/cli" >}}) | Command-line interface |
| [Quick Start]({{< relref "/quick-start" >}}) | Get started in 5 minutes |
| [Examples]({{< relref "/examples" >}}) | Sample Kubernetes projects |
| [FAQ]({{< relref "/faq" >}}) | Frequently asked questions |

## Installation

```bash
go install github.com/lex00/wetwire-k8s-go@latest
```

## Quick Example

```go
var WebApp = appsv1.Deployment{
    Name:     "web",
    Replicas: 3,
    Image:    "nginx:latest",
}
```
