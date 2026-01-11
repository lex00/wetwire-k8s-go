package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8105: ImagePullPolicy explicit
// This file contains no violations

// Good: Container with ImagePullPolicy
var ContainerWithImagePullPolicy = corev1.Container{
	Name:            "app",
	Image:           "nginx:1.21",
	ImagePullPolicy: corev1.PullIfNotPresent,
}

// Good: Container with Always ImagePullPolicy
var ContainerWithAlways = corev1.Container{
	Name:            "app",
	Image:           "nginx:latest",
	ImagePullPolicy: corev1.PullAlways,
}

// Good: Pod with all containers having ImagePullPolicy
var PodAllImagePullPolicy = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:            "app",
				Image:           "nginx:1.21",
				ImagePullPolicy: corev1.PullIfNotPresent,
			},
			corev1.Container{
				Name:            "sidecar",
				Image:           "busybox:1.35",
				ImagePullPolicy: corev1.PullNever,
			},
		},
	},
}
