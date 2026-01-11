package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8101: Selector label mismatch
// This file contains violations

// Bad: Selector has labels that template doesn't have
var DeploymentMismatchLabels = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":     "myapp",
				"version": "v1",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "myapp",
					// Missing "version" label
				},
			},
		},
	},
}

// Bad: Completely different labels
var DeploymentWrongLabels = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "frontend",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "backend", // Different value
				},
			},
		},
	},
}

// Bad: StatefulSet with label mismatch
var StatefulSetMismatch = appsv1.StatefulSet{
	Spec: appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "database",
				"tier": "db",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "database",
					// Missing "tier" label
				},
			},
		},
	},
}

// Bad: DaemonSet with no template labels
var DaemonSetNoLabels = appsv1.DaemonSet{
	Spec: appsv1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "monitoring",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
		},
	},
}
