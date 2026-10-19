package composition

import "github.com/eirikur-ari/pidshooter/internal/application/apperror"

// lazyRunner runs each request on a runner it builds when Run is called.
type lazyRunner[Request any, Runner interface{ Run(Request) error }] struct {
	build      func() (Runner, error)
	errHandler *apperror.Handler
}

// Run builds the runner and runs request on it. If the build fails, it handles
// that error and returns it without running request.
func (l *lazyRunner[Request, Runner]) Run(request Request) error {
	runner, err := l.build()
	if err != nil {
		// Without a runner the request cannot run, whatever the severity.
		_ = l.errHandler.Handle(err)
		return err
	}
	return runner.Run(request)
}
