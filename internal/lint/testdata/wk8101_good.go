package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8101: Selector label mismatch
// This file contains no violations

// Good: Selector labels match template labels
var DeploymentMatchingLabels = appsv1.Deployment{
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
					"app":     "myapp",
					"version": "v1",
				},
			},
		},
	},
}

// Good: Template has extra labels (subset is OK)
var DeploymentExtraLabels = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "myapp",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":         "myapp",
					"version":     "v1",
					"environment": "prod",
				},
			},
		},
	},
}

// Good: StatefulSet with matching labels
var StatefulSetMatching = appsv1.StatefulSet{
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
					"app":  "database",
					"tier": "db",
				},
			},
		},
	},
}

// Good: Using shared label map
var sharedLabels = map[string]string{
	"app":     "frontend",
	"version": "v2",
}

var DeploymentSharedLabels = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: sharedLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: sharedLabels,
			},
		},
	},
}
