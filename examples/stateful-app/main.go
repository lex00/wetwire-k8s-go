// Package main demonstrates StatefulSet patterns.
//
// This example shows:
// - StatefulSet with persistent storage
// - Headless Service for stable network identities
// - PodDisruptionBudget for high availability
// - Init containers for cluster bootstrap
//
// Use case: A PostgreSQL database cluster with replicas.
package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func main() {}

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Common Labels
// =============================================================================

var appLabels = map[string]string{
	"app.kubernetes.io/name":       "postgres",
	"app.kubernetes.io/component":  "database",
	"app.kubernetes.io/managed-by": "wetwire-k8s",
}

var selectorLabels = map[string]string{
	"app.kubernetes.io/name": "postgres",
}

// =============================================================================
// Headless Service - Required for StatefulSet
// =============================================================================

// PostgresHeadless provides stable DNS names for pods.
// Each pod gets: <pod-name>.<service-name>.<namespace>.svc.cluster.local
var PostgresHeadless = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "postgres-headless",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: corev1.ServiceSpec{
		// ClusterIP: None makes this a headless service
		ClusterIP: "None",
		Selector:  selectorLabels,
		Ports: []corev1.ServicePort{
			{
				Name:       "postgres",
				Port:       5432,
				TargetPort: intstr.FromInt(5432),
			},
		},
		// Publish not-ready addresses for peer discovery during bootstrap
		PublishNotReadyAddresses: true,
	},
}

// PostgresService provides a stable endpoint for client connections.
var PostgresService = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "postgres",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: selectorLabels,
		Ports: []corev1.ServicePort{
			{
				Name:       "postgres",
				Port:       5432,
				TargetPort: intstr.FromInt(5432),
			},
		},
	},
}

// =============================================================================
// StatefulSet
// =============================================================================

// PostgresStatefulSet manages a PostgreSQL cluster with persistent storage.
var PostgresStatefulSet = appsv1.StatefulSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "postgres",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: appsv1.StatefulSetSpec{
		// serviceName must match the headless service
		ServiceName: "postgres-headless",
		Replicas:    ptr(int32(3)),
		Selector: &metav1.LabelSelector{
			MatchLabels: selectorLabels,
		},
		// OrderedReady ensures pods are created in order (0, 1, 2)
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		// RollingUpdate with partition for canary deployments
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
				// Partition: pods with ordinal >= partition are updated
				Partition: ptr(int32(0)),
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: selectorLabels,
			},
			Spec: corev1.PodSpec{
				// Spread pods across nodes
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: selectorLabels,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				// Init container to configure replication
				InitContainers: []corev1.Container{
					{
						Name:  "init-postgres",
						Image: "postgres:16-alpine",
						Command: []string{
							"sh", "-c",
							`# Determine if this is a primary or replica based on ordinal
							ORDINAL=${HOSTNAME##*-}
							if [ "$ORDINAL" = "0" ]; then
								echo "Primary node, no additional setup needed"
							else
								echo "Replica node, will connect to primary"
							fi`,
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "data", MountPath: "/var/lib/postgresql/data"},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:  "postgres",
						Image: "postgres:16-alpine",
						Ports: []corev1.ContainerPort{
							{Name: "postgres", ContainerPort: 5432},
						},
						Env: []corev1.EnvVar{
							{Name: "POSTGRES_DB", Value: "app"},
							{
								Name: "POSTGRES_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "postgres-credentials",
										},
										Key: "password",
									},
								},
							},
							{Name: "PGDATA", Value: "/var/lib/postgresql/data/pgdata"},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("250m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "data", MountPath: "/var/lib/postgresql/data"},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"pg_isready", "-U", "postgres"},
								},
							},
							InitialDelaySeconds: 30,
							PeriodSeconds:       10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"pg_isready", "-U", "postgres"},
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},
					},
				},
				// Graceful shutdown
				TerminationGracePeriodSeconds: ptr(int64(30)),
			},
		},
		// VolumeClaimTemplates create PVCs for each pod
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					// Uncomment to specify a storage class
					// StorageClassName: ptr("standard"),
				},
			},
		},
	},
}

// =============================================================================
// Credentials Secret
// =============================================================================

// PostgresCredentials stores database credentials.
// In production, use external secrets management (Vault, AWS Secrets Manager, etc.)
var PostgresCredentials = corev1.Secret{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Secret",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "postgres-credentials",
		Namespace: "default",
		Labels:    appLabels,
	},
	Type: corev1.SecretTypeOpaque,
	StringData: map[string]string{
		// IMPORTANT: Change this in production!
		"password": "change-me-in-production",
	},
}

// =============================================================================
// PodDisruptionBudget - Ensure availability during disruptions
// =============================================================================

// PostgresPDB ensures at least 2 replicas are available during disruptions.
var PostgresPDB = policyv1.PodDisruptionBudget{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "policy/v1",
		Kind:       "PodDisruptionBudget",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "postgres-pdb",
		Namespace: "default",
		Labels:    appLabels,
	},
	Spec: policyv1.PodDisruptionBudgetSpec{
		// At least 2 pods must be available
		MinAvailable: ptr(intstr.FromInt(2)),
		Selector: &metav1.LabelSelector{
			MatchLabels: selectorLabels,
		},
	},
}
