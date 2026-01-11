package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8103: Container name required
// This file contains violations

// Bad: Container without name
var ContainerNoName = corev1.Container{
	Image: "nginx:1.21",
	// Missing Name field
}

// Bad: Pod with container without name
var PodContainerNoName = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Image: "nginx:1.21",
				// Missing Name field
			},
		},
	},
}

// Bad: Init container without name
var PodInitContainerNoName = corev1.Pod{
	Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Image: "busybox:1.35",
				// Missing Name field
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
