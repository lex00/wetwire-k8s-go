package lint

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	t.Run("should create text formatter by default", func(t *testing.T) {
		formatter := NewFormatter(FormatText)
		assert.IsType(t, &TextFormatter{}, formatter)
	})

	t.Run("should create JSON formatter", func(t *testing.T) {
		formatter := NewFormatter(FormatJSON)
		assert.IsType(t, &JSONFormatter{}, formatter)
	})

	t.Run("should create GitHub formatter", func(t *testing.T) {
		formatter := NewFormatter(FormatGitHub)
		assert.IsType(t, &GitHubFormatter{}, formatter)
	})

	t.Run("should default to text formatter for unknown format", func(t *testing.T) {
		formatter := NewFormatter("unknown")
		assert.IsType(t, &TextFormatter{}, formatter)
	})
}

func TestTextFormatter_Format(t *testing.T) {
	formatter := &TextFormatter{}

	t.Run("should format no issues", func(t *testing.T) {
		result := &LintResult{
			Issues:          []Issue{},
			TotalFiles:      5,
			FilesWithIssues: 0,
			ErrorCount:      0,
			WarningCount:    0,
			InfoCount:       0,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No issues found")
	})

	t.Run("should format issues with summary", func(t *testing.T) {
		result := &LintResult{
			Issues: []Issue{
				{
					File:     "test1.go",
					Line:     10,
					Column:   5,
					Rule:     "WK8001",
					Message:  "Test error",
					Severity: SeverityError,
				},
				{
					File:     "test2.go",
					Line:     20,
					Column:   3,
					Rule:     "WK8002",
					Message:  "Test warning",
					Severity: SeverityWarning,
				},
				{
					File:     "test1.go",
					Line:     15,
					Column:   8,
					Rule:     "WK8003",
					Message:  "Test info",
					Severity: SeverityInfo,
				},
			},
			TotalFiles:      2,
			FilesWithIssues: 2,
			ErrorCount:      1,
			WarningCount:    1,
			InfoCount:       1,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()
		// Check issues are formatted
		assert.Contains(t, output, "test1.go:10:5")
		assert.Contains(t, output, "error")
		assert.Contains(t, output, "[WK8001]")
		assert.Contains(t, output, "Test error")

		// Check summary
		assert.Contains(t, output, "Found 3 issue(s) in 2 file(s)")
		assert.Contains(t, output, "1 error(s)")
		assert.Contains(t, output, "1 warning(s)")
		assert.Contains(t, output, "1 info")
	})

	t.Run("should sort issues by file and line", func(t *testing.T) {
		result := &LintResult{
			Issues: []Issue{
				{File: "b.go", Line: 20, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
				{File: "a.go", Line: 10, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
				{File: "b.go", Line: 10, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
			},
			TotalFiles:      2,
			FilesWithIssues: 2,
			ErrorCount:      3,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(output, "\n")

		// Find issue lines (skip empty lines)
		var issueLines []string
		for _, line := range lines {
			if strings.Contains(line, ".go:") {
				issueLines = append(issueLines, line)
			}
		}

		// Check sorting
		assert.Contains(t, issueLines[0], "a.go:10")
		assert.Contains(t, issueLines[1], "b.go:10")
		assert.Contains(t, issueLines[2], "b.go:20")
	})
}

func TestJSONFormatter_Format(t *testing.T) {
	formatter := &JSONFormatter{}

	t.Run("should format as valid JSON", func(t *testing.T) {
		result := &LintResult{
			Issues: []Issue{
				{
					File:     "test.go",
					Line:     10,
					Column:   5,
					Rule:     "WK8001",
					Message:  "Test error",
					Severity: SeverityError,
				},
			},
			TotalFiles:      1,
			FilesWithIssues: 1,
			ErrorCount:      1,
			WarningCount:    0,
			InfoCount:       0,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		// Verify valid JSON
		var output JSONOutput
		err = json.Unmarshal(buf.Bytes(), &output)
		require.NoError(t, err)

		assert.Len(t, output.Issues, 1)
		assert.Equal(t, "test.go", output.Issues[0].File)
		assert.Equal(t, 10, output.Issues[0].Line)
		assert.Equal(t, "WK8001", output.Issues[0].Rule)
		assert.Equal(t, 1, output.TotalFiles)
		assert.Equal(t, 1, output.FilesWithIssues)
		assert.Equal(t, 1, output.ErrorCount)
		assert.Equal(t, 0, output.WarningCount)
		assert.Equal(t, 0, output.InfoCount)
	})

	t.Run("should handle empty issues", func(t *testing.T) {
		result := &LintResult{
			Issues:          nil,
			TotalFiles:      5,
			FilesWithIssues: 0,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		var output JSONOutput
		err = json.Unmarshal(buf.Bytes(), &output)
		require.NoError(t, err)

		// Should have empty array, not null
		assert.NotNil(t, output.Issues)
		assert.Len(t, output.Issues, 0)
	})
}

func TestGitHubFormatter_Format(t *testing.T) {
	formatter := &GitHubFormatter{}

	t.Run("should format no issues", func(t *testing.T) {
		result := &LintResult{
			Issues:          []Issue{},
			TotalFiles:      5,
			FilesWithIssues: 0,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No issues found")
	})

	t.Run("should format issues as GitHub annotations", func(t *testing.T) {
		result := &LintResult{
			Issues: []Issue{
				{
					File:     "test.go",
					Line:     10,
					Column:   5,
					Rule:     "WK8001",
					Message:  "Test error",
					Severity: SeverityError,
				},
				{
					File:     "test.go",
					Line:     20,
					Column:   3,
					Rule:     "WK8002",
					Message:  "Test warning",
					Severity: SeverityWarning,
				},
				{
					File:     "test.go",
					Line:     30,
					Column:   1,
					Rule:     "WK8003",
					Message:  "Test info",
					Severity: SeverityInfo,
				},
			},
			TotalFiles:      1,
			FilesWithIssues: 1,
			ErrorCount:      1,
			WarningCount:    1,
			InfoCount:       1,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()

		// Check GitHub annotations format
		assert.Contains(t, output, "::error file=test.go,line=10,col=5,title=WK8001::Test error")
		assert.Contains(t, output, "::warning file=test.go,line=20,col=3,title=WK8002::Test warning")
		assert.Contains(t, output, "::notice file=test.go,line=30,col=1,title=WK8003::Test info")

		// Check summary
		assert.Contains(t, output, "Found 3 issue(s) in 1 file(s)")
	})

	t.Run("should sort issues by file and line", func(t *testing.T) {
		result := &LintResult{
			Issues: []Issue{
				{File: "b.go", Line: 20, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
				{File: "a.go", Line: 10, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
				{File: "b.go", Line: 10, Column: 1, Rule: "WK8001", Message: "msg", Severity: SeverityError},
			},
			TotalFiles:      2,
			FilesWithIssues: 2,
			ErrorCount:      3,
		}

		var buf bytes.Buffer
		err := formatter.Format(result, &buf)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(output, "\n")

		// Find annotation lines
		var annotations []string
		for _, line := range lines {
			if strings.HasPrefix(line, "::") && strings.Contains(line, "file=") {
				annotations = append(annotations, line)
			}
		}

		// Check sorting
		assert.Contains(t, annotations[0], "file=a.go,line=10")
		assert.Contains(t, annotations[1], "file=b.go,line=10")
		assert.Contains(t, annotations[2], "file=b.go,line=20")
	})
}

func TestSummarizeByRule(t *testing.T) {
	t.Run("should summarize issues by rule", func(t *testing.T) {
		issues := []Issue{
			{Rule: "WK8001", Message: "msg1", Severity: SeverityError},
			{Rule: "WK8001", Message: "msg2", Severity: SeverityError},
			{Rule: "WK8002", Message: "msg3", Severity: SeverityWarning},
			{Rule: "WK8003", Message: "msg4", Severity: SeverityInfo},
		}

		rules := []Rule{
			{ID: "WK8001", Description: "Rule 1", Severity: SeverityError},
			{ID: "WK8002", Description: "Rule 2", Severity: SeverityWarning},
			{ID: "WK8003", Description: "Rule 3", Severity: SeverityInfo},
		}

		summary := SummarizeByRule(issues, rules)

		// Should have 3 unique rules
		assert.Len(t, summary, 3)

		// Find summaries by rule ID
		var wk8001, wk8002, wk8003 *IssueSummary
		for i := range summary {
			switch summary[i].RuleID {
			case "WK8001":
				wk8001 = &summary[i]
			case "WK8002":
				wk8002 = &summary[i]
			case "WK8003":
				wk8003 = &summary[i]
			}
		}

		require.NotNil(t, wk8001)
		assert.Equal(t, 2, wk8001.Count)
		assert.Equal(t, "Rule 1", wk8001.Description)
		assert.Equal(t, SeverityError, wk8001.Severity)

		require.NotNil(t, wk8002)
		assert.Equal(t, 1, wk8002.Count)

		require.NotNil(t, wk8003)
		assert.Equal(t, 1, wk8003.Count)
	})

	t.Run("should handle empty issues", func(t *testing.T) {
		issues := []Issue{}
		rules := AllRules()

		summary := SummarizeByRule(issues, rules)
		assert.Empty(t, summary)
	})

	t.Run("should sort by rule ID", func(t *testing.T) {
		issues := []Issue{
			{Rule: "WK8003", Message: "msg"},
			{Rule: "WK8001", Message: "msg"},
			{Rule: "WK8002", Message: "msg"},
		}

		rules := AllRules()
		summary := SummarizeByRule(issues, rules)

		assert.Equal(t, "WK8001", summary[0].RuleID)
		assert.Equal(t, "WK8002", summary[1].RuleID)
		assert.Equal(t, "WK8003", summary[2].RuleID)
	})
}

func TestIssueSummary_Fields(t *testing.T) {
	summary := IssueSummary{
		RuleID:      "WK8001",
		Count:       5,
		Severity:    SeverityError,
		Description: "Test description",
	}

	assert.Equal(t, "WK8001", summary.RuleID)
	assert.Equal(t, 5, summary.Count)
	assert.Equal(t, SeverityError, summary.Severity)
	assert.Equal(t, "Test description", summary.Description)
}
