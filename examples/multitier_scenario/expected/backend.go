// Package expected provides gold standard Kubernetes resources for the multi-tier scenario.
package expected

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// backendLabels identifies the backend tier pods
var backendLabels = map[string]string{
	"app":  "ecommerce",
	"tier": "backend",
}

// BackendDeployment runs the e-commerce REST API backend.
// Handles business logic and database interactions with 2 replicas.
var BackendDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "backend",
		Namespace: "ecommerce",
		Labels:    backendLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(2)),
		Selector: &metav1.LabelSelector{
			MatchLabels: backendLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: backendLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "backend",
						Image: "ecommerce/backend:latest",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8080,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name: "DATABASE_HOST",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "app-config",
										},
										Key: "database.host",
									},
								},
							},
							{
								Name: "DATABASE_PORT",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "app-config",
										},
										Key: "database.port",
									},
								},
							},
							{
								Name: "DATABASE_NAME",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "app-config",
										},
										Key: "database.name",
									},
								},
							},
							{
								Name: "DATABASE_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "db-credentials",
										},
										Key: "username",
									},
								},
							},
							{
								Name: "DATABASE_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "db-credentials",
										},
										Key: "password",
									},
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("300m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// BackendService exposes the backend API to other pods in the cluster.
// Uses ClusterIP for internal-only access.
var BackendService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "backend",
		Namespace: "ecommerce",
		Labels:    backendLabels,
	},
	Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		},
		Selector: backendLabels,
	},
}
