<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

Common issues and solutions when using wetwire-k8s-go.

## Build Issues

### No resources found

**Symptom:** `wetwire-k8s build` produces no output.

**Possible causes:**

1. **Resources not at top level**

   Resources must be declared as top-level package variables:

   ```go
   // Wrong - inside function
   func main() {
       deployment := &appsv1.Deployment{...}
   }

   // Correct - top-level variable
   var MyDeployment = &appsv1.Deployment{...}
   ```

2. **Unrecognized package alias**

   Use standard package aliases that wetwire-k8s recognizes:

   ```go
   // Recognized aliases
   import (
       appsv1 "k8s.io/api/apps/v1"
       corev1 "k8s.io/api/core/v1"
       batchv1 "k8s.io/api/batch/v1"
       networkingv1 "k8s.io/api/networking/v1"
   )

   // Not recognized - custom alias
   import (
       apps "k8s.io/api/apps/v1"  // Won't be discovered
   )
   ```

3. **Wrong directory**

   Ensure you're running from the correct directory:

   ```bash
   wetwire-k8s build ./path/to/k8s
   ```

4. **Test files only**

   Files ending in `_test.go` are skipped. Ensure your resources are in regular `.go` files.

### Parse errors

**Symptom:** `build failed: discovery failed: failed to parse file`

**Solutions:**

1. Check for Go syntax errors:

   ```bash
   go build ./...
   ```

2. Ensure all imports are valid:

   ```bash
   go mod tidy
   ```

3. Check for unresolved dependencies:

   ```bash
   go mod download
   ```

### Missing apiVersion or kind

**Symptom:** Generated YAML has `apiVersion: v1` and `kind: Unknown`

**Cause:** The resource type couldn't be determined from the package.

**Solution:** Use qualified types with recognized package aliases:

```go
// Wrong - unqualified type
var MyPod = Pod{...}

// Correct - qualified type
var MyPod = &corev1.Pod{...}
```

## Type Errors

### Pointer field errors

**Symptom:** `cannot use 3 (type int) as type *int32`

**Solution:** Use pointer helpers for pointer fields:

```go
// Define a helper
func ptr[T any](v T) *T { return &v }

// Use it for pointer fields
Replicas: ptr(int32(3)),
Timeout:  ptr(int64(30)),
```

Common pointer fields:
- `Replicas` (*int32)
- `TerminationGracePeriodSeconds` (*int64)
- `RevisionHistoryLimit` (*int32)
- `ActiveDeadlineSeconds` (*int64)

### IntOrString fields

**Symptom:** `cannot use 80 (type int) as type intstr.IntOrString`

**Solution:** Use `intstr.FromInt()` or `intstr.FromString()`:

```go
import "k8s.io/apimachinery/pkg/util/intstr"

// For integer values
TargetPort: intstr.FromInt(80),

// For string values (port names)
TargetPort: intstr.FromString("http"),
```

### Missing ObjectMeta fields

**Symptom:** `unknown field 'Name' in struct literal`

**Cause:** ObjectMeta is not in the expected location.

**Solution:** Check the struct hierarchy:

```go
// Deployment uses ObjectMeta (not Metadata)
var MyDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{  // Correct field name
        Name: "my-app",
    },
}

// Pod template uses ObjectMeta too
Template: corev1.PodTemplateSpec{
    ObjectMeta: metav1.ObjectMeta{  // Not "Metadata"
        Labels: map[string]string{...},
    },
}
```

## Lint Errors

### WK8001: Non-flat declaration

**Symptom:** `resource declaration should be at package level`

**Solution:** Move resource declarations to the top level:

```go
// Wrong
func createDeployment() *appsv1.Deployment {
    return &appsv1.Deployment{...}
}

// Correct
var MyDeployment = &appsv1.Deployment{...}
```

### WK8002: Missing labels

**Symptom:** `resource should have labels for identification`

**Solution:** Add labels to metadata:

```go
ObjectMeta: metav1.ObjectMeta{
    Name: "my-app",
    Labels: map[string]string{
        "app": "my-app",
    },
}
```

### WK8003: Selector mismatch

**Symptom:** `selector labels do not match pod template labels`

**Solution:** Ensure selector matches template labels:

```go
var MyDeployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{
            MatchLabels: map[string]string{"app": "myapp"},  // Must match
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: map[string]string{"app": "myapp"},   // Must match
            },
        },
    },
}
```

Or use a shared variable:

```go
var appLabels = map[string]string{"app": "myapp"}

var MyDeployment = &appsv1.Deployment{
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

### WK8004: Resource without limits

**Symptom:** `container should specify resource limits`

**Solution:** Add resource requests and limits:

```go
Container: corev1.Container{
    Resources: corev1.ResourceRequirements{
        Requests: corev1.ResourceList{
            corev1.ResourceCPU:    resource.MustParse("100m"),
            corev1.ResourceMemory: resource.MustParse("128Mi"),
        },
        Limits: corev1.ResourceList{
            corev1.ResourceCPU:    resource.MustParse("500m"),
            corev1.ResourceMemory: resource.MustParse("512Mi"),
        },
    },
}
```

## Validation Errors

### Invalid API version

**Symptom:** `unknown apiVersion: apps/v1beta1`

**Solution:** Use stable API versions:

```go
// Wrong - deprecated
import appsv1beta1 "k8s.io/api/apps/v1beta1"

// Correct - stable
import appsv1 "k8s.io/api/apps/v1"
```

### Required field missing

**Symptom:** `spec.selector is required`

**Solution:** Add the required field:

```go
var MyDeployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{  // Required
            MatchLabels: map[string]string{"app": "myapp"},
        },
        Template: corev1.PodTemplateSpec{...},
    },
}
```

### Invalid field value

**Symptom:** `spec.replicas must be greater than 0`

**Solution:** Check field constraints in Kubernetes documentation.

## Import Errors

### Unsupported resource type

**Symptom:** `unsupported kind: CustomResourceDefinition`

**Cause:** The importer only supports built-in Kubernetes types.

**Solution:** Manually convert CRDs or use raw YAML.

### Malformed YAML

**Symptom:** `failed to parse YAML: yaml: line N: did not find expected key`

**Solutions:**

1. Validate YAML syntax:
   ```bash
   kubectl apply --dry-run=client -f manifest.yaml
   ```

2. Check for tabs (YAML uses spaces):
   ```bash
   cat -A manifest.yaml | grep -E '^\t'
   ```

3. Ensure proper indentation (2 spaces recommended)

## Dependency Issues

### Circular dependency detected

**Symptom:** `cycle detected: A -> B -> A`

**Solution:** Refactor to break the cycle:

```go
// Wrong - circular
var A = &corev1.ConfigMap{
    Data: map[string]string{
        "key": B.Name,  // A depends on B
    },
}

var B = &corev1.ConfigMap{
    Data: map[string]string{
        "key": A.Name,  // B depends on A
    },
}

// Correct - break cycle with shared value
var sharedName = "shared-config"

var A = &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{Name: "a"},
    Data: map[string]string{
        "key": sharedName,
    },
}

var B = &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{Name: "b"},
    Data: map[string]string{
        "key": sharedName,
    },
}
```

### Unresolved reference

**Symptom:** `reference to undefined resource: MyConfig`

**Solutions:**

1. Check spelling and case sensitivity
2. Ensure referenced resource is in the same package
3. Verify the resource is a top-level variable

## Performance Issues

### Slow discovery

**Symptom:** Build takes a long time on large codebases.

**Solutions:**

1. Build only the k8s directory:
   ```bash
   wetwire-k8s build ./k8s
   ```

2. Exclude unrelated files:
   ```bash
   # Move k8s definitions to dedicated directory
   mkdir k8s
   mv *_k8s.go k8s/
   ```

### High memory usage

**Symptom:** Process uses excessive memory on large projects.

**Solutions:**

1. Split resources into multiple packages
2. Build packages separately
3. Avoid deeply nested references

## Getting Help

If you're still stuck:

1. Check the [FAQ](FAQ.md) for common questions
2. Review [examples](../examples/) for working code
3. Run with verbose output:
   ```bash
   wetwire-k8s build -v
   ```
4. File an issue at https://github.com/lex00/wetwire-k8s-go/issues

When reporting issues, include:

- Go version: `go version`
- wetwire-k8s version: `wetwire-k8s --version`
- Minimal reproducing example
- Full error output
