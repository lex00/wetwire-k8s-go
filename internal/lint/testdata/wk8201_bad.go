package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8201: Missing resource limits
// This file contains violations

// Bad: Container with no resource limits
var DeploymentNoLimits = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:1.21",
						// No Resources specified
					},
				},
			},
		},
	},
}

// Bad: Container with requests but no limits
var PodOnlyRequests = corev1.Pod{
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
					// No Limits
				},
			},
		},
	},
}

// Bad: Container with empty resource requirements
var PodEmptyResources = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:      "app",
				Image:     "nginx:1.21",
				Resources: corev1.ResourceRequirements{},
			},
		},
	},
}

// Bad: Multiple containers, one missing limits
var PodMixedResources = corev1.Pod{
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
				// No resources
			},
		},
	},
}
