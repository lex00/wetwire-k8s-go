package prometheus

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var PrometheusK8sServiceAccount = corev1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ServiceAccount",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "prometheus-k8s",
		Namespace: "monitoring",
		Labels: map[string]string{
			"app.kubernetes.io/component": "prometheus",
			"app.kubernetes.io/instance":  "k8s",
			"app.kubernetes.io/name":      "prometheus",
			"app.kubernetes.io/part-of":   "kube-prometheus",
			"app.kubernetes.io/version":   "3.9.1",
		},
	},
}
