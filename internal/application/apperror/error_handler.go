package apperror

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Handler logs errors passed to Handle and decides whether the caller must
// still treat the operation as failed.
type Handler struct {
	logger outbound.Logger
}

// NewHandler returns a Handler that logs through logger.
func NewHandler(logger outbound.Logger) *Handler {
	return &Handler{logger: logger}
}

// Handle logs err at the level its Severity calls for, then reports whether
// the caller must still treat the operation as failed: a Warning- or
// Error-severity err is absorbed (nil); a Fatal-severity err, or one of
// unknown severity, is returned unchanged. Safe to call more than once on
// the same err — it is logged only once.
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

// logOnce logs err at the level severity calls for, and marks appErr, if
// non-nil, as logged.
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
