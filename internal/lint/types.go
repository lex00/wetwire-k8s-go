package lint

import (
	"go/ast"
	"go/token"

	corelint "github.com/lex00/wetwire-core-go/lint"
)

// Severity is an alias to the core lint severity type.
type Severity = corelint.Severity

// Severity constants from core lint package.
const (
	SeverityError   = corelint.SeverityError
	SeverityWarning = corelint.SeverityWarning
	SeverityInfo    = corelint.SeverityInfo
)

// Issue represents a lint issue found in the code.
type Issue struct {
	Rule     string   // Rule ID, e.g., "WK8001"
	Message  string   // Human-readable message
	File     string   // Source file path
	Line     int      // Line number
	Column   int      // Column number
	Severity Severity // Issue severity
}

// Rule represents a lint rule that can check Go source code.
type Rule struct {
	ID          string                                              // e.g., "WK8001"
	Name        string                                              // Human-readable name
	Description string                                              // Full description
	Severity    Severity                                            // Error, Warning, or Info
	Check       func(file *ast.File, fset *token.FileSet) []Issue  // Function to check the rule
	Fix         func(file *ast.File, issue Issue) error            // Optional auto-fix function
}

// Config represents the linter configuration.
type Config struct {
	DisabledRules []string // List of rule IDs to disable
	MinSeverity   Severity // Minimum severity to report
}

// Context provides context for rule execution.
type Context struct {
	FileSet *token.FileSet
	File    *ast.File
	Config  *Config
}
