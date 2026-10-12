package runner

import "github.com/eirikur-ari/pidshooter/internal/application/apperror"

func newFatalError(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityFatal, message, nil)
}

func newWarning(message string) *apperror.Error {
	return apperror.NewError(apperror.CodeUnknown, apperror.SeverityWarning, message, nil)
}
