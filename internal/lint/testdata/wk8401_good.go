package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8401: File size limits
// This file is under the resource limit (<= 20 resources)

var MyDeployment = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{},
}

var MyService = corev1.Service{
	Spec: corev1.ServiceSpec{},
}

var MyConfigMap = corev1.ConfigMap{
	Data: map[string]string{"key": "value"},
}
