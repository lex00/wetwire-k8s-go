// Package differ provides semantic comparison of Kubernetes manifests.
// It can be used by any domain that outputs K8s-style YAML (GCP Config Connector,
// ACK, ASO, standard K8s resources).
package differ

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/lex00/wetwire-core-go/domain"
	"gopkg.in/yaml.v3"
)

// K8sDiffer implements domain.Differ for Kubernetes manifests.
type K8sDiffer struct{}

// Compile-time interface check
var _ domain.Differ = (*K8sDiffer)(nil)

// New creates a new K8sDiffer.
func New() *K8sDiffer {
	return &K8sDiffer{}
}

// Diff compares two K8s manifest files and returns their semantic differences.
func (d *K8sDiffer) Diff(ctx *domain.Context, file1, file2 string, opts domain.DiffOpts) (*domain.DiffResult, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file1, err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file2, err)
	}

	return Compare(data1, data2, opts)
}

// Compare compares two YAML manifest contents and returns semantic differences.
func Compare(data1, data2 []byte, opts domain.DiffOpts) (*domain.DiffResult, error) {
	resources1, err := parseManifests(data1)
	if err != nil {
		return nil, fmt.Errorf("parse first manifest: %w", err)
	}

	resources2, err := parseManifests(data2)
	if err != nil {
		return nil, fmt.Errorf("parse second manifest: %w", err)
	}

	return compareResources(resources1, resources2, opts), nil
}

// resource represents a parsed K8s resource.
type resource struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   metadata               `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
	raw        map[string]interface{}
}

type metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// key returns a unique identifier for the resource.
func (r resource) key() string {
	parts := []string{r.APIVersion, r.Kind}
	if r.Metadata.Namespace != "" {
		parts = append(parts, r.Metadata.Namespace)
	}
	parts = append(parts, r.Metadata.Name)
	return strings.Join(parts, "/")
}

// displayKey returns a shorter key for display purposes.
func (r resource) displayKey() string {
	if r.Metadata.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s", r.Kind, r.Metadata.Namespace, r.Metadata.Name)
	}
	return fmt.Sprintf("%s/%s", r.Kind, r.Metadata.Name)
}

// parseManifests parses multi-document YAML into resources.
func parseManifests(data []byte) ([]resource, error) {
	var resources []resource

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var raw map[string]interface{}
		err := decoder.Decode(&raw)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		if raw == nil {
			continue
		}

		r := resource{raw: raw}
		if v, ok := raw["apiVersion"].(string); ok {
			r.APIVersion = v
		}
		if v, ok := raw["kind"].(string); ok {
			r.Kind = v
		}
		if meta, ok := raw["metadata"].(map[string]interface{}); ok {
			if name, ok := meta["name"].(string); ok {
				r.Metadata.Name = name
			}
			if ns, ok := meta["namespace"].(string); ok {
				r.Metadata.Namespace = ns
			}
		}
		if spec, ok := raw["spec"].(map[string]interface{}); ok {
			r.Spec = spec
		}

		resources = append(resources, r)
	}

	return resources, nil
}

// compareResources compares two sets of resources.
func compareResources(resources1, resources2 []resource, opts domain.DiffOpts) *domain.DiffResult {
	// Build maps for lookup
	map1 := make(map[string]resource)
	for _, r := range resources1 {
		map1[r.key()] = r
	}

	map2 := make(map[string]resource)
	for _, r := range resources2 {
		map2[r.key()] = r
	}

	var entries []domain.DiffEntry

	// Find removed (in 1, not in 2)
	for key, r := range map1 {
		if _, exists := map2[key]; !exists {
			entries = append(entries, domain.DiffEntry{
				Resource: r.displayKey(),
				Type:     r.Kind,
				Action:   "removed",
			})
		}
	}

	// Find added (in 2, not in 1)
	for key, r := range map2 {
		if _, exists := map1[key]; !exists {
			entries = append(entries, domain.DiffEntry{
				Resource: r.displayKey(),
				Type:     r.Kind,
				Action:   "added",
			})
		}
	}

	// Find modified (in both, but different)
	for key, r1 := range map1 {
		if r2, exists := map2[key]; exists {
			changes := compareSpecs(r1.raw, r2.raw, "", opts)
			if len(changes) > 0 {
				entries = append(entries, domain.DiffEntry{
					Resource: r1.displayKey(),
					Type:     r1.Kind,
					Action:   "modified",
					Changes:  changes,
				})
			}
		}
	}

	// Sort for consistent output
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Action != entries[j].Action {
			// Order: added, modified, removed
			order := map[string]int{"added": 0, "modified": 1, "removed": 2}
			return order[entries[i].Action] < order[entries[j].Action]
		}
		return entries[i].Resource < entries[j].Resource
	})

	// Calculate summary
	var added, removed, modified int
	for _, e := range entries {
		switch e.Action {
		case "added":
			added++
		case "removed":
			removed++
		case "modified":
			modified++
		}
	}

	return &domain.DiffResult{
		Entries: entries,
		Summary: domain.DiffSummary{
			Added:    added,
			Removed:  removed,
			Modified: modified,
			Total:    added + removed + modified,
		},
	}
}

// compareSpecs compares two spec maps and returns a list of changes.
func compareSpecs(v1, v2 interface{}, path string, opts domain.DiffOpts) []string {
	var changes []string

	// Handle nil cases
	if v1 == nil && v2 == nil {
		return nil
	}
	if v1 == nil {
		if path != "" {
			return []string{fmt.Sprintf("%s: added", path)}
		}
		return nil
	}
	if v2 == nil {
		if path != "" {
			return []string{fmt.Sprintf("%s: removed", path)}
		}
		return nil
	}

	// Type mismatch
	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		if path != "" {
			return []string{fmt.Sprintf("%s: type changed", path)}
		}
		return nil
	}

	switch val1 := v1.(type) {
	case map[string]interface{}:
		val2 := v2.(map[string]interface{})
		// Check for keys in v1 not in v2
		for k := range val1 {
			if _, exists := val2[k]; !exists {
				subPath := joinPath(path, k)
				changes = append(changes, fmt.Sprintf("%s: removed", subPath))
			}
		}
		// Check for keys in v2 not in v1, or changed
		for k, v2Val := range val2 {
			subPath := joinPath(path, k)
			if v1Val, exists := val1[k]; !exists {
				changes = append(changes, fmt.Sprintf("%s: added", subPath))
			} else {
				changes = append(changes, compareSpecs(v1Val, v2Val, subPath, opts)...)
			}
		}

	case []interface{}:
		val2 := v2.([]interface{})
		if opts.IgnoreOrder {
			// Compare as sets
			if !equalSets(val1, val2) {
				changes = append(changes, fmt.Sprintf("%s: changed", path))
			}
		} else {
			// Compare element by element
			if len(val1) != len(val2) {
				changes = append(changes, fmt.Sprintf("%s: length changed (%d -> %d)", path, len(val1), len(val2)))
			} else {
				for i := range val1 {
					subPath := fmt.Sprintf("%s[%d]", path, i)
					changes = append(changes, compareSpecs(val1[i], val2[i], subPath, opts)...)
				}
			}
		}

	default:
		// Scalar comparison
		if !reflect.DeepEqual(v1, v2) {
			if path != "" {
				changes = append(changes, fmt.Sprintf("%s: %v -> %v", path, v1, v2))
			}
		}
	}

	return changes
}

// joinPath joins path components.
func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

// equalSets checks if two slices contain the same elements (ignoring order).
func equalSets(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	// Simple approach: convert to strings and sort
	strA := make([]string, len(a))
	strB := make([]string, len(b))
	for i, v := range a {
		strA[i] = fmt.Sprintf("%v", v)
	}
	for i, v := range b {
		strB[i] = fmt.Sprintf("%v", v)
	}
	sort.Strings(strA)
	sort.Strings(strB)

	for i := range strA {
		if strA[i] != strB[i] {
			return false
		}
	}
	return true
}
