package fake

// ProcessKiller adapts a plain function to a Kill method, for tests.
type ProcessKiller func(pid int, name string, protected bool) (shouldReap bool, err error)

func (f ProcessKiller) Kill(pid int, name string, protected bool) (shouldReap bool, err error) {
	return f(pid, name, protected)
}
