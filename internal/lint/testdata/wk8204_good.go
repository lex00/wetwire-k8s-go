package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8204: RunAsNonRoot
// This file contains no violations

// Helper function for bool pointer
func ptrBool8204Good(b bool) *bool {
	return &b
}

// Good: Container with RunAsNonRoot = true
var ContainerRunAsNonRootTrue = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8204Good(true),
	},
}

// Good: All containers with RunAsNonRoot
var PodAllRunAsNonRoot = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot: ptrBool8204Good(true),
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot: ptrBool8204Good(true),
				},
			},
		},
	},
}
