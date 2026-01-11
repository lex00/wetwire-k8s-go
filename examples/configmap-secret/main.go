// Package main demonstrates ConfigMap and Secret usage with the wetwire pattern.
// This example shows how to define configuration and secrets, then mount them
// into pods using volume mounts and environment variables.
package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Shared Labels
// =============================================================================

// appLabels identifies all resources in this example
var appLabels = map[string]string{
	"app": "config-demo",
}

// =============================================================================
// ConfigMap - Application Configuration
// =============================================================================

// AppConfig holds application configuration as key-value pairs
var AppConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "app-config",
		Labels: appLabels,
	},
	Data: map[string]string{
		// Simple key-value configuration
		"LOG_LEVEL":    "info",
		"ENABLE_DEBUG": "false",
		"MAX_WORKERS":  "10",

		// Multi-line configuration file
		"config.yaml": `server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s
logging:
  level: info
  format: json
database:
  pool_size: 10
  idle_timeout: 5m
`,
	},
}

// NginxConfig holds nginx-specific configuration
var NginxConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "nginx-config",
		Labels: appLabels,
	},
	Data: map[string]string{
		"nginx.conf": `events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name localhost;

        location / {
            root /usr/share/nginx/html;
            index index.html;
        }

        location /health {
            return 200 'ok';
            add_header Content-Type text/plain;
        }
    }
}
`,
		"index.html": `<!DOCTYPE html>
<html>
<head>
    <title>Config Demo</title>
</head>
<body>
    <h1>Configuration Demo Application</h1>
    <p>This page is served by nginx with custom configuration.</p>
</body>
</html>
`,
	},
}

// =============================================================================
// Secret - Sensitive Configuration
// =============================================================================

// AppSecrets holds sensitive configuration data
// Note: In production, secrets should be managed externally (Vault, AWS Secrets Manager, etc.)
var AppSecrets = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "app-secrets",
		Labels: appLabels,
	},
	Type: corev1.SecretTypeOpaque,
	// StringData is automatically base64 encoded by Kubernetes
	StringData: map[string]string{
		"DATABASE_URL":      "postgres://user:password@db.example.com:5432/mydb",
		"API_KEY":           "sk-example-api-key-12345",
		"ENCRYPTION_KEY":    "aes256-secret-key-here",
		"credentials.json": `{
  "client_id": "my-client-id",
  "client_secret": "my-client-secret",
  "token_url": "https://auth.example.com/token"
}`,
	},
}

// TLSSecret holds TLS certificate and key for HTTPS
var TLSSecret = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "tls-secret",
		Labels: appLabels,
	},
	Type: corev1.SecretTypeTLS,
	// These would contain actual certificate data in production
	StringData: map[string]string{
		"tls.crt": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
		"tls.key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
	},
}

// =============================================================================
// Deployment - Using ConfigMaps and Secrets
// =============================================================================

// AppDeployment demonstrates various ways to use ConfigMaps and Secrets
var AppDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "config-demo",
		Labels: appLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(1)),
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
						Name:  "app",
						Image: "nginx:1.25-alpine",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
							},
						},
						// Environment variables from ConfigMap
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
							{
								Name: "MAX_WORKERS",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: AppConfig.Name,
										},
										Key: "MAX_WORKERS",
									},
								},
							},
							// Environment variables from Secret
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
							{
								Name: "API_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: AppSecrets.Name,
										},
										Key: "API_KEY",
									},
								},
							},
						},
						// All ConfigMap keys as environment variables
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: AppConfig.Name,
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							// Mount ConfigMap as directory
							{
								Name:      "config-volume",
								MountPath: "/etc/app",
								ReadOnly:  true,
							},
							// Mount nginx config
							{
								Name:      "nginx-config",
								MountPath: "/etc/nginx/nginx.conf",
								SubPath:   "nginx.conf",
								ReadOnly:  true,
							},
							{
								Name:      "nginx-config",
								MountPath: "/usr/share/nginx/html/index.html",
								SubPath:   "index.html",
								ReadOnly:  true,
							},
							// Mount Secret as directory
							{
								Name:      "secrets-volume",
								MountPath: "/etc/secrets",
								ReadOnly:  true,
							},
							// Mount TLS certificates
							{
								Name:      "tls-certs",
								MountPath: "/etc/tls",
								ReadOnly:  true,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					// ConfigMap as volume
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
					// Nginx config as volume
					{
						Name: "nginx-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: NginxConfig.Name,
								},
							},
						},
					},
					// Secret as volume
					{
						Name: "secrets-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: AppSecrets.Name,
								// Restrict file permissions
								DefaultMode: ptr(int32(0400)),
							},
						},
					},
					// TLS Secret as volume
					{
						Name: "tls-certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  TLSSecret.Name,
								DefaultMode: ptr(int32(0400)),
							},
						},
					},
				},
			},
		},
	},
}

// main is required for package main but not used - wetwire-k8s discovers resources from variable declarations
func main() {}
