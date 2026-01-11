package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8202: Privileged containers
// This file contains no violations

// Helper function for bool pointer
func ptrBoolGood(b bool) *bool {
	return &b
}

// Good: Non-privileged container
var PodNotPrivileged = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBoolGood(false),
				},
			},
		},
	},
}

// Good: No security context (defaults to non-privileged)
var PodNoSecurityContext = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				// No SecurityContext, defaults to non-privileged
			},
		},
	},
}

// Good: Multiple containers, all non-privileged
var PodAllNonPrivileged = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					Privileged:               ptrBoolGood(false),
					RunAsNonRoot:             ptrBoolGood(true),
					AllowPrivilegeEscalation: ptrBoolGood(false),
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptrBoolGood(false),
				},
			},
		},
	},
}
