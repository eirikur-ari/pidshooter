package inbound

import (
	"fmt"
	"io"
)

// ErrorCode categorizes an Error, letting callers branch on failure kind
// without string-matching Error().
type ErrorCode int

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeInvalidConfig
	ErrorCodeProcessDiscoveryFailed
	ErrorCodeNoProcessesFound
	ErrorCodeGameFailed
	ErrorCodeScoreLoadFailed
	ErrorCodeScoreSaveFailed
)

// ErrorSeverity tells the presentation layer whether an Error should abort
// the caller's operation or merely be reported alongside a completed one.
type ErrorSeverity int

const (
	// ErrorSeverityFatal is the zero value: an ErrorSeverity left unset
	// defaults to blocking rather than silently being treated as a mere
	// warning.
	ErrorSeverityFatal ErrorSeverity = iota
	ErrorSeverityWarning
)

// Error wraps an underlying error with a Code, a Severity, and a
// human-readable Message. A Fatal-severity Error is returned to the
// caller for handling; a Warning-severity one is never returned — it's
// printed at its point of origin instead.
type Error struct {
	Code     ErrorCode
	Severity ErrorSeverity
	Message  string
	Err      error
}

// NewError constructs an Error wrapping err with the given code, severity, and message.
func NewError(code ErrorCode, severity ErrorSeverity, message string, err error) *Error {
	return &Error{Code: code, Severity: severity, Message: message, Err: err}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Fprint writes e to w, prefixed by its kind ("warning: " or "error: ").
func (e *Error) Fprint(w io.Writer) {
	label := "warning"
	if e.Severity == ErrorSeverityFatal {
		label = "error"
	}
	fmt.Fprintf(w, "%s: %v\n", label, e)
}
