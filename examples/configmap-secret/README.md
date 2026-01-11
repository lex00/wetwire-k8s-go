# ConfigMap and Secret Example

Demonstrates ConfigMap and Secret usage with the wetwire pattern, showing various ways to inject configuration into pods.

## Overview

This example deploys:

- **AppConfig** - ConfigMap with application settings and configuration files
- **NginxConfig** - ConfigMap with nginx server configuration
- **AppSecrets** - Secret with sensitive credentials
- **TLSSecret** - Secret with TLS certificate and key
- **AppDeployment** - Pod that uses all ConfigMaps and Secrets

## Configuration Injection Methods

This example demonstrates multiple ways to use ConfigMaps and Secrets:

### 1. Single Environment Variable from ConfigMap

```go
Env: []corev1.EnvVar{
    {
        Name: "LOG_LEVEL",
        ValueFrom: &corev1.EnvVarSource{
            ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: AppConfig.Name,
                },
                Key: "LOG_LEVEL",
            },
        },
    },
}
```

### 2. Single Environment Variable from Secret

```go
Env: []corev1.EnvVar{
    {
        Name: "DATABASE_URL",
        ValueFrom: &corev1.EnvVarSource{
            SecretKeyRef: &corev1.SecretKeySelector{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: AppSecrets.Name,
                },
                Key: "DATABASE_URL",
            },
        },
    },
}
```

### 3. All Keys as Environment Variables

```go
EnvFrom: []corev1.EnvFromSource{
    {
        ConfigMapRef: &corev1.ConfigMapEnvSource{
            LocalObjectReference: corev1.LocalObjectReference{
                Name: AppConfig.Name,
            },
        },
    },
}
```

### 4. Mount as Directory

```go
VolumeMounts: []corev1.VolumeMount{
    {
        Name:      "config-volume",
        MountPath: "/etc/app",
        ReadOnly:  true,
    },
}

Volumes: []corev1.Volume{
    {
        Name: "config-volume",
        VolumeSource: corev1.VolumeSource{
            ConfigMap: &corev1.ConfigMapVolumeSource{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: AppConfig.Name,
                },
            },
        },
    },
}
```

### 5. Mount Single File with SubPath

```go
VolumeMounts: []corev1.VolumeMount{
    {
        Name:      "nginx-config",
        MountPath: "/etc/nginx/nginx.conf",
        SubPath:   "nginx.conf",
        ReadOnly:  true,
    },
}
```

## Wetwire Pattern Highlights

### Resource References

ConfigMap and Secret names are referenced directly:

```go
ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
    LocalObjectReference: corev1.LocalObjectReference{
        Name: AppConfig.Name,  // Direct reference
    },
    Key: "LOG_LEVEL",
}
```

This creates a dependency from the Deployment to the ConfigMap, ensuring the ConfigMap is created first.

### Secret Types

Different secret types for different use cases:

```go
// Opaque secret for arbitrary data
Type: corev1.SecretTypeOpaque

// TLS secret with specific keys
Type: corev1.SecretTypeTLS
```

## Build

Generate the Kubernetes manifests:

```bash
wetwire-k8s build ./examples/configmap-secret -o config-demo.yaml
```

## Deploy

```bash
kubectl apply -f config-demo.yaml
```

## Verify

Check the resources:

```bash
kubectl get configmaps
kubectl get secrets
kubectl get deployments config-demo
kubectl get pods -l app=config-demo
```

Verify configuration is mounted:

```bash
# Check environment variables
kubectl exec deploy/config-demo -- env | grep -E 'LOG_LEVEL|DATABASE_URL'

# Check mounted files
kubectl exec deploy/config-demo -- ls -la /etc/app
kubectl exec deploy/config-demo -- cat /etc/app/config.yaml
```

## Cleanup

```bash
kubectl delete -f config-demo.yaml
```

## Resources

| Resource | Kind | Description |
|----------|------|-------------|
| app-config | ConfigMap | Application settings |
| nginx-config | ConfigMap | Nginx configuration files |
| app-secrets | Secret | Database and API credentials |
| tls-secret | Secret | TLS certificate and key |
| config-demo | Deployment | Application using all configs |

## Security Notes

1. **Never commit real secrets** - This example uses placeholder values
2. **Use external secret management** - Consider Vault, AWS Secrets Manager, or sealed-secrets
3. **Restrict file permissions** - Secrets are mounted with mode 0400
4. **Use read-only mounts** - ConfigMaps and Secrets are mounted as read-only

## Best Practices

1. **Separate concerns** - Different ConfigMaps for different purposes
2. **Use StringData for secrets** - Kubernetes handles base64 encoding
3. **Reference resources directly** - Use `AppConfig.Name` instead of string literals
4. **Mount as read-only** - Prevent accidental modifications
5. **Set file permissions** - Especially important for secrets
