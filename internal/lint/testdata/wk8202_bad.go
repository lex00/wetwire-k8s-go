package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8202: Privileged containers
// This file contains violations

// Helper function for bool pointer
func ptrBool(b bool) *bool {
	return &b
}

// Bad: Privileged container
var PodPrivileged = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBool(true), // Privileged mode
				},
			},
		},
	},
}

// Bad: Init container privileged
var PodPrivilegedInit = corev1.Pod{
	Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Name:  "init",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBool(true),
				},
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

// Bad: Multiple containers, one privileged
var PodMixedPrivileged = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBool(false),
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBool(true), // Privileged
				},
			},
		},
	},
}
