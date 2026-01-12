// Package main demonstrates HorizontalPodAutoscaler patterns.
//
// This example shows:
// - CPU-based autoscaling
// - Memory-based autoscaling
// - Custom metrics scaling
// - Scaling behavior configuration
//
// Use case: Web application that scales based on load.
package main

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Common Labels
// =============================================================================

var appLabels = map[string]string{
	"app.kubernetes.io/name":       "web-app",
	"app.kubernetes.io/managed-by": "wetwire-k8s",
}

// =============================================================================
// Deployment - Target for HPA
// =============================================================================

// WebAppDeployment is the target deployment for autoscaling.
var WebAppDeployment = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "web-app",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(2)), // Initial replicas, HPA will manage this
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app.kubernetes.io/name": "web-app"},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app.kubernetes.io/name": "web-app"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "web-app",
						Image: "nginx:1.25",
						Ports: []corev1.ContainerPort{
							{ContainerPort: 80},
						},
						// IMPORTANT: Resource requests are REQUIRED for HPA
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// =============================================================================
// HPA with CPU and Memory metrics
// =============================================================================

// WebAppHPA scales based on CPU and memory utilization.
var WebAppHPA = autoscalingv2.HorizontalPodAutoscaler{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "autoscaling/v2",
		Kind:       "HorizontalPodAutoscaler",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "web-app-hpa",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "web-app",
		},
		MinReplicas: ptr(int32(2)),
		MaxReplicas: 10,
		Metrics: []autoscalingv2.MetricSpec{
			{
				// Scale based on CPU utilization
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr(int32(70)),
					},
				},
			},
			{
				// Scale based on memory utilization
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceMemory,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr(int32(80)),
					},
				},
			},
		},
		Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
			ScaleUp: &autoscalingv2.HPAScalingRules{
				// Scale up aggressively
				StabilizationWindowSeconds: ptr(int32(0)),
				Policies: []autoscalingv2.HPAScalingPolicy{
					{
						Type:          autoscalingv2.PercentScalingPolicy,
						Value:         100,
						PeriodSeconds: 15,
					},
					{
						Type:          autoscalingv2.PodsScalingPolicy,
						Value:         4,
						PeriodSeconds: 15,
					},
				},
				SelectPolicy: ptr(autoscalingv2.MaxChangePolicySelect),
			},
			ScaleDown: &autoscalingv2.HPAScalingRules{
				// Scale down conservatively
				StabilizationWindowSeconds: ptr(int32(300)),
				Policies: []autoscalingv2.HPAScalingPolicy{
					{
						Type:          autoscalingv2.PercentScalingPolicy,
						Value:         10,
						PeriodSeconds: 60,
					},
				},
				SelectPolicy: ptr(autoscalingv2.MinChangePolicySelect),
			},
		},
	},
}

// =============================================================================
// HPA with custom metrics
// =============================================================================

func main() {}

// WebAppHPACustomMetrics scales based on custom metrics (requires metrics server).
var WebAppHPACustomMetrics = autoscalingv2.HorizontalPodAutoscaler{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "autoscaling/v2",
		Kind:       "HorizontalPodAutoscaler",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "web-app-hpa-custom",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "web-app",
		},
		MinReplicas: ptr(int32(2)),
		MaxReplicas: 20,
		Metrics: []autoscalingv2.MetricSpec{
			{
				// CPU as baseline
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr(int32(50)),
					},
				},
			},
			{
				// Custom metric: requests per second per pod
				// Requires metrics server (e.g., Prometheus Adapter)
				Type: autoscalingv2.PodsMetricSourceType,
				Pods: &autoscalingv2.PodsMetricSource{
					Metric: autoscalingv2.MetricIdentifier{
						Name: "http_requests_per_second",
					},
					Target: autoscalingv2.MetricTarget{
						Type:         autoscalingv2.AverageValueMetricType,
						AverageValue: ptr(resource.MustParse("1000")),
					},
				},
			},
		},
	},
}
