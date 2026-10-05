package apperror

import "errors"

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

// In reports whether the given error is an Error with this Code. It looks
// through plain wrapping but not into the cause of another Error, which is a
// separately classified problem.
func (c Code) In(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == c
}

// Severity tells the caller how an Error affects the operation it occurred in.
type Severity int

const (
	// SeverityUnknown marks an error of undetermined severity.
	SeverityUnknown Severity = iota
	// SeverityFatal marks an error that aborts the operation.
	SeverityFatal
	// SeverityError marks a non-fatal error that ends the current operation
	// early.
	SeverityError
	// SeverityWarning marks a non-fatal error reported alongside a completed
	// operation.
	SeverityWarning
)

// Error classifies a failure and wraps its underlying cause.
type Error struct {
	Code     Code
	Severity Severity
	message  string
	cause    error
	logged   bool
}

// NewError returns an Error that wraps cause.
func NewError(code Code, severity Severity, message string, cause error) *Error {
	return &Error{Code: code, Severity: severity, message: message, cause: cause}
}

// Error returns the message and the cause joined as "message: cause", or
// whichever of them is present.
func (e *Error) Error() string {
	if e.cause == nil {
		return e.message
	}

	if e.message == "" {
		return e.cause.Error()
	}

	return e.message + ": " + e.cause.Error()
}

// Unwrap returns the cause.
func (e *Error) Unwrap() error {
	return e.cause
}

// severity returns the Severity, or SeverityUnknown if the Error is nil.
func (e *Error) severity() Severity {
	if e == nil {
		return SeverityUnknown
	}

	return e.Severity
}
