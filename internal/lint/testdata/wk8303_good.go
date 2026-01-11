package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// WK8303: PodDisruptionBudget
// This file contains no violations - HA deployment with matching PDB

// Helper function for int32 pointer
func ptrInt32_8303Good(i int32) *int32 {
	return &i
}

// Good: HA deployment with matching PDB
var HADeploymentWithPDB = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-with-pdb",
		Labels: map[string]string{
			"app": "ha-app-with-pdb",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8303Good(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "ha-app-with-pdb",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "ha-app-with-pdb",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:1.21",
					},
				},
			},
		},
	},
}

// PDB for the HA deployment
var HAAppPDB = policyv1.PodDisruptionBudget{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ha-app-with-pdb-pdb",
	},
	Spec: policyv1.PodDisruptionBudgetSpec{
		MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 2},
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "ha-app-with-pdb",
			},
		},
	},
}

// Good: Single replica deployment (no PDB needed)
var SingleReplicaDeployment = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "single-app",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrInt32_8303Good(1),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "nginx:1.21",
					},
				},
			},
		},
	},
}
