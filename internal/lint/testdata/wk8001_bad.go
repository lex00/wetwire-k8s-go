package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8001: Resources must be top-level variable declarations
// This file contains violations

// Bad: Function returning a resource
func CreateDeployment(name string) appsv1.Deployment {
	return appsv1.Deployment{
		Metadata: corev1.ObjectMeta{Name: name},
	}
}

// Bad: Variable assigned from function call
var MyDeploy = CreateDeployment("app")

// Bad: Nested inside a function
func setupResources() {
	var NestedPod = corev1.Pod{
		Metadata: corev1.ObjectMeta{Name: "nested"},
	}
	_ = NestedPod
}
