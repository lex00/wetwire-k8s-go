package crd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Fetcher handles fetching CRDs from various sources.
type Fetcher struct {
	// CacheDir is the directory for caching downloaded CRDs
	CacheDir string

	// HTTPClient is the HTTP client for downloading CRDs
	HTTPClient *http.Client
}

// NewFetcher creates a new CRD fetcher.
func NewFetcher(cacheDir string) *Fetcher {
	return &Fetcher{
		CacheDir: cacheDir,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GitHubContent represents a file/directory in GitHub's contents API response.
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

// ConfigConnectorSource represents the Config Connector CRD source.
const ConfigConnectorSource = "https://api.github.com/repos/GoogleCloudPlatform/k8s-config-connector/contents/crds"

// ConfigConnectorRawBase is the raw content base URL.
const ConfigConnectorRawBase = "https://raw.githubusercontent.com/GoogleCloudPlatform/k8s-config-connector/master/crds"

// FetchConfigConnector downloads Config Connector CRDs to the cache directory.
// Returns the path to the directory containing the CRDs.
func (f *Fetcher) FetchConfigConnector(ctx context.Context) (string, error) {
	return f.FetchFromGitHubDir(ctx, ConfigConnectorSource, "config-connector")
}

// FetchFromGitHubDir fetches all YAML files from a GitHub directory.
func (f *Fetcher) FetchFromGitHubDir(ctx context.Context, apiURL string, subdir string) (string, error) {
	// Create cache directory
	cacheDir := filepath.Join(f.CacheDir, subdir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Check if we have a recent cache
	cacheMarker := filepath.Join(cacheDir, ".cache_timestamp")
	if f.isCacheValid(cacheMarker, 24*time.Hour) {
		return cacheDir, nil
	}

	// Fetch directory listing
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch directory listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Download each YAML file
	downloaded := 0
	for _, item := range contents {
		if item.Type != "file" {
			continue
		}

		ext := strings.ToLower(filepath.Ext(item.Name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		destPath := filepath.Join(cacheDir, item.Name)

		// Check if file already exists and is recent
		if info, err := os.Stat(destPath); err == nil {
			if time.Since(info.ModTime()) < 24*time.Hour {
				downloaded++
				continue
			}
		}

		if err := f.downloadFile(ctx, item.DownloadURL, destPath); err != nil {
			return "", fmt.Errorf("failed to download %s: %w", item.Name, err)
		}
		downloaded++
	}

	// Update cache marker
	if err := os.WriteFile(cacheMarker, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "Warning: failed to write cache marker: %v\n", err)
	}

	fmt.Printf("Downloaded %d CRD files to %s\n", downloaded, cacheDir)
	return cacheDir, nil
}

// FetchFile downloads a single CRD file.
func (f *Fetcher) FetchFile(ctx context.Context, url string, destPath string) error {
	return f.downloadFile(ctx, url, destPath)
}

// downloadFile downloads a file from a URL to a local path.
func (f *Fetcher) downloadFile(ctx context.Context, url string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temporary file
	tmpPath := destPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	file.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Rename to final path
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// isCacheValid checks if a cache marker file indicates the cache is still valid.
func (f *Fetcher) isCacheValid(markerPath string, maxAge time.Duration) bool {
	info, err := os.Stat(markerPath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < maxAge
}

// ListCRDFiles lists all CRD YAML files in a directory.
func ListCRDFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// CRDSource represents a source for CRDs.
type CRDSource struct {
	// Type is the source type: "directory", "github", "url"
	Type string

	// Path is the local directory path (for "directory" type)
	Path string

	// URL is the remote URL (for "github" or "url" type)
	URL string

	// Domain is the domain prefix for generated types
	Domain string
}

// ParseCRDSource parses a CRD source string into a CRDSource.
// Formats:
//   - Local path: "./crds" or "/path/to/crds"
//   - GitHub: "github:owner/repo/path"
//   - Config Connector shorthand: "config-connector"
func ParseCRDSource(source string) CRDSource {
	switch {
	case source == "config-connector":
		return CRDSource{
			Type:   "github",
			URL:    ConfigConnectorSource,
			Domain: "cnrm",
		}

	case strings.HasPrefix(source, "github:"):
		// Parse github:owner/repo/path format
		parts := strings.TrimPrefix(source, "github:")
		return CRDSource{
			Type: "github",
			URL:  "https://api.github.com/repos/" + parts,
		}

	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		return CRDSource{
			Type: "url",
			URL:  source,
		}

	default:
		return CRDSource{
			Type: "directory",
			Path: source,
		}
	}
}

// Resolve resolves a CRDSource to a local directory path.
// For remote sources, this downloads to the cache directory.
func (s *CRDSource) Resolve(ctx context.Context, fetcher *Fetcher) (string, error) {
	switch s.Type {
	case "directory":
		// Verify directory exists
		info, err := os.Stat(s.Path)
		if err != nil {
			return "", fmt.Errorf("CRD directory not found: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("CRD path is not a directory: %s", s.Path)
		}
		return s.Path, nil

	case "github":
		// Derive subdir from URL
		subdir := "crds"
		if s.Domain != "" {
			subdir = s.Domain
		}
		return fetcher.FetchFromGitHubDir(ctx, s.URL, subdir)

	case "url":
		return "", fmt.Errorf("direct URL sources not yet implemented")

	default:
		return "", fmt.Errorf("unknown source type: %s", s.Type)
	}
}
