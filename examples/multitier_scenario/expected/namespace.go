// Package expected provides gold standard Kubernetes resources for the multi-tier scenario.
package expected

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// EcommerceNamespace defines the namespace for the e-commerce application.
var EcommerceNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ecommerce",
		Labels: map[string]string{
			"app":  "ecommerce",
			"tier": "production",
		},
	},
}

// EcommerceQuota defines resource quotas for the e-commerce namespace.
// Limits: 20 pods, 10 CPU cores, 20Gi memory
var EcommerceQuota = &corev1.ResourceQuota{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ecommerce-quota",
		Namespace: "ecommerce",
	},
	Spec: corev1.ResourceQuotaSpec{
		Hard: corev1.ResourceList{
			corev1.ResourcePods:   resource.MustParse("20"),
			corev1.ResourceCPU:    resource.MustParse("10"),
			corev1.ResourceMemory: resource.MustParse("20Gi"),
		},
	},
}
