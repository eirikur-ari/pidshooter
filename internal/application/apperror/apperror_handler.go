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
func (h Handler) Handle(err error) error {
	var appErr *Error
	errors.As(err, &appErr)
	if appErr.Severity == SeverityWarning {
		h.logger.Warn(err.Error())
		return nil
	}
	h.logger.Error(err.Error())
	if appErr.Severity == SeverityFatal {
		return err
	}
	return nil
}