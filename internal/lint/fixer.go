package lint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// FixResult represents the result of a fix operation.
type FixResult struct {
	File         string // File that was fixed
	Rule         string // Rule that was applied
	Fixed        bool   // Whether the fix was applied
	Error        error  // Error if fix failed
	Description  string // Description of what was fixed
}

// Fixer handles auto-fix operations for lint issues.
type Fixer struct {
	config *Config
}

// NewFixer creates a new Fixer with the given configuration.
func NewFixer(config *Config) *Fixer {
	if config == nil {
		config = &Config{
			MinSeverity: SeverityInfo, // Include all severities for fixing
		}
	}
	return &Fixer{config: config}
}

// FixFile attempts to fix all fixable issues in a file.
// Returns the list of fixes that were applied.
func (f *Fixer) FixFile(filePath string) ([]FixResult, error) {
	var results []FixResult

	// Read the original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Track if any fixes were made
	modified := false

	// Apply WK8105 fixes (ImagePullPolicy)
	fixResults, changed := f.fixWK8105(file, fset, filePath)
	results = append(results, fixResults...)
	if changed {
		modified = true
	}

	// Apply WK8002 fixes (deeply nested structures)
	fixResults, changed = f.fixWK8002(file, fset, filePath)
	results = append(results, fixResults...)
	if changed {
		modified = true
	}

	// Write the modified file if any fixes were made
	if modified {
		var buf bytes.Buffer
		cfg := printer.Config{
			Mode:     printer.UseSpaces | printer.TabIndent,
			Tabwidth: 8,
		}
		if err := cfg.Fprint(&buf, fset, file); err != nil {
			return results, fmt.Errorf("failed to format file %s: %w", filePath, err)
		}

		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return results, fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return results, nil
}

// fixWK8105 fixes missing ImagePullPolicy on containers.
// Sets ImagePullPolicy to "Always" for :latest or untagged images,
// "IfNotPresent" for tagged images.
func (f *Fixer) fixWK8105(file *ast.File, fset *token.FileSet, filePath string) ([]FixResult, bool) {
	var results []FixResult
	modified := false

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if this is a Container struct
		if !isContainerType(compLit) {
			return true
		}

		// Check if ImagePullPolicy is already set
		hasImagePullPolicy := false
		var imageValue string
		var imageField *ast.KeyValueExpr

		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if key.Name == "ImagePullPolicy" {
				hasImagePullPolicy = true
			} else if key.Name == "Image" {
				imageField = kv
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					imageValue = strings.Trim(lit.Value, `"`)
				}
			}
		}

		// If ImagePullPolicy is not set and we have an image, add it
		if !hasImagePullPolicy && imageValue != "" {
			// Determine the correct policy based on image tag
			policy := determineImagePullPolicy(imageValue)

			// Create new field for ImagePullPolicy
			newField := &ast.KeyValueExpr{
				Key:   ast.NewIdent("ImagePullPolicy"),
				Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, policy)},
			}

			// Insert after the Image field if we found it
			if imageField != nil {
				// Find the position of the Image field and insert after it
				newElts := make([]ast.Expr, 0, len(compLit.Elts)+1)
				for _, elt := range compLit.Elts {
					newElts = append(newElts, elt)
					if elt == imageField {
						newElts = append(newElts, newField)
					}
				}
				compLit.Elts = newElts
			} else {
				// Just append if Image field not found
				compLit.Elts = append(compLit.Elts, newField)
			}

			modified = true
			pos := fset.Position(compLit.Pos())
			results = append(results, FixResult{
				File:        filePath,
				Rule:        "WK8105",
				Fixed:       true,
				Description: fmt.Sprintf("Added ImagePullPolicy: %q for image %q at line %d", policy, imageValue, pos.Line),
			})
		}

		return true
	})

	return results, modified
}

// determineImagePullPolicy returns the appropriate ImagePullPolicy based on the image tag.
func determineImagePullPolicy(image string) string {
	// Check if image uses :latest or has no tag
	if strings.HasSuffix(image, ":latest") {
		return "Always"
	}

	// Check if image has no tag (defaults to :latest)
	if !strings.Contains(image, ":") && !strings.Contains(image, "@") {
		return "Always"
	}

	// Tagged image - use IfNotPresent
	return "IfNotPresent"
}

// fixWK8002 fixes deeply nested structures by extracting them to variables.
// This is a more complex fix that extracts nested composite literals.
func (f *Fixer) fixWK8002(file *ast.File, fset *token.FileSet, filePath string) ([]FixResult, bool) {
	var results []FixResult
	modified := false

	// Find all top-level variable declarations with deep nesting
	var newDecls []ast.Decl
	var extractedVars []*ast.GenDecl

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			newDecls = append(newDecls, decl)
			continue
		}

		// Check each spec in the declaration
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				if i < len(valueSpec.Values) {
					depth := calculateNestingDepth(valueSpec.Values[i])
					if depth > 5 {
						// Extract nested structures
						extracted, newValue := f.extractNestedStructures(valueSpec.Values[i], name.Name, fset)
						if len(extracted) > 0 {
							extractedVars = append(extractedVars, extracted...)
							valueSpec.Values[i] = newValue
							modified = true

							pos := fset.Position(name.Pos())
							results = append(results, FixResult{
								File:        filePath,
								Rule:        "WK8002",
								Fixed:       true,
								Description: fmt.Sprintf("Extracted %d nested structure(s) from %s at line %d", len(extracted), name.Name, pos.Line),
							})
						}
					}
				}
			}
		}

		newDecls = append(newDecls, decl)
	}

	// Insert extracted variables before the declarations that use them
	if len(extractedVars) > 0 {
		// Find the first var declaration and insert before it
		var finalDecls []ast.Decl
		insertedExtracted := false

		for _, decl := range newDecls {
			if !insertedExtracted {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
					// Insert extracted vars before this declaration
					for _, extracted := range extractedVars {
						finalDecls = append(finalDecls, extracted)
					}
					insertedExtracted = true
				}
			}
			finalDecls = append(finalDecls, decl)
		}

		// If we haven't inserted yet (no var decls found), append at end
		if !insertedExtracted {
			for _, extracted := range extractedVars {
				finalDecls = append(finalDecls, extracted)
			}
		}

		file.Decls = finalDecls
	}

	return results, modified
}

// extractNestedStructures extracts deeply nested composite literals to separate variables.
// Returns the extracted declarations and the modified expression.
func (f *Fixer) extractNestedStructures(expr ast.Expr, parentName string, fset *token.FileSet) ([]*ast.GenDecl, ast.Expr) {
	var extracted []*ast.GenDecl
	counter := 0

	var extractRecursive func(e ast.Expr, depth int, fieldName string) ast.Expr
	extractRecursive = func(e ast.Expr, depth int, fieldName string) ast.Expr {
		switch node := e.(type) {
		case *ast.CompositeLit:
			// Process children first
			for i, elt := range node.Elts {
				switch elem := elt.(type) {
				case *ast.KeyValueExpr:
					keyName := ""
					if ident, ok := elem.Key.(*ast.Ident); ok {
						keyName = ident.Name
					}
					elem.Value = extractRecursive(elem.Value, depth+1, keyName)
					node.Elts[i] = elem
				default:
					node.Elts[i] = extractRecursive(elt, depth+1, "")
				}
			}

			// If we're at depth 4 or more and have nested structures, extract
			if depth >= 4 && len(node.Elts) > 0 {
				counter++
				varName := fmt.Sprintf("%s%s%d", parentName, fieldName, counter)
				if fieldName == "" {
					varName = fmt.Sprintf("%sNested%d", parentName, counter)
				}

				// Create a new variable declaration for this structure
				newDecl := &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names:  []*ast.Ident{ast.NewIdent(varName)},
							Values: []ast.Expr{node},
						},
					},
				}
				extracted = append(extracted, newDecl)

				// Return a reference to the new variable
				return ast.NewIdent(varName)
			}

			return node

		case *ast.UnaryExpr:
			node.X = extractRecursive(node.X, depth, fieldName)
			return node

		default:
			return e
		}
	}

	result := extractRecursive(expr, 0, "")
	return extracted, result
}

// FixDirectory attempts to fix all fixable issues in all Go files in a directory.
func (f *Fixer) FixDirectory(dir string) ([]FixResult, error) {
	var allResults []FixResult

	// Get list of Go files
	linter := NewLinter(f.config)
	result, err := linter.LintWithResult(dir)
	if err != nil {
		return nil, err
	}

	// Group issues by file
	fileIssues := make(map[string][]Issue)
	for _, issue := range result.Issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	// Fix each file that has fixable issues
	for filePath, issues := range fileIssues {
		// Check if file has fixable issues
		hasFixable := false
		for _, issue := range issues {
			if isFixableRule(issue.Rule) {
				hasFixable = true
				break
			}
		}

		if hasFixable {
			results, err := f.FixFile(filePath)
			if err != nil {
				allResults = append(allResults, FixResult{
					File:  filePath,
					Fixed: false,
					Error: err,
				})
				continue
			}
			allResults = append(allResults, results...)
		}
	}

	return allResults, nil
}

// isFixableRule returns true if the rule supports auto-fix.
func isFixableRule(ruleID string) bool {
	fixableRules := map[string]bool{
		"WK8105": true, // ImagePullPolicy
		"WK8002": true, // Deeply nested structures
		// WK8006 is NOT fixable - it just warns about :latest, user must choose version
	}
	return fixableRules[ruleID]
}

// FixableRules returns a list of rule IDs that support auto-fix.
func FixableRules() []string {
	return []string{"WK8002", "WK8105"}
}
