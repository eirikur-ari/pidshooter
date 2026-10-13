package game

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

// killFailure records a target that was not killed, along with the error that caused the failure.
type killFailure struct {
	Name string
	PID  int
	Err  error
}

// toError returns the failure as a warning that the target could not be killed.
func (f killFailure) toError() *apperror.Error {
	message := fmt.Sprintf("could not kill %s (PID %d)", f.Name, f.PID)
	return apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, message, f.Err)
}
