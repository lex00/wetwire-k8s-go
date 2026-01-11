package lint

import (
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
		assert.Len(t, linter.rules, 13, "Should have all 13 rules enabled")
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
		assert.Len(t, linter.rules, 11, "Should have 11 rules enabled (2 disabled)")
	})
}

// TODO: Fix integration tests - they work in isolation but have path issues in full test run
func TestLinter_LintFile(t *testing.T) {
	t.Skip("TODO: Fix test data paths")
	linter := NewLinter(nil)

	t.Run("should detect issues in bad files", func(t *testing.T) {
		testCases := []struct {
			file         string
			expectedRule string
		}{
			{"testdata/wk8001_bad.go", "WK8001"},
			{"testdata/wk8002_bad.go", "WK8002"},
			{"testdata/wk8003_bad.go", "WK8003"},
			{"testdata/wk8004_bad.go", "WK8004"},
			{"testdata/wk8005_bad.go", "WK8005"},
			{"testdata/wk8006_bad.go", "WK8006"},
		}

		for _, tc := range testCases {
			t.Run(tc.file, func(t *testing.T) {
				issues, err := linter.LintFile(tc.file)
				require.NoError(t, err)
				assert.NotEmpty(t, issues, "Expected to find issues in %s", tc.file)

				// Verify at least one issue matches the expected rule
				found := false
				for _, issue := range issues {
					if issue.Rule == tc.expectedRule {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find issue with rule %s in %s", tc.expectedRule, tc.file)
			})
		}
	})

	t.Run("should not find issues in good files", func(t *testing.T) {
		testCases := []string{
			"testdata/wk8001_good.go",
			"testdata/wk8002_good.go",
			"testdata/wk8003_good.go",
			"testdata/wk8004_good.go",
			"testdata/wk8005_good.go",
			"testdata/wk8006_good.go",
		}

		for _, file := range testCases {
			t.Run(file, func(t *testing.T) {
				issues, err := linter.LintFile(file)
				require.NoError(t, err)
				assert.Empty(t, issues, "Expected no issues in %s", file)
			})
		}
	})
}

func TestLinter_LintDirectory(t *testing.T) {
	t.Skip("TODO: Fix test data paths")
	linter := NewLinter(nil)

	t.Run("should lint all files in directory", func(t *testing.T) {
		issues, err := linter.LintDirectory("testdata")
		require.NoError(t, err)

		// We expect to find issues from all the bad files
		assert.NotEmpty(t, issues, "Expected to find issues in testdata directory")

		// Count issues by rule
		ruleCounts := make(map[string]int)
		for _, issue := range issues {
			ruleCounts[issue.Rule]++
		}

		// We should have issues from all 13 rules
		assert.Greater(t, ruleCounts["WK8001"], 0, "Expected WK8001 issues")
		assert.Greater(t, ruleCounts["WK8002"], 0, "Expected WK8002 issues")
		assert.Greater(t, ruleCounts["WK8003"], 0, "Expected WK8003 issues")
		assert.Greater(t, ruleCounts["WK8004"], 0, "Expected WK8004 issues")
		assert.Greater(t, ruleCounts["WK8005"], 0, "Expected WK8005 issues")
		assert.Greater(t, ruleCounts["WK8006"], 0, "Expected WK8006 issues")
		assert.Greater(t, ruleCounts["WK8041"], 0, "Expected WK8041 issues")
		assert.Greater(t, ruleCounts["WK8042"], 0, "Expected WK8042 issues")
		assert.Greater(t, ruleCounts["WK8101"], 0, "Expected WK8101 issues")
		assert.Greater(t, ruleCounts["WK8102"], 0, "Expected WK8102 issues")
		assert.Greater(t, ruleCounts["WK8201"], 0, "Expected WK8201 issues")
		assert.Greater(t, ruleCounts["WK8202"], 0, "Expected WK8202 issues")
		assert.Greater(t, ruleCounts["WK8301"], 0, "Expected WK8301 issues")
	})
}

func TestLinter_Lint(t *testing.T) {
	t.Skip("TODO: Fix test data paths")
	linter := NewLinter(nil)

	t.Run("should lint file when path is file", func(t *testing.T) {
		issues, err := linter.Lint("testdata/wk8001_bad.go")
		require.NoError(t, err)
		assert.NotEmpty(t, issues)
	})

	t.Run("should lint directory when path is directory", func(t *testing.T) {
		issues, err := linter.Lint("testdata")
		require.NoError(t, err)
		assert.NotEmpty(t, issues)
	})
}

func TestLinter_LintWithResult(t *testing.T) {
	t.Skip("TODO: Fix test data paths")
	linter := NewLinter(nil)

	t.Run("should return detailed result", func(t *testing.T) {
		result, err := linter.LintWithResult("testdata")
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
	t.Skip("TODO: Fix test data paths")
	t.Run("should filter issues by minimum severity", func(t *testing.T) {
		// Create linter that only reports errors
		config := &Config{
			MinSeverity: SeverityError,
		}
		linter := NewLinter(config)

		issues, err := linter.LintFile("testdata/wk8001_bad.go")
		require.NoError(t, err)

		// All issues should be errors or higher
		for _, issue := range issues {
			assert.GreaterOrEqual(t, issue.Severity, SeverityError)
		}
	})
}

func TestLinter_DisabledRules(t *testing.T) {
	t.Skip("TODO: Fix test data paths")
	t.Run("should not report issues for disabled rules", func(t *testing.T) {
		config := &Config{
			DisabledRules: []string{"WK8001"},
		}
		linter := NewLinter(config)

		issues, err := linter.LintFile("testdata/wk8001_bad.go")
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

	// The linter should have 4 rules (13 - 2 disabled)
	assert.Len(t, linter.rules, 11)
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
	// Should have all 13 rules enabled by default
	assert.Len(t, linter.rules, 13)
}

func TestLinter_DisableAllRules(t *testing.T) {
	config := &Config{
		DisabledRules: []string{"WK8001", "WK8002", "WK8003", "WK8004", "WK8005", "WK8006", "WK8041", "WK8042", "WK8101", "WK8102", "WK8201", "WK8202", "WK8301"},
	}
	linter := NewLinter(config)
	assert.Len(t, linter.rules, 0)
}
