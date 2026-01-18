package argo

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ArgocdRedisDeployment = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "argocd-redis",
		Labels: map[string]string{
			"app.kubernetes.io/component": "redis",
			"app.kubernetes.io/name":      "argocd-redis",
			"app.kubernetes.io/part-of":   "argocd",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/name": "argocd-redis",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "argocd-redis",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "redis",
						Image: "redis:8.2.3-alpine",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 6379,
							},
						},
					},
				},
			},
		},
	},
}
