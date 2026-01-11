package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8102: Missing labels
// This file contains no violations

// Good: Deployment with labels
var DeploymentWithLabels = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "labeled-deploy",
		Labels: map[string]string{
			"app":     "myapp",
			"version": "v1",
		},
	},
}

// Good: Service with labels
var ServiceWithLabels = corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "labeled-svc",
		Labels: map[string]string{
			"app":       "myapp",
			"component": "api",
		},
	},
}

// Good: ConfigMap with labels
var ConfigMapWithLabels = corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "labeled-cm",
		Labels: map[string]string{
			"app":    "myapp",
			"config": "application",
		},
	},
}

// Good: Pod with labels
var PodWithLabels = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name: "labeled-pod",
		Labels: map[string]string{
			"app":  "myapp",
			"tier": "frontend",
		},
	},
}
