package apperror

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/util"
)

// Code categorizes the kind of failure an Error represents.
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

// Handle logs err at the level its Severity calls for, then reports whether
// the caller must still treat the operation as failed: a Warning- or
// Error-severity err is absorbed (nil); a Fatal-severity err, or one of
// unknown severity, is returned unchanged. Safe to call more than once on
// the same err — it is logged only once.
func Handle(err error) error {
	if err == nil {
		return nil
	}

	var appErr *Error
	errors.As(err, &appErr)

	severity := severityOf(err)

	if appErr == nil || !appErr.logged {
		logOnce(err, severity, appErr)
	}

	if severity == SeverityFatal || severity == SeverityUnknown {
		return err
	}

	return nil
}

// logOnce logs err at the level severity calls for, and marks appErr, if
// non-nil, as logged.
func logOnce(err error, severity Severity, appErr *Error) {
	logger := util.NewLogger()

	switch severity {

	case SeverityWarning:
		logger.Warn(err.Error())
	default:
		logger.Error(err.Error())
	}
	if appErr != nil {
		appErr.logged = true
	}
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
