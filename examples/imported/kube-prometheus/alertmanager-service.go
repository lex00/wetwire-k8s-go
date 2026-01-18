package prometheus

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AlertmanagerMainService = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "alertmanager-main",
		Namespace: "monitoring",
		Labels: map[string]string{
			"app.kubernetes.io/component": "alert-router",
			"app.kubernetes.io/instance":  "main",
			"app.kubernetes.io/name":      "alertmanager",
			"app.kubernetes.io/part-of":   "kube-prometheus",
			"app.kubernetes.io/version":   "0.30.0",
		},
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{
			"app.kubernetes.io/component": "alert-router",
			"app.kubernetes.io/instance":  "main",
			"app.kubernetes.io/name":      "alertmanager",
			"app.kubernetes.io/part-of":   "kube-prometheus",
		},
		Ports: []corev1.ServicePort{
			{
				Name: "web",
				Port: 9093,
			},
			{
				Name: "reloader-web",
				Port: 8080,
			},
		},
	},
}
