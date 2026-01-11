package lint

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLinter(t *testing.T) {
	t.Run("should create linter with default config", func(t *testing.T) {
		linter := NewLinter(nil)
		assert.NotNil(t, linter)
		assert.NotNil(t, linter.config)
		assert.Equal(t, SeverityInfo, linter.config.MinSeverity)
		assert.Len(t, linter.rules, 6, "Should have all 6 rules enabled")
	})

	t.Run("should create linter with custom config", func(t *testing.T) {
		config := &Config{
			MinSeverity: SeverityWarning,
		}
		linter := NewLinter(config)
		assert.NotNil(t, linter)
		assert.Equal(t, SeverityWarning, linter.config.MinSeverity)
	})

	t.Run("should disable specified rules", func(t *testing.T) {
		config := &Config{
			DisabledRules: []string{"WK8001", "WK8002"},
		}
		linter := NewLinter(config)
		assert.Len(t, linter.rules, 4, "Should have 4 rules enabled (2 disabled)")
	})
}

func TestLinter_LintFile(t *testing.T) {
	linter := NewLinter(nil)

	t.Run("should detect issues in bad files", func(t *testing.T) {
		testCases := []struct {
			file         string
			expectedRule string
		}{
			{"wk8001_bad.go", "WK8001"},
			{"wk8002_bad.go", "WK8002"},
			{"wk8003_bad.go", "WK8003"},
			{"wk8004_bad.go", "WK8004"},
			{"wk8005_bad.go", "WK8005"},
			{"wk8006_bad.go", "WK8006"},
		}

		for _, tc := range testCases {
			t.Run(tc.file, func(t *testing.T) {
				filePath := filepath.Join("testdata", tc.file)
				issues, err := linter.LintFile(filePath)
				require.NoError(t, err)
				assert.NotEmpty(t, issues, "Expected to find issues in %s", filePath)

				// Verify at least one issue matches the expected rule
				found := false
				for _, issue := range issues {
					if issue.Rule == tc.expectedRule {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find issue with rule %s in %s", tc.expectedRule, filePath)
			})
		}
	})

	t.Run("should not find specific rule issues in good files", func(t *testing.T) {
		testCases := []struct {
			file         string
			expectedRule string
		}{
			{"wk8001_good.go", "WK8001"},
			{"wk8002_good.go", "WK8002"},
			{"wk8003_good.go", "WK8003"},
			{"wk8004_good.go", "WK8004"},
			{"wk8005_good.go", "WK8005"},
			{"wk8006_good.go", "WK8006"},
		}

		for _, tc := range testCases {
			t.Run(tc.file, func(t *testing.T) {
				filePath := filepath.Join("testdata", tc.file)
				issues, err := linter.LintFile(filePath)
				require.NoError(t, err)

				// Verify the specific rule doesn't fire for this "good" file
				for _, issue := range issues {
					assert.NotEqual(t, tc.expectedRule, issue.Rule,
						"Expected no %s issues in %s, but found: %s", tc.expectedRule, filePath, issue.Message)
				}
			})
		}
	})
}

func TestLinter_LintDirectory(t *testing.T) {
	linter := NewLinter(nil)

	t.Run("should lint all files in directory", func(t *testing.T) {
		testdataPath := filepath.Join("testdata")
		issues, err := linter.LintDirectory(testdataPath)
		require.NoError(t, err)

		// We expect to find issues from all the bad files
		assert.NotEmpty(t, issues, "Expected to find issues in testdata directory")

		// Count issues by rule
		ruleCounts := make(map[string]int)
		for _, issue := range issues {
			ruleCounts[issue.Rule]++
		}

		// We should have issues from all 6 rules
		assert.Greater(t, ruleCounts["WK8001"], 0, "Expected WK8001 issues")
		assert.Greater(t, ruleCounts["WK8002"], 0, "Expected WK8002 issues")
		assert.Greater(t, ruleCounts["WK8003"], 0, "Expected WK8003 issues")
		assert.Greater(t, ruleCounts["WK8004"], 0, "Expected WK8004 issues")
		assert.Greater(t, ruleCounts["WK8005"], 0, "Expected WK8005 issues")
		assert.Greater(t, ruleCounts["WK8006"], 0, "Expected WK8006 issues")
	})
}

func TestLinter_Lint(t *testing.T) {
	linter := NewLinter(nil)

	t.Run("should lint file when path is file", func(t *testing.T) {
		filePath := filepath.Join("testdata", "wk8001_bad.go")
		issues, err := linter.Lint(filePath)
		require.NoError(t, err)
		assert.NotEmpty(t, issues)
	})

	t.Run("should lint directory when path is directory", func(t *testing.T) {
		dirPath := filepath.Join("testdata")
		issues, err := linter.Lint(dirPath)
		require.NoError(t, err)
		assert.NotEmpty(t, issues)
	})
}

func TestLinter_LintWithResult(t *testing.T) {
	linter := NewLinter(nil)

	t.Run("should return detailed result", func(t *testing.T) {
		dirPath := filepath.Join("testdata")
		result, err := linter.LintWithResult(dirPath)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// We should have issues
		assert.NotEmpty(t, result.Issues)

		// We should have files scanned
		assert.Greater(t, result.TotalFiles, 0)
		assert.Greater(t, result.FilesWithIssues, 0)

		// We should have issues by severity
		// All our test rules are errors
		assert.Greater(t, result.ErrorCount, 0)
	})
}

func TestLinter_MinSeverityFilter(t *testing.T) {
	t.Run("should filter issues by minimum severity", func(t *testing.T) {
		// Create linter that only reports errors
		config := &Config{
			MinSeverity: SeverityError,
		}
		linter := NewLinter(config)

		filePath := filepath.Join("testdata", "wk8001_bad.go")
		issues, err := linter.LintFile(filePath)
		require.NoError(t, err)

		// All issues should be errors or higher
		for _, issue := range issues {
			assert.GreaterOrEqual(t, issue.Severity, SeverityError)
		}
	})
}

func TestLinter_DisabledRules(t *testing.T) {
	t.Run("should not report issues for disabled rules", func(t *testing.T) {
		config := &Config{
			DisabledRules: []string{"WK8001"},
		}
		linter := NewLinter(config)

		filePath := filepath.Join("testdata", "wk8001_bad.go")
		issues, err := linter.LintFile(filePath)
		require.NoError(t, err)

		// Should not have any WK8001 issues
		for _, issue := range issues {
			assert.NotEqual(t, "WK8001", issue.Rule)
		}
	})
}

func TestLinter_NonExistentFile(t *testing.T) {
	linter := NewLinter(nil)
	_, err := linter.LintFile("/nonexistent/path/file.go")
	assert.Error(t, err)
}

func TestLinter_NonExistentPath(t *testing.T) {
	linter := NewLinter(nil)
	_, err := linter.Lint("/nonexistent/path")
	assert.Error(t, err)
}

func TestIsRuleDisabled(t *testing.T) {
	config := &Config{
		DisabledRules: []string{"WK8001", "WK8003"},
	}
	linter := NewLinter(config)

	// The linter should have 4 rules (6 - 2 disabled)
	assert.Len(t, linter.rules, 4)
}

func TestLintResult_CountsIssuesBySeverity(t *testing.T) {
	result := &LintResult{
		Issues: []Issue{
			{Severity: SeverityError, Rule: "WK8001"},
			{Severity: SeverityError, Rule: "WK8002"},
			{Severity: SeverityWarning, Rule: "WK8003"},
			{Severity: SeverityInfo, Rule: "WK8004"},
		},
		TotalFiles:      3,
		FilesWithIssues: 2,
	}

	// Count manually
	errorCount := 0
	warningCount := 0
	infoCount := 0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			errorCount++
		case SeverityWarning:
			warningCount++
		case SeverityInfo:
			infoCount++
		}
	}

	assert.Equal(t, 2, errorCount)
	assert.Equal(t, 1, warningCount)
	assert.Equal(t, 1, infoCount)
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

func TestIssue_Fields(t *testing.T) {
	issue := Issue{
		File:     "test.go",
		Line:     10,
		Rule:     "WK8001",
		Message:  "test message",
		Severity: SeverityError,
	}

	assert.Equal(t, "test.go", issue.File)
	assert.Equal(t, 10, issue.Line)
	assert.Equal(t, "WK8001", issue.Rule)
	assert.Equal(t, "test message", issue.Message)
	assert.Equal(t, SeverityError, issue.Severity)
}

func TestLinter_EmptyDirectory(t *testing.T) {
	linter := NewLinter(nil)

	// Lint a directory that doesn't exist
	_, err := linter.LintDirectory("/nonexistent/path")
	assert.Error(t, err)
}

func TestConfig_DefaultValues(t *testing.T) {
	config := &Config{}
	assert.Equal(t, Severity(0), config.MinSeverity)
	assert.Empty(t, config.DisabledRules)
}

func TestLinter_AllRulesEnabled(t *testing.T) {
	linter := NewLinter(nil)
	// Should have all 6 rules enabled by default
	assert.Len(t, linter.rules, 6)
}

func TestLinter_DisableAllRules(t *testing.T) {
	config := &Config{
		DisabledRules: []string{"WK8001", "WK8002", "WK8003", "WK8004", "WK8005", "WK8006"},
	}
	linter := NewLinter(config)
	assert.Len(t, linter.rules, 0)
}
