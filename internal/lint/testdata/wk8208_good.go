package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8208: No host PID
// This file contains no violations

// Good: Pod without host PID (default)
var PodNoHostPID = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Good: Pod explicitly not using host PID
var PodExplicitNoHostPID = corev1.Pod{
	Spec: corev1.PodSpec{
		HostPID: false,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}
