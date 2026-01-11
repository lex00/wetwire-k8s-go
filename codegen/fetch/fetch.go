package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Fetcher handles fetching and caching Kubernetes OpenAPI schemas.
type Fetcher struct {
	cacheDir string
	client   *http.Client
}

// NewFetcher creates a new schema fetcher with the given cache directory.
func NewFetcher(cacheDir string) *Fetcher {
	return &Fetcher{
		cacheDir: cacheDir,
		client:   &http.Client{},
	}
}

// GetSchemaURL returns the URL for downloading a specific Kubernetes version's OpenAPI schema.
func GetSchemaURL(version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version cannot be empty")
	}

	// Ensure version has 'v' prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json", version), nil
}

// FetchSchema fetches the OpenAPI schema for the given Kubernetes version.
// It uses a local cache if available, otherwise downloads from GitHub.
func (f *Fetcher) FetchSchema(ctx context.Context, version string) (*Schema, error) {
	// Check cache first
	cachedPath := filepath.Join(f.cacheDir, fmt.Sprintf("%s.json", version))
	if _, err := os.Stat(cachedPath); err == nil {
		return f.LoadSchemaFromFile(cachedPath)
	}

	// Download schema
	url, err := GetSchemaURL(version)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download schema: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download schema: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse schema
	var schema Schema
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Cache the schema
	if err := os.MkdirAll(f.cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(cachedPath, body, 0644); err != nil {
		return nil, fmt.Errorf("failed to cache schema: %w", err)
	}

	return &schema, nil
}

// FetchSchemaToFile downloads the schema and saves it to the specified file path.
func (f *Fetcher) FetchSchemaToFile(ctx context.Context, version string, outputPath string) error {
	schema, err := f.FetchSchema(ctx, version)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}

// LoadSchemaFromFile loads a schema from a JSON file.
func (f *Fetcher) LoadSchemaFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	return &schema, nil
}
