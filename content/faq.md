---
title: "FAQ"
---

# Frequently Asked Questions

## General

### What is wetwire-k8s-go?

wetwire-k8s-go is a Go implementation of the wetwire pattern for Kubernetes manifests. It allows you to define Kubernetes resources using native Go code with full type safety, IDE support, and AI-assisted development.

### Why use Go code instead of YAML?

Go code provides several advantages:

- **Type safety** - Catch errors at compile time, not deployment time
- **IDE support** - Full autocomplete, go-to-definition, refactoring
- **Code reuse** - Functions, variables, and modules for DRY principles
- **Testing** - Unit test your infrastructure definitions
- **AI generation** - AI can generate and modify code more reliably than YAML
- **Refactoring** - Rename and reorganize with IDE support

### How does wetwire-k8s-go differ from other tools?

| Tool | Approach | wetwire-k8s-go Difference |
|------|----------|---------------------------|
| **kubectl + YAML** | Manual YAML editing | Type-safe Go, IDE support |
| **Helm** | Template engine | Native Go, no templating |
| **Kustomize** | YAML patching | Programmatic composition |
| **Pulumi** | Imperative SDK | Declarative, lint-enforced patterns |
| **CDK8s** | Imperative constructors | Flat, AI-optimized declarations |

### Is this production-ready?

wetwire-k8s-go generates standard Kubernetes YAML that can be applied with `kubectl`. The tool itself is in active development. Review generated manifests before deploying to production.

---

## Installation

### How do I install wetwire-k8s?

See [README.md](../README.md#installation) for installation instructions.

### What are the prerequisites?

- Go 1.23 or later
- Basic familiarity with Kubernetes concepts
- (Optional) `kubectl` for diff and deployment

### How do I update to the latest version?

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

Or update your `go.mod`:

```bash
go get -u github.com/lex00/wetwire-k8s-go
```

---

## Usage

### How do I get started?

1. Initialize a new project:
   ```bash
   wetwire-k8s init --example
   ```

2. Edit the generated `main.go` or create new `.go` files

3. Build manifests:
   ```bash
   wetwire-k8s build -o manifests.yaml
   ```

4. Apply to cluster:
   ```bash
   kubectl apply -f manifests.yaml
   ```

### How do I define a Deployment?

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

### How do I reference other resources?

Use direct field references:

```go
var MyConfigMap = corev1.ConfigMap{
    Metadata: corev1.ObjectMeta{
        Name: "app-config",
    },
    Data: map[string]string{
        "config.yaml": "...",
    },
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Volumes: []corev1.Volume{
                    {
                        Name: "config",
                        VolumeSource: corev1.VolumeSource{
                            ConfigMap: &corev1.ConfigMapVolumeSource{
                                Name: MyConfigMap.Metadata.Name,  // Direct reference
                            },
                        },
                    },
                },
            },
        },
    },
}
```

### Can I use functions and loops?

The wetwire pattern encourages flat, declarative code. The linter will flag imperative constructs like loops and conditionals inside resource definitions.

For repetitive patterns, extract to variables or use helper functions OUTSIDE resource definitions:

```go
// Good - helper function returns value
func createContainer(name, image string) corev1.Container {
    return corev1.Container{
        Name:  name,
        Image: image,
    }
}

var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        Containers: []corev1.Container{
            createContainer("app", "myapp:latest"),
            createContainer("sidecar", "proxy:latest"),
        },
    },
}
```

### How do I import existing YAML?

```bash
wetwire-k8s import -o k8s.go existing-manifests.yaml
```

This converts YAML to Go code. Review and run `wetwire-k8s lint --fix` to apply wetwire patterns.

---

## Kubernetes-specific

### How do I handle multiple namespaces?

Define namespace explicitly on each resource:

```go
var DevDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name:      "myapp",
        Namespace: "development",
    },
    // ...
}

var ProdDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name:      "myapp",
        Namespace: "production",
    },
    // ...
}
```

Or use a variable:

```go
const targetNamespace = "production"

var MyDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name:      "myapp",
        Namespace: targetNamespace,
    },
    // ...
}
```

### How do I set resource limits?

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "app",
                        Image: "myapp:latest",
                        Resources: corev1.ResourceRequirements{
                            Requests: corev1.ResourceList{
                                "cpu":    "100m",
                                "memory": "128Mi",
                            },
                            Limits: corev1.ResourceList{
                                "cpu":    "500m",
                                "memory": "512Mi",
                            },
                        },
                    },
                },
            },
        },
    },
}
```

### How do I use Secrets?

Define the Secret and reference it:

```go
var MySecret = corev1.Secret{
    Metadata: corev1.ObjectMeta{
        Name: "db-credentials",
    },
    Type: "Opaque",
    StringData: map[string]string{
        "username": "admin",
        "password": "secret",  // Use external secret management in production
    },
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name: "app",
                        Env: []corev1.EnvVar{
                            {
                                Name: "DB_USER",
                                ValueFrom: &corev1.EnvVarSource{
                                    SecretKeyRef: &corev1.SecretKeySelector{
                                        LocalObjectReference: corev1.LocalObjectReference{
                                            Name: MySecret.Metadata.Name,
                                        },
                                        Key: "username",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    },
}
```

### What about StatefulSets and DaemonSets?

They work the same way as Deployments:

```go
import appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"

var MyStatefulSet = appsv1.StatefulSet{
    Metadata: corev1.ObjectMeta{
        Name: "database",
    },
    Spec: appsv1.StatefulSetSpec{
        ServiceName: "db-service",
        Replicas:    ptrInt32(3),
        // ...
    },
}

var MyDaemonSet = appsv1.DaemonSet{
    Metadata: corev1.ObjectMeta{
        Name: "log-collector",
    },
    Spec: appsv1.DaemonSetSpec{
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{"app": "logs"},
        },
        // ...
    },
}
```

### How do I use Custom Resources (CRDs)?

Custom Resource Definitions are not yet supported. This is planned for a future release.

---

## Design mode

### What is design mode?

Design mode is AI-assisted development where you describe what you want in natural language, and Claude generates the Go code.

### How do I use design mode?

**Recommended:** Use Claude Code with the MCP server:

1. Install wetwire-k8s: `go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest`
2. Configure MCP in Claude Code settings:
   ```json
   {
     "mcpServers": {
       "wetwire-k8s": {
         "command": "wetwire-k8s-mcp"
       }
     }
   }
   ```
3. Start Claude Code in your project directory
4. Ask Claude to generate Kubernetes resources

**Alternative:** Use the CLI:

```bash
export ANTHROPIC_API_KEY=your-key
wetwire-k8s design "Create a deployment for nginx"
```

### Do I need an API key?

- **Claude Code + MCP:** No, Claude Code handles authentication
- **CLI design command:** Yes, set `ANTHROPIC_API_KEY` environment variable

### What can I ask for?

Examples:

- "Create a Deployment for nginx with 3 replicas"
- "Add a Service that exposes port 80"
- "Create a StatefulSet for PostgreSQL with persistent storage"
- "Generate a complete web application stack with frontend, backend, and database"
- "Add resource limits and health checks to the deployment"

Be specific about requirements. Claude can ask clarifying questions.

### How accurate is the generated code?

Generated code is:
- Syntactically correct Go
- Follows wetwire patterns (flat, declarative)
- Validated against Kubernetes schemas
- Auto-fixed by the linter

Review generated code before deploying, especially for production.

---

## Linting

### What does the linter check?

The linter enforces:
- Flat, declarative patterns (no nested constructors, loops, conditionals)
- Top-level resource declarations
- Direct field references (no function calls for dependencies)
- Label selector consistency
- Required fields
- Security best practices
- Resource limits
- Naming conventions

See [Lint Rules](/lint-rules/) for complete rule list.

### How do I auto-fix issues?

```bash
wetwire-k8s lint --fix
```

Many issues can be auto-fixed. Manual fixes may be needed for complex violations.

### Can I disable specific rules?

Yes, with `--disable`:

```bash
wetwire-k8s lint --disable WK8001,WK8002
```

Or in `.wetwire-k8s.yaml`:

```yaml
lint:
  disabled_rules:
    - WK8001
    - WK8002
```

### Why is the linter flagging my working code?

The linter enforces patterns that enable static analysis and AI generation. Working code may not follow these patterns. Consider:

1. Does the code use direct field references (not function calls)?
2. Are resources declared at top level?
3. Is the code flat and declarative?

If you have a valid reason for imperative code, you can disable specific rules.

---

## Validation

### What's the difference between lint and validate?

- **lint** - Checks Go code for wetwire patterns and best practices
- **validate** - Checks resources against Kubernetes schemas (required fields, types, API versions)

Both are run during `build` by default.

### What Kubernetes versions are supported?

wetwire-k8s validates against Kubernetes 1.28 by default. Specify a different version with `--k8s-version`:

```bash
wetwire-k8s validate --k8s-version 1.30
```

Supported versions: 1.24 through 1.31.

### Can I validate without building?

Yes:

```bash
wetwire-k8s validate
```

This parses Go code and validates resources without generating YAML.

---

## Troubleshooting

### "Cannot find package" error

Make sure dependencies are installed:

```bash
go mod download
```

And that you're importing from the correct package:

```go
import (
    appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"
    corev1 "github.com/lex00/wetwire-k8s-go/resources/core/v1"
)
```

### "Undefined: ptrInt32" error

Many Kubernetes fields are pointers. Define a helper function:

```go
func ptrInt32(i int32) *int32 { return &i }
func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool { return &b }
```

Or use the helpers from wetwire-core-go:

```go
import "github.com/lex00/wetwire-core-go/k8s/helpers"

Replicas: helpers.Int32Ptr(3),
```

### Linter says "resource not discoverable"

Resources MUST be top-level variables:

```go
// Good
var MyPod = corev1.Pod{...}

// Bad - not discoverable
func createPod() corev1.Pod {
    return corev1.Pod{...}
}
```

### Generated YAML is in wrong order

Build generates resources in dependency order. If order is still wrong, there may be a circular dependency or the dependency graph isn't detecting all references.

Check that you're using direct field references (not function calls).

### "Invalid API version" error

Make sure you're using the correct import:

```go
// Correct
import appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"

// Wrong - different version
import appsv1beta1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1beta1"
```

### How do I debug build issues?

Enable verbose output:

```bash
wetwire-k8s build --verbose
```

This shows:
- Discovered resources
- Dependency graph
- Validation results
- Generation steps

---

## Comparison

### How does this compare to Helm?

| Aspect | Helm | wetwire-k8s-go |
|--------|------|----------------|
| **Syntax** | Go templates in YAML | Native Go |
| **Type safety** | No | Yes |
| **IDE support** | Limited | Full |
| **Package management** | Charts | Go modules |
| **Templating** | String templating | Go composition |
| **AI generation** | Difficult | Optimized |

Use Helm if you need a package ecosystem. Use wetwire-k8s-go if you want type safety and AI-assisted development.

### How does this compare to Kustomize?

| Aspect | Kustomize | wetwire-k8s-go |
|--------|-----------|----------------|
| **Approach** | YAML patching | Programmatic |
| **Type safety** | No | Yes |
| **Composition** | Overlays | Go code |
| **IDE support** | Limited | Full |

Use Kustomize if you want to stay in YAML. Use wetwire-k8s-go if you want programmatic composition with type safety.

### How does this compare to Pulumi?

| Aspect | Pulumi | wetwire-k8s-go |
|--------|--------|----------------|
| **Style** | Imperative | Declarative |
| **State** | Pulumi service | None (kubectl) |
| **Control flow** | Full programming | Lint-restricted |
| **AI generation** | Complex | Simple patterns |

Use Pulumi if you need stateful deployment orchestration. Use wetwire-k8s-go if you want declarative, AI-friendly code with kubectl.

### How does this compare to CDK8s?

| Aspect | CDK8s | wetwire-k8s-go |
|--------|-------|----------------|
| **Style** | Imperative constructors | Flat declarations |
| **Languages** | Multiple | Go only |
| **Nesting** | Deep | Flat |
| **AI generation** | Complex | Optimized |

Use CDK8s if you need multi-language support. Use wetwire-k8s-go if you want AI-optimized, flat patterns.

---

## Getting help

### Where can I find more documentation?

- [CLI Reference](/cli/) - Complete command documentation
- [Lint Rules](/lint-rules/) - All lint rules with examples
- [CLAUDE.md](../CLAUDE.md) - AI assistant context
- [Wetwire Specification](https://github.com/lex00/wetwire/docs/WETWIRE_SPEC.md) - Core patterns

### How do I report bugs?

Open an issue on GitHub: https://github.com/lex00/wetwire-k8s-go/issues

Include:
- wetwire-k8s version (`wetwire-k8s --version`)
- Go version (`go version`)
- Command that failed
- Error output
- Minimal reproduction example

### How do I request features?

Open a feature request on GitHub: https://github.com/lex00/wetwire-k8s-go/issues

Describe:
- What you're trying to accomplish
- Why existing features don't work
- Example of desired syntax/behavior

### Can I contribute?

Yes! See [Contributing](/contributing/) for guidelines.

Areas where contributions are especially welcome:
- Additional lint rules
- Kubernetes version support
- Import optimizations
- Documentation improvements
- Example projects
