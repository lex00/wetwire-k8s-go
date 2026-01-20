// Package crd provides parsing for Kubernetes Custom Resource Definitions.
//
// Unlike the standard Kubernetes OpenAPI schema (Swagger 2.0), CRDs embed
// OpenAPI v3 schemas in their spec.versions[].schema.openAPIV3Schema field.
// This package parses those schemas to enable code generation for CRD types.
package crd

// CRD represents a Kubernetes Custom Resource Definition.
type CRD struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       CRDSpec  `yaml:"spec" json:"spec"`
}

// Metadata contains CRD metadata.
type Metadata struct {
	Name        string            `yaml:"name" json:"name"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// CRDSpec is the spec portion of a CRD.
type CRDSpec struct {
	Group    string       `yaml:"group" json:"group"`
	Names    CRDNames     `yaml:"names" json:"names"`
	Scope    string       `yaml:"scope" json:"scope"`
	Versions []CRDVersion `yaml:"versions" json:"versions"`
}

// CRDNames contains the naming information for a CRD.
type CRDNames struct {
	Kind       string   `yaml:"kind" json:"kind"`
	Plural     string   `yaml:"plural" json:"plural"`
	Singular   string   `yaml:"singular,omitempty" json:"singular,omitempty"`
	ShortNames []string `yaml:"shortNames,omitempty" json:"shortNames,omitempty"`
	Categories []string `yaml:"categories,omitempty" json:"categories,omitempty"`
}

// CRDVersion represents a version entry in a CRD.
type CRDVersion struct {
	Name                     string           `yaml:"name" json:"name"`
	Served                   bool             `yaml:"served" json:"served"`
	Storage                  bool             `yaml:"storage,omitempty" json:"storage,omitempty"`
	Deprecated               bool             `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	DeprecationWarning       string           `yaml:"deprecationWarning,omitempty" json:"deprecationWarning,omitempty"`
	Schema                   *CRDVersionSchema `yaml:"schema,omitempty" json:"schema,omitempty"`
	AdditionalPrinterColumns []PrinterColumn  `yaml:"additionalPrinterColumns,omitempty" json:"additionalPrinterColumns,omitempty"`
}

// CRDVersionSchema wraps the OpenAPI v3 schema.
type CRDVersionSchema struct {
	OpenAPIV3Schema *OpenAPIV3Schema `yaml:"openAPIV3Schema,omitempty" json:"openAPIV3Schema,omitempty"`
}

// PrinterColumn defines an additional column for kubectl get output.
type PrinterColumn struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	JSONPath    string `yaml:"jsonPath" json:"jsonPath"`
	Priority    int    `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// OpenAPIV3Schema represents an OpenAPI v3 schema as embedded in CRDs.
// This is a subset of the full OpenAPI v3 spec, covering what CRDs support.
type OpenAPIV3Schema struct {
	// Type is the JSON Schema type (string, number, integer, boolean, array, object)
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Description provides documentation for this schema
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Properties defines the properties of an object type
	Properties map[string]*OpenAPIV3Schema `yaml:"properties,omitempty" json:"properties,omitempty"`

	// Required lists properties that must be present
	Required []string `yaml:"required,omitempty" json:"required,omitempty"`

	// Items defines the schema for array items
	Items *OpenAPIV3Schema `yaml:"items,omitempty" json:"items,omitempty"`

	// AdditionalProperties defines the schema for map values
	// Can be a bool (true = any type allowed) or a schema
	AdditionalProperties *OpenAPIV3Schema `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`

	// Format provides additional type information (e.g., "int32", "int64", "date-time")
	Format string `yaml:"format,omitempty" json:"format,omitempty"`

	// Enum restricts values to a fixed set
	Enum []interface{} `yaml:"enum,omitempty" json:"enum,omitempty"`

	// Default specifies the default value
	Default interface{} `yaml:"default,omitempty" json:"default,omitempty"`

	// Pattern is a regex pattern for string validation
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"`

	// Minimum is the minimum value for numbers
	Minimum *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`

	// Maximum is the maximum value for numbers
	Maximum *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`

	// MinLength is the minimum length for strings
	MinLength *int64 `yaml:"minLength,omitempty" json:"minLength,omitempty"`

	// MaxLength is the maximum length for strings
	MaxLength *int64 `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`

	// MinItems is the minimum number of array items
	MinItems *int64 `yaml:"minItems,omitempty" json:"minItems,omitempty"`

	// MaxItems is the maximum number of array items
	MaxItems *int64 `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`

	// OneOf specifies that the schema must match exactly one of the sub-schemas
	OneOf []*OpenAPIV3Schema `yaml:"oneOf,omitempty" json:"oneOf,omitempty"`

	// AnyOf specifies that the schema must match at least one of the sub-schemas
	AnyOf []*OpenAPIV3Schema `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`

	// AllOf specifies that the schema must match all of the sub-schemas
	AllOf []*OpenAPIV3Schema `yaml:"allOf,omitempty" json:"allOf,omitempty"`

	// Not specifies that the schema must not match
	Not *OpenAPIV3Schema `yaml:"not,omitempty" json:"not,omitempty"`

	// Nullable indicates the field can be null
	Nullable bool `yaml:"nullable,omitempty" json:"nullable,omitempty"`

	// XKubernetesPreserveUnknownFields allows unknown fields
	XKubernetesPreserveUnknownFields *bool `yaml:"x-kubernetes-preserve-unknown-fields,omitempty" json:"x-kubernetes-preserve-unknown-fields,omitempty"`

	// XKubernetesIntOrString allows the field to be an integer or string
	XKubernetesIntOrString bool `yaml:"x-kubernetes-int-or-string,omitempty" json:"x-kubernetes-int-or-string,omitempty"`

	// XKubernetesEmbeddedResource indicates this is an embedded API resource
	XKubernetesEmbeddedResource bool `yaml:"x-kubernetes-embedded-resource,omitempty" json:"x-kubernetes-embedded-resource,omitempty"`

	// XKubernetesListMapKeys specifies the keys for list map types
	XKubernetesListMapKeys []string `yaml:"x-kubernetes-list-map-keys,omitempty" json:"x-kubernetes-list-map-keys,omitempty"`

	// XKubernetesListType specifies the list type (atomic, set, or map)
	XKubernetesListType string `yaml:"x-kubernetes-list-type,omitempty" json:"x-kubernetes-list-type,omitempty"`

	// XKubernetesMapType specifies the map type (atomic or granular)
	XKubernetesMapType string `yaml:"x-kubernetes-map-type,omitempty" json:"x-kubernetes-map-type,omitempty"`

	// XKubernetesValidations contains CEL validation rules
	XKubernetesValidations []ValidationRule `yaml:"x-kubernetes-validations,omitempty" json:"x-kubernetes-validations,omitempty"`
}

// ValidationRule represents a CEL validation rule.
type ValidationRule struct {
	Rule              string `yaml:"rule" json:"rule"`
	Message           string `yaml:"message,omitempty" json:"message,omitempty"`
	MessageExpression string `yaml:"messageExpression,omitempty" json:"messageExpression,omitempty"`
}
