package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8041: Hardcoded API keys/tokens
// This file contains violations

// Bad: Bearer token in env var
var PodWithBearerToken = corev1.Pod{
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "app",
				Image: "myapp",
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "AUTH_HEADER",
						Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // Bearer token
					},
				},
			},
		},
	},
}

// Bad: API key pattern in config
var ConfigMapWithAPIKey = corev1.ConfigMap{
	Data: map[string]string{
		"config.yaml": "api_key=sk_live_1234567890abcdef", // API key pattern
	},
}

// Bad: Token pattern in secret (ironically)
var ConfigWithToken = corev1.ConfigMap{
	Data: map[string]string{
		"app.conf": "token: ghp_1234567890abcdefghijklmnopqrstuv", // GitHub token pattern
	},
}

// Bad: AWS access key pattern
var ConfigWithAWSKey = corev1.ConfigMap{
	Data: map[string]string{
		"aws.conf": "aws_access_key_id=AKIAIOSFODNN7EXAMPLE", // AWS access key
	},
}
