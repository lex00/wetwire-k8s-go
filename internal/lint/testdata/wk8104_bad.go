package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8104: Port name recommended
// This file contains violations

// Bad: Container port without name
var ContainerUnnamedPort = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	Ports: []corev1.ContainerPort{
		corev1.ContainerPort{
			ContainerPort: 80,
			// Missing Name field
		},
	},
}

// Bad: Multiple ports, one unnamed
var ContainerMixedPorts = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	Ports: []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 80,
		},
		corev1.ContainerPort{
			ContainerPort: 443,
			// Missing Name field
		},
	},
}

// Bad: Service port without name
var ServiceUnnamedPort = corev1.Service{
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Port: 80,
				// Missing Name field
			},
		},
	},
}
