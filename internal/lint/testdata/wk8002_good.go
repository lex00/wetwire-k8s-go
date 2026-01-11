package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8002: Avoid deeply nested inline structures (max depth 5)
// This file contains compliant code

// Good: Extracted to variables
var AppEnvVars = []corev1.EnvVar{
	{
		Name:  "CONFIG",
		Value: "value",
	},
}

var AppContainer = corev1.Container{
	Name:  "app",
	Image: "nginx:latest",
	Env:   AppEnvVars,
}

var AppPodSpec = corev1.PodSpec{
	Containers: []corev1.Container{AppContainer},
}

var AppPodTemplate = corev1.PodTemplateSpec{
	Spec: AppPodSpec,
}

var FlatDeployment = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: AppPodTemplate,
	},
}
