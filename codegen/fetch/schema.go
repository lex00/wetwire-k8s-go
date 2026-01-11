package fetch

// Schema represents a Kubernetes OpenAPI/Swagger schema.
type Schema struct {
	Swagger     string                 `json:"swagger"`
	Info        Info                   `json:"info"`
	Paths       map[string]interface{} `json:"paths,omitempty"`
	Definitions map[string]Definition  `json:"definitions"`
}

// Info contains metadata about the API.
type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// Definition represents a schema definition for a Kubernetes type.
type Definition struct {
	Type                 string                `json:"type"`
	Description          string                `json:"description,omitempty"`
	Properties           map[string]Property   `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	XKubernetesGroupVersionKind []GroupVersionKind `json:"x-kubernetes-group-version-kind,omitempty"`
}

// Property represents a property in a schema definition.
type Property struct {
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Ref         string                 `json:"$ref,omitempty"`
	Items       *Property              `json:"items,omitempty"`
	AdditionalProperties *Property     `json:"additionalProperties,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
}

// GroupVersionKind identifies a Kubernetes resource type.
type GroupVersionKind struct {
	Group   string `json:"group"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
}
