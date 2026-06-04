package validator

// Severity describes the importance of a diagnostic.
type Severity string

const (
	SeverityError Severity = "error"
)

// Diagnostic describes a validation issue in Monaco-friendly coordinates.
type Diagnostic struct {
	Severity  Severity
	Message   string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}
