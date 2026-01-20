package registry

import (
	"testing"
)

func TestDefaultRegistryHasBuiltins(t *testing.T) {
	// DefaultRegistry should have builtins registered via init()
	tests := []struct {
		typeName string
		expected bool
	}{
		// Core v1 types
		{"corev1.Pod", true},
		{"corev1.Service", true},
		{"corev1.ConfigMap", true},
		{"corev1.Secret", true},
		{"corev1.Namespace", true},

		// Apps v1 types
		{"appsv1.Deployment", true},
		{"appsv1.StatefulSet", true},
		{"appsv1.DaemonSet", true},

		// Batch v1 types
		{"batchv1.Job", true},
		{"batchv1.CronJob", true},

		// Networking v1 types
		{"networkingv1.Ingress", true},
		{"networkingv1.NetworkPolicy", true},

		// RBAC v1 types
		{"rbacv1.Role", true},
		{"rbacv1.ClusterRole", true},

		// Unknown types
		{"unknown.Type", false},
		{"foov1.Bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := DefaultRegistry.IsKnownType(tt.typeName)
			if result != tt.expected {
				t.Errorf("IsKnownType(%q) = %v, want %v", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	r := NewRegistry()

	// Register a custom type
	r.Register(TypeInfo{
		Package:    "computev1beta1",
		Group:      "compute.cnrm.cloud.google.com",
		Version:    "v1beta1",
		Kind:       "ComputeInstance",
		APIVersion: "compute.cnrm.cloud.google.com/v1beta1",
		Domain:     "cnrm",
	})

	// Should be found by qualified name
	if !r.IsKnownType("computev1beta1.ComputeInstance") {
		t.Error("Expected registered type to be found by qualified name")
	}

	// Should be found by kind (backward compatibility)
	if !r.IsKnownType("ComputeInstance") {
		t.Error("Expected registered type to be found by kind")
	}

	// Package should be known
	if !r.IsKnownPackage("computev1beta1") {
		t.Error("Expected package to be known")
	}

	// Get type info
	info, ok := r.GetTypeInfo("computev1beta1.ComputeInstance")
	if !ok {
		t.Fatal("Expected to get type info")
	}

	if info.Kind != "ComputeInstance" {
		t.Errorf("Kind = %q, want %q", info.Kind, "ComputeInstance")
	}

	if info.Domain != "cnrm" {
		t.Errorf("Domain = %q, want %q", info.Domain, "cnrm")
	}
}

func TestAPIVersionForPackage(t *testing.T) {
	tests := []struct {
		pkg      string
		expected string
	}{
		{"corev1", "v1"},
		{"appsv1", "apps/v1"},
		{"batchv1", "batch/v1"},
		{"networkingv1", "networking.k8s.io/v1"},
		{"rbacv1", "rbac.authorization.k8s.io/v1"},
		{"autoscalingv2", "autoscaling/v2"},
		{"unknownpkg", ""},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			result := DefaultRegistry.APIVersionForPackage(tt.pkg)
			if result != tt.expected {
				t.Errorf("APIVersionForPackage(%q) = %q, want %q", tt.pkg, result, tt.expected)
			}
		})
	}
}

func TestIsKnownPackage(t *testing.T) {
	tests := []struct {
		pkg      string
		expected bool
	}{
		{"corev1", true},
		{"appsv1", true},
		{"batchv1", true},
		{"networkingv1", true},
		{"rbacv1", true},
		{"storagev1", true},
		{"policyv1", true},
		{"autoscalingv1", true},
		{"autoscalingv2", true},
		{"unknownpkg", false},
		{"foov1", false},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			result := DefaultRegistry.IsKnownPackage(tt.pkg)
			if result != tt.expected {
				t.Errorf("IsKnownPackage(%q) = %v, want %v", tt.pkg, result, tt.expected)
			}
		})
	}
}

func TestRegistryClear(t *testing.T) {
	r := NewRegistry()
	r.Register(TypeInfo{
		Package:    "testpkg",
		Kind:       "TestKind",
		APIVersion: "test/v1",
	})

	if !r.IsKnownType("testpkg.TestKind") {
		t.Error("Expected type to be registered")
	}

	r.Clear()

	if r.IsKnownType("testpkg.TestKind") {
		t.Error("Expected type to be cleared")
	}
}

func TestListPackages(t *testing.T) {
	pkgs := DefaultRegistry.ListPackages()

	// Should have at least the standard K8s packages
	expectedPkgs := []string{"corev1", "appsv1", "batchv1"}

	for _, expected := range expectedPkgs {
		found := false
		for _, pkg := range pkgs {
			if pkg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected package %q not found in list", expected)
		}
	}
}

func TestRegisterCRDTypes(t *testing.T) {
	r := NewRegistry()

	crdTypes := []CRDTypeInfo{
		{
			Package:    "computev1beta1",
			Group:      "compute.cnrm.cloud.google.com",
			Version:    "v1beta1",
			Kind:       "ComputeInstance",
			APIVersion: "compute.cnrm.cloud.google.com/v1beta1",
		},
		{
			Package:    "storagev1beta1",
			Group:      "storage.cnrm.cloud.google.com",
			Version:    "v1beta1",
			Kind:       "StorageBucket",
			APIVersion: "storage.cnrm.cloud.google.com/v1beta1",
		},
	}

	r.RegisterCRDTypes("cnrm", crdTypes)

	// Both types should be registered
	if !r.IsKnownType("computev1beta1.ComputeInstance") {
		t.Error("Expected ComputeInstance to be registered")
	}

	if !r.IsKnownType("storagev1beta1.StorageBucket") {
		t.Error("Expected StorageBucket to be registered")
	}

	// Both should have domain set
	info, _ := r.GetTypeInfo("computev1beta1.ComputeInstance")
	if info.Domain != "cnrm" {
		t.Errorf("Domain = %q, want %q", info.Domain, "cnrm")
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	// Run concurrent registrations and lookups
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			r.Register(TypeInfo{
				Package: "testpkg",
				Kind:    "TestKind",
			})
			r.IsKnownType("testpkg.TestKind")
			r.ListPackages()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
