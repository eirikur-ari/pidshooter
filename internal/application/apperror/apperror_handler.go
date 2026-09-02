package apperror

import "errors"

// Logger is the minimal reporting capability a Handler needs.
type Logger interface {
	Warn(msg string)
	Error(msg string)
}

// Handler logs an *Error at the level its Severity calls for, then reports
// whether the caller must still treat the operation as failed.
type Handler struct {
	logger Logger
}

// NewHandler constructs a Handler that reports through logger.
func NewHandler(logger Logger) Handler {
	return Handler{logger: logger}
}

// Handle logs err at the level its Severity calls for, then reports whether
// the caller must still treat the operation as failed. A Warning-severity
// error is absorbed here, so Handle reports success. A Fatal-severity error
// is returned after being logged, so the caller can terminate the program.
// An err that is not (and does not wrap) an *Error has SeverityUnknown,
// which — like SeverityError — is logged at error level and absorbed.
func (h Handler) Handle(err error) error {
	if err == nil {
		return nil
	}

	switch severityOf(err) {
	case SeverityFatal:
		h.logger.Error(err.Error())
		return err
	case SeverityWarning:
		h.logger.Warn(err.Error())
	default:
		h.logger.Error(err.Error())
	}
	return nil
}

// severityOf reports err's Severity if it is (or wraps) an *Error, or
// SeverityUnknown otherwise.
func severityOf(err error) Severity {
	var appErr *Error
	errors.As(err, &appErr)
	if appErr == nil {
		return SeverityUnknown
	}
	return appErr.Severity
}