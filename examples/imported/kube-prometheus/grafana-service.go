package prometheus

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var GrafanaService = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "grafana",
		Namespace: "monitoring",
		Labels: map[string]string{
			"app.kubernetes.io/component": "grafana",
			"app.kubernetes.io/name":      "grafana",
			"app.kubernetes.io/part-of":   "kube-prometheus",
			"app.kubernetes.io/version":   "12.3.1",
		},
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{
			"app.kubernetes.io/component": "grafana",
			"app.kubernetes.io/name":      "grafana",
			"app.kubernetes.io/part-of":   "kube-prometheus",
		},
		Ports: []corev1.ServicePort{
			{
				Name: "http",
				Port: 3000,
			},
		},
	},
}
