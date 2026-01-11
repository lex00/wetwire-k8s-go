package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMap that other resources depend on
var AppConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "app-config",
	},
}

// Deployment that references AppConfig
var WebDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "web-deployment",
	},
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: AppConfig.Name,
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

// Service that references WebDeployment
var WebService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "web-service",
	},
	Spec: corev1.ServiceSpec{
		Selector: WebDeployment.Spec.Template.Labels,
	},
}
