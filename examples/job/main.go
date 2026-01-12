// Package main demonstrates Kubernetes Job patterns.
//
// This example shows:
// - Simple Job: One-time execution (e.g., database migration)
// - Parallel Job: Process multiple items concurrently
// - Indexed Job: Each pod gets a unique index
// - Job with Init Container: Setup before processing
//
// Use cases: Database migrations, batch processing, report generation
package main

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {}

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Simple Job - One-time execution
// =============================================================================

// DatabaseMigrationJob runs database migrations once.
var DatabaseMigrationJob = batchv1.Job{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "batch/v1",
		Kind:       "Job",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "database-migration",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "database-migration",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: batchv1.JobSpec{
		BackoffLimit:            ptr(int32(3)),
		ActiveDeadlineSeconds:   ptr(int64(600)),
		TTLSecondsAfterFinished: ptr(int32(3600)),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "database-migration",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:    "migrate",
						Image:   "myapp/migrations:v1.2.0",
						Command: []string{"python", "manage.py", "migrate"},
						Env: []corev1.EnvVar{
							{
								Name: "DATABASE_URL",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "database-credentials",
										},
										Key: "url",
									},
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// =============================================================================
// Parallel Job - Process multiple items
// =============================================================================

// BatchProcessorJob processes work items in parallel.
var BatchProcessorJob = batchv1.Job{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "batch/v1",
		Kind:       "Job",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "batch-processor",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "batch-processor",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: batchv1.JobSpec{
		Completions:  ptr(int32(10)),
		Parallelism:  ptr(int32(3)),
		BackoffLimit: ptr(int32(5)),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "batch-processor",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:    "processor",
						Image:   "myapp/processor:v1.0.0",
						Command: []string{"python", "process_batch.py"},
						Env: []corev1.EnvVar{
							{
								Name:  "QUEUE_URL",
								Value: "redis://redis:6379/0",
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	},
}

// =============================================================================
// Indexed Job - Each pod gets a unique index
// =============================================================================

// ReportGeneratorJob generates reports with indexed completion mode.
var ReportGeneratorJob = batchv1.Job{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "batch/v1",
		Kind:       "Job",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "report-generator",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "report-generator",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: batchv1.JobSpec{
		CompletionMode: ptr(batchv1.IndexedCompletion),
		Completions:    ptr(int32(5)),
		Parallelism:    ptr(int32(5)),
		BackoffLimit:   ptr(int32(2)),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "report-generator",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:  "generator",
						Image: "myapp/report-generator:v2.0.0",
						// JOB_COMPLETION_INDEX is automatically set (0, 1, 2, ...)
						Command: []string{"sh", "-c", "python generate_report.py --region=$JOB_COMPLETION_INDEX"},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "output",
								MountPath: "/output",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "output",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "reports-pvc",
							},
						},
					},
				},
			},
		},
	},
}

// =============================================================================
// Job with Init Container - Setup before processing
// =============================================================================

// DataImportJob downloads data before processing.
var DataImportJob = batchv1.Job{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "batch/v1",
		Kind:       "Job",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "data-import",
		Namespace: "default",
		Labels: map[string]string{
			"app.kubernetes.io/name":       "data-import",
			"app.kubernetes.io/managed-by": "wetwire-k8s",
		},
	},
	Spec: batchv1.JobSpec{
		BackoffLimit:          ptr(int32(2)),
		ActiveDeadlineSeconds: ptr(int64(3600)),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name": "data-import",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				InitContainers: []corev1.Container{
					{
						Name:    "download",
						Image:   "curlimages/curl:8.5.0",
						Command: []string{"sh", "-c", "curl -o /data/input.csv https://example.com/data.csv"},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "data", MountPath: "/data"},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "import",
						Image:   "myapp/importer:v1.0.0",
						Command: []string{"python", "import.py", "/data/input.csv"},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "data", MountPath: "/data"},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
			},
		},
	},
}
