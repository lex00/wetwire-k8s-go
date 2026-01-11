package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8003: No duplicate resource names in same namespace
// This file contains compliant code

// Good: Different names
var PodA = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app-1",
		Namespace: "default",
	},
}

var PodB = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app-2",
		Namespace: "default",
	},
}

// Good: Same names but different namespaces
var DevPod = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app",
		Namespace: "dev",
	},
}

var ProdPod = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name:      "my-app",
		Namespace: "prod",
	},
}
