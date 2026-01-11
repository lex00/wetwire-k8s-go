package lint

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Linter represents the lint engine.
type Linter struct {
	config *Config
	rules  []Rule
}

// NewLinter creates a new linter with the given configuration.
func NewLinter(config *Config) *Linter {
	if config == nil {
		config = &Config{
			MinSeverity: SeverityInfo,
		}
	}

	// Get all available rules
	allRules := AllRules()

	// Filter out disabled rules
	var enabledRules []Rule
	for _, rule := range allRules {
		if !isRuleDisabled(rule.ID, config.DisabledRules) {
			enabledRules = append(enabledRules, rule)
		}
	}

	return &Linter{
		config: config,
		rules:  enabledRules,
	}
}

// isRuleDisabled checks if a rule ID is in the disabled list.
func isRuleDisabled(ruleID string, disabledRules []string) bool {
	for _, disabled := range disabledRules {
		if disabled == ruleID {
			return true
		}
	}
	return false
}

// LintFile lints a single Go source file.
func (l *Linter) LintFile(filePath string) ([]Issue, error) {
	// Create a new file set for position information
	fset := token.NewFileSet()

	// Parse the Go source file
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	var allIssues []Issue

	// Run each rule
	for _, rule := range l.rules {
		issues := rule.Check(file, fset)

		// Filter by minimum severity
		// Note: Lower severity values are more severe (Error=0, Warning=1, Info=2)
		for _, issue := range issues {
			if issue.Severity <= l.config.MinSeverity {
				allIssues = append(allIssues, issue)
			}
		}
	}

	return allIssues, nil
}

// LintDirectory lints all Go files in a directory recursively.
func (l *Linter) LintDirectory(dir string) ([]Issue, error) {
	var allIssues []Issue

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Lint this file
		issues, err := l.LintFile(path)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to lint %s: %v\n", path, err)
			return nil
		}

		allIssues = append(allIssues, issues...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return allIssues, nil
}

// Lint lints a file or directory.
// If the path is a file, it lints that file.
// If the path is a directory, it lints all Go files in that directory recursively.
func (l *Linter) Lint(path string) ([]Issue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if info.IsDir() {
		return l.LintDirectory(path)
	}

	return l.LintFile(path)
}

// LintResult represents the result of a lint operation.
type LintResult struct {
	Issues       []Issue
	TotalFiles   int
	FilesWithIssues int
	ErrorCount   int
	WarningCount int
	InfoCount    int
}

// LintWithResult lints a file or directory and returns a detailed result.
func (l *Linter) LintWithResult(path string) (*LintResult, error) {
	issues, err := l.Lint(path)
	if err != nil {
		return nil, err
	}

	result := &LintResult{
		Issues: issues,
	}

	// Count issues by severity
	filesWithIssues := make(map[string]bool)
	for _, issue := range issues {
		filesWithIssues[issue.File] = true

		switch issue.Severity {
		case SeverityError:
			result.ErrorCount++
		case SeverityWarning:
			result.WarningCount++
		case SeverityInfo:
			result.InfoCount++
		}
	}

	result.FilesWithIssues = len(filesWithIssues)

	// Count total files
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			totalFiles := 0
			filepath.Walk(path, func(p string, i os.FileInfo, e error) error {
				if e == nil && !i.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
					totalFiles++
				}
				return nil
			})
			result.TotalFiles = totalFiles
		} else {
			result.TotalFiles = 1
		}
	}

	return result, nil
}
