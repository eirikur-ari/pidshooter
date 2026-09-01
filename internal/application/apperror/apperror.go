package apperror

// Code categorizes an Error, letting callers branch on failure kind
// without string-matching Error().
type Code int

const (
	CodeUnknown Code = iota
	CodeInvalidConfig
	CodeProcessDiscoveryFailed
	CodeNoProcessesFound
	CodeGameFailed
	CodeScoreLoadFailed
	CodeScoreSaveFailed
	CodeKillFailed
)

// Severity tells the caller whether an Error should abort its operation,
// end it early while being logged as an error, or merely be reported
// alongside a completed one.
type Severity int

const (
	// SeverityFatal is the zero value: a Severity left unset defaults to
	// blocking rather than silently being treated as a mere warning.
	SeverityFatal Severity = iota
	// SeverityError is non-fatal — the program keeps running — but ends
	// the current operation early, so it's logged at error level rather
	// than warning level.
	SeverityError
	SeverityWarning
)

// Error wraps an underlying error with a Code, a Severity, and a
// human-readable Message.
type Error struct {
	Code     Code
	Severity Severity
	Message  string
	Err      error
}

// NewError constructs an Error wrapping err with the given code, severity, and message.
func NewError(code Code, severity Severity, message string, err error) *Error {
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
