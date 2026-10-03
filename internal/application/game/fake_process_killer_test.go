package game

// FakeProcessKiller adapts a plain function to a Kill method, for tests.
type FakeProcessKiller func(pid int, name string, protected bool) (shouldReap bool, err error)

func (f FakeProcessKiller) Kill(pid int, name string, protected bool) (shouldReap bool, err error) {
	return f(pid, name, protected)
}
