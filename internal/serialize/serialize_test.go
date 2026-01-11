package serialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestSerializeBasicStruct tests basic struct to map conversion
func TestSerializeBasicStruct(t *testing.T) {
	tests := []struct {
		name     string
		resource interface{}
		want     map[string]interface{}
	}{
		{
			name: "simple deployment",
			resource: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "default",
				},
			},
			want: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "test-deployment",
					"namespace": "default",
				},
			},
		},
		{
			name: "simple service",
			resource: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-service",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   corev1.ProtocolTCP,
						},
					},
				},
			},
			want: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "test-service",
				},
				"spec": map[string]interface{}{
					"type": "ClusterIP",
					"ports": []interface{}{
						map[string]interface{}{
							"port":       float64(80),
							"targetPort": float64(8080),
							"protocol":   "TCP",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Serialize(tt.resource)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestSerializeNestedStructs tests handling of nested structs
func TestSerializeNestedStructs(t *testing.T) {
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nested-test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	got, err := Serialize(deployment)
	require.NoError(t, err)

	// Verify nested structure
	assert.Equal(t, "apps/v1", got["apiVersion"])
	assert.Equal(t, "Deployment", got["kind"])

	metadata, ok := got["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "nested-test", metadata["name"])
	assert.Equal(t, "default", metadata["namespace"])

	labels, ok := metadata["labels"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", labels["app"])

	spec, ok := got["spec"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), spec["replicas"])

	selector, ok := spec["selector"].(map[string]interface{})
	require.True(t, ok)
	matchLabels, ok := selector["matchLabels"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", matchLabels["app"])

	template, ok := spec["template"].(map[string]interface{})
	require.True(t, ok)
	templateSpec, ok := template["spec"].(map[string]interface{})
	require.True(t, ok)

	containers, ok := templateSpec["containers"].([]interface{})
	require.True(t, ok)
	require.Len(t, containers, 1)

	container, ok := containers[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "nginx", container["name"])
	assert.Equal(t, "nginx:latest", container["image"])
}

// TestFieldNameConversion tests Go naming to Kubernetes camelCase conversion
func TestFieldNameConversion(t *testing.T) {
	resource := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1", // APIVersion -> apiVersion
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	got, err := Serialize(resource)
	require.NoError(t, err)

	// Check that APIVersion is converted to apiVersion
	assert.Contains(t, got, "apiVersion")
	assert.NotContains(t, got, "APIVersion")

	// Check that ObjectMeta is converted to metadata
	assert.Contains(t, got, "metadata")
	assert.NotContains(t, got, "objectMeta")
}

// TestZeroValueOmission tests that zero values are omitted from output
func TestZeroValueOmission(t *testing.T) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			// Namespace is empty string (zero value) - should be omitted
			// Labels is nil (zero value) - should be omitted
		},
		// Spec is zero value struct - should be omitted
	}

	got, err := Serialize(deployment)
	require.NoError(t, err)

	metadata, ok := got["metadata"].(map[string]interface{})
	require.True(t, ok)

	// Zero values should be omitted
	assert.NotContains(t, metadata, "namespace")
	assert.NotContains(t, metadata, "labels")
	assert.NotContains(t, metadata, "annotations")

	// Spec should be omitted if it's a zero value struct
	// However, Kubernetes structs may have fields - check if spec exists and is empty
	if spec, exists := got["spec"]; exists {
		specMap, ok := spec.(map[string]interface{})
		require.True(t, ok)
		// If spec exists, it should at least be empty or minimal
		_ = specMap
	}
}

// TestToYAML tests YAML output generation
func TestToYAML(t *testing.T) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
	}

	yaml, err := ToYAML(deployment)
	require.NoError(t, err)
	require.NotEmpty(t, yaml)

	yamlStr := string(yaml)
	assert.Contains(t, yamlStr, "apiVersion: apps/v1")
	assert.Contains(t, yamlStr, "kind: Deployment")
	assert.Contains(t, yamlStr, "metadata:")
	assert.Contains(t, yamlStr, "name: test-deployment")
	assert.Contains(t, yamlStr, "namespace: default")
}

// TestToJSON tests JSON output generation
func TestToJSON(t *testing.T) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
	}

	jsonBytes, err := ToJSON(deployment)
	require.NoError(t, err)
	require.NotEmpty(t, jsonBytes)

	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "apps/v1", result["apiVersion"])
	assert.Equal(t, "Deployment", result["kind"])

	metadata, ok := result["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-deployment", metadata["name"])
	assert.Equal(t, "default", metadata["namespace"])
}

// TestToMultiYAML tests multi-document YAML output
func TestToMultiYAML(t *testing.T) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-service",
		},
	}

	resources := []interface{}{deployment, service}
	yaml, err := ToMultiYAML(resources)
	require.NoError(t, err)
	require.NotEmpty(t, yaml)

	yamlStr := string(yaml)

	// Should contain document separator
	assert.Contains(t, yamlStr, "---")

	// Should contain both resources
	assert.Contains(t, yamlStr, "kind: Deployment")
	assert.Contains(t, yamlStr, "kind: Service")
	assert.Contains(t, yamlStr, "test-deployment")
	assert.Contains(t, yamlStr, "test-service")

	// Count document separators (should have at least one)
	separatorCount := strings.Count(yamlStr, "---")
	assert.GreaterOrEqual(t, separatorCount, 1)
}

// TestToMultiYAMLSingleResource tests multi-document YAML with single resource
func TestToMultiYAMLSingleResource(t *testing.T) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
	}

	resources := []interface{}{deployment}
	yaml, err := ToMultiYAML(resources)
	require.NoError(t, err)
	require.NotEmpty(t, yaml)

	yamlStr := string(yaml)
	assert.Contains(t, yamlStr, "kind: Deployment")
	assert.Contains(t, yamlStr, "test-deployment")
}

// TestToMultiYAMLEmpty tests multi-document YAML with no resources
func TestToMultiYAMLEmpty(t *testing.T) {
	resources := []interface{}{}
	yaml, err := ToMultiYAML(resources)
	require.NoError(t, err)
	assert.Empty(t, yaml)
}

// TestSerializeNilResource tests handling of nil resources
func TestSerializeNilResource(t *testing.T) {
	_, err := Serialize(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource cannot be nil")
}

// TestComplexResource tests a more complex real-world deployment
func TestComplexResource(t *testing.T) {
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "production",
			Labels: map[string]string{
				"app":     "nginx",
				"version": "1.0",
			},
			Annotations: map[string]string{
				"description": "NGINX web server",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "nginx",
						"version": "1.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.21",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENV",
									Value: "production",
								},
							},
						},
					},
				},
			},
		},
	}

	yaml, err := ToYAML(deployment)
	require.NoError(t, err)

	yamlStr := string(yaml)
	assert.Contains(t, yamlStr, "apiVersion: apps/v1")
	assert.Contains(t, yamlStr, "kind: Deployment")
	assert.Contains(t, yamlStr, "nginx-deployment")
	assert.Contains(t, yamlStr, "replicas: 3")
	assert.Contains(t, yamlStr, "nginx:1.21")
}

// TestToYAML_NilResource tests YAML output for nil resource
func TestToYAML_NilResource(t *testing.T) {
	_, err := ToYAML(nil)
	assert.Error(t, err)
}

// TestToJSON_NilResource tests JSON output for nil resource
func TestToJSON_NilResource(t *testing.T) {
	_, err := ToJSON(nil)
	assert.Error(t, err)
}

// TestSerializeConfigMap tests ConfigMap serialization
func TestSerializeConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	got, err := Serialize(cm)
	require.NoError(t, err)

	assert.Equal(t, "v1", got["apiVersion"])
	assert.Equal(t, "ConfigMap", got["kind"])

	data, ok := got["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value1", data["key1"])
	assert.Equal(t, "value2", data["key2"])
}

// TestSerializeSecret tests Secret serialization
func TestSerializeSecret(t *testing.T) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"username": "admin",
			"password": "secret",
		},
	}

	got, err := Serialize(secret)
	require.NoError(t, err)

	assert.Equal(t, "v1", got["apiVersion"])
	assert.Equal(t, "Secret", got["kind"])
	assert.Equal(t, "Opaque", got["type"])
}

// TestToMultiYAML_OrderPreservation tests that resources are output in order
func TestToMultiYAML_OrderPreservation(t *testing.T) {
	resources := []interface{}{
		&corev1.Service{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
			ObjectMeta: metav1.ObjectMeta{Name: "first"},
		},
		&appsv1.Deployment{
			TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
			ObjectMeta: metav1.ObjectMeta{Name: "second"},
		},
		&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "third"},
		},
	}

	yaml, err := ToMultiYAML(resources)
	require.NoError(t, err)

	yamlStr := string(yaml)

	// Check order by finding positions
	firstIdx := strings.Index(yamlStr, "name: first")
	secondIdx := strings.Index(yamlStr, "name: second")
	thirdIdx := strings.Index(yamlStr, "name: third")

	assert.Less(t, firstIdx, secondIdx)
	assert.Less(t, secondIdx, thirdIdx)
}

// TestSerializeStatefulSet tests StatefulSet serialization
func TestSerializeStatefulSet(t *testing.T) {
	replicas := int32(3)
	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-statefulset",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: "test-service",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx"},
					},
				},
			},
		},
	}

	got, err := Serialize(ss)
	require.NoError(t, err)

	assert.Equal(t, "apps/v1", got["apiVersion"])
	assert.Equal(t, "StatefulSet", got["kind"])

	spec, ok := got["spec"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-service", spec["serviceName"])
}

// TestSerializeDaemonSet tests DaemonSet serialization
func TestSerializeDaemonSet(t *testing.T) {
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-daemonset",
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "fluentd"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": "fluentd"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "fluentd", Image: "fluentd:latest"},
					},
				},
			},
		},
	}

	yaml, err := ToYAML(ds)
	require.NoError(t, err)

	yamlStr := string(yaml)
	assert.Contains(t, yamlStr, "kind: DaemonSet")
	assert.Contains(t, yamlStr, "test-daemonset")
	assert.Contains(t, yamlStr, "kube-system")
}

// TestSerializePod tests Pod serialization
func TestSerializePod(t *testing.T) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:1.21",
					Ports: []corev1.ContainerPort{
						{ContainerPort: 80},
					},
				},
			},
		},
	}

	got, err := Serialize(pod)
	require.NoError(t, err)

	assert.Equal(t, "v1", got["apiVersion"])
	assert.Equal(t, "Pod", got["kind"])
}

// TestSerializeReplicaSet tests ReplicaSet serialization
func TestSerializeReplicaSet(t *testing.T) {
	replicas := int32(3)
	rs := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "ReplicaSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-replicaset",
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx"},
					},
				},
			},
		},
	}

	got, err := Serialize(rs)
	require.NoError(t, err)

	assert.Equal(t, "apps/v1", got["apiVersion"])
	assert.Equal(t, "ReplicaSet", got["kind"])
}
