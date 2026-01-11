package discover

// Resource represents a discovered Kubernetes resource in the source code.
type Resource struct {
	Name         string   // Variable name
	Type         string   // e.g., "apps/v1.Deployment"
	File         string   // Source file path
	Line         int      // Line number
	Dependencies []string // Referenced resource names
}
