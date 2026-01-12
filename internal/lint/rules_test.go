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

	t.Run("should have all 26 rules", func(t *testing.T) {
		assert.Len(t, rules, 26, "Expected 26 rules")
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

func TestWK8103_ContainerNameRequired(t *testing.T) {
	rule := RuleWK8103()

	t.Run("should detect containers without name", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8103_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8103" {
				found = true
				assert.Contains(t, issue.Message, "Name", "Expected message about Name field")
			}
		}
		assert.True(t, found, "Expected to find WK8103 violation")
	})

	t.Run("should pass for containers with name", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8103_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8104_PortNameRecommended(t *testing.T) {
	rule := RuleWK8104()

	t.Run("should detect ports without name", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8104_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8104" {
				found = true
				assert.Contains(t, issue.Message, "Name", "Expected message about Name field")
			}
		}
		assert.True(t, found, "Expected to find WK8104 violation")
	})

	t.Run("should pass for ports with name", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8104_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8105_ImagePullPolicyExplicit(t *testing.T) {
	rule := RuleWK8105()

	t.Run("should detect missing ImagePullPolicy", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8105_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8105" {
				found = true
				assert.Contains(t, issue.Message, "ImagePullPolicy", "Expected message about ImagePullPolicy")
			}
		}
		assert.True(t, found, "Expected to find WK8105 violation")
	})

	t.Run("should pass for containers with ImagePullPolicy", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8105_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8203_ReadOnlyRootFilesystem(t *testing.T) {
	rule := RuleWK8203()

	t.Run("should detect missing ReadOnlyRootFilesystem", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8203_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8203" {
				found = true
				assert.Contains(t, issue.Message, "ReadOnlyRootFilesystem", "Expected message about ReadOnlyRootFilesystem")
			}
		}
		assert.True(t, found, "Expected to find WK8203 violation")
	})

	t.Run("should pass for containers with ReadOnlyRootFilesystem", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8203_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8204_RunAsNonRoot(t *testing.T) {
	rule := RuleWK8204()

	t.Run("should detect missing RunAsNonRoot", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8204_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8204" {
				found = true
				assert.Contains(t, issue.Message, "RunAsNonRoot", "Expected message about RunAsNonRoot")
			}
		}
		assert.True(t, found, "Expected to find WK8204 violation")
	})

	t.Run("should pass for containers with RunAsNonRoot", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8204_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8205_DropCapabilities(t *testing.T) {
	rule := RuleWK8205()

	t.Run("should detect missing drop capabilities", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8205_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8205" {
				found = true
				assert.Contains(t, issue.Message, "capabilities", "Expected message about capabilities")
			}
		}
		assert.True(t, found, "Expected to find WK8205 violation")
	})

	t.Run("should pass for containers dropping capabilities", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8205_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8207_NoHostNetwork(t *testing.T) {
	rule := RuleWK8207()

	t.Run("should detect HostNetwork usage", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8207_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8207" {
				found = true
				assert.Contains(t, issue.Message, "HostNetwork", "Expected message about HostNetwork")
			}
		}
		assert.True(t, found, "Expected to find WK8207 violation")
	})

	t.Run("should pass for pods without HostNetwork", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8207_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8208_NoHostPID(t *testing.T) {
	rule := RuleWK8208()

	t.Run("should detect HostPID usage", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8208_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8208" {
				found = true
				assert.Contains(t, issue.Message, "HostPID", "Expected message about HostPID")
			}
		}
		assert.True(t, found, "Expected to find WK8208 violation")
	})

	t.Run("should pass for pods without HostPID", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8208_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8209_NoHostIPC(t *testing.T) {
	rule := RuleWK8209()

	t.Run("should detect HostIPC usage", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8209_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8209" {
				found = true
				assert.Contains(t, issue.Message, "HostIPC", "Expected message about HostIPC")
			}
		}
		assert.True(t, found, "Expected to find WK8209 violation")
	})

	t.Run("should pass for pods without HostIPC", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8209_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8302_ReplicasMinimum(t *testing.T) {
	rule := RuleWK8302()

	t.Run("should detect single replica deployments", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8302_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8302" {
				found = true
				assert.Contains(t, issue.Message, "replicas", "Expected message about replicas")
			}
		}
		assert.True(t, found, "Expected to find WK8302 violation")
	})

	t.Run("should pass for multi-replica deployments", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8302_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8303_PodDisruptionBudget(t *testing.T) {
	rule := RuleWK8303()

	t.Run("should detect HA deployments without PDB", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8303_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8303" {
				found = true
				assert.Contains(t, issue.Message, "PodDisruptionBudget", "Expected message about PodDisruptionBudget")
			}
		}
		assert.True(t, found, "Expected to find WK8303 violation")
	})

	t.Run("should pass for HA deployments with PDB", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8303_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8304_AntiAffinityRecommended(t *testing.T) {
	rule := RuleWK8304()

	t.Run("should detect HA deployments without anti-affinity", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8304_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8304" {
				found = true
				assert.Contains(t, issue.Message, "anti-affinity", "Expected message about anti-affinity")
			}
		}
		assert.True(t, found, "Expected to find WK8304 violation")
	})

	t.Run("should pass for HA deployments with anti-affinity", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8304_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}

func TestWK8401_FileSizeLimits(t *testing.T) {
	rule := RuleWK8401()

	t.Run("should detect file with too many resources", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8401_bad.go")
		issues := rule.Check(file, fset)

		assert.NotEmpty(t, issues, "Expected to find issues in bad file")

		found := false
		for _, issue := range issues {
			if issue.Rule == "WK8401" {
				found = true
				assert.Contains(t, issue.Message, "resources", "Expected message about resource count")
			}
		}
		assert.True(t, found, "Expected to find WK8401 violation")
	})

	t.Run("should pass for file with few resources", func(t *testing.T) {
		fset, file := parseTestFile(t, "testdata/wk8401_good.go")
		issues := rule.Check(file, fset)

		assert.Empty(t, issues, "Expected no issues in good file")
	})
}
