package runner

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// List lists the processes that match search patterns.
type List struct{}

// NewList returns a List.
func NewList() *List {
	return &List{}
}

// Run returns an error stating that listing is unimplemented.
func (*List) Run(inbound.ListRequest) error {
	return errors.New("unimplemented")
}
