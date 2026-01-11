# wetwire-k8s-go

Go implementation of wetwire for Kubernetes manifests.

## Overview

wetwire-k8s-go enables defining Kubernetes resources using Go code, with full type safety and IDE support.

## Installation

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

## Quick Start

```go
package main

import (
    "github.com/lex00/wetwire-k8s-go/resources/apps/v1"
    corev1 "github.com/lex00/wetwire-k8s-go/resources/core/v1"
)

func main() {
    deployment := v1.Deployment{
        Metadata: corev1.ObjectMeta{
            Name: "my-app",
        },
        Spec: v1.DeploymentSpec{
            Replicas: 3,
            // ...
        },
    }
}
```

## Commands

- `wetwire-k8s build` - Generate Kubernetes manifests from Go code
- `wetwire-k8s lint` - Lint Kubernetes definitions
- `wetwire-k8s validate` - Validate against Kubernetes schemas
- `wetwire-k8s test` - Run synthesis tests with AI personas

## License

Apache 2.0
