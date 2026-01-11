package build

import (
	"fmt"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
)

// TopologicalSort sorts resources in topological order using Kahn's algorithm.
// Resources with no dependencies come first, followed by resources that depend on them.
// Returns an error if a cycle is detected.
func TopologicalSort(resources []discover.Resource) ([]discover.Resource, error) {
	// Handle empty input
	if len(resources) == 0 {
		return []discover.Resource{}, nil
	}

	// First, detect cycles
	if err := DetectCycles(resources); err != nil {
		return nil, err
	}

	// Build a map of resources by name for quick lookup
	resourceMap := make(map[string]discover.Resource)
	for _, r := range resources {
		resourceMap[r.Name] = r
	}

	// Calculate in-degree (number of incoming edges) for each resource
	inDegree := make(map[string]int)
	for _, r := range resources {
		if _, exists := inDegree[r.Name]; !exists {
			inDegree[r.Name] = 0
		}
		for range r.Dependencies {
			inDegree[r.Name]++
		}
	}

	// Initialize queue with all resources that have no dependencies (in-degree = 0)
	queue := []string{}
	for _, r := range resources {
		if inDegree[r.Name] == 0 {
			queue = append(queue, r.Name)
		}
	}

	// Process resources in topological order
	var sorted []discover.Resource
	for len(queue) > 0 {
		// Dequeue a resource with no dependencies
		current := queue[0]
		queue = queue[1:]

		// Add it to the sorted list
		sorted = append(sorted, resourceMap[current])

		// For each resource that depends on the current resource,
		// decrease its in-degree
		for _, r := range resources {
			for _, dep := range r.Dependencies {
				if dep == current {
					inDegree[r.Name]--
					// If in-degree becomes 0, add to queue
					if inDegree[r.Name] == 0 {
						queue = append(queue, r.Name)
					}
				}
			}
		}
	}

	// If we couldn't sort all resources, there must be a cycle
	// (This should not happen since we detect cycles first, but let's be safe)
	if len(sorted) != len(resources) {
		return nil, fmt.Errorf("topological sort failed: cycle detected (sorted %d of %d resources)", len(sorted), len(resources))
	}

	return sorted, nil
}
