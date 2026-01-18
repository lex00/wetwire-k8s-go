// Package expected provides gold standard Kubernetes resources for the multi-tier scenario.
package expected

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppConfig stores application configuration data.
// Contains database connection parameters used by the backend.
var AppConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "app-config",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app": "ecommerce",
		},
	},
	Data: map[string]string{
		"database.host": "postgres",
		"database.port": "5432",
		"database.name": "ecommerce",
	},
}

// DBCredentials stores sensitive database credentials.
// Referenced by backend deployment for database authentication.
var DBCredentials = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "db-credentials",
		Namespace: "ecommerce",
		Labels: map[string]string{
			"app": "ecommerce",
		},
	},
	Type: corev1.SecretTypeOpaque,
	StringData: map[string]string{
		"username": "postgres",
		"password": "changeme",
	},
}
