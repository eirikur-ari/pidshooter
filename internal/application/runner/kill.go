package runner

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// Kill kills the processes that match search patterns.
type Kill struct{}

// NewKill returns a Kill.
func NewKill() *Kill {
	return &Kill{}
}

// Run returns an error stating that killing is unimplemented.
func (*Kill) Run(inbound.KillRequest) error {
	return errors.New("unimplemented")
}
