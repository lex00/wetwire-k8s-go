// Package expected provides gold standard Kubernetes resources for the multi-tier scenario.
package expected

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// frontendLabels identifies the frontend tier pods
var frontendLabels = map[string]string{
	"app":  "ecommerce",
	"tier": "frontend",
}

// FrontendDeployment runs the e-commerce web frontend.
// Serves the user-facing web application with 3 replicas for high availability.
var FrontendDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "frontend",
		Namespace: "ecommerce",
		Labels:    frontendLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(3)),
		Selector: &metav1.LabelSelector{
			MatchLabels: frontendLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: frontendLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "frontend",
						Image: "ecommerce/frontend:latest",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8080,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// FrontendService exposes the frontend to external traffic via LoadBalancer.
// Maps external port 80 to container port 8080.
var FrontendService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "frontend",
		Namespace: "ecommerce",
		Labels:    frontendLabels,
	},
	Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeLoadBalancer,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
		},
		Selector: frontendLabels,
	},
}
