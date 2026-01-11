package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to parse test files
func parseTestFile(t *testing.T, filename string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	require.NoError(t, err, "Failed to parse test file")
	return fset, file
}

func TestWK8001_TopLevelResourceDeclarations(t *testing.T) {
	rule := RuleWK8001()

	t.Run("should detect non-top-level resources", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8001_bad.go")
		issues := rule.Check(file, fset)

		// Should detect the variable assigned from function call
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8001" {
				found = true
				assert.Contains(t, issue.Message, "function call", "Expected message about function call")
			}
		}
		assert.True(t, found, "Expected to find WK8001 violation")
	})

	t.Run("should pass for top-level declarations", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8001_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8002_AvoidDeeplyNestedStructures(t *testing.T) {
	rule := RuleWK8002()

	t.Run("should detect deeply nested structures", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8002_bad.go")
		issues := rule.Check(file, fset)

		// Should detect nesting depth > 5
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8002" {
				found = true
				assert.Contains(t, issue.Message, "depth", "Expected message about nesting depth")
			}
		}
		assert.True(t, found, "Expected to find WK8002 violation")
	})

	t.Run("should pass for flat structures", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8002_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8003_NoDuplicateResourceNames(t *testing.T) {
	rule := RuleWK8003()

	t.Run("should detect duplicate names in same namespace", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8003_bad.go")
		issues := rule.Check(file, fset)

		// Should detect duplicate names
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8003" {
				found = true
				assert.Contains(t, issue.Message, "Duplicate", "Expected message about duplicate")
			}
		}
		assert.True(t, found, "Expected to find WK8003 violation")
	})

	t.Run("should pass for unique names or different namespaces", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8003_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8004_CircularDependencyDetection(t *testing.T) {
	rule := RuleWK8004()

	t.Run("should detect circular dependencies", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8004_bad.go")
		issues := rule.Check(file, fset)

		// Should detect circular dependencies
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8004" {
				found = true
				assert.Contains(t, issue.Message, "Circular", "Expected message about circular dependency")
			}
		}
		assert.True(t, found, "Expected to find WK8004 violation")
	})

	t.Run("should pass for linear dependencies", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8004_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8005_FlagHardcodedSecrets(t *testing.T) {
	rule := RuleWK8005()

	t.Run("should detect hardcoded secrets", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8005_bad.go")
		issues := rule.Check(file, fset)

		// Should detect hardcoded secrets
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8005" {
				found = true
				assert.Regexp(t, "(?i)(secret|password|key|token)", issue.Message, "Expected message about secrets")
			}
		}
		assert.True(t, found, "Expected to find WK8005 violation")
	})

	t.Run("should pass for secret references and non-sensitive values", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8005_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8006_FlagLatestImageTags(t *testing.T) {
	rule := RuleWK8006()

	t.Run("should detect :latest tags", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8006_bad.go")
		issues := rule.Check(file, fset)

		// Should detect :latest tags
		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		// Check that we found the specific violation
		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8006" {
				found = true
				assert.Contains(t, issue.Message, "latest", "Expected message about :latest tag")
			}
		}
		assert.True(t, found, "Expected to find WK8006 violation")
	})

	t.Run("should pass for specific version tags", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8006_good.go")
		issues := rule.Check(file, fset)

		// Should not find any issues
		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestAllRules(t *testing.T) {
	rules := AllRules()

	t.Run("should have all 6 rules", func(t *testing.T) {
		assert.Len(t, rules, 6, "Expected 6 rules")
	})

	t.Run("all rules should have required fields", func(t *testing.T) {
		for _, rule := range rules {
			assert.NotEmpty(t, rule.ID, "Rule ID should not be empty")
			assert.NotEmpty(t, rule.Name, "Rule name should not be empty")
			assert.NotEmpty(t, rule.Description, "Rule description should not be empty")
			assert.NotNil(t, rule.Check, "Rule Check function should not be nil")
		}
	})

	t.Run("rule IDs should be unique", func(t *testing.T) {
		ids := make(map[string]bool)
		for _, rule := range rules {
			assert.False(t, ids[rule.ID], "Duplicate rule ID: %s", rule.ID)
			ids[rule.ID] = true
		}
	})

	t.Run("rule IDs should follow WK8xxx format", func(t *testing.T) {
		for _, rule := range rules {
			assert.Regexp(t, `^WK8\d{3}$`, rule.ID, "Rule ID should match WK8xxx format")
		}
	})
}
