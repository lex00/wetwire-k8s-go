package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Simple deployment with no dependencies
var SimpleDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "simple-app",
	},
}

// Simple service with no dependencies
var SimpleService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "simple-svc",
	},
}
