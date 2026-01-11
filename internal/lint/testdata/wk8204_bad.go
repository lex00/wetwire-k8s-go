package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8204: RunAsNonRoot
// This file contains violations

// Helper function for bool pointer
func ptrBool8204(b bool) *bool {
	return &b
}

// Bad: Container without RunAsNonRoot
var ContainerNoRunAsNonRoot = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		ReadOnlyRootFilesystem: ptrBool8204(true),
		// Missing RunAsNonRoot
	},
}

// Bad: Container with RunAsNonRoot = false
var ContainerRunAsNonRootFalse = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8204(false),
	},
}

// Bad: Container without SecurityContext (no RunAsNonRoot)
var ContainerNoSecurityContext8204 = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	// No SecurityContext
}
