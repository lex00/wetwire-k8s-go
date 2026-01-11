package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8003: No duplicate resource names in same namespace
// This file contains violations

// Bad: Duplicate names in default namespace
var Pod1 = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app",
		Namespace: "default",
	},
}

var Pod2 = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app", // Duplicate name in same namespace
		Namespace: "default",
	},
}

// Bad: Duplicate names without namespace (defaults to "default")
var Service1 = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name: "my-service",
	},
}

var Service2 = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name: "my-service", // Duplicate
	},
}
