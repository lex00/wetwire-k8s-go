# Lint Rules

This document describes all lint rules for wetwire-k8s-go.

## Overview

The wetwire-k8s linter enforces flat, declarative patterns optimized for AI generation and human readability. Rules check for structural patterns, Kubernetes best practices, and security issues.

## Rule naming convention

All rules follow the format: `WK8xxx`

- **W** - Wetwire
- **K8** - Kubernetes
- **xxx** - Three-digit number

## Rule index

| Rule | Description | Severity | Auto-fix |
|------|-------------|----------|----------|
| [WK8001](#wk8001-top-level-resource-declarations) | Resources must be top-level declarations | Error | No |
| [WK8002](#wk8002-no-nested-constructors) | No nested resource constructors | Error | Yes |
| [WK8003](#wk8003-direct-field-references) | Use direct field references, not function calls | Error | No |
| [WK8004](#wk8004-no-loops-in-resources) | No loops inside resource definitions | Error | No |
| [WK8005](#wk8005-no-conditionals-in-resources) | No conditionals inside resource definitions | Error | No |
| [WK8006](#wk8006-flat-variable-references) | Extract nested structures to variables | Warning | Yes |
| [WK8101](#wk8101-selector-label-match) | Selector labels must match template labels | Error | No |
| [WK8102](#wk8102-required-metadata-name) | Metadata.Name is required | Error | No |
| [WK8103](#wk8103-container-name-required) | Container name is required | Error | No |
| [WK8104](#wk8104-port-name-recommended) | Named ports are recommended | Warning | No |
| [WK8105](#wk8105-image-pull-policy) | ImagePullPolicy should be explicit | Warning | Yes |
| [WK8201](#wk8201-resource-limits-required) | Resource limits should be set | Warning | No |
| [WK8202](#wk8202-security-context-recommended) | SecurityContext is recommended | Warning | No |
| [WK8203](#wk8203-read-only-root-filesystem) | ReadOnlyRootFilesystem should be true | Warning | No |
| [WK8204](#wk8204-run-as-non-root) | RunAsNonRoot should be true | Warning | No |
| [WK8205](#wk8205-drop-capabilities) | Drop unnecessary Linux capabilities | Warning | No |
| [WK8206](#wk8206-no-privileged-containers) | Privileged containers should not be used | Error | No |
| [WK8207](#wk8207-no-host-network) | HostNetwork should not be used | Warning | No |
| [WK8208](#wk8208-no-host-pid) | HostPID should not be used | Warning | No |
| [WK8209](#wk8209-no-host-ipc) | HostIPC should not be used | Warning | No |
| [WK8301](#wk8301-health-checks-recommended) | Health checks (liveness/readiness) recommended | Warning | No |
| [WK8302](#wk8302-replicas-minimum) | Deployments should have at least 2 replicas | Info | No |
| [WK8303](#wk8303-pod-disruption-budget) | PodDisruptionBudget recommended for HA | Info | No |
| [WK8304](#wk8304-anti-affinity-recommended) | Pod anti-affinity recommended for HA | Info | No |

---

## Rules

### WK8001: Top-level resource declarations

**Description:** Kubernetes resources MUST be declared as top-level variables. This enables resource discovery via AST parsing without code execution.

**Severity:** Error

**Auto-fix:** No

**Why:** Resource discovery relies on finding top-level variable declarations. Nested or dynamically created resources cannot be discovered statically.

**Bad:**

```go
func CreateDeployment(name string) appsv1.Deployment {
    return appsv1.Deployment{
        Metadata: corev1.ObjectMeta{Name: name},
    }
}

var myDeploy = CreateDeployment("app")  // Not discoverable
```

**Good:**

```go
var MyDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{Name: "app"},
}
```

---

### WK8002: No nested constructors

**Description:** Resource definitions MUST NOT use nested constructor calls or struct literals for other resources. Extract nested resources to separate variables.

**Severity:** Error

**Auto-fix:** Yes (extracts to variable)

**Why:** Flat declarations are easier to read, modify, and analyze. Nested structures complicate dependency tracking.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "app",
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
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    Ports: []corev1.ContainerPort{
        {ContainerPort: 80},
    },
}

var AppPodSpec = corev1.PodSpec{
    Containers: []corev1.Container{AppContainer},
}

var AppPodTemplate = corev1.PodTemplateSpec{
    Spec: AppPodSpec,
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: AppPodTemplate,
    },
}
```

**Auto-fix:** Automatically extracts nested structures to variables with generated names.

---

### WK8003: Direct field references

**Description:** Resource references MUST use direct field access (e.g., `MyPod.Metadata.Name`), not function calls (e.g., `getName(MyPod)`).

**Severity:** Error

**Auto-fix:** No

**Why:** Direct field references enable static dependency analysis. Function calls require execution to resolve dependencies.

**Bad:**

```go
func getServiceName(svc corev1.Service) string {
    return svc.Metadata.Name
}

var MyIngress = networkingv1.Ingress{
    Spec: networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Backend: networkingv1.IngressBackend{
                                    Service: &networkingv1.IngressServiceBackend{
                                        Name: getServiceName(MyService),  // Function call
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

**Good:**

```go
var MyIngress = networkingv1.Ingress{
    Spec: networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Backend: networkingv1.IngressBackend{
                                    Service: &networkingv1.IngressServiceBackend{
                                        Name: MyService.Metadata.Name,  // Direct reference
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

---

### WK8004: No loops in resources

**Description:** Resource definitions MUST NOT contain loops (for, range). Extract repetitive patterns to variables or use helper functions that return values.

**Severity:** Error

**Auto-fix:** No

**Why:** Loops prevent static analysis and make code execution-dependent.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: func() []corev1.Container {
                    containers := []corev1.Container{}
                    for i := 0; i < 3; i++ {  // Loop
                        containers = append(containers, corev1.Container{
                            Name:  fmt.Sprintf("app-%d", i),
                            Image: "nginx:latest",
                        })
                    }
                    return containers
                }(),
            },
        },
    },
}
```

**Good:**

```go
var App0Container = corev1.Container{
    Name:  "app-0",
    Image: "nginx:latest",
}

var App1Container = corev1.Container{
    Name:  "app-1",
    Image: "nginx:latest",
}

var App2Container = corev1.Container{
    Name:  "app-2",
    Image: "nginx:latest",
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    App0Container,
                    App1Container,
                    App2Container,
                },
            },
        },
    },
}
```

**Note:** For truly dynamic patterns, define the slice before the resource declaration:

```go
var appContainers = []corev1.Container{
    {Name: "app-0", Image: "nginx:latest"},
    {Name: "app-1", Image: "nginx:latest"},
    {Name: "app-2", Image: "nginx:latest"},
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: appContainers,
            },
        },
    },
}
```

---

### WK8005: No conditionals in resources

**Description:** Resource definitions MUST NOT contain conditional logic (if/else, switch). Extract conditional logic outside resource declarations.

**Severity:** Error

**Auto-fix:** No

**Why:** Conditionals make resource definitions execution-dependent and harder to analyze.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: func() *int32 {
            if os.Getenv("ENV") == "prod" {
                return ptrInt32(5)
            }
            return ptrInt32(1)
        }(),
    },
}
```

**Good:**

```go
var replicaCount = func() *int32 {
    if os.Getenv("ENV") == "prod" {
        return ptrInt32(5)
    }
    return ptrInt32(1)
}()

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: replicaCount,
    },
}
```

---

### WK8006: Flat variable references

**Description:** Complex nested structures SHOULD be extracted to separate variables for improved readability.

**Severity:** Warning

**Auto-fix:** Yes (extracts to variable)

**Why:** Flat declarations are easier to read, modify, and reuse.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Volumes: []corev1.Volume{
                    {
                        Name: "config",
                        VolumeSource: corev1.VolumeSource{
                            ConfigMap: &corev1.ConfigMapVolumeSource{
                                LocalObjectReference: corev1.LocalObjectReference{
                                    Name: "app-config",
                                },
                                Items: []corev1.KeyToPath{
                                    {Key: "config.yaml", Path: "config.yaml"},
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

**Good:**

```go
var ConfigVolume = corev1.Volume{
    Name: "config",
    VolumeSource: corev1.VolumeSource{
        ConfigMap: &corev1.ConfigMapVolumeSource{
            LocalObjectReference: corev1.LocalObjectReference{
                Name: "app-config",
            },
            Items: []corev1.KeyToPath{
                {Key: "config.yaml", Path: "config.yaml"},
            },
        },
    },
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Volumes: []corev1.Volume{ConfigVolume},
            },
        },
    },
}
```

**Auto-fix:** Extracts nested structures automatically.

---

### WK8101: Selector label match

**Description:** Deployment/StatefulSet/DaemonSet selector labels MUST match template pod labels.

**Severity:** Error

**Auto-fix:** No (ambiguous which to change)

**Why:** Kubernetes requires selector and template labels to match. Mismatch causes deployment failure.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{
                "app": "myapp",
                "version": "v1",
            },
        },
        Template: corev1.PodTemplateSpec{
            Metadata: corev1.ObjectMeta{
                Labels: map[string]string{
                    "app": "myapp",  // Missing "version" label
                },
            },
        },
    },
}
```

**Good:**

```go
var appLabels = map[string]string{
    "app":     "myapp",
    "version": "v1",
}

var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &corev1.LabelSelector{
            MatchLabels: appLabels,
        },
        Template: corev1.PodTemplateSpec{
            Metadata: corev1.ObjectMeta{
                Labels: appLabels,
            },
        },
    },
}
```

---

### WK8102: Required metadata name

**Description:** All Kubernetes resources MUST have `Metadata.Name` set.

**Severity:** Error

**Auto-fix:** No

**Why:** Kubernetes requires all resources to have a name.

**Bad:**

```go
var MyPod = corev1.Pod{
    Metadata: corev1.ObjectMeta{
        Namespace: "default",
        // Missing Name
    },
    Spec: corev1.PodSpec{...},
}
```

**Good:**

```go
var MyPod = corev1.Pod{
    Metadata: corev1.ObjectMeta{
        Name:      "my-pod",
        Namespace: "default",
    },
    Spec: corev1.PodSpec{...},
}
```

---

### WK8103: Container name required

**Description:** All containers MUST have a `Name` field.

**Severity:** Error

**Auto-fix:** No

**Why:** Kubernetes requires all containers to have a name.

**Bad:**

```go
var AppContainer = corev1.Container{
    Image: "nginx:latest",
    // Missing Name
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
}
```

---

### WK8104: Port name recommended

**Description:** Container and Service ports SHOULD be named for better documentation and service mesh support.

**Severity:** Warning

**Auto-fix:** No

**Why:** Named ports improve clarity and enable features in service meshes like Istio.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    Ports: []corev1.ContainerPort{
        {ContainerPort: 80},  // Unnamed
    },
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    Ports: []corev1.ContainerPort{
        {Name: "http", ContainerPort: 80},
    },
}
```

---

### WK8105: ImagePullPolicy explicit

**Description:** `ImagePullPolicy` SHOULD be explicitly set rather than relying on defaults.

**Severity:** Warning

**Auto-fix:** Yes (sets to `IfNotPresent` for tagged images, `Always` for `:latest`)

**Why:** Explicit configuration prevents surprises from Kubernetes default behavior changes.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    // ImagePullPolicy not set
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:            "app",
    Image:           "nginx:latest",
    ImagePullPolicy: "Always",
}
```

**Auto-fix:** Sets to `Always` for `:latest` images, `IfNotPresent` for tagged images.

---

### WK8201: Resource limits required

**Description:** Containers SHOULD specify resource requests and limits.

**Severity:** Warning

**Auto-fix:** No

**Why:** Resource limits prevent containers from consuming excessive cluster resources and enable proper scheduling.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    // No Resources specified
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
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
}
```

---

### WK8202: SecurityContext recommended

**Description:** Pods and containers SHOULD define a SecurityContext.

**Severity:** Warning

**Auto-fix:** No

**Why:** SecurityContext enables security hardening (non-root users, read-only filesystem, etc.).

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    // No SecurityContext
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot:             ptrBool(true),
        RunAsUser:                ptrInt64(1000),
        ReadOnlyRootFilesystem:   ptrBool(true),
        AllowPrivilegeEscalation: ptrBool(false),
    },
}
```

---

### WK8203: ReadOnlyRootFilesystem

**Description:** Containers SHOULD set `ReadOnlyRootFilesystem: true` in SecurityContext.

**Severity:** Warning

**Auto-fix:** No

**Why:** Read-only filesystems reduce attack surface and prevent container compromise.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot: ptrBool(true),
        // ReadOnlyRootFilesystem not set
    },
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot:           ptrBool(true),
        ReadOnlyRootFilesystem: ptrBool(true),
    },
}
```

---

### WK8204: RunAsNonRoot

**Description:** Containers SHOULD set `RunAsNonRoot: true` in SecurityContext.

**Severity:** Warning

**Auto-fix:** No

**Why:** Running as root increases security risk. Non-root users limit potential damage from container compromise.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        ReadOnlyRootFilesystem: ptrBool(true),
        // RunAsNonRoot not set
    },
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot:           ptrBool(true),
        RunAsUser:              ptrInt64(1000),
        ReadOnlyRootFilesystem: ptrBool(true),
    },
}
```

---

### WK8205: Drop capabilities

**Description:** Containers SHOULD drop unnecessary Linux capabilities.

**Severity:** Warning

**Auto-fix:** No

**Why:** Dropping capabilities reduces attack surface by removing unnecessary privileges.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot: ptrBool(true),
        // Capabilities not set
    },
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        RunAsNonRoot: ptrBool(true),
        Capabilities: &corev1.Capabilities{
            Drop: []string{"ALL"},
        },
    },
}
```

---

### WK8206: No privileged containers

**Description:** Containers MUST NOT run in privileged mode unless absolutely necessary.

**Severity:** Error

**Auto-fix:** No

**Why:** Privileged containers have full access to the host and pose significant security risks.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        Privileged: ptrBool(true),  // Privileged
    },
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    SecurityContext: &corev1.SecurityContext{
        Privileged: ptrBool(false),
    },
}
```

---

### WK8207: No host network

**Description:** Pods SHOULD NOT use `HostNetwork: true` unless required.

**Severity:** Warning

**Auto-fix:** No

**Why:** Host network access bypasses network policies and exposes the pod to host-level network risks.

**Bad:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostNetwork: true,  // Host network
        Containers:  []corev1.Container{AppContainer},
    },
}
```

**Good:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostNetwork: false,
        Containers:  []corev1.Container{AppContainer},
    },
}
```

---

### WK8208: No host PID

**Description:** Pods SHOULD NOT use `HostPID: true` unless required.

**Severity:** Warning

**Auto-fix:** No

**Why:** Host PID namespace access allows viewing and potentially interfering with host processes.

**Bad:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostPID:    true,  // Host PID
        Containers: []corev1.Container{AppContainer},
    },
}
```

**Good:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostPID:    false,
        Containers: []corev1.Container{AppContainer},
    },
}
```

---

### WK8209: No host IPC

**Description:** Pods SHOULD NOT use `HostIPC: true` unless required.

**Severity:** Warning

**Auto-fix:** No

**Why:** Host IPC namespace access enables inter-process communication with host processes, increasing security risk.

**Bad:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostIPC:    true,  // Host IPC
        Containers: []corev1.Container{AppContainer},
    },
}
```

**Good:**

```go
var MyPod = corev1.Pod{
    Spec: corev1.PodSpec{
        HostIPC:    false,
        Containers: []corev1.Container{AppContainer},
    },
}
```

---

### WK8301: Health checks recommended

**Description:** Containers SHOULD define liveness and readiness probes.

**Severity:** Warning

**Auto-fix:** No

**Why:** Health checks enable Kubernetes to detect and recover from failures automatically.

**Bad:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    // No health checks
}
```

**Good:**

```go
var AppContainer = corev1.Container{
    Name:  "app",
    Image: "nginx:latest",
    LivenessProbe: &corev1.Probe{
        ProbeHandler: corev1.ProbeHandler{
            HTTPGet: &corev1.HTTPGetAction{
                Path: "/healthz",
                Port: intstr.FromInt(8080),
            },
        },
        InitialDelaySeconds: 10,
        PeriodSeconds:       10,
    },
    ReadinessProbe: &corev1.Probe{
        ProbeHandler: corev1.ProbeHandler{
            HTTPGet: &corev1.HTTPGetAction{
                Path: "/ready",
                Port: intstr.FromInt(8080),
            },
        },
        InitialDelaySeconds: 5,
        PeriodSeconds:       5,
    },
}
```

---

### WK8302: Replicas minimum

**Description:** Deployments SHOULD have at least 2 replicas for high availability.

**Severity:** Info

**Auto-fix:** No

**Why:** Multiple replicas enable zero-downtime updates and resilience to node failures.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(1),  // Single replica
    },
}
```

**Good:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(3),  // Multiple replicas
    },
}
```

---

### WK8303: PodDisruptionBudget

**Description:** High-availability deployments SHOULD have a PodDisruptionBudget.

**Severity:** Info

**Auto-fix:** No

**Why:** PodDisruptionBudgets ensure minimum availability during voluntary disruptions (node drains, updates).

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(5),
    },
}
// No PodDisruptionBudget
```

**Good:**

```go
var MyDeployment = appsv1.Deployment{
    Metadata: corev1.ObjectMeta{
        Name: "my-app",
        Labels: map[string]string{"app": "my-app"},
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(5),
    },
}

var MyPDB = policyv1.PodDisruptionBudget{
    Metadata: corev1.ObjectMeta{
        Name: "my-app-pdb",
    },
    Spec: policyv1.PodDisruptionBudgetSpec{
        MinAvailable: ptrIntOrString(intstr.FromInt(3)),
        Selector: &corev1.LabelSelector{
            MatchLabels: map[string]string{"app": "my-app"},
        },
    },
}
```

---

### WK8304: Anti-affinity recommended

**Description:** High-availability deployments SHOULD use pod anti-affinity to spread across nodes.

**Severity:** Info

**Auto-fix:** No

**Why:** Pod anti-affinity prevents all replicas from running on the same node, improving resilience.

**Bad:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(3),
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                // No affinity rules
                Containers: []corev1.Container{AppContainer},
            },
        },
    },
}
```

**Good:**

```go
var MyDeployment = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptrInt32(3),
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Affinity: &corev1.Affinity{
                    PodAntiAffinity: &corev1.PodAntiAffinity{
                        PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
                            {
                                Weight: 100,
                                PodAffinityTerm: corev1.PodAffinityTerm{
                                    LabelSelector: &corev1.LabelSelector{
                                        MatchLabels: map[string]string{"app": "my-app"},
                                    },
                                    TopologyKey: "kubernetes.io/hostname",
                                },
                            },
                        },
                    },
                },
                Containers: []corev1.Container{AppContainer},
            },
        },
    },
}
```

---

## Severity levels

- **Error** - Must be fixed, blocks build
- **Warning** - Should be fixed, doesn't block build
- **Info** - Optional best practice suggestion

## Disabling rules

Disable specific rules with `--disable`:

```bash
wetwire-k8s lint --disable WK8201,WK8202
```

Or in `.wetwire-k8s.yaml`:

```yaml
lint:
  disabled_rules:
    - WK8201
    - WK8202
```

## See also

- [CLI Reference](CLI.md) - Lint command documentation
- [FAQ](FAQ.md) - Common linting questions
- [CLAUDE.md](../CLAUDE.md) - AI assistant context
