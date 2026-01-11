package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8303: PodDisruptionBudget
// This file contains violations - HA deployment without PDB

// Helper function for int32 pointer
func ptrInt32_8303(i int32) *int32 {
	return &i
}

// Bad: HA deployment (3+ replicas) without PDB in the same file
var HADeploymentNoPDB = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app",
		Labels: map[string]string{
			"app": "ha-app",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8303(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "ha-app",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "ha-app",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:1.21",
					},
				},
			},
		},
	},
}

// No PDB is defined in this file
