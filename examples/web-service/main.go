// Package main demonstrates a web service with Deployment, Service, and Ingress.
// This example shows a complete production-ready web application setup with
// proper resource references between components.
package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Application Configuration
// =============================================================================

// appName is the application identifier used across all resources
const appName = "webapp"

// appLabels are shared labels for all resources
var appLabels = map[string]string{
	"app":     appName,
	"version": "v1",
}

// =============================================================================
// Deployment
// =============================================================================

// WebAppDeployment runs the web application containers
var WebAppDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:   appName,
		Labels: appLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(3)),
		Selector: &metav1.LabelSelector{
			MatchLabels: appLabels,
		},
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: ptr(intstr.FromString("25%")),
				MaxSurge:       ptr(intstr.FromString("25%")),
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: appLabels,
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
					"prometheus.io/port":   "8080",
					"prometheus.io/path":   "/metrics",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  appName,
						Image: "nginx:1.25-alpine",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromString("http"),
								},
							},
							InitialDelaySeconds: 10,
							PeriodSeconds:       10,
							TimeoutSeconds:      5,
							FailureThreshold:    3,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.FromString("http"),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
							TimeoutSeconds:      3,
							FailureThreshold:    3,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							ReadOnlyRootFilesystem:   ptr(true),
							RunAsNonRoot:             ptr(true),
							RunAsUser:                ptr(int64(1000)),
							AllowPrivilegeEscalation: ptr(false),
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{
					FSGroup: ptr(int64(1000)),
				},
				TerminationGracePeriodSeconds: ptr(int64(30)),
			},
		},
	},
}

// =============================================================================
// Service
// =============================================================================

// WebAppService exposes the web application within the cluster
var WebAppService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:   appName,
		Labels: appLabels,
	},
	Spec: corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: WebAppDeployment.Spec.Selector.MatchLabels,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromString("http"),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	},
}

// =============================================================================
// Ingress
// =============================================================================

// ingressClassName specifies the ingress controller to use
var ingressClassName = "nginx"

// WebAppIngress routes external traffic to the service
var WebAppIngress = &networkingv1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name:   appName,
		Labels: appLabels,
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/ssl-redirect": "true",
			"nginx.ingress.kubernetes.io/use-regex":    "false",
		},
	},
	Spec: networkingv1.IngressSpec{
		IngressClassName: &ingressClassName,
		TLS: []networkingv1.IngressTLS{
			{
				Hosts:      []string{"webapp.example.com"},
				SecretName: "webapp-tls",
			},
		},
		Rules: []networkingv1.IngressRule{
			{
				Host: "webapp.example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: ptr(networkingv1.PathTypePrefix),
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: WebAppService.Name,
										Port: networkingv1.ServiceBackendPort{
											Name: "http",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

// main is required for package main but not used - wetwire-k8s discovers resources from variable declarations
func main() {}
