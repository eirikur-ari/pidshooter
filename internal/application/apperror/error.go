package apperror

// Code categorizes the kind of failure an Error represents.
type Code int

const (
	CodeUnknown Code = iota
	CodeInvalidConfig
	CodeProcessDiscoveryFailed
	CodeProcessNotFound
	CodeGameFailed
	CodeKillFailed
	CodeStoreLoadFailed
	CodeStoreSaveFailed
)

// Severity tells the caller whether an Error should abort its operation,
// end it early while being logged as an error, or merely be reported
// alongside a completed one.
type Severity int

const (
	// SeverityFatal marks an error that aborts the operation.
	SeverityFatal Severity = iota
	// SeverityError marks a non-fatal error that ends the current operation early.
	SeverityError
	// SeverityWarning marks a non-fatal error reported alongside a completed operation.
	SeverityWarning
	// SeverityUnknown marks an error of undetermined severity.
	SeverityUnknown
)

// Error wraps an underlying error with a Code, a Severity, and a
// human-readable Message.
type Error struct {
	Code     Code
	Severity Severity
	Message  string
	Cause    error
	logged   bool
}

// NewError constructs an Error wrapping error cause with the given error code, error severity, and error message.
func NewError(code Code, severity Severity, message string, cause error) *Error {
	return &Error{Code: code, Severity: severity, Message: message, Cause: cause}
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	if e.Message == "" {
		return e.Cause.Error()
	}

	return e.Message + ": " + e.Cause.Error()
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// severity reports e's Severity, or SeverityUnknown if e is nil.
func (e *Error) severity() Severity {
	if e == nil {
		return SeverityUnknown
	}

	return e.Severity
}
