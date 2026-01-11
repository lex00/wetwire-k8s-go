package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8001: Resources must be top-level variable declarations
// This file contains compliant code

// Good: Top-level variable with composite literal
var MyDeployment = appsv1.Deployment{
	Metadata: corev1.ObjectMeta{Name: "app"},
}

// Good: Top-level variable with explicit type
var MyPod corev1.Pod = corev1.Pod{
	Metadata: corev1.ObjectMeta{Name: "my-pod"},
}
