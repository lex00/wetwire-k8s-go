package build

import (
	"fmt"
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
)

// ValidateReferences checks that all resource dependencies reference existing resources.
// It returns an error if any resource references a non-existent resource or itself.
func ValidateReferences(resources []discover.Resource) error {
	// Build a set of all resource names
	resourceNames := make(map[string]bool)
	for _, r := range resources {
		resourceNames[r.Name] = true
	}

	// Check each resource's dependencies
	var errors []string
	for _, r := range resources {
		for _, dep := range r.Dependencies {
			// Check for self-reference
			if dep == r.Name {
				errors = append(errors, fmt.Sprintf("resource %q references itself", r.Name))
				continue
			}

			// Check if dependency exists
			if !resourceNames[dep] {
				errors = append(errors, fmt.Sprintf("resource %q references non-existent resource %q", r.Name, dep))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// DetectCycles detects circular dependencies in the resource graph.
// It returns an error if any cycles are found.
func DetectCycles(resources []discover.Resource) error {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, r := range resources {
		graph[r.Name] = r.Dependencies
	}

	// Track visited nodes and nodes in current path
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// Helper function to detect cycle using DFS
	var hasCycle func(node string, path []string) (bool, []string)
	hasCycle = func(node string, path []string) (bool, []string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		// Visit all dependencies
		for _, dep := range graph[node] {
			if !visited[dep] {
				if cycle, cyclePath := hasCycle(dep, path); cycle {
					return true, cyclePath
				}
			} else if recStack[dep] {
				// Found a cycle - build the cycle path
				cycleStart := -1
				for i, n := range path {
					if n == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					return true, append(path[cycleStart:], dep)
				}
				return true, append(path, dep)
			}
		}

		recStack[node] = false
		return false, nil
	}

	// Check each resource for cycles
	for _, r := range resources {
		if !visited[r.Name] {
			if cycle, cyclePath := hasCycle(r.Name, []string{}); cycle {
				return fmt.Errorf("cycle detected: %s", strings.Join(cyclePath, " -> "))
			}
		}
	}

	return nil
}
