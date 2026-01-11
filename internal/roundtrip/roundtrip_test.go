package roundtrip_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-k8s-go/internal/roundtrip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	// Navigate from internal/roundtrip to testdata/roundtrip/examples
	return filepath.Join("..", "..", "testdata", "roundtrip", "examples")
}

// TestParseMultiDocYAML tests the YAML parsing functionality
func TestParseMultiDocYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected number of documents
		wantErr  bool
	}{
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
		{
			name: "single document",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`,
			expected: 1,
		},
		{
			name: "multiple documents",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test3
`,
			expected: 3,
		},
		{
			name: "empty documents between valid ones",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test1
---
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test2
`,
			expected: 2,
		},
		{
			name:    "invalid YAML",
			input:   "{{invalid}}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := roundtrip.ParseMultiDocYAML([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, docs, tt.expected)
		})
	}
}

// TestNormalizeDocument tests YAML document normalization
func TestNormalizeDocument(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "remove empty values",
			input: map[string]interface{}{
				"name":   "test",
				"labels": map[string]interface{}{},
				"data":   nil,
			},
			expected: map[string]interface{}{
				"name": "test",
			},
		},
		{
			name: "normalize numeric types",
			input: map[string]interface{}{
				"replicas": 3,
				"port":     float64(80),
			},
			expected: map[string]interface{}{
				"replicas": int64(3),
				"port":     int64(80),
			},
		},
		{
			name: "nested maps",
			input: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "",
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test",
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
			},
		},
		{
			name: "slices with empty items",
			input: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1"},
					map[string]interface{}{},
					map[string]interface{}{"name": "item2"},
				},
			},
			expected: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1"},
					map[string]interface{}{"name": "item2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundtrip.NormalizeDocument(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompare tests the comparison functionality
func TestCompare(t *testing.T) {
	tests := []struct {
		name       string
		original   []map[string]interface{}
		result     []map[string]interface{}
		normalize  bool
		equivalent bool
		diffCount  int
	}{
		{
			name:       "identical documents",
			original:   []map[string]interface{}{{"apiVersion": "v1", "kind": "ConfigMap"}},
			result:     []map[string]interface{}{{"apiVersion": "v1", "kind": "ConfigMap"}},
			normalize:  true,
			equivalent: true,
			diffCount:  0,
		},
		{
			name:       "different document count",
			original:   []map[string]interface{}{{"apiVersion": "v1"}},
			result:     []map[string]interface{}{{"apiVersion": "v1"}, {"apiVersion": "v1"}},
			normalize:  true,
			equivalent: false,
			diffCount:  1,
		},
		{
			name:       "missing field",
			original:   []map[string]interface{}{{"apiVersion": "v1", "kind": "ConfigMap"}},
			result:     []map[string]interface{}{{"apiVersion": "v1"}},
			normalize:  true,
			equivalent: false,
			diffCount:  1,
		},
		{
			name:       "added field",
			original:   []map[string]interface{}{{"apiVersion": "v1"}},
			result:     []map[string]interface{}{{"apiVersion": "v1", "kind": "ConfigMap"}},
			normalize:  true,
			equivalent: false,
			diffCount:  1,
		},
		{
			name:       "modified field",
			original:   []map[string]interface{}{{"apiVersion": "v1", "kind": "ConfigMap"}},
			result:     []map[string]interface{}{{"apiVersion": "v1", "kind": "Secret"}},
			normalize:  true,
			equivalent: false,
			diffCount:  1,
		},
		{
			name: "numeric type normalization",
			original: []map[string]interface{}{
				{"replicas": 3},
			},
			result: []map[string]interface{}{
				{"replicas": float64(3)},
			},
			normalize:  true,
			equivalent: true,
			diffCount:  0,
		},
		{
			name: "nested differences",
			original: []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name": "test",
						"labels": map[string]interface{}{
							"app": "myapp",
						},
					},
				},
			},
			result: []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name": "test",
						"labels": map[string]interface{}{
							"app": "different",
						},
					},
				},
			},
			normalize:  true,
			equivalent: false,
			diffCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equivalent, diffs := roundtrip.Compare(tt.original, tt.result, tt.normalize)
			assert.Equal(t, tt.equivalent, equivalent)
			assert.Len(t, diffs, tt.diffCount)
		})
	}
}

// TestNormalizeYAML tests the YAML normalization function
func TestNormalizeYAML(t *testing.T) {
	input := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  labels:
    app: myapp
data:
  key1: value1
`)

	normalized, err := roundtrip.NormalizeYAML(input)
	require.NoError(t, err)
	assert.NotEmpty(t, normalized)

	// Parse the normalized output to verify structure
	docs, err := roundtrip.ParseMultiDocYAML(normalized)
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "v1", docs[0]["apiVersion"])
	assert.Equal(t, "ConfigMap", docs[0]["kind"])
}

// TestDiffSummary tests the difference summary formatting
func TestDiffSummary(t *testing.T) {
	tests := []struct {
		name        string
		differences []roundtrip.Difference
		wantSubstr  []string
	}{
		{
			name:        "no differences",
			differences: []roundtrip.Difference{},
			wantSubstr:  []string{"No differences"},
		},
		{
			name: "missing field",
			differences: []roundtrip.Difference{
				{Path: "metadata.name", Type: roundtrip.DiffTypeMissing, Original: "test"},
			},
			wantSubstr: []string{"MISSING", "metadata.name"},
		},
		{
			name: "added field",
			differences: []roundtrip.Difference{
				{Path: "spec.replicas", Type: roundtrip.DiffTypeAdded, Result: 3},
			},
			wantSubstr: []string{"ADDED", "spec.replicas"},
		},
		{
			name: "modified field",
			differences: []roundtrip.Difference{
				{Path: "kind", Type: roundtrip.DiffTypeModified, Original: "ConfigMap", Result: "Secret"},
			},
			wantSubstr: []string{"MODIFIED", "kind", "ConfigMap", "Secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := roundtrip.DiffSummary(tt.differences)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, summary, substr)
			}
		})
	}
}

// TestFieldPath tests the FieldPath type
func TestFieldPath(t *testing.T) {
	tests := []struct {
		name     string
		path     roundtrip.FieldPath
		expected string
	}{
		{
			name:     "empty path",
			path:     roundtrip.FieldPath{},
			expected: "",
		},
		{
			name:     "single segment",
			path:     roundtrip.FieldPath{"metadata"},
			expected: "metadata",
		},
		{
			name:     "multiple segments",
			path:     roundtrip.FieldPath{"metadata", "labels", "app"},
			expected: "metadata.labels.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.path.String())
		})
	}
}

// TestFieldPathAppend tests FieldPath.Append
func TestFieldPathAppend(t *testing.T) {
	path := roundtrip.FieldPath{"metadata"}
	newPath := path.Append("name")

	assert.Equal(t, "metadata.name", newPath.String())
	assert.Equal(t, "metadata", path.String()) // Original should be unchanged
}

// TestReferenceYAMLFiles tests that reference YAML files can be parsed
func TestReferenceYAMLFiles(t *testing.T) {
	files := []string{
		"deployment.yaml",
		"service.yaml",
		"configmap.yaml",
		"multi-document.yaml",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(testdataDir(), file)
			data, err := os.ReadFile(path)
			require.NoError(t, err, "failed to read %s", path)

			docs, err := roundtrip.ParseMultiDocYAML(data)
			require.NoError(t, err, "failed to parse %s", path)
			assert.NotEmpty(t, docs, "no documents found in %s", path)

			// Verify each document has required fields
			for i, doc := range docs {
				assert.Contains(t, doc, "apiVersion", "doc %d in %s missing apiVersion", i, file)
				assert.Contains(t, doc, "kind", "doc %d in %s missing kind", i, file)
			}
		})
	}
}

// TestReferenceYAMLNormalization tests that reference files can be normalized
func TestReferenceYAMLNormalization(t *testing.T) {
	files := []string{
		"deployment.yaml",
		"service.yaml",
		"configmap.yaml",
		"multi-document.yaml",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(testdataDir(), file)
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			normalized, err := roundtrip.NormalizeYAML(data)
			require.NoError(t, err)
			assert.NotEmpty(t, normalized)

			// Verify normalized YAML is still valid
			docs, err := roundtrip.ParseMultiDocYAML(normalized)
			require.NoError(t, err)
			assert.NotEmpty(t, docs)
		})
	}
}

// TestMultiDocumentYAML tests multi-document YAML handling
func TestMultiDocumentYAML(t *testing.T) {
	path := filepath.Join(testdataDir(), "multi-document.yaml")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	docs, err := roundtrip.ParseMultiDocYAML(data)
	require.NoError(t, err)
	assert.Len(t, docs, 4, "multi-document.yaml should have 4 documents")

	// Verify document kinds
	expectedKinds := []string{"Namespace", "ConfigMap", "Deployment", "Service"}
	for i, kind := range expectedKinds {
		assert.Equal(t, kind, docs[i]["kind"], "document %d should be %s", i, kind)
	}
}

// TestEdgeCases tests edge cases in comparison
func TestEdgeCases(t *testing.T) {
	t.Run("empty values", func(t *testing.T) {
		original := []map[string]interface{}{
			{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"data":       map[string]interface{}{},
			},
		}
		result := []map[string]interface{}{
			{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
		}

		equivalent, _ := roundtrip.Compare(original, result, true)
		assert.True(t, equivalent, "empty maps should be normalized away")
	})

	t.Run("nested structures", func(t *testing.T) {
		original := []map[string]interface{}{
			{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "app",
									"image": "nginx:latest",
								},
							},
						},
					},
				},
			},
		}
		result := []map[string]interface{}{
			{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "app",
									"image": "nginx:latest",
								},
							},
						},
					},
				},
			},
		}

		equivalent, diffs := roundtrip.Compare(original, result, true)
		assert.True(t, equivalent)
		assert.Empty(t, diffs)
	})

	t.Run("array length mismatch", func(t *testing.T) {
		original := []map[string]interface{}{
			{
				"containers": []interface{}{
					map[string]interface{}{"name": "c1"},
					map[string]interface{}{"name": "c2"},
				},
			},
		}
		result := []map[string]interface{}{
			{
				"containers": []interface{}{
					map[string]interface{}{"name": "c1"},
				},
			},
		}

		equivalent, diffs := roundtrip.Compare(original, result, true)
		assert.False(t, equivalent)
		assert.NotEmpty(t, diffs)
	})
}

// TestComparisonModes tests different comparison modes
func TestComparisonModes(t *testing.T) {
	t.Run("strict mode with extra whitespace", func(t *testing.T) {
		input1 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`)
		input2 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test

`)
		docs1, err := roundtrip.ParseMultiDocYAML(input1)
		require.NoError(t, err)
		docs2, err := roundtrip.ParseMultiDocYAML(input2)
		require.NoError(t, err)

		// Should be equivalent after normalization
		equivalent, _ := roundtrip.Compare(docs1, docs2, true)
		assert.True(t, equivalent)
	})
}
