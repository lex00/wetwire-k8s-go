package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8104: Port name recommended
// This file contains no violations

// Good: Container port with name
var ContainerNamedPort = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	Ports: []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 80,
		},
	},
}

// Good: Multiple named ports
var ContainerMultipleNamedPorts = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
	Ports: []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 80,
		},
		corev1.ContainerPort{
			Name:          "https",
			ContainerPort: 443,
		},
	},
}

// Good: Service port with name
var ServiceNamedPort = corev1.Service{
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Name: "http",
				Port: 80,
			},
		},
	},
}

// Good: Container with no ports (no violation)
var ContainerNoPorts = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
}
