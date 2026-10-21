package apperror

import "errors"

// Code categorizes the kind of failure an Error represents.
type Code int

const (
	CodeUnknown Code = iota
	CodeInvalidConfig
	CodeProcessNotFound
	CodeGameFailed
	CodeKillFailed
	CodeStoreFailed
	CodePermissionDenied
)

// In reports whether the given error is an Error with this Code. It looks
// through plain wrapping but not into the cause of another Error, which is a
// separately classified problem.
func (c Code) In(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == c
}
