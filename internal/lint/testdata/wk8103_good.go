package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8103: Container name required
// This file contains no violations

// Good: Container with name
var ContainerWithName = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
}

// Good: Pod with named container
var PodContainerWithName = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Good: Init container with name
var PodInitContainerWithName = corev1.Pod{
	Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Name:  "init",
				Image: "busybox:1.35",
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
