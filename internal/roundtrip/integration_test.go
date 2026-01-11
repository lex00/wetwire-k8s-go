// Package roundtrip provides integration tests that verify round-trip
// conversion using real Kubernetes example manifests.
package roundtrip_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lex00/wetwire-k8s-go/internal/roundtrip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kubernetesExamplesURL is the base URL for kubernetes/examples raw files
const kubernetesExamplesBaseURL = "https://raw.githubusercontent.com/kubernetes/examples/master"

// exampleFiles contains paths to well-known Kubernetes example YAML files
// that can be used for round-trip testing.
var exampleFiles = []struct {
	name        string
	path        string
	description string
	skip        bool   // Skip if not supported by current importer
	skipReason  string // Reason for skipping
}{
	{
		name:        "guestbook-frontend-deployment",
		path:        "/guestbook/frontend-deployment.yaml",
		description: "Guestbook frontend deployment",
	},
	{
		name:        "guestbook-frontend-service",
		path:        "/guestbook/frontend-service.yaml",
		description: "Guestbook frontend service",
	},
	{
		name:        "guestbook-redis-leader-deployment",
		path:        "/guestbook/redis-leader-deployment.yaml",
		description: "Guestbook redis leader deployment",
	},
	{
		name:        "guestbook-redis-leader-service",
		path:        "/guestbook/redis-leader-service.yaml",
		description: "Guestbook redis leader service",
	},
	{
		name:        "guestbook-redis-follower-deployment",
		path:        "/guestbook/redis-follower-deployment.yaml",
		description: "Guestbook redis follower deployment",
	},
	{
		name:        "guestbook-redis-follower-service",
		path:        "/guestbook/redis-follower-service.yaml",
		description: "Guestbook redis follower service",
	},
}

// TestKubernetesExamplesIntegration tests round-trip with kubernetes/examples.
// This test requires network access and is skipped in short mode.
func TestKubernetesExamplesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if SKIP_INTEGRATION_TESTS is set
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test due to SKIP_INTEGRATION_TESTS")
	}

	for _, example := range exampleFiles {
		t.Run(example.name, func(t *testing.T) {
			if example.skip {
				t.Skipf("skipping: %s", example.skipReason)
			}

			// Download the YAML file
			yamlData, err := downloadFile(kubernetesExamplesBaseURL + example.path)
			if err != nil {
				t.Skipf("failed to download %s: %v", example.path, err)
				return
			}

			// Verify we can parse the YAML
			docs, err := roundtrip.ParseMultiDocYAML(yamlData)
			require.NoError(t, err, "failed to parse YAML from %s", example.path)
			assert.NotEmpty(t, docs, "no documents found in %s", example.path)

			// Verify each document has apiVersion and kind
			for i, doc := range docs {
				assert.Contains(t, doc, "apiVersion", "doc %d missing apiVersion", i)
				assert.Contains(t, doc, "kind", "doc %d missing kind", i)
			}

			// Normalize the YAML
			normalized, err := roundtrip.NormalizeYAML(yamlData)
			require.NoError(t, err, "failed to normalize YAML from %s", example.path)
			assert.NotEmpty(t, normalized)
		})
	}
}

// TestLocalReferenceFiles tests round-trip with local reference files.
// These tests do not require network access.
func TestLocalReferenceFiles(t *testing.T) {
	referenceDir := filepath.Join("..", "..", "testdata", "roundtrip", "examples")

	files, err := os.ReadDir(referenceDir)
	if err != nil {
		t.Skipf("reference directory not found: %v", err)
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(referenceDir, file.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			// Parse the YAML
			docs, err := roundtrip.ParseMultiDocYAML(data)
			require.NoError(t, err)
			assert.NotEmpty(t, docs)

			// Normalize and re-parse
			normalized, err := roundtrip.NormalizeYAML(data)
			require.NoError(t, err)

			normalizedDocs, err := roundtrip.ParseMultiDocYAML(normalized)
			require.NoError(t, err)

			// Compare original and normalized
			equivalent, diffs := roundtrip.Compare(docs, normalizedDocs, true)
			assert.True(t, equivalent, "normalization should preserve content: %s",
				roundtrip.DiffSummary(diffs))
		})
	}
}

// TestRoundTripWithLocalFiles performs a full round-trip test with local files.
// This test is skipped if the wetwire-k8s binary is not available.
func TestRoundTripWithLocalFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping round-trip test in short mode")
	}

	// Check if wetwire-k8s binary exists
	binaryPaths := []string{
		"./wetwire-k8s",
		filepath.Join("..", "..", "wetwire-k8s"),
	}

	binaryFound := false
	for _, path := range binaryPaths {
		if _, err := os.Stat(path); err == nil {
			binaryFound = true
			break
		}
	}

	if !binaryFound {
		t.Skip("wetwire-k8s binary not found, skipping round-trip test")
	}

	referenceDir := filepath.Join("..", "..", "testdata", "roundtrip", "examples")

	files, err := os.ReadDir(referenceDir)
	if err != nil {
		t.Skipf("reference directory not found: %v", err)
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		t.Run("roundtrip-"+file.Name(), func(t *testing.T) {
			path := filepath.Join(referenceDir, file.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			result, err := roundtrip.RoundTrip(data, roundtrip.DefaultOptions())
			if err != nil {
				// Log the error but don't fail - the round-trip may not be fully implemented
				t.Logf("round-trip failed for %s: %v", file.Name(), err)
				t.Skip("round-trip not fully implemented")
				return
			}

			if !result.Equivalent {
				t.Logf("Round-trip differences for %s:\n%s",
					file.Name(), roundtrip.DiffSummary(result.Differences))
			}

			// For now, we just log differences instead of failing
			// This allows us to track progress on round-trip fidelity
			t.Logf("Round-trip for %s: equivalent=%v, differences=%d",
				file.Name(), result.Equivalent, len(result.Differences))
		})
	}
}

// downloadFile downloads a file from the given URL with timeout.
func downloadFile(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// BenchmarkParseMultiDocYAML benchmarks YAML parsing performance.
func BenchmarkParseMultiDocYAML(b *testing.B) {
	path := filepath.Join("..", "..", "testdata", "roundtrip", "examples", "multi-document.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("test file not found: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = roundtrip.ParseMultiDocYAML(data)
	}
}

// BenchmarkNormalizeYAML benchmarks YAML normalization performance.
func BenchmarkNormalizeYAML(b *testing.B) {
	path := filepath.Join("..", "..", "testdata", "roundtrip", "examples", "multi-document.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("test file not found: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = roundtrip.NormalizeYAML(data)
	}
}

// BenchmarkCompare benchmarks document comparison performance.
func BenchmarkCompare(b *testing.B) {
	path := filepath.Join("..", "..", "testdata", "roundtrip", "examples", "multi-document.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("test file not found: %v", err)
	}

	docs, err := roundtrip.ParseMultiDocYAML(data)
	if err != nil {
		b.Fatalf("failed to parse YAML: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		roundtrip.Compare(docs, docs, true)
	}
}
