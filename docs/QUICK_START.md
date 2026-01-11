# Quick Start Guide

Get started with wetwire-k8s-go in 5 minutes. This guide walks you through defining Kubernetes resources in Go and generating YAML manifests.

## Prerequisites

- Go 1.21 or later
- Basic familiarity with Kubernetes resources
- Optional: kubectl for applying manifests

## Installation

Install the wetwire-k8s CLI:

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

Verify the installation:

```bash
wetwire-k8s --help
```

## Step 1: Initialize a Project

Create a new directory and initialize a Go module:

```bash
mkdir my-k8s-app
cd my-k8s-app
go mod init github.com/myorg/my-k8s-app
```

Add the wetwire-k8s-go dependency:

```bash
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
```

## Step 2: Define Your First Resource

Create a file named `main.go` with a simple Deployment:

```go
package main

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function for int32 pointers
func ptr[T any](v T) *T { return &v }

// NginxDeployment defines a simple nginx deployment
var NginxDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "nginx",
        Namespace: "default",
        Labels: map[string]string{
            "app": "nginx",
        },
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
        Selector: &metav1.LabelSelector{
            MatchLabels: map[string]string{
                "app": "nginx",
            },
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: map[string]string{
                    "app": "nginx",
                },
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "nginx",
                        Image: "nginx:1.21",
                        Ports: []corev1.ContainerPort{
                            {ContainerPort: 80},
                        },
                    },
                },
            },
        },
    },
}
```

## Step 3: Build the Manifest

Generate YAML from your Go code:

```bash
wetwire-k8s build
```

This outputs the Kubernetes YAML manifest to stdout. Save it to a file:

```bash
wetwire-k8s build -o manifests.yaml
```

## Step 4: Deploy to Kubernetes

Apply the generated manifest to your cluster:

```bash
kubectl apply -f manifests.yaml
```

## Adding More Resources

The wetwire pattern encourages flat, declarative resource definitions. Add a Service in the same file:

```go
// NginxService exposes the nginx deployment
var NginxService = &corev1.Service{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "nginx",
        Namespace: "default",
    },
    Spec: corev1.ServiceSpec{
        Selector: NginxDeployment.Spec.Selector.MatchLabels, // Reference the deployment's labels
        Ports: []corev1.ServicePort{
            {
                Port:       80,
                TargetPort: intstr.FromInt(80),
            },
        },
        Type: corev1.ServiceTypeClusterIP,
    },
}
```

Note how the Service references the Deployment's labels directly. This creates a dependency that wetwire-k8s tracks automatically.

## Key Concepts

### Flat Declarations

Resources are defined as top-level package variables, not nested in functions:

```go
// Good - flat, top-level declaration
var MyDeployment = &appsv1.Deployment{...}

// Bad - nested in function
func createDeployment() *appsv1.Deployment {
    return &appsv1.Deployment{...}
}
```

### Direct References

Reference other resources directly instead of duplicating values:

```go
// Good - direct reference
Selector: MyDeployment.Spec.Selector.MatchLabels

// Bad - duplicated values
Selector: map[string]string{"app": "myapp"}
```

### Type Safety

Use the official k8s.io/api types for full IDE support:

- Autocomplete for all fields
- Compile-time type checking
- Go-to-definition for types

## Linting

Check your code follows wetwire patterns:

```bash
wetwire-k8s lint
```

Auto-fix common issues:

```bash
wetwire-k8s lint --fix
```

## Validation

Validate resources against Kubernetes schemas:

```bash
wetwire-k8s validate
```

Validate for a specific Kubernetes version:

```bash
wetwire-k8s validate --k8s-version 1.28
```

## Next Steps

- See [examples/](../examples/) for complete working examples
- Read [INTERNALS.md](INTERNALS.md) for architecture details
- Check [CLI.md](CLI.md) for complete command reference
- Review [LINT_RULES.md](LINT_RULES.md) for coding standards

## Common Patterns

### Pointer Helpers

Many Kubernetes fields require pointers. Use a generic helper:

```go
func ptr[T any](v T) *T { return &v }

// Usage
Replicas: ptr(int32(3))
```

### Sharing Labels

Extract labels to avoid duplication:

```go
var appLabels = map[string]string{
    "app":     "myapp",
    "version": "v1",
}

var MyDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Labels: appLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{
            MatchLabels: appLabels,
        },
        // ...
    },
}
```

### Multi-Environment Configurations

Create separate files for different environments:

```
k8s/
  common.go      # Shared resources
  dev.go         # Development overrides
  prod.go        # Production overrides
```

Build specific environments using Go build tags or separate directories.
