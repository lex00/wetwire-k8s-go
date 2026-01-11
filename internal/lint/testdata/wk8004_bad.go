package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8004: Circular dependency detection
// This file contains violations

// Bad: Circular dependency A -> B -> A
var PodA = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name: PodB.Metadata.Name, // References PodB
	},
}

var PodB = corev1.Pod{
	Metadata: corev1.ObjectMeta{
		Name: PodA.Metadata.Name, // References PodA
	},
}

// Bad: Circular dependency A -> B -> C -> A
var ConfigA = corev1.ConfigMap{
	Metadata: corev1.ObjectMeta{
		Name: ConfigC.Metadata.Name,
	},
}

var ConfigB = corev1.ConfigMap{
	Metadata: corev1.ObjectMeta{
		Name: ConfigA.Metadata.Name,
	},
}

var ConfigC = corev1.ConfigMap{
	Metadata: corev1.ObjectMeta{
		Name: ConfigB.Metadata.Name,
	},
}
