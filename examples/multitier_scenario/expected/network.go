// Package expected provides gold standard Kubernetes resources for the multi-tier scenario.
package expected

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackendNetworkPolicy restricts network access to the backend tier.
// Only allows ingress traffic from frontend tier pods for security.
var BackendNetworkPolicy = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "backend-policy",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app": "ecommerce",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"tier": "backend",
			},
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"tier": "frontend",
							},
						},
					},
				},
			},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
		},
	},
}

// FrontendHPA auto-scales the frontend deployment based on CPU usage.
// Scales between 3-10 replicas to handle traffic variations.
var FrontendHPA = &autoscalingv2.HorizontalPodAutoscaler{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "frontend-hpa",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "frontend",
		},
	},
	Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "frontend",
		},
		MinReplicas: ptr(int32(3)),
		MaxReplicas: 10,
		Metrics: []autoscalingv2.MetricSpec{
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr(int32(70)),
					},
				},
			},
		},
	},
}
