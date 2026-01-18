package argo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var ArgocdServerService = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "argocd-server",
		Labels: map[string]string{
			"app.kubernetes.io/component": "server",
			"app.kubernetes.io/name":      "argocd-server",
			"app.kubernetes.io/part-of":   "argocd",
		},
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{
			"app.kubernetes.io/name": "argocd-server",
		},
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt32(8080),
			},
			{
				Name:       "https",
				Port:       443,
				TargetPort: intstr.FromInt32(8080),
			},
		},
	},
}
