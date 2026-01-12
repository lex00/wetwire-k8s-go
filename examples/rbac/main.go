// Package main demonstrates RBAC (Role-Based Access Control) patterns.
//
// This example shows:
// - ServiceAccount: Identity for pods
// - Role: Namespace-scoped permissions
// - ClusterRole: Cluster-wide permissions
// - RoleBinding: Binds Role to subjects
// - ClusterRoleBinding: Binds ClusterRole to subjects
//
// Use case: Application that needs to read ConfigMaps and Secrets
// in its namespace, plus cluster-wide read access to Nodes.
package main

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// Common Labels
// =============================================================================

func main() {}

var commonLabels = map[string]string{
	"app.kubernetes.io/name":       "my-app",
	"app.kubernetes.io/managed-by": "wetwire-k8s",
}

// =============================================================================
// ServiceAccount
// =============================================================================

// AppServiceAccount provides identity for application pods.
var AppServiceAccount = corev1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ServiceAccount",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "app-service-account",
		Namespace: "default",
		Labels:    commonLabels,
	},
}

// =============================================================================
// Role - Namespace-scoped permissions
// =============================================================================

// AppRole grants namespace-scoped permissions for reading configs and secrets.
var AppRole = rbacv1.Role{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "Role",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "app-role",
		Namespace: "default",
		Labels:    commonLabels,
	},
	Rules: []rbacv1.PolicyRule{
		{
			// Read ConfigMaps
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			// Read Secrets
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			// Manage own pods (for sidecars, etc.)
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			// Read and update own service
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list", "watch", "update", "patch"},
		},
	},
}

// =============================================================================
// RoleBinding - Binds Role to ServiceAccount
// =============================================================================

// AppRoleBinding binds the AppRole to the AppServiceAccount.
var AppRoleBinding = rbacv1.RoleBinding{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "RoleBinding",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "app-role-binding",
		Namespace: "default",
		Labels:    commonLabels,
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "app-service-account",
			Namespace: "default",
		},
	},
	RoleRef: rbacv1.RoleRef{
		Kind:     "Role",
		Name:     "app-role",
		APIGroup: "rbac.authorization.k8s.io",
	},
}

// =============================================================================
// ClusterRole - Cluster-wide permissions
// =============================================================================

// NodeReaderClusterRole grants cluster-wide read access to nodes and namespaces.
var NodeReaderClusterRole = rbacv1.ClusterRole{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "ClusterRole",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:   "node-reader",
		Labels: commonLabels,
	},
	Rules: []rbacv1.PolicyRule{
		{
			// Read nodes (for scheduling decisions, node info)
			APIGroups: []string{""},
			Resources: []string{"nodes"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			// Read namespaces (for multi-tenant awareness)
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
			Verbs:     []string{"get", "list", "watch"},
		},
	},
}

// =============================================================================
// ClusterRoleBinding - Binds ClusterRole to ServiceAccount
// =============================================================================

// NodeReaderBinding binds the NodeReaderClusterRole to the AppServiceAccount.
var NodeReaderBinding = rbacv1.ClusterRoleBinding{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "ClusterRoleBinding",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:   "app-node-reader-binding",
		Labels: commonLabels,
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "app-service-account",
			Namespace: "default",
		},
	},
	RoleRef: rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     "node-reader",
		APIGroup: "rbac.authorization.k8s.io",
	},
}
