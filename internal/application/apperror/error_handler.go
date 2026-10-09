package apperror

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Handler logs errors according to their severity.
type Handler struct {
	logger outbound.Logger
}

// NewHandler returns a Handler that logs through logger.
func NewHandler(logger outbound.Logger) *Handler {
	return &Handler{logger: logger}
}

// Handle logs the given error at the level its Severity calls for, then reports
// whether the caller must still treat the operation as failed: an error of
// Warning or Error severity is absorbed (nil); one of Fatal or unknown severity
// is returned unchanged. An Error handled more than once is logged only once.
func (h *Handler) Handle(err error) error {
	if err == nil {
		return nil
	}

	var appErr *Error
	errors.As(err, &appErr)

	severity := appErr.severity()

	if appErr == nil || !appErr.logged {
		h.logOnce(err, severity, appErr)
	}

	if severity == SeverityFatal || severity == SeverityUnknown {
		return err
	}

	return nil
}

// HandleAll handles each of the given errors as Handle does, and returns the
// first one the caller must still treat as a failure, or nil if there is none.
// Every error is handled, even after one that is returned.
func (h *Handler) HandleAll(errs []error) error {
	var first error
	for _, err := range errs {
		if failure := h.Handle(err); failure != nil && first == nil {
			first = failure
		}
	}

	return first
}

// logOnce logs the error at the level its severity calls for, and marks the
// Error, if any, as logged.
func (h *Handler) logOnce(err error, severity Severity, appErr *Error) {
	switch severity {
	case SeverityWarning:
		h.logger.Warn(err.Error())
	default:
		h.logger.Error(err.Error())
	}

	if appErr != nil {
		appErr.logged = true
	}
}
