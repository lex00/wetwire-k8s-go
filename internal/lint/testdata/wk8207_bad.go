package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8207: No host network
// This file contains violations

// Bad: Pod using host network
var PodHostNetwork = corev1.Pod{
	Spec: corev1.PodSpec{
		HostNetwork: true,
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
			},
		},
	},
}

// Bad: Pod with host network in a deployment context
var PodTemplateHostNetwork = corev1.PodSpec{
	HostNetwork: true,
	Containers: []corev1.Container{
		corev1.Container{
			Name:  "app",
			Image: "nginx:1.21",
		},
	},
}
