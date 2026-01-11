package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8006: Flag :latest image tags
// This file contains compliant code

// Good: Using specific version tag
var DeploymentWithVersion = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: "nginx:1.21.6", // Specific version
					},
				},
			},
		},
	},
}

// Good: Using SHA digest
var DeploymentWithSHA = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: "nginx@sha256:abc123def456", // SHA digest
					},
				},
			},
		},
	},
}

// Good: Init container with specific version
var PodWithVersionedInit = corev1.Pod{
	Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:  "init",
				Image: "busybox:1.35.0", // Specific version
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "app",
				Image: "nginx:1.21.6",
			},
		},
	},
}
