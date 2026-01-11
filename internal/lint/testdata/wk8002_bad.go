package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8002: Avoid deeply nested inline structures (max depth 5)
// This file contains violations

// Bad: Nesting depth > 5
var DeeplyNestedDeployment = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{ // depth 1
		Template: corev1.PodTemplateSpec{ // depth 2
			Spec: corev1.PodSpec{ // depth 3
				Containers: []corev1.Container{ // depth 4
					{ // depth 5
						Name:  "app",
						Image: "nginx:latest",
						Env: []corev1.EnvVar{ // depth 6 - exceeds limit
							{
								Name:  "CONFIG",
								Value: "value",
							},
						},
					},
				},
			},
		},
	},
}
