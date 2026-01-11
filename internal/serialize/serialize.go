package serialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Serialize converts a Go struct (Kubernetes resource) to a map[string]interface{}.
// It handles field name conversion from Go naming to Kubernetes camelCase conventions,
// recursively processes nested structs, and omits zero values.
func Serialize(resource interface{}) (map[string]interface{}, error) {
	if resource == nil {
		return nil, errors.New("resource cannot be nil")
	}

	// Convert to JSON first (which handles k8s types correctly with json tags),
	// then parse back to map[string]interface{}
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	// Clean up the result by removing zero values
	result = cleanZeroValues(result)

	return result, nil
}

// ToYAML converts a Kubernetes resource to YAML format.
func ToYAML(resource interface{}) ([]byte, error) {
	data, err := Serialize(resource)
	if err != nil {
		return nil, err
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return yamlBytes, nil
}

// ToJSON converts a Kubernetes resource to JSON format.
func ToJSON(resource interface{}) ([]byte, error) {
	data, err := Serialize(resource)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return jsonBytes, nil
}

// ToMultiYAML converts multiple Kubernetes resources to a multi-document YAML format,
// separated by "---" delimiters.
func ToMultiYAML(resources []interface{}) ([]byte, error) {
	if len(resources) == 0 {
		return []byte{}, nil
	}

	var documents []string
	for i, resource := range resources {
		yamlBytes, err := ToYAML(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize resource %d: %w", i, err)
		}

		// Trim trailing whitespace
		doc := strings.TrimSpace(string(yamlBytes))
		documents = append(documents, doc)
	}

	// Join with document separator
	result := strings.Join(documents, "\n---\n")
	return []byte(result), nil
}

// cleanZeroValues recursively removes zero values from a map
func cleanZeroValues(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if isZeroValue(value) {
			continue
		}

		// Recursively clean nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			cleaned := cleanZeroValues(nestedMap)
			if len(cleaned) > 0 {
				result[key] = cleaned
			}
			continue
		}

		// Recursively clean slices
		if slice, ok := value.([]interface{}); ok {
			cleanedSlice := cleanSlice(slice)
			if len(cleanedSlice) > 0 {
				result[key] = cleanedSlice
			}
			continue
		}

		result[key] = value
	}

	return result
}

// cleanSlice recursively cleans zero values from a slice
func cleanSlice(slice []interface{}) []interface{} {
	var result []interface{}

	for _, item := range slice {
		if isZeroValue(item) {
			continue
		}

		// Recursively clean nested maps
		if nestedMap, ok := item.(map[string]interface{}); ok {
			cleaned := cleanZeroValues(nestedMap)
			if len(cleaned) > 0 {
				result = append(result, cleaned)
			}
			continue
		}

		// Recursively clean nested slices
		if nestedSlice, ok := item.([]interface{}); ok {
			cleaned := cleanSlice(nestedSlice)
			if len(cleaned) > 0 {
				result = append(result, cleaned)
			}
			continue
		}

		result = append(result, item)
	}

	return result
}

// isZeroValue checks if a value is a zero value that should be omitted
func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}
