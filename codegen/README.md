# Kubernetes Schema Code Generator

This package provides tools to fetch Kubernetes OpenAPI schemas and generate Go resource type definitions.

## Features

- **Schema Fetching**: Download OpenAPI specs from the Kubernetes GitHub repository
- **Schema Caching**: Cache downloaded schemas locally for faster subsequent runs
- **Schema Parsing**: Parse Kubernetes OpenAPI definitions into structured resource types
- **Code Generation**: Generate Go structs with proper JSON/YAML tags for all Kubernetes resources

## Architecture

```
codegen/
├── fetch/         # Schema fetching and caching
├── parse/         # Schema parsing and type extraction
├── generate/      # Go code generation
└── schemas/       # Cached schema files
```

## Usage

```go
package main

import (
    "context"
    "github.com/lex00/wetwire-k8s-go/codegen/fetch"
    "github.com/lex00/wetwire-k8s-go/codegen/parse"
    "github.com/lex00/wetwire-k8s-go/codegen/generate"
)

func main() {
    ctx := context.Background()

    // 1. Fetch schema
    fetcher := fetch.NewFetcher("schemas")
    schema, _ := fetcher.FetchSchema(ctx, "v1.28.0")

    // 2. Parse resources
    parser := parse.NewParser()
    resources, _ := parser.ParseResourceTypes(schema)

    // 3. Generate code
    generator := generate.NewGenerator("resources")
    _ = generator.GenerateResources(resources)
}
```

## Generated Code Structure

```
resources/
├── apps/v1/
│   ├── deployment.go
│   └── statefulset.go
├── core/v1/
│   ├── pod.go
│   └── service.go
└── ...
```

## Testing

Run all tests:
```bash
go test ./codegen/...
```

Run integration tests:
```bash
go test -v ./codegen/integration_test.go
```
