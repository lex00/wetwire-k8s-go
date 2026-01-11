package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8005: Flag hardcoded secrets in env vars
// This file contains violations

// Bad: Hardcoded password
var PodWithPassword = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "DB_PASSWORD",
						Value: "hardcoded-password", // Hardcoded secret
					},
				},
			},
		},
	},
}

// Bad: Hardcoded API key
var PodWithAPIKey = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "API_KEY",
						Value: "sk_live_abc123", // Hardcoded secret
					},
				},
			},
		},
	},
}

// Bad: Hardcoded token
var PodWithToken = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "AUTH_TOKEN",
						Value: "my-secret-token", // Hardcoded secret
					},
				},
			},
		},
	},
}
