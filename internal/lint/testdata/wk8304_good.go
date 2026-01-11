package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WK8304: Anti-affinity recommended
// This file contains no violations

// Helper function for int32 pointer
func ptrInt32_8304Good(i int32) *int32 {
	return &i
}

// Good: HA deployment with pod anti-affinity
var HADeploymentWithAntiAffinity = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-with-affinity",
		Labels: map[string]string{
			"app": "ha-app-with-affinity",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8304Good(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "ha-app-with-affinity",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "ha-app-with-affinity",
				},
			},
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							corev1.WeightedPodAffinityTerm{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "ha-app-with-affinity",
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
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

// Good: Single replica deployment (no anti-affinity needed)
var SingleReplicaNoAntiAffinity = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "single-app",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8304Good(1),
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

// Good: HA deployment with required anti-affinity
var HADeploymentRequiredAntiAffinity = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-required",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8304Good(3),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "ha-app-required",
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
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
