package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8005: Flag hardcoded secrets in env vars
// This file contains compliant code

// Good: Using secret reference
var PodWithSecretRef = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					{
						Name: "DB_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "db-secret",
								},
								Key: "password",
							},
						},
					},
				},
			},
		},
	},
}

// Good: Non-sensitive hardcoded values
var PodWithConfig = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					{
						Name:  "APP_NAME",
						Value: "my-application", // Not a secret
					},
					{
						Name:  "LOG_LEVEL",
						Value: "info", // Not a secret
					},
				},
			},
		},
	},
}
