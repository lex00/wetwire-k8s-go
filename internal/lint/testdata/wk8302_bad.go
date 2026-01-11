package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8302: Replicas minimum
// This file contains violations

// Helper function for int32 pointer
func ptrInt32_8302(i int32) *int32 {
	return &i
}

// Bad: Deployment with only 1 replica
var DeploymentSingleReplica = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8302(1),
		Template: corev1.PodTemplateSpec{
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

// Bad: Deployment with 0 replicas
var DeploymentZeroReplicas = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8302(0),
		Template: corev1.PodTemplateSpec{
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

// Bad: Deployment without replicas (defaults to 1)
var DeploymentNoReplicas = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
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
