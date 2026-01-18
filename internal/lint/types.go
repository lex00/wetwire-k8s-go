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

// Issue is an alias to the core lint Issue type.
type Issue = corelint.Issue

// Rule represents a lint rule that can check Go source code.
type Rule struct {
	ID          string                                              // e.g., "WK8001"
	Name        string                                              // Human-readable name
	Description string                                              // Full description
	Severity    Severity                                            // Error, Warning, or Info
	Check       func(file *ast.File, fset *token.FileSet) []Issue  // Function to check the rule
	Fix         func(file *ast.File, issue Issue) error            // Optional auto-fix function
}

// Config is an alias to the core lint Config type.
type Config = corelint.Config

// Context provides context for rule execution.
type Context struct {
	FileSet *token.FileSet
	File    *ast.File
	Config  *Config
}
