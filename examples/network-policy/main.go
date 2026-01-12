// Package main demonstrates NetworkPolicy patterns.
//
// This example shows:
// - Default deny all (baseline security)
// - Ingress traffic control (incoming connections)
// - Egress traffic control (outgoing connections)
// - Pod-to-pod communication rules
// - Cross-namespace access (monitoring)
//
// Use case: A web application with frontend, backend, and database tiers
// where each tier can only communicate with adjacent tiers.
package main

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// Default Deny All - Baseline security
// =============================================================================

// DefaultDenyAll blocks all ingress and egress traffic by default.
var DefaultDenyAll = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default-deny-all",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		// Applies to all pods in the namespace
		PodSelector: metav1.LabelSelector{},
		// Deny all ingress and egress by default
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
	},
}

// =============================================================================
// Frontend Policy - Allow external ingress, backend egress
// =============================================================================

// FrontendPolicy allows external traffic in, backend traffic out.
var FrontendPolicy = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "frontend-policy",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "frontend",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"tier": "frontend"},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				// Allow traffic from anywhere (external users)
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(80))},
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(443))},
				},
			},
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				// Allow traffic to backend tier only
				To: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "backend"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(8080))},
				},
			},
			{
				// Allow DNS resolution
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolUDP), Port: ptr(intstr.FromInt(53))},
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(53))},
				},
			},
		},
	},
}

// =============================================================================
// Backend Policy - Allow frontend ingress, database egress
// =============================================================================

// BackendPolicy allows frontend traffic in, database traffic out.
var BackendPolicy = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "backend-policy",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "backend",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"tier": "backend"},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				// Only allow traffic from frontend tier
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "frontend"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(8080))},
				},
			},
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				// Allow traffic to database tier only
				To: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "database"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(5432))},
				},
			},
			{
				// Allow DNS resolution
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolUDP), Port: ptr(intstr.FromInt(53))},
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(53))},
				},
			},
		},
	},
}

// =============================================================================
// Database Policy - Allow backend ingress only
// =============================================================================

// DatabasePolicy allows only backend traffic in, minimal egress.
var DatabasePolicy = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "database-policy",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "database",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"tier": "database"},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				// Only allow traffic from backend tier
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "backend"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(5432))},
				},
			},
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				// Allow DNS resolution only
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolUDP), Port: ptr(intstr.FromInt(53))},
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(53))},
				},
			},
		},
	},
}

// =============================================================================
// Cross-Namespace Policy - Allow monitoring access
// =============================================================================

// AllowMonitoring allows Prometheus scraping from monitoring namespace.
var AllowMonitoring = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-monitoring",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: networkingv1.NetworkPolicySpec{
		// Applies to all pods
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				// Allow Prometheus scraping from monitoring namespace
				From: []networkingv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "monitoring"},
						},
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "prometheus"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(9090))},
				},
			},
		},
	},
}

func main() {}

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }
