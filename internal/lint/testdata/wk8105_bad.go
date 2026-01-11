package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8105: ImagePullPolicy explicit
// This file contains violations

// Bad: Container without ImagePullPolicy
var ContainerNoImagePullPolicy = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	// Missing ImagePullPolicy field
}

// Bad: Pod with container without ImagePullPolicy
var PodNoImagePullPolicy = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				// Missing ImagePullPolicy field
			},
		},
	},
}

// Bad: Multiple containers, one without ImagePullPolicy
var PodMixedImagePullPolicy = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:            "app",
				Image:           "nginx:1.21",
				ImagePullPolicy: corev1.PullIfNotPresent,
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				// Missing ImagePullPolicy field
			},
		},
	},
}
