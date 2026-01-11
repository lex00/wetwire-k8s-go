package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8302: Replicas minimum
// This file contains no violations

// Helper function for int32 pointer
func ptrInt32_8302Good(i int32) *int32 {
	return &i
}

// Good: Deployment with 2 replicas
var DeploymentTwoReplicas = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8302Good(2),
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

// Good: Deployment with 3 replicas
var DeploymentThreeReplicas = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8302Good(3),
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

// Good: Deployment with 5 replicas
var DeploymentFiveReplicas = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8302Good(5),
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
