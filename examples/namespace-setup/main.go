// Package main demonstrates a production-ready namespace setup.
//
// This example shows:
// - Namespace with labels and annotations
// - ResourceQuota for resource governance
// - LimitRange for default limits
// - NetworkPolicy for isolation
// - RBAC for access control
//
// Use case: Multi-tenant cluster where each team gets an isolated namespace
// with resource limits and network isolation.
package main

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// Constants
// =============================================================================

const namespaceName = "team-alpha"

var teamLabels = map[string]string{
	"team":                          "alpha",
	"app.kubernetes.io/managed-by": "wetwire-k8s",
}

// =============================================================================
// Namespace
// =============================================================================

// TeamNamespace creates an isolated namespace for the team.
var TeamNamespace = corev1.Namespace{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Namespace",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: namespaceName,
		Labels: map[string]string{
			"name":                          namespaceName,
			"team":                          "alpha",
			"environment":                   "production",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
		Annotations: map[string]string{
			"owner":       "team-alpha@example.com",
			"cost-center": "engineering",
		},
	},
}

// =============================================================================
// ResourceQuota - Limit total resources in namespace
// =============================================================================

// ComputeQuota limits total resources the namespace can use.
var ComputeQuota = corev1.ResourceQuota{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ResourceQuota",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "compute-quota",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Spec: corev1.ResourceQuotaSpec{
		Hard: corev1.ResourceList{
			// CPU limits
			corev1.ResourceRequestsCPU: resource.MustParse("10"),
			corev1.ResourceLimitsCPU:   resource.MustParse("20"),
			// Memory limits
			corev1.ResourceRequestsMemory: resource.MustParse("20Gi"),
			corev1.ResourceLimitsMemory:   resource.MustParse("40Gi"),
			// Object count limits
			corev1.ResourcePods:                   resource.MustParse("50"),
			corev1.ResourceServices:               resource.MustParse("20"),
			corev1.ResourceSecrets:                resource.MustParse("100"),
			corev1.ResourceConfigMaps:             resource.MustParse("100"),
			corev1.ResourcePersistentVolumeClaims: resource.MustParse("10"),
			// Storage limits
			corev1.ResourceRequestsStorage: resource.MustParse("100Gi"),
		},
	},
}

// =============================================================================
// LimitRange - Default and max limits for pods
// =============================================================================

// DefaultLimits sets default resource limits for containers.
var DefaultLimits = corev1.LimitRange{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "LimitRange",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default-limits",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Spec: corev1.LimitRangeSpec{
		Limits: []corev1.LimitRangeItem{
			{
				Type: corev1.LimitTypeContainer,
				// Default values if not specified
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
				// Default requests if not specified
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				// Maximum allowed
				Max: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
				// Minimum required
				Min: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			},
			{
				Type: corev1.LimitTypePersistentVolumeClaim,
				Max: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("20Gi"),
				},
				Min: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	},
}

// =============================================================================
// NetworkPolicies
// =============================================================================

// DefaultDenyAll denies all traffic by default.
var DefaultDenyAll = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default-deny-all",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
	},
}

// AllowDNS allows DNS resolution to kube-dns.
var AllowDNS = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-dns",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				To: []networkingv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{},
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"k8s-app": "kube-dns"},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: ptr(corev1.ProtocolUDP), Port: ptr(intstr.FromInt(53))},
					{Protocol: ptr(corev1.ProtocolTCP), Port: ptr(intstr.FromInt(53))},
				},
			},
		},
	},
}

// AllowSameNamespace allows pods in the namespace to communicate.
var AllowSameNamespace = networkingv1.NetworkPolicy{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-same-namespace",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{From: []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}},
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{To: []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}},
		},
	},
}

// =============================================================================
// RBAC
// =============================================================================

// TeamServiceAccount provides identity for team workloads.
var TeamServiceAccount = corev1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ServiceAccount",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "team-workload",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
}

// DeveloperRole grants developers full access to workload resources.
var DeveloperRole = rbacv1.Role{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "Role",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "namespace-developer",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Rules: []rbacv1.PolicyRule{
		{
			// Full access to workload resources
			APIGroups: []string{"", "apps", "batch"},
			Resources: []string{
				"pods", "pods/log", "pods/exec",
				"deployments", "replicasets", "statefulsets", "daemonsets",
				"jobs", "cronjobs",
				"services", "configmaps", "secrets", "persistentvolumeclaims",
			},
			Verbs: []string{"*"},
		},
		{
			// Read-only for HPA and events
			APIGroups: []string{"autoscaling", ""},
			Resources: []string{"horizontalpodautoscalers", "events"},
			Verbs:     []string{"get", "list", "watch"},
		},
	},
}

// DeveloperBinding binds the developer role to team members.
var DeveloperBinding = rbacv1.RoleBinding{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "RoleBinding",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "team-developers",
		Namespace: namespaceName,
		Labels:    teamLabels,
	},
	Subjects: []rbacv1.Subject{
		{
			// Bind to a group (configure in your identity provider)
			Kind:     "Group",
			Name:     "team-alpha-developers",
			APIGroup: "rbac.authorization.k8s.io",
		},
	},
	RoleRef: rbacv1.RoleRef{
		Kind:     "Role",
		Name:     "namespace-developer",
		APIGroup: "rbac.authorization.k8s.io",
	},
}

func main() {}

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }
