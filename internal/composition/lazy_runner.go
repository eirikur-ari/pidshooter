package composition

// lazyRunner runs each request on a runner it builds when Run is called.
type lazyRunner[Request any, Runner interface{ Run(Request) error }] struct {
	build func() (Runner, error)
}

// Run builds the runner and runs request on it. If the build fails, it returns
// that error without running request.
func (l *lazyRunner[Request, Runner]) Run(request Request) error {
	runner, err := l.build()
	if err != nil {
		return err
	}
	return runner.Run(request)
}
