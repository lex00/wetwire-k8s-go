package testdata

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// WK8004: Circular dependency detection
// This file contains compliant code

// Good: Linear dependencies (no cycles)
var MyService = corev1.Service{
	Metadata: corev1.ObjectMeta{
		Name: "my-service",
	},
}

var MyPod = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name: "my-pod",
		Labels: map[string]string{
			"service": MyService.Metadata.Name, // Pod -> Service (no cycle)
		},
	},
}

var MyIngress = networkingv1.Ingress{
	Spec: networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: MyService.Metadata.Name, // Ingress -> Service (no cycle)
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
