# Test Fixture Sources

This document provides attribution for the test fixtures used in the comprehensive test suite.

## Attribution

The test fixtures in this directory are inspired by and adapted from the following sources:

### ContainerSolutions/kubernetes-examples
- **Repository:** https://github.com/ContainerSolutions/kubernetes-examples
- **License:** Apache 2.0
- **Used for:** Minimal, focused examples demonstrating specific Kubernetes features
- **Patterns adapted:** Pod probes, resource limits, security contexts, volumes

### kubernetes/examples (Official)
- **Repository:** https://github.com/kubernetes/examples
- **License:** Apache 2.0
- **Used for:** Production-like patterns and best practices
- **Patterns adapted:** StatefulSet with PVC templates, DaemonSet patterns, multi-container pods

### Kubernetes Official Documentation
- **Source:** https://kubernetes.io/docs/
- **License:** CC BY 4.0
- **Used for:** Standard API object structures and field references

## Fixture Categories

| Directory | Count | Description |
|-----------|-------|-------------|
| pods/ | 6 | Pod variations (basic, probes, resources, volumes, init containers, security) |
| workloads/ | 5 | Deployment, StatefulSet, DaemonSet, ReplicaSet |
| services/ | 5 | ClusterIP, NodePort, LoadBalancer, Headless, ExternalName |
| config/ | 3 | ConfigMap, Secret, TLS Secret |
| storage/ | 3 | PersistentVolume, PersistentVolumeClaim, StorageClass |
| networking/ | 2 | Ingress, NetworkPolicy |
| batch/ | 2 | Job, CronJob |
| rbac/ | 5 | Role, ClusterRole, RoleBinding, ClusterRoleBinding, ServiceAccount |
| autoscaling/ | 4 | HPA, PDB, ResourceQuota, LimitRange |

**Total: 35 fixtures**

## License Compliance

All fixtures are either:
1. Original creations following Kubernetes API specifications
2. Adaptations from Apache 2.0 licensed sources with proper attribution

The test fixtures themselves are released under the same license as the wetwire-k8s-go project (Apache 2.0).
