package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

// Dispatcher runs what a request describes.
type Dispatcher struct {
	play       inbound.PlayRunner
	list       inbound.ListRunner
	kill       inbound.KillRunner
	errHandler *apperror.Handler
}

// NewDispatcher returns a Dispatcher that runs play, list and kill requests
// with the given runners and handles errors with errHandler.
func NewDispatcher(
	play inbound.PlayRunner,
	list inbound.ListRunner,
	kill inbound.KillRunner,
	errHandler *apperror.Handler,
) *Dispatcher {
	return &Dispatcher{play: play, list: list, kill: kill, errHandler: errHandler}
}

// Run runs what request describes on the matching runner. It returns an error
// if request describes nothing or if the run could not complete.
func (d *Dispatcher) Run(request inbound.RunRequest) error {
	switch {
	case request.Play != nil:
		return d.play.Run(*request.Play)
	case request.List != nil:
		return d.list.Run(*request.List)
	case request.Kill != nil:
		return d.kill.Run(*request.Kill)
	default:
		return apperror.NewError(apperror.CodeUnknown, apperror.SeverityFatal, "nothing to run", nil)
	}
}

// HandleError logs err at the level its severity calls for. It returns nil if
// the error is absorbed, and err unchanged otherwise.
func (d *Dispatcher) HandleError(err error) error {
	return d.errHandler.Handle(err)
}
