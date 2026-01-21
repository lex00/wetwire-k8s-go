package main

import (
	appsv1 "github.com/lex00/wetwire-k8s-go/resources/apps/v1"
	autoscalingv2 "github.com/lex00/wetwire-k8s-go/resources/autoscaling/v2"
	corev1 "github.com/lex00/wetwire-k8s-go/resources/core/v1"
	networkingv1 "github.com/lex00/wetwire-k8s-go/resources/networking/v1"
)

// Helper function for pointer to int32
func ptrInt32(i int32) *int32 {
	return &i
}

// Helper function for pointer to string
func ptrString(s string) *string {
	return &s
}

// Namespace for the entire application
var EcommerceNamespace = corev1.Namespace{
	Metadata: corev1.ObjectMeta{
		Name: "ecommerce",
		Labels: map[string]string{
			"app": "ecommerce",
		},
	},
}

// Database configuration stored safely in a Secret
var DatabaseSecret = corev1.Secret{
	Metadata: corev1.ObjectMeta{
		Name:      "postgres-secret",
		Namespace: "ecommerce",
	},
	Type: "Opaque",
	StringData: map[string]string{
		"POSTGRES_USER":     "ecommerce_user",
		"POSTGRES_PASSWORD": "changeme123",
		"POSTGRES_DB":       "ecommerce_db",
	},
}

// ConfigMap for application configuration
var AppConfig = corev1.ConfigMap{
	Metadata: corev1.ObjectMeta{
		Name:      "app-config",
		Namespace: "ecommerce",
	},
	Data: map[string]string{
		"DATABASE_HOST": "postgres-service",
		"DATABASE_PORT": "5432",
		"BACKEND_URL":   "http://backend-service:8080",
	},
}

// PostgreSQL StatefulSet for persistent database
var PostgresStatefulSet = appsv1.StatefulSet{
	Metadata: corev1.ObjectMeta{
		Name:      "postgres",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "database",
		},
	},
	Spec: appsv1.StatefulSetSpec{
		ServiceName: "postgres-service",
		Replicas:    ptrInt32(1),
		Selector: &corev1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "ecommerce",
				"tier": "database",
			},
		},
		Template: corev1.PodTemplateSpec{
			Metadata: corev1.ObjectMeta{
				Labels: map[string]string{
					"app":  "ecommerce",
					"tier": "database",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "postgres",
						Image: "postgres:15-alpine",
						Ports: []corev1.ContainerPort{
							{ContainerPort: 5432, Name: "postgres"},
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								SecretRef: &corev1.SecretEnvSource{
									Name: DatabaseSecret.Metadata.Name,
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[string]string{
								"cpu":    "500m",
								"memory": "1Gi",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "postgres-storage",
								MountPath: "/var/lib/postgresql/data",
							},
						},
					},
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				Metadata: corev1.ObjectMeta{
					Name: "postgres-storage",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []string{"ReadWriteOnce"},
					Resources: corev1.VolumeResourceRequirements{
						Requests: map[string]string{
							"storage": "10Gi",
						},
					},
				},
			},
		},
	},
}

// Service for PostgreSQL - internal only
var PostgresService = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name:      "postgres-service",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "database",
		},
	},
	Spec: corev1.ServiceSpec{
		Type: "ClusterIP",
		Ports: []corev1.ServicePort{
			{
				Port:       5432,
				TargetPort: "5432",
				Name:       "postgres",
			},
		},
		Selector: map[string]string{
			"app":  "ecommerce",
			"tier": "database",
		},
	},
}

// Backend API Deployment
var BackendDeployment = appsv1.Deployment{
	Metadata: corev1.ObjectMeta{
		Name:      "backend",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "backend",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32(2),
		Selector: &corev1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "ecommerce",
				"tier": "backend",
			},
		},
		Template: corev1.PodTemplateSpec{
			Metadata: corev1.ObjectMeta{
				Labels: map[string]string{
					"app":  "ecommerce",
					"tier": "backend",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "backend",
						Image: "ecommerce/backend:latest",
						Ports: []corev1.ContainerPort{
							{ContainerPort: 8080, Name: "http"},
						},
						Env: []corev1.EnvVar{
							{
								Name: "DATABASE_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Name: DatabaseSecret.Metadata.Name,
										Key:  "POSTGRES_USER",
									},
								},
							},
							{
								Name: "DATABASE_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Name: DatabaseSecret.Metadata.Name,
										Key:  "POSTGRES_PASSWORD",
									},
								},
							},
							{
								Name: "DATABASE_NAME",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Name: DatabaseSecret.Metadata.Name,
										Key:  "POSTGRES_DB",
									},
								},
							},
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									Name: AppConfig.Metadata.Name,
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[string]string{
								"cpu":    "300m",
								"memory": "512Mi",
							},
						},
					},
				},
			},
		},
	},
}

// Service for Backend API - internal only
var BackendService = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name:      "backend-service",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "backend",
		},
	},
	Spec: corev1.ServiceSpec{
		Type: "ClusterIP",
		Ports: []corev1.ServicePort{
			{
				Port:       8080,
				TargetPort: "8080",
				Name:       "http",
			},
		},
		Selector: map[string]string{
			"app":  "ecommerce",
			"tier": "backend",
		},
	},
}

// Frontend Web UI Deployment
var FrontendDeployment = appsv1.Deployment{
	Metadata: corev1.ObjectMeta{
		Name:      "frontend",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "frontend",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32(3),
		Selector: &corev1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "ecommerce",
				"tier": "frontend",
			},
		},
		Template: corev1.PodTemplateSpec{
			Metadata: corev1.ObjectMeta{
				Labels: map[string]string{
					"app":  "ecommerce",
					"tier": "frontend",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "frontend",
						Image: "ecommerce/frontend:latest",
						Ports: []corev1.ContainerPort{
							{ContainerPort: 8080, Name: "http"},
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									Name: AppConfig.Metadata.Name,
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[string]string{
								"cpu":    "200m",
								"memory": "256Mi",
							},
						},
					},
				},
			},
		},
	},
}

// Service for Frontend - LoadBalancer for internet access
var FrontendService = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name:      "frontend-service",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "frontend",
		},
	},
	Spec: corev1.ServiceSpec{
		Type: "LoadBalancer",
		Ports: []corev1.ServicePort{
			{
				Port:       80,
				TargetPort: "8080",
				Name:       "http",
			},
		},
		Selector: map[string]string{
			"app":  "ecommerce",
			"tier": "frontend",
		},
	},
}

// Horizontal Pod Autoscaler for Frontend - scales based on CPU
var FrontendHPA = autoscalingv2.HorizontalPodAutoscaler{
	Metadata: corev1.ObjectMeta{
		Name:      "frontend-hpa",
		Namespace: "ecommerce",
	},
	Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       FrontendDeployment.Metadata.Name,
		},
		MinReplicas: ptrInt32(3),
		MaxReplicas: 10,
		Metrics: []autoscalingv2.MetricSpec{
			{
				Type: "Resource",
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: "cpu",
					Target: autoscalingv2.MetricTarget{
						Type:               "Utilization",
						AverageUtilization: ptrInt32(70),
					},
				},
			},
		},
	},
}

// Network Policy - Allow frontend to talk to backend
var FrontendToBackendPolicy = networkingv1.NetworkPolicy{
	Metadata: corev1.ObjectMeta{
		Name:      "frontend-to-backend",
		Namespace: "ecommerce",
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: corev1.LabelSelector{
			MatchLabels: map[string]string{
				"tier": "backend",
			},
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &corev1.LabelSelector{
							MatchLabels: map[string]string{
								"tier": "frontend",
							},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Port: ptrString("8080"),
					},
				},
			},
		},
		PolicyTypes: []string{"Ingress"},
	},
}

// Network Policy - Allow backend to talk to database
var BackendToDatabasePolicy = networkingv1.NetworkPolicy{
	Metadata: corev1.ObjectMeta{
		Name:      "backend-to-database",
		Namespace: "ecommerce",
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: corev1.LabelSelector{
			MatchLabels: map[string]string{
				"tier": "database",
			},
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &corev1.LabelSelector{
							MatchLabels: map[string]string{
								"tier": "backend",
							},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Port: ptrString("5432"),
					},
				},
			},
		},
		PolicyTypes: []string{"Ingress"},
	},
}
