package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8006: Flag :latest image tags
// This file contains violations

// Bad: Using :latest tag
var DeploymentWithLatest = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:latest", // Using :latest
					},
				},
			},
		},
	},
}

// Bad: No tag (defaults to :latest)
var DeploymentWithNoTag = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx", // No tag, defaults to :latest
					},
				},
			},
		},
	},
}

// Bad: Init container with :latest
var PodWithLatestInit = corev1.Pod{
	Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Name:  "init",
				Image: "busybox:latest", // Using :latest
			},
		},
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}
