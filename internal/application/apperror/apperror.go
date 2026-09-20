package apperror

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/util"
)

// Code categorizes an Error, letting callers branch on failure kind
// without string-matching Error().
type Code int

const (
	CodeUnknown Code = iota
	CodeInvalidConfig
	CodeProcessDiscoveryFailed
	CodeProcessNotFound
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
	// SeverityUnknown marks any unknown error.
	SeverityUnknown
)

// Error wraps an underlying error with a Code, a Severity, and a
// human-readable Message.
type Error struct {
	Code     Code
	Severity Severity
	Message  string
	Cause    error
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

// Handle logs err at the level its Severity calls for, then reports whether
// the caller must still treat the operation as failed. A Warning-severity
// error is absorbed here, so Handle reports success. A Fatal-severity error
// is returned after being logged, so the caller can terminate the program.
// An err that is not (and does not wrap) an *Error has SeverityUnknown,
// which — like SeverityError — is logged at error level and absorbed.
func Handle(err error) error {
	if err == nil {
		return nil
	}

	logger := util.NewLogger()
	switch severityOf(err) {
	case SeverityFatal:
		logger.Error(err.Error())
		return err
	case SeverityWarning:
		logger.Warn(err.Error())
	default:
		logger.Error(err.Error())
	}
	return nil
}

// severityOf reports error Severity if it is (or wraps) an *Error, or
// SeverityUnknown otherwise.
func severityOf(err error) Severity {
	var appErr *Error
	errors.As(err, &appErr)
	if appErr == nil {
		return SeverityUnknown
	}
	return appErr.Severity
}
