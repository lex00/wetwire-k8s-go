package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8304: Anti-affinity recommended
// This file contains violations - HA deployment without anti-affinity

// Helper function for int32 pointer
func ptrInt32_8304(i int32) *int32 {
	return &i
}

// Bad: HA deployment without anti-affinity
var HADeploymentNoAntiAffinity = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-no-affinity",
		Labels: map[string]string{
			"app": "ha-app-no-affinity",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8304(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "ha-app-no-affinity",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "ha-app-no-affinity",
				},
			},
			Spec: corev1.PodSpec{
				// No Affinity configured
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

// Bad: HA deployment with only node affinity (no pod anti-affinity)
var HADeploymentOnlyNodeAffinity = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-node-affinity",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8304(3),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						// Only node affinity, no pod anti-affinity
					},
				},
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
