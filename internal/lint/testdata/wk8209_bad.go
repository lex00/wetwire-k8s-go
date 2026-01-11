package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8209: No host IPC
// This file contains violations

// Bad: Pod using host IPC namespace
var PodHostIPC = corev1.Pod{
	Spec: corev1.PodSpec{
		HostIPC: true,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Bad: Pod with host IPC in a deployment context
var PodTemplateHostIPC = corev1.PodSpec{
	HostIPC: true,
	Containers: []corev1.Container{
		corev1.Container{
			Name:  "app",
			Image: "nginx:1.21",
		},
	},
}
