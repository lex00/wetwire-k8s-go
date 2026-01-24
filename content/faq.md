---
title: "FAQ"
---

# Frequently Asked Questions

## General

<details>
<summary>What is wetwire-k8s-go?</summary>

wetwire-k8s-go is a Go implementation of the wetwire pattern for Kubernetes manifests. It allows you to define Kubernetes resources using native Go code with full type safety, IDE support, and AI-assisted development.
</details>

<details>
<summary>Why use Go code instead of YAML?</summary>

Go code provides several advantages:

- **Type safety** - Catch errors at compile time, not deployment time
- **IDE support** - Full autocomplete, go-to-definition, refactoring
- **Code reuse** - Functions, variables, and modules for DRY principles
- **Testing** - Unit test your infrastructure definitions
- **AI generation** - AI can generate and modify code more reliably than YAML
- **Refactoring** - Rename and reorganize with IDE support
</details>

<details>
<summary>How does wetwire-k8s-go differ from other tools?</summary>

| Tool | Approach | wetwire-k8s-go Difference |
|------|----------|---------------------------|
| **kubectl + YAML** | Manual YAML editing | Type-safe Go, IDE support |
| **Helm** | Template engine | Native Go, no templating |
| **Kustomize** | YAML patching | Programmatic composition |
| **Pulumi** | Imperative SDK | Declarative, lint-enforced patterns |
| **CDK8s** | Imperative constructors | Flat, AI-optimized declarations |
</details>

<details>
<summary>Is this production-ready?</summary>

wetwire-k8s-go generates standard Kubernetes YAML that can be applied with `kubectl`. The tool itself is in active development. Review generated manifests before deploying to production.
</details>

---

## Installation

<details>
<summary>How do I install wetwire-k8s?</summary>

```bash
go install github.com/lex00/wetwire-k8s-go@latest
```

Requires Go 1.23 or later.
</details>

<details>
<summary>What are the prerequisites?</summary>

- Go 1.23 or later
- Basic familiarity with Kubernetes concepts
- (Optional) `kubectl` for diff and deployment
</details>

<details>
<summary>How do I update to the latest version?</summary>

```bash
go install github.com/lex00/wetwire-k8s-go@latest
```

Or update your `go.mod`:

```bash
go get -u github.com/lex00/wetwire-k8s-go
```
</details>

---

## Usage

<details>
<summary>How do I get started?</summary>

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
</details>

<details>
<summary>How do I define a Deployment?</summary>

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
</details>

<details>
<summary>How do I reference other resources?</summary>

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
</details>

<details>
<summary>Can I use functions and loops?</summary>

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
</details>

<details>
<summary>How do I import existing YAML?</summary>

```bash
wetwire-k8s import -o k8s.go existing-manifests.yaml
```

This converts YAML to Go code. Review and run `wetwire-k8s lint --fix` to apply wetwire patterns.
</details>

---

## GitOps and CI/CD

<details>
<summary>How do I integrate with my GitOps workflow?</summary>

wetwire-k8s-go fits naturally into GitOps workflows:

1. **Store Go source in Git** - Your Kubernetes definitions are versioned Go code
2. **CI builds manifests** - Run `wetwire-k8s build -o manifests.yaml` in your pipeline
3. **Commit generated YAML** - Either commit to the same repo or a separate GitOps repo
4. **ArgoCD/Flux syncs** - Point your GitOps tool at the generated manifests

Example GitHub Actions step:
```yaml
- name: Build manifests
  run: |
    wetwire-k8s build -o manifests/
    git add manifests/
    git commit -m "Update generated manifests" || true
```
</details>

<details>
<summary>Can I import existing YAML manifests?</summary>

Yes, use the import command:

```bash
# Import a single file
wetwire-k8s import -o k8s.go existing-manifests.yaml

# Import multiple files
wetwire-k8s import -o k8s.go manifests/*.yaml

# Import with custom package name
wetwire-k8s import --package myapp -o k8s.go manifests.yaml
```

After importing, run `wetwire-k8s lint --fix` to apply wetwire patterns and best practices.
</details>

<details>
<summary>How does the linter help catch errors?</summary>

The linter enforces patterns that prevent common Kubernetes mistakes:

- **Label selector mismatches** - Ensures Deployment selectors match Pod template labels
- **Missing required fields** - Catches missing container names, image tags, etc.
- **Security issues** - Warns about privileged containers, missing resource limits
- **Naming conventions** - Enforces consistent resource naming
- **Direct references** - Ensures dependencies are explicit and traceable

Run `wetwire-k8s lint` before committing to catch issues early. Use `--fix` to auto-fix many problems.
</details>

<details>
<summary>What's the recommended project structure?</summary>

```
my-app/
├── go.mod
├── go.sum
├── k8s/
│   ├── deployment.go    # Deployment definitions
│   ├── service.go       # Service definitions
│   ├── configmap.go     # ConfigMaps and Secrets
│   └── ingress.go       # Ingress rules
├── manifests/           # Generated YAML (optional, for GitOps)
│   └── all.yaml
└── Makefile
```

Organize by resource type or by application component. The build command discovers all resources regardless of file organization.
</details>

<details>
<summary>How do I handle Helm chart generation?</summary>

wetwire-k8s-go generates plain Kubernetes YAML, not Helm charts. However, you can:

1. **Use wetwire-k8s for your own apps** - Generate manifests directly
2. **Combine with Helm** - Use Helm for third-party charts, wetwire-k8s for your code
3. **Template with Go** - Use Go's text/template if you need parameterization

For values that change between environments, use Go variables or build-time flags:

```go
var Namespace = os.Getenv("K8S_NAMESPACE")
if Namespace == "" {
    Namespace = "default"
}
```
</details>

---

## Kubernetes-specific

<details>
<summary>How do I handle multiple namespaces?</summary>

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
</details>

<details>
<summary>How do I set resource limits?</summary>

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
</details>

<details>
<summary>How do I use Secrets?</summary>

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
</details>

<details>
<summary>What about StatefulSets and DaemonSets?</summary>

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
</details>

<details>
<summary>How do I use Custom Resources (CRDs)?</summary>

Custom Resource Definitions are not yet supported. This is planned for a future release.
</details>

---

## Design mode

<details>
<summary>What is design mode?</summary>

Design mode is AI-assisted development where you describe what you want in natural language, and Claude generates the Go code.
</details>

<details>
<summary>How do I use design mode?</summary>

**Recommended:** Use Claude Code with the MCP server:

1. Install wetwire-k8s: `go install github.com/lex00/wetwire-k8s-go@latest`
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
</details>

<details>
<summary>Do I need an API key?</summary>

- **Claude Code + MCP:** No, Claude Code handles authentication
- **CLI design command:** Yes, set `ANTHROPIC_API_KEY` environment variable
</details>

<details>
<summary>What can I ask for?</summary>

Examples:

- "Create a Deployment for nginx with 3 replicas"
- "Add a Service that exposes port 80"
- "Create a StatefulSet for PostgreSQL with persistent storage"
- "Generate a complete web application stack with frontend, backend, and database"
- "Add resource limits and health checks to the deployment"

Be specific about requirements. Claude can ask clarifying questions.
</details>

<details>
<summary>How accurate is the generated code?</summary>

Generated code is:
- Syntactically correct Go
- Follows wetwire patterns (flat, declarative)
- Validated against Kubernetes schemas
- Auto-fixed by the linter

Review generated code before deploying, especially for production.
</details>

---

## Linting

<details>
<summary>What does the linter check?</summary>

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
</details>

<details>
<summary>How do I auto-fix issues?</summary>

```bash
wetwire-k8s lint --fix
```

Many issues can be auto-fixed. Manual fixes may be needed for complex violations.
</details>

<details>
<summary>Can I disable specific rules?</summary>

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
</details>

<details>
<summary>Why is the linter flagging my working code?</summary>

The linter enforces patterns that enable static analysis and AI generation. Working code may not follow these patterns. Consider:

1. Does the code use direct field references (not function calls)?
2. Are resources declared at top level?
3. Is the code flat and declarative?

If you have a valid reason for imperative code, you can disable specific rules.
</details>

---

## Validation

<details>
<summary>What's the difference between lint and validate?</summary>

- **lint** - Checks Go code for wetwire patterns and best practices
- **validate** - Checks resources against Kubernetes schemas (required fields, types, API versions)

Both are run during `build` by default.
</details>

<details>
<summary>What Kubernetes versions are supported?</summary>

wetwire-k8s validates against Kubernetes 1.28 by default. Specify a different version with `--k8s-version`:

```bash
wetwire-k8s validate --k8s-version 1.30
```

Supported versions: 1.24 through 1.31.
</details>

<details>
<summary>Can I validate without building?</summary>

Yes:

```bash
wetwire-k8s validate
```

This parses Go code and validates resources without generating YAML.
</details>

---

## Troubleshooting

<details>
<summary>"Cannot find package" error</summary>

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
</details>

<details>
<summary>"Undefined: ptrInt32" error</summary>

Many Kubernetes fields are pointers. Define a helper function:

```go
func ptrInt32(i int32) *int32 { return &i }
func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool { return &b }
```
</details>

<details>
<summary>Linter says "resource not discoverable"</summary>

Resources MUST be top-level variables:

```go
// Good
var MyPod = corev1.Pod{...}

// Bad - not discoverable
func createPod() corev1.Pod {
    return corev1.Pod{...}
}
```
</details>

<details>
<summary>Generated YAML is in wrong order</summary>

Build generates resources in dependency order. If order is still wrong, there may be a circular dependency or the dependency graph isn't detecting all references.

Check that you're using direct field references (not function calls).
</details>

<details>
<summary>"Invalid API version" error</summary>

Make sure you're using the correct import:

```go
// Correct
import appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"

// Wrong - different version
import appsv1beta1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1beta1"
```
</details>

<details>
<summary>How do I debug build issues?</summary>

Enable verbose output:

```bash
wetwire-k8s build --verbose
```

This shows:
- Discovered resources
- Dependency graph
- Validation results
- Generation steps
</details>

---

## Getting help

<details>
<summary>Where can I find more documentation?</summary>

- [CLI Reference](/cli/) - Complete command documentation
- [Lint Rules](/lint-rules/) - All lint rules with examples
- [Quick Start](/quick-start/) - Getting started guide
- [Examples](/examples/) - Real-world manifest patterns
</details>

<details>
<summary>How do I report bugs?</summary>

Open an issue on GitHub: https://github.com/lex00/wetwire-k8s-go/issues

Include:
- wetwire-k8s version (`wetwire-k8s --version`)
- Go version (`go version`)
- Command that failed
- Error output
- Minimal reproduction example
</details>

<details>
<summary>How do I request features?</summary>

Open a feature request on GitHub: https://github.com/lex00/wetwire-k8s-go/issues

Describe:
- What you're trying to accomplish
- Why existing features don't work
- Example of desired syntax/behavior
</details>

<details>
<summary>Can I contribute?</summary>

Yes! See [Contributing](/contributing/) for guidelines.

Areas where contributions are especially welcome:
- Additional lint rules
- Kubernetes version support
- Import optimizations
- Documentation improvements
- Example projects
</details>
