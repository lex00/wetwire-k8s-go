package lint

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// OutputFormat represents the output format for lint results.
type OutputFormat string

const (
	// FormatText outputs human-readable text.
	FormatText OutputFormat = "text"
	// FormatJSON outputs JSON.
	FormatJSON OutputFormat = "json"
	// FormatGitHub outputs GitHub Actions format.
	FormatGitHub OutputFormat = "github"
)

// Formatter formats lint results for output.
type Formatter interface {
	Format(result *LintResult, w io.Writer) error
}

// NewFormatter creates a new formatter for the given format.
func NewFormatter(format OutputFormat) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatGitHub:
		return &GitHubFormatter{}
	default:
		return &TextFormatter{}
	}
}

// TextFormatter formats results as human-readable text.
type TextFormatter struct{}

// Format implements Formatter.
func (f *TextFormatter) Format(result *LintResult, w io.Writer) error {
	if len(result.Issues) == 0 {
		fmt.Fprintln(w, "No issues found.")
		return nil
	}

	// Sort issues by file, then line
	sortedIssues := make([]Issue, len(result.Issues))
	copy(sortedIssues, result.Issues)
	sort.Slice(sortedIssues, func(i, j int) bool {
		if sortedIssues[i].File != sortedIssues[j].File {
			return sortedIssues[i].File < sortedIssues[j].File
		}
		return sortedIssues[i].Line < sortedIssues[j].Line
	})

	// Print issues
	for _, issue := range sortedIssues {
		fmt.Fprintf(w, "%s:%d:%d: %s [%s] %s\n",
			issue.File,
			issue.Line,
			issue.Column,
			issue.Severity.String(),
			issue.Rule,
			issue.Message,
		)
	}

	// Print summary
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Found %d issue(s) in %d file(s):\n",
		len(result.Issues),
		result.FilesWithIssues,
	)
	if result.ErrorCount > 0 {
		fmt.Fprintf(w, "  - %d error(s)\n", result.ErrorCount)
	}
	if result.WarningCount > 0 {
		fmt.Fprintf(w, "  - %d warning(s)\n", result.WarningCount)
	}
	if result.InfoCount > 0 {
		fmt.Fprintf(w, "  - %d info\n", result.InfoCount)
	}

	return nil
}

// JSONFormatter formats results as JSON.
type JSONFormatter struct{}

// JSONOutput represents the JSON output structure.
type JSONOutput struct {
	Issues       []Issue `json:"issues"`
	TotalFiles   int     `json:"total_files"`
	FilesWithIssues int     `json:"files_with_issues"`
	ErrorCount   int     `json:"error_count"`
	WarningCount int     `json:"warning_count"`
	InfoCount    int     `json:"info_count"`
}

// Format implements Formatter.
func (f *JSONFormatter) Format(result *LintResult, w io.Writer) error {
	output := JSONOutput{
		Issues:       result.Issues,
		TotalFiles:   result.TotalFiles,
		FilesWithIssues: result.FilesWithIssues,
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		InfoCount:    result.InfoCount,
	}

	// Ensure Issues is not nil (empty array instead)
	if output.Issues == nil {
		output.Issues = []Issue{}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// GitHubFormatter formats results for GitHub Actions.
type GitHubFormatter struct{}

// Format implements Formatter.
func (f *GitHubFormatter) Format(result *LintResult, w io.Writer) error {
	if len(result.Issues) == 0 {
		fmt.Fprintln(w, "No issues found.")
		return nil
	}

	// Sort issues by file, then line
	sortedIssues := make([]Issue, len(result.Issues))
	copy(sortedIssues, result.Issues)
	sort.Slice(sortedIssues, func(i, j int) bool {
		if sortedIssues[i].File != sortedIssues[j].File {
			return sortedIssues[i].File < sortedIssues[j].File
		}
		return sortedIssues[i].Line < sortedIssues[j].Line
	})

	// Print GitHub Actions annotations
	for _, issue := range sortedIssues {
		level := "error"
		switch issue.Severity {
		case SeverityWarning:
			level = "warning"
		case SeverityInfo:
			level = "notice"
		}

		fmt.Fprintf(w, "::%s file=%s,line=%d,col=%d,title=%s::%s\n",
			level,
			issue.File,
			issue.Line,
			issue.Column,
			issue.Rule,
			issue.Message,
		)
	}

	// Print summary
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Found %d issue(s) in %d file(s)\n",
		len(result.Issues),
		result.FilesWithIssues,
	)

	return nil
}

// IssueSummary represents a summary of issues by rule.
type IssueSummary struct {
	RuleID      string
	Count       int
	Severity    Severity
	Description string
}

// SummarizeByRule creates a summary of issues grouped by rule.
func SummarizeByRule(issues []Issue, rules []Rule) []IssueSummary {
	ruleCounts := make(map[string]int)
	ruleMap := make(map[string]Rule)

	// Build rule map
	for _, rule := range rules {
		ruleMap[rule.ID] = rule
	}

	// Count issues by rule
	for _, issue := range issues {
		ruleCounts[issue.Rule]++
	}

	// Create summary
	var summary []IssueSummary
	for ruleID, count := range ruleCounts {
		rule := ruleMap[ruleID]
		summary = append(summary, IssueSummary{
			RuleID:      ruleID,
			Count:       count,
			Severity:    rule.Severity,
			Description: rule.Description,
		})
	}

	// Sort by rule ID
	sort.Slice(summary, func(i, j int) bool {
		return summary[i].RuleID < summary[j].RuleID
	})

	return summary
}
