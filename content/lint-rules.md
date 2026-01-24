---
title: "Lint Rules"
---


This document describes all lint rules for wetwire-k8s-go.

## Overview

The wetwire-k8s linter enforces flat, declarative patterns optimized for AI generation and human readability. Rules check for structural patterns, Kubernetes best practices, and security issues.

**Currently implemented: 26 rules** (15 structural/naming + 11 security/availability best practices)

## Rule naming convention

All rules follow the format: `WK8xxx`

- **W** - Wetwire
- **K8** - Kubernetes
- **xxx** - Three-digit number

## Rule index

| Rule | Description | Severity | Auto-fix |
|------|-------------|----------|----------|
| [WK8001](#wk8001-top-level-resource-declarations) | Resources must be top-level declarations | Error | No |
| [WK8002](#wk8002-avoid-deeply-nested-structures) | Avoid deeply nested inline structures | Error | No |
| [WK8003](#wk8003-no-duplicate-resource-names) | No duplicate resource names | Error | No |
| [WK8004](#wk8004-circular-dependency-detection) | Circular dependency detection | Error | No |
| [WK8005](#wk8005-flag-hardcoded-secrets) | Flag hardcoded secrets in env vars | Error | No |
| [WK8006](#wk8006-flag-latest-image-tags) | Flag :latest image tags | Error | No |
| [WK8041](#wk8041-hardcoded-api-keystokens) | Hardcoded API keys/tokens detected | Error | No |
| [WK8042](#wk8042-private-key-headers) | Private key headers detected | Error | No |
| [WK8101](#wk8101-selector-label-mismatch) | Selector labels must match template labels | Error | No |
| [WK8102](#wk8102-missing-labels) | Resources should have metadata labels | Warning | No |
| [WK8103](#wk8103-container-name-required) | Containers must have a Name field | Error | No |
| [WK8104](#wk8104-port-name-recommended) | Container and Service ports should be named | Warning | No |
| [WK8105](#wk8105-imagepullpolicy-explicit) | ImagePullPolicy should be explicitly set | Warning | Yes |
| [WK8201](#wk8201-missing-resource-limits) | Containers should have resource limits | Warning | No |
| [WK8202](#wk8202-privileged-containers) | Containers should not run in privileged mode | Error | No |
| [WK8203](#wk8203-readonlyrootfilesystem) | Containers should set ReadOnlyRootFilesystem | Warning | No |
| [WK8204](#wk8204-runasnonroot) | Containers should set RunAsNonRoot | Warning | No |
| [WK8205](#wk8205-drop-capabilities) | Containers should drop Linux capabilities | Warning | No |
| [WK8207](#wk8207-no-host-network) | Pods should not use HostNetwork | Warning | No |
| [WK8208](#wk8208-no-host-pid) | Pods should not use HostPID | Warning | No |
| [WK8209](#wk8209-no-host-ipc) | Pods should not use HostIPC | Warning | No |
| [WK8301](#wk8301-missing-health-probes) | Containers should have health probes | Warning | No |
| [WK8302](#wk8302-replicas-minimum) | Deployments should have 2+ replicas | Info | No |
| [WK8303](#wk8303-poddisruptionbudget) | HA deployments should have a PDB | Info | No |
| [WK8304](#wk8304-anti-affinity-recommended) | HA deployments should use pod anti-affinity | Info | No |
| [WK8401](#wk8401-file-size-limits) | Files should not exceed 20 resources | Warning | No |

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

### WK8102: Missing labels

**Description:** Resources SHOULD have metadata labels for organization and selection.

**Severity:** Warning

**Auto-fix:** No

**Why:** Labels enable querying, grouping, and managing resources effectively.

---

### WK8201: Missing resource limits

**Description:** Containers SHOULD specify resource limits (CPU, memory).

**Severity:** Warning

**Auto-fix:** No

**Why:** Resource limits prevent containers from consuming excessive cluster resources.

**Good:**

```go
var DeploymentWithLimits = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "app",
                        Image: "nginx:1.21",
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

---

### WK8202: Privileged containers

**Description:** Containers MUST NOT run in privileged mode.

**Severity:** Error

**Auto-fix:** No

**Why:** Privileged containers have full access to the host and pose significant security risks.

---

### WK8203: ReadOnlyRootFilesystem

**Description:** Containers SHOULD set `ReadOnlyRootFilesystem: true` in SecurityContext.

**Severity:** Warning

**Why:** Read-only filesystems reduce attack surface and prevent container compromise.

---

### WK8204: RunAsNonRoot

**Description:** Containers SHOULD set `RunAsNonRoot: true` in SecurityContext.

**Severity:** Warning

**Why:** Running as root increases security risk. Non-root users limit potential damage from container compromise.

---

### WK8205: Drop capabilities

**Description:** Containers SHOULD drop unnecessary Linux capabilities.

**Severity:** Warning

**Why:** Dropping capabilities reduces attack surface by removing unnecessary privileges.

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

### WK8207: No host network

**Description:** Pods SHOULD NOT use `HostNetwork: true` unless required.

**Severity:** Warning

**Why:** Host network access bypasses network policies and exposes the pod to host-level network risks.

---

### WK8208: No host PID

**Description:** Pods SHOULD NOT use `HostPID: true` unless required.

**Severity:** Warning

**Why:** Host PID namespace access allows viewing and potentially interfering with host processes.

---

### WK8209: No host IPC

**Description:** Pods SHOULD NOT use `HostIPC: true` unless required.

**Severity:** Warning

**Why:** Host IPC namespace access enables inter-process communication with host processes, increasing security risk.

---

### WK8301: Missing health probes

**Description:** Containers SHOULD have both liveness and readiness probes.

**Severity:** Warning

**Why:** Health probes enable Kubernetes to detect and recover from failures automatically.

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

**Why:** Multiple replicas enable zero-downtime updates and resilience to node failures.

---

### WK8303: PodDisruptionBudget

**Description:** High-availability deployments SHOULD have a PodDisruptionBudget.

**Severity:** Info

**Why:** PodDisruptionBudgets ensure minimum availability during voluntary disruptions (node drains, updates).

---

### WK8304: Anti-affinity recommended

**Description:** High-availability deployments SHOULD use pod anti-affinity to spread across nodes.

**Severity:** Info

**Why:** Pod anti-affinity prevents all replicas from running on the same node, improving resilience.

---

### WK8401: File size limits

**Description:** Files SHOULD NOT exceed 20 Kubernetes resources. Large files are harder to navigate and review. Consider splitting resources by concern (networking, compute, storage, etc.).

**Severity:** Warning

**Why:** Modular code organization improves readability and maintainability. Per the wetwire spec, files should stay under 500 lines and 20 resources.

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

- [CLI Reference](/cli/) - Lint command documentation
- [FAQ](/faq/) - Common linting questions
- [CLAUDE.md](../CLAUDE.md) - AI assistant context
