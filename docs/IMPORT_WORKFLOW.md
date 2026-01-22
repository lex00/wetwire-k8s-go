<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

This guide walks through the complete workflow for importing existing Kubernetes YAML manifests into wetwire-k8s-go code.

## Overview

The `wetwire-k8s import` command converts existing YAML/JSON manifests to idiomatic Go code that follows wetwire patterns. This is essential for:

- Migrating from kubectl/YAML workflows to wetwire-k8s-go
- Converting Helm chart outputs to Go code
- Importing kustomize-generated manifests
- Onboarding existing Kubernetes resources

## Basic Import Workflow

### Step 1: Prepare Your YAML

Ensure your YAML files are valid Kubernetes manifests:

```bash
# Validate with kubectl
kubectl apply --dry-run=client -f my-resources.yaml

# Or use kubeconform
kubeconform my-resources.yaml
```

### Step 2: Run Import Command

Import a single YAML file:

```bash
wetwire-k8s import my-resources.yaml
```

This outputs Go code to stdout. To save to a file:

```bash
wetwire-k8s import -o k8s.go my-resources.yaml
```

### Step 3: Review Generated Code

The importer generates:

1. **Package declaration** - Defaults to `main`
2. **Imports** - All necessary k8s.io/api packages
3. **Helper functions** - Generic pointer helper `ptr[T any](v T) *T`
4. **Resource variables** - Top-level `var` declarations for each resource
5. **Shared values** - Extracted common labels, selectors, etc. (if `--optimize`)

Example output:

```go
package main

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ptr[T any](v T) *T { return &v }

var appLabels = map[string]string{
    "app": "nginx",
}

var NginxDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:   "nginx",
        Labels: appLabels,
    },
    // ... rest of deployment spec
}
```

### Step 4: Run Linter and Fix Issues

After importing, run the linter to identify pattern violations:

```bash
wetwire-k8s lint k8s.go
```

Auto-fix issues where possible:

```bash
wetwire-k8s lint --fix k8s.go
```

### Step 5: Build and Verify

Generate YAML from the imported Go code:

```bash
wetwire-k8s build -o output.yaml
```

Compare with original:

```bash
diff my-resources.yaml output.yaml
```

Minor differences are normal (field ordering, zero value omission).

## Import Options

### Custom Package Name

Generate code for a specific package:

```bash
wetwire-k8s import --package k8s -o k8s/resources.go manifests.yaml
```

### Variable Name Prefix

Add a prefix to all generated variable names:

```bash
wetwire-k8s import --var-prefix Prod -o prod.go manifests.yaml
```

This generates:

```go
var ProdNginxDeployment = &appsv1.Deployment{...}
var ProdNginxService = &corev1.Service{...}
```

### Disable Optimizations

Skip extracting shared values:

```bash
wetwire-k8s import --optimize=false -o k8s.go manifests.yaml
```

Use this when you want literal translations without refactoring.

## Common Patterns

### Multi-Document YAML

Import files with multiple resources separated by `---`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 3
  # ... deployment spec
---
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  # ... service spec
```

Import command:

```bash
wetwire-k8s import multi-resource.yaml -o k8s.go
```

Output:

```go
var AppConfig = &corev1.ConfigMap{...}
var AppDeployment = &appsv1.Deployment{...}
var AppService = &corev1.Service{...}
```

### Importing from stdin

Import from kubectl output:

```bash
kubectl get deployment nginx -o yaml | wetwire-k8s import - -o nginx.go
```

Import from helm template:

```bash
helm template myapp ./chart | wetwire-k8s import - -o myapp.go
```

Import from kustomize:

```bash
kustomize build ./overlays/prod | wetwire-k8s import - -o prod.go
```

### Directory of YAML Files

Import multiple files by concatenating them:

```bash
cat manifests/*.yaml | wetwire-k8s import - -o k8s.go
```

Or import each file separately:

```bash
for file in manifests/*.yaml; do
    wetwire-k8s import "$file" -o "k8s/$(basename ${file%.yaml}).go"
done
```

## Edge Cases and Solutions

### Unsupported Resource Types

**Problem:** Custom Resources (CRDs) may not import correctly.

**Solution:** Import generates best-effort code. Manually adjust using `unstructured.Unstructured`:

```go
import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

var MyCustomResource = &unstructured.Unstructured{
    Object: map[string]interface{}{
        "apiVersion": "example.com/v1",
        "kind":       "MyResource",
        "metadata": map[string]interface{}{
            "name": "my-resource",
        },
        "spec": map[string]interface{}{
            // ... spec fields
        },
    },
}
```

### Complex Field Expressions

**Problem:** Computed values or expressions cannot be represented in YAML.

**Solution:** After import, refactor to use Go expressions:

```go
// Before (literal value)
Replicas: ptr(int32(3))

// After (computed)
var baseReplicas = 3
var NginxDeployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(baseReplicas)),
    },
}
```

### Duplicate Resource Names

**Problem:** Multiple resources with the same kind and name generate name conflicts.

**Solution 1:** Use `--var-prefix`:

```bash
wetwire-k8s import dev.yaml --var-prefix Dev -o dev.go
wetwire-k8s import prod.yaml --var-prefix Prod -o prod.go
```

**Solution 2:** Manually rename variables after import.

### Missing Dependencies

**Problem:** Imports reference resources not in the YAML file.

**Solution:** Import all related files together or manually add missing resources:

```bash
# Import application stack together
cat app-deployment.yaml app-service.yaml app-configmap.yaml | \
  wetwire-k8s import - -o app.go
```

### Namespace Inconsistencies

**Problem:** Some resources have namespace, others don't.

**Solution:** After import, standardize namespaces:

```go
const appNamespace = "production"

var NginxDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "nginx",
        Namespace: appNamespace,  // Use constant
    },
}
```

## Post-Import Cleanup

### 1. Extract Common Configurations

Identify repeated values and extract to variables:

```go
// Before
var Deployment1 = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
    },
}
var Deployment2 = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
    },
}

// After
var defaultReplicas = int32(3)
var Deployment1 = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(defaultReplicas),
    },
}
var Deployment2 = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(defaultReplicas),
    },
}
```

### 2. Create Direct References

Replace string references with direct field references:

```go
// Before
var NginxService = &corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: map[string]string{"app": "nginx"},
    },
}

// After (assuming NginxDeployment exists)
var NginxService = &corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: NginxDeployment.Spec.Selector.MatchLabels,
    },
}
```

### 3. Add Comments

Document each resource:

```go
// NginxDeployment runs the nginx web server with 3 replicas
var NginxDeployment = &appsv1.Deployment{...}

// NginxService exposes the nginx deployment on port 80
var NginxService = &corev1.Service{...}
```

### 4. Organize by Section

Group related resources:

```go
// =============================================================================
// Configuration
// =============================================================================

var AppConfig = &corev1.ConfigMap{...}
var AppSecrets = &corev1.Secret{...}

// =============================================================================
// Application
// =============================================================================

var AppDeployment = &appsv1.Deployment{...}
var AppService = &corev1.Service{...}
```

### 5. Validate Generated Code

Run the full validation pipeline:

```bash
# Lint for pattern violations
wetwire-k8s lint --fix k8s.go

# Validate against Kubernetes schemas
wetwire-k8s validate k8s.go

# Build and verify output
wetwire-k8s build -o output.yaml
kubectl apply --dry-run=client -f output.yaml
```

## Advanced Workflows

### Incremental Migration

Migrate resources gradually:

1. **Start with stateless services** (Deployments, Services)
2. **Add configuration** (ConfigMaps, Secrets)
3. **Import stateful components** (StatefulSets, PVCs)
4. **Finally import infrastructure** (Namespaces, RBAC)

### Environment-Specific Imports

Import different environments to separate files:

```bash
# Development
kubectl get all -n dev -o yaml | wetwire-k8s import - --var-prefix Dev -o dev.go

# Production
kubectl get all -n prod -o yaml | wetwire-k8s import - --var-prefix Prod -o prod.go
```

Then extract common base:

```go
// common.go
var appLabels = map[string]string{"app": "myapp"}

// dev.go
var DevDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "myapp",
        Namespace: "dev",
        Labels:    appLabels,
    },
}

// prod.go
var ProdDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "myapp",
        Namespace: "prod",
        Labels:    appLabels,
    },
}
```

### Helm Chart Conversion

Convert rendered Helm charts:

```bash
# Render chart with values
helm template myapp ./chart -f values.yaml > rendered.yaml

# Import
wetwire-k8s import rendered.yaml -o myapp.go

# Clean up Helm annotations
sed -i '/helm.sh\/chart/d' myapp.go
sed -i '/app.kubernetes.io\/managed-by: Helm/d' myapp.go
```

### Kustomize Migration

Import kustomize overlays:

```bash
# Base resources
kustomize build ./base | wetwire-k8s import - -o base.go

# Development overlay
kustomize build ./overlays/dev | wetwire-k8s import - --var-prefix Dev -o dev.go

# Production overlay
kustomize build ./overlays/prod | wetwire-k8s import - --var-prefix Prod -o prod.go
```

## Troubleshooting

### Import Fails with Parse Error

**Error:** `Failed to parse YAML: ...`

**Solution:** Validate YAML syntax:

```bash
yamllint my-resources.yaml
kubectl apply --dry-run=client -f my-resources.yaml
```

### Generated Code Doesn't Compile

**Error:** `undefined: SomeType`

**Solution:** Missing import or unsupported type. Check import statements and API versions.

### Build Output Differs from Original

**Expected:** Minor differences in field ordering and zero values.

**Problem:** Significant semantic differences.

**Solution:** Review the original YAML for defaulted fields. Kubernetes admission controllers may add defaults.

### Lint Reports Many Violations

**Expected:** Imported YAML may not follow wetwire patterns.

**Solution:** Run auto-fix, then manually address remaining issues:

```bash
wetwire-k8s lint --fix k8s.go
wetwire-k8s lint k8s.go  # Review remaining issues
```

## Best Practices

1. **Import incrementally** - Start with simple resources, build confidence
2. **Use --optimize** - Let the importer extract shared values
3. **Run lint --fix** - Automatically fix common issues
4. **Add comments** - Document what each resource does
5. **Test round-trip** - Build YAML and compare with original
6. **Validate early** - Use kubectl --dry-run before deploying
7. **Version control** - Commit imported code, review diffs
8. **Refactor gradually** - Don't try to perfect everything at once

## Next Steps

- Review [CLI.md](CLI.md) for complete import command reference
- See [ADOPTION.md](ADOPTION.md) for migration strategies
- Check [EXAMPLES.md](EXAMPLES.md) for common import patterns
- Read [DEVELOPERS.md](DEVELOPERS.md) for extending import functionality
