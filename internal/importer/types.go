package importer

// Options configures the import operation.
type Options struct {
	PackageName string
	VarPrefix   string
	Optimize    bool
}

// DefaultOptions returns the default import options.
func DefaultOptions() Options {
	return Options{
		PackageName: "main",
		VarPrefix:   "",
		Optimize:    true,
	}
}

// Result contains the output of the import operation.
type Result struct {
	GoCode        string
	ResourceCount int
	Warnings      []string
}

// ResourceInfo contains parsed information about a Kubernetes resource.
type ResourceInfo struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	RawData    map[string]interface{}
}
