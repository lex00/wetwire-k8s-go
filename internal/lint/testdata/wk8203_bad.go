package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8203: ReadOnlyRootFilesystem
// This file contains violations

// Helper function for bool pointer
func ptrBool8203(b bool) *bool {
	return &b
}

// Bad: Container without ReadOnlyRootFilesystem
var ContainerNoReadOnlyFS = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8203(true),
		// Missing ReadOnlyRootFilesystem
	},
}

// Bad: Container with ReadOnlyRootFilesystem = false
var ContainerReadOnlyFSFalse = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot:           ptrBool8203(true),
		ReadOnlyRootFilesystem: ptrBool8203(false),
	},
}

// Bad: Container without SecurityContext (no ReadOnlyRootFilesystem)
var ContainerNoSecurityContext8203 = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	// No SecurityContext
}
