package registry

// init registers all built-in Kubernetes types when the package is imported.
func init() {
	DefaultRegistry.RegisterBuiltins()
}

// RegisterBuiltins registers all standard Kubernetes API types.
func (r *Registry) RegisterBuiltins() {
	// Core v1 (no group, just "v1")
	r.registerCoreV1Types()

	// Apps v1
	r.registerAppsV1Types()

	// Batch v1
	r.registerBatchV1Types()

	// Networking v1
	r.registerNetworkingV1Types()

	// RBAC v1
	r.registerRBACv1Types()

	// Storage v1
	r.registerStorageV1Types()

	// Policy v1
	r.registerPolicyV1Types()

	// Autoscaling v1
	r.registerAutoscalingV1Types()

	// Autoscaling v2
	r.registerAutoscalingV2Types()
}

func (r *Registry) registerCoreV1Types() {
	types := []TypeInfo{
		{Package: "corev1", Group: "", Version: "v1", Kind: "Pod", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Service", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "ConfigMap", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Secret", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Namespace", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "ServiceAccount", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "PersistentVolume", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "PersistentVolumeClaim", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Node", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Endpoints", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Event", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "LimitRange", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "ResourceQuota", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "ReplicationController", APIVersion: "v1"},
		// Supporting types (not top-level resources, but useful for discovery)
		{Package: "corev1", Group: "", Version: "v1", Kind: "PodTemplateSpec", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Container", APIVersion: "v1"},
		{Package: "corev1", Group: "", Version: "v1", Kind: "Volume", APIVersion: "v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerAppsV1Types() {
	types := []TypeInfo{
		{Package: "appsv1", Group: "apps", Version: "v1", Kind: "Deployment", APIVersion: "apps/v1"},
		{Package: "appsv1", Group: "apps", Version: "v1", Kind: "StatefulSet", APIVersion: "apps/v1"},
		{Package: "appsv1", Group: "apps", Version: "v1", Kind: "DaemonSet", APIVersion: "apps/v1"},
		{Package: "appsv1", Group: "apps", Version: "v1", Kind: "ReplicaSet", APIVersion: "apps/v1"},
		{Package: "appsv1", Group: "apps", Version: "v1", Kind: "ControllerRevision", APIVersion: "apps/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerBatchV1Types() {
	types := []TypeInfo{
		{Package: "batchv1", Group: "batch", Version: "v1", Kind: "Job", APIVersion: "batch/v1"},
		{Package: "batchv1", Group: "batch", Version: "v1", Kind: "CronJob", APIVersion: "batch/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerNetworkingV1Types() {
	types := []TypeInfo{
		{Package: "networkingv1", Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", APIVersion: "networking.k8s.io/v1"},
		{Package: "networkingv1", Group: "networking.k8s.io", Version: "v1", Kind: "IngressClass", APIVersion: "networking.k8s.io/v1"},
		{Package: "networkingv1", Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy", APIVersion: "networking.k8s.io/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerRBACv1Types() {
	types := []TypeInfo{
		{Package: "rbacv1", Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role", APIVersion: "rbac.authorization.k8s.io/v1"},
		{Package: "rbacv1", Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding", APIVersion: "rbac.authorization.k8s.io/v1"},
		{Package: "rbacv1", Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole", APIVersion: "rbac.authorization.k8s.io/v1"},
		{Package: "rbacv1", Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding", APIVersion: "rbac.authorization.k8s.io/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerStorageV1Types() {
	types := []TypeInfo{
		{Package: "storagev1", Group: "storage.k8s.io", Version: "v1", Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
		{Package: "storagev1", Group: "storage.k8s.io", Version: "v1", Kind: "VolumeAttachment", APIVersion: "storage.k8s.io/v1"},
		{Package: "storagev1", Group: "storage.k8s.io", Version: "v1", Kind: "CSIDriver", APIVersion: "storage.k8s.io/v1"},
		{Package: "storagev1", Group: "storage.k8s.io", Version: "v1", Kind: "CSINode", APIVersion: "storage.k8s.io/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerPolicyV1Types() {
	types := []TypeInfo{
		{Package: "policyv1", Group: "policy", Version: "v1", Kind: "PodDisruptionBudget", APIVersion: "policy/v1"},
		{Package: "policyv1", Group: "policy", Version: "v1", Kind: "PodSecurityPolicy", APIVersion: "policy/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerAutoscalingV1Types() {
	types := []TypeInfo{
		{Package: "autoscalingv1", Group: "autoscaling", Version: "v1", Kind: "HorizontalPodAutoscaler", APIVersion: "autoscaling/v1"},
		{Package: "autoscalingv1", Group: "autoscaling", Version: "v1", Kind: "Scale", APIVersion: "autoscaling/v1"},
	}
	r.RegisterBulk(types)
}

func (r *Registry) registerAutoscalingV2Types() {
	types := []TypeInfo{
		{Package: "autoscalingv2", Group: "autoscaling", Version: "v2", Kind: "HorizontalPodAutoscaler", APIVersion: "autoscaling/v2"},
	}
	r.RegisterBulk(types)
}

// RegisterCRDTypes registers types from parsed CRD definitions.
// This is called by the CRD codegen to register generated types.
func (r *Registry) RegisterCRDTypes(domain string, types []CRDTypeInfo) {
	for _, t := range types {
		info := TypeInfo{
			Package:    t.Package,
			Group:      t.Group,
			Version:    t.Version,
			Kind:       t.Kind,
			APIVersion: t.APIVersion,
			Domain:     domain,
		}
		r.Register(info)
	}
}

// CRDTypeInfo contains information needed to register a CRD type.
type CRDTypeInfo struct {
	Package    string // Go package alias (e.g., "computev1beta1")
	Group      string // Full API group (e.g., "compute.cnrm.cloud.google.com")
	Version    string // API version (e.g., "v1beta1")
	Kind       string // Resource kind (e.g., "ComputeInstance")
	APIVersion string // Full apiVersion (e.g., "compute.cnrm.cloud.google.com/v1beta1")
}
