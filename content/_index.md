---
title: "Wetwire Kubernetes"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-k8s-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-k8s-go)
[![CI](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-k8s-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/lex00/wetwire-k8s-go/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-k8s-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-k8s-go)](https://goreportcard.com/report/github.com/lex00/wetwire-k8s-go)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Semantic linting for Kubernetes manifests.

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
