package prometheus

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var GrafanaDeployment = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
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
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr.To[int32](1),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/component": "grafana",
				"app.kubernetes.io/name":      "grafana",
				"app.kubernetes.io/part-of":   "kube-prometheus",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/component": "grafana",
					"app.kubernetes.io/name":      "grafana",
					"app.kubernetes.io/part-of":   "kube-prometheus",
					"app.kubernetes.io/version":   "12.3.1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "grafana",
						Image: "grafana/grafana:12.3.1",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 3000,
							},
						},
					},
				},
			},
		},
	},
}
