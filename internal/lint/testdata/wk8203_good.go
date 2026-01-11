package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8203: ReadOnlyRootFilesystem
// This file contains no violations

// Helper function for bool pointer
func ptrBool8203Good(b bool) *bool {
	return &b
}

// Good: Container with ReadOnlyRootFilesystem = true
var ContainerReadOnlyFSTrue = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot:           ptrBool8203Good(true),
		ReadOnlyRootFilesystem: ptrBool8203Good(true),
	},
}

// Good: All containers with ReadOnlyRootFilesystem
var PodAllReadOnlyFS = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					ReadOnlyRootFilesystem: ptrBool8203Good(true),
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					ReadOnlyRootFilesystem: ptrBool8203Good(true),
				},
			},
		},
	},
}
