package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8041: Hardcoded API keys/tokens
// This file contains no violations

// Good: Using secret reference
var PodWithSecretRef = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name: "AUTH_HEADER",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "api-tokens",
								},
								Key: "bearer-token",
							},
						},
					},
				},
			},
		},
	},
}

// Good: ConfigMap with safe values
var ConfigMapSafe = corev1.ConfigMap{
	Data: map[string]string{
		"config.yaml": "api_url=https://api.example.com",
	},
}

// Good: Environment variables without sensitive patterns
var PodWithSafeEnv = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "APP_NAME",
						Value: "my-application",
					},
					corev1.EnvVar{
						Name:  "LOG_LEVEL",
						Value: "info",
					},
				},
			},
		},
	},
}
