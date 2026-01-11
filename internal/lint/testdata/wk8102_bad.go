package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8102: Missing labels
// This file contains violations

// Bad: Deployment with no metadata labels
var DeploymentNoLabels = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "no-labels-deploy",
		// Labels is nil/empty
	},
}

// Bad: Service with empty labels
var ServiceEmptyLabels = corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "no-labels-svc",
		Labels: map[string]string{},
	},
}

// Bad: ConfigMap with no labels
var ConfigMapNoLabels = corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "no-labels-cm",
	},
}

// Bad: Pod with nil labels
var PodNilLabels = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "no-labels-pod",
		Labels: nil,
	},
}
