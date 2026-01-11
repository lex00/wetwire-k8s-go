package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8205: Drop capabilities
// This file contains violations

// Helper function for bool pointer
func ptrBool8205(b bool) *bool {
	return &b
}

// Bad: Container without dropping capabilities
var ContainerNoDropCapabilities = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8205(true),
		// No Capabilities.Drop
	},
}

// Bad: Container with empty drop list
var ContainerEmptyDropCapabilities = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8205(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{},
		},
	},
}

// Bad: Container without SecurityContext
var ContainerNoSecurityContext8205 = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	// No SecurityContext
}
