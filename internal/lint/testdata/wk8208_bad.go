package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8208: No host PID
// This file contains violations

// Bad: Pod using host PID namespace
var PodHostPID = corev1.Pod{
	Spec: corev1.PodSpec{
		HostPID: true,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Bad: Pod with host PID in a deployment context
var PodTemplateHostPID = corev1.PodSpec{
	HostPID: true,
	Containers: []corev1.Container{
		corev1.Container{
			Name:  "app",
			Image: "nginx:1.21",
		},
	},
}
