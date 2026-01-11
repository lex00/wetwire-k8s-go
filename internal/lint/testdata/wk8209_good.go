package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8209: No host IPC
// This file contains no violations

// Good: Pod without host IPC (default)
var PodNoHostIPC = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Good: Pod explicitly not using host IPC
var PodExplicitNoHostIPC = corev1.Pod{
	Spec: corev1.PodSpec{
		HostIPC: false,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}
