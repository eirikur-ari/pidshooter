package game

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
)

// killDud records a target whose backing process was already gone before a
// kill could land on it.
type killDud struct {
	Name string
	PID  int
}

// toError returns the dud as a warning that the target's process was already gone.
func (d killDud) toError() *apperror.Error {
	message := fmt.Sprintf("%s (PID %d) ran away before it could be killed", d.Name, d.PID)
	return apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, message, nil)
}
