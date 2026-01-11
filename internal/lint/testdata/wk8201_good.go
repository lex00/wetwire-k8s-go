package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8201: Missing resource limits
// This file contains no violations

// Good: Container with both requests and limits
var DeploymentWithLimits = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:1.21",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu":    "100m",
								"memory": "128Mi",
							},
							Limits: corev1.ResourceList{
								"cpu":    "500m",
								"memory": "512Mi",
							},
						},
					},
				},
			},
		},
	},
}

// Good: Pod with resources on all containers
var PodWithAllResources = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "nginx:1.21",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    "100m",
						"memory": "128Mi",
					},
					Limits: corev1.ResourceList{
						"cpu":    "500m",
						"memory": "512Mi",
					},
				},
			},
			corev1.Container{
				Name:  "sidecar",
				Image: "busybox:1.35",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    "50m",
						"memory": "64Mi",
					},
					Limits: corev1.ResourceList{
						"cpu":    "200m",
						"memory": "256Mi",
					},
				},
			},
		},
	},
}
