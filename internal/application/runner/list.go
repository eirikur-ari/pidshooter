package runner

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// List lists the processes that match search patterns.
type List struct {
	errHandler *apperror.Handler
}

// NewList returns a List that handles its errors with the given error handler.
func NewList(errHandler *apperror.Handler) *List {
	return &List{errHandler: errHandler}
}

// Run returns an error stating that listing is unimplemented, after handling
// it.
func (l *List) Run(inbound.ListRequest) error {
	return l.errHandler.Handle(errors.New("unimplemented"))
}
