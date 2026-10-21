package runner

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// Kill kills the processes that match search patterns.
type Kill struct {
	errHandler *apperror.Handler
}

// NewKill returns a Kill that handles its errors with the given error handler.
func NewKill(errHandler *apperror.Handler) *Kill {
	return &Kill{errHandler: errHandler}
}

// Run returns an error stating that killing is unimplemented, after handling
// it.
func (k *Kill) Run(inbound.KillRequest) error {
	return k.errHandler.Handle(errors.New("unimplemented"))
}
