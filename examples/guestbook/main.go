// Package main demonstrates a multi-tier guestbook application using wetwire pattern.
// This example shows a Redis backend with a web frontend, following Kubernetes
// guestbook tutorial patterns.
package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Helper function for pointer values
func ptr[T any](v T) *T { return &v }

// =============================================================================
// Shared Labels
// =============================================================================

// redisLeaderLabels identifies the Redis leader (master) pods
var redisLeaderLabels = map[string]string{
	"app":  "redis",
	"role": "leader",
	"tier": "backend",
}

// redisFollowerLabels identifies the Redis follower (replica) pods
var redisFollowerLabels = map[string]string{
	"app":  "redis",
	"role": "follower",
	"tier": "backend",
}

// frontendLabels identifies the guestbook frontend pods
var frontendLabels = map[string]string{
	"app":  "guestbook",
	"tier": "frontend",
}

// =============================================================================
// Redis Leader (Master)
// =============================================================================

// RedisLeaderDeployment runs the Redis leader instance
var RedisLeaderDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "redis-leader",
		Labels: redisLeaderLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(1)),
		Selector: &metav1.LabelSelector{
			MatchLabels: redisLeaderLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: redisLeaderLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "leader",
						Image: "docker.io/redis:7.0",
						Ports: []corev1.ContainerPort{
							{
								Name:          "redis",
								ContainerPort: 6379,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// RedisLeaderService exposes the Redis leader to other pods
var RedisLeaderService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "redis-leader",
		Labels: redisLeaderLabels,
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Port:       6379,
				TargetPort: intstr.FromInt(6379),
			},
		},
		Selector: redisLeaderLabels,
	},
}

// =============================================================================
// Redis Followers (Replicas)
// =============================================================================

// RedisFollowerDeployment runs Redis follower instances that replicate from leader
var RedisFollowerDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "redis-follower",
		Labels: redisFollowerLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(2)),
		Selector: &metav1.LabelSelector{
			MatchLabels: redisFollowerLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: redisFollowerLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "follower",
						Image: "us-docker.pkg.dev/google-samples/containers/gke/gb-redis-follower:v2",
						Ports: []corev1.ContainerPort{
							{
								Name:          "redis",
								ContainerPort: 6379,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// RedisFollowerService exposes the Redis followers for read operations
var RedisFollowerService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "redis-follower",
		Labels: redisFollowerLabels,
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Port:       6379,
				TargetPort: intstr.FromInt(6379),
			},
		},
		Selector: redisFollowerLabels,
	},
}

// =============================================================================
// Guestbook Frontend
// =============================================================================

// FrontendDeployment runs the guestbook web application
var FrontendDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "frontend",
		Labels: frontendLabels,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptr(int32(3)),
		Selector: &metav1.LabelSelector{
			MatchLabels: frontendLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: frontendLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "php-redis",
						Image: "us-docker.pkg.dev/google-samples/containers/gke/gb-frontend:v5",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "GET_HOSTS_FROM",
								Value: "dns",
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
	},
}

// FrontendService exposes the frontend to external traffic
var FrontendService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "frontend",
		Labels: frontendLabels,
	},
	Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeLoadBalancer,
		Ports: []corev1.ServicePort{
			{
				Port:       80,
				TargetPort: intstr.FromInt(80),
			},
		},
		Selector: frontendLabels,
	},
}

// main is required for package main but not used - wetwire-k8s discovers resources from variable declarations
func main() {}
