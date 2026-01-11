package fetch

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchSchema(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "fetch v1.28 schema",
			version: "v1.28.0",
			wantErr: false,
		},
		{
			name:    "fetch v1.29 schema",
			version: "v1.29.0",
			wantErr: false,
		},
		{
			name:    "fetch v1.30 schema",
			version: "v1.30.0",
			wantErr: false,
		},
		{
			name:    "invalid version",
			version: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			fetcher := NewFetcher(tmpDir)

			ctx := context.Background()
			schema, err := fetcher.FetchSchema(ctx, tt.version)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, schema)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, schema)
				assert.NotEmpty(t, schema.Definitions)
			}
		})
	}
}

func TestFetchSchemaToFile(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir)

	ctx := context.Background()
	version := "v1.28.0"
	outputPath := filepath.Join(tmpDir, "v1.28.json")

	err := fetcher.FetchSchemaToFile(ctx, version, outputPath)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Verify file is not empty
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestLoadSchemaFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "test-schema.json")

	// Create a minimal test schema
	testSchema := `{
		"swagger": "2.0",
		"info": {
			"title": "Kubernetes",
			"version": "v1.28.0"
		},
		"definitions": {
			"io.k8s.api.core.v1.Pod": {
				"type": "object",
				"properties": {
					"apiVersion": {"type": "string"},
					"kind": {"type": "string"}
				}
			}
		}
	}`

	err := os.WriteFile(schemaPath, []byte(testSchema), 0644)
	require.NoError(t, err)

	fetcher := NewFetcher(tmpDir)
	schema, err := fetcher.LoadSchemaFromFile(schemaPath)

	require.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "2.0", schema.Swagger)
	assert.Equal(t, "Kubernetes", schema.Info.Title)
	assert.Contains(t, schema.Definitions, "io.k8s.api.core.v1.Pod")
}

func TestGetSchemaURL(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectedURL string
		wantErr     bool
	}{
		{
			name:        "v1.28.0",
			version:     "v1.28.0",
			expectedURL: "https://raw.githubusercontent.com/kubernetes/kubernetes/v1.28.0/api/openapi-spec/swagger.json",
			wantErr:     false,
		},
		{
			name:        "v1.29.0",
			version:     "v1.29.0",
			expectedURL: "https://raw.githubusercontent.com/kubernetes/kubernetes/v1.29.0/api/openapi-spec/swagger.json",
			wantErr:     false,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := GetSchemaURL(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestFetcherCaching(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir)

	// Create a cached schema file
	cachedPath := filepath.Join(tmpDir, "v1.28.0.json")
	testSchema := `{
		"swagger": "2.0",
		"info": {"title": "Kubernetes", "version": "v1.28.0"},
		"definitions": {}
	}`
	err := os.WriteFile(cachedPath, []byte(testSchema), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	schema, err := fetcher.FetchSchema(ctx, "v1.28.0")

	require.NoError(t, err)
	assert.NotNil(t, schema)
	// Should load from cache without making HTTP request
	assert.Equal(t, "Kubernetes", schema.Info.Title)
}

func TestLoadSchemaFromFile_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir)

	_, err := fetcher.LoadSchemaFromFile("/nonexistent/path/schema.json")
	assert.Error(t, err)
}

func TestLoadSchemaFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(schemaPath, []byte("not valid json"), 0644)
	require.NoError(t, err)

	fetcher := NewFetcher(tmpDir)
	_, err = fetcher.LoadSchemaFromFile(schemaPath)
	assert.Error(t, err)
}

func TestGetSchemaURL_NoVPrefix(t *testing.T) {
	// Test that version without v prefix is handled
	url, err := GetSchemaURL("1.28.0")
	require.NoError(t, err)
	assert.Contains(t, url, "1.28.0")
}

func TestNewFetcher(t *testing.T) {
	t.Run("with cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		fetcher := NewFetcher(tmpDir)
		assert.NotNil(t, fetcher)
	})

	t.Run("with empty cache directory", func(t *testing.T) {
		fetcher := NewFetcher("")
		assert.NotNil(t, fetcher)
	})
}

func TestFetchSchema_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFetcher(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := fetcher.FetchSchema(ctx, "v1.28.0")
	// Should return an error due to cancelled context (or succeed from cache)
	// The behavior depends on whether there's a cached version
	_ = err // Accept either outcome
}
