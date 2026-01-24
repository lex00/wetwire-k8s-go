---
title: "Adoption"
---

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

This guide helps teams migrate from existing Kubernetes tooling to wetwire-k8s-go. Whether you're coming from kubectl/raw YAML, Helm, kustomize, or other IaC tools, this document provides migration strategies and comparisons.

## Overview

wetwire-k8s-go is designed for teams who want:

- Type-safe Kubernetes resource definitions
- Full IDE support (autocomplete, refactoring, go-to-definition)
- AI-assisted development with Claude Code
- Elimination of YAML editing
- Compile-time validation

## Migration from kubectl/Raw YAML

### Current Workflow

```bash
# Edit YAML files manually
vim deployment.yaml
vim service.yaml

# Apply to cluster
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Or apply directory
kubectl apply -f ./manifests/
```

### Migration Strategy

#### Step 1: Import Existing YAML

Convert your YAML files to Go code:

```bash
# Single file
wetwire-k8s import deployment.yaml -o k8s/deployment.go

# Multiple files
cat manifests/*.yaml | wetwire-k8s import - -o k8s/resources.go

# From cluster
kubectl get all -n myapp -o yaml | wetwire-k8s import - -o k8s/myapp.go
```

#### Step 2: Set Up Project Structure

```bash
# Initialize Go module
mkdir myapp-k8s
cd myapp-k8s
go mod init github.com/myorg/myapp-k8s

# Add dependencies
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest

# Organize files
mkdir k8s
mv resources.go k8s/
```

#### Step 3: Refactor and Improve

```bash
# Run linter
wetwire-k8s lint --fix k8s/

# Build and verify
wetwire-k8s build -o output.yaml
diff original.yaml output.yaml
```

#### Step 4: Integrate into Workflow

```bash
# Build manifests
wetwire-k8s build -o manifests.yaml

# Apply to cluster
kubectl apply -f manifests.yaml

# Or pipe directly
wetwire-k8s build | kubectl apply -f -
```

### New Workflow

```bash
# Edit Go code in your IDE
code k8s/deployment.go

# Build manifests
wetwire-k8s build -o manifests.yaml

# Apply to cluster
kubectl apply -f manifests.yaml
```

### Benefits Over Raw YAML

| Raw YAML | wetwire-k8s-go |
|----------|----------------|
| No validation until kubectl apply | Compile-time type checking |
| Manual editing prone to typos | IDE autocomplete prevents errors |
| Copy-paste for similar resources | Go functions for reuse |
| No refactoring support | Rename variables across files |
| Diff tools show YAML changes | Git diff shows semantic changes |

## Migration from Helm

### Current Workflow

```bash
# Install from chart
helm install myapp ./chart -f values.yaml

# Upgrade release
helm upgrade myapp ./chart -f values.yaml

# Manage multiple environments
helm install myapp-dev ./chart -f values-dev.yaml
helm install myapp-prod ./chart -f values-prod.yaml
```

### Migration Strategy

#### Option A: Convert Rendered Charts

**Best for:** Simple charts without heavy templating

```bash
# Render chart
helm template myapp ./chart -f values.yaml > rendered.yaml

# Import to Go
wetwire-k8s import rendered.yaml -o k8s/myapp.go

# Clean up Helm annotations
# (Remove helm.sh/chart, app.kubernetes.io/managed-by, etc.)
```

#### Option B: Rewrite from Scratch

**Best for:** Complex charts where templates obscure structure

1. Review chart's resource definitions
2. Write Go code using wetwire patterns
3. Extract values to Go constants/variables
4. Use Go conditionals instead of Helm templates

Example transformation:

**Helm template:**

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "myapp.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "myapp.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
        {{- if .Values.resources }}
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
        {{- end }}
```

**wetwire-k8s-go:**

```go
// k8s/deployment.go
package main

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ptr[T any](v T) *T { return &v }

// Configuration (replaces values.yaml)
const (
    appName      = "myapp"
    appVersion   = "1.0.0"
    replicaCount = 3
)

var appLabels = map[string]string{
    "app.kubernetes.io/name":    appName,
    "app.kubernetes.io/version": appVersion,
}

var MyAppDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:   appName,
        Labels: appLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(replicaCount)),
        Selector: &metav1.LabelSelector{
            MatchLabels: appLabels,
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: appLabels,
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  appName,
                        Image: "myrepo/myapp:" + appVersion,
                        Resources: corev1.ResourceRequirements{
                            Requests: corev1.ResourceList{
                                corev1.ResourceCPU:    resource.MustParse("100m"),
                                corev1.ResourceMemory: resource.MustParse("128Mi"),
                            },
                        },
                    },
                },
            },
        },
    },
}
```

#### Step 3: Environment-Specific Configuration

**Helm approach:** Multiple values files

**wetwire approach:** Go build tags or separate packages

```go
// k8s/common/config.go
package common

const AppName = "myapp"

// k8s/dev/deployment.go
//go:build dev

package main

import "github.com/myorg/myapp-k8s/k8s/common"

const (
    namespace = "dev"
    replicas  = 1
)

// k8s/prod/deployment.go
//go:build prod

package main

import "github.com/myorg/myapp-k8s/k8s/common"

const (
    namespace = "production"
    replicas  = 5
)
```

Build for specific environment:

```bash
# Development
go build -tags dev && wetwire-k8s build -o dev.yaml

# Production
go build -tags prod && wetwire-k8s build -o prod.yaml
```

### Benefits Over Helm

| Helm | wetwire-k8s-go |
|------|----------------|
| Template syntax (Go templates) | Native Go code |
| Values files (YAML) | Go constants/variables |
| Functions and helpers | Go functions |
| Tiller/Helm required | No special tooling |
| Chart repository | Git repository |
| Limited type checking | Full type safety |
| Templating errors at runtime | Compile-time errors |

## Migration from kustomize

### Current Workflow

```
# Directory structure
kustomize/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/
    ├── dev/
    │   ├── kustomization.yaml
    │   └── patch-replicas.yaml
    └── prod/
        ├── kustomization.yaml
        └── patch-replicas.yaml

# Build and apply
kustomize build overlays/dev | kubectl apply -f -
kustomize build overlays/prod | kubectl apply -f -
```

### Migration Strategy

#### Step 1: Import Base Resources

```bash
# Import base
kustomize build base | wetwire-k8s import - -o k8s/base.go
```

#### Step 2: Create Environment Variations

Instead of kustomize patches, use Go code:

```go
// k8s/common.go
package main

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ptr[T any](v T) *T { return &v }

// createDeployment creates a deployment with environment-specific settings
func createDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
    labels := map[string]string{"app": name}

    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
            Labels:    labels,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: ptr(replicas),
            Selector: &metav1.LabelSelector{
                MatchLabels: labels,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: labels,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  name,
                            Image: "myapp:latest",
                        },
                    },
                },
            },
        },
    }
}

// k8s/dev.go
//go:build dev

package main

var MyDeployment = createDeployment("myapp", "dev", 1)

// k8s/prod.go
//go:build prod

package main

var MyDeployment = createDeployment("myapp", "production", 5)
```

#### Alternative: Separate Directories

```
k8s/
├── common/
│   └── base.go          # Shared configuration
├── dev/
│   └── resources.go     # Dev resources
└── prod/
    └── resources.go     # Prod resources
```

### Benefits Over kustomize

| kustomize | wetwire-k8s-go |
|-----------|----------------|
| Strategic merge patches | Direct Go code modification |
| JSON patches | Type-safe changes |
| Limited validation | Compile-time validation |
| YAML-based | Go-based |
| Complex overlay rules | Simple Go imports |

## Migration Best Practices

### 1. Start Small

Begin with non-critical resources:

- ConfigMaps and Secrets
- Simple Deployments and Services
- Gradually migrate to StatefulSets, Ingresses, etc.

### 2. Parallel Run

Run both systems in parallel during transition:

```bash
# Old way
kubectl apply -f old-manifests/

# New way (different namespace)
wetwire-k8s build | kubectl apply -n test -f -

# Compare
kubectl diff -f <(wetwire-k8s build)
```

### 3. Automate Validation

Add CI checks:

```yaml
# .github/workflows/validate.yml
- name: Lint
  run: wetwire-k8s lint --severity error

- name: Build
  run: wetwire-k8s build -o manifests.yaml

- name: Validate
  run: kubectl apply --dry-run=client -f manifests.yaml
```

### 4. Document Patterns

Create team guidelines:

- How to structure Go files
- Naming conventions
- Where to put environment-specific config
- How to handle secrets

### 5. Train Team

Ensure team understands:

- Go basics (if coming from YAML-only background)
- wetwire patterns (flat declarations, direct references)
- Build and deployment workflow
- Using the linter

## Common Migration Challenges

### Challenge 1: Dynamic Values

**Problem:** Helm/kustomize allow runtime value injection

**Solution:** Use environment variables or build tags

```go
// Read from environment
var replicas = getEnvInt("REPLICAS", 3)

func getEnvInt(key string, defaultVal int) int32 {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.Atoi(val); err == nil {
            return int32(i)
        }
    }
    return int32(defaultVal)
}
```

### Challenge 2: Complex Templating

**Problem:** Heavy Helm template logic difficult to translate

**Solution:** Rewrite using Go logic

```go
// Instead of Helm if/else
var containers []corev1.Container
if enableMetrics {
    containers = append(containers, metricsContainer)
}
containers = append(containers, appContainer)
```

### Challenge 3: Secret Management

**Problem:** Secrets in YAML files (bad practice anyway)

**Solution:** Use external secret management

```go
// Reference external secrets
var AppSecret = &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name: "app-secret",
        Annotations: map[string]string{
            "external-secrets.io/backend": "vault",
        },
    },
}
```

Or use sealed secrets, SOPS, etc.

### Challenge 4: Team Resistance

**Problem:** Team prefers YAML

**Solution:**
- Show IDE autocomplete benefits
- Demonstrate compile-time validation
- Highlight AI assistance with Claude Code
- Start with volunteers, expand gradually

## Next Steps

- Read [Import Workflow](/import-workflow/) for detailed import instructions
- See [Examples](/examples/) for common patterns
- Check [Developers](/developers/) for development setup
- Review [Quick Start](/quick-start/) for getting started
