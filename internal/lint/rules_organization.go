package lint

import (
	"fmt"
	"go/ast"
	"go/token"
)

// MaxResourcesPerFile is the maximum recommended resources per file.
const MaxResourcesPerFile = 20

// RuleWK8401 checks for file size limits (resources per file).
func RuleWK8401() Rule {
	return Rule{
		ID:          "WK8401",
		Name:        "File size limits",
		Description: "Files should not exceed 20 resources",
		Severity:    SeverityWarning,
		Check:       checkWK8401,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8401(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Count K8s resources in the file
	resourceCount := 0

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if any value is a K8s resource type
			for _, value := range valueSpec.Values {
				if isK8sResourceType(value) {
					resourceCount++
				}
			}
		}
	}

	if resourceCount > MaxResourcesPerFile {
		pos := fset.Position(file.Package)
		issues = append(issues, Issue{
			Rule:     "WK8401",
			Message:  fmt.Sprintf("File contains %d resources (max %d), consider splitting into smaller files", resourceCount, MaxResourcesPerFile),
			File:     pos.Filename,
			Line:     1,
			Column:   1,
			Severity: SeverityWarning,
		})
	}

	return issues
}
