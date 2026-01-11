package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8207: No host network
// This file contains no violations

// Good: Pod without host network (default)
var PodNoHostNetwork = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Good: Pod explicitly not using host network
var PodExplicitNoHostNetwork = corev1.Pod{
	Spec: corev1.PodSpec{
		HostNetwork: false,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}
