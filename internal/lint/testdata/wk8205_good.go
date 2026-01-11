package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8205: Drop capabilities
// This file contains no violations

// Helper function for bool pointer
func ptrBool8205Good(b bool) *bool {
	return &b
}

// Good: Container dropping all capabilities
var ContainerDropAllCapabilities = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8205Good(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	},
}

// Good: Container dropping specific capabilities
var ContainerDropSpecificCapabilities = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	SecurityContext: &corev1.SecurityContext{
		RunAsNonRoot: ptrBool8205Good(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"NET_RAW", "SYS_ADMIN"},
		},
	},
}

// Good: All containers dropping capabilities
var PodAllDropCapabilities = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
			},
		},
	},
}
