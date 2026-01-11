package roundtrip

// FieldPath represents a path to a field in a YAML document.
type FieldPath []string

// String returns the dot-separated path representation.
func (fp FieldPath) String() string {
	if len(fp) == 0 {
		return ""
	}
	result := fp[0]
	for i := 1; i < len(fp); i++ {
		result += "." + fp[i]
	}
	return result
}

// Append returns a new FieldPath with the given segment appended.
func (fp FieldPath) Append(segment string) FieldPath {
	newPath := make(FieldPath, len(fp)+1)
	copy(newPath, fp)
	newPath[len(fp)] = segment
	return newPath
}

// ComparisonMode specifies how YAML documents should be compared.
type ComparisonMode int

const (
	// ComparisonModeStrict requires exact equivalence of all fields.
	ComparisonModeStrict ComparisonMode = iota

	// ComparisonModeSemantic allows minor variations that don't affect meaning.
	// For example, different key ordering, numeric type differences (int vs float).
	ComparisonModeSemantic

	// ComparisonModeSubset checks that result contains all fields from original.
	// Allows extra fields in result that weren't in original.
	ComparisonModeSubset
)

// TestCase represents a single round-trip test case.
type TestCase struct {
	Name        string   // Name of the test case
	InputFile   string   // Path to input YAML file
	Description string   // Description of what's being tested
	Tags        []string // Tags for filtering tests (e.g., "deployment", "service")
}

// TestResult represents the result of running a test case.
type TestResult struct {
	TestCase    TestCase // The test case that was run
	Passed      bool     // Whether the test passed
	Result      *Result  // The round-trip result
	Error       error    // Error if the test failed
	Duration    int64    // Duration in milliseconds
	Differences []Difference
}

// TestSuite represents a collection of test cases.
type TestSuite struct {
	Name        string     // Name of the test suite
	Description string     // Description of the test suite
	TestCases   []TestCase // Test cases in the suite
}
