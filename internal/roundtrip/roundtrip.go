// Package roundtrip provides utilities for testing round-trip conversion
// of Kubernetes YAML manifests through the wetwire import/build pipeline.
//
// A round-trip test verifies that:
// YAML -> Import (Go code) -> Build -> YAML produces semantically equivalent output.
package roundtrip

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Difference represents a single difference between two YAML documents.
type Difference struct {
	Path     string      // JSON path to the differing field (e.g., "metadata.labels.app")
	Original interface{} // Value in original document
	Result   interface{} // Value in result document
	Type     DiffType    // Type of difference
}

// DiffType represents the type of difference found.
type DiffType string

const (
	DiffTypeMissing  DiffType = "missing"  // Field exists in original but not in result
	DiffTypeAdded    DiffType = "added"    // Field exists in result but not in original
	DiffTypeModified DiffType = "modified" // Field exists in both but values differ
)

// Result contains the output of a round-trip operation.
type Result struct {
	OriginalYAML []byte       // The original input YAML
	GoCode       string       // The generated Go code
	ResultYAML   []byte       // The YAML produced after building the Go code
	Equivalent   bool         // Whether the original and result are semantically equivalent
	Differences  []Difference // List of differences if not equivalent
}

// Options configures the round-trip operation.
type Options struct {
	PackageName string // Go package name for generated code (default: "main")
	Normalize   bool   // Whether to normalize YAML before comparison (default: true)
}

// DefaultOptions returns the default round-trip options.
func DefaultOptions() Options {
	return Options{
		PackageName: "main",
		Normalize:   true,
	}
}

// RoundTrip performs a complete round-trip conversion:
// YAML input -> Go code -> YAML output
//
// This is a high-level function that coordinates the import and build process.
// It returns the result along with any semantic differences found.
func RoundTrip(yamlInput []byte, opts Options) (*Result, error) {
	if len(yamlInput) == 0 {
		return &Result{
			OriginalYAML: yamlInput,
			Equivalent:   true,
		}, nil
	}

	// Parse input YAML to validate it
	originalDocs, err := ParseMultiDocYAML(yamlInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input YAML: %w", err)
	}

	if len(originalDocs) == 0 {
		return &Result{
			OriginalYAML: yamlInput,
			Equivalent:   true,
		}, nil
	}

	// Create a temporary directory for the round-trip
	tmpDir, err := os.MkdirTemp("", "roundtrip-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write input YAML to temp file
	inputFile := filepath.Join(tmpDir, "input.yaml")
	if err := os.WriteFile(inputFile, yamlInput, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// Generate Go code using the importer
	goCode, err := importYAML(inputFile, opts.PackageName)
	if err != nil {
		return nil, fmt.Errorf("import failed: %w", err)
	}

	// Write Go code to temp file
	goFile := filepath.Join(tmpDir, "resources.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Go file: %w", err)
	}

	// Build YAML from Go code
	resultYAML, err := buildYAML(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Parse result YAML
	resultDocs, err := ParseMultiDocYAML(resultYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result YAML: %w", err)
	}

	// Compare documents
	equivalent, differences := Compare(originalDocs, resultDocs, opts.Normalize)

	return &Result{
		OriginalYAML: yamlInput,
		GoCode:       goCode,
		ResultYAML:   resultYAML,
		Equivalent:   equivalent,
		Differences:  differences,
	}, nil
}

// ParseMultiDocYAML parses a multi-document YAML byte slice into a slice of documents.
func ParseMultiDocYAML(data []byte) ([]map[string]interface{}, error) {
	var docs []map[string]interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if doc == nil || len(doc) == 0 {
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// Compare performs semantic comparison between original and result YAML documents.
// Returns true if documents are equivalent, along with a list of differences if not.
func Compare(original, result []map[string]interface{}, normalize bool) (bool, []Difference) {
	var differences []Difference

	// Handle different document counts
	if len(original) != len(result) {
		differences = append(differences, Difference{
			Path:     "",
			Original: len(original),
			Result:   len(result),
			Type:     DiffTypeModified,
		})
		return false, differences
	}

	// Compare each document
	for i := range original {
		origDoc := original[i]
		resultDoc := result[i]

		if normalize {
			origDoc = NormalizeDocument(origDoc)
			resultDoc = NormalizeDocument(resultDoc)
		}

		docDiffs := compareDocuments(origDoc, resultDoc, fmt.Sprintf("doc[%d]", i))
		differences = append(differences, docDiffs...)
	}

	return len(differences) == 0, differences
}

// NormalizeDocument normalizes a YAML document for comparison.
// It performs the following normalizations:
// - Sorts map keys alphabetically
// - Removes empty/nil values
// - Normalizes numeric types
func NormalizeDocument(doc map[string]interface{}) map[string]interface{} {
	return normalizeValue(doc).(map[string]interface{})
}

// NormalizeYAML parses and normalizes YAML for comparison.
// It returns normalized YAML that can be compared byte-for-byte.
func NormalizeYAML(yamlData []byte) ([]byte, error) {
	docs, err := ParseMultiDocYAML(yamlData)
	if err != nil {
		return nil, err
	}

	var normalizedDocs []map[string]interface{}
	for _, doc := range docs {
		normalizedDocs = append(normalizedDocs, NormalizeDocument(doc))
	}

	return serializeDocuments(normalizedDocs)
}

// normalizeValue recursively normalizes a value for comparison.
func normalizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			normalized := normalizeValue(v)
			// Skip nil and empty values
			if !isEmptyValue(normalized) {
				result[k] = normalized
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result

	case []interface{}:
		var result []interface{}
		for _, item := range val {
			normalized := normalizeValue(item)
			if !isEmptyValue(normalized) {
				result = append(result, normalized)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result

	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val

	default:
		return val
	}
}

// isEmptyValue checks if a value should be considered empty.
func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return len(val) == 0
	case []interface{}:
		return len(val) == 0
	case string:
		return val == ""
	default:
		return false
	}
}

// compareDocuments compares two documents and returns differences.
func compareDocuments(original, result map[string]interface{}, pathPrefix string) []Difference {
	var differences []Difference

	// Get all keys from both documents
	allKeys := make(map[string]bool)
	for k := range original {
		allKeys[k] = true
	}
	for k := range result {
		allKeys[k] = true
	}

	// Sort keys for consistent ordering
	sortedKeys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		path := key
		if pathPrefix != "" {
			path = pathPrefix + "." + key
		}

		origVal, origExists := original[key]
		resultVal, resultExists := result[key]

		if origExists && !resultExists {
			differences = append(differences, Difference{
				Path:     path,
				Original: origVal,
				Result:   nil,
				Type:     DiffTypeMissing,
			})
			continue
		}

		if !origExists && resultExists {
			differences = append(differences, Difference{
				Path:     path,
				Original: nil,
				Result:   resultVal,
				Type:     DiffTypeAdded,
			})
			continue
		}

		// Both exist - compare values
		diffs := compareValues(origVal, resultVal, path)
		differences = append(differences, diffs...)
	}

	return differences
}

// compareValues compares two values and returns differences.
func compareValues(original, result interface{}, path string) []Difference {
	var differences []Difference

	// Handle nil cases
	if original == nil && result == nil {
		return nil
	}
	if original == nil && result != nil {
		return []Difference{{
			Path:     path,
			Original: nil,
			Result:   result,
			Type:     DiffTypeAdded,
		}}
	}
	if original != nil && result == nil {
		return []Difference{{
			Path:     path,
			Original: original,
			Result:   nil,
			Type:     DiffTypeMissing,
		}}
	}

	// Type-based comparison
	switch origVal := original.(type) {
	case map[string]interface{}:
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			return []Difference{{
				Path:     path,
				Original: original,
				Result:   result,
				Type:     DiffTypeModified,
			}}
		}
		return compareDocuments(origVal, resultMap, path)

	case []interface{}:
		resultSlice, ok := result.([]interface{})
		if !ok {
			return []Difference{{
				Path:     path,
				Original: original,
				Result:   result,
				Type:     DiffTypeModified,
			}}
		}
		return compareSlices(origVal, resultSlice, path)

	default:
		if !valuesEqual(original, result) {
			return []Difference{{
				Path:     path,
				Original: original,
				Result:   result,
				Type:     DiffTypeModified,
			}}
		}
	}

	return differences
}

// compareSlices compares two slices and returns differences.
func compareSlices(original, result []interface{}, path string) []Difference {
	var differences []Difference

	// Compare lengths
	if len(original) != len(result) {
		return []Difference{{
			Path:     path,
			Original: fmt.Sprintf("length=%d", len(original)),
			Result:   fmt.Sprintf("length=%d", len(result)),
			Type:     DiffTypeModified,
		}}
	}

	// Compare elements
	for i := range original {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		diffs := compareValues(original[i], result[i], elemPath)
		differences = append(differences, diffs...)
	}

	return differences
}

// valuesEqual compares two scalar values for equality.
func valuesEqual(a, b interface{}) bool {
	// Handle numeric type differences (int vs float64 from YAML parsing)
	aNum, aIsNum := toFloat64(a)
	bNum, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// Compare strings
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	if aIsStr && bIsStr {
		return aStr == bStr
	}

	// Use reflect.DeepEqual for other types
	return reflect.DeepEqual(a, b)
}

// toFloat64 converts a numeric value to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// serializeDocuments serializes documents back to multi-document YAML.
func serializeDocuments(docs []map[string]interface{}) ([]byte, error) {
	if len(docs) == 0 {
		return []byte{}, nil
	}

	var documents []string
	for _, doc := range docs {
		yamlBytes, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		documents = append(documents, strings.TrimSpace(string(yamlBytes)))
	}

	return []byte(strings.Join(documents, "\n---\n")), nil
}

// importYAML runs the wetwire-k8s import command on a YAML file.
func importYAML(inputFile, packageName string) (string, error) {
	// Find the wetwire-k8s binary
	binary, err := findBinary()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(binary, "import", "-p", packageName, inputFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("import command failed: %s\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// buildYAML runs the wetwire-k8s build command on a directory.
func buildYAML(dir string) ([]byte, error) {
	// Find the wetwire-k8s binary
	binary, err := findBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binary, "build", dir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("build command failed: %s\n%s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// findBinary finds the wetwire-k8s binary.
func findBinary() (string, error) {
	// Check for binary in current directory first
	if _, err := os.Stat("./wetwire-k8s"); err == nil {
		return "./wetwire-k8s", nil
	}

	// Check in PATH
	binary, err := exec.LookPath("wetwire-k8s")
	if err == nil {
		return binary, nil
	}

	// Check in project root
	projectRoot := findProjectRoot()
	if projectRoot != "" {
		binaryPath := filepath.Join(projectRoot, "wetwire-k8s")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("wetwire-k8s binary not found")
}

// findProjectRoot attempts to find the project root directory.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// DiffSummary returns a human-readable summary of differences.
func DiffSummary(differences []Difference) string {
	if len(differences) == 0 {
		return "No differences found"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Found %d difference(s):\n", len(differences)))

	for _, diff := range differences {
		switch diff.Type {
		case DiffTypeMissing:
			summary.WriteString(fmt.Sprintf("  - MISSING: %s (was: %v)\n", diff.Path, diff.Original))
		case DiffTypeAdded:
			summary.WriteString(fmt.Sprintf("  + ADDED: %s (now: %v)\n", diff.Path, diff.Result))
		case DiffTypeModified:
			summary.WriteString(fmt.Sprintf("  ~ MODIFIED: %s (was: %v, now: %v)\n", diff.Path, diff.Original, diff.Result))
		}
	}

	return summary.String()
}
